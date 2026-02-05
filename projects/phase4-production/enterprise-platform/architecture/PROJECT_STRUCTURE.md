# 企业级Agent平台 - 项目结构

## 目录结构

```
enterprise-platform/
├── architecture/                   # 架构设计
│   ├── proto/                     # gRPC协议定义
│   │   ├── agent.proto           # Agent服务协议
│   │   ├── task.proto            # 任务服务协议
│   │   ├── tool.proto            # 工具服务协议
│   │   ├── user.proto            # 用户服务协议
│   │   ├── tenant.proto          # 租户服务协议
│   │   ├── cost.proto            # 成本服务协议
│   │   └── monitoring.proto      # 监控服务协议
│   ├── diagrams/                 # 架构图
│   └── decisions/                # 架构决策记录(ADR)
│
├── services/                       # 微服务实现
│   ├── agent/                     # Agent服务
│   │   ├── cmd/                  # 启动入口
│   │   ├── internal/             # 内部实现
│   │   │   ├── handler/         # gRPC处理器
│   │   │   ├── service/         # 业务逻辑
│   │   │   ├── repository/      # 数据访问
│   │   │   ├── model/           # 数据模型
│   │   │   └── llm/             # LLM客户端
│   │   ├── proto/                # 生成的gRPC代码
│   │   ├── config/               # 配置文件
│   │   ├── migrations/           # 数据库迁移
│   │   └── Dockerfile
│   │
│   ├── task/                      # 任务服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── queue/           # 任务队列
│   │   │   └── scheduler/       # 任务调度
│   │   ├── proto/
│   │   ├── config/
│   │   └── Dockerfile
│   │
│   ├── tool/                      # 工具服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── registry/        # 工具注册表
│   │   │   └── executor/        # 工具执行器
│   │   ├── proto/
│   │   ├── config/
│   │   └── Dockerfile
│   │
│   ├── user/                      # 用户服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── auth/            # 认证逻辑
│   │   │   └── jwt/             # JWT Token
│   │   ├── proto/
│   │   ├── config/
│   │   └── Dockerfile
│   │
│   ├── tenant/                    # 租户服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── quota/           # 配额管理
│   │   │   └── isolation/       # 数据隔离
│   │   ├── proto/
│   │   ├── config/
│   │   └── Dockerfile
│   │
│   ├── cost/                      # 成本服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── calculator/      # 成本计算
│   │   │   ├── forecast/        # 成本预测
│   │   │   └── alert/           # 成本告警
│   │   ├── proto/
│   │   ├── config/
│   │   └── Dockerfile
│   │
│   └── monitoring/                # 监控服务
│       ├── cmd/
│       ├── internal/
│       │   ├── handler/
│       │   ├── service/
│       │   ├── collector/       # 指标收集
│       │   └── health/          # 健康检查
│       ├── proto/
│       ├── config/
│       └── Dockerfile
│
├── pkg/                           # 共享库
│   ├── logger/                    # 日志库
│   ├── errors/                    # 错误处理
│   ├── tracing/                   # 分布式追踪
│   ├── metrics/                   # 指标收集
│   ├── middleware/                # gRPC中间件
│   ├── database/                  # 数据库工具
│   ├── cache/                     # 缓存工具
│   └── utils/                     # 通用工具
│
├── deploy/                        # 部署配置
│   ├── kubernetes/                # K8s配置
│   │   ├── base/                 # 基础配置
│   │   │   ├── agent-service.yaml
│   │   │   ├── task-service.yaml
│   │   │   ├── tool-service.yaml
│   │   │   ├── user-service.yaml
│   │   │   ├── tenant-service.yaml
│   │   │   ├── cost-service.yaml
│   │   │   └── monitoring-service.yaml
│   │   ├── overlays/             # 环境特定配置
│   │   │   ├── dev/
│   │   │   ├── staging/
│   │   │   └── production/
│   │   └── hpa/                  # 自动扩缩容配置
│   ├── docker-compose/            # Docker Compose配置
│   │   ├── docker-compose.yml
│   │   └── docker-compose.dev.yml
│   ├── helm/                      # Helm Charts
│   └── terraform/                 # 基础设施即代码
│
├── monitoring/                    # 监控配置
│   ├── prometheus/
│   │   ├── prometheus.yml        # Prometheus配置
│   │   ├── rules/                # 告警规则
│   │   └── targets/              # 监控目标
│   ├── grafana/
│   │   ├── dashboards/           # Grafana面板
│   │   └── datasources/          # 数据源配置
│   ├── alertmanager/
│   │   └── alertmanager.yml      # 告警管理配置
│   └── jaeger/
│       └── jaeger.yml            # Jaeger配置
│
├── logging/                       # 日志配置
│   ├── elasticsearch/
│   ├── logstash/
│   │   ├── logstash.conf
│   │   └── pipelines/
│   └── kibana/
│       └── dashboards/
│
├── scripts/                       # 脚本
│   ├── generate-proto.sh         # 生成gRPC代码
│   ├── deploy.sh                 # 部署脚本
│   ├── migrate-db.sh             # 数据库迁移
│   └── health-check.sh           # 健康检查
│
├── tests/                         # 测试
│   ├── integration/              # 集成测试
│   ├── e2e/                      # 端到端测试
│   ├── performance/              # 性能测试
│   └── stress/                   # 压力测试
│
├── docs/                          # 文档
│   ├── api/                      # API文档
│   ├── architecture/             # 架构文档
│   ├── deployment/               # 部署文档
│   ├── development/              # 开发文档
│   └── operations/               # 运维文档
│
├── go.mod                         # Go模块定义
├── go.sum
├── Makefile                       # Make命令
├── README.md
└── .gitignore
```

## 核心目录说明

### 1. services/

每个微服务遵循相同的结构：

```
service-name/
├── cmd/                    # 程序入口
│   └── main.go            # 启动文件
├── internal/               # 内部实现（不可导出）
│   ├── handler/           # gRPC handler实现
│   ├── service/           # 业务逻辑层
│   ├── repository/        # 数据访问层
│   └── model/             # 数据模型
├── proto/                  # 生成的gRPC代码
├── config/                 # 配置文件
│   ├── config.yaml        # 默认配置
│   ├── config.dev.yaml    # 开发环境
│   └── config.prod.yaml   # 生产环境
├── migrations/             # 数据库迁移文件
│   ├── 001_initial.up.sql
│   └── 001_initial.down.sql
├── Dockerfile              # Docker镜像构建
├── go.mod                  # 服务依赖
└── README.md               # 服务文档
```

### 2. pkg/

共享库，可被所有服务导入：

```go
// 使用示例
import (
    "github.com/agent-learning/enterprise-platform/pkg/logger"
    "github.com/agent-learning/enterprise-platform/pkg/tracing"
    "github.com/agent-learning/enterprise-platform/pkg/metrics"
)
```

### 3. deploy/

部署配置，支持多种部署方式：

- **kubernetes/**: Kubernetes原生配置
- **docker-compose/**: 本地开发部署
- **helm/**: Helm包管理
- **terraform/**: 云基础设施

### 4. monitoring/

监控体系配置：

- **Prometheus**: 指标收集
- **Grafana**: 可视化面板
- **Alertmanager**: 告警管理
- **Jaeger**: 分布式追踪

### 5. logging/

ELK日志栈配置：

- **Elasticsearch**: 日志存储和搜索
- **Logstash**: 日志处理管道
- **Kibana**: 日志可视化

---

## 开发流程

### 1. 生成gRPC代码

```bash
make proto-gen
```

### 2. 启动开发环境

```bash
make dev-up
```

### 3. 运行测试

```bash
make test
```

### 4. 构建镜像

```bash
make docker-build
```

### 5. 部署到K8s

```bash
make k8s-deploy ENV=dev
```

---

## 环境变量

每个服务支持以下环境变量：

```bash
# 服务配置
SERVICE_NAME=agent-service
SERVICE_PORT=8080
GRPC_PORT=9090

# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_NAME=agent_db
DB_USER=agent
DB_PASSWORD=secret

# Redis配置
REDIS_HOST=redis
REDIS_PORT=6379

# Consul配置
CONSUL_HOST=consul
CONSUL_PORT=8500

# 监控配置
PROMETHEUS_PORT=9091
JAEGER_ENDPOINT=http://jaeger:14268

# 日志配置
LOG_LEVEL=info
LOG_FORMAT=json
```

---

## 技术栈版本

| 技术 | 版本 |
|------|------|
| Go | 1.21+ |
| gRPC | 1.60+ |
| PostgreSQL | 15+ |
| Redis | 7+ |
| Consul | 1.17+ |
| Prometheus | 2.48+ |
| Grafana | 10.2+ |
| Jaeger | 1.52+ |
| Elasticsearch | 8.11+ |
| Kubernetes | 1.28+ |
| Docker | 24+ |

---

**版本**: v1.0.0
**日期**: 2026-01-30
