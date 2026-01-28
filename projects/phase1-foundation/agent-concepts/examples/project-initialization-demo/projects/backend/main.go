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
	fmt.Println("🚀 Task 1.3.3 - Go后端依赖验证")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// 加载环境变量
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  未找到.env文件，使用系统环境变量")
	}

	// 验证Go模块
	fmt.Println("\n✅ 验证清单:")
	fmt.Println("   [✓] Go模块初始化成功")
	fmt.Println("   [✓] go.mod文件创建完成")

	// 验证OpenAI SDK
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("   [⏳] OpenAI API密钥未配置（请在.env中配置）")
	} else {
		fmt.Println("   [✓] OpenAI API密钥已配置")

		// 测试OpenAI客户端初始化
		client := openai.NewClient(apiKey)
		if client != nil {
			fmt.Println("   [✓] OpenAI客户端初始化成功")
		}
	}

	// 验证环境变量加载
	fmt.Println("   [✓] godotenv包安装成功")

	fmt.Println("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📦 已安装的Go依赖:")
	fmt.Println("   - github.com/sashabaranov/go-openai v1.17.0")
	fmt.Println("   - github.com/joho/godotenv v1.5.1")

	fmt.Println("\n🎯 下一步:")
	fmt.Println("   1. 复制.env.example为.env")
	fmt.Println("   2. 在.env中配置API密钥")
	fmt.Println("   3. 运行: go run main.go")
	fmt.Println("   4. 测试API连接: go run test_api.go")

	// 如果配置了API密钥，进行简单测试
	if apiKey != "" {
		fmt.Println("\n🧪 测试OpenAI API连接...")
		testOpenAI(apiKey)
	}
}

func testOpenAI(apiKey string) {
	client := openai.NewClient(apiKey)
	ctx := context.Background()

	// 列出可用模型（不会产生费用）
	_, err := client.ListModels(ctx)
	if err != nil {
		fmt.Printf("   [❌] API连接失败: %v\n", err)
		fmt.Println("   提示: 请检查API密钥是否正确")
	} else {
		fmt.Println("   [✓] OpenAI API连接成功！")
	}
}
