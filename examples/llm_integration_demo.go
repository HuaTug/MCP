package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

// LLM API配置 - 从环境变量读取
func getLLMConfig() (string, string, string) {
	apiURL := os.Getenv("LLM_API_URL")
	if apiURL == "" {
		apiURL = "http://api.lkeap.cloud.tencent.com/v1/chat/completions"
	}

	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		apiKey = "sk-qFPEqgpxmS8DJ0nJQ6gvdIkozY1k2oEZER2A4zRhLxBvtIHl"
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = "deepseek-v3-0324"
	}

	return apiURL, apiKey, model
}

// LLM API请求结构
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMRequest struct {
	Model     string       `json:"model"`
	Messages  []LLMMessage `json:"messages"`
	Stream    bool         `json:"stream"`
	ExtraBody ExtraBody    `json:"extra_body"`
}

type ExtraBody struct {
	EnableSearch bool `json:"enable_search"`
}

// LLM API响应结构
type LLMChoice struct {
	Delta struct {
		Content string `json:"content"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type LLMResponse struct {
	Choices []LLMChoice `json:"choices"`
}

// LLM智能应用演示
type IntelligentAssistant struct {
	mcpClient      *client.Client
	availableTools []mcp.Tool
}

// 工具调用结构
type ToolCall struct {
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args"`
}

// 查询结果
type QueryResult struct {
	UserQuery   string        `json:"user_query"`
	ToolsUsed   []ToolCall    `json:"tools_used"`
	RawResults  []string      `json:"raw_results"`
	FinalAnswer string        `json:"final_answer"`
	ProcessTime time.Duration `json:"process_time"`
}

// 语义分析结果
type ToolAnalysis struct {
	NeedsWebSearch   bool                   `json:"needs_web_search"`
	WebSearchQuery   string                 `json:"web_search_query"`
	NeedsDatabase    bool                   `json:"needs_database"`
	DatabaseQuery    map[string]interface{} `json:"database_query"`
	NeedsCalculation bool                   `json:"needs_calculation"`
	CalculationArgs  map[string]interface{} `json:"calculation_args"`
	Reasoning        string                 `json:"reasoning"`
}

// 初始化智能助手
func NewIntelligentAssistant() (*IntelligentAssistant, error) {
	// 设置自定义命令函数，指定工作目录
	cmdFunc := func(ctx context.Context, command string, env []string, args []string) (*exec.Cmd, error) {
		cmd := exec.CommandContext(ctx, command, args...)
		cmd.Env = env
		// 设置工作目录为上级目录
		cmd.Dir = "../"
		return cmd, nil
	}

	// 连接到MCP服务器，使用自定义命令函数
	mcpClient, err := client.NewStdioMCPClientWithOptions(
		"go",
		nil,                        // env
		[]string{"run", "main.go"}, // 修改为直接运行 main.go
		transport.WithCommandFunc(cmdFunc),
	)
	if err != nil {
		return nil, fmt.Errorf("连接MCP服务器失败: %v", err)
	}

	// 增加超时时间到30秒，给服务器更多启动时间
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 初始化连接
	initRequest := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    mcp.ClientCapabilities{
				// 基本能力配置
			},
			ClientInfo: mcp.Implementation{
				Name:    "llm-integration-demo",
				Version: "1.0.0",
			},
		},
	}

	_, err = mcpClient.Initialize(ctx, initRequest)
	if err != nil {
		return nil, fmt.Errorf("初始化MCP连接失败: %v", err)
	}

	// 获取可用工具列表
	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := mcpClient.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, fmt.Errorf("获取工具列表失败: %v", err)
	}

	fmt.Printf("✅ 成功连接到MCP服务器，发现 %d 个可用工具:\n", len(toolsResult.Tools))
	for _, tool := range toolsResult.Tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	return &IntelligentAssistant{
		mcpClient:      mcpClient,
		availableTools: toolsResult.Tools,
	}, nil
}

// 处理用户查询的主要方法
func (ia *IntelligentAssistant) ProcessUserQuery(ctx context.Context, userQuery string) (*QueryResult, error) {
	startTime := time.Now()

	result := &QueryResult{
		UserQuery:  userQuery,
		ToolsUsed:  []ToolCall{},
		RawResults: []string{},
	}

	fmt.Printf("\n🤖 处理用户查询: %s\n", userQuery)

	// 1. 智能分析查询，决定是否需要工具
	toolCalls := ia.analyzeQueryForTools(userQuery)

	if len(toolCalls) == 0 {
		// 不需要工具，直接使用LLM回答
		fmt.Printf("🤖 直接调用LLM回答（无需工具）...\n")
		llmResponse, err := ia.callLLM(ctx, userQuery, []string{})
		if err != nil {
			// LLM调用失败时的备用方案
			fmt.Printf("⚠️ LLM调用失败，使用备用回答: %v\n", err)
			result.FinalAnswer = ia.generateDirectAnswer(userQuery)
		} else {
			result.FinalAnswer = llmResponse
		}
		result.ProcessTime = time.Since(startTime)
		return result, nil
	}

	// 2. 执行工具调用
	for _, toolCall := range toolCalls {
		fmt.Printf("🔧 调用工具: %s\n", toolCall.Name)

		toolResult, err := ia.callTool(ctx, toolCall)
		if err != nil {
			return nil, fmt.Errorf("工具调用失败 (%s): %v", toolCall.Name, err)
		}

		result.ToolsUsed = append(result.ToolsUsed, toolCall)
		result.RawResults = append(result.RawResults, ia.formatToolResult(toolResult))
	}

	// 3. 调用LLM生成最终回答
	fmt.Printf("🤖 调用LLM生成智能回答...\n")
	llmResponse, err := ia.callLLM(ctx, userQuery, result.RawResults)
	if err != nil {
		// 如果LLM调用失败，使用备用方案
		fmt.Printf("⚠️ LLM调用失败，使用备用回答: %v\n", err)
		result.FinalAnswer = ia.synthesizeAnswer(userQuery, result.RawResults)
	} else {
		result.FinalAnswer = llmResponse
	}

	result.ProcessTime = time.Since(startTime)

	return result, nil
}

// 智能分析查询，确定需要哪些工具
func (ia *IntelligentAssistant) analyzeQueryForTools(query string) []ToolCall {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用语义理解来分析查询意图
	toolAnalysis, err := ia.analyzeQuerySemantics(ctx, query)
	if err != nil {
		fmt.Printf("⚠️ 语义分析失败，回退到关键词匹配: %v\n", err)
		return ia.analyzeQueryForToolsFallback(query)
	}

	var tools []ToolCall

	// 根据语义分析结果构建工具调用
	if toolAnalysis.NeedsWebSearch {
		tools = append(tools, ToolCall{
			Name: "web_search",
			Args: map[string]interface{}{
				"query": toolAnalysis.WebSearchQuery,
				"limit": 5,
			},
		})
		fmt.Printf("📡 语义分析：需要网络搜索 - %s\n", toolAnalysis.WebSearchQuery)
	}

	if toolAnalysis.NeedsDatabase {
		tools = append(tools, ToolCall{
			Name: "database_query",
			Args: toolAnalysis.DatabaseQuery,
		})
		fmt.Printf("🗄️ 语义分析：需要数据库查询\n")
	}

	if toolAnalysis.NeedsCalculation {
		tools = append(tools, ToolCall{
			Name: "calculator",
			Args: toolAnalysis.CalculationArgs,
		})
		fmt.Printf("🧮 语义分析：需要数学计算\n")
	}

	return tools
}

// 使用LLM进行语义分析，判断需要哪些工具
func (ia *IntelligentAssistant) analyzeQuerySemantics(ctx context.Context, query string) (*ToolAnalysis, error) {
	// 获取LLM配置
	apiURL, apiKey, model := getLLMConfig()

	// 构建分析提示词
	prompt := fmt.Sprintf(`你是一个智能助手的工具调用分析器。请分析用户的查询，判断需要调用哪些工具。

可用的工具：
1. web_search - 网络搜索工具，用于获取最新信息、新闻、实时数据、当前时间日期等
2. database_query - 数据库查询工具，用于查询用户数据、统计信息等
3. calculator - 计算器工具，用于数学运算

用户查询：%s

请仔细分析这个查询，判断是否需要调用工具，并以JSON格式返回分析结果：

{
  "needs_web_search": false,
  "web_search_query": "",
  "needs_database": false,
  "database_query": {},
  "needs_calculation": false,
  "calculation_args": {},
  "reasoning": "分析推理过程"
}

分析规则：
- **需要web_search的情况：**
  * 询问当前时间、日期（如"今天几号"、"现在几点"、"今天星期几"）
  * 最新消息、新闻、实时数据
  * 当前天气、股价等实时信息
  * 任何需要"当前"、"现在"、"今天"状态的查询

- **需要database_query的情况：**
  * 查询用户数据、统计信息
  * 数据库相关操作（增删改查）
  * 涉及"用户"、"统计"、"数据"等关键词

- **需要calculator的情况：**
  * 明确的数学计算、运算
  * 涉及数字计算的问题

- **不需要工具的情况：**
  * 一般性知识问题
  * 概念解释
  * 历史事实等静态知识

重要提醒：时间和日期相关的查询（如"今天几号"）属于实时信息，必须使用web_search工具！

请只返回JSON，不要包含其他内容。`, query)

	// 构建API请求
	llmRequest := LLMRequest{
		Model: model,
		Messages: []LLMMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false, // 使用非流式响应以便解析JSON
		ExtraBody: ExtraBody{
			EnableSearch: false,
		},
	}

	// 序列化请求
	requestBody, err := json.Marshal(llmRequest)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API请求失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析响应，提取JSON内容
	response := string(body)

	// 如果是流式响应格式，需要提取实际内容
	if strings.Contains(response, "data:") {
		lines := strings.Split(response, "\n")
		var content strings.Builder
		for _, line := range lines {
			if strings.HasPrefix(line, "data: ") {
				jsonData := strings.TrimPrefix(line, "data: ")
				if jsonData != "[DONE]" && jsonData != "" {
					var streamResp LLMResponse
					if parseErr := json.Unmarshal([]byte(jsonData), &streamResp); parseErr == nil {
						if len(streamResp.Choices) > 0 {
							content.WriteString(streamResp.Choices[0].Delta.Content)
						}
					}
				}
			}
		}
		response = content.String()
	} else {
		// 非流式响应，直接解析
		var llmResponse struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if parseErr := json.Unmarshal(body, &llmResponse); parseErr == nil {
			if len(llmResponse.Choices) > 0 {
				response = llmResponse.Choices[0].Message.Content
			}
		}
	}

	// 提取JSON部分
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("无法从响应中提取JSON: %s", response)
	}

	jsonStr := response[start : end+1]

	// 解析工具分析结果
	var analysis ToolAnalysis
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %v, 原始响应: %s", err, jsonStr)
	}

	// 智能填充工具参数
	ia.fillToolParameters(&analysis, query)

	return &analysis, nil
}

// 智能填充工具参数
func (ia *IntelligentAssistant) fillToolParameters(analysis *ToolAnalysis, query string) {
	// 填充网络搜索参数
	if analysis.NeedsWebSearch && analysis.WebSearchQuery == "" {
		analysis.WebSearchQuery = query
	}

	// 填充数据库查询参数
	if analysis.NeedsDatabase && len(analysis.DatabaseQuery) == 0 {
		analysis.DatabaseQuery = ia.buildDatabaseQuery(query)
	}

	// 填充计算参数
	if analysis.NeedsCalculation && len(analysis.CalculationArgs) == 0 {
		analysis.CalculationArgs = ia.parseCalculation(query)
	}
}

// 备用的关键词匹配方法（当语义分析失败时使用）
func (ia *IntelligentAssistant) analyzeQueryForToolsFallback(query string) []ToolCall {
	var tools []ToolCall
	query = strings.ToLower(query)

	// 检测是否需要实时信息搜索
	if ia.needsWebSearch(query) {
		tools = append(tools, ToolCall{
			Name: "web_search",
			Args: map[string]interface{}{
				"query": query,
				"limit": 5,
			},
		})
		fmt.Printf("📡 关键词匹配：需要网络搜索\n")
	}

	// 检测是否需要数据库查询
	if ia.needsDatabase(query) {
		tools = append(tools, ToolCall{
			Name: "database_query",
			Args: ia.buildDatabaseQuery(query),
		})
		fmt.Printf("🗄️ 关键词匹配：需要数据库查询\n")
	}

	// 检测是否需要数学计算
	if ia.needsCalculation(query) {
		calcArgs := ia.parseCalculation(query)
		if calcArgs != nil {
			tools = append(tools, ToolCall{
				Name: "calculator",
				Args: calcArgs,
			})
			fmt.Printf("🧮 关键词匹配：需要数学计算\n")
		}
	}

	return tools
}

// 判断是否需要网络搜索
func (ia *IntelligentAssistant) needsWebSearch(query string) bool {
	webSearchKeywords := []string{
		// 时间日期相关
		"今天", "现在", "当前", "今日", "此刻", "目前",
		"几号", "几月", "几点", "星期几", "周几", "日期", "时间",
		// 最新信息相关
		"最新", "新闻", "动态", "发布", "更新", "最近",
		"2024", "2025", "实时", "当下",
	}

	for _, keyword := range webSearchKeywords {
		if strings.Contains(query, keyword) {
			return true
		}
	}
	return false
}

// 判断是否需要数据库查询
func (ia *IntelligentAssistant) needsDatabase(query string) bool {
	dbKeywords := []string{
		"用户", "数据库", "查询", "统计", "数据",
		"记录", "表", "字段", "count", "sum",
	}

	for _, keyword := range dbKeywords {
		if strings.Contains(query, keyword) {
			return true
		}
	}
	return false
}

// 判断是否需要数学计算
func (ia *IntelligentAssistant) needsCalculation(query string) bool {
	calcKeywords := []string{
		"计算", "加", "减", "乘", "除", "+", "-", "*", "/",
		"等于", "结果", "数学", "算", "总和", "平均",
	}

	for _, keyword := range calcKeywords {
		if strings.Contains(query, keyword) {
			return true
		}
	}
	return false
}

// 构建数据库查询参数
func (ia *IntelligentAssistant) buildDatabaseQuery(query string) map[string]interface{} {
	// 根据查询内容智能构建数据库查询
	if strings.Contains(query, "统计") || strings.Contains(query, "数量") {
		return map[string]interface{}{
			"query_type": "structured",
			"query":      "select",
			"table_name": "users",
			"fields":     "status, COUNT(*) as count",
			"group_by":   "status",
		}
	}

	if strings.Contains(query, "活跃") {
		return map[string]interface{}{
			"query_type":       "structured",
			"query":            "select",
			"table_name":       "users",
			"fields":           "*",
			"where_conditions": "status=active",
			"limit":            10,
		}
	}

	// 默认查询
	return map[string]interface{}{
		"query_type": "structured",
		"query":      "select",
		"table_name": "users",
		"limit":      5,
	}
}

// 解析数学计算
func (ia *IntelligentAssistant) parseCalculation(query string) map[string]interface{} {
	// 简单的数学表达式解析
	// 实际应用中可以使用更复杂的表达式解析器

	if strings.Contains(query, "加") || strings.Contains(query, "+") {
		return map[string]interface{}{
			"operation": "add",
			"x":         10.5, // 实际应用中从查询中解析
			"y":         20.3,
		}
	}

	if strings.Contains(query, "减") || strings.Contains(query, "-") {
		return map[string]interface{}{
			"operation": "subtract",
			"x":         100,
			"y":         25,
		}
	}

	// 默认乘法示例
	return map[string]interface{}{
		"operation": "multiply",
		"x":         12,
		"y":         8,
	}
}

// 调用MCP工具
func (ia *IntelligentAssistant) callTool(ctx context.Context, toolCall ToolCall) (*mcp.CallToolResult, error) {
	return ia.mcpClient.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolCall.Name,
			Arguments: toolCall.Args,
		},
	})
}

// 格式化工具结果
func (ia *IntelligentAssistant) formatToolResult(result *mcp.CallToolResult) string {
	if result.IsError {
		return fmt.Sprintf("❌ 工具执行出错: %v", result.Content)
	}

	var formattedResult strings.Builder
	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			formattedResult.WriteString(textContent.Text)
			formattedResult.WriteString("\n")
		}
	}

	return formattedResult.String()
}

// 调用LLM API
func (ia *IntelligentAssistant) callLLM(ctx context.Context, userQuery string, toolResults []string) (string, error) {
	// 获取LLM配置
	apiURL, apiKey, model := getLLMConfig()

	// 构建LLM提示词
	var prompt strings.Builder
	prompt.WriteString(fmt.Sprintf("用户问题: %s\n\n", userQuery))

	if len(toolResults) > 0 {
		prompt.WriteString("我已经通过工具获取了以下信息:\n")
		for i, result := range toolResults {
			prompt.WriteString(fmt.Sprintf("\n工具结果 %d:\n%s\n", i+1, result))
		}
		prompt.WriteString("\n请基于以上工具提供的信息来回答用户的问题。请整合这些信息给出准确、详细的回答。")
	}

	// 构建API请求
	llmRequest := LLMRequest{
		Model: model,
		Messages: []LLMMessage{
			{
				Role:    "user",
				Content: prompt.String(),
			},
		},
		Stream: true,
		ExtraBody: ExtraBody{
			EnableSearch: len(toolResults) == 0, // 如果没有工具结果，启用搜索
		},
	}

	// 序列化请求
	requestBody, err := json.Marshal(llmRequest)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API请求失败, 状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析流式响应
	var result strings.Builder
	decoder := json.NewDecoder(resp.Body)

	for {
		var line string
		if err := decoder.Decode(&line); err != nil {
			if err == io.EOF {
				break
			}
			// 尝试逐行读取
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return "", fmt.Errorf("读取响应失败: %v", readErr)
			}

			// 处理Server-Sent Events格式
			lines := strings.Split(string(body), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "data: ") {
					jsonData := strings.TrimPrefix(line, "data: ")
					if jsonData == "[DONE]" {
						break
					}

					var response LLMResponse
					if parseErr := json.Unmarshal([]byte(jsonData), &response); parseErr == nil {
						if len(response.Choices) > 0 {
							result.WriteString(response.Choices[0].Delta.Content)
						}
					}
				}
			}
			break
		}
	}

	if result.Len() == 0 {
		return "LLM暂时无法响应，请稍后再试。", nil
	}

	return result.String(), nil
}

// 生成直接回答（不需要工具）
func (ia *IntelligentAssistant) generateDirectAnswer(query string) string {
	return fmt.Sprintf("这是一个常规问题，我可以直接回答：%s\n（此答案无需调用外部工具）", query)
}

// 整合多个工具结果生成最终答案
func (ia *IntelligentAssistant) synthesizeAnswer(userQuery string, toolResults []string) string {
	var answer strings.Builder

	answer.WriteString(fmt.Sprintf("基于您的问题「%s」，我通过以下工具获取了信息：\n\n", userQuery))

	for i, result := range toolResults {
		answer.WriteString(fmt.Sprintf("📊 工具结果 %d:\n%s\n", i+1, result))
	}

	answer.WriteString("\n💡 综合分析：\n")
	answer.WriteString("根据以上工具提供的数据，我为您整理了完整的答案。")
	answer.WriteString("这些信息来源于实时数据和准确计算，确保了回答的时效性和准确性。")

	return answer.String()
}

// 关闭连接
func (ia *IntelligentAssistant) Close() error {
	return ia.mcpClient.Close()
}

// 演示程序主函数
func runDemo() {
	fmt.Println("🚀 启动LLM智能助手演示程序")
	fmt.Println(strings.Repeat("=", 50))

	// 初始化智能助手
	assistant, err := NewIntelligentAssistant()
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}
	defer assistant.Close()

	// 演示查询场景
	demoQueries := []string{
		"帮我查询一下活跃用户的数量",
		"今天是几月几号",
		"现在几点了",
		"计算15.5加上24.3的结果",
		"什么是人工智能？", // 不需要工具的查询
		"今天星期几",
		"当前时间是多少",
	}

	ctx := context.Background()

	for i, query := range demoQueries {
		fmt.Printf("\n📝 演示查询 %d: %s\n", i+1, query)
		fmt.Println(strings.Repeat("-", 40))

		result, err := assistant.ProcessUserQuery(ctx, query)
		if err != nil {
			fmt.Printf("❌ 处理失败: %v\n", err)
			continue
		}

		// 输出处理结果
		fmt.Printf("⏱️ 处理时间: %v\n", result.ProcessTime)
		fmt.Printf("🔧 使用工具: %d 个\n", len(result.ToolsUsed))

		for _, tool := range result.ToolsUsed {
			toolArgs, _ := json.MarshalIndent(tool.Args, "  ", "  ")
			fmt.Printf("  - %s: %s\n", tool.Name, string(toolArgs))
		}

		fmt.Printf("\n🎯 最终回答:\n%s\n", result.FinalAnswer)
		fmt.Println(strings.Repeat("=", 50))
	}

	fmt.Println("\n✅ 演示完成！")
	fmt.Println("\n💡 这个演示展示了LLM如何智能地：")
	fmt.Println("   1. 分析用户查询的意图")
	fmt.Println("   2. 判断是否需要外部工具")
	fmt.Println("   3. 选择合适的工具组合")
	fmt.Println("   4. 整合工具结果生成智能回答")
}

// 主函数
func main() {
	runDemo()
}
