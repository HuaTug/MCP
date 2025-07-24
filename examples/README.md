# LLM与MCP工具集成演示

这个示例展示了如何将MCP工具服务器与**真实的LLM API**智能地集成，让LLM能够根据用户查询自动判断并调用合适的工具。

## 🎯 演示功能

### 智能工具选择
LLM会根据用户查询的内容，智能地判断是否需要调用外部工具：

1. **实时信息查询** - 自动调用 `web_search` 工具
2. **数据库查询** - 自动调用 `database_query` 工具  
3. **数学计算** - 自动调用 `calculator` 工具
4. **常规问答** - 直接调用LLM回答，无需工具

### 工作流程
```
用户提问 → 查询分析 → 工具选择 → 工具调用 → LLM整合 → 智能回答
```

## 🚀 运行演示

### 前置条件

1. **启动MCP服务器**：
```bash
cd /path/to/mcp-demo-server
go run main.go
```

2. **配置API密钥**（可选）：
```bash
cd examples
# 编辑配置文件，设置你的LLM API密钥
source config.sh
```

3. **运行演示**：
```bash
go mod init mcp-demo-integration
go get github.com/mark3labs/mcp-go
go run llm_integration_demo.go
```

### � API配置

演示程序使用**腾讯云深度求索API**，您可以：

1. **使用默认配置**（限量试用）
2. **配置自己的API密钥**：
   ```bash
   export LLM_API_KEY="your-api-key"
   export LLM_API_URL="your-api-endpoint"
   export LLM_MODEL="your-model-name"
   ```

## �📊 演示场景

### 场景1: 数据库查询 + LLM分析
```
用户输入: "帮我查询一下活跃用户的数量"

🔄 处理流程:
1. LLM分析: 检测到数据库查询需求
2. 调用工具: database_query
3. 工具结果: 返回用户统计数据
4. LLM整合: 基于真实数据生成专业分析

🎯 最终回答: LLM基于工具数据的智能分析和建议
```

### 场景2: 实时搜索 + LLM总结
```
用户输入: "2025年Go语言有什么最新特性？"

🔄 处理流程:
1. LLM分析: 检测到需要最新信息
2. 调用工具: web_search  
3. 工具结果: 返回最新搜索结果
4. LLM整合: 总结和分析搜索内容

🎯 最终回答: LLM基于搜索结果的专业技术总结
```

### 场景3: 数学计算 + LLM解释
```
用户输入: "计算15.5加上24.3的结果"

🔄 处理流程:
1. LLM分析: 检测到数学计算需求
2. 调用工具: calculator
3. 工具结果: 精确计算结果
4. LLM整合: 解释计算过程

🎯 最终回答: 准确结果 + LLM的专业解释
```

### 场景4: 纯LLM问答
```
用户输入: "什么是人工智能？"

🔄 处理流程:
1. LLM分析: 常规知识问题，无需外部工具
2. 直接调用: LLM API（启用搜索增强）
3. LLM回答: 基于训练知识和实时搜索的综合回答

🎯 最终回答: LLM的专业知识回答
```

## 🔧 核心技术实现

### 1. 智能查询分析
```go
func (ia *IntelligentAssistant) analyzeQueryForTools(query string) []ToolCall {
    // 多维度关键词检测
    if ia.needsWebSearch(query) {
        // 检测: "最新", "现在", "2025"等时效性关键词
    }
    if ia.needsDatabase(query) {
        // 检测: "用户", "查询", "统计"等数据关键词  
    }
    if ia.needsCalculation(query) {
        // 检测: "计算", "加减乘除"等数学关键词
    }
}
```

### 2. 真实LLM API集成
```go
func (ia *IntelligentAssistant) callLLM(ctx context.Context, userQuery string, toolResults []string) (string, error) {
    // 构建增强提示词
    prompt := fmt.Sprintf(`
用户问题: %s

工具查询结果:
%s

请基于工具提供的信息回答用户问题。
`, userQuery, strings.Join(toolResults, "\n"))

    // 调用腾讯云深度求索API
    llmRequest := LLMRequest{
        Model: "deepseek-v3-0324",
        Messages: []LLMMessage{{
            Role: "user", 
            Content: prompt,
        }},
        Stream: true,
        ExtraBody: ExtraBody{
            EnableSearch: true, // 启用搜索增强
        },
    }
    
    // 发送HTTP请求到LLM API...
}
```

### 3. 智能结果整合
- **有工具结果**: LLM基于工具数据进行专业分析
- **无工具结果**: LLM直接回答并启用搜索增强
- **错误处理**: LLM调用失败时使用备用回答机制

## 🎮 快速体验

### 方法1: 一键运行脚本
```bash
cd examples
./run_demo.sh
```

### 方法2: 手动运行
```bash
# 1. 启动MCP服务器
cd /path/to/mcp-demo-server
go run main.go &

# 2. 配置环境变量
cd examples
source config.sh

# 3. 运行演示
go run llm_integration_demo.go
```

### 方法3: 自定义配置
```bash
# 使用你自己的API密钥
export LLM_API_KEY="your-api-key"
export LLM_API_URL="your-api-endpoint"
export LLM_MODEL="your-model"

go run llm_integration_demo.go
```

## 💡 扩展应用

### 1. 添加更多LLM提供商
```go
// 支持OpenAI
if model == "gpt-4" {
    apiURL = "https://api.openai.com/v1/chat/completions"
}

// 支持Claude
if model == "claude-3" {
    apiURL = "https://api.anthropic.com/v1/messages"
}

// 支持本地LLM
if strings.HasPrefix(model, "local-") {
    apiURL = "http://localhost:11434/api/chat"
}
```

### 2. 工具链组合
```go
// 智能工具链：先搜索，再分析，最后计算
func (ia *IntelligentAssistant) executeToolChain(query string) {
    // 1. 网络搜索获取最新信息
    searchResults := ia.callTool("web_search", searchArgs)
    
    // 2. 基于搜索结果查询相关数据
    dbResults := ia.callTool("database_query", dbArgs)
    
    // 3. 对数据进行计算分析
    calcResults := ia.callTool("calculator", calcArgs)
    
    // 4. LLM整合所有结果
    return ia.callLLM(query, []string{searchResults, dbResults, calcResults})
}
```

### 3. 智能缓存机制
```go
// 缓存相似查询的结果
type QueryCache struct {
    cache map[string]CacheEntry
    mutex sync.RWMutex
}

// 语义相似度匹配
func (qc *QueryCache) findSimilarQuery(query string) (*CacheEntry, bool) {
    // 使用向量相似度或关键词匹配
    // 返回缓存的工具调用结果
}
```

## 🎯 生产应用建议

### 1. 安全控制
- **API密钥管理**: 使用密钥管理服务
- **访问控制**: 限制工具调用权限
- **数据脱敏**: 敏感数据处理

### 2. 性能优化  
- **并发调用**: 多工具并行执行
- **结果缓存**: 避免重复调用
- **连接池**: 复用数据库连接

### 3. 监控运维
- **调用链跟踪**: 记录完整调用路径
- **性能指标**: API延迟、成功率等
- **错误告警**: 异常情况及时通知

### 4. 用户体验
- **流式输出**: 实时显示LLM回答
- **进度指示**: 显示工具调用状态
- **交互优化**: 支持多轮对话

---

## 🎉 总结

这个演示展示了MCP协议的强大能力：

✅ **智能工具选择** - LLM自动判断何时需要外部帮助  
✅ **真实API集成** - 与生产级LLM API无缝对接  
✅ **动态能力扩展** - 通过工具无限扩展LLM能力  
✅ **智能结果整合** - LLM基于工具数据生成专业回答  

通过MCP，您的LLM应用不再局限于训练数据，而是成为了一个能够**主动获取信息、处理数据、实时学习**的智能助手！🚀

## 🔧 核心技术点

### 1. 智能查询分析
```go
func (ia *IntelligentAssistant) analyzeQueryForTools(query string) []ToolCall {
    // 检测关键词，判断查询类型
    if ia.needsWebSearch(query) {
        // 需要实时信息
    }
    if ia.needsDatabase(query) {
        // 需要数据库查询
    }
    if ia.needsCalculation(query) {
        // 需要数学计算
    }
}
```

### 2. 动态工具调用
```go
func (ia *IntelligentAssistant) callTool(ctx context.Context, toolCall ToolCall) (*mcp.CallToolResult, error) {
    return ia.mcpClient.CallTool(ctx, mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Name:      toolCall.Name,
            Arguments: toolCall.Args,
        },
    })
}
```

### 3. 结果整合
```go
func (ia *IntelligentAssistant) synthesizeAnswer(userQuery string, toolResults []string) string {
    // 将多个工具结果整合成连贯的回答
}
```

## 💡 扩展思路

### 1. 增加更多工具判断逻辑
- 文件操作检测
- API调用检测  
- 复杂数据分析检测

### 2. 工具链组合
- 先搜索再计算
- 先查询再分析
- 多步骤工具协作

### 3. 上下文记忆
- 会话历史管理
- 工具调用缓存
- 智能推荐下一步操作

### 4. 实际LLM集成
- OpenAI GPT集成
- Claude API集成
- 本地LLM集成

## 🎯 生产应用建议

1. **错误处理** - 增加完善的错误处理和重试机制
2. **性能优化** - 工具调用缓存和并发优化
3. **安全控制** - 工具访问权限和数据脱敏
4. **监控日志** - 详细的调用日志和性能监控
5. **配置管理** - 灵活的工具配置和参数管理

这个演示展示了MCP协议的强大之处：让LLM能够智能地扩展自己的能力边界！
