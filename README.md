# MCP Demo Server - 高级数据库工具服务器

这是一个基于 [MCP Go SDK](https://github.com/mark3labs/mcp-go) 开发的高级 Model Context Protocol (MCP) 服务器，专门为 LLM 提供强大的数据库操作、网络搜索和计算功能。

## 🚀 项目概述

本项目实现了一个功能丰富的 MCP 服务器，为大语言模型（LLM）提供以下核心能力：

- **🗄️ 数据库操作** - 支持 MySQL 数据库的复杂查询、CRUD 操作
- **🔍 网络搜索** - 集成 Google 搜索 API 进行实时信息检索  
- **🧮 数学计算** - 基础四则运算功能
- **🔧 多连接管理** - 支持多个数据库连接的并发管理

## 📖 什么是 MCP？

**Model Context Protocol (MCP)** 是一个开放标准协议，专门为 AI 应用程序与外部数据源和工具之间建立安全、可控的连接而设计。它为大语言模型（LLM）提供了一种标准化的方式来访问和交互外部系统，同时保持安全性和用户控制。

### 🎯 MCP 的核心价值

- **🔒 安全可控** - 严格的权限控制和安全边界
- **📏 标准化** - 统一的协议规范，确保互操作性
- **🔌 可扩展** - 灵活的架构支持各种工具和数据源
- **🎮 用户控制** - 用户完全控制 LLM 可以访问的资源

## 🏗️ MCP 核心概念

### 1. *    return tools
}
```

### 🚀 完整演示示例

我们提供了一个完整的演示程序，展示LLM如何智能调用MCP工具：

👉 **[查看完整演示代码](./examples/README.md)**

运行演示：
```bash
# 1. 启动MCP服务器
go run main.go

# 2. 在新终端运行演示
cd examples
go mod init mcp-demo-integration  
go get github.com/mark3labs/mcp-go
go run llm_integration_demo.go
```

演示将展示以下智能场景：
- ✅ 自动数据库查询："查询活跃用户数量"
- ✅ 实时信息搜索："2025年Go最新特性"  
- ✅ 数学计算："计算15.5+24.3"
- ✅ 常规问答："什么是人工智能"

## � 高级配置Tools (工具)**
类似于 API 的 POST 端点，为 LLM 提供执行操作的能力：
- 执行计算和业务逻辑
- 产生副作用（如数据修改）
- 接受结构化参数输入
- 返回结构化结果

**示例用途：**
- 数据库查询和更新
- 文件操作
- API 调用
- 复杂计算

### 2. **📚 Resources (资源)**
类似于 API 的 GET 端点，向 LLM 暴露数据：
- **静态资源** - 固定 URI 的数据源
- **动态资源** - 使用 URI 模板的参数化数据源
- 只读访问，不产生副作用

**示例用途：**
- 配置信息
- 文件内容
- 数据库表结构
- 实时状态信息

### 3. **💬 Prompts (提示模板)**
预定义的 LLM 交互模式：
- 可重用的对话模板
- 参数化提示内容
- 标准化的 LLM 指令

**示例用途：**
- 代码审查模板
- 数据分析指南
- 问题诊断流程

### 4. **🖥️ Server (服务器)**
MCP 协议的实现核心：
- 处理连接管理
- 消息路由和协议兼容
- 工具、资源和提示的注册管理

## 📁 项目结构

```
mcp-demo-server/
├── go.mod              # Go 模块依赖定义
├── go.sum              # 依赖版本锁定
├── main.go             # 主服务器实现
├── README.md           # 项目文档
└── demo.db             # SQLite 示例数据库（运行时生成）
```

## ⚡ 快速开始

### 1. 环境准备

确保您的系统已安装：
- **Go 1.21+** 
- **MySQL 8.0+** (可选，支持 SQLite)
- **Git**

### 2. 克隆和安装

```bash
# 克隆项目
git clone <repository-url>
cd mcp-demo-server

# 安装依赖
go mod tidy
```

### 3. 数据库配置

#### 选项 A: 使用 MySQL（推荐）
```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE mcp_demo;
```

修改 `main.go` 中的数据库配置：
```go
config := DatabaseConfig{
    Driver:   "mysql",
    Host:     "localhost",
    Port:     3306,
    Database: "mcp_demo",
    Username: "root",     // 您的用户名
    Password: "root",     // 您的密码
}
```

#### 选项 B: 使用 SQLite（简单部署）
配置已内置，无需额外设置。服务器启动时会自动创建 `demo.db` 文件。

### 4. Google 搜索配置（可选）

如需启用网络搜索功能，请设置环境变量：
```bash
export GOOGLE_API_KEY="your-google-api-key"
export GOOGLE_SEARCH_ENGINE_ID="your-search-engine-id"
```

### 5. 启动服务器

```bash
# 编译并运行
go run main.go

# 或编译后运行
go build -o mcp-server
./mcp-server
```

服务器启动后将通过标准输入输出（stdio）协议等待客户端连接。

## 🛠️ 服务器功能详解

### 核心工具 (Tools)

#### 1. 🧮 calculator - 数学计算器
**功能**: 执行基本四则运算

**参数**:
- `operation` (string, 必需): 运算类型
  - `add` - 加法
  - `subtract` - 减法  
  - `multiply` - 乘法
  - `divide` - 除法
- `x` (number, 必需): 第一个操作数
- `y` (number, 必需): 第二个操作数

**使用示例**:
```json
{
  "name": "calculator",
  "arguments": {
    "operation": "add",
    "x": 15.5,
    "y": 24.3
  }
}
```

#### 2. 🗄️ database_query - 高级数据库查询
**功能**: 提供多种数据库查询模式，支持原始 SQL、结构化查询和模型查询

**核心参数**:
- `query_type` (string): 查询类型
  - `raw` - 原始 SQL 查询（仅支持 SELECT）
  - `structured` - 结构化查询构建器
  - `model` - 预定义模型查询
- `query` (string, 必需): 查询内容
- `database` (string): 数据库连接名称（默认: "default"）

**结构化查询专属参数**:
- `table_name` (string): 目标表名
- `fields` (string): 查询字段（默认: "*"）
- `where_conditions` (string): WHERE 条件
- `order_by` (string): 排序规则
- `limit` (number): 结果限制数量
- `offset` (number): 分页偏移量
- `group_by` (string): 分组字段
- `having` (string): HAVING 条件
- `join_tables` (string): JSON 格式的关联表信息

**使用示例**:

**原始 SQL 查询**:
```json
{
  "name": "database_query",
  "arguments": {
    "query_type": "raw",
    "query": "SELECT * FROM users WHERE status = 'active' LIMIT 10"
  }
}
```

**结构化查询**:
```json
{
  "name": "database_query", 
  "arguments": {
    "query_type": "structured",
    "query": "select",
    "table_name": "users",
    "fields": "id,name,email,status",
    "where_conditions": "status=active,created_at>2024-01-01",
    "order_by": "created_at DESC",
    "limit": 20
  }
}
```

**模型查询**:
```json
{
  "name": "database_query",
  "arguments": {
    "query_type": "model", 
    "model_name": "users",
    "query": "active"
  }
}
```

#### 3. 🔍 web_search - 网络搜索
**功能**: 使用 Google Custom Search API 进行实时网络搜索

**参数**:
- `query` (string, 必需): 搜索关键词
- `limit` (number): 结果数量限制（默认: 10，最大: 20）

**使用示例**:
```json
{
  "name": "web_search",
  "arguments": {
    "query": "Go programming language tutorial",
    "limit": 5
  }
}
```

### 数据库功能特性

#### 🔗 多连接管理
- 支持同时连接多个数据库
- 连接池自动管理和优化
- 支持 MySQL、PostgreSQL、SQLite

#### 🛡️ 安全特性
- SQL 注入防护
- 只读查询限制（原始 SQL 模式）
- 参数化查询支持
- 操作权限验证

#### 📊 查询构建器
结构化查询支持复杂的 SQL 构建：

**WHERE 条件格式**:
```bash
# 简单格式
field1=value1,field2>value2,field3!=value3

# JSON 格式  
{"field1": "value1", "field2": "value2"}
```

**JOIN 操作格式**:
```json
[
  {
    "table": "orders", 
    "on": "users.id=orders.user_id",
    "type": "LEFT"
  }
]
```

## 🎛️ 客户端集成

### Claude Desktop 集成

在 Claude Desktop 配置文件中添加：

```json
{
  "mcpServers": {
    "database-tools": {
      "command": "go",
      "args": ["run", "/path/to/mcp-demo-server/main.go"],
      "env": {
        "GOOGLE_API_KEY": "your-api-key",
        "GOOGLE_SEARCH_ENGINE_ID": "your-search-engine-id"
      }
    }
  }
}
```

### 自定义 LLM 客户端

```go
package main

import (
    "context"
    "github.com/mark3labs/mcp-go/client"
    "github.com/mark3labs/mcp-go/mcp"
)

func main() {
    // 创建 stdio 客户端
    c, err := client.NewStdioMCPClient(
        "go", []string{"run", "/path/to/main.go"},
    )
    if err != nil {
        panic(err)
    }
    defer c.Close()

    ctx := context.Background()
    
    // 初始化连接
    if err := c.Initialize(ctx); err != nil {
        panic(err)
    }

    // 调用数据库查询工具
    result, err := c.CallTool(ctx, mcp.CallToolRequest{
        Params: mcp.CallToolRequestParams{
            Name: "database_query",
            Arguments: map[string]interface{}{
                "query_type": "structured",
                "query": "select", 
                "table_name": "users",
                "limit": 10,
            },
        },
    })
}
```

## 🤖 LLM智能工具调用集成

### 集成原理

当用户向LLM提问时，LLM会自动判断是否需要调用外部工具来获取信息：

```
用户提问 → LLM分析 → 判断知识边界 → 选择合适工具 → 调用MCP服务 → 整合结果 → 回答用户
```

### 🎯 触发场景示例

#### 1. 实时信息查询
```
用户：2025年最新的Go语言特性有哪些？
LLM思考：我的训练数据可能不包含2025年的最新信息
动作：调用 web_search 工具搜索最新信息
```

#### 2. 数据库信息查询
```
用户：帮我查看一下用户表中活跃用户的统计信息
LLM思考：这需要查询具体的数据库
动作：调用 database_query 工具查询数据
```

#### 3. 复杂计算
```
用户：计算复合利率：本金10000，年利率5.5%，复利10年
LLM思考：这需要精确的数学计算
动作：调用 calculator 工具进行计算
```

### 🔧 集成实现方式

#### 方式1: Claude Desktop集成（推荐新手）

**步骤1: 配置Claude Desktop**

找到Claude Desktop配置文件位置：
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Linux**: `~/.config/claude/claude_desktop_config.json`

**步骤2: 添加MCP服务器配置**

```json
{
  "mcpServers": {
    "database-tools": {
      "command": "go",
      "args": ["run", "/path/to/mcp-demo-server/main.go"],
      "env": {
        "GOOGLE_API_KEY": "your-google-api-key",
        "GOOGLE_SEARCH_ENGINE_ID": "your-search-engine-id",
        "DB_HOST": "localhost",
        "DB_PORT": "3306",
        "DB_NAME": "mcp_demo",
        "DB_USER": "root",
        "DB_PASSWORD": "password"
      }
    }
  }
}
```

**步骤3: 重启Claude Desktop**

配置完成后，Claude会自动连接到您的MCP服务器，可以智能调用工具。

#### 方式2: API集成（开发者推荐）

创建一个智能LLM应用，集成多个MCP服务器：

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "strings"

    "github.com/mark3labs/mcp-go/client"
    "github.com/mark3labs/mcp-go/mcp"
)

// LLM应用结构
type IntelligentLLMApp struct {
    mcpClient     client.Client
    availableTools []mcp.Tool
}

// 初始化应用
func NewIntelligentLLMApp() (*IntelligentLLMApp, error) {
    // 连接MCP服务器
    mcpClient, err := client.NewStdioMCPClient(
        "go", []string{"run", "./main.go"},
    )
    if err != nil {
        return nil, err
    }

    ctx := context.Background()
    if err := mcpClient.Initialize(ctx); err != nil {
        return nil, err
    }

    // 获取可用工具
    toolsResult, err := mcpClient.ListTools(ctx)
    if err != nil {
        return nil, err
    }

    return &IntelligentLLMApp{
        mcpClient:     mcpClient,
        availableTools: toolsResult.Tools,
    }, nil
}

// 智能处理用户查询
func (app *IntelligentLLMApp) ProcessQuery(ctx context.Context, userQuery string) (string, error) {
    // 1. 分析用户查询，判断是否需要工具
    toolNeeded := app.analyzeQueryForTools(userQuery)
    
    if toolNeeded == nil {
        // 直接使用LLM回答
        return app.callLLM(ctx, userQuery, nil)
    }

    // 2. 调用相应的MCP工具
    toolResult, err := app.callTool(ctx, toolNeeded)
    if err != nil {
        return "", fmt.Errorf("工具调用失败: %v", err)
    }

    // 3. 将工具结果与原查询一起发送给LLM
    enhancedPrompt := fmt.Sprintf(`
用户问题: %s

工具查询结果:
%s

请基于以上工具提供的信息来回答用户的问题。如果工具结果与问题相关，请整合这些信息给出准确回答。
`, userQuery, app.formatToolResult(toolResult))

    return app.callLLM(ctx, enhancedPrompt, toolResult)
}

// 分析查询是否需要工具
func (app *IntelligentLLMApp) analyzeQueryForTools(query string) *ToolCall {
    query = strings.ToLower(query)
    
    // 实时信息查询
    if strings.Contains(query, "最新") || strings.Contains(query, "今天") || 
       strings.Contains(query, "现在") || strings.Contains(query, "当前") ||
       strings.Contains(query, "2024") || strings.Contains(query, "2025") {
        return &ToolCall{
            Name: "web_search",
            Args: map[string]interface{}{
                "query": query,
                "limit": 5,
            },
        }
    }
    
    // 数据库查询
    if strings.Contains(query, "用户") || strings.Contains(query, "数据库") ||
       strings.Contains(query, "查询") || strings.Contains(query, "统计") {
        return &ToolCall{
            Name: "database_query",
            Args: map[string]interface{}{
                "query_type": "structured",
                "query": "select",
                "table_name": "users",
                "limit": 10,
            },
        }
    }
    
    // 数学计算
    if strings.Contains(query, "计算") || strings.Contains(query, "加") ||
       strings.Contains(query, "减") || strings.Contains(query, "乘") ||
       strings.Contains(query, "除") {
        // 这里可以解析具体的数学表达式
        return app.parseCalculation(query)
    }
    
    return nil
}

// 工具调用结构
type ToolCall struct {
    Name string
    Args map[string]interface{}
}

// 调用工具
func (app *IntelligentLLMApp) callTool(ctx context.Context, toolCall *ToolCall) (*mcp.CallToolResult, error) {
    return app.mcpClient.CallTool(ctx, mcp.CallToolRequest{
        Params: mcp.CallToolRequestParams{
            Name:      toolCall.Name,
            Arguments: toolCall.Args,
        },
    })
}

// 格式化工具结果
func (app *IntelligentLLMApp) formatToolResult(result *mcp.CallToolResult) string {
    if result.IsError {
        return fmt.Sprintf("工具执行出错: %v", result.Content)
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

// 调用LLM (这里需要集成具体的LLM API)
func (app *IntelligentLLMApp) callLLM(ctx context.Context, prompt string, toolResult *mcp.CallToolResult) (string, error) {
    // 这里集成您选择的LLM API (OpenAI, Claude, etc.)
    // 示例代码...
    return "基于工具结果，LLM处理后的智能回答", nil
}

// 解析计算请求
func (app *IntelligentLLMApp) parseCalculation(query string) *ToolCall {
    // 简单的数学表达式解析
    // 实际应用中可以使用更复杂的NLP解析
    return &ToolCall{
        Name: "calculator",
        Args: map[string]interface{}{
            "operation": "add",
            "x":         10,
            "y":         20,
        },
    }
}
```

### 🎯 实际应用场景演示

#### 场景1: 智能数据分析助手

```
用户: "分析一下最近注册的用户趋势"

LLM思考过程:
1. 识别这是数据分析请求
2. 需要查询数据库获取用户注册数据
3. 调用database_query工具

工具调用:
{
  "name": "database_query",
  "arguments": {
    "query_type": "structured", 
    "query": "select",
    "table_name": "users",
    "fields": "created_at, COUNT(*) as count",
    "where_conditions": "created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)",
    "group_by": "DATE(created_at)",
    "order_by": "created_at DESC"
  }
}

LLM整合回答:
"根据数据库查询结果，最近30天的用户注册趋势如下：
[具体分析数据和图表描述]
建议：[基于数据的专业建议]"
```

#### 场景2: 实时信息搜索

```
用户: "Go 1.23版本有什么新特性？"

LLM思考过程:
1. 识别这是关于最新技术信息的查询
2. 我的训练数据可能不包含最新版本信息
3. 调用web_search工具获取最新信息

工具调用:
{
  "name": "web_search", 
  "arguments": {
    "query": "Go 1.23 new features changelog",
    "limit": 5
  }
}

LLM整合回答:
"基于最新搜索结果，Go 1.23版本的主要新特性包括：
[搜索结果整合和专业分析]"
```

### 💡 智能工具选择策略

MCP系统支持LLM根据查询内容智能选择最合适的工具：

```go
// 智能工具选择逻辑示例
func (app *IntelligentLLMApp) selectBestTools(query string) []ToolCall {
    var tools []ToolCall
    
    // 多工具组合使用
    if strings.Contains(query, "最新") && strings.Contains(query, "用户") {
        // 先搜索最新信息，再查询数据库
        tools = append(tools, 
            ToolCall{Name: "web_search", Args: buildSearchArgs(query)},
            ToolCall{Name: "database_query", Args: buildDBArgs(query)},
        )
    }
    
    return tools
}
```

## � 高级配置 