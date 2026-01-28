# Go后端项目 - OpenAI SDK 集成

Task 1.3.3 Go依赖安装和验证示例。

## ✅ 已完成

- [x] 创建Go模块
- [x] 安装OpenAI Go SDK
- [x] 安装其他必要依赖（godotenv）
- [x] 创建验证测试脚本

## 📦 依赖列表

### Go依赖
- **github.com/sashabaranov/go-openai** v1.17.0 - OpenAI Go SDK
- **github.com/joho/godotenv** v1.5.1 - 环境变量加载

## 🚀 快速开始

### 1. 初始化Go模块

```bash
# 已完成，go.mod已创建
go mod tidy
```

### 2. 下载依赖

```bash
go mod download
```

### 3. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入API密钥：
```env
OPENAI_API_KEY=your_openai_api_key_here
PORT=8080
```

### 4. 运行验证程序

```bash
# 基础验证
go run main.go

# API连接测试（需要配置API密钥）
go run test_api.go
```

## 📁 项目结构

```
backend/
├── go.mod              # Go模块配置
├── go.sum              # 依赖校验和（运行后生成）
├── .env.example        # 环境变量示例
├── main.go             # 主程序（依赖验证）
└── test_api.go         # API连接测试
```

## 🧪 验证清单

### 1. Go模块验证

```bash
# 验证go.mod
cat go.mod

# 下载依赖
go mod download

# 验证依赖
go mod verify
```

### 2. 编译验证

```bash
# 编译主程序
go build -o backend main.go

# 运行编译后的程序
./backend
```

### 3. API连接测试

```bash
# 配置API密钥后运行
go run test_api.go
```

预期输出：
```
🧪 测试OpenAI API连接
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ API密钥已加载

测试1: 列出可用模型...
✓ 成功! 找到 XX 个模型

测试2: 测试Chat Completion API...
✓ 成功!
回复: Hello from Go!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✨ 所有测试通过！OpenAI Go SDK工作正常
```

## 📚 OpenAI Go SDK 使用示例

### 基础聊天完成

```go
package main

import (
    "context"
    "fmt"
    "os"

    openai "github.com/sashabaranov/go-openai"
)

func main() {
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

    resp, err := client.CreateChatCompletion(
        context.Background(),
        openai.ChatCompletionRequest{
            Model: openai.GPT3Dot5Turbo,
            Messages: []openai.ChatCompletionMessage{
                {
                    Role:    openai.ChatMessageRoleUser,
                    Content: "Hello!",
                },
            },
        },
    )

    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    fmt.Println(resp.Choices[0].Message.Content)
}
```

### 流式响应

```go
stream, err := client.CreateChatCompletionStream(
    ctx,
    openai.ChatCompletionRequest{
        Model: openai.GPT3Dot5Turbo,
        Messages: messages,
        Stream: true,
    },
)

defer stream.Close()

for {
    response, err := stream.Recv()
    if errors.Is(err, io.EOF) {
        break
    }
    if err != nil {
        return err
    }
    fmt.Print(response.Choices[0].Delta.Content)
}
```

## 🔑 API密钥获取

### OpenAI API密钥
1. 访问 https://platform.openai.com/
2. 注册/登录账号
3. 进入 API Keys 页面
4. 创建新的API密钥
5. 复制密钥到 `.env` 文件

## 🛠️ 常用命令

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 验证依赖
go mod verify

# 查看依赖树
go mod graph

# 运行程序
go run main.go

# 构建可执行文件
go build -o backend main.go

# 运行测试
go test ./...

# 代码格式化
gofmt -w .

# 代码检查
golangci-lint run
```

## 🐛 常见问题

### 1. 依赖下载失败

```bash
# 设置GOPROXY
export GOPROXY=https://goproxy.cn,direct

# 或设置到环境变量
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.bashrc
source ~/.bashrc
```

### 2. API连接失败

- 检查API密钥是否正确
- 检查网络连接
- 检查API配额是否用尽
- 检查防火墙设置

### 3. 导入路径错误

确保go.mod中的module名称正确：
```go
module github.com/agent-learning/backend
```

## 📖 相关文档

- [OpenAI Go SDK](https://github.com/sashabaranov/go-openai)
- [Go Modules](https://go.dev/blog/using-go-modules)
- [OpenAI API文档](https://platform.openai.com/docs/api-reference)
- [godotenv](https://github.com/joho/godotenv)

## 🎯 下一步

1. ✅ 完成依赖安装
2. ✅ 验证Go SDK
3. ⏳ 实现业务逻辑
4. ⏳ 添加单元测试
5. ⏳ 集成CI/CD

---

**创建日期**: 2026-01-28
**任务来源**: phase1-tasks.md - 1.3.3 依赖安装和验证
