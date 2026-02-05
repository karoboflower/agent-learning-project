# Task 4.1.7 完成 - 实现日志系统

**完成日期**: 2026-01-30
**任务**: 实现日志系统（Day 18-20）

---

## ✅ 完成内容

### 1. 结构化日志库 ✅

**文件**: `monitoring/logging/internal/logger/logger.go` (~420行)

**核心功能**:

```
Logger - 结构化日志器
├── 基于zap高性能日志库
├── JSON/Console两种输出格式
├── 5个日志级别（Debug/Info/Warn/Error/Fatal）
├── 结构化字段支持
├── 上下文传播
└── 全局单例

日志配置
├── 日志级别
├── 输出路径
├── 编码格式（JSON/Console）
├── 是否启用调用��信息
├── 是否启用堆栈跟踪
├── 服务名称和版本
└── 环境（development/staging/production）

字段类型
├── String - 字符串
├── Int/Int64 - 整数
├── Float64 - 浮点数
├── Bool - 布尔
├── Duration - 时长
├── Error - 错误
└── Any - 任意类型

上下文字段自动提取
├── request_id - 请求ID
├── tenant_id - 租户ID
├── user_id - 用户ID
├── trace_id - 追踪ID
└── span_id - Span ID
```

**使用示例**:
```go
// 创建日���器
config := &logger.Config{
    Level:       logger.InfoLevel,
    Environment: "production",
    Encoding:    "json",
    ServiceName: "agent-service",
}
log, _ := logger.NewLogger(config)

// 记录日志
log.Info("Agent execution started",
    logger.String("agent_id", "agent-001"),
    logger.String("tenant_id", "tenant-123"),
    logger.Duration("timeout", 30*time.Second),
)

// 带上下文
log.WithContext(ctx).Error("Execution failed",
    logger.Error(err),
)

// 格式化日志
log.Infof("Processing %d tasks", taskCount)
```

**日志输出示例**:
```json
{
  "timestamp": "2026-01-30T10:15:30.123Z",
  "level": "info",
  "logger": "agent-service",
  "message": "Agent execution started",
  "service": "agent-service",
  "version": "1.0.0",
  "env": "production",
  "agent_id": "agent-001",
  "tenant_id": "tenant-123",
  "request_id": "20260130101530-abc12345",
  "timeout": 30,
  "caller": "agent/executor.go:42"
}
```

### 2. 日志中间件 ✅

**文件**: `monitoring/logging/internal/middleware/logging_middleware.go` (~220行)

**核心组件**:

```
HTTP日志中间件
├── 自动记录请求开始/完成
├── 请求ID生成和传播
├── 记录请求方法、路径、状态码
├── 记录响应时间和大小
└── 上下文注入

gRPC一元调用拦截器
├── 记录方法调用
├── 从metadata提取信息
├── 记录执行时间
├── 错误日志记录
└── 状态码记录

gRPC流式调用拦截器
├── 记录流开始/完成
├── 标识客户端/服务端流
├── 记录流执行时间
└── 错误处理

恢复中间件
├── 捕获panic
├── 记录panic信息
├── 返回500错误
└── 防止服务崩溃
```

**使用示例**:
```go
// HTTP中间件
log := logger.GetGlobalLogger()
mux := http.NewServeMux()
handler := middleware.HTTPLoggingMiddleware(log)(mux)
handler = middleware.RecoveryMiddleware(log)(handler)

http.ListenAndServe(":8080", handler)

// gRPC拦截器
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(middleware.GRPCUnaryLoggingInterceptor(log)),
    grpc.StreamInterceptor(middleware.GRPCStreamLoggingInterceptor(log)),
)
```

### 3. Elasticsearch配置 ✅

**文件**: `monitoring/logging/elasticsearch/elasticsearch.yml` (~100行)

**核心配置**:

```
集群配置
├── 集群名称: enterprise-platform-logs
├── 节点名称: es-node-1
├── 发现模式: single-node
└── 安全配置: 关闭（开发环境）

索引模板（app-logs-*）
├── 3个主分片
├── 1个副本
├── 5秒刷新间隔
├── 异步translog
└── ILM策略关联

字段映射
├── timestamp (date) - 时间戳
├── level (keyword) - 日志级别
├── message (text) - 消息内容
├── service (keyword) - 服务名称
├── tenant_id (keyword) - 租户ID
├── request_id (keyword) - 请求ID
├── trace_id (keyword) - 追踪ID
├── method (keyword) - HTTP方法
├── path (keyword) - 请求路径
├── status (integer) - 状态码
├── duration (long) - 执行时长
├── error (text) - 错误信息
└── stacktrace (text) - 堆栈跟踪

索引生命周期策略
├── Hot阶段（0天）: 自动滚动（50GB/1天/1亿文档）
├── Warm阶段（7天）: 强制合并+收缩
├── Cold阶段（30天）: 冻结
└── Delete阶段（90天）: 删除
```

**生命周期示意**:
```
Day 0-7:   Hot (写入+查询) → 50GB或1天后滚动
Day 7-30:  Warm (只读+优化) → 强制合并为1段
Day 30-90: Cold (归档) → 冻结索引
Day 90+:   Delete → 自动删除
```

### 4. Logstash配置 ✅

**文件**: `monitoring/logging/logstash/pipeline.conf` (~180行)

**数据流**:

```
Input（3种输入源）
├── Beats (5044端口) - 从Filebeat接收
├── TCP (5000端口) - 直接TCP输入
└── HTTP (8080端口) - HTTP API输入

Filter（12个过滤器）
├── JSON解析 - 解析JSON日志
├── 字段提升 - 提升到顶层
├── 时间戳解析 - ISO8601格式
├── 日志级别标准化 - 转小写
├── 地理位置信息 - GeoIP
├── 类型转换 - status/duration
├── 标签添加 - error/warning
├── 敏感信息脱敏 - password/token/secret
├── 错误类型提取 - error_class
├── 处理时间戳 - processing_time
├── 元数据添加 - index_prefix
└── Ruby代码处理 - 自定义逻辑

Output（3种输出）
├── Elasticsearch - 主输出（app-logs-YYYY.MM.DD）
├── Elasticsearch - 错误日志（error-logs-YYYY.MM.DD）
└── Kafka（可选） - 消息队列
```

**脱敏规则**:
```
password="secret123" → password=***REDACTED***
token="abc123xyz"    → token=***REDACTED***
secret="mykey"       → secret=***REDACTED***
```

### 5. Filebeat配置 ✅

**文件**: `monitoring/logging/filebeat.yml` (~120行)

**输入源**:

```
应用日志
├── 路径: /var/log/enterprise-platform/**/*.log
├── JSON解析
├── 多行合并
├── 字段添加（log_type, environment）
└── 每10秒扫描

Docker容器日志
├── 路径: /var/lib/docker/containers/*/*.log
├── Docker元数据
├── JSON解析
└── 容器信息

Kubernetes Pod日志（可选）
├── Hints自动发现
├── Kubernetes元数据
└── Pod/Container信息
```

**处理器**:
```
add_host_metadata - 添加主机信息
add_docker_metadata - 添加Docker元数据
add_kubernetes_metadata - 添加K8s元数据
drop_fields - 删除不需要的字段
rename - 字段重命名
add_tags - 添加标签
```

**输出配置**:
- 主输出：Logstash (5044端口)
- 备选：直接输出到Elasticsearch
- 备选：输出到Kafka

### 6. Kibana配置和Dashboard ✅

**文件**: `monitoring/logging/kibana/kibana.yml` (~160行)

**核心组件**:

```
索引模式
└── app-logs-* (时间字段: @timestamp)

保存的搜索
└── Error Logs - 错误日志搜索

可视化（3个）
├── Log Level Distribution - 日志级别分布（饼图）
├── Service Request Rate - 服务请求率（折线图）
└── Error Trend - 错误趋势（面积图）

Dashboard
├── Application Logs Overview
├── 13个面板
├── 30秒自动刷新
└── 最近24小时

Watcher告警
├── 错误率告警
├── 5分钟检查间隔
├── 阈值: 10个错误/5分钟
├── Email + Webhook通知
└── 自动触发
```

**Dashboard面板**:
1. 日志级别分布
2. 错误趋势
3. 服务请求率
4. Top错误
5. 慢查询
6. 租户请求分布
7. API响应时间
8. 错误堆栈
9. 请求方法分布
10. 状态码分布
11. 地理位置分布
12. 日志时间线
13. 最近错误

### 7. Fluentd配置 ✅

**文件**: `monitoring/logging/fluentd/fluent.conf` (~180行)

**Fluentd特性**（替代Filebeat+Logstash）:

```
输入源（4种）
├── tail - 文件尾随
├── forward - 转发（24224端口）
├── http - HTTP API（9880端口）
└── syslog - 系统日志（5140端口）

过滤器（7个）
├── record_transformer - 添加字段
├── parser - JSON解析
├── record_modifier - 脱敏
├── kubernetes_metadata - K8s元数据
├── record_transformer - 日志级别标准化
├── geoip - 地理位置
└── 自定义Ruby代码

输出（4种）
├── Elasticsearch - 主输出
├── Elasticsearch - 错误日志
├── Kafka（可选）
└── S3归档（可选）

监控
├── monitor_agent (24220端口) - 监控API
├── prometheus (24231端口) - Prometheus指标
└── prometheus_output_monitor - 输出监控
```

**优势对比**:
```
Fluentd方案:
✅ 单一组件，配置简单
✅ 内存占用更小（Ruby实现）
✅ 插件丰富（500+）
✅ 原生Kubernetes支持
✅ 统一的配置格式

Filebeat+Logstash方案:
✅ Elastic官方支持
✅ 更好的Elasticsearch集成
✅ Logstash丰富的过滤器
✅ Beats生态系统
```

### 8. 日志聚合器 ✅

**文件**: `monitoring/logging/internal/aggregator/log_aggregator.go` (~380行)

**核心功能**:

```
LogAggregator - 日志聚合器
├── 基于Elasticsearch客户端
├── 日志搜索和查询
├── 多维度聚合
├── 统计分析
└── 健康检查

搜索功能
├── 时间范围过滤
├── 日志级别过滤
├── 服务过滤
├── 租户过滤
├── 用户过滤
├── 请求ID过滤
├── 消息全文搜索
├── 分页和排序
└── 高亮显示

聚合功能
├── 按日志级别聚合
├── 按服务聚合
├── 按时间聚合（时间线）
├── 错误统计
├── Top N查询
└── 多维度组合

特殊查询
├── 根据请求ID获取完整日志链路
├── 根据追踪ID获取分布式追踪日志
├── 错误聚类分析
└── 慢查询分析

管理功能
├── 删除旧日志
├── 索引优化
├── 健康检查
└── 连接管理
```

**使用示例**:
```go
// 创建聚合器
aggregator, _ := aggregator.NewLogAggregator("http://localhost:9200")

// 搜索日志
query := &aggregator.LogQuery{
    StartTime: time.Now().Add(-1 * time.Hour),
    EndTime:   time.Now(),
    Level:     "error",
    Service:   "agent-service",
    Limit:     100,
}
logs, total, _ := aggregator.SearchLogs(ctx, query)

// 按级别聚合
stats, _ := aggregator.AggregateByLevel(ctx, startTime, endTime)
// 结果: {"error": 150, "warn": 320, "info": 5000}

// 时间线聚合
timeline, _ := aggregator.AggregateTimeline(ctx, startTime, endTime, "1h")
// 结果: [{time: "10:00", count: 1200}, {time: "11:00", count: 1350}, ...]

// 根据请求ID获取完整链路
logs, _ := aggregator.GetLogsByRequestID(ctx, "request-123")
// 返回该请求的所有日志，按时间排序

// 错误统计
errorStats, _ := aggregator.GetErrorStats(ctx, startTime, endTime)
// 结果: {total: 150, by_service: {...}, by_error_type: {...}}
```

### 9. Docker Compose日志栈 ✅

**文件**: `monitoring/logging/docker-compose.elk.yml` (~100行)

**组件**（5个）:

```
Elasticsearch
├── 镜像: 8.10.2
├── 单节点模式
├── 512MB堆内存
├── 端口: 9200, 9300
└── 持久化存储

Logstash
├── 镜像: 8.10.2
├── 256MB堆内存
├── 端口: 5000, 5044, 9600
├── Pipeline配置
└── 依赖Elasticsearch

Kibana
├── 镜像: 8.10.2
├── 端口: 5601
├── Dashboard配置
├── 依赖Elasticsearch
└── 中文界面

Filebeat
├── 镜像: 8.10.2
├── 读取Docker容器日志
├── 发送到Logstash
└── 依赖Logstash

Fluentd（可选）
├── 镜像: v1.16-1
├── 端口: 24224, 9880, 24231
├── Prometheus指标
└── 依赖Elasticsearch
```

**启动命令**:
```bash
# 启动ELK栈
cd monitoring/logging
docker-compose -f docker-compose.elk.yml up -d

# 查看状态
docker-compose -f docker-compose.elk.yml ps

# 查看日志
docker-compose -f docker-compose.elk.yml logs -f elasticsearch

# 停止
docker-compose -f docker-compose.elk.yml down
```

### 10. 日志归档和清理 ✅

**文件**: `monitoring/logging/scripts/archive-logs.sh` (~180行)

**功能**:

```
归档策略
├── 检查Elasticsearch连接
├── 强制合并旧索引（提高压缩率）
├── 创建快照到S3
├── 删除过期索引
└── 生成归档报告

配置参数
├── ARCHIVE_DAYS=30 - 归档阈值
├── DELETE_DAYS=90 - 删除阈值
├── S3_BUCKET - S3存储桶
└── S3_REGION - S3区域

执行流程
1. 健康检查
2. 强制合并（>7天的索引）
3. S3快照归档（>30天的索引）
4. 删除索引（>90天的索引）
5. 生成报告
```

**Cron任务**:
```cron
# 每天凌晨2点归档
0 2 * * * /opt/.../archive-logs.sh

# 每周日凌晨3点优化
0 3 * * 0 curl -X POST "http://localhost:9200/_optimize"

# 每月1号生成报告
0 4 1 * * /opt/.../generate-monthly-report.sh

# 每小时健康检查
0 * * * * curl -sf "http://localhost:9200/_cluster/health"
```

---

## 🎯 核心亮点

### 1. 完整的日志链路

```
应用代码
    ↓ (结构化日志库)
日志文件/stdout
    ↓ (Filebeat/Fluentd)
Logstash
    ↓ (解析+过滤+脱敏)
Elasticsearch
    ↓ (存储+索引)
Kibana
    ↓ (可视化+告警)
用户
```

### 2. 多种部署方案

**方案A: Filebeat + Logstash + Elasticsearch + Kibana (ELK)**
```
优势: Elastic官方支持，功能强大
劣势: 组件多，资源占用较高
适用: 大规模生产环境
```

**方案B: Fluentd + Elasticsearch + Kibana (FEK)**
```
优势: 配置简单，资源占用低
劣势: Logstash功能更丰富
适用: 中小规模环境
```

**方案C: Loki + Promtail + Grafana**
```
优势: 轻量级，成本低
劣势: 功能相对简单
适用: 日志量不大的环境
```

### 3. 智能日志管理

**索引生命周期**:
```
Day 0-7:   Hot → 频繁写入和查询
Day 7-30:  Warm → 只读，优化存储
Day 30-90: Cold → 归档，冻结
Day 90+:   Delete → 自动删除
```

**存储优化**:
- 自动滚动：50GB或1天
- 强制合并：1个段
- 索引收缩：减少分片
- 快照备份：S3归档

**成本估算**:
```
假设日志量：100GB/天

热数据（7天）:   700GB × $0.10/GB = $70
温数据（23天）:  2.3TB × $0.05/GB = $115
冷数据（60天）:  6TB × $0.01/GB = $60
总成本: $245/月

S3归档: 6TB × $0.004/GB = $24/月
合计: $269/月
```

### 4. 强大的查询能力

**全文搜索**:
```
message:"connection timeout"
```

**复杂查询**:
```
level:error AND service:agent-service AND tenant_id:tenant-123
```

**时间范围**:
```
@timestamp:[now-1h TO now]
```

**通配符**:
```
path:/api/*/execute
```

**正则表达式**:
```
error:/timeout|connection.*failed/
```

**聚合查询**:
```
按服务分组，计算错误率
按时间分组，绘制趋势图
Top 10错误类型
95分位响应时间
```

### 5. 敏感信息保护

**自动脱敏**:
```
原始: "User password is abc123"
脱敏: "User password is ***REDACTED***"

原始: "token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
脱敏: "token=***REDACTED***"
```

**字段加密**（可扩展）:
- 用户密码
- API密钥
- OAuth令牌
- 信用卡号
- 身份证号

### 6. 分布式追踪集成

**追踪ID传播**:
```
HTTP Request
├── trace_id: abc123xyz
├── span_id: span-001
└── request_id: req-20260130

日志自动关联
├── Agent日志: trace_id=abc123xyz
├── Task日志: trace_id=abc123xyz
├── Tool日志: trace_id=abc123xyz
└── Cost日志: trace_id=abc123xyz

一键查询完整链路
└── aggregator.GetLogsByTraceID("abc123xyz")
```

---

## 📊 日志架构

```
┌─────────────────────────────────────────────────┐
│              应用服务层                          │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │ Agent   │ │ Task    │ │ Cost    │ ...       │
│  └────┬────┘ └────┬────┘ └────┬────┘           │
│       │ 结构化日志  │           │                │
└───────┼───────────┼───────────┼────────────────┘
        │           │           │
        ▼           ▼           ▼
┌─────────────────────────────────────────────────┐
│            日志收集层                            │
│  ┌──────────┐              ┌──────────┐        │
│  │ Filebeat │              │ Fluentd  │        │
│  └─────┬────┘              └─────┬────┘        │
└────────┼─────────────────────────┼─────────────┘
         │                         │
         ▼                         ▼
┌─────────────────────────────────────────────────┐
│            日志处理层                            │
│  ┌──────────────────────────────────────┐      │
│  │         Logstash Pipeline            │      │
│  │  ├─ JSON解析                         │      │
│  │  ├─ 字段转换                         │      │
│  │  ├─ 敏感信息脱敏                     │      │
│  │  ├─ 地理位置添加                     │      │
│  │  └─ 标签和元数据                     │      │
│  └──────────────┬───────────────────────┘      │
└─────────────────┼──────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────┐
│            日志存储层                            │
│  ┌────────────────────────────────────┐        │
│  │        Elasticsearch Cluster       │        │
│  │  ┌────────────────────────────┐    │        │
│  │  │ app-logs-2026.01.30 (Hot)  │    │        │
│  │  │ app-logs-2026.01.23 (Warm) │    │        │
│  │  │ app-logs-2025.12.30 (Cold) │    │        │
│  │  └────────────────────────────┘    │        │
│  └────────────┬───────────────────────┘        │
└───────────────┼──────────────────────────────────┘
                │
      ┌─────────┴─────────┐
      │                   │
      ▼                   ▼
┌─────────────┐   ┌──────────────┐
│   Kibana    │   │ Log          │
│  Dashboard  │   │ Aggregator   │
│   + Alert   │   │    API       │
└─────────────┘   └──────────────┘
      │                   │
      └─────────┬─────────┘
                ▼
           ┌─────────┐
           │  Users  │
           └─────────┘
```

---

## 🔧 使用指南

### 1. 启动日志系统

```bash
# 进入日志目录
cd monitoring/logging

# 启动ELK栈
docker-compose -f docker-compose.elk.yml up -d

# 等待服务启动（约2分钟）
docker-compose -f docker-compose.elk.yml ps

# 检查Elasticsearch
curl http://localhost:9200/_cluster/health

# 访问Kibana
open http://localhost:5601
```

### 2. 应用中集成日志

```go
package main

import (
    "github.com/agent-learning/enterprise-platform/monitoring/logging/internal/logger"
    "github.com/agent-learning/enterprise-platform/monitoring/logging/internal/middleware"
)

func main() {
    // 配置日志
    config := &logger.Config{
        Level:       logger.InfoLevel,
        Environment: "production",
        Encoding:    "json",
        OutputPaths: []string{
            "stdout",
            "/var/log/enterprise-platform/agent-service.log",
        },
        ServiceName:    "agent-service",
        ServiceVersion: "1.0.0",
    }

    log, _ := logger.NewLogger(config)
    logger.SetGlobalLogger(log)

    // HTTP服务器
    mux := http.NewServeMux()

    // 添加日志中间件
    handler := middleware.HTTPLoggingMiddleware(log)(mux)
    handler = middleware.RecoveryMiddleware(log)(handler)

    // 业务处理
    mux.HandleFunc("/api/execute", func(w http.ResponseWriter, r *http.Request) {
        // 使用带上下文的日志
        log.WithContext(r.Context()).Info("Executing agent",
            logger.String("agent_id", "agent-001"),
        )

        // 业务逻辑...

        log.WithContext(r.Context()).Info("Execution completed")
    })

    http.ListenAndServe(":8080", handler)
}
```

### 3. 在Kibana中查询日志

```
1. 打开Kibana: http://localhost:5601

2. 创建索引模式:
   - Management → Index Patterns → Create
   - 索引模式: app-logs-*
   - 时间字段: @timestamp

3. 搜索日志:
   - Discover → 选择 app-logs-*
   - 查询示例:
     * level:error
     * service:agent-service AND tenant_id:tenant-123
     * message:"timeout" AND @timestamp:[now-1h TO now]

4. 创建可视化:
   - Visualize → Create → 选择类型
   - 配置数据源和聚合
   - 保存可视化

5. 创建Dashboard:
   - Dashboard → Create → Add panels
   - 添加保存的可视化
   - 配置布局和刷新间隔
```

### 4. 使用日志聚合API

```go
// 创建聚合器
aggregator, _ := aggregator.NewLogAggregator("http://localhost:9200")

// 搜索错误日志
query := &aggregator.LogQuery{
    StartTime: time.Now().Add(-24 * time.Hour),
    EndTime:   time.Now(),
    Level:     "error",
    Service:   "agent-service",
    Limit:     100,
}
logs, total, _ := aggregator.SearchLogs(ctx, query)

fmt.Printf("Found %d error logs\n", total)
for _, log := range logs {
    fmt.Printf("[%s] %s: %s\n", log.Timestamp, log.Level, log.Message)
}

// 获取错误统计
stats, _ := aggregator.GetErrorStats(ctx, startTime, endTime)
fmt.Printf("Total errors: %d\n", stats.TotalErrors)
for service, count := range stats.ByService {
    fmt.Printf("  %s: %d\n", service, count)
}

// 根据请求ID获取完整链路
logs, _ = aggregator.GetLogsByRequestID(ctx, "request-123")
```

### 5. 配置日志归档

```bash
# 设置环境变量
export ELASTICSEARCH_HOST="localhost:9200"
export ARCHIVE_DAYS=30
export DELETE_DAYS=90
export S3_BUCKET="enterprise-platform-logs"
export S3_REGION="us-west-2"

# 手动执行归档
./monitoring/logging/scripts/archive-logs.sh

# 配置Cron任务
crontab -e
# 添加: 0 2 * * * /opt/.../archive-logs.sh

# 查看归档报告
cat /var/log/log-archive.log
```

---

## 🚀 下一步

**Task 4.1.8 - 实现追踪系统（Day 21-22）**:
- Jaeger/Zipkin集成
- OpenTelemetry配置
- 分布式追踪
- 性能分析
- 链路可视化

---

## 📁 文件清单

```
monitoring/logging/
├── internal/
│   ├── logger/
│   │   └── logger.go                  ✅ 结构化日志库（420行）
│   ├── middleware/
│   │   └── logging_middleware.go      ✅ 日志中间件（220行）
│   └── aggregator/
│       └── log_aggregator.go          ✅ 日志聚合器（380行）
├── elasticsearch/
│   └── elasticsearch.yml              ✅ ES配置（100行）
├── logstash/
│   └── pipeline.conf                  ✅ Logstash配置（180行）
├── kibana/
│   └── kibana.yml                     ✅ Kibana配置（160行）
├── fluentd/
│   └── fluent.conf                    ✅ Fluentd配置（180行）
├── scripts/
│   ├── archive-logs.sh                ✅ 归档脚本（180行）
│   └── crontab                        ✅ Cron配置（10行）
├── filebeat.yml                       ✅ Filebeat配置（120行）
├── docker-compose.elk.yml             ✅ Docker Compose（100行）
└── README.md                           📝 待添加
```

**总代码量**: ~2,050行

---

**版本**: v1.0.0
**状态**: ✅ Task 4.1.7 完成
**输出**: 完整日志系统、结构化日志、ELK Stack、日志聚合、归档清理

## 🎉 Task 4.1.7 日志系统实现完成！

实现了完整的企业级日志系统：
- ✅ 结构化日志库（基于zap，5个级别，上下文传播）
- ✅ 日志中间件（HTTP/gRPC，自动记录，panic恢复）
- ✅ ELK Stack（Elasticsearch + Logstash + Kibana）
- ✅ Fluentd（替代方案，更轻量）
- ✅ 日志聚合器（搜索、聚合、统计、链路追踪）
- ✅ 索引生命周期（Hot→Warm→Cold→Delete）
- ✅ 日志归档（S3快照，90天清理）
- ✅ 敏感信息脱敏（password/token/secret）
- ✅ 分布式追踪集成（trace_id/span_id）
- ✅ Docker Compose一键部署

**从日志收集到分析，全链路覆盖！**
