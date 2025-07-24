#!/bin/bash

echo "🚀 LLM与MCP工具集成演示"
echo "=============================="

# 设置环境变量
export LLM_API_URL="http://api.lkeap.cloud.tencent.com/v1/chat/completions"
export LLM_API_KEY="sk-qFPEqgpxmS8DJ0nJQ6gvdIkozY1k2oEZER2A4zRhLxBvtIHl"
export LLM_MODEL="deepseek-v3-0324"

# 设置Google搜索API配置
export GOOGLE_API_KEY="AIzaSyCJffa8kg0c1_Ef7zl18QUMZVvqGwBVtrM"
export GOOGLE_SEARCH_ENGINE_ID="e6676dbfd052c4ecf"

echo "✅ 环境变量已设置"
echo "📡 API端点: $LLM_API_URL"
echo "🤖 模型: $LLM_MODEL"
echo "🔍 Google API: ${GOOGLE_API_KEY:0:10}..."
echo ""

# 检查MCP服务器是否运行
echo "🔍 检查MCP服务器状态..."
if pgrep -f "go run.*main.go" > /dev/null; then
    echo "✅ MCP服务器正在运行"
else
    echo "❌ MCP服务器未运行，正在启动..."
    cd .. && go run main.go &
    sleep 3
    echo "✅ MCP服务器已启动"
fi

echo ""
echo "🎯 开始运行LLM集成演示..."
echo "=============================="

# 运行演示程序
go run llm_integration_demo.go
