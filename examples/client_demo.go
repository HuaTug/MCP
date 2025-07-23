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
		Name:    "MCP Go 客户端示例",
		Version: "1.0.0",
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
	demonstrateTools(ctx, c, serverInfo)

	// 演示资源功能
	demonstrateResources(ctx, c, serverInfo)

	// 演示提示模板功能
	demonstratePrompts(ctx, c, serverInfo)

	fmt.Println("客户端演示完成")
}

func demonstrateTools(ctx context.Context, c *client.Client, serverInfo *mcp.InitializeResult) {
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
				"operation": "add",
				"x":         10,
				"y":         5,
			},
		},
	}

	calcResult, err := c.CallTool(ctx, calcRequest)
	if err != nil {
		log.Printf("调用计算器工具失败: %v", err)
	} else {
		fmt.Printf("计算器结果: %v\n", calcResult.Content)
	}

	// 演示系统信息工具
	fmt.Println("\n--- 获取系统信息 ---")
	sysRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      "system_info",
			Arguments: map[string]interface{}{},
		},
	}

	sysResult, err := c.CallTool(ctx, sysRequest)
	if err != nil {
		log.Printf("获取系统信息失败: %v", err)
	} else {
		fmt.Printf("系统信息: %v\n", sysResult.Content)
	}

	// 演示时间工具
	fmt.Println("\n--- 获取当前时间 ---")
	timeRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "current_time",
			Arguments: map[string]interface{}{
				"format":   "2006-01-02 15:04:05",
				"timezone": "Asia/Shanghai",
			},
		},
	}

	timeResult, err := c.CallTool(ctx, timeRequest)
	if err != nil {
		log.Printf("获取时间失败: %v", err)
	} else {
		fmt.Printf("当前时间: %v\n", timeResult.Content)
	}
}

func demonstrateResources(ctx context.Context, c *client.Client, serverInfo *mcp.InitializeResult) {
	if serverInfo.Capabilities.Resources == nil {
		fmt.Println("服务器不支持资源")
		return
	}

	fmt.Println("\n=== 资源演示 ===")

	// 列出可用资源
	resourcesRequest := mcp.ListResourcesRequest{}
	resourcesResult, err := c.ListResources(ctx, resourcesRequest)
	if err != nil {
		log.Printf("获取资源列表失败: %v", err)
		return
	}

	fmt.Printf("服务器提供 %d 个资源:\n", len(resourcesResult.Resources))
	for i, resource := range resourcesResult.Resources {
		fmt.Printf("  %d. %s - %s\n", i+1, resource.URI, resource.Name)
	}

	// 读取服务器状态资源
	fmt.Println("\n--- 读取服务器状态 ---")
	statusRequest := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "server://status",
		},
	}

	statusResult, err := c.ReadResource(ctx, statusRequest)
	if err != nil {
		log.Printf("读取服务器状态失败: %v", err)
	} else {
		fmt.Printf("服务器状态: %v\n", statusResult.Contents)
	}

	// 读取配置资源
	fmt.Println("\n--- 读取配置信息 ---")
	configRequest := mcp.ReadResourceRequest{
		Params: mcp.ReadResourceParams{
			URI: "config://debug",
		},
	}

	configResult, err := c.ReadResource(ctx, configRequest)
	if err != nil {
		log.Printf("读取配置失败: %v", err)
	} else {
		fmt.Printf("配置信息: %v\n", configResult.Contents)
	}
}

func demonstratePrompts(ctx context.Context, c *client.Client, serverInfo *mcp.InitializeResult) {
	if serverInfo.Capabilities.Prompts == nil {
		fmt.Println("服务器不支持提示模板")
		return
	}

	fmt.Println("\n=== 提示模板演示 ===")

	// 列出可用提示模板
	promptsRequest := mcp.ListPromptsRequest{}
	promptsResult, err := c.ListPrompts(ctx, promptsRequest)
	if err != nil {
		log.Printf("获取提示模板列表失败: %v", err)
		return
	}

	fmt.Printf("服务器提供 %d 个提示模板:\n", len(promptsResult.Prompts))
	for i, prompt := range promptsResult.Prompts {
		fmt.Printf("  %d. %s - %s\n", i+1, prompt.Name, prompt.Description)
	}

	// 获取代码审查提示
	fmt.Println("\n--- 获取代码审查提示 ---")
	codeReviewRequest := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "code_review",
			Arguments: map[string]string{
				"language": "Go",
				"focus":    "security",
			},
		},
	}

	codeReviewResult, err := c.GetPrompt(ctx, codeReviewRequest)
	if err != nil {
		log.Printf("获取代码审查提示失败: %v", err)
	} else {
		fmt.Printf("代码审查提示: %s\n", codeReviewResult.Description)
		for i, msg := range codeReviewResult.Messages {
			fmt.Printf("  消息 %d (%s): %v\n", i+1, msg.Role, msg.Content)
		}
	}

	// 获取数据分析提示
	fmt.Println("\n--- 获取数据分析提示 ---")
	dataAnalysisRequest := mcp.GetPromptRequest{
		Params: mcp.GetPromptParams{
			Name: "data_analysis",
			Arguments: map[string]string{
				"data_type": "销售数据",
			},
		},
	}

	dataAnalysisResult, err := c.GetPrompt(ctx, dataAnalysisRequest)
	if err != nil {
		log.Printf("获取数据分析提示失败: %v", err)
	} else {
		fmt.Printf("数据分析提示: %s\n", dataAnalysisResult.Description)
		for i, msg := range dataAnalysisResult.Messages {
			fmt.Printf("  消息 %d (%s): %v\n", i+1, msg.Role, msg.Content)
		}
	}
}
