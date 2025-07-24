package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DatabaseManager struct {
	connections map[string]*gorm.DB
	mutex       sync.RWMutex
}

var dbManager *DatabaseManager

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:100;not null"`
	Email     string    `json:"email" gorm:"size:100;uniqueIndex;not null"`
	Status    string    `json:"status" gorm:"size:20;default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	DSN      string `json:"dsn"`
}

func init() {
	dbManager = &DatabaseManager{
		connections: make(map[string]*gorm.DB),
	}

	initDefaultDatabase()
}

func initDefaultDatabase() {
	config := DatabaseConfig{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Database: "mcp_demo",
		Username: "root",
		Password: "root",
		DSN:      "root:root@tcp(localhost:3306)/mcp_demo?charset=utf8mb4&parseTime=True&loc=Local",
	}

	err := dbManager.AddConnection("default", config)
	if err != nil {
		log.Fatalf("初始化默认数据库连接失败: %v", err)
	}

	db, err := dbManager.GetConnection("default")
	if err != nil {
		log.Fatalf("获取默认数据库连接失败: %v", err)
	}

	err = db.AutoMigrate(&User{})
	if err != nil {
		log.Printf("自动迁移失败: %v", err)
		return
	}

	// 插入示例数据

	log.Println("默认数据库连接和自动迁移成功")
}

func (dm *DatabaseManager) GetConnection(name string) (*gorm.DB, error) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	if db, exists := dm.connections[name]; exists {
		return db, nil
	}
	return nil, fmt.Errorf("数据库连接 %s 不存在", name)
}

func (dm *DatabaseManager) AddConnection(name string, config DatabaseConfig) error {
	var dsn string
	if config.DSN != "" {
		dsn = config.DSN
	} else {
		charset := "utf8mb4"
		if charset == "" {
			charset = "utf8"
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			config.Username, config.Password, config.Host, config.Port, config.Database, charset)
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(mysql.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败 %s: %v", name, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("无法获取数据库连接 %s: %v", name, err)
	}

	err = sqlDB.Ping()
	if err != nil {
		return fmt.Errorf("无法连接到数据库 %s: %v", name, err)
	}

	//设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	dm.mutex.Lock()
	dm.connections[name] = db
	dm.mutex.Unlock()

	setData(db) // 插入示例数据
	log.Printf("MySQL数据库连接 %s 已添加", name)
	return nil
}

// 插入示例数据
func setData(db *gorm.DB) {
	var userCount int64
	db.Model(&User{}).Count(&userCount)
	if userCount == 0 {
		// 插入示例用户
		users := []User{
			{Name: "张三", Email: "zhangsan@example.com", Status: "active"},
			{Name: "李四", Email: "lisi@example.com", Status: "inactive"},
			{Name: "王五", Email: "wangwu@example.com", Status: "active"},
		}
		db.Create(&users)
		log.Println("插入示例用户数据")
	}
}

func main() {
	// 创建MCP服务器
	mcpServer := server.NewMCPServer(
		"Advance Go MCP server",
		"2.0.0",
		server.WithResourceCapabilities(true, true), // 支持静态和动态资源
		server.WithPromptCapabilities(true),         // 支持提示模板
		server.WithToolCapabilities(true),           // 支持工具
		server.WithRecovery(),                       // 错误恢复
		server.WithLogging(),                        // 启用日志
	)

	// 注册基础工具
	registerTools(mcpServer)

	// 注册高级工具
	//registerAdvancedTools(mcpServer)

	// 启动服务器
	log.Println("启动MCP服务器...")
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

// 注册工具
func registerTools(s *server.MCPServer) {
	// 计算器工具
	calculatorTool := mcp.NewTool("calculator",
		mcp.WithDescription("执行基本数学运算"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("要执行的运算 (add, subtract, multiply, divide)"),
			mcp.Enum("add", "subtract", "multiply", "divide"),
		),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("第一个数字"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("第二个数字"),
		),
	)
	s.AddTool(calculatorTool, handleCalculator)

	// 增强的数据库查询工具
	dbQueryTool := mcp.NewTool("database_query",
		mcp.WithDescription("执行数据库查询，支持原始SQL和结构化查询"),
		mcp.WithString("query_type",
			mcp.DefaultString("raw"),
			mcp.Description("查询类型: raw(原始SQL), structured(结构化查询), model(模型查询)"),
			mcp.Enum("raw", "structured", "model"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("查询内容：raw类型为SQL语句，structured类型为操作类型(select/count/insert/update/delete)，model类型为操作名称"),
		),
		mcp.WithString("database",
			mcp.DefaultString("default"),
			mcp.Description("数据库连接名称"),
		),
		mcp.WithString("table_name",
			mcp.Description("表名(structured查询必需)"),
		),
		mcp.WithString("fields",
			mcp.DefaultString("*"),
			mcp.Description("要查询的字段，多个字段用逗号分隔，默认为*"),
		),
		mcp.WithString("where_conditions",
			mcp.Description("WHERE条件，格式：field1=value1,field2>value2 或 JSON格式"),
		),
		mcp.WithString("order_by",
			mcp.Description("排序字段，格式：field1 ASC,field2 DESC"),
		),
		mcp.WithNumber("limit",
			mcp.Description("限制返回记录数"),
		),
		mcp.WithNumber("offset",
			mcp.Description("偏移量，用于分页"),
		),
		mcp.WithString("group_by",
			mcp.Description("分组字段"),
		),
		mcp.WithString("having",
			mcp.Description("HAVING条件"),
		),
		mcp.WithString("join_tables",
			mcp.Description("关联表信息，JSON格式：[{\"table\":\"table2\",\"on\":\"table1.id=table2.user_id\",\"type\":\"LEFT\"}]"),
		),
		mcp.WithString("model_name",
			mcp.Description("模型名称(model查询类型使用)"),
		),
	)
	s.AddTool(dbQueryTool, handleDatabaseQuery)

	// 搜索工具
	searchTool := mcp.NewTool("web_search",
		mcp.WithDescription("网络搜索"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("搜索关键词"),
		),
		mcp.WithNumber("limit",
			mcp.DefaultNumber(10),
			mcp.Description("结果数量限制"),
		),
	)
	s.AddTool(searchTool, handleWebSearch)
}

// 计算器工具处理函数
func handleCalculator(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	operation, err := request.RequireString("operation")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	x, err := request.RequireFloat("x")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	y, err := request.RequireFloat("y")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var result float64
	switch operation {
	case "add":
		result = x + y
	case "subtract":
		result = x - y
	case "multiply":
		result = x * y
	case "divide":
		if y == 0 {
			return mcp.NewToolResultError("除数不能为零"), nil
		}
		result = x / y
	default:
		return mcp.NewToolResultError("不支持的运算"), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("计算结果: %.2f %s %.2f = %.2f", x, operation, y, result)), nil
}

// 数据库查询工具处理函数
func handleDatabaseQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	queryType := request.GetString("query_type", "raw")
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	database := request.GetString("database", "default")

	// 获取MySQL数据库连接
	db, err := dbManager.GetConnection(database)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var result string

	switch queryType {
	case "raw":
		result, err = executeRawQuery(db, query)
	case "structured":
		result, err = executeStructuredQuery(db, request)
	case "model":
		modelName := request.GetString("model_name", "")
		result, err = executeModelQuery(db, modelName, query)
	default:
		return mcp.NewToolResultError("不支持的查询类型: " + queryType), nil
	}

	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(result), nil
}

func executeStructuredQuery(db *gorm.DB, request mcp.CallToolRequest) (string, error) {
	operation := request.GetString("query", "")
	tableName := request.GetString("table_name", "")

	if tableName == "" {
		return "", fmt.Errorf("结构化查询必须指定table_name参数")
	}

	switch strings.ToLower(operation) {
	case "select":
		return executeStructuredSelect(db, request)
	case "count":
		return executeStructuredCount(db, request)
	case "insert":
		return executeStructuredInsert(db, request)
	case "update":
		return executeStructuredUpdate(db, request)
	case "delete":
		return executeStructuredDelete(db, request)
	default:
		return "", fmt.Errorf("不支持的结构化查询操作: %s", operation)
	}
}

// 结构化SELECT查询
func executeStructuredSelect(db *gorm.DB, request mcp.CallToolRequest) (string, error) {
	tableName := request.GetString("table_name", "")
	fields := request.GetString("fields", "*")
	whereConditions := request.GetString("where_conditions", "")
	orderBy := request.GetString("order_by", "")
	groupBy := request.GetString("group_by", "")
	having := request.GetString("having", "")
	joinTables := request.GetString("join_tables", "")
	limit := request.GetInt("limit", 0)
	offset := request.GetInt("offset", 0)

	// 构建查询
	query := db.Table(tableName)

	// 处理字段选择
	if fields != "*" {
		query = query.Select(fields)
	}

	// 处理JOIN
	if joinTables != "" {
		var joins []map[string]interface{}
		if err := json.Unmarshal([]byte(joinTables), &joins); err == nil {
			for _, join := range joins {
				joinType := join["type"].(string)
				joinTable := join["table"].(string)
				joinOn := join["on"].(string)
				query = query.Joins(fmt.Sprintf("%s JOIN %s ON %s", joinType, joinTable, joinOn))
			}
		}
	}

	// 处理WHERE条件
	if whereConditions != "" {
		query = applyWhereConditions(query, whereConditions)
	}

	// 处理GROUP BY
	if groupBy != "" {
		query = query.Group(groupBy)
	}

	// 处理HAVING
	if having != "" {
		query = query.Having(having)
	}

	// 处理ORDER BY
	if orderBy != "" {
		query = query.Order(orderBy)
	}

	// 处理LIMIT和OFFSET
	if limit > 0 {
		query = query.Limit(int(limit))
	}
	if offset > 0 {
		query = query.Offset(int(offset))
	}

	// 执行查询
	var results []map[string]interface{}
	err := query.Find(&results).Error
	if err != nil {
		return "", err
	}

	// 格式化结果
	if len(results) == 0 {
		return fmt.Sprintf("表 %s 查询结果为空", tableName), nil
	}

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("表 %s 查询成功，返回 %d 条记录：\n%s", tableName, len(results), string(jsonData)), nil
}

// 结构化COUNT查询
func executeStructuredCount(db *gorm.DB, request mcp.CallToolRequest) (string, error) {
	tableName := request.GetString("table_name", "")
	whereConditions := request.GetString("where_conditions", "")
	groupBy := request.GetString("group_by", "")

	query := db.Table(tableName)

	// 处理WHERE条件
	if whereConditions != "" {
		query = applyWhereConditions(query, whereConditions)
	}

	// 处理GROUP BY
	if groupBy != "" {
		query = query.Group(groupBy)
		var results []map[string]interface{}
		err := query.Select(groupBy + ", COUNT(*) as count").Find(&results).Error
		if err != nil {
			return "", err
		}

		jsonData, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("表 %s 分组统计结果：\n%s", tableName, string(jsonData)), nil
	}

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("表 %s 记录总数：%d", tableName, count), nil
}

// 结构化INSERT查询
func executeStructuredInsert(db *gorm.DB, request mcp.CallToolRequest) (string, error) {
	tableName := request.GetString("table_name", "")
	fields := request.GetString("fields", "")

	if fields == "" {
		return "", fmt.Errorf("INSERT操作必须指定fields参数，格式：{\"field1\":\"value1\",\"field2\":\"value2\"}")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(fields), &data); err != nil {
		return "", fmt.Errorf("fields参数格式错误，必须是有效的JSON格式")
	}

	result := db.Table(tableName).Create(&data)
	if result.Error != nil {
		return "", result.Error
	}

	return fmt.Sprintf("成功向表 %s 插入 %d 条记录", tableName, result.RowsAffected), nil
}

// 结构化UPDATE查询
func executeStructuredUpdate(db *gorm.DB, request mcp.CallToolRequest) (string, error) {
	tableName := request.GetString("table_name", "")
	fields := request.GetString("fields", "")
	whereConditions := request.GetString("where_conditions", "")

	if fields == "" {
		return "", fmt.Errorf("UPDATE操作必须指定fields参数")
	}
	if whereConditions == "" {
		return "", fmt.Errorf("UPDATE操作必须指定where_conditions参数，以防止误操作")
	}

	var updateData map[string]interface{}
	if err := json.Unmarshal([]byte(fields), &updateData); err != nil {
		return "", fmt.Errorf("fields参数格式错误，必须是有效的JSON格式")
	}

	query := db.Table(tableName)
	query = applyWhereConditions(query, whereConditions)

	result := query.Updates(updateData)
	if result.Error != nil {
		return "", result.Error
	}

	return fmt.Sprintf("成功更新表 %s 中的 %d 条记录", tableName, result.RowsAffected), nil
}

// 结构化DELETE查询
func executeStructuredDelete(db *gorm.DB, request mcp.CallToolRequest) (string, error) {
	tableName := request.GetString("table_name", "")
	whereConditions := request.GetString("where_conditions", "")

	if whereConditions == "" {
		return "", fmt.Errorf("DELETE操作必须指定where_conditions参数，以防止误删除所有数据")
	}

	query := db.Table(tableName)
	query = applyWhereConditions(query, whereConditions)

	result := query.Delete(nil)
	if result.Error != nil {
		return "", result.Error
	}

	return fmt.Sprintf("成功从表 %s 删除 %d 条记录", tableName, result.RowsAffected), nil
}

// 应用WHERE条件的辅助函数
func applyWhereConditions(query *gorm.DB, whereConditions string) *gorm.DB {
	// 尝试解析为JSON格式
	var jsonConditions map[string]interface{}
	if err := json.Unmarshal([]byte(whereConditions), &jsonConditions); err == nil {
		// JSON格式条件
		for field, value := range jsonConditions {
			query = query.Where(fmt.Sprintf("%s = ?", field), value)
		}
		return query
	}

	// 简单格式条件：field1=value1,field2>value2
	conditions := strings.Split(whereConditions, ",")
	for _, condition := range conditions {
		condition = strings.TrimSpace(condition)
		if condition == "" {
			continue
		}

		// 支持的操作符
		operators := []string{">=", "<=", "!=", "<>", ">", "<", "=", "LIKE", "like"}
		var field, operator, value string

		for _, op := range operators {
			if strings.Contains(condition, op) {
				parts := strings.SplitN(condition, op, 2)
				if len(parts) == 2 {
					field = strings.TrimSpace(parts[0])
					operator = op
					value = strings.TrimSpace(parts[1])
					break
				}
			}
		}

		if field != "" && operator != "" && value != "" {
			// 移除值两边的引号
			value = strings.Trim(value, "'\"")
			query = query.Where(fmt.Sprintf("%s %s ?", field, operator), value)
		}
	}

	return query
}

func executeRawQuery(db *gorm.DB, query string) (string, error) {
	// 安全检查：只允许SELECT查询
	queryLower := strings.ToLower(strings.TrimSpace(query))
	if !strings.HasPrefix(queryLower, "select") {
		return "", fmt.Errorf("只允许执行SELECT查询")
	}

	var results []map[string]interface{}
	err := db.Raw(query).Scan(&results).Error
	if err != nil {
		return "", err
	}

	// 格式化结果
	if len(results) == 0 {
		return "MySQL查询结果为空", nil
	}

	// 转化为JSON格式
	jsonData, err := json.MarshalIndent(results, "", "")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("MySQL查询成功,返回 %d 条记录:\n%s", len(results), string(jsonData)), nil
}

func executeModelQuery(db *gorm.DB, modelName, operation string) (string, error) {
	switch strings.ToLower(modelName) {
	case "users":
		return queryUsers(db, operation)
	default:
		return "", fmt.Errorf("不支持的模型查询: %s", modelName)
	}
	return "", nil
}

// 查询用户数据
func queryUsers(db *gorm.DB, operation string) (string, error) {
	var users []User
	var err error

	switch strings.ToLower(operation) {
	case "all", "list":
		err = db.Find(&users).Error
	case "active":
		err = db.Where("status = ?", "active").Find(&users).Error
	case "inactive":
		err = db.Where("status = ?", "inactive").Find(&users).Error
	case "count":
		var count int64
		err = db.Model(&User{}).Count(&count).Error
		if err != nil {
			return "", fmt.Errorf("查询用户数量失败: %v", err)
		}
		return fmt.Sprintf("用户总数: %d", count), nil
	case "recent":
		err = db.Order("created_at DESC").Limit(10).Find(&users).Error
	default:
		return "", fmt.Errorf("不支持的用户查询操作: %s", operation)
	}

	if err != nil {
		return "", fmt.Errorf("查询用户失败: %v", err)
	}

	if len(users) == 0 {
		return "未找到用户数据", nil
	}

	// 转换为JSON格式
	jsonData, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化用户数据失败: %v", err)
	}

	return fmt.Sprintf("查询成功,返回 %d 条记录:\n%s", len(users), string(jsonData)), nil
}

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// GoogleSearchResponse represents the response from GoogleSearchResponse API
type GoogleSearchResponse struct {
	Items []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"items"`
	SearchInfoformation struct {
		TotalResults string `json:"totalResults"`
	} `json:"searchInformation"`
}

// 网络搜索工具处理函数
func handleWebSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	limit, err := request.RequireFloat("limit")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 限制搜索结果数量，避免过多请求
	if limit > 20 {
		limit = 20
	}
	if limit < 1 {
		limit = 10
	}

	// 执行真实的网络搜索
	results, err := performWebSearch(ctx, query, int(limit))
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索失败: %v", err)), nil
	}

	// 格式化搜索结果
	var resultText strings.Builder
	resultText.WriteString(fmt.Sprintf("网络搜索结果 - 关键词: '%s'\n", query))
	resultText.WriteString(fmt.Sprintf("找到 %d 条结果:\n\n", len(results)))

	for i, result := range results {
		resultText.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Title))
		resultText.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		if result.Snippet != "" {
			// 限制摘要长度
			snippet := result.Snippet
			if len(snippet) > 200 {
				snippet = snippet[:200] + "..."
			}
			resultText.WriteString(fmt.Sprintf("   摘要: %s\n", snippet))
		}
		resultText.WriteString("\n")
	}

	if len(results) == 0 {
		resultText.WriteString("未找到相关搜索结果")
	}

	return mcp.NewToolResultText(resultText.String()), nil
}

// performWebSearch performs actual web search using DuckDuckGo API
func performWebSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	searchEngineID := os.Getenv("GOOGLE_SEARCH_ENGINE_ID")

	if apiKey == "" || searchEngineID == "" {
		return nil, fmt.Errorf("未配置Google API密钥或搜索引擎ID")
	}

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 构建Google Custom Search API URL
	searchURL := fmt.Sprintf(
		"https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&num=%d",
		apiKey,
		searchEngineID,
		url.QueryEscape(query),
		limit,
	)

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置User-Agent
	req.Header.Set("User-Agent", "MCP-Client/1.0")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google搜索API返回错误状态: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取Google搜索API响应失败: %v", err)
	}

	// 解析JSON响应
	var googleResp GoogleSearchResponse
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("解析Google搜索API响应失败: %v", err)
	}

	var results []SearchResult
	for _, item := range googleResp.Items {
		if len(results) >= limit {
			break
		}

		results = append(results, SearchResult{
			Title:   item.Title,
			URL:     item.Link,
			Snippet: item.Snippet,
		})
	}

	return results, nil
}
