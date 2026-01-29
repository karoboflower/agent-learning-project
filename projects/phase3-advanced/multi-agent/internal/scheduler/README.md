# Task Scheduler

> 多Agent任务调度器 - Agent注册、任务分配和队列管理

## 📦 功能特性

- **Agent注册管理**: 注册、注销、能力描述、状态管理
- **多种分配策略**: 基于能力、负载均衡、优先级、轮询
- **优先级队列**: 自动按优先级排序任务
- **任务生命周期管理**: 提交、分配、执行、完成/失败/取消
- **心跳监控**: 自动检测离线Agent
- **并发安全**: 所有操作线程安全
- **统计信息**: 实时统计Agent和任务状态

## 🚀 快速开始

### 创建调度器

```go
import "github.com/agent-learning/multi-agent/internal/scheduler"

// 使用默认配置
config := scheduler.DefaultSchedulerConfig()
s := scheduler.NewScheduler(config)

// 启动调度器
if err := s.Start(); err != nil {
    log.Fatalf("Failed to start scheduler: %v", err)
}

// 程序退出时停止
defer s.Stop()
```

### 注册Agent

```go
agent := &scheduler.Agent{
    ID:           "agent-001",
    Name:         "Code Review Agent",
    Capabilities: []string{"code_review", "syntax_check"},
    MaxTasks:     10,
}

if err := s.RegisterAgent(agent); err != nil {
    log.Fatalf("Failed to register agent: %v", err)
}
```

### 提交任务

```go
task := &scheduler.Task{
    ID:                   "task-001",
    Type:                 "code_review",
    Priority:             8,  // 1-10, 10最高
    RequiredCapabilities: []string{"code_review"},
    Metadata: map[string]interface{}{
        "pr_number": 123,
        "repo":      "my-repo",
    },
}

if err := s.SubmitTask(task); err != nil {
    log.Fatalf("Failed to submit task: %v", err)
}
```

### 任务生命周期

```go
// 任务会自动分配给合适的Agent
// 也可以手动触发分配
agentID, err := s.AssignTask("task-001")

// Agent完成任务后
if err := s.CompleteTask("task-001"); err != nil {
    log.Printf("Failed to complete task: %v", err)
}

// 或标记失败
if err := s.FailTask("task-001"); err != nil {
    log.Printf("Failed to mark task as failed: %v", err)
}

// 或取消任务
if err := s.CancelTask("task-001"); err != nil {
    log.Printf("Failed to cancel task: %v", err)
}
```

## 📚 核心概念

### 1. Agent状态

| 状态 | 说明 |
|------|------|
| IDLE | 空闲，可接受任务 |
| BUSY | 忙碌，已达最大任务数 |
| OFFLINE | 离线，心跳超时 |
| MAINTENANCE | 维护中，不接受任务 |

```go
// 更新Agent状态
s.UpdateAgentStatus("agent-001", scheduler.AgentStatusMaintenance)

// 更新心跳
s.UpdateAgentHeartbeat("agent-001")
```

### 2. 任务状态

| 状态 | 说明 |
|------|------|
| PENDING | 等待分配 |
| ASSIGNED | 已分配给Agent |
| RUNNING | 执行中 |
| COMPLETED | 已完成 |
| FAILED | 失败 |
| CANCELLED | 已取消 |

### 3. 分配策略

#### 基于能力 (CAPABILITY)

根据任务所需能力匹配Agent，在匹配的Agent中选择负载最低的。

```go
config := &scheduler.SchedulerConfig{
    AllocationStrategy: scheduler.StrategyCapability,
}
s := scheduler.NewScheduler(config)
```

**特点**:
- 确保Agent具有所需能力
- 在候选Agent中负载均衡
- 适合能力差异大的场景

#### 负载均衡 (LOAD_BALANCE)

选择负载最低的Agent分配任务。

```go
config := &scheduler.SchedulerConfig{
    AllocationStrategy: scheduler.StrategyLoadBalance,
}
s := scheduler.NewScheduler(config)
```

**负载计算**:
```go
load = (显式负载 + 任务数负载) / 2
任务数负载 = 当前任务数 / 最大任务数
```

**特点**:
- 平衡Agent工作负载
- 避免单个Agent过载
- 适合Agent能力相近的场景

#### 基于优先级 (PRIORITY)

根据任务优先级选择Agent：
- 高优先级(≥8)：分配给负载最低的Agent
- 中低优先级：分配给负载适中的Agent

```go
config := &scheduler.SchedulerConfig{
    AllocationStrategy: scheduler.StrategyPriority,
}
s := scheduler.NewScheduler(config)
```

**特点**:
- 优先保证高优先级任务
- 避免所有任务都集中在最空闲的Agent
- 适合有明确优先级的场景

#### 轮询 (ROUND_ROBIN)

按顺序轮流分配给各个Agent。

```go
config := &scheduler.SchedulerConfig{
    AllocationStrategy: scheduler.StrategyRoundRobin,
}
s := scheduler.NewScheduler(config)
```

**特点**:
- 简单公平
- 不考虑负载差异
- 适合任务时长相近的场景

### 4. 任务队列

内置优先级队列，自动按优先级排序。

```go
// 查看队列大小
queueSize := s.GetQueueSize()

// 查看待处理任务
pendingTasks := s.ListTasksByStatus(scheduler.TaskStatusPending)
```

**特点**:
- 自动优先级排序
- 支持最大容量限制
- 并发安全
- 支持阻塞等待

### 5. 心跳机制

定期检查Agent心跳，自动标记超时Agent为离线。

```go
config := &scheduler.SchedulerConfig{
    HeartbeatInterval: 30 * time.Second,  // 检查间隔
    HeartbeatTimeout:  90 * time.Second,  // 超时时间
}
```

Agent需要定期发送心跳：
```go
// Agent每30秒调用一次
s.UpdateAgentHeartbeat("agent-001")
```

## 🎯 使用场景

### 场景1: 代码审查系统

```go
// 注册不同能力的Agent
agents := []*scheduler.Agent{
    {
        ID:           "reviewer-1",
        Name:         "Syntax Reviewer",
        Capabilities: []string{"syntax_check"},
        MaxTasks:     5,
    },
    {
        ID:           "reviewer-2",
        Name:         "Security Reviewer",
        Capabilities: []string{"security_check"},
        MaxTasks:     3,
    },
    {
        ID:           "reviewer-3",
        Name:         "Full Reviewer",
        Capabilities: []string{"syntax_check", "security_check", "quality_check"},
        MaxTasks:     8,
    },
}

for _, agent := range agents {
    s.RegisterAgent(agent)
}

// 提交不同类型的审查任务
syntaxTask := &scheduler.Task{
    ID:                   "review-001",
    Type:                 "code_review",
    Priority:             5,
    RequiredCapabilities: []string{"syntax_check"},
}
s.SubmitTask(syntaxTask)

securityTask := &scheduler.Task{
    ID:                   "review-002",
    Type:                 "security_audit",
    Priority:             9,  // 高优先级
    RequiredCapabilities: []string{"security_check"},
}
s.SubmitTask(securityTask)
```

### 场景2: 数据处理管道

```go
// 使用负载均衡策略
config := &scheduler.SchedulerConfig{
    AllocationStrategy: scheduler.StrategyLoadBalance,
    MaxQueueSize:       1000,
    WorkerCount:        10,
}
s := scheduler.NewScheduler(config)
s.Start()

// 注册处理Agent
for i := 0; i < 20; i++ {
    agent := &scheduler.Agent{
        ID:           fmt.Sprintf("processor-%d", i),
        Name:         fmt.Sprintf("Data Processor %d", i),
        Capabilities: []string{"data_processing"},
        MaxTasks:     5,
    }
    s.RegisterAgent(agent)
}

// 批量提交数据处理任务
for i := 0; i < 1000; i++ {
    task := &scheduler.Task{
        ID:       fmt.Sprintf("process-%d", i),
        Type:     "data_process",
        Priority: 5,
        Metadata: map[string]interface{}{
            "data_id": i,
        },
    }
    s.SubmitTask(task)
}
```

### 场景3: 紧急任务处理

```go
// 使用优先级策略
config := &scheduler.SchedulerConfig{
    AllocationStrategy: scheduler.StrategyPriority,
}
s := scheduler.NewScheduler(config)
s.Start()

// 普通任务
normalTask := &scheduler.Task{
    ID:       "normal-001",
    Type:     "report",
    Priority: 5,
}
s.SubmitTask(normalTask)

// 紧急任务 - 会优先分配给负载最低的Agent
urgentTask := &scheduler.Task{
    ID:       "urgent-001",
    Type:     "alert",
    Priority: 10,
}
s.SubmitTask(urgentTask)
```

## 🔧 配置选项

```go
type SchedulerConfig struct {
    MaxQueueSize       int                // 最大队列大小 (默认: 1000)
    AllocationStrategy AllocationStrategy // 分配策略 (默认: LOAD_BALANCE)
    HeartbeatInterval  time.Duration      // 心跳检查间隔 (默认: 30s)
    HeartbeatTimeout   time.Duration      // 心跳超时时间 (默认: 90s)
    WorkerCount        int                // 工作协程数 (默认: 5)
}

// 自定义配置
config := &scheduler.SchedulerConfig{
    MaxQueueSize:       5000,
    AllocationStrategy: scheduler.StrategyCapability,
    HeartbeatInterval:  15 * time.Second,
    HeartbeatTimeout:   45 * time.Second,
    WorkerCount:        10,
}
```

## 📊 统计信息

```go
stats := s.GetStatistics()

fmt.Printf("Total Agents: %d\n", stats.TotalAgents)
fmt.Printf("Idle Agents: %d\n", stats.AgentsByStatus[scheduler.AgentStatusIdle])
fmt.Printf("Busy Agents: %d\n", stats.AgentsByStatus[scheduler.AgentStatusBusy])

fmt.Printf("Total Tasks: %d\n", stats.TotalTasks)
fmt.Printf("Pending: %d\n", stats.TasksByStatus["PENDING"])
fmt.Printf("Running: %d\n", stats.TasksByStatus["RUNNING"])
fmt.Printf("Completed: %d\n", stats.TasksByStatus["COMPLETED"])

fmt.Printf("Queue Size: %d\n", stats.QueueSize)
fmt.Printf("Strategy: %s\n", stats.AllocationStrategy)
```

## 📝 最佳实践

### 1. 合理设置MaxTasks

```go
agent := &scheduler.Agent{
    ID:       "agent-001",
    MaxTasks: 10,  // 根据Agent处理能力设置
}
```

建议：
- CPU密集型任务：MaxTasks = CPU核心数 × 2
- I/O密集型任务：MaxTasks = 20-50
- 混合型任务：MaxTasks = 10-20

### 2. 及时更新心跳

```go
// Agent应每30秒更新一次心跳
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for range ticker.C {
    if err := s.UpdateAgentHeartbeat(agentID); err != nil {
        log.Printf("Failed to update heartbeat: %v", err)
    }
}
```

### 3. 处理任务失败

```go
// Agent执行任务
task, _ := getNextTask()

if err := executeTask(task); err != nil {
    // 标记失败
    s.FailTask(task.ID)

    // 可选：重新提交任务
    task.Metadata["retry_count"] = retryCount + 1
    if retryCount < 3 {
        s.SubmitTask(task)
    }
} else {
    // 标记完成
    s.CompleteTask(task.ID)
}
```

### 4. 监控队列大小

```go
// 定期检查队列
if s.GetQueueSize() > 800 {  // 80%容量
    log.Warn("Queue is almost full, consider adding more agents")
}
```

### 5. 动态调整策略

```go
// 根据系统负载动态调整策略
stats := s.GetStatistics()
avgLoad := calculateAverageLoad(stats)

if avgLoad > 0.8 {
    // 高负载时使用负载均衡
    s.SetAllocationStrategy(scheduler.StrategyLoadBalance)
} else {
    // 低负载时使用能力匹配
    s.SetAllocationStrategy(scheduler.StrategyCapability)
}
```

### 6. 优雅关闭

```go
// 捕获信号
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

<-sigChan

// 停止接受新任务
// 等待当前任务完成
// 关闭调度器
if err := s.Stop(); err != nil {
    log.Printf("Failed to stop scheduler: %v", err)
}
```

## 🧪 测试

```bash
cd projects/phase3-advanced/multi-agent/internal/scheduler
go test -v
```

## 📖 API文档

### Agent管理

- `RegisterAgent(agent *Agent) error` - 注册Agent
- `UnregisterAgent(agentID string) error` - 注销Agent
- `GetAgent(agentID string) (*Agent, error)` - 获取Agent信息
- `ListAgents() []*Agent` - 列出所有Agent
- `UpdateAgentStatus(agentID string, status AgentStatus) error` - 更新Agent状态
- `UpdateAgentHeartbeat(agentID string) error` - 更新心跳

### 任务管理

- `SubmitTask(task *Task) error` - 提交任务
- `AssignTask(taskID string) (string, error)` - 手动分配任务
- `CompleteTask(taskID string) error` - 完成任务
- `FailTask(taskID string) error` - 标记失败
- `CancelTask(taskID string) error` - 取消任务
- `GetTask(taskID string) (*Task, error)` - 获取任务信息
- `ListTasks() []*Task` - 列出所有任务
- `ListTasksByStatus(status TaskStatus) []*Task` - 按状态列出任务
- `GetAgentTasks(agentID string) []*Task` - 获取Agent的任务

### 调度器

- `Start() error` - 启动调度器
- `Stop() error` - 停止调度器
- `GetStatistics() *Statistics` - 获取统计信息
- `SetAllocationStrategy(strategy AllocationStrategy)` - 设置分配策略
- `GetAllocationStrategy() AllocationStrategy` - 获取当前策略

## 🔗 相关模块

- [Task Decomposer](../task-decomposer/README.md) - 任务分解器
- [Protocol](../../protocol/README.md) - 通信协议

---

**版本**: 1.0.0
**许可证**: MIT
