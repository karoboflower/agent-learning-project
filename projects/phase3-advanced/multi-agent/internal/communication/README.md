# Communication Module

> Agent通信模块 - WebSocket服务器、消息路由和连接管理

## 📦 功能特性

- **WebSocket服务器**: 高性能的WebSocket通信服务
- **连接管理**: Agent连接的注册、管理和监控
- **消息路由**: 灵活的消息路由和处理机制
- **消息确认**: 可靠的消息确认机制
- **消息广播**: 支持全局广播和定向广播
- **心跳机制**: 自动检测离线Agent
- **并发安全**: 所有操作线程安全
- **异步处理**: 异步消息队列和worker池

## 🚀 快速开始

### 创建WebSocket服务器

```go
import "github.com/agent-learning/multi-agent/internal/communication"

// 使用默认配置
config := communication.DefaultWebSocketConfig()
server := communication.NewWebSocketServer(config)

// 启动服务器
if err := server.Start(); err != nil {
    log.Fatalf("Failed to start server: %v", err)
}

// 程序退出时停止
defer server.Stop()
```

### 注册消息处理器

```go
// 注册TASK_REQUEST消息处理器
server.RegisterMessageHandler("TASK_REQUEST", func(msg *communication.Message) error {
    log.Printf("Received task request from %s", msg.From)

    // 处理任务请求
    taskID := msg.Payload["task_id"].(string)
    taskType := msg.Payload["task_type"].(string)

    // 发送响应
    response := communication.NewMessageBuilder().
        SetType("TASK_ACCEPT").
        SetFrom("server").
        SetTo(msg.From).
        AddPayloadField("task_id", taskID).
        AddPayloadField("accepted_at", time.Now().Format(time.RFC3339)).
        Build()

    return server.SendMessage(response)
})
```

### Agent连接

Agent通过WebSocket连接到服务器：

```javascript
// JavaScript客户端示例
const ws = new WebSocket('ws://localhost:8080/ws?agent_id=agent-001');

ws.onopen = () => {
    console.log('Connected to server');

    // 发送消息
    const msg = {
        message_id: 'msg-001',
        type: 'HEARTBEAT',
        from: 'agent-001',
        to: 'server',
        timestamp: new Date().toISOString(),
        payload: {
            status: 'ACTIVE',
            load: 0.5
        }
    };

    ws.send(JSON.stringify(msg));
};

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log('Received:', msg);

    // 处理消息...
};
```

### 发送消息

```go
// 发送单播消息
msg := communication.NewMessageBuilder().
    SetMessageID("msg-001").
    SetType("TASK_REQUEST").
    SetFrom("server").
    SetTo("agent-001").
    SetPriority(8).
    AddPayloadField("task_id", "task-123").
    AddPayloadField("task_type", "code_review").
    Build()

server.SendMessage(msg)

// 广播消息给所有Agent
broadcastMsg := communication.NewMessageBuilder().
    SetType("BROADCAST").
    SetFrom("server").
    SetTo("broadcast").
    AddPayloadField("announcement", "System maintenance at 2AM").
    Build()

server.BroadcastMessage(broadcastMsg)
```

## 📚 核心概念

### 1. 连接管理

#### Connection

单个Agent的WebSocket连接：

```go
type Connection struct {
    ID            string           // 连接ID
    AgentID       string           // Agent ID
    Conn          *websocket.Conn  // WebSocket连接
    Status        ConnectionStatus // 连接状态
    ConnectedAt   time.Time        // 连接时间
    LastHeartbeat time.Time        // 最后心跳时间
    SendChan      chan []byte      // 发送通道
}
```

**连接状态**:
- `CONNECTED`: 已连接
- `DISCONNECTED`: 已断开
- `RECONNECTING`: 重连中

#### ConnectionManager

管理所有Agent连接：

```go
cm := server.GetConnectionManager()

// 获取所有连接
conns := cm.ListConnections()

// 获取活跃连接
active := cm.GetActiveConnections()

// 获取特定Agent的连接
conn, err := cm.GetConnectionByAgent("agent-001")

// 检查心跳超时
timeoutConns := cm.CheckHeartbeat(60 * time.Second)
```

### 2. 消息路由

#### MessageRouter

根据消息类型路由到对应处理器：

```go
router := server.GetRouter()

// 注册处理器
router.RegisterHandler("TASK_REQUEST", handleTaskRequest)
router.RegisterHandler("TASK_COMPLETE", handleTaskComplete)
router.RegisterHandler("HEARTBEAT", handleHeartbeat)

// 检查是否有处理器
if router.HasHandler("TASK_REQUEST") {
    log.Println("Task request handler registered")
}

// 获取处理器数量
count := router.GetHandlerCount()
```

#### MessageDispatcher

异步消息分发器：

```go
dispatcher := server.GetDispatcher()

// 查看队列大小
inQueueSize := dispatcher.GetInQueueSize()
outQueueSize := dispatcher.GetOutQueueSize()

// 发送给特定Agent
dispatcher.SendToAgent("agent-001", msg)

// 发送给多个Agent
dispatcher.SendToAgents([]string{"agent-001", "agent-002"}, msg)

// 广播
dispatcher.BroadcastMessage(msg)
```

### 3. 消息确认

确保消息可靠传递：

```go
ackMgr := communication.NewAckManager(10 * time.Second)

// 注册消息
ackMgr.RegisterMessage("msg-001")

// 发送消息...

// 等待确认
ack, err := ackMgr.WaitForAck("msg-001")
if err != nil {
    log.Printf("Ack timeout: %v", err)
} else {
    log.Printf("Ack status: %s", ack.Status)
}

// 确认消息（在接收到响应时）
ackMgr.Confirm("msg-001", true, "")

// 清理过期确认
count := ackMgr.CleanupExpired(5 * time.Minute)
```

**确认状态**:
- `PENDING`: 等待确认
- `CONFIRMED`: 已确认
- `TIMEOUT`: 超时
- `FAILED`: 失败

### 4. 消息格式

统一的消息格式：

```go
type Message struct {
    MessageID string                 `json:"message_id"`
    Type      string                 `json:"type"`
    From      string                 `json:"from"`
    To        string                 `json:"to"`
    Timestamp string                 `json:"timestamp"`
    Priority  int                    `json:"priority,omitempty"`
    Payload   map[string]interface{} `json:"payload"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

**消息构建**:

```go
// 使用Builder模式
msg := communication.NewMessageBuilder().
    SetMessageID(uuid.New().String()).
    SetType("TASK_REQUEST").
    SetFrom("coordinator").
    SetTo("agent-001").
    SetPriority(8).
    AddPayloadField("task_id", "task-123").
    AddPayloadField("task_type", "code_review").
    AddMetadataField("correlation_id", "trace-456").
    Build()

// 序列化
data, err := communication.SerializeMessage(msg)

// 反序列化
msg, err := communication.DeserializeMessage(data)
```

**消息验证**:

```go
validator := communication.NewMessageValidator()

// 验证消息
if err := validator.Validate(msg); err != nil {
    log.Printf("Invalid message: %v", err)
}

// 验证负载
requiredFields := []string{"task_id", "task_type"}
if err := validator.ValidatePayload(msg, requiredFields); err != nil {
    log.Printf("Invalid payload: %v", err)
}
```

### 5. 心跳机制

自动发送Ping/Pong保持连接：

```go
config := &communication.WebSocketConfig{
    PingInterval: 30 * time.Second,  // 每30秒发送ping
    PongTimeout:  60 * time.Second,  // 60秒无pong则超时
}

server := communication.NewWebSocketServer(config)
```

Agent端自动处理Pong响应，服务器自动检测超时连接。

## 🎯 使用场景

### 场景1: 任务分配系统

```go
// 服务器端
server := communication.NewWebSocketServer(nil)
server.Start()

// 注册任务请求处理器
server.RegisterMessageHandler("TASK_ACCEPT", func(msg *communication.Message) error {
    taskID := msg.Payload["task_id"].(string)
    agentID := msg.From

    log.Printf("Agent %s accepted task %s", agentID, taskID)

    // 更新任务状态
    updateTaskStatus(taskID, "ASSIGNED", agentID)

    return nil
})

// 分配任务
func assignTask(taskID, agentID string) error {
    msg := communication.NewMessageBuilder().
        SetType("TASK_REQUEST").
        SetFrom("server").
        SetTo(agentID).
        SetPriority(8).
        AddPayloadField("task_id", taskID).
        AddPayloadField("task_type", "code_review").
        AddPayloadField("timeout", 300).
        Build()

    return server.SendMessage(msg)
}
```

### 场景2: 实时状态监控

```go
// 收集Agent状态
server.RegisterMessageHandler("HEARTBEAT", func(msg *communication.Message) error {
    agentID := msg.From
    status := msg.Payload["status"].(string)
    load := msg.Payload["load"].(float64)

    // 更新Agent状态
    updateAgentStatus(agentID, status, load)

    return nil
})

// 定期查询所有Agent状态
func monitorAgents() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        cm := server.GetConnectionManager()
        conns := cm.GetActiveConnections()

        for _, conn := range conns {
            if !conn.IsAlive(60 * time.Second) {
                log.Printf("Agent %s heartbeat timeout", conn.AgentID)
                // 标记Agent离线
                markAgentOffline(conn.AgentID)
            }
        }
    }
}
```

### 场景3: 广播通知

```go
// 系统维护通知
func broadcastMaintenance() {
    msg := communication.NewMessageBuilder().
        SetType("SYSTEM_NOTIFICATION").
        SetFrom("server").
        SetTo("broadcast").
        SetPriority(10).
        AddPayloadField("type", "maintenance").
        AddPayloadField("message", "System maintenance in 5 minutes").
        AddPayloadField("scheduled_at", time.Now().Add(5*time.Minute).Format(time.RFC3339)).
        Build()

    server.BroadcastMessage(msg)
}

// 定向通知特定Agent组
func notifyAgentGroup(agentIDs []string, message string) {
    msg := communication.NewMessageBuilder().
        SetType("GROUP_NOTIFICATION").
        SetFrom("server").
        AddPayloadField("message", message).
        Build()

    dispatcher := server.GetDispatcher()
    dispatcher.SendToAgents(agentIDs, msg)
}
```

### 场景4: 可靠消息传递

```go
// 使用消息确认机制
ackMgr := communication.NewAckManager(10 * time.Second)

func sendTaskWithAck(taskID, agentID string) error {
    msg := communication.NewMessageBuilder().
        SetMessageID(uuid.New().String()).
        SetType("TASK_REQUEST").
        SetFrom("server").
        SetTo(agentID).
        AddPayloadField("task_id", taskID).
        Build()

    // 注册等待确认
    ackMgr.RegisterMessage(msg.MessageID)

    // 发送消息
    if err := server.SendMessage(msg); err != nil {
        return err
    }

    // 等待确认
    ack, err := ackMgr.WaitForAck(msg.MessageID)
    if err != nil {
        log.Printf("Task %s assignment failed: %v", taskID, err)
        return err
    }

    if ack.Status == communication.AckStatusConfirmed {
        log.Printf("Task %s successfully assigned to %s", taskID, agentID)
        return nil
    }

    return fmt.Errorf("task assignment failed: %s", ack.Error)
}
```

## 🔧 配置选项

```go
type WebSocketConfig struct {
    Host              string        // 监听地址 (默认: 0.0.0.0)
    Port              int           // 监听端口 (默认: 8080)
    ReadBufferSize    int           // 读缓冲区大小 (默认: 1024)
    WriteBufferSize   int           // 写缓冲区大小 (默认: 1024)
    HandshakeTimeout  time.Duration // 握手超时 (默认: 10s)
    ReadTimeout       time.Duration // 读超时 (默认: 60s)
    WriteTimeout      time.Duration // 写超时 (默认: 10s)
    PingInterval      time.Duration // Ping间隔 (默认: 30s)
    PongTimeout       time.Duration // Pong超时 (默认: 60s)
    MessageQueueSize  int           // 消息队列大小 (默认: 1000)
    WorkerPoolSize    int           // Worker池大小 (默认: 10)
}

// 自定义配置
config := &communication.WebSocketConfig{
    Host:             "0.0.0.0",
    Port:             9000,
    PingInterval:     15 * time.Second,
    PongTimeout:      30 * time.Second,
    MessageQueueSize: 5000,
    WorkerPoolSize:   20,
}

server := communication.NewWebSocketServer(config)
```

## 📊 监控和统计

### 健康检查

```bash
curl http://localhost:8080/health
```

响应：
```json
{
    "status": "healthy",
    "connections": 15,
    "active_connections": 12,
    "in_queue_size": 3,
    "out_queue_size": 1
}
```

### 连接统计

```go
cm := server.GetConnectionManager()

// 总连接数
total := cm.GetConnectionCount()

// 按状态统计
stats := cm.GetConnectionCountByStatus()
fmt.Printf("Connected: %d\n", stats[communication.ConnectionStatusConnected])
fmt.Printf("Disconnected: %d\n", stats[communication.ConnectionStatusDisconnected])

// 活跃连接
active := len(cm.GetActiveConnections())
```

### 消息统计

```go
dispatcher := server.GetDispatcher()

// 队列大小
inQueue := dispatcher.GetInQueueSize()
outQueue := dispatcher.GetOutQueueSize()

// 处理器数量
handlerCount := server.GetRouter().GetHandlerCount()

// 确认统计
ackMgr := communication.NewAckManager(10 * time.Second)
ackStats := ackMgr.GetAckStats()
fmt.Printf("Pending: %d\n", ackStats[communication.AckStatusPending])
fmt.Printf("Confirmed: %d\n", ackStats[communication.AckStatusConfirmed])
fmt.Printf("Failed: %d\n", ackStats[communication.AckStatusFailed])
```

## 📝 最佳实践

### 1. 合理设置超时

```go
config := &communication.WebSocketConfig{
    HandshakeTimeout: 10 * time.Second,   // 握手超时
    ReadTimeout:      60 * time.Second,   // 读超时
    WriteTimeout:     10 * time.Second,   // 写超时
    PingInterval:     30 * time.Second,   // Ping间隔
    PongTimeout:      60 * time.Second,   // Pong超时（建议是PingInterval的2倍）
}
```

### 2. 处理连接断开

```go
// 监控连接断开事件
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        cm := server.GetConnectionManager()
        timeoutConns := cm.CheckHeartbeat(60 * time.Second)

        for _, connID := range timeoutConns {
            log.Printf("Connection %s timeout", connID)
            // 清理资源
            // 更新Agent状态
            // 重新分配任务等
        }
    }
}()
```

### 3. 优雅关闭

```go
// 捕获信号
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

<-sigChan

// 发送关闭通知
shutdownMsg := communication.NewMessageBuilder().
    SetType("SERVER_SHUTDOWN").
    SetFrom("server").
    SetTo("broadcast").
    AddPayloadField("message", "Server shutting down").
    Build()

server.BroadcastMessage(shutdownMsg)

// 等待消息发送完成
time.Sleep(1 * time.Second)

// 停止服务器
server.Stop()
```

### 4. 错误处理

```go
server.RegisterMessageHandler("TASK_REQUEST", func(msg *communication.Message) error {
    // 验证消息
    validator := communication.NewMessageValidator()
    if err := validator.Validate(msg); err != nil {
        log.Printf("Invalid message: %v", err)
        // 发送错误响应
        sendErrorResponse(msg.From, "VALIDATION_ERROR", err.Error())
        return err
    }

    // 处理任务...

    return nil
})

func sendErrorResponse(to, errorCode, errorMsg string) {
    errMsg := communication.NewMessageBuilder().
        SetType("ERROR").
        SetFrom("server").
        SetTo(to).
        AddPayloadField("error_code", errorCode).
        AddPayloadField("error_message", errorMsg).
        Build()

    server.SendMessage(errMsg)
}
```

### 5. 监控队列积压

```go
// 定期检查队列
func monitorQueues() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        dispatcher := server.GetDispatcher()

        inQueue := dispatcher.GetInQueueSize()
        outQueue := dispatcher.GetOutQueueSize()

        if inQueue > 800 {  // 80%容量
            log.Warn("Incoming queue almost full")
        }

        if outQueue > 800 {
            log.Warn("Outgoing queue almost full")
        }
    }
}
```

## 🧪 测试

```bash
cd projects/phase3-advanced/multi-agent/internal/communication
go test -v
```

## 📖 API文档

### WebSocketServer

- `Start() error` - 启动服务器
- `Stop() error` - 停止服务器
- `RegisterMessageHandler(messageType string, handler MessageHandler)` - 注册处理器
- `SendMessage(msg *Message) error` - 发送消息
- `BroadcastMessage(msg *Message) error` - 广播消息
- `GetConnectionManager() *ConnectionManager` - 获取连接管理器
- `GetRouter() *MessageRouter` - 获取路由器
- `GetDispatcher() *MessageDispatcher` - 获取分发器

### ConnectionManager

- `AddConnection(conn *Connection) error` - 添加连接
- `RemoveConnection(connID string) error` - 移除连接
- `GetConnection(connID string) (*Connection, error)` - 获取连接
- `GetConnectionByAgent(agentID string) (*Connection, error)` - 按Agent获取
- `ListConnections() []*Connection` - 列出所有连接
- `GetActiveConnections() []*Connection` - 获取活跃连接
- `BroadcastToAll(data []byte) error` - 全局广播
- `BroadcastToAgents(agentIDs []string, data []byte) error` - 定向广播
- `CheckHeartbeat(timeout time.Duration) []string` - 检查心跳

### MessageRouter

- `RegisterHandler(messageType string, handler MessageHandler)` - 注册处理器
- `UnregisterHandler(messageType string)` - 注销处理器
- `Route(msg *Message) error` - 路由消息
- `HasHandler(messageType string) bool` - 检查处理器
- `GetHandlerCount() int` - 获取处理器数量

### AckManager

- `RegisterMessage(messageID string)` - 注册消息
- `Confirm(messageID string, success bool, errorMsg string) error` - 确认消息
- `WaitForAck(messageID string) (*Acknowledgment, error)` - 等待确认
- `GetAck(messageID string) (*Acknowledgment, error)` - 获取确认
- `CleanupExpired(expireAfter time.Duration) int` - 清理过期
- `GetPendingCount() int` - 获取待确认数
- `GetAckStats() map[AckStatus]int` - 获取统计

## 🔗 相关模块

- [Task Scheduler](../scheduler/README.md) - 任务调度器
- [Task Decomposer](../task-decomposer/README.md) - 任务分解器
- [Protocol](../../protocol/README.md) - 通信协议

---

**版本**: 1.0.0
**许可证**: MIT
