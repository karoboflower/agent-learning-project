# Go Agent API 快速启动指南

## 📋 前置要求

- Go 1.21+
- Docker和Docker Compose (可选，用于Redis和PostgreSQL)
- OpenAI API Key

## 🚀 快速开始

### 1. 克隆项目

```bash
cd agent-learning-project/projects/phase2-practice/go-agent-api
```

### 2. 配置环境变量

```bash
# 复制示例配置
cp .env.example .env

# 编辑 .env 文件，填入必要配置
vim .env
```

**最小配置**:
```env
OPENAI_API_KEY=your_openai_api_key_here
SERVER_PORT=8080
```

### 3. 启动依赖服务 (可选)

如果你想使用Redis和PostgreSQL:

```bash
# 启动Docker服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

或使用Makefile:
```bash
make docker-up
```

### 4. 下载Go依赖

```bash
go mod download
go mod tidy
```

或使用Makefile:
```bash
make install
```

### 5. 运行服务

```bash
go run cmd/server/main.go
```

或使用Makefile:
```bash
make run
```

### 6. 测试服务

打开新终端，测试健康检查:

```bash
curl http://localhost:8080/health
```

预期响应:
```json
{
  "status": "ok",
  "service": "go-agent-api"
}
```

## 🧪 API测试

### 创建Agent

```bash
curl -X POST http://localhost:8080/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-first-agent",
    "type": "general",
    "config": {
      "model": "gpt-4",
      "temperature": 0.7,
      "max_tokens": 2000
    }
  }'
```

保存返回的agent ID，例如: `agent-123`

### 提交任务

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent-123",
    "type": "query",
    "input": "What is Go programming language?",
    "priority": 1
  }'
```

保存返回的task ID，例如: `task-456`

### 查询任务状态

```bash
curl http://localhost:8080/api/v1/tasks/task-456
```

### 获取任务结果

```bash
curl http://localhost:8080/api/v1/tasks/task-456/result
```

### 获取统计信息

```bash
curl http://localhost:8080/api/v1/tasks/stats
```

## 📊 Makefile命令

```bash
make help           # 显示所有可用命令
make install        # 安装依赖
make build          # 编译应用
make run            # 运行应用
make test           # 运行测试
make test-coverage  # 运行测试并生成覆盖率报告
make clean          # 清理构建产物
make docker-up      # 启动Docker服务
make docker-down    # 停止Docker服务
make docker-logs    # 查看Docker日志
make fmt            # 格式化代码
make vet            # 运行go vet
make dev            # 启动开发环境(Docker + 服务)
```

## 🔍 常见问题

### Q: 服务启动失败

**A**: 检查以下几点:
1. OpenAI API Key是否正确配置
2. 端口8080是否被占用
3. 查看错误日志

### Q: Redis连接失败

**A**: 服务会自动降级到内存模式，不影响核心功能。如需Redis:
```bash
docker-compose up -d redis
```

### Q: PostgreSQL连接失败

**A**: 数据库是可选的，不影响服务运行。如需数据持久化:
```bash
docker-compose up -d postgres
```

### Q: 任务一直处于pending状态

**A**: 检查:
1. Agent是否创建成功
2. 调度器是否正常启动
3. OpenAI API是否可访问

## 📝 开发建议

### 推荐开发流程

1. 启动依赖服务:
   ```bash
   make docker-up
   ```

2. 运行服务:
   ```bash
   make run
   ```

3. 在另一个终端测试API:
   ```bash
   ./scripts/test-api.sh
   ```

### 代码规范

运行格式化和检查:
```bash
make fmt
make vet
```

### 测试

运行所有测试:
```bash
make test
```

生成覆盖率报告:
```bash
make test-coverage
```

## 🎯 下一步

- 阅读 [README.md](README.md) 了解完整功能
- 阅读 [Task-2.3-README.md](Task-2.3-README.md) 了解实现细节
- 查看 [API文档](#) (待生成Swagger文档)

## 💡 提示

- 首次运行可能需要下载Go依赖，请耐心等待
- 建议使用Go 1.21或更高版本
- 如遇到依赖问题，运行 `go mod tidy`
- 生产环境建议配置Redis和PostgreSQL
