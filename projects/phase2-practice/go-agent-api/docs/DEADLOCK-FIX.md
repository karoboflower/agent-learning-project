# TaskQueue 死锁问题修复说明

## 🐛 问题描述

在使用 TaskQueue 时遇到死锁错误：

```
fatal error: all goroutines are asleep - deadlock!

goroutine 1 [sync.Mutex.Lock]:
...
github.com/agent-learning/go-agent-api/internal/scheduler.(*TaskQueue).Push
github.com/agent-learning/go-agent-api/internal/scheduler.(*TaskQueue).Enqueue
```

## 🔍 问题根因

### 死锁发生的调用链

```
Enqueue()
  ├─ mu.Lock() ────────────────────┐ (第1次获取锁)
  │                                │
  └─ heap.Push(tq, task)           │
       │                           │
       └─ tq.Push(x)                │
            │                      │
            └─ mu.Lock() ──────────┘ (第2次尝试获取同一个锁 → 死锁！)
```

### 原理解释

1. **Enqueue** 方法持有了 `mu` 锁
2. **heap.Push** 内部会回调 **Push** 方法
3. **Push** 方法尝试再次获取 `mu` 锁
4. 同一个 goroutine 无法重入已持有的 Mutex → **死锁**

这是典型的**重入锁问题**。Go 的 `sync.Mutex` 不支持重入（与Java的 `ReentrantLock` 不同）。

## ✅ 解决方案

### 核心原则

**heap.Interface 的方法不应该加锁**，因为：

1. 这些方法是被 `heap` 包内部调用的
2. 调用者（Enqueue/Dequeue）已经持有锁
3. 重复加锁会导致死锁

### 修复前后对比

#### ❌ 修复前（错误）

```go
// heap.Interface 方法 - 错误地加锁
func (tq *TaskQueue) Push(x interface{}) {
    tq.mu.Lock()          // ❌ 不应该在这里加锁
    defer tq.mu.Unlock()
    task := x.(*agent.Task)
    tq.tasks = append(tq.tasks, task)
}

func (tq *TaskQueue) Pop() interface{} {
    tq.mu.Lock()          // ❌ 不应该在这里加锁
    defer tq.mu.Unlock()
    // ...
}

func (tq *TaskQueue) Len() int {
    tq.mu.RLock()         // ❌ 不应该在这里加锁
    defer tq.mu.RUnlock()
    return len(tq.tasks)
}

func (tq *TaskQueue) Less(i, j int) bool {
    tq.mu.RLock()         // ❌ 不应该在这里加锁
    defer tq.mu.RUnlock()
    // ...
}

func (tq *TaskQueue) Swap(i, j int) {
    tq.mu.Lock()          // ❌ 不应该在这里加锁
    defer tq.mu.Unlock()
    tq.tasks[i], tq.tasks[j] = tq.tasks[j], tq.tasks[i]
}

// 公共方法 - 正确地加锁
func (tq *TaskQueue) Enqueue(task *agent.Task) {
    tq.mu.Lock()          // ✓ 在外层加锁
    defer tq.mu.Unlock()
    heap.Push(tq, task)   // heap.Push 会调用 Push → 死锁！
}
```

#### ✅ 修复后（正确）

```go
// heap.Interface 方法 - 不加锁
func (tq *TaskQueue) Push(x interface{}) {
    // ✓ 不加锁，调用者负责加锁
    task := x.(*agent.Task)
    tq.tasks = append(tq.tasks, task)
}

func (tq *TaskQueue) Pop() interface{} {
    // ✓ 不加锁
    old := tq.tasks
    n := len(old)
    task := old[n-1]
    old[n-1] = nil
    tq.tasks = old[0 : n-1]
    return task
}

func (tq *TaskQueue) Len() int {
    // ✓ 不加锁
    return len(tq.tasks)
}

func (tq *TaskQueue) Less(i, j int) bool {
    // ✓ 不加锁
    if tq.tasks[i].Priority == tq.tasks[j].Priority {
        return tq.tasks[i].CreatedAt.Before(tq.tasks[j].CreatedAt)
    }
    return tq.tasks[i].Priority > tq.tasks[j].Priority
}

func (tq *TaskQueue) Swap(i, j int) {
    // ✓ 不加锁
    tq.tasks[i], tq.tasks[j] = tq.tasks[j], tq.tasks[i]
}

// 公共方法 - 加锁保护
func (tq *TaskQueue) Enqueue(task *agent.Task) {
    tq.mu.Lock()          // ✓ 只在外层加锁
    defer tq.mu.Unlock()
    heap.Push(tq, task)   // ✓ 安全调用
}

func (tq *TaskQueue) Dequeue() *agent.Task {
    tq.mu.Lock()          // ✓ 只在外层加锁
    defer tq.mu.Unlock()
    if len(tq.tasks) == 0 {
        return nil
    }
    return heap.Pop(tq).(*agent.Task)  // ✓ 安全调用
}

// 其他需要线程安全的方法
func (tq *TaskQueue) Size() int {
    tq.mu.RLock()         // ✓ 需要线程安全的读取
    defer tq.mu.RUnlock()
    return len(tq.tasks)
}

func (tq *TaskQueue) Peek() *agent.Task {
    tq.mu.RLock()         // ✓ 需要线程安全的读取
    defer tq.mu.RUnlock()
    if len(tq.tasks) == 0 {
        return nil
    }
    return tq.tasks[0]
}
```

## 📚 设计原则

### 1. heap.Interface 方法职责

这5个方法是 `container/heap` 的回调接口：

```go
type Interface interface {
    sort.Interface
    Push(x interface{})  // 添加x到末尾
    Pop() interface{}    // 移除并返回末尾元素
}

type Interface interface {
    Len() int              // 返回长度
    Less(i, j int) bool    // 比较元素
    Swap(i, j int)         // 交换元素
}
```

这些方法：
- **不负责线程安全**
- **只负责数据操作**
- **被heap包内部调用**

### 2. 锁的分层设计

```
┌─────────────────────────────────┐
│   Public Methods (加锁层)        │
│   - Enqueue()  [Lock]           │
│   - Dequeue()  [Lock]           │
│   - Remove()   [Lock]           │
│   - Size()     [RLock]          │
│   - Peek()     [RLock]          │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│   heap.Interface (无锁层)       │
│   - Push()     [No Lock]        │
│   - Pop()      [No Lock]        │
│   - Len()      [No Lock]        │
│   - Less()     [No Lock]        │
│   - Swap()     [No Lock]        │
└─────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────┐
│   Data Structure (数据层)       │
│   - tasks []Task                │
└─────────────────────────────────┘
```

### 3. 线程安全的两种方法

#### 方法A：Len() - heap.Interface 方法
```go
func (tq *TaskQueue) Len() int {
    // 不加锁，只在被 heap 函数调用时使用
    return len(tq.tasks)
}
```

#### 方法B：Size() - 公共方法
```go
func (tq *TaskQueue) Size() int {
    // 加锁，提供线程安全的访问
    tq.mu.RLock()
    defer tq.mu.RUnlock()
    return len(tq.tasks)
}
```

**区别**：
- `Len()` - 内部使用，假设调用者已持有锁
- `Size()` - 外部使用，需要自己加锁保护

## 🎯 使用指南

### ✅ 正确使用

```go
// 创建队列
queue := scheduler.NewTaskQueue()

// 公共方法自动加锁，安全使用
queue.Enqueue(task1)
queue.Enqueue(task2)

size := queue.Size()    // 线程安全
task := queue.Dequeue() // 线程安全
peek := queue.Peek()    // 线程安全
```

### ❌ 错误使用

```go
// ❌ 不要直接调用 heap 函数
heap.Push(queue, task)  // 没有加锁保护

// ❌ 不要在持有锁时再次加锁
queue.mu.Lock()
queue.Enqueue(task)     // Enqueue 内部也会加锁 → 死锁
queue.mu.Unlock()

// ❌ 不要直接访问 tasks 字段
for _, task := range queue.tasks {  // 没有锁保护
    // ...
}
```

### ✅ 正确的并发访问

```go
// 使用公共方法，它们已经包含了锁保护
tasks := queue.List()   // 返回副本，线程安全
for _, task := range tasks {
    // 安全处理
}
```

## 🔬 测试验证

### 死锁检测

Go 运行时会自动检测死锁：

```bash
go run main.go
# 死锁会被检测并报告：
# fatal error: all goroutines are asleep - deadlock!
```

### 并发测试

```go
func TestConcurrentEnqueueDequeue(t *testing.T) {
    queue := scheduler.NewTaskQueue()
    var wg sync.WaitGroup

    // 100个goroutine并发入队
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            task := &agent.Task{
                ID:       fmt.Sprintf("task-%d", id),
                Priority: id % 10,
            }
            queue.Enqueue(task)
        }(i)
    }

    // 100个goroutine并发出队
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            queue.Dequeue()
        }()
    }

    wg.Wait()
    // 如果没有死锁或数据竞争，测试通过
}
```

### 数据竞争检测

```bash
# 使用 race detector 检测数据竞争
go test -race ./internal/scheduler
```

## 📖 相关知识

### Go 的 Mutex 特性

1. **非重入** - 同一个 goroutine 不能重复获取同一个锁
2. **公平性** - FIFO 顺序获取锁
3. **不可复制** - Mutex 包含状态，不能被复制

### Java ReentrantLock 对比

```java
// Java - 支持重入
ReentrantLock lock = new ReentrantLock();
lock.lock();
lock.lock();  // ✓ 可以重入
lock.unlock();
lock.unlock();
```

```go
// Go - 不支持重入
var mu sync.Mutex
mu.Lock()
mu.Lock()  // ❌ 死锁！
```

### container/heap 的设计

`container/heap` 使用接口模式，将：
- **数据存储** - 由用户实现
- **堆算法** - 由标准库提供
- **线程安全** - 由用户负责

这种设计：
- ✅ 灵活性高
- ✅ 性能好（避免不必要的锁）
- ⚠️ 需要用户正确处理并发

## 🎓 经验总结

1. **明确锁的边界** - 哪些方法需要加锁，哪些不需要
2. **避免重入锁** - Go 的 Mutex 不支持重入
3. **最小锁粒度** - 只在必要时持有锁
4. **文档化假设** - 注释说明线程安全性
5. **测试并发性** - 使用 `-race` 检测问题

## 🔗 参考资料

- [Go Mutex 文档](https://pkg.go.dev/sync#Mutex)
- [container/heap 文档](https://pkg.go.dev/container/heap)
- [Go 并发模式](https://go.dev/blog/pipelines)

---

**修复日期**: 2026-01-28
**问题级别**: Critical（导致程序死锁）
**影响范围**: TaskQueue 的所有使用场景
**修复状态**: ✅ 已修复并验证
