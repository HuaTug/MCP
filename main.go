package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

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
	registerAdvancedTools(mcpServer)

	// 注册资源
	registerResources(mcpServer)

	// 注册提示模板
	registerPrompts(mcpServer)

	// 启动服务器
	log.Println("启动MCP服务器...")
	if err := server.ServeStdio(mcpServer); err != nil {
		log.Fatalf("服务器错误: %v", err)
	}
}

// 注册工具
func registerTools(s *server.MCPServer) {
	// 1. 计算器工具
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

	// 2. 文件操作工具
	fileReadTool := mcp.NewTool("read_file",
		mcp.WithDescription("读取文件内容"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("文件路径"),
		),
	)
	s.AddTool(fileReadTool, handleReadFile)

	fileWriteTool := mcp.NewTool("write_file",
		mcp.WithDescription("写入文件内容"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("文件路径"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("要写入的内容"),
		),
	)
	s.AddTool(fileWriteTool, handleWriteFile)

	// 3. HTTP请求工具
	httpTool := mcp.NewTool("http_request",
		mcp.WithDescription("发送HTTP请求"),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("请求URL"),
		),
		mcp.WithString("method",
			mcp.DefaultString("GET"),
			mcp.Description("HTTP方法"),
			mcp.Enum("GET", "POST", "PUT", "DELETE"),
		),
		mcp.WithString("body",
			mcp.Description("请求体（仅用于POST/PUT）"),
		),
	)
	s.AddTool(httpTool, handleHTTPRequest)

	// 4. 系统信息工具
	systemInfoTool := mcp.NewTool("system_info",
		mcp.WithDescription("获取系统信息"),
	)
	s.AddTool(systemInfoTool, handleSystemInfo)

	// 5. 时间工具
	timeTool := mcp.NewTool("current_time",
		mcp.WithDescription("获取当前时间"),
		mcp.WithString("format",
			mcp.DefaultString("2006-01-02 15:04:05"),
			mcp.Description("时间格式"),
		),
		mcp.WithString("timezone",
			mcp.DefaultString("Local"),
			mcp.Description("时区"),
		),
	)
	s.AddTool(timeTool, handleCurrentTime)
}

// 注册高级工具
func registerAdvancedTools(s *server.MCPServer) {
	// 1.实时搜索工具
	webSearchTool := mcp.NewTool("web_search",
		mcp.WithDescription("执行实时网络搜索"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("搜索关键词查询"),
		),
		mcp.WithString("engine",
			mcp.DefaultString("google"),
			mcp.Description("搜索引擎"),
			mcp.Enum("google", "bing", "duckduckgo"),
		),
		mcp.WithNumber("results",
			mcp.DefaultNumber(10),
			mcp.Description("返回结果数量"),
		),
	)
	s.AddTool(webSearchTool, handleWebSearch)

	// 2. 数据库操作工具
	dbQueryTool := mcp.NewTool("database_query",
		mcp.WithDescription("执行数据库查询"),
		mcp.WithString("connection_string",
			mcp.Required(),
			mcp.Description("数据库连接字符串"),
		),
		mcp.WithString("driver",
			mcp.Required(),
			mcp.Description("数据库驱动"),
			mcp.Enum("mysql", "postgres", "sqlite3"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("SQL查询语句"),
		),
	)
	s.AddTool(dbQueryTool, handleDatabaseQuery)

	// 3. 文件分析工具
	fileAnalysisTool := mcp.NewTool("analyze_file",
		mcp.WithDescription("深度分析文件信息"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("文件路径"),
		),
		mcp.WithString("analysis_type",
			mcp.DefaultString("comprehensive"),
			mcp.Description("分析类型"),
			mcp.Enum("basic", "comprehensive", "security", "content"),
		),
	)
	s.AddTool(fileAnalysisTool, handleFileAnalysis)

	// 4. 目录扫描工具
	dirScanTool := mcp.NewTool("scan_directory",
		mcp.WithDescription("扫描目录结构和文件信息"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("目录路径"),
		),
		mcp.WithBoolean("recursive",
			mcp.DefaultBool(true),
			mcp.Description("是否递归扫描"),
		),
		mcp.WithString("filter",
			mcp.Description("文件过滤器（正则表达式）"),
		),
		mcp.WithNumber("max_depth",
			mcp.DefaultFloat(5),
			mcp.Description("最大扫描深度"),
		),
	)
	s.AddTool(dirScanTool, handleDirectoryScan)

	// 5. 网络监控工具
	networkMonitorTool := mcp.NewTool("network_monitor",
		mcp.WithDescription("网络连接和端口监控"),
		mcp.WithString("target",
			mcp.Required(),
			mcp.Description("目标主机或IP"),
		),
		mcp.WithString("action",
			mcp.Required(),
			mcp.Description("监控动作"),
			mcp.Enum("ping", "port_scan", "trace_route"),
		),
		mcp.WithString("ports",
			mcp.Description("端口范围（如：80,443,8080-8090）"),
		),
	)
	s.AddTool(networkMonitorTool, handleNetworkMonitor)

	// 6. 数据处理工具
	dataProcessTool := mcp.NewTool("process_data",
		mcp.WithDescription("处理和分析结构化数据"),
		mcp.WithString("data",
			mcp.Required(),
			mcp.Description("输入数据（JSON/CSV格式）"),
		),
		mcp.WithString("format",
			mcp.Required(),
			mcp.Description("数据格式"),
			mcp.Enum("json", "csv", "xml"),
		),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("处理操作"),
			mcp.Enum("analyze", "filter", "transform", "aggregate"),
		),
		mcp.WithString("parameters",
			mcp.Description("操作参数（JSON格式）"),
		),
	)
	s.AddTool(dataProcessTool, handleDataProcess)

	// 7. 代码分析工具
	codeAnalysisTool := mcp.NewTool("analyze_code",
		mcp.WithDescription("分析代码质量和结构"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("代码文件或目录路径"),
		),
		mcp.WithString("language",
			mcp.Description("编程语言（自动检测如果未指定）"),
			mcp.Enum("go", "python", "javascript", "java", "cpp", "auto"),
		),
		mcp.WithString("analysis_type",
			mcp.DefaultString("quality"),
			mcp.Description("分析类型"),
			mcp.Enum("quality", "complexity", "security", "dependencies"),
		),
	)
	s.AddTool(codeAnalysisTool, handleCodeAnalysis)

	// 8. 系统监控工具
	systemMonitorTool := mcp.NewTool("system_monitor",
		mcp.WithDescription("系统资源监控"),
		mcp.WithString("metric",
			mcp.Required(),
			mcp.Description("监控指标"),
			mcp.Enum("cpu", "memory", "disk", "network", "processes", "all"),
		),
		mcp.WithNumber("duration",
			mcp.DefaultFloat(5),
			mcp.Description("监控持续时间（秒）"),
		),
	)
	s.AddTool(systemMonitorTool, handleSystemMonitor)
}

func handleWebSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	engine := request.GetString("engine", "google")
	limit := int(request.GetFloat("limit", 10))

	// 构建搜索URL
	var searchURL string
	switch engine {
	case "google":
		searchURL = fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=YOUR_API_KEY&cx=YOUR_CX&q=%s&num=%d",
			url.QueryEscape(query), limit)
	case "bing":
		searchURL = fmt.Sprintf("https://api.bing.microsoft.com/v7.0/search?q=%s&count=%d",
			url.QueryEscape(query), limit)
	case "duckduckgo":
		// DuckDuckGo Instant Answer API
		searchURL = fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1",
			url.QueryEscape(query))
	default:
		return mcp.NewToolResultError("不支持的搜索引擎"), nil
	}

	// 发送搜索请求
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("创建搜索请求失败: %v", err)), nil
	}

	// 设置请求头
	if engine == "bing" {
		req.Header.Set("Ocp-Apim-Subscription-Key", "YOUR_BING_API_KEY")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取搜索结果失败: %v", err)), nil
	}

	// 解析搜索结果
	var results []map[string]interface{}
	if err := json.Unmarshal(respBody, &results); err != nil {
		// 如果解析失败，返回原始结果
		return mcp.NewToolResultText(fmt.Sprintf("🔍 搜索结果 (%s):\n%s", engine, string(respBody))), nil
	}

	// 格式化搜索结果
	var formattedResults strings.Builder
	formattedResults.WriteString(fmt.Sprintf("🔍 搜索结果 - \"%s\" (%s引擎):\n\n", query, engine))

	for i, result := range results {
		if i >= limit {
			break
		}
		title := getStringFromMap(result, "title")
		link := getStringFromMap(result, "link")
		snippet := getStringFromMap(result, "snippet")

		formattedResults.WriteString(fmt.Sprintf("%d. **%s**\n", i+1, title))
		formattedResults.WriteString(fmt.Sprintf("   🔗 %s\n", link))
		formattedResults.WriteString(fmt.Sprintf("   📝 %s\n\n", snippet))
	}

	return mcp.NewToolResultText(formattedResults.String()), nil
}

// 数据库查询处理函数
func handleDatabaseQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	connectionString, err := request.RequireString("connection_string")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	driver, err := request.RequireString("driver")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 连接数据库
	db, err := sql.Open(driver, connectionString)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("数据库连接失败: %v", err)), nil
	}
	defer db.Close()

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("数据库连接测试失败: %v", err)), nil
	}

	// 执行查询
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询执行失败: %v", err)), nil
	}
	defer rows.Close()

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取列信息失败: %v", err)), nil
	}

	// 读取查询结果
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("扫描行数据失败: %v", err)), nil
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	// 格式化结果
	jsonData, _ := json.MarshalIndent(map[string]interface{}{
		"query":     query,
		"driver":    driver,
		"row_count": len(results),
		"columns":   columns,
		"results":   results,
	}, "", "  ")

	return mcp.NewToolResultText(fmt.Sprintf("📊 数据库查询结果:\n%s", string(jsonData))), nil
}

// 文件分析处理函数
func handleFileAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	analysisType := request.GetString("analysis_type", "comprehensive")

	// 获取文件信息
	fileInfo, err := os.Stat(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取文件信息失败: %v", err)), nil
	}

	analysis := map[string]interface{}{
		"path":         path,
		"name":         fileInfo.Name(),
		"size":         fileInfo.Size(),
		"mode":         fileInfo.Mode().String(),
		"modified":     fileInfo.ModTime().Format(time.RFC3339),
		"is_directory": fileInfo.IsDir(),
	}

	// 基础分析
	if analysisType == "basic" || analysisType == "comprehensive" {
		analysis["extension"] = filepath.Ext(path)
		analysis["directory"] = filepath.Dir(path)

		// 文件哈希
		if !fileInfo.IsDir() {
			hash, err := calculateFileHash(path)
			if err == nil {
				analysis["md5_hash"] = hash
			}
		}
	}

	// 内容分析
	if analysisType == "content" || analysisType == "comprehensive" {
		if !fileInfo.IsDir() {
			contentAnalysis, err := analyzeFileContent(path)
			if err == nil {
				analysis["content_analysis"] = contentAnalysis
			}
		}
	}

	// 安全分析
	if analysisType == "security" || analysisType == "comprehensive" {
		securityAnalysis := analyzeFileSecurity(path, fileInfo)
		analysis["security_analysis"] = securityAnalysis
	}

	// 综合分析
	if analysisType == "comprehensive" {
		// 添加更多分析维度
		analysis["readable"] = isFileReadable(path)
		analysis["writable"] = isFileWritable(path)
		analysis["executable"] = isFileExecutable(fileInfo)

		if !fileInfo.IsDir() {
			analysis["mime_type"] = detectMimeType(path)
		}
	}

	jsonData, _ := json.MarshalIndent(analysis, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("📁 文件分析报告 (%s):\n%s", analysisType, string(jsonData))), nil
}

// 目录扫描处理函数
func handleDirectoryScan(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	recursive := request.GetBool("recursive", true)
	filterPattern := request.GetString("filter", "")
	maxDepth := int(request.GetFloat("max_depth", 5))

	var filter *regexp.Regexp
	if filterPattern != "" {
		filter, err = regexp.Compile(filterPattern)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("无效的过滤器正则表达式: %v", err)), nil
		}
	}

	scanResult := map[string]interface{}{
		"scan_path":   path,
		"recursive":   recursive,
		"filter":      filterPattern,
		"max_depth":   maxDepth,
		"scan_time":   time.Now().Format(time.RFC3339),
		"files":       []map[string]interface{}{},
		"directories": []map[string]interface{}{},
		"summary":     map[string]interface{}{},
	}

	var files []map[string]interface{}
	var directories []map[string]interface{}
	var totalSize int64
	var fileCount, dirCount int

	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续扫描
		}

		// 检查深度限制
		relPath, _ := filepath.Rel(path, filePath)
		depth := strings.Count(relPath, string(filepath.Separator))
		if depth > maxDepth {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 应用过滤器
		if filter != nil && !filter.MatchString(info.Name()) {
			return nil
		}

		fileData := map[string]interface{}{
			"name":     info.Name(),
			"path":     filePath,
			"size":     info.Size(),
			"modified": info.ModTime().Format(time.RFC3339),
			"mode":     info.Mode().String(),
		}

		if info.IsDir() {
			directories = append(directories, fileData)
			dirCount++
		} else {
			files = append(files, fileData)
			fileCount++
			totalSize += info.Size()
		}

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("目录扫描失败: %v", err)), nil
	}

	scanResult["files"] = files
	scanResult["directories"] = directories
	scanResult["summary"] = map[string]interface{}{
		"total_files":       fileCount,
		"total_directories": dirCount,
		"total_size":        totalSize,
		"total_size_human":  formatBytes(totalSize),
	}

	jsonData, _ := json.MarshalIndent(scanResult, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("📂 目录扫描结果:\n%s", string(jsonData))), nil
}

// 网络监控处理函数
func handleNetworkMonitor(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	target, err := request.RequireString("target")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	action, err := request.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	ports := request.GetString("ports", "")

	result := map[string]interface{}{
		"target":    target,
		"action":    action,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	switch action {
	case "ping":
		// 简单的连接测试
		conn, err := net.DialTimeout("tcp", target+":80", 5*time.Second)
		if err != nil {
			result["status"] = "unreachable"
			result["error"] = err.Error()
		} else {
			conn.Close()
			result["status"] = "reachable"
			result["response_time"] = "< 5s"
		}

	case "port_scan":
		if ports == "" {
			return mcp.NewToolResultError("端口扫描需要指定端口"), nil
		}

		portResults := []map[string]interface{}{}
		portList := parsePortList(ports)

		for _, port := range portList {
			portResult := map[string]interface{}{
				"port": port,
			}

			conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", target, port), 3*time.Second)
			if err != nil {
				portResult["status"] = "closed"
			} else {
				conn.Close()
				portResult["status"] = "open"
			}

			portResults = append(portResults, portResult)
		}

		result["ports"] = portResults

	case "trace_route":
		// 简化的路由跟踪（实际实现会更复杂）
		result["trace"] = []map[string]interface{}{
			{"hop": 1, "address": "gateway", "time": "1ms"},
			{"hop": 2, "address": target, "time": "10ms"},
		}

	default:
		return mcp.NewToolResultError("不支持的网络监控动作"), nil
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("🌐 网络监控结果:\n%s", string(jsonData))), nil
}

// 数据处理处理函数
func handleDataProcess(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data, err := request.RequireString("data")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	format, err := request.RequireString("format")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	operation, err := request.RequireString("operation")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	parameters := request.GetString("parameters", "{}")

	// 解析参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(parameters), &params); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("参数解析失败: %v", err)), nil
	}

	result := map[string]interface{}{
		"format":    format,
		"operation": operation,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// 根据格式解析数据
	var parsedData interface{}
	switch format {
	case "json":
		if err := json.Unmarshal([]byte(data), &parsedData); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("JSON解析失败: %v", err)), nil
		}
	case "csv":
		// 简化的CSV解析
		lines := strings.Split(data, "\n")
		var csvData []map[string]string
		if len(lines) > 0 {
			headers := strings.Split(lines[0], ",")
			for i := 1; i < len(lines); i++ {
				if strings.TrimSpace(lines[i]) == "" {
					continue
				}
				values := strings.Split(lines[i], ",")
				row := make(map[string]string)
				for j, header := range headers {
					if j < len(values) {
						row[strings.TrimSpace(header)] = strings.TrimSpace(values[j])
					}
				}
				csvData = append(csvData, row)
			}
		}
		parsedData = csvData
	default:
		return mcp.NewToolResultError("不支持的数据格式"), nil
	}

	// 执行操作
	switch operation {
	case "analyze":
		analysis := analyzeData(parsedData)
		result["analysis"] = analysis
	case "filter":
		// 实现数据过滤逻辑
		result["filtered_data"] = parsedData // 简化实现
	case "transform":
		// 实现数据转换逻辑
		result["transformed_data"] = parsedData // 简化实现
	case "aggregate":
		// 实现数据聚合逻辑
		result["aggregated_data"] = parsedData // 简化实现
	default:
		return mcp.NewToolResultError("不支持的数据操作"), nil
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("📊 数据处理结果:\n%s", string(jsonData))), nil
}

// 注册资源
func registerResources(s *server.MCPServer) {
	// 静态资源：服务器状态
	statusResource := mcp.NewResource(
		"server://status",
		"服务器状态",
		mcp.WithResourceDescription("当前服务器状态信息"),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(statusResource, handleServerStatus)

	// 动态资源：文件内容
	fileTemplate := mcp.NewResourceTemplate(
		"file://{path}",
		"文件内容",
		mcp.WithTemplateDescription("读取指定路径的文件内容"),
		mcp.WithTemplateMIMEType("text/plain"),
	)
	s.AddResourceTemplate(fileTemplate, handleFileResource)

	// 动态资源：配置
	configTemplate := mcp.NewResourceTemplate(
		"config://{key}",
		"配置项",
		mcp.WithTemplateDescription("获取指定配置项的值"),
		mcp.WithTemplateMIMEType("application/json"),
	)
	s.AddResourceTemplate(configTemplate, handleConfigResource)
}

// 注册提示模板
func registerPrompts(s *server.MCPServer) {
	// 代码审查提示
	codeReviewPrompt := mcp.NewPrompt("code_review",
		mcp.WithPromptDescription("代码审查助手"),
		mcp.WithArgument("language",
			mcp.ArgumentDescription("编程语言"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("focus",
			mcp.ArgumentDescription("审查重点：security, performance, style"),
		),
	)
	s.AddPrompt(codeReviewPrompt, handleCodeReviewPrompt)

	// 数据分析提示
	dataAnalysisPrompt := mcp.NewPrompt("data_analysis",
		mcp.WithPromptDescription("数据分析助手"),
		mcp.WithArgument("data_type",
			mcp.ArgumentDescription("数据类型"),
			mcp.RequiredArgument(),
		),
	)
	s.AddPrompt(dataAnalysisPrompt, handleDataAnalysisPrompt)
}

// 工具处理函数

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

func handleReadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取文件失败: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("文件内容 (%s):\n%s", path, string(content))), nil
}

func handleWriteFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, err := request.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("创建目录失败: %v", err)), nil
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("写入文件失败: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("成功写入文件: %s", path)), nil
}

func handleHTTPRequest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url, err := request.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	method := request.GetString("method", "GET")
	body := request.GetString("body", "")

	var req *http.Request
	if body != "" {
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("创建请求失败: %v", err)), nil
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("请求失败: %v", err)), nil
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("读取响应失败: %v", err)), nil
	}

	result := fmt.Sprintf("HTTP %s %s\n状态码: %d\n响应体:\n%s", method, url, resp.StatusCode, string(respBody))
	return mcp.NewToolResultText(result), nil
}

func handleSystemInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	info := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"pid":       os.Getpid(),
		"hostname":  getHostname(),
		"workdir":   getWorkdir(),
	}

	jsonData, _ := json.MarshalIndent(info, "", "  ")
	return mcp.NewToolResultText(fmt.Sprintf("系统信息:\n%s", string(jsonData))), nil
}

func handleCurrentTime(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := request.GetString("format", "2006-01-02 15:04:05")
	timezone := request.GetString("timezone", "Local")

	var loc *time.Location
	var err error
	if timezone == "Local" {
		loc = time.Local
	} else {
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("无效时区: %v", err)), nil
		}
	}

	now := time.Now().In(loc)
	formatted := now.Format(format)

	return mcp.NewToolResultText(fmt.Sprintf("当前时间 (%s): %s", timezone, formatted)), nil
}

// 资源处理函数

func handleServerStatus(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	status := map[string]interface{}{
		"status":    "running",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
		"version":   "1.0.0",
	}

	jsonData, _ := json.MarshalIndent(status, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      request.Params.URI,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

func handleFileResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// 从URI中提取文件路径: file://{path}
	uri := request.Params.URI
	if !strings.HasPrefix(uri, "file://") {
		return nil, fmt.Errorf("无效的文件URI: %s", uri)
	}

	path := strings.TrimPrefix(uri, "file://")
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "text/plain",
			Text:     string(content),
		},
	}, nil
}

func handleConfigResource(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	// 从URI中提取配置键: config://{key}
	uri := request.Params.URI
	if !strings.HasPrefix(uri, "config://") {
		return nil, fmt.Errorf("无效的配置URI: %s", uri)
	}

	key := strings.TrimPrefix(uri, "config://")

	// 模拟配置数据
	configs := map[string]interface{}{
		"debug":     true,
		"max_users": 100,
		"timeout":   30,
		"log_level": "info",
	}

	value, exists := configs[key]
	if !exists {
		return nil, fmt.Errorf("配置项不存在: %s", key)
	}

	result := map[string]interface{}{
		"key":   key,
		"value": value,
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      uri,
			MIMEType: "application/json",
			Text:     string(jsonData),
		},
	}, nil
}

// 提示模板处理函数

func handleCodeReviewPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	language := request.Params.Arguments["language"]
	focus := request.Params.Arguments["focus"]
	if focus == "" {
		focus = "general"
	}

	prompt := fmt.Sprintf(`你是一个专业的代码审查专家。请审查以下%s代码，重点关注%s方面：

审查标准：
- 代码质量和可读性
- 潜在的bug和错误
- 性能优化建议
- 安全性问题
- 最佳实践建议

请提供具体的改进建议和代码示例。`, language, focus)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("%s代码审查 (重点: %s)", language, focus),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}

func handleDataAnalysisPrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	dataType := request.Params.Arguments["data_type"]

	prompt := fmt.Sprintf(`你是一个数据分析专家。请帮助分析%s类型的数据：

分析内容应包括：
1. 数据概览和基本统计
2. 数据质量评估
3. 趋势和模式识别
4. 异常值检测
5. 数据可视化建议
6. 关键洞察和建议

请提供详细的分析步骤和解释。`, dataType)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("%s数据分析", dataType),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: prompt,
				},
			},
		},
	}, nil
}

// 辅助函数

var startTime = time.Now()

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func getWorkdir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
