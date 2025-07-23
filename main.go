package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 创建MCP服务器
	mcpServer := server.NewMCPServer(
		"Go Demo MCP Server",
		"1.0.0",
		server.WithResourceCapabilities(true, true), // 支持静态和动态资源
		server.WithPromptCapabilities(true),         // 支持提示模板
		server.WithToolCapabilities(true),           // 支持工具
		server.WithRecovery(),                       // 错误恢复
		server.WithLogging(),                        // 启用日志
	)

	// 注册工具
	registerTools(mcpServer)

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
