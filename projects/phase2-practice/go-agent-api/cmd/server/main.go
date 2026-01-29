package main

import (
	"fmt"
	"log"
	"time"

	"github.com/agent-learning/go-agent-api/internal/agent"
	"github.com/agent-learning/go-agent-api/internal/scheduler"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("🎯 TaskQueue 测试 Demo")
	fmt.Println("=" + repeat("=", 70))
	fmt.Println()

	// 测试 1: 基础队列操作
	demo1BasicQueueOperations()

	// 测试 2: 优先级队列
	demo2PriorityQueue()

	// 测试 3: 并发安全测试
	demo3ConcurrentSafety()

	fmt.Println()
	fmt.Println("✅ 所有测试完成!")
}

// 测试1: 基础队列操作
func demo1BasicQueueOperations() {
	fmt.Println("📋 测试 1: 基础队列操作")
	fmt.Println("-" + repeat("-", 70))

	// 创建队列
	queue := scheduler.NewTaskQueue()
	fmt.Printf("✓ 创建队列成功，当前大小: %d\n", queue.Size())

	// 添加任务
	tasks := []*agent.Task{
		createTask("task-1", "第一个任务", 1),
		createTask("task-2", "第二个任务", 1),
		createTask("task-3", "第三个任务", 1),
	}

	fmt.Println("\n📥 添加任务到队列:")
	for _, task := range tasks {
		queue.Enqueue(task)
		fmt.Printf("  ✓ 添加任务: %s (优先级: %d)\n", task.ID, task.Priority)
	}
	fmt.Printf("队列大小: %d\n", queue.Size())

	// 查看堆顶
	peek := queue.Peek()
	if peek != nil {
		fmt.Printf("\n👀 查看堆顶任务: %s\n", peek.ID)
	}

	// 出队
	fmt.Println("\n📤 从队列取出任务:")
	for queue.Size() > 0 {
		task := queue.Dequeue()
		if task != nil {
			fmt.Printf("  ✓ 取出任务: %s (优先级: %d)\n", task.ID, task.Priority)
		}
	}
	fmt.Printf("队列大小: %d\n", queue.Size())

	fmt.Println()
}

// 测试2: 优先级队列
func demo2PriorityQueue() {
	fmt.Println("🎯 测试 2: 优先级调度")
	fmt.Println("-" + repeat("-", 70))

	queue := scheduler.NewTaskQueue()

	// 创建不同优先级的任务
	tasks := []*agent.Task{
		createTask("low-1", "低优先级任务1", 1),
		createTask("high-1", "高优先级任务1", 10),
		createTask("medium-1", "中优先级任务1", 5),
		createTask("low-2", "低优先级任务2", 1),
		createTask("high-2", "高优先级任务2", 10),
		createTask("medium-2", "中优先级任务2", 5),
	}

	// 乱序添加
	fmt.Println("\n📥 按乱序添加任务:")
	for _, task := range tasks {
		queue.Enqueue(task)
		fmt.Printf("  添加: %-15s 优先级: %2d\n", task.ID, task.Priority)
	}

	fmt.Println("\n📤 按优先级顺序取出:")
	order := 1
	for queue.Size() > 0 {
		task := queue.Dequeue()
		if task != nil {
			fmt.Printf("  %d. %-15s 优先级: %2d (%s)\n",
				order, task.ID, task.Priority, task.Input)
			order++
		}
	}

	fmt.Println("\n💡 观察: 高优先级任务(10)优先执行，其次是中优先级(5)，最后是低优先级(1)")
	fmt.Println()
}

// 测试3: 并发安全
func demo3ConcurrentSafety() {
	fmt.Println("🔒 测试 3: 并发安全性")
	fmt.Println("-" + repeat("-", 70))

	queue := scheduler.NewTaskQueue()

	// 模拟多个goroutine并发操作
	numGoroutines := 10
	tasksPerGoroutine := 10

	fmt.Printf("\n🚀 启动 %d 个goroutine，每个添加 %d 个任务\n",
		numGoroutines, tasksPerGoroutine)

	// 启动多个生产者
	done := make(chan bool, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < tasksPerGoroutine; j++ {
				task := createTask(
					fmt.Sprintf("g%d-task%d", id, j),
					fmt.Sprintf("Goroutine %d 的任务 %d", id, j),
					(id+j)%5+1, // 优先级 1-5
				)
				queue.Enqueue(task)
				time.Sleep(time.Millisecond) // 模拟一些延迟
			}
			done <- true
		}(i)
	}

	// 等待所有生产者完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	expectedTotal := numGoroutines * tasksPerGoroutine
	actualTotal := queue.Size()

	fmt.Printf("\n📊 统计:")
	fmt.Printf("\n  预期任务数: %d", expectedTotal)
	fmt.Printf("\n  实际任务数: %d", actualTotal)

	if expectedTotal == actualTotal {
		fmt.Println("\n  ✅ 并发安全测试通过！没有数据丢失或竞争")
	} else {
		fmt.Println("\n  ❌ 检测到数据不一致")
	}

	// 测试并发读取
	fmt.Println("\n🔍 测试并发读取:")
	readers := 5
	readDone := make(chan bool, readers)

	for i := 0; i < readers; i++ {
		go func(id int) {
			// 多次读取队列大小和peek
			for j := 0; j < 10; j++ {
				_ = queue.Size()
				_ = queue.Peek()
				time.Sleep(time.Millisecond)
			}
			readDone <- true
		}(i)
	}

	// 等待所有读者完成
	for i := 0; i < readers; i++ {
		<-readDone
	}

	fmt.Println("  ✅ 并发读取测试通过！")

	// 清空队列
	fmt.Printf("\n🧹 清空队列 (%d 个任务)...\n", queue.Size())
	count := 0
	for queue.Size() > 0 {
		queue.Dequeue()
		count++
	}
	fmt.Printf("  ✓ 已移除 %d 个任务\n", count)

	fmt.Println()
}

// 辅助函数：创建测试任务
func createTask(id, input string, priority int) *agent.Task {
	return &agent.Task{
		ID:        id,
		AgentID:   uuid.New().String(),
		Type:      agent.TaskTypeQuery,
		Input:     input,
		Status:    agent.TaskStatusPending,
		Priority:  priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// 辅助函数：重复字符串
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// 可视化堆结构（额外功能）
func visualizeQueue(queue *scheduler.TaskQueue) {
	tasks := queue.List()
	if len(tasks) == 0 {
		fmt.Println("  (空队列)")
		return
	}

	fmt.Println("\n  堆结构可视化:")
	fmt.Println("  " + repeat("-", 40))

	// 简单的树形展示（仅显示前几层）
	levels := [][]int{
		{0},           // 第0层：根节点
		{1, 2},        // 第1层：2个节点
		{3, 4, 5, 6},  // 第2层：4个节点
	}

	for levelNum, level := range levels {
		indent := repeat("  ", 3-levelNum)
		fmt.Print(indent)

		for _, idx := range level {
			if idx < len(tasks) {
				task := tasks[idx]
				fmt.Printf("[%s:P%d] ", task.ID[:6], task.Priority)
			}
		}
		fmt.Println()
	}

	if len(tasks) > 7 {
		fmt.Printf("  ... 还有 %d 个任务\n", len(tasks)-7)
	}
	fmt.Println()
}

// 性能测试（可选）
func demoBenchmark() {
	fmt.Println("⚡ 性能基准测试")
	fmt.Println("-" + repeat("-", 70))

	queue := scheduler.NewTaskQueue()
	numTasks := 10000

	// 测试入队性能
	fmt.Printf("\n📥 测试入队性能 (%d 个任务)...\n", numTasks)
	start := time.Now()

	for i := 0; i < numTasks; i++ {
		task := createTask(
			fmt.Sprintf("task-%d", i),
			fmt.Sprintf("测试任务 %d", i),
			i%10+1,
		)
		queue.Enqueue(task)
	}

	enqueueTime := time.Since(start)
	fmt.Printf("  入队耗时: %v\n", enqueueTime)
	fmt.Printf("  平均每个: %v\n", enqueueTime/time.Duration(numTasks))
	fmt.Printf("  吞吐量: %.0f ops/sec\n",
		float64(numTasks)/enqueueTime.Seconds())

	// 测试出队性能
	fmt.Printf("\n📤 测试出队性能 (%d 个任务)...\n", numTasks)
	start = time.Now()

	count := 0
	for queue.Size() > 0 {
		queue.Dequeue()
		count++
	}

	dequeueTime := time.Since(start)
	fmt.Printf("  出队耗时: %v\n", dequeueTime)
	fmt.Printf("  平均每个: %v\n", dequeueTime/time.Duration(count))
	fmt.Printf("  吞吐量: %.0f ops/sec\n",
		float64(count)/dequeueTime.Seconds())

	fmt.Println()
}

// 添加一个初始化日志函数
func init() {
	log.SetFlags(0) // 移除默认的时间戳
}
