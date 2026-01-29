# Aggregator Module

> 结果聚合器 - Agent任务结果收集、验证、合并和冲突解决

## 📦 功能特性

- **结果收集**: 接收和存储Agent执行结果
- **结果验证**: 多规则验证系统
- **结果存储**: 高效的结果存储和检索
- **结果合并**: 6种合并策略
- **冲突检测**: 自动检测字段冲突
- **冲突解决**: 4种冲突解决策略
- **置信度计算**: 智能置信度评估
- **并发安全**: 所有操作线程安全

## 🚀 快速开始

### 创建聚合器

```go
import "github.com/agent-learning/multi-agent/internal/aggregator"

// 创建聚合器
aggregator := aggregator.NewResultAggregator(
    aggregator.MergeStrategyVoting,           // 合并策略: 投票法
    aggregator.ConflictResolutionVoting,      // 冲突解决: 投票解决
)

// 配置最少结果数
aggregator.GetMerger().SetMinResults(2)

// 配置置信度阈值
aggregator.GetMerger().SetConfidenceThreshold(0.7)
```

### 添加结果

```go
result := &aggregator.TaskResult{
    ID:      "result-001",
    TaskID:  "task-001",
    AgentID: "agent-001",
    Data: map[string]interface{}{
        "answer": "A",
        "confidence": 95.5,
    },
    Score:     85,
    CreatedAt: time.Now(),
}

// 添加结果（自动验证）
if err := aggregator.AddResult(result); err != nil {
    log.Printf("Failed to add result: %v", err)
}
```

### 聚合结果

```go
// 聚合任务的所有结果
aggregated, err := aggregator.AggregateTask("task-001")
if err != nil {
    log.Fatalf("Aggregation failed: %v", err)
}

// 查看合并后的数据
fmt.Printf("Merged data: %+v\n", aggregated.MergedData)

// 查看冲突
for _, conflict := range aggregated.Conflicts {
    fmt.Printf("Conflict in field '%s': %v\n", conflict.Field, conflict.Values)
    fmt.Printf("Resolution: %s\n", conflict.Resolution)
}

// 查看置信度
fmt.Printf("Confidence: %.2f\n", aggregated.Confidence)
```

## 📚 核心概念

### 1. 结果存储

#### ResultStore

存储和管理所有Agent执行结果：

```go
store := aggregator.NewResultStore()

// 添加结果
store.AddResult(result)

// 获取结果
result, err := store.GetResult("result-001")

// 获取任务的所有结果
results := store.GetResultsByTask("task-001")

// 按状态查询
validatedResults := store.GetResultsByStatus(aggregator.ResultStatusValidated)

// 更新结果
store.UpdateResult(result)

// 删除结果
store.DeleteResult("result-001")

// 统计
count := store.GetResultCount()
countByTask := store.GetResultCountByTask("task-001")
```

**结果状态**:
- `PENDING`: 待处理
- `VALIDATED`: 已验证
- `REJECTED`: 已拒绝
- `MERGED`: 已合并

### 2. 结果验证

#### ResultValidator

多规则验证系统：

```go
validator := aggregator.NewResultValidator()

// 添加必需字段规则
validator.AddRule(&aggregator.RequiredFieldsRule{
    Fields: []string{"answer", "confidence"},
})

// 添加数据类型规则
validator.AddRule(&aggregator.DataTypeRule{
    Field:        "answer",
    ExpectedType: "string",
})

validator.AddRule(&aggregator.DataTypeRule{
    Field:        "confidence",
    ExpectedType: "number",
})

// 添加分数范围规则
validator.AddRule(&aggregator.ScoreRangeRule{
    MinScore: 0,
    MaxScore: 100,
})

// 验证并标记结果
if err := validator.ValidateAndMark(result); err != nil {
    log.Printf("Validation failed: %v", err)
    // 结果被标记为 REJECTED
} else {
    // 结果被标记为 VALIDATED
}

// 批量验证
errors := validator.ValidateMultiple(results)
for resultID, err := range errors {
    log.Printf("Result %s failed: %v", resultID, err)
}
```

**内置验证规则**:

1. **RequiredFieldsRule** - 必需字段验证
2. **DataTypeRule** - 数据类型验证（string, number, boolean, object, array）
3. **ScoreRangeRule** - 分数范围验证

**自定义验证规则**:

```go
type CustomRule struct {}

func (r *CustomRule) Name() string {
    return "CustomRule"
}

func (r *CustomRule) Validate(result *aggregator.TaskResult) error {
    // 自定义验证逻辑
    if someCondition {
        return fmt.Errorf("validation failed")
    }
    return nil
}

validator.AddRule(&CustomRule{})
```

### 3. 结果合并

#### ResultMerger

支持6种合并策略：

**1. 投票法 (VOTING)**

选择出现次数最多的值：

```go
// Agent-001: answer="A"
// Agent-002: answer="A"
// Agent-003: answer="B"
// 结果: answer="A" (2票 vs 1票)
```

**2. 平均法 (AVERAGING)**

对数值类型求平均：

```go
// Agent-001: value=10
// Agent-002: value=20
// Agent-003: value=30
// 结果: value=20.0 (平均值)
```

**3. 加权法 (WEIGHTED)**

使用结果分数作为权重：

```go
// Agent-001: value=10, score=50
// Agent-002: value=20, score=100
// 结果: value=16.67 (加权平均)
```

**4. 一致性法 (CONSENSUS)**

只保留所有Agent一致的字段：

```go
// Agent-001: {agreed="yes", disagreed="A"}
// Agent-002: {agreed="yes", disagreed="B"}
// 结果: {agreed="yes"} (只保留一致的)
```

**5. 优先级法 (PRIORITY)**

使用分数最高的结果：

```go
// Agent-001: answer="A", score=80
// Agent-002: answer="B", score=95  <- 选择这个
// Agent-003: answer="C", score=70
// 结果: answer="B"
```

**6. 最高分法 (HIGHEST_SCORE)**

与优先级法相同，使用最高分结果。

**示例**:

```go
merger := aggregator.NewResultMerger(
    aggregator.MergeStrategyVoting,
    aggregator.ConflictResolutionVoting,
)

merger.SetMinResults(2)                    // 最少2个结果
merger.SetConfidenceThreshold(0.7)         // 置信度阈值70%

aggregated, err := merger.Merge("task-001", results)
```

### 4. 冲突检测与解决

#### 冲突检测

自动检测字段值不一致：

```go
// Agent-001: answer="A"
// Agent-002: answer="B"
// 检测到冲突: field="answer", values=["A", "B"]
```

#### 冲突解决策略

**1. 投票解决 (VOTING)**

选择出现次数最多的值：

```go
// Agent-001: answer="A"
// Agent-002: answer="A"
// Agent-003: answer="B"
// 解决: answer="A" (2票获胜)
```

**2. 多数解决 (MAJORITY)**

与投票解决相同。

**3. 高分解决 (HIGH_SCORE)**

选择分数最高的Agent的值：

```go
// Agent-001: answer="A", score=80
// Agent-002: answer="B", score=95  <- 选择这个
// 解决: answer="B" (最高分)
```

**4. 手动解决 (MANUAL)**

标记为需要人工介入：

```go
conflict.Resolution = "Manual resolution required"
```

**查看冲突**:

```go
for _, conflict := range aggregated.Conflicts {
    fmt.Printf("Field: %s\n", conflict.Field)
    fmt.Printf("Values: %v\n", conflict.Values)
    fmt.Printf("Agents: %v\n", conflict.AgentIDs)
    fmt.Printf("Resolution: %s\n", conflict.Resolution)
    fmt.Printf("Resolved at: %v\n", conflict.ResolvedAt)
}
```

### 5. 置信度计算

综合多个因素计算置信度：

**因素1: 结果数量** (30%权重)
- 越多结果，置信度越高
- 相对于minResults的比例

**因素2: 平均分数** (40%权重)
- 所有���果的平均分数
- 分数越高，置信度越高

**因素3: 冲突数量** (30%权重)
- 冲突越少，置信度越高
- 冲突数 / 字段数 的反比

**最终置信度**: 加权平均，范围 [0.0, 1.0]

```go
confidence := aggregated.Confidence
if confidence >= 0.9 {
    fmt.Println("Very high confidence")
} else if confidence >= 0.7 {
    fmt.Println("High confidence")
} else if confidence >= 0.5 {
    fmt.Println("Medium confidence")
} else {
    fmt.Println("Low confidence")
}
```

## 🎯 使用场景

### 场景1: 代码审查任务

多个Agent审查同一段代码：

```go
aggregator := aggregator.NewResultAggregator(
    aggregator.MergeStrategyVoting,
    aggregator.ConflictResolutionVoting,
)

// 配置验证规则
validator := aggregator.GetValidator()
validator.AddRule(&aggregator.RequiredFieldsRule{
    Fields: []string{"issues_found", "severity", "recommendation"},
})

// Agent-001的结果
result1 := &aggregator.TaskResult{
    ID:      "result-001",
    TaskID:  "code-review-123",
    AgentID: "agent-001",
    Data: map[string]interface{}{
        "issues_found":   3,
        "severity":       "medium",
        "recommendation": "refactor",
    },
    Score: 85,
}

// Agent-002的结果
result2 := &aggregator.TaskResult{
    ID:      "result-002",
    TaskID:  "code-review-123",
    AgentID: "agent-002",
    Data: map[string]interface{}{
        "issues_found":   2,
        "severity":       "medium",
        "recommendation": "refactor",
    },
    Score: 90,
}

// Agent-003的结果
result3 := &aggregator.TaskResult{
    ID:      "result-003",
    TaskID:  "code-review-123",
    AgentID: "agent-003",
    Data: map[string]interface{}{
        "issues_found":   3,
        "severity":       "low",
        "recommendation": "refactor",
    },
    Score: 80,
}

// 添加结果
aggregator.AddResult(result1)
aggregator.AddResult(result2)
aggregator.AddResult(result3)

// 聚合
aggregated, _ := aggregator.AggregateTask("code-review-123")

// 结果:
// issues_found: 3 (投票: 2 vs 1)
// severity: "medium" (投票: 2 vs 1)
// recommendation: "refactor" (一致)
// 冲突: issues_found, severity
// 置信度: ~0.75
```

### 场景2: 数据分析任务

多个Agent分析同一数据集：

```go
aggregator := aggregator.NewResultAggregator(
    aggregator.MergeStrategyAveraging,  // 使用平均法
    aggregator.ConflictResolutionHighScore,
)

// Agent-001的分析
result1 := &aggregator.TaskResult{
    ID:      "result-001",
    TaskID:  "data-analysis-456",
    AgentID: "agent-001",
    Data: map[string]interface{}{
        "mean":     10.5,
        "median":   10.0,
        "std_dev":  2.3,
    },
    Score: 88,
}

// Agent-002的分析
result2 := &aggregator.TaskResult{
    ID:      "result-002",
    TaskID:  "data-analysis-456",
    AgentID: "agent-002",
    Data: map[string]interface{}{
        "mean":     10.8,
        "median":   10.0,
        "std_dev":  2.5,
    },
    Score: 92,
}

aggregator.AddResult(result1)
aggregator.AddResult(result2)

aggregated, _ := aggregator.AggregateTask("data-analysis-456")

// 结果:
// mean: 10.65 (平均)
// median: 10.0 (一致)
// std_dev: 2.4 (平均)
// 置信度: ~0.85 (高分数，少冲突)
```

### 场景3: 问答任务

多个Agent回答同一问题：

```go
aggregator := aggregator.NewResultAggregator(
    aggregator.MergeStrategyWeighted,  // 使用加权法
    aggregator.ConflictResolutionHighScore,
)

// 专家Agent (高分)
expertResult := &aggregator.TaskResult{
    ID:      "result-001",
    TaskID:  "qa-789",
    AgentID: "expert-agent",
    Data: map[string]interface{}{
        "answer": "Option B is correct because...",
        "confidence": 0.95,
    },
    Score: 95,  // 高权重
}

// 普通Agent (低分)
normalResult1 := &aggregator.TaskResult{
    ID:      "result-002",
    TaskID:  "qa-789",
    AgentID: "normal-agent-1",
    Data: map[string]interface{}{
        "answer": "I think it's Option A",
        "confidence": 0.6,
    },
    Score: 60,  // 低权重
}

normalResult2 := &aggregator.TaskResult{
    ID:      "result-003",
    TaskID:  "qa-789",
    AgentID: "normal-agent-2",
    Data: map[string]interface{}{
        "answer": "Option B seems right",
        "confidence": 0.7,
    },
    Score: 70,
}

aggregator.AddResult(expertResult)
aggregator.AddResult(normalResult1)
aggregator.AddResult(normalResult2)

aggregated, _ := aggregator.AggregateTask("qa-789")

// 加权合并会更倾向于高分Agent的结果
// 冲突会被高分Agent的答案解决
```

### 场景4: 一致性要求的任务

要求所有Agent结果一致：

```go
aggregator := aggregator.NewResultAggregator(
    aggregator.MergeStrategyConsensus,  // 使用一致性法
    aggregator.ConflictResolutionManual,
)

// 所有结果必须一致，否则需要人工介入

aggregated, err := aggregator.AggregateTask("critical-task-001")
if err != nil {
    log.Fatalf("Aggregation failed: %v", err)
}

if len(aggregated.Conflicts) > 0 {
    // 有冲突，需要人工解决
    for _, conflict := range aggregated.Conflicts {
        fmt.Printf("Manual resolution required for field: %s\n", conflict.Field)
        // 通知人工审核
        notifyHumanReview(conflict)
    }
} else {
    // 所有Agent一致，可以自动采用
    fmt.Println("All agents agree, proceeding automatically")
}
```

## 🔧 高级用法

### 序列化和反序列化

```go
// 序列化结果
data, err := aggregator.SerializeResult(result)
if err != nil {
    log.Fatal(err)
}

// 保存到文件或发送到网络...

// 反序列化结果
result, err := aggregator.DeserializeResult(data)
if err != nil {
    log.Fatal(err)
}

// 序列化聚合结果
aggregatedData, err := aggregator.SerializeAggregatedResult(aggregated)

// 反序列化聚合结果
aggregated, err := aggregator.DeserializeAggregatedResult(aggregatedData)
```

### 直接使用组件

```go
// 单独使用存储
store := aggregator.NewResultStore()
store.AddResult(result)

// 单独使用验证器
validator := aggregator.NewResultValidator()
validator.AddRule(&aggregator.RequiredFieldsRule{Fields: []string{"answer"}})
validator.ValidateAndMark(result)

// 单独使用合并器
merger := aggregator.NewResultMerger(
    aggregator.MergeStrategyVoting,
    aggregator.ConflictResolutionVoting,
)
aggregated, err := merger.Merge("task-001", results)
```

### 监控和统计

```go
// 获取统计信息
totalResults := aggregator.GetStore().GetResultCount()
taskResults := aggregator.GetStore().GetResultCountByTask("task-001")

// 按状态查询
validated := aggregator.GetStore().GetResultsByStatus(aggregator.ResultStatusValidated)
rejected := aggregator.GetStore().GetResultsByStatus(aggregator.ResultStatusRejected)

fmt.Printf("Total: %d, Task: %d, Validated: %d, Rejected: %d\n",
    totalResults, taskResults, len(validated), len(rejected))
```

## 📝 最佳实践

### 1. 选择合适的合并策略

```go
// 投票法 - 适用于离散选项（A/B/C）
MergeStrategyVoting

// 平均法 - 适用于数值结果
MergeStrategyAveraging

// 加权法 - 有专家Agent或信任度不同
MergeStrategyWeighted

// 一致性法 - 关键任务，要求全部一致
MergeStrategyConsensus

// 优先级法 - 信任最优秀的Agent
MergeStrategyPriority
```

### 2. 设置合适的验证规则

```go
validator := aggregator.GetValidator()

// 基本字段验证
validator.AddRule(&aggregator.RequiredFieldsRule{
    Fields: []string{"result", "confidence"},
})

// 类型验证
validator.AddRule(&aggregator.DataTypeRule{
    Field:        "result",
    ExpectedType: "string",
})

// 分数验证
validator.AddRule(&aggregator.ScoreRangeRule{
    MinScore: 0,
    MaxScore: 100,
})
```

### 3. 处理低置信度结果

```go
aggregated, _ := aggregator.AggregateTask("task-001")

if aggregated.Confidence < 0.5 {
    // 低置信度，需要更多结果或人工审核
    log.Printf("Low confidence %.2f, requesting more results", aggregated.Confidence)

    // 请求更多Agent执行任务
    requestMoreResults("task-001")

    // 或通知人工审核
    notifyHumanReview(aggregated)
}
```

### 4. 处理冲突

```go
aggregated, _ := aggregator.AggregateTask("task-001")

for _, conflict := range aggregated.Conflicts {
    if conflict.Resolution == "Manual resolution required" {
        // 需要人工解决
        log.Printf("Manual resolution needed for %s", conflict.Field)

        // 展示给用户选择
        showConflictToUser(conflict)
    } else {
        // 自动解决了
        log.Printf("Conflict in %s resolved: %s", conflict.Field, conflict.Resolution)
    }
}
```

### 5. 错误处理

```go
// 添加结果时处理验证失败
if err := aggregator.AddResult(result); err != nil {
    log.Printf("Result validation failed: %v", err)
    // 结果仍然被存储，但状态为REJECTED

    // 可以查看错误详情
    retrieved, _ := aggregator.GetResult(result.ID)
    fmt.Printf("Error: %s\n", retrieved.Error)
    fmt.Printf("Status: %s\n", retrieved.Status)
}

// 聚合时处理错误
aggregated, err := aggregator.AggregateTask("task-001")
if err != nil {
    if err.Error() == "insufficient results" {
        // 结果不足，等待更多
        waitForMoreResults()
    } else if err.Error() == "no validated results available" {
        // 没有有效结果
        log.Error("All results failed validation")
    }
}
```

## 🧪 测试

```bash
cd projects/phase3-advanced/multi-agent/internal/aggregator
go test -v
```

## 📖 API文档

### ResultStore

- `AddResult(result *TaskResult) error` - 添加结果
- `GetResult(resultID string) (*TaskResult, error)` - 获取结果
- `GetResultsByTask(taskID string) []*TaskResult` - 获取任务的所有结果
- `UpdateResult(result *TaskResult) error` - 更新结果
- `DeleteResult(resultID string) error` - 删除结果
- `GetAllResults() []*TaskResult` - 获取所有结果
- `GetResultCount() int` - 获取结果数量
- `GetResultCountByTask(taskID string) int` - 获取任务的结果数量
- `GetResultsByStatus(status ResultStatus) []*TaskResult` - 按状态查询

### ResultValidator

- `AddRule(rule ValidationRule)` - 添加验证规则
- `Validate(result *TaskResult) error` - 验证结果
- `ValidateAndMark(result *TaskResult) error` - 验证并标记状态
- `ValidateMultiple(results []*TaskResult) map[string]error` - 批量验证

### ResultMerger

- `SetMinResults(min int)` - 设置最少结果数
- `SetConfidenceThreshold(threshold float64)` - 设置置信度阈值
- `Merge(taskID string, results []*TaskResult) (*AggregatedResult, error)` - 合并结果

### ResultAggregator

- `AddResult(result *TaskResult) error` - 添加结果
- `AggregateTask(taskID string) (*AggregatedResult, error)` - 聚合任务结果
- `GetResult(resultID string) (*TaskResult, error)` - 获取结果
- `GetResultsByTask(taskID string) []*TaskResult` - 获取任务的所有结果
- `GetStore() *ResultStore` - 获取存储
- `GetValidator() *ResultValidator` - 获取验证器
- `GetMerger() *ResultMerger` - 获取合并器

### 工具函数

- `SerializeResult(result *TaskResult) ([]byte, error)` - 序列化结果
- `DeserializeResult(data []byte) (*TaskResult, error)` - 反序列化结果
- `SerializeAggregatedResult(result *AggregatedResult) ([]byte, error)` - 序列化聚合结果
- `DeserializeAggregatedResult(data []byte) (*AggregatedResult, error)` - 反序列化聚合结果

## 🔗 相关模块

- [Task Scheduler](../scheduler/README.md) - 任务调度器
- [Communication](../communication/README.md) - 通信模块
- [Task Decomposer](../task-decomposer/README.md) - 任务分解器
- [Protocol](../../protocol/README.md) - 通信协议

---

**版本**: 1.0.0
**许可证**: MIT
