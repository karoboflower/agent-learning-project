# 多Agent协作系统 - 通信协议设计

> Multi-Agent Communication Protocol Specification v1.0

## 📋 概述

本文档定义了多Agent协作系统中Agent之间的通信协议，包括消息格式、消息类型、通信流程等。

## 🎯 设计目标

1. **简单性** - 协议易于理解和实现
2. **可扩展性** - 支持未来功能扩展
3. **可靠性** - 确保消息可靠传递
4. **高效性** - 最小化通信开销
5. **安全性** - 支持消息验证和加密

## 📦 消息格式

### 基础消息结构

所有消息遵循统一的JSON格式：

```json
{
  "message_id": "uuid",
  "type": "message_type",
  "from": "agent_id",
  "to": "agent_id or broadcast",
  "timestamp": "2026-01-28T10:00:00Z",
  "priority": 1,
  "payload": {},
  "metadata": {}
}
```

### 字段说明

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `message_id` | string | ✅ | 消息唯一标识符（UUID） |
| `type` | string | ✅ | 消息类型（见消息类型章节） |
| `from` | string | ✅ | 发送者Agent ID |
| `to` | string | ✅ | 接收者Agent ID或"broadcast" |
| `timestamp` | string | ✅ | ISO 8601格式时间戳 |
| `priority` | int | ❌ | 消息优先级（1-10，默认5） |
| `payload` | object | ✅ | 消息负载（具体内容见各消息类型） |
| `metadata` | object | ❌ | 元数据（用于扩展） |

## 📨 消息类型

### 1. 任务请求消息 (TASK_REQUEST)

Agent请求另一个Agent执行任务。

```json
{
  "message_id": "req-001",
  "type": "TASK_REQUEST",
  "from": "coordinator-agent",
  "to": "worker-agent-1",
  "timestamp": "2026-01-28T10:00:00Z",
  "priority": 5,
  "payload": {
    "task_id": "task-001",
    "task_type": "code_review",
    "input": "function add(a, b) { return a + b; }",
    "requirements": {
      "language": "javascript",
      "check_security": true
    },
    "timeout": 300,
    "callback_url": "http://coordinator/callback"
  },
  "metadata": {
    "parent_task_id": "main-task-001",
    "correlation_id": "corr-001"
  }
}
```

**Payload 字段**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `task_id` | string | 任务唯一标识符 |
| `task_type` | string | 任务类型 |
| `input` | string/object | 任务输入数据 |
| `requirements` | object | 任务要求 |
| `timeout` | int | 超时时间（秒） |
| `callback_url` | string | 回调URL（可选） |

### 2. 任务接受消息 (TASK_ACCEPT)

Agent接受任务请求。

```json
{
  "message_id": "acc-001",
  "type": "TASK_ACCEPT",
  "from": "worker-agent-1",
  "to": "coordinator-agent",
  "timestamp": "2026-01-28T10:00:01Z",
  "priority": 5,
  "payload": {
    "task_id": "task-001",
    "estimated_duration": 60,
    "accepted_at": "2026-01-28T10:00:01Z"
  }
}
```

### 3. 任务拒绝消息 (TASK_REJECT)

Agent拒绝任务请求。

```json
{
  "message_id": "rej-001",
  "type": "TASK_REJECT",
  "from": "worker-agent-1",
  "to": "coordinator-agent",
  "timestamp": "2026-01-28T10:00:01Z",
  "priority": 5,
  "payload": {
    "task_id": "task-001",
    "reason": "CAPABILITY_MISMATCH",
    "message": "This agent does not support code_review tasks",
    "suggested_agents": ["worker-agent-2", "worker-agent-3"]
  }
}
```

**拒绝原因码**：

| 原因码 | 说明 |
|--------|------|
| `CAPABILITY_MISMATCH` | 能力不匹配 |
| `RESOURCE_UNAVAILABLE` | 资源不足 |
| `OVERLOADED` | 负载过高 |
| `MAINTENANCE` | 维护中 |
| `INVALID_REQUEST` | 无效请求 |

### 4. 任务完成消息 (TASK_COMPLETE)

Agent完成任务。

```json
{
  "message_id": "cmp-001",
  "type": "TASK_COMPLETE",
  "from": "worker-agent-1",
  "to": "coordinator-agent",
  "timestamp": "2026-01-28T10:01:00Z",
  "priority": 5,
  "payload": {
    "task_id": "task-001",
    "status": "SUCCESS",
    "output": {
      "review_result": "PASS",
      "issues": [],
      "suggestions": ["Consider adding input validation"]
    },
    "duration": 59,
    "completed_at": "2026-01-28T10:01:00Z"
  }
}
```

**任务状态**：

| 状态 | 说明 |
|------|------|
| `SUCCESS` | 成功完成 |
| `FAILED` | 失败 |
| `PARTIAL` | 部分完成 |
| `TIMEOUT` | 超时 |

### 5. 任务失败消息 (TASK_FAILED)

Agent执行任务失败。

```json
{
  "message_id": "fail-001",
  "type": "TASK_FAILED",
  "from": "worker-agent-1",
  "to": "coordinator-agent",
  "timestamp": "2026-01-28T10:01:00Z",
  "priority": 8,
  "payload": {
    "task_id": "task-001",
    "error_code": "EXECUTION_ERROR",
    "error_message": "Failed to parse input code",
    "error_details": {
      "line": 1,
      "column": 20,
      "expected": "function body"
    },
    "retry_possible": true
  }
}
```

### 6. 心跳消息 (HEARTBEAT)

Agent定期发送心跳表明存活。

```json
{
  "message_id": "hb-001",
  "type": "HEARTBEAT",
  "from": "worker-agent-1",
  "to": "coordinator-agent",
  "timestamp": "2026-01-28T10:00:00Z",
  "priority": 1,
  "payload": {
    "status": "ACTIVE",
    "load": 0.45,
    "tasks_running": 3,
    "tasks_queued": 5,
    "capabilities": ["code_review", "refactoring"]
  }
}
```

**Agent状态**：

| 状态 | 说明 |
|------|------|
| `ACTIVE` | 活跃 |
| `IDLE` | 空闲 |
| `BUSY` | 忙碌 |
| `MAINTENANCE` | 维护中 |
| `ERROR` | 错误 |

### 7. 状态查询消息 (STATUS_QUERY)

查询Agent或任务状态。

```json
{
  "message_id": "sq-001",
  "type": "STATUS_QUERY",
  "from": "coordinator-agent",
  "to": "worker-agent-1",
  "timestamp": "2026-01-28T10:00:00Z",
  "priority": 3,
  "payload": {
    "query_type": "TASK_STATUS",
    "task_id": "task-001"
  }
}
```

### 8. 状态响应消息 (STATUS_RESPONSE)

响应状态查询。

```json
{
  "message_id": "sr-001",
  "type": "STATUS_RESPONSE",
  "from": "worker-agent-1",
  "to": "coordinator-agent",
  "timestamp": "2026-01-28T10:00:01Z",
  "priority": 3,
  "payload": {
    "query_id": "sq-001",
    "task_id": "task-001",
    "status": "RUNNING",
    "progress": 45,
    "estimated_remaining": 30
  }
}
```

### 9. 广播消息 (BROADCAST)

向所有Agent广播消息。

```json
{
  "message_id": "bc-001",
  "type": "BROADCAST",
  "from": "coordinator-agent",
  "to": "broadcast",
  "timestamp": "2026-01-28T10:00:00Z",
  "priority": 7,
  "payload": {
    "event": "SYSTEM_SHUTDOWN",
    "message": "System will shutdown in 5 minutes",
    "countdown": 300
  }
}
```

## 🔄 通信流程

### 典型任务执行流程

```
Coordinator                 Worker Agent
    │                            │
    ├──── TASK_REQUEST ────────▶│
    │                            │
    │◀──── TASK_ACCEPT ─────────┤
    │                            │
    │                            ├─ 执行任务
    │                            │
    │◀──── STATUS_RESPONSE ─────┤ (可选进度更新)
    │                            │
    │                            ├─ 完成任务
    │                            │
    │◀──── TASK_COMPLETE ───────┤
    │                            │
```

### 任务拒绝流程

```
Coordinator                 Worker Agent
    │                            │
    ├──── TASK_REQUEST ────────▶│
    │                            │
    │◀──── TASK_REJECT ─────────┤
    │                            │
    ├──── TASK_REQUEST ────────▶│ (发送给建议的Agent)
    │                            │
```

### 心跳机制

```
Agent                      Coordinator
  │                            │
  ├──── HEARTBEAT ───────────▶│
  │                            │
  ├──── HEARTBEAT ───────────▶│ (每30秒)
  │                            │
  ├──── HEARTBEAT ───────────▶│
  │                            │
```

如果连续3次心跳失败，Agent被标记为不可用。

## 🛡️ 错误处理

### 错误消息格式

```json
{
  "message_id": "err-001",
  "type": "ERROR",
  "from": "worker-agent-1",
  "to": "coordinator-agent",
  "timestamp": "2026-01-28T10:00:00Z",
  "priority": 9,
  "payload": {
    "error_type": "PROTOCOL_ERROR",
    "error_code": "INVALID_MESSAGE_FORMAT",
    "error_message": "Missing required field: task_id",
    "original_message_id": "req-001",
    "severity": "ERROR"
  }
}
```

### 错误类型

| 错误类型 | 说明 |
|----------|------|
| `PROTOCOL_ERROR` | 协议错误 |
| `VALIDATION_ERROR` | 验证错误 |
| `EXECUTION_ERROR` | 执行错误 |
| `TIMEOUT_ERROR` | 超时错误 |
| `RESOURCE_ERROR` | 资源错误 |

### 错误严重级别

| 级别 | 说明 | 处理方式 |
|------|------|----------|
| `INFO` | 信息 | 记录日志 |
| `WARNING` | 警告 | 记录日志，可能重试 |
| `ERROR` | 错误 | 记录日志，通知相关Agent |
| `CRITICAL` | 严重 | 立即处理，可能需要人工介入 |

## 🔐 安全机制

### 消息签名

每个消息可以包含签名以验证发送者身份：

```json
{
  "message_id": "req-001",
  "type": "TASK_REQUEST",
  "from": "coordinator-agent",
  "to": "worker-agent-1",
  "timestamp": "2026-01-28T10:00:00Z",
  "signature": "SHA256:abc123...",
  "payload": {...}
}
```

### 消息加密

敏感消息的payload可以加密：

```json
{
  "message_id": "req-001",
  "type": "TASK_REQUEST",
  "from": "coordinator-agent",
  "to": "worker-agent-1",
  "timestamp": "2026-01-28T10:00:00Z",
  "encrypted": true,
  "encryption_algorithm": "AES-256-GCM",
  "payload": "encrypted_base64_data..."
}
```

## 📊 消息优先级

| 优先级 | 级别 | 用途 |
|--------|------|------|
| 1-2 | 最低 | 心跳、日志 |
| 3-4 | 低 | 状态查询、统计 |
| 5-6 | 普通 | 常规任务 |
| 7-8 | 高 | 重要任务、错误通知 |
| 9-10 | 最高 | 紧急任务、系统事件 |

## 🌐 传输协议

### WebSocket

主要通信方式，支持：
- 双向通信
- 实时消息推送
- 连接保活

**连接端点**：`ws://host:port/agent/ws`

**连接参数**：
- `agent_id` - Agent标识
- `auth_token` - 认证令牌

### HTTP (备用)

用于非实时通信：

**发送消息**：
```
POST /api/v1/messages
Content-Type: application/json

{消息JSON}
```

**拉取消息**：
```
GET /api/v1/messages?agent_id=xxx&since=timestamp
```

## 📏 限制和约束

| 项目 | 限制 |
|------|------|
| 消息大小 | 最大 1MB |
| 消息队列长度 | 最大 10000 条 |
| 心跳间隔 | 30 秒 |
| 心跳超时 | 90 秒（3次失败） |
| 任务超时 | 默认 300 秒，可配置 |
| 重试次数 | 最多 3 次 |

## 🔍 消息追踪

使用 `correlation_id` 追踪相关消息：

```json
{
  "message_id": "req-001",
  "type": "TASK_REQUEST",
  "metadata": {
    "correlation_id": "trace-001",
    "parent_message_id": "main-001",
    "trace_level": 1
  }
}
```

## 📝 最佳实践

### 1. 消息ID生成
```go
messageID := uuid.New().String()
```

### 2. 时间戳格式
```go
timestamp := time.Now().Format(time.RFC3339)
```

### 3. 消息验证
```go
func ValidateMessage(msg *Message) error {
    if msg.MessageID == "" {
        return errors.New("message_id is required")
    }
    if msg.Type == "" {
        return errors.New("type is required")
    }
    // ...
    return nil
}
```

### 4. 错误处理
```go
if err != nil {
    sendErrorMessage(msg.From, "EXECUTION_ERROR", err.Error())
}
```

## 🎯 使用示例

### 发送任务请求

```go
msg := &Message{
    MessageID: uuid.New().String(),
    Type:      "TASK_REQUEST",
    From:      "coordinator",
    To:        "worker-1",
    Timestamp: time.Now().Format(time.RFC3339),
    Priority:  5,
    Payload: map[string]interface{}{
        "task_id":   "task-001",
        "task_type": "code_review",
        "input":     "...",
    },
}

err := sendMessage(msg)
```

### 处理接收消息

```go
func handleMessage(msg *Message) error {
    switch msg.Type {
    case "TASK_REQUEST":
        return handleTaskRequest(msg)
    case "TASK_COMPLETE":
        return handleTaskComplete(msg)
    case "HEARTBEAT":
        return handleHeartbeat(msg)
    default:
        return fmt.Errorf("unknown message type: %s", msg.Type)
    }
}
```

## 📚 参考实现

协议的Go语言实现见：
- `protocol/message.go` - 消息定义
- `protocol/types.go` - 类型定义
- `protocol/validator.go` - 消息验证
- `protocol/serializer.go` - 序列化/反序列化

---

**版本**: 1.0.0
**最后更新**: 2026-01-28
**维护者**: Multi-Agent System Team
