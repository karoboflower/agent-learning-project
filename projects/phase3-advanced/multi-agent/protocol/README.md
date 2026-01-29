# Multi-Agent Communication Protocol

> Agent间通信协议的Go语言实现

## 📦 安装

```bash
go get github.com/agent-learning/multi-agent/protocol
```

## 🚀 快速开始

### 创建消息

```go
import "github.com/agent-learning/multi-agent/protocol"

// 创建任务请求消息
msg := protocol.NewMessage(
    protocol.MessageTypeTaskRequest,
    "coordinator",
    "worker-1",
)

// 设置负载
payload := &protocol.TaskRequestPayload{
    TaskID:   "task-001",
    TaskType: "code_review",
    Input:    "function add(a, b) { return a + b; }",
    Timeout:  300,
}

payloadMap, _ := protocol.SerializePayload(payload)
msg.Payload = payloadMap
```

### 验证消息

```go
validator := protocol.NewValidator()
if err := validator.Validate(msg); err != nil {
    log.Fatalf("Invalid message: %v", err)
}
```

### 序列化消息

```go
serializer := protocol.NewSerializer()

// 序列化为JSON
data, err := serializer.Serialize(msg)
if err != nil {
    log.Fatalf("Serialization failed: %v", err)
}

// 发送消息
sendToAgent(data)
```

### 反序列化消息

```go
// 接收消息
data := receiveFromAgent()

// 反序列化
msg, err := serializer.Deserialize(data)
if err != nil {
    log.Fatalf("Deserialization failed: %v", err)
}

// 处理消息
handleMessage(msg)
```

## 📚 消息类型

### 任务请求 (TASK_REQUEST)

```go
msg := protocol.NewMessage(
    protocol.MessageTypeTaskRequest,
    "from_agent",
    "to_agent",
)

payload := &protocol.TaskRequestPayload{
    TaskID:   "task-001",
    TaskType: "code_review",
    Input:    "code here",
    Timeout:  300,
}

payloadMap, _ := protocol.SerializePayload(payload)
msg.Payload = payloadMap
```

### 任务接受 (TASK_ACCEPT)

```go
msg := protocol.NewMessage(
    protocol.MessageTypeTaskAccept,
    "worker_agent",
    "coordinator",
)

payload := &protocol.TaskAcceptPayload{
    TaskID:            "task-001",
    EstimatedDuration: 60,
    AcceptedAt:        time.Now().Format(time.RFC3339),
}

payloadMap, _ := protocol.SerializePayload(payload)
msg.Payload = payloadMap
```

### 任务完成 (TASK_COMPLETE)

```go
msg := protocol.NewMessage(
    protocol.MessageTypeTaskComplete,
    "worker_agent",
    "coordinator",
)

payload := &protocol.TaskCompletePayload{
    TaskID:      "task-001",
    Status:      protocol.TaskStatusSuccess,
    Output:      map[string]interface{}{"result": "success"},
    Duration:    59,
    CompletedAt: time.Now().Format(time.RFC3339),
}

payloadMap, _ := protocol.SerializePayload(payload)
msg.Payload = payloadMap
```

### 心跳 (HEARTBEAT)

```go
msg := protocol.NewMessage(
    protocol.MessageTypeHeartbeat,
    "worker_agent",
    "coordinator",
)
msg.Priority = 1 // 低优先级

payload := &protocol.HeartbeatPayload{
    Status:       protocol.AgentStatusActive,
    Load:         0.45,
    TasksRunning: 3,
    TasksQueued:  5,
    Capabilities: []string{"code_review", "refactoring"},
}

payloadMap, _ := protocol.SerializePayload(payload)
msg.Payload = payloadMap
```

## 🔍 消息验证

验证器会检查：
- 必需字段
- 字段类型
- 字段值范围
- 时间戳格式
- 负载完整性

```go
validator := protocol.NewValidator()

// 严格模式（默认）
validator.SetStrictMode(true)

// 验证消息
if err := validator.Validate(msg); err != nil {
    // 处理验证错误
    log.Printf("Validation error: %v", err)
}
```

## 📊 优先级

消息优先级范围：1-10

| 优先级 | 用途 |
|--------|------|
| 1-2 | 心跳、日志 |
| 3-4 | 状态查询 |
| 5-6 | 常规任务 |
| 7-8 | 重要任务 |
| 9-10 | 紧急任务 |

```go
msg.Priority = 8 // 高优先级
```

## 🏷️ 元数据

使用元数据进行消息追踪：

```go
// 设置元数据
msg.SetMetadata("correlation_id", "trace-001")
msg.SetMetadata("parent_task_id", "main-001")

// 获取元数据
if correlationID, ok := msg.GetMetadata("correlation_id"); ok {
    log.Printf("Correlation ID: %v", correlationID)
}
```

## 🔐 安全特性

### 消息签名

```go
// 添加签名
msg.Signature = generateSignature(msg)
```

### 消息加密

```go
// 标记为加密消息
msg.Encrypted = true
msg.EncryptionAlgorithm = "AES-256-GCM"
```

## 📝 最佳实践

### 1. 总是验证接收的消息

```go
validator := protocol.NewValidator()
if err := validator.Validate(msg); err != nil {
    return fmt.Errorf("invalid message: %w", err)
}
```

### 2. 设置适当的优先级

```go
// 心跳消息
msg.Priority = 1

// 紧急任务
msg.Priority = 9
```

### 3. 使用元数据进行追踪

```go
msg.SetMetadata("correlation_id", correlationID)
msg.SetMetadata("trace_level", traceLevel)
```

### 4. 处理错误情况

```go
func handleMessage(msg *protocol.Message) error {
    if err := validateMessage(msg); err != nil {
        sendErrorMessage(msg.From, "VALIDATION_ERROR", err.Error())
        return err
    }
    // ...
}
```

## 🧪 测试

```bash
go test ./protocol
```

## 📖 API文档

完整的API文档见：[docs/architecture/multi-agent-protocol.md](../../../docs/architecture/multi-agent-protocol.md)

## 📋 示例

更多示例见：[examples/](../examples/)

---

**版本**: 1.0.0
**许可证**: MIT
