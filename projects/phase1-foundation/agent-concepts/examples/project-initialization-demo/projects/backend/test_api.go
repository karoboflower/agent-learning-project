package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	fmt.Println("🧪 测试OpenAI API连接")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❌ 错误: 无法加载.env文件")
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ 错误: OPENAI_API_KEY未设置")
	}

	fmt.Println("✓ API密钥已加载")

	// 创建客户端
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// 测试1: 列出模型
	fmt.Println("\n测试1: 列出可用模型...")
	models, err := client.ListModels(ctx)
	if err != nil {
		log.Fatalf("❌ 失败: %v\n", err)
	}
	fmt.Printf("✓ 成功! 找到 %d 个模型\n", len(models.Models))

	// 测试2: 简单的聊天完成
	fmt.Println("\n测试2: 测试Chat Completion API...")
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Say 'Hello from Go!'",
				},
			},
			MaxTokens: 20,
		},
	)

	if err != nil {
		log.Fatalf("❌ 失败: %v\n", err)
	}

	fmt.Println("✓ 成功!")
	fmt.Printf("回复: %s\n", resp.Choices[0].Message.Content)

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✨ 所有测试通过！OpenAI Go SDK工作正常")
}
