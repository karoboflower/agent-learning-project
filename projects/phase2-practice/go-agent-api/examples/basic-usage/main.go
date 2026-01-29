package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/agent-learning/go-agent-api/internal/agent"
	"github.com/agent-learning/go-agent-api/internal/scheduler"
)

func main() {
	fmt.Println("🚀 Go Agent API - 基础使用示例")
	fmt.Println(strings.Repeat("=", 52))
	fmt.Println()

	demo1BasicQueueOperations()
	demo2PriorityDemo()
	demo3StatusManagement()
	printSummary()
}

func demo1BasicQueueOperations() {
	fmt.Println("📋 示例 1: 任务队列基础操作")
	fmt.Println(strings.Repeat("-", 52))

	// 1.1 创建任务队列
	queue := scheduler.NewTaskQueue()
	fmt.Printf("✓ 创建空队列，当前长度: %d\n\n", queue.Len())

	// 1.2 创建三个不同优先级的任务
	task1 := &agent.Task{
		ID:        "task-1",
		AgentID:   "agent-001",
		Type:      agent.TaskTypeQuery,
		Input:     "普通任务 - 优先级 1",
		Status:    agent.TaskStatusPending,
		Priority:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	task2 := &agent.Task{
		ID:        "task-2",
		AgentID:   "agent-001",
		Type:      agent.TaskTypeQuery,
		Input:     "高优先级任务 - 优先级 10",
		Status:    agent.TaskStatusPending,
		Priority:  10,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	task3 := &agent.Task{
		ID:        "task-3",
		AgentID:   "agent-001",
		Type:      agent.TaskTypeQuery,
		Input:     "中等优先级任务 - 优先级 5",
		Status:    agent.TaskStatusPending,
		Priority:  5,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1.3 将任务加入队列
	fmt.Println("📥 添加任务到队列:")
	queue.Enqueue(task1)
	fmt.Printf("  • 任务 %s 已加入 (优先级: %d)\n", task1.ID, task1.Priority)

	queue.Enqueue(task2)
	fmt.Printf("  • 任务 %s 已加入 (优先级: %d)\n", task2.ID, task2.Priority)

	queue.Enqueue(task3)
	fmt.Printf("  • 任务 %s 已加入 (优先级: %d)\n", task3.ID, task3.Priority)

	fmt.Printf("\n✓ 队列当前长度: %d\n", queue.Len())

	// 1.4 查看堆顶任务（最高优先级）
	topTask := queue.Peek()
	if topTask != nil {
		fmt.Printf("\n🔝 堆顶任务（最高优先级）:\n")
		fmt.Printf("  ID: %s\n", topTask.ID)
		fmt.Printf("  输入: %s\n", topTask.Input)
		fmt.Printf("  优先级: %d\n\n", topTask.Priority)
	}

	// 1.5 按优先级顺序出队
	fmt.Println("📤 按优先级顺序出队:")
	executionOrder := 1
	for queue.Len() > 0 {
		task := queue.Dequeue()
		if task != nil {
			fmt.Printf("  %d. [%s] %s (优先级: %d)\n",
				executionOrder, task.ID, task.Input, task.Priority)
			executionOrder++
		}
	}

	fmt.Printf("\n✓ 队列已清空，当前长度: %d\n\n", queue.Len())
}

func demo2PriorityDemo() {
	fmt.Println("🎯 示例 2: 优先级调度演示")
	fmt.Println(strings.Repeat("-", 52))

	queue := scheduler.NewTaskQueue()

	// 模拟实际场景的任务
	tasks := []struct {
		id       string
		input    string
		priority int
	}{
		{"urgent-bug", "紧急：修复生产环境 Bug", 10},
		{"code-review", "代码审查：PR #123", 7},
		{"feature-dev", "开发新功能：用户管理", 5},
		{"routine-test", "常规测试：回归测试", 3},
		{"documentation", "更新文档：API 文档", 2},
	}

	fmt.Println("📋 任务列表（提交顺序）:")
	for i, t := range tasks {
		task := &agent.Task{
			ID:        t.id,
			AgentID:   "agent-001",
			Type:      agent.TaskTypeQuery,
			Input:     t.input,
			Status:    agent.TaskStatusPending,
			Priority:  t.priority,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		queue.Enqueue(task)
		fmt.Printf("%d. [优先级:%2d] %s - %s\n", i+1, t.priority, t.id, t.input)
	}

	fmt.Println("\n🔄 执行顺序（按优先级）:")
	execOrder := 1
	for queue.Len() > 0 {
		task := queue.Dequeue()
		if task != nil {
			fmt.Printf("%d. [优先级:%2d] %s - %s\n",
				execOrder, task.Priority, task.ID, task.Input)
			execOrder++
		}
	}
	fmt.Println()
}

func demo3StatusManagement() {
	fmt.Println("📊 示例 3: 任务状态管理")
	fmt.Println(strings.Repeat("-", 52))

	queue := scheduler.NewTaskQueue()

	// 添加不同状态的任务
	for i := 1; i <= 3; i++ {
		status := agent.TaskStatusPending
		if i == 3 {
			status = agent.TaskStatusRunning
		}

		queue.Enqueue(&agent.Task{
			ID:        fmt.Sprintf("status-task-%d", i),
			AgentID:   "agent-001",
			Status:    status,
			Priority:  1,
			Input:     fmt.Sprintf("任务 %d", i),
			CreatedAt: time.Now(),
		})
	}

	// 按状态查询任务
	pendingTasks := queue.GetByStatus(agent.TaskStatusPending)
	runningTasks := queue.GetByStatus(agent.TaskStatusRunning)

	fmt.Printf("待处理任务数量: %d\n", len(pendingTasks))
	for _, task := range pendingTasks {
		fmt.Printf("  • %s: %s\n", task.ID, task.Input)
	}

	fmt.Printf("\n运行中任务数量: %d\n", len(runningTasks))
	for _, task := range runningTasks {
		fmt.Printf("  • %s: %s\n", task.ID, task.Input)
	}

	// 更新任务状态
	fmt.Println("\n🔄 更新任务状态:")
	updated := queue.UpdateTaskStatus("status-task-1", agent.TaskStatusRunning)
	if updated {
		fmt.Println("✓ 任务 status-task-1 状态已更新为 Running")
	}

	// 再次查询
	pendingTasks = queue.GetByStatus(agent.TaskStatusPending)
	runningTasks = queue.GetByStatus(agent.TaskStatusRunning)
	fmt.Printf("\n更新后:\n")
	fmt.Printf("  待处理任务: %d\n", len(pendingTasks))
	fmt.Printf("  运行中任务: %d\n\n", len(runningTasks))
}

func printSummary() {
	fmt.Println(strings.Repeat("=", 52))
	fmt.Println("✅ 所有示例执行完成！")
	fmt.Println(strings.Repeat("=", 52))

	fmt.Println("\n📚 学到的知识点:")
	fmt.Println("1. ✓ 任务队列的创建和基本操作")
	fmt.Println("2. ✓ 优先级调度机制（高优先级优先执行）")
	fmt.Println("3. ✓ 任务状态管理和查询")
	fmt.Println("4. ✓ 堆数据结构的实际应用")

	fmt.Println("\n🎓 下一步建议:")
	fmt.Println("• 阅读完整的调度器文档: docs/SCHEDULER-DEEP-DIVE.md")
	fmt.Println("• 运行完整的 API 服务器: go run cmd/server/main.go")
	fmt.Println("• 尝试 HTTP API 调用: curl http://localhost:8080/health")
}
