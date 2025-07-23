# MCP Go 服务器开发指南

这是一个完整的 MCP (Model Context Protocol) Go 服务器示例，展示了如何使用 Go 开发 MCP 服务并与 LLM 集成。

## 什么是 MCP？

MCP (Model Context Protocol) 是一个专门为 LLM 应用程序设计的协议，它让 LLM 能够安全、标准化地访问外部数据源和工具。可以把它想象成专为 LLM 交互设计的 Web API。

## 核心概念

### 1. **Server (服务器)**
- 处理连接管理、协议兼容性和消息路由的核心接口

### 2. **Tools (工具)**
- 为 LLM 提供功能（类似 POST 端点）
- 执行计算和产生副作用
- 例如：计算器、文件操作、API 调用

### 3. **Resources (资源)**
- 向 LLM 暴露数据（类似 GET 端点）
- 静态资源（固定URI）和动态资源（使用URI模板）
- 例如：文件内容、配置信息、数据库查询

### 4. **Prompts (提示模板)**
- 定义 LLM 交互模式的可重用模板
- 例如：代码审查模板、数据分析模板

## 项目结构

```
mcp-demo-server/
├── go.mod              # Go 模块定义
├── main.go             # 主服务器代码
├── README.md           # 使用说明
└── examples/           # 示例配置文件
```

## 快速开始

### 1. 安装依赖

```bash
# 初始化 Go 模块
go mod init mcp-demo-server

# 安装 MCP Go SDK
go get github.com/mark3labs/mcp-go
```

### 2. 运行服务器

```bash
# 编译并运行
go run main.go
```

### 3. 通过 stdio 与服务器交互

服务器通过标准输入输出进行通信。你可以使用支持 MCP 的客户端连接：

```bash
# 使用 MCP 客户端连接
mcp-client --stdio "go run main.go"
```

## 服务器功能

### 工具 (Tools)

本服务器提供以下工具：

#### 1. calculator - 计算器
- **描述**: 执行基本数学运算
- **参数**:
  - `operation` (string): 运算类型 (add, subtract, multiply, divide)
  - `x` (number): 第一个数字
  - `y` (number): 第二个数字

#### 2. read_file - 文件读取
- **描述**: 读取文件内容
- **参数**:
  - `path` (string): 文件路径

#### 3. write_file - 文件写入
- **描述**: 写入文件内容
- **参数**:
  - `path` (string): 文件路径
  - `content` (string): 要写入的内容

#### 4. http_request - HTTP 请求
- **描述**: 发送 HTTP 请求
- **参数**:
  - `url` (string): 请求 URL
  - `method` (string): HTTP 方法 (默认: GET)
  - `body` (string): 请求体（可选）

#### 5. system_info - 系统信息
- **描述**: 获取系统信息
- **参数**: 无

#### 6. current_time - 当前时间
- **描述**: 获取当前时间
- **参数**:
  - `format` (string): 时间格式 (默认: "2006-01-02 15:04:05")
  - `timezone` (string): 时区 (默认: "Local")

### 资源 (Resources)

#### 静态资源
- `server://status` - 服务器状态信息

#### 动态资源
- `file://{path}` - 读取指定路径的文件内容
- `config://{key}` - 获取配置项的值

### 提示模板 (Prompts)

#### 1. code_review - 代码审查
- **描述**: 代码审查助手
- **参数**:
  - `language` (required): 编程语言
  - `focus` (optional): 审查重点

#### 2. data_analysis - 数据分析
- **描述**: 数据分析助手
- **参数**:
  - `data_type` (required): 数据类型

## 在 LLM 中使用

### Claude Desktop 集成

1. 在 Claude Desktop 的配置文件中添加：

```json
{
  "mcpServers": {
    "go-demo-server": {
      "command": "go",
      "args": ["run", "/path/to/your/main.go"],
      "env": {}
    }
  }
}
```

### 其他 LLM 客户端

大多数支持 MCP 的 LLM 客户端都可以通过 stdio 协议连接：

```bash
# 通用连接方式
your-llm-client --mcp-server "go run main.go"
```

## 使用示例

### 1. 使用计算器工具

```json
{
  "method": "tools/call",
  "params": {
    "name": "calculator",
    "arguments": {
      "operation": "add",
      "x": 10,
      "y": 5
    }
  }
}
```

### 2. 读取文件

```json
{
  "method": "tools/call", 
  "params": {
    "name": "read_file",
    "arguments": {
      "path": "./README.md"
    }
  }
}
```

### 3. 获取服务器状态

```json
{
  "method": "resources/read",
  "params": {
    "uri": "server://status"
  }
}
```

### 4. 使用代码审查提示

```json
{
  "method": "prompts/get",
  "params": {
    "name": "code_review",
    "arguments": {
      "language": "Go",
      "focus": "security"
    }
  }
}
```

## 扩展服务器

### 添加新工具

```go
// 添加新工具
newTool := mcp.NewTool("my_tool",
    mcp.WithDescription("我的自定义工具"),
    mcp.WithString("param1",
        mcp.Required(),
        mcp.Description("参数描述"),
    ),
)
s.AddTool(newTool, handleMyTool)

// 实现处理函数
func handleMyTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    param1, err := request.RequireString("param1")
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }
    
    // 处理逻辑
    result := "处理结果"
    
    return mcp.NewToolResultText(result), nil
}
```

### 添加新资源

```go
// 静态资源
resource := mcp.NewResource(
    "my://resource",
    "我的资源",
    mcp.WithResourceDescription("资源描述"),
    mcp.WithMIMEType("application/json"),
)
s.AddResource(resource, handleMyResource)

// 动态资源模板
template := mcp.NewResourceTemplate(
    "my://{id}/data",
    "动态资源",
    mcp.WithTemplateDescription("动态资源描述"),
)
s.AddResourceTemplate(template, handleMyTemplate)
```

### 添加提示模板

```go
prompt := mcp.NewPrompt("my_prompt",
    mcp.WithPromptDescription("我的提示模板"),
    mcp.WithArgument("param1",
        mcp.ArgumentDescription("参数描述"),
        mcp.RequiredArgument(),
    ),
)
s.AddPrompt(prompt, handleMyPrompt)
```

## 高级功能

### 会话管理

```go
// 启用会话管理
s := server.NewMCPServer(
    "Server Name",
    "1.0.0",
    server.WithToolCapabilities(true),
)

// 实现会话
type MySession struct {
    id           string
    notifChannel chan mcp.JSONRPCNotification
    isInitialized bool
}

// 注册会话
session := &MySession{
    id:           "user-123",
    notifChannel: make(chan mcp.JSONRPCNotification, 10),
}
s.RegisterSession(context.Background(), session)
```

### HTTP 传输

```go
// 启用 HTTP 传输而不是 stdio
httpServer := server.NewStreamableHTTPServer(mcpServer)
log.Printf("HTTP server listening on :8080/mcp")
if err := httpServer.Start(":8080"); err != nil {
    log.Fatalf("Server error: %v", err)
}
```

### 采样支持

```go
// 启用采样功能（调用 LLM）
mcpServer.EnableSampling()

// 在工具中使用采样
func handleAskLLM(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    question, _ := request.RequireString("question")
    
    samplingRequest := mcp.CreateMessageRequest{
        CreateMessageParams: mcp.CreateMessageParams{
            Messages: []mcp.SamplingMessage{
                {
                    Role: mcp.RoleUser,
                    Content: mcp.TextContent{
                        Type: "text",
                        Text: question,
                    },
                },
            },
            MaxTokens: 1000,
        },
    }
    
    serverFromCtx := server.ServerFromContext(ctx)
    result, err := serverFromCtx.RequestSampling(ctx, samplingRequest)
    // 处理结果...
}
```

## 部署

### Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mcp-server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mcp-server .
CMD ["./mcp-server"]
```

### 系统服务

```bash
# 创建 systemd 服务文件
sudo tee /etc/systemd/system/mcp-server.service << EOF
[Unit]
Description=MCP Go Server
After=network.target

[Service]
Type=simple
User=mcp
WorkingDirectory=/opt/mcp-server
ExecStart=/opt/mcp-server/mcp-server
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 启用并启动服务
sudo systemctl enable mcp-server
sudo systemctl start mcp-server
```

## 故障排除

### 常见问题

1. **连接失败**
   - 检查服务器是否正确启动
   - 确认客户端和服务器使用相同的传输协议

2. **工具调用失败**
   - 检查参数是否正确
   - 查看服务器日志获取错误信息

3. **资源访问失败**
   - 确认 URI 格式正确
   - 检查资源是否存在

### 调试

启用详细日志：

```go
mcpServer := server.NewMCPServer(
    "Server Name",
    "1.0.0",
    server.WithLogging(),  // 启用日志
)
```

## 参考资料

- [MCP 官方文档](https://modelcontextprotocol.io)
- [MCP Go SDK 文档](https://pkg.go.dev/github.com/mark3labs/mcp-go)
- [MCP 规范](https://spec.modelcontextprotocol.io)

## 许可证

本项目采用 MIT 许可证。 