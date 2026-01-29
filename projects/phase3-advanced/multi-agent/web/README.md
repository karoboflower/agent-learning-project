# Multi-Agent Web Interface

> 多Agent协作系统Web界面

## 📦 功能特性

### Agent管理界面
- ✅ Agent列表展示
- ✅ Agent状态显示（活跃/空闲/离线）
- ✅ Agent注册界面
- ✅ Agent能力展示
- ✅ Agent负载监控
- ✅ Agent详情查看
- ✅ Agent删除操作

### 任务监控界面
- ✅ 任务列表展示
- ✅ 任务状态显示（待分配/执行中/已完成）
- ✅ 任务创建界面
- ✅ 任务分配可视化
- ✅ 任务优先级管理
- ✅ 任务进度监控
- ✅ 任务详情查看

### 结果展示界面
- ✅ 结果列表展示
- ✅ 结果状态显示
- ✅ 结果对比功能
- ✅ 结果导出（JSON格式）
- ✅ 置信度统计
- ✅ 冲突展示
- ✅ 聚合结果展示

### 实时更新
- ✅ WebSocket实时通信
- ✅ Agent状态实时更新
- ✅ 任务状态实时更新
- ✅ 结果实时推送
- ✅ 自动重连机制

## 🚀 快速开始

### 1. 启动服务器

```bash
cd projects/phase3-advanced/multi-agent
go run cmd/server/main.go
```

服务器将在以下端口启动：
- WebSocket: `ws://localhost:8080/ws`
- Web UI: `http://localhost:8080`
- API: `http://localhost:8080/api`

### 2. 访问Web界面

在浏览器中打开：
```
http://localhost:8080
```

### 3. 注册Agent

**方式1: 通过Web界面**

1. 点击"Agent管理"标签
2. 点击"注册新Agent"按钮
3. 填写Agent信息：
   - Agent ID
   - 名称
   - 能力（逗号分隔）
   - 最大任务数
4. 点击"注册"提交

**方式2: 通过API**

```bash
curl -X POST http://localhost:8080/api/agents \
  -H "Content-Type: application/json" \
  -d '{
    "id": "agent-001",
    "name": "Code Review Agent",
    "capabilities": ["code_review", "testing"],
    "max_tasks": 5
  }'
```

**方式3: 通过WebSocket**

```javascript
// 连接WebSocket
const ws = new WebSocket('ws://localhost:8080/ws?agent_id=agent-001');

// 发送注册消息
ws.send(JSON.stringify({
    message_id: 'msg-001',
    type: 'AGENT_REGISTER',
    from: 'agent-001',
    to: 'server',
    timestamp: new Date().toISOString(),
    payload: {
        name: 'Code Review Agent',
        capabilities: ['code_review', 'testing'],
        max_tasks: 5
    }
}));
```

### 4. 创建任务

**方式1: 通过Web界面**

1. 点击"任务监控"标签
2. 点击"创建任务"按钮
3. 填写任务信息：
   - 任务ID
   - 任务类型
   - 优先级
   - 描述
   - 所需能力
4. 点击"创建"提交

**方式2: 通过API**

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "id": "task-001",
    "type": "code_review",
    "priority": 8,
    "description": "Review pull request #123",
    "capabilities": ["code_review"]
  }'
```

### 5. 提交结果

Agent通过WebSocket提交任务结果：

```javascript
ws.send(JSON.stringify({
    message_id: 'msg-002',
    type: 'TASK_RESULT',
    from: 'agent-001',
    to: 'server',
    timestamp: new Date().toISOString(),
    payload: {
        task_id: 'task-001',
        data: {
            result: 'APPROVED',
            issues_found: 2,
            confidence: 0.95
        },
        score: 90
    }
}));
```

### 6. 查看聚合结果

1. 点击"结果展示"标签
2. 在"结果对比"区域选择任务
3. 点击"对比结果"按钮
4. 查看：
   - 各Agent的结果
   - 合并后的结果
   - 检测到的冲突
   - 置信度评分

## 📚 界面说明

### Agent管理

**统计卡片**：
- 总Agent数
- 活跃Agent（正在执行任务）
- 空闲Agent（可接受任务）
- 离线Agent（连接断开）

**Agent列表**：
- ID、名称、状态
- 能力标签
- 当前负载百分比
- 任务数（当前/最大）
- 最后心跳时间
- 操作按钮（详情/删除）

### 任务监控

**统计卡片**：
- 总任务数
- 待分配任务
- 执行中任务
- 已完成任务

**任务分配可视化**：
- 柱状图显示各Agent的任务分配情况
- 直观展示负载分布

**任务列表**：
- ID、类型、状态
- 优先级（1-10）
- 分配给哪个Agent
- 进度条
- 创建时间
- 操作按钮（详情）

### 结果展示

**统计卡片**：
- 总结果数
- 已验证结果
- 已聚合结果
- 平均置信度

**结果对比**：
- 选择任务查看所有Agent的结果
- 并排对比各结果的数据
- 查看合并后的最终结果
- 查看检测到的冲突及解决方案
- 查看置信度评分

**结果列表**：
- 结果ID、任务ID、Agent
- 状态、分数、置信度
- 创建时间
- 操作按钮（详情）

## 🔧 API文档

### Agent API

**列出所有Agent**
```
GET /api/agents
```

**注册Agent**
```
POST /api/agents
Content-Type: application/json

{
  "id": "agent-001",
  "name": "Agent Name",
  "capabilities": ["capability1", "capability2"],
  "max_tasks": 5
}
```

**获取Agent详情**
```
GET /api/agents/{agent_id}
```

**删除Agent**
```
DELETE /api/agents/{agent_id}
```

### Task API

**列出所有任务**
```
GET /api/tasks
```

**创建任务**
```
POST /api/tasks
Content-Type: application/json

{
  "id": "task-001",
  "type": "code_review",
  "priority": 8,
  "description": "Task description",
  "capabilities": ["code_review"]
}
```

**获取任务详情**
```
GET /api/tasks/{task_id}
```

### Result API

**列出所有结果**
```
GET /api/results
```

**获取结果详情**
```
GET /api/results/{result_id}
```

**获取聚合结果**
```
GET /api/results/aggregate/{task_id}
```

## 🎯 WebSocket消息协议

### 连接

```
ws://localhost:8080/ws?agent_id={agent_id}
```

### 消息格式

所有消息使用JSON格式：

```json
{
  "message_id": "msg-001",
  "type": "MESSAGE_TYPE",
  "from": "sender_id",
  "to": "receiver_id",
  "timestamp": "2026-01-29T10:00:00Z",
  "payload": {}
}
```

### 消息类型

**Agent注册**
```json
{
  "type": "AGENT_REGISTER",
  "payload": {
    "name": "Agent Name",
    "capabilities": ["cap1", "cap2"],
    "max_tasks": 5
  }
}
```

**心跳**
```json
{
  "type": "HEARTBEAT",
  "payload": {
    "status": "ACTIVE",
    "load": 0.5,
    "tasks_running": 2
  }
}
```

**任务请求（服务器发送）**
```json
{
  "type": "TASK_REQUEST",
  "payload": {
    "task_id": "task-001",
    "task_type": "code_review",
    "priority": 8,
    "description": "..."
  }
}
```

**任务结果提交**
```json
{
  "type": "TASK_RESULT",
  "payload": {
    "task_id": "task-001",
    "data": {
      "result": "...",
      "confidence": 0.95
    },
    "score": 90
  }
}
```

**任务状态更新**
```json
{
  "type": "TASK_STATUS",
  "payload": {
    "task_id": "task-001",
    "status": "RUNNING",
    "progress": 50
  }
}
```

### 广播事件

服务器会广播以下事件：

- `AGENT_REGISTERED` - Agent注册成功
- `AGENT_STATUS_UPDATE` - Agent状态更新
- `TASK_CREATED` - 任务创建
- `TASK_ASSIGNED` - 任务已分配
- `TASK_STATUS_UPDATE` - 任务状态更新
- `RESULT_SUBMITTED` - 结果已提交
- `RESULT_AGGREGATED` - 结果已聚合

## 💡 使用示例

### 完整工作流程

1. **启动服务器**
```bash
go run cmd/server/main.go
```

2. **打开Web界面**
```
http://localhost:8080
```

3. **注册2个Agent**
   - Agent-001: 能力 [code_review]
   - Agent-002: 能力 [code_review]

4. **创建任务**
   - 任务类型: code_review
   - 优先级: 8
   - 所需能力: code_review

5. **观察任务分配**
   - 系统自动分配任务给Agent-001
   - 在"任务监控"页面查看分配情况

6. **Agent提交结果**
   - Agent-001提交结果
   - Agent-002提交结果

7. **查看聚合结果**
   - 在"结果展示"页面选择任务
   - 点击"对比结果"
   - 查看合并结果和置信度

### 模拟Agent客户端

创建一个简单的Agent客户端：

```javascript
// agent-client.js
const WebSocket = require('ws');

const agentId = 'agent-001';
const ws = new WebSocket(`ws://localhost:8080/ws?agent_id=${agentId}`);

ws.on('open', () => {
    console.log('Connected to server');

    // 注册Agent
    ws.send(JSON.stringify({
        message_id: generateId(),
        type: 'AGENT_REGISTER',
        from: agentId,
        to: 'server',
        timestamp: new Date().toISOString(),
        payload: {
            name: 'Test Agent',
            capabilities: ['code_review', 'testing'],
            max_tasks: 5
        }
    }));

    // 定期发送心跳
    setInterval(() => {
        ws.send(JSON.stringify({
            message_id: generateId(),
            type: 'HEARTBEAT',
            from: agentId,
            to: 'server',
            timestamp: new Date().toISOString(),
            payload: {
                status: 'ACTIVE',
                load: 0.3,
                tasks_running: 1
            }
        }));
    }, 30000);
});

ws.on('message', (data) => {
    const msg = JSON.parse(data);
    console.log('Received:', msg);

    if (msg.type === 'TASK_REQUEST') {
        // 模拟任务执行
        setTimeout(() => {
            // 提交结果
            ws.send(JSON.stringify({
                message_id: generateId(),
                type: 'TASK_RESULT',
                from: agentId,
                to: 'server',
                timestamp: new Date().toISOString(),
                payload: {
                    task_id: msg.payload.task_id,
                    data: {
                        result: 'APPROVED',
                        issues_found: 0,
                        confidence: 0.95
                    },
                    score: 90
                }
            }));
        }, 3000);
    }
});

function generateId() {
    return 'msg-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
}
```

运行：
```bash
node agent-client.js
```

## 🎨 界面截图

（实际使用时会显示美观的现代化界面）

### Agent管理页面
- 统计卡片显示Agent概况
- 表格展示所有Agent的详细信息
- 能力标签彩色显示
- 实时更新心跳时间

### 任务监控页面
- 任务分配柱状图
- 任务列表带进度条
- 状态标签颜色区分

### 结果展示页面
- 结果对比卡片式布局
- 合并结果JSON格式展示
- 冲突检测红色高亮

## 🔗 相关模块

- [Communication Module](../../internal/communication/README.md)
- [Task Scheduler](../../internal/scheduler/README.md)
- [Result Aggregator](../../internal/aggregator/README.md)

---

**版本**: 1.0.0
**技术栈**: HTML5, CSS3, JavaScript, WebSocket
