package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// MCP 客户端示例
// 用法: go run examples/client_demo.go
func main() {
	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 创建 stdio 传输，连接到服务器
	stdioTransport := transport.NewStdio("go", nil, "run", "../main.go")

	// 创建客户端
	c := client.NewClient(stdioTransport)

	// 启动客户端
	if err := c.Start(ctx); err != nil {
		log.Fatalf("启动客户端失败: %v", err)
	}
	defer c.Close()

	// 初始化客户端
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "Advance Go MCP client",
		Version: "2.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	fmt.Println("正在初始化 MCP 客户端...")
	serverInfo, err := c.Initialize(ctx, initRequest)
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	fmt.Printf("已连接到服务器: %s (版本 %s)\n", serverInfo.ServerInfo.Name, serverInfo.ServerInfo.Version)
	fmt.Printf("服务器能力: %+v\n", serverInfo.Capabilities)

	// 演示工具功能
	demonstratCalculateTools(ctx, c, serverInfo)

	demonstrateEnhancedDatabaseQueries(ctx, c, serverInfo)

	demonstrateWebSearch(ctx, c, serverInfo)

	fmt.Println("客户端演示完成")
}

func demonstratCalculateTools(ctx context.Context, c *client.Client, serverInfo *mcp.InitializeResult) {
	if serverInfo.Capabilities.Tools == nil {
		fmt.Println("服务器不支持工具")
		return
	}

	fmt.Println("\n=== 工具演示 ===")

	// 列出可用工具
	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := c.ListTools(ctx, toolsRequest)
	if err != nil {
		log.Printf("获取工具列表失败: %v", err)
		return
	}

	fmt.Printf("服务器提供 %d 个工具:\n", len(toolsResult.Tools))
	for i, tool := range toolsResult.Tools {
		fmt.Printf("  %d. %s - %s\n", i+1, tool.Name, tool.Description)
	}

	// 演示计算器工具
	fmt.Println("\n--- 使用计算器工具 ---")
	calcRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "calculator",
			Arguments: map[string]interface{}{
				"operation": "multiply",
				"x":         199999349349,
				"y":         4384535757535,
			},
		},
	}

	calcResult, err := c.CallTool(ctx, calcRequest)
	if err != nil {
		log.Printf("调用计算器工具失败: %v", err)
	} else {
		fmt.Printf("计算器结果: %v\n", calcResult.Content)
	}

}

// 演示增强的数据库查询功能
func demonstrateEnhancedDatabaseQueries(ctx context.Context, c *client.Client, serverInfo *mcp.InitializeResult) {
	if serverInfo.Capabilities.Tools == nil {
		fmt.Println("服务器不支持工具")
		return
	}

	fmt.Println("\n=== 增强数据库查询工具演示 ===")

	// 1. 原始SQL查询
	fmt.Println("\n--- 1. 原始SQL查询 ---")
	rawQueryRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type": "raw",
				"query":      "SELECT id, name, email FROM users WHERE status = 'active' LIMIT 3",
				"database":   "default",
			},
		},
	}

	result, err := c.CallTool(ctx, rawQueryRequest)
	if err != nil {
		log.Printf("原始SQL查询失败: %v", err)
	} else {
		fmt.Printf("查询结果:\n%v\n", result.Content)
	}

	// 2. 结构化SELECT查询 - 基本查询
	fmt.Println("\n--- 2. 结构化SELECT查询 - 基本查询 ---")
	structuredSelectRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type":  "structured",
				"query":       "select",
				"table_name":  "users",
				"fields":      "id, name, email, status",
				"limit":       5,
			},
		},
	}

	result, err = c.CallTool(ctx, structuredSelectRequest)
	if err != nil {
		log.Printf("结构化SELECT查询失败: %v", err)
	} else {
		fmt.Printf("查询结果:\n%v\n", result.Content)
	}

	// 3. 结构化SELECT查询 - 带条件和排序
	fmt.Println("\n--- 3. 结构化SELECT查询 - 带条件和排序 ---")
	conditionalSelectRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type":       "structured",
				"query":            "select",
				"table_name":       "users",
				"fields":           "name, email, created_at",
				"where_conditions": "status=active",
				"order_by":         "created_at DESC",
				"limit":            3,
			},
		},
	}

	result, err = c.CallTool(ctx, conditionalSelectRequest)
	if err != nil {
		log.Printf("条件查询失败: %v", err)
	} else {
		fmt.Printf("查询结果:\n%v\n", result.Content)
	}

	// 4. 结构化COUNT查询
	fmt.Println("\n--- 4. 结构化COUNT查询 ---")
	countRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type":       "structured",
				"query":            "count",
				"table_name":       "users",
				"where_conditions": "status=active",
			},
		},
	}

	result, err = c.CallTool(ctx, countRequest)
	if err != nil {
		log.Printf("COUNT查询失败: %v", err)
	} else {
		fmt.Printf("统计结果:\n%v\n", result.Content)
	}

	// 5. 结构化COUNT查询 - 分组统计
	fmt.Println("\n--- 5. 结构化COUNT查询 - 分组统计 ---")
	groupCountRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type": "structured",
				"query":      "count",
				"table_name": "users",
				"group_by":   "status",
			},
		},
	}

	result, err = c.CallTool(ctx, groupCountRequest)
	if err != nil {
		log.Printf("分组统计查询失败: %v", err)
	} else {
		fmt.Printf("分组统计结果:\n%v\n", result.Content)
	}

	// 6. JSON格式条件查询
	fmt.Println("\n--- 6. JSON格式条件查询 ---")
	jsonConditionRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type":       "structured",
				"query":            "select",
				"table_name":       "users",
				"fields":           "*",
				"where_conditions": `{"status":"active"}`,
				"limit":            2,
			},
		},
	}

	result, err = c.CallTool(ctx, jsonConditionRequest)
	if err != nil {
		log.Printf("JSON条件查询失败: %v", err)
	} else {
		fmt.Printf("查询结果:\n%v\n", result.Content)
	}

	// 7. 复杂条件查询
	fmt.Println("\n--- 7. 复杂条件查询 ---")
	complexConditionRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type":       "structured",
				"query":            "select",
				"table_name":       "users",
				"fields":           "id, name, email",
				"where_conditions": "id>1,status=active",
				"order_by":         "id ASC",
				"limit":            3,
			},
		},
	}

	result, err = c.CallTool(ctx, complexConditionRequest)
	if err != nil {
		log.Printf("复杂条件查询失败: %v", err)
	} else {
		fmt.Printf("查询结果:\n%v\n", result.Content)
	}

	// 8. 演示INSERT操作（注意：这会实际插入数据）
	fmt.Println("\n--- 8. 结构化INSERT操作 ---")
	insertRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type": "structured",
				"query":      "insert",
				"table_name": "users",
				"fields":     `{"name":"测试用户","email":"test@example.com","status":"active"}`,
			},
		},
	}

	result, err = c.CallTool(ctx, insertRequest)
	if err != nil {
		log.Printf("INSERT操作失败: %v", err)
	} else {
		fmt.Printf("插入结果:\n%v\n", result.Content)
	}

	// 9. 演示UPDATE操作
	fmt.Println("\n--- 9. 结构化UPDATE操作 ---")
	updateRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type":       "structured",
				"query":            "update",
				"table_name":       "users",
				"fields":           `{"status":"updated"}`,
				"where_conditions": "email=test@example.com",
			},
		},
	}

	result, err = c.CallTool(ctx, updateRequest)
	if err != nil {
		log.Printf("UPDATE操作失败: %v", err)
	} else {
		fmt.Printf("更新结果:\n%v\n", result.Content)
	}

	// 10. 验证更新结果
	fmt.Println("\n--- 10. 验证更新结果 ---")
	verifyRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "database_query",
			Arguments: map[string]interface{}{
				"query_type":       "structured",
				"query":            "select",
				"table_name":       "users",
				"where_conditions": "email=test@example.com",
			},
		},
	}

	result, err = c.CallTool(ctx, verifyRequest)
	if err != nil {
		log.Printf("验证查询失败: %v", err)
	} else {
		fmt.Printf("验证结果:\n%v\n", result.Content)
	}
}


// 演示网络搜索工具
func demonstrateWebSearch(ctx context.Context, c *client.Client, serverInfo *mcp.InitializeResult) {
	if serverInfo.Capabilities.Tools == nil {
		fmt.Println("服务器不支持工具")
		return
	}

	fmt.Println("\n=== 网络搜索工具演示 ===")

	// 1. 基本搜索 - 使用默认结果数量
	fmt.Println("\n--- 基本搜索：Go语言 ---")
	basicSearchRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query": "Go programming language tutorial",
			},
		},
	}

	result, err := c.CallTool(ctx, basicSearchRequest)
	if err != nil {
		log.Printf("执行基本搜索失败: %v", err)
	} else {
		fmt.Printf("搜索结果:\n%v\n", result.Content)
	}

	// 2. 限制结果数量的搜索
	fmt.Println("\n--- 限制结果搜索：Python机器学习（限制5条结果）---")
	limitedSearchRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query": "Python machine learning frameworks",
				"limit": 5,
			},
		},
	}

	result, err = c.CallTool(ctx, limitedSearchRequest)
	if err != nil {
		log.Printf("执行限制结果搜索失败: %v", err)
	} else {
		fmt.Printf("搜索结果:\n%v\n", result.Content)
	}

	// 3. 技术相关搜索
	fmt.Println("\n--- 技术搜索：Docker容器化 ---")
	techSearchRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query": "Docker containerization best practices",
				"limit": 8,
			},
		},
	}

	result, err = c.CallTool(ctx, techSearchRequest)
	if err != nil {
		log.Printf("执行技术搜索失败: %v", err)
	} else {
		fmt.Printf("搜索结果:\n%v\n", result.Content)
	}

	// 4. 中文搜索示例
	fmt.Println("\n--- 中文搜索：人工智能发展 ---")
	chineseSearchRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "web_search",
			Arguments: map[string]interface{}{
				"query": "人工智能发展趋势 2024",
				"limit": 6,
			},
		},
	}

	result, err = c.CallTool(ctx, chineseSearchRequest)
	if err != nil {
		log.Printf("执行中文搜索失败: %v", err)
	} else {
		fmt.Printf("搜索结果:\n%v\n", result.Content)
	}
}
