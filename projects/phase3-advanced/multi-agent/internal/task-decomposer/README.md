# Task Decomposer

> 任务分解器 - 将复杂任务分解为可执行的子任务

## 📦 功能特性

- **多种分解策略**: 支持依赖、优先级、能力和混合四种分解策略
- **复杂度分析**: 自动分析任务复杂度，推荐合适的分解策略
- **依赖图管理**: 自动构建任务依赖图，检测循环依赖
- **拓扑排序**: 计算任务执行顺序和并行执行组
- **规则系统**: 支持自定义分解规则，内置代码审查和文档处理规则
- **子任务生成**: 支持顺序、并行、流水线三种生成模式

## 🚀 快速开始

### 创建任务并分解

```go
import "github.com/agent-learning/multi-agent/internal/task-decomposer"

// 创建任务
task := decomposer.NewTask("task-001", "code_review", "Review PR #123")
task.Priority = 5
task.AddCapability("syntax_analysis")
task.AddCapability("quality_analysis")
task.AddCapability("security_analysis")

// 创建分解器
config := decomposer.DefaultConfig()
d := decomposer.NewDecomposer(config)

// 分解任务
result, err := d.Decompose(task)
if err != nil {
    log.Fatalf("Decomposition failed: %v", err)
}

// 查看结果
fmt.Printf("Generated %d sub-tasks\n", len(result.SubTasks))
fmt.Printf("Strategy used: %s\n", result.Strategy)
fmt.Printf("Max level: %d\n", result.Metadata["max_level"])
```

## 📚 核心概念

### 1. 任务复杂度

任务复杂度分为4个等级：

| 复杂度 | 说明 | 分数范围 |
|--------|------|----------|
| Simple | 简单任务，不需要分解 | < 2.0 |
| Moderate | 中等复杂度 | 2.0 - 4.9 |
| Complex | 复杂任务 | 5.0 - 7.9 |
| VeryComplex | 非常复杂 | ≥ 8.0 |

复杂度计算考虑因素：
- 依赖数量（权重0.3）
- 所需能力数量（权重0.3）
- 要求数量（权重0.2）
- 任务类型（权重0.2）

```go
analyzer := decomposer.NewComplexityAnalyzer()
complexity := analyzer.Analyze(task)

if complexity >= decomposer.ComplexityComplex {
    fmt.Println("This is a complex task, decomposition recommended")
}
```

### 2. 分解策略

#### 2.1 基于依赖的分解 (DEPENDENCY)

适用于有明确依赖关系的任务。

```go
config := &decomposer.DecomposerConfig{
    Strategy: decomposer.StrategyDependency,
    MaxDepth: 3,
}

d := decomposer.NewDecomposer(config)
result, _ := d.Decompose(task)
```

**特点**：
- 为每个依赖创建子任务
- 创建主任务依赖于所有依赖子任务
- 自动构建依赖图

**示例**：
```
任务A (依赖: B, C)
  ├── 子任务1: 处理依赖B (Level 0)
  ├── 子任务2: 处理依赖C (Level 0)
  └── 子任务3: 执行主任务 (Level 1, 依赖: 1, 2)
```

#### 2.2 基于优先级的分解 (PRIORITY)

按执行阶段分解任务。

```go
config := &decomposer.DecomposerConfig{
    Strategy: decomposer.StrategyPriority,
}

d := decomposer.NewDecomposer(config)
result, _ := d.Decompose(task)
```

**特点**：
- 分为准备、执行、验证三个阶段
- 按阶段顺序执行
- 优先级递增

**示例**：
```
任务A
  ├── 准备阶段 (Priority: 4, Level 0)
  ├── 执行阶段 (Priority: 5, Level 1)
  └── 验证阶段 (Priority: 6, Level 2)
```

#### 2.3 基于能力的分解 (CAPABILITY)

根据所需Agent能力分解任务。

```go
task := decomposer.NewTask("task-001", "analysis", "Analyze code")
task.AddCapability("syntax_check")
task.AddCapability("quality_check")
task.AddCapability("security_check")

config := &decomposer.DecomposerConfig{
    Strategy: decomposer.StrategyCapability,
}

d := decomposer.NewDecomposer(config)
result, _ := d.Decompose(task)
```

**特点**：
- 为每个能力创建独立子任务
- 子任务可并行执行
- 创建聚合任务收集结果

**示例**：
```
任务A (能力: syntax, quality, security)
  ├── 子任务1: syntax_check (Level 0, 可并行)
  ├── 子任务2: quality_check (Level 0, 可并行)
  ├── 子任务3: security_check (Level 0, 可并行)
  └── 子任务4: aggregate (Level 1, 依赖: 1,2,3)
```

#### 2.4 混合策略 (HYBRID)

综合多种策略，根据任务特点自动选择。

```go
config := &decomposer.DecomposerConfig{
    Strategy:           decomposer.StrategyHybrid,
    ComplexityAnalysis: true,
}

d := decomposer.NewDecomposer(config)
result, _ := d.Decompose(task)
```

**特点**：
- 自动分析任务复杂度
- 应用预定义规则
- 根据复杂度选择策略

**决策逻辑**：
```
VeryComplex → Dependency分解
Complex → Capability分解
Moderate → Priority分解
Simple → 不分解
```

### 3. 自定义分解规则

```go
// 创建自定义规则
rule := &decomposer.DecompositionRule{
    Name:     "api_test",
    TaskType: "api_test",
    Condition: func(t *decomposer.Task) bool {
        return t.Type == "api_test"
    },
    Decompose: func(t *decomposer.Task) ([]*decomposer.SubTask, error) {
        return []*decomposer.SubTask{
            {
                ID:          t.ID + "-prepare",
                Type:        "prepare_test_data",
                Description: "Prepare test data",
                Priority:    t.Priority,
                Level:       0,
            },
            {
                ID:           t.ID + "-execute",
                Type:         "execute_tests",
                Description:  "Execute API tests",
                Priority:     t.Priority,
                Dependencies: []string{t.ID + "-prepare"},
                Level:        1,
            },
            {
                ID:           t.ID + "-verify",
                Type:         "verify_results",
                Description:  "Verify test results",
                Priority:     t.Priority,
                Dependencies: []string{t.ID + "-execute"},
                Level:        2,
            },
        }, nil
    },
    Priority: 10,
}

// 注册规则
d.RegisterRule(rule)
```

### 4. 依赖图操作

```go
// 获取依赖图
graph := result.Graph

// 检查循环依赖
if graph.HasCycle() {
    log.Fatal("Circular dependency detected!")
}

// 拓扑排序
sorted, err := graph.TopologicalSort()
if err != nil {
    log.Fatalf("Sort failed: %v", err)
}

fmt.Println("Execution order:", sorted)

// 获取并行执行组
parallelGroups := graph.GetParallelTasks()
for level, tasks := range parallelGroups {
    fmt.Printf("Level %d: %v (can run in parallel)\n", level, tasks)
}
```

### 5. 复杂度分析报告

```go
analyzer := decomposer.NewComplexityAnalyzer()

// 生成详细报告
report := analyzer.GenerateReport(task)

fmt.Printf("Task: %s\n", report.TaskID)
fmt.Printf("Complexity: %v\n", report.Complexity)
fmt.Printf("Score: %.2f\n", report.Score)
fmt.Printf("Recommended Strategy: %s\n", report.RecommendedStrategy)
fmt.Printf("Estimated Sub-tasks: %d\n", report.EstimatedSubTasks)

// 查看影响因素
fmt.Println("\nFactors:")
for factor, value := range report.Factors {
    fmt.Printf("  %s: %.2f\n", factor, value)
}

// 查看建议
fmt.Println("\nRecommendations:")
for _, rec := range report.Recommendations {
    fmt.Printf("  - %s\n", rec)
}
```

## 🎯 使用场景

### 场景1: 代码审查任务

```go
task := decomposer.NewTask("review-001", "code_review", "Review PR #123")
task.AddCapability("syntax_analysis")
task.AddCapability("quality_analysis")
task.AddCapability("security_analysis")

d := decomposer.NewDecomposer(nil) // 使用默认配置
result, _ := d.Decompose(task)

// 自动分解为：
// 1. syntax_check (Level 0)
// 2. quality_check (Level 1, 依赖syntax_check)
// 3. security_check (Level 1, 依赖syntax_check)
```

### 场景2: 文档处理任务

```go
task := decomposer.NewTask("doc-001", "document_processing", "Process contract.pdf")

d := decomposer.NewDecomposer(nil)
result, _ := d.Decompose(task)

// 自动分解为：
// 1. parse (Level 0)
// 2. analyze (Level 1, 依赖parse)
// 3. summarize (Level 2, 依赖analyze)
```

### 场景3: 数据分析任务

```go
task := decomposer.NewTask("analysis-001", "data_analysis", "Analyze sales data")
task.AddCapability("data_cleaning")
task.AddCapability("statistical_analysis")
task.AddCapability("visualization")
task.Priority = 7

config := &decomposer.DecomposerConfig{
    Strategy:          decomposer.StrategyCapability,
    ComplexityAnalysis: true,
}

d := decomposer.NewDecomposer(config)
result, _ := d.Decompose(task)

// 并行执行数据清洗、统计分析、可视化
// 最后聚合结果
```

## 📊 子任务生成模式

### 顺序模式 (Sequential)

```go
generator := decomposer.NewSubTaskGenerator()
subTasks, _ := generator.GenerateWithPattern(task, "sequential")

// 生成: prepare → execute → verify
```

### 并行模式 (Parallel)

```go
generator := decomposer.NewSubTaskGenerator()
subTasks, _ := generator.GenerateWithPattern(task, "parallel")

// 生成: task1, task2, task3 (同一层级)
```

### 流水线模式 (Pipeline)

```go
generator := decomposer.NewSubTaskGenerator()
subTasks, _ := generator.GenerateWithPattern(task, "pipeline")

// 生成: input → process → output
```

## 🔧 配置选项

```go
config := &decomposer.DecomposerConfig{
    Strategy:           decomposer.StrategyHybrid,
    MaxDepth:           3,    // 最大分解深度
    MinSubTasks:        2,    // 最小子任务数
    MaxSubTasks:        10,   // 最大子任务数
    ParallelThreshold:  3,    // 并行阈值
    ComplexityAnalysis: true, // 启用复杂度分析
}

d := decomposer.NewDecomposer(config)
```

## 📝 最佳实践

### 1. 始终检查错误

```go
result, err := d.Decompose(task)
if err != nil {
    log.Fatalf("Decomposition failed: %v", err)
}
```

### 2. 验证任务有效性

```go
if task.ID == "" || task.Type == "" {
    log.Fatal("Invalid task: ID and Type are required")
}
```

### 3. 检查循环依赖

```go
if result.Graph.HasCycle() {
    log.Fatal("Circular dependency detected!")
}
```

### 4. 利用复杂度分析

```go
if task.GetComplexity() == decomposer.ComplexitySimple {
    // 简单任务，直接执行
} else {
    // 复杂任务，进行分解
}
```

### 5. 合理设置优先级

```go
task.Priority = 5  // 普通任务
// 或
task.Priority = 9  // 紧急任务
```

## 🧪 测试

```bash
go test ./internal/task-decomposer -v
```

## 📖 相关文档

- [Multi-Agent Protocol](../../protocol/README.md)
- [Architecture Documentation](../../../../docs/architecture/multi-agent-protocol.md)

---

**版本**: 1.0.0
**许可证**: MIT
