# Go Agent API Service

> Go后端的Agent API服务，支持多Agent协作、任务调度和工具调用

[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-1.9-green.svg)](https://gin-gonic.com/)
[![Status](https://img.shields.io/badge/Status-In%20Progress-yellow.svg)]()

## ✨ 功能特性

- 🤖 **Agent服务** - 支持多Agent实例管理
- 📋 **任务调度** - 智能任务分发和优先级管理
- 🔧 **工具调用** - 灵活的工具注册和调用机制
- 💾 **状态管理** - Redis状态存储和同步
- 📊 **任务历史** - PostgreSQL持久化存储
- 🌐 **RESTful API** - 完整的HTTP接口
- 📖 **API文档** - Swagger自动生成文档

## 🚀 快速开始

### 前置条件

- Go 1.21+
- Redis
- PostgreSQL
- OpenAI API Key

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，填入配置
```

### 3. 启动服务

```bash
go run cmd/server/main.go
```

### 4. 访问API

```
API Base URL: http://localhost:8080
Swagger Docs: http://localhost:8080/swagger/index.html
```

## 📦 项目结构

```
go-agent-api/
├── cmd/
│   └── server/
│       └── main.go              # 应用入口
├── internal/
│   ├── agent/                   # Agent核心
│   │   ├── agent.go            # Agent接口和实现
│   │   ├── registry.go         # Agent注册表
│   │   └── types.go            # Agent类型定义
│   ├── scheduler/               # 任务调度
│   │   ├── scheduler.go        # 调度器实现
│   │   ├── queue.go            # 任务队列
│   │   └── types.go            # 任务类型定义
│   ├── state/                   # 状态管理
│   │   ├── redis.go            # Redis状态存储
│   │   └── manager.go          # 状态管理器
│   ├── tools/                   # 工具系统
│   │   ├── tool.go             # 工具接口
│   │   ├── registry.go         # 工具注册
│   │   ├── search.go           # 搜索工具
│   │   ├── code.go             # 代码工具
│   │   └── file.go             # 文件工具
│   ├── api/                     # API接口
│   │   ├── handlers/           # 处理器
│   │   ├── middleware/         # 中间件
│   │   └── routes.go           # 路由定义
│   ├── database/                # 数据库
│   │   ├── postgres.go         # PostgreSQL连接
│   │   └── models.go           # 数据模型
│   └── config/                  # 配置
│       └── config.go           # 配置加载
├── pkg/                         # 公共包
│   └── utils/                  # 工具函数
├── docs/                        # Swagger文档
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── README.md
```

## 🎯 核心功能

### 1. Agent管理

创建和管理多个Agent实例：

```go
// 创建Agent
POST /api/v1/agents
{
  "name": "code-assistant",
  "type": "general",
  "config": {
    "model": "gpt-4",
    "temperature": 0.7
  }
}

// 获取Agent列表
GET /api/v1/agents

// 获取Agent详情
GET /api/v1/agents/:id

// 删除Agent
DELETE /api/v1/agents/:id
```

### 2. 任务管理

提交和管理任务：

```go
// 提交任务
POST /api/v1/tasks
{
  "agent_id": "agent-uuid",
  "type": "code_review",
  "input": "...",
  "priority": 1
}

// 获取任务状态
GET /api/v1/tasks/:id

// 获取任务结果
GET /api/v1/tasks/:id/result

// 取消任务
DELETE /api/v1/tasks/:id
```

### 3. 工具调用

Agent可调用的工具：

- 🔍 **搜索工具** - 网络搜索和信息检索
- 💻 **代码工具** - 代码执行和分析
- 📁 **文件工具** - 文件读写和操作

### 4. 状态管理

- Agent状态实时同步到Redis
- 任务历史持久化到PostgreSQL
- 支持状态恢复和容错

## 🏗️ 技术架构

### Agent执行流程

```
客户端请求
    ↓
API接口层
    ↓
任务调度器
    ↓
Agent执行引擎
    ↓
工具调用系统
    ↓
结果返回
```

### 数据流

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  API Layer  │
└──────┬──────┘
       │
       ▼
┌─────────────┐      ┌─────────────┐
│  Scheduler  │◄────►│    Redis    │ (状态)
└──────┬──────┘      └─────────────┘
       │
       ▼
┌─────────────┐      ┌─────────────┐
│    Agent    │◄────►│  PostgreSQL │ (历史)
└──────┬──────┘      └─────────────┘
       │
       ▼
┌─────────────┐
│    Tools    │
└─────────────┘
```

## 🔐 环境变量

| 变量名 | 说明 | 必需 | 默认值 |
|--------|------|------|--------|
| `SERVER_PORT` | 服务端口 | ❌ | 8080 |
| `OPENAI_API_KEY` | OpenAI API密钥 | ✅ | - |
| `OPENAI_MODEL` | OpenAI模型 | ❌ | gpt-4 |
| `REDIS_HOST` | Redis主机 | ❌ | localhost |
| `REDIS_PORT` | Redis端口 | ❌ | 6379 |
| `POSTGRES_HOST` | PostgreSQL主机 | ❌ | localhost |
| `POSTGRES_PORT` | PostgreSQL端口 | ❌ | 5432 |
| `MAX_CONCURRENT_AGENTS` | 最大并发Agent数 | ❌ | 10 |

## 📖 API文档

### Agent API

#### 创建Agent

```http
POST /api/v1/agents
Content-Type: application/json

{
  "name": "my-agent",
  "type": "general",
  "config": {
    "model": "gpt-4",
    "temperature": 0.7,
    "max_tokens": 2000
  }
}
```

**响应**：
```json
{
  "id": "agent-uuid",
  "name": "my-agent",
  "type": "general",
  "status": "idle",
  "created_at": "2026-01-28T10:00:00Z"
}
```

### Task API

#### 提交任务

```http
POST /api/v1/tasks
Content-Type: application/json

{
  "agent_id": "agent-uuid",
  "type": "query",
  "input": "What is Go?",
  "priority": 1,
  "tools": ["search", "code"]
}
```

**响应**：
```json
{
  "id": "task-uuid",
  "agent_id": "agent-uuid",
  "status": "pending",
  "created_at": "2026-01-28T10:00:00Z"
}
```

## 🧪 测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试
go test -tags=integration ./...

# 查看测试覆盖率
go test -cover ./...
```

## 🛣️ 开发路线图

### 已完成 ✅
- [ ] Task 2.3.1 - 项目初始化
- [ ] Task 2.3.2 - Agent服务接口
- [ ] Task 2.3.3 - 任务调度器
- [ ] Task 2.3.4 - 状态管理
- [ ] Task 2.3.5 - 工具调用机制
- [ ] Task 2.3.6 - API接口
- [ ] Task 2.3.7 - 数据库集成和优化

### 计划中 📋
- [ ] WebSocket实时通信
- [ ] 多Agent协作
- [ ] 更多工具集成
- [ ] 性能监控和追踪
- [ ] Docker部署

## 🤝 贡献

欢迎提交Issue和Pull Request！

## 📝 许可证

MIT License

---

**完成日期**: 2026-01-28
**版本**: v1.0.0 (In Progress)
**状态**: 🚧 开发中
