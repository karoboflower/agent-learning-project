# Go依赖问题修复指南

## 问题说明

由于Go项目依赖需要从网络下载，可能会遇到以下问题：
1. 网络连接问题
2. Go代理配置问题
3. 模块缓存问题

## 解决方案

### 方案1: 使用国内镜像 (推荐)

```bash
# 临时设置（仅当前终端）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=sum.golang.google.cn

# 或使用阿里云镜像
export GOPROXY=https://mirrors.aliyun.com/goproxy/,direct

# 下载依赖
go mod tidy
go mod download
```

### 方案2: 永久配置Go代理

```bash
# Linux/macOS
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.bashrc
echo 'export GOSUMDB=sum.golang.google.cn' >> ~/.bashrc
source ~/.bashrc

# 或写入 ~/.bash_profile 或 ~/.zshrc
```

### 方案3: 使用go env配置

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn
```

查看当前配置:
```bash
go env | grep GOPROXY
go env | grep GOSUMDB
```

### ��案4: 清理缓存后重试

```bash
# 清理模块缓存
go clean -modcache

# 重新下载
go mod tidy
go mod download
```

## 验证依赖

下载完成后，验证依赖:

```bash
# 查看依赖列表
go list -m all

# 验证依赖完整性
go mod verify

# 查看依赖图
go mod graph
```

## 项目依赖列表

本项目主要依赖:

```
github.com/gin-gonic/gin v1.9.1           # Web框架
github.com/go-redis/redis/v8 v8.11.5      # Redis客户端
github.com/lib/pq v1.10.9                 # PostgreSQL驱动
github.com/sashabaranov/go-openai v1.17.9 # OpenAI SDK
github.com/google/uuid v1.5.0             # UUID生成
github.com/joho/godotenv v1.5.1           # 环境变量加载
```

## 离线安装 (可选)

如果网络完全无法访问，可以使用vendor模式:

```bash
# 下载所有依赖到vendor目录
go mod vendor

# 使用vendor模式运行
go run -mod=vendor cmd/server/main.go

# 使用vendor模式构建
go build -mod=vendor -o bin/server cmd/server/main.go
```

## 常见错误处理

### 错误1: "could not import"

```bash
# 解决方法
go mod tidy
```

### 错误2: "checksum mismatch"

```bash
# 清理并重新下载
go clean -modcache
go mod download
```

### 错误3: "dial tcp: i/o timeout"

```bash
# 切换代理
go env -w GOPROXY=https://goproxy.cn,direct
go mod download
```

## 最小化依赖版本

如果某些依赖版本有问题，可以降级:

```bash
# 降级到特定版本
go get github.com/gin-gonic/gin@v1.9.0

# 查看可用版本
go list -m -versions github.com/gin-gonic/gin
```

## 简化版本 (无外部依赖)

如果依赖问题无法解决，项目已经设计为可以在没有Redis和PostgreSQL的情况下运行:

1. **Redis不可用** - 自动降级到内存缓存
2. **PostgreSQL不可用** - 跳过数据持久化
3. **只需要OpenAI** - 最小配置只需要OPENAI_API_KEY

最小启动配置:
```env
OPENAI_API_KEY=your_key_here
```

## 测试依赖安装

运行简单测试确认依赖正常:

```bash
# 编译测试（不运行）
go test -c ./internal/agent

# 如果编译成功，说明依赖安装正确
```

## 获取帮助

如果以上方法都无法解决问题:

1. 检查Go版本: `go version` (需要1.21+)
2. 检查网络连接: `ping goproxy.cn`
3. 查看详细错误: `go get -v github.com/gin-gonic/gin`
4. 查看Go环境: `go env`

## 推荐的完整安装流程

```bash
# 1. 配置代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=sum.golang.google.cn

# 2. 清理缓存
go clean -modcache

# 3. 下载依赖
cd go-agent-api
go mod tidy
go mod download

# 4. 验证
go mod verify

# 5. 尝试编译
go build cmd/server/main.go

# 6. 如果成功，运行测试
go test ./internal/agent

# 7. 运行服务
go run cmd/server/main.go
```

## 备注

- 项目代码本身不需要修改
- 所有依赖问题都是下载和配置问题
- 建议优先使用方案1或方案3
- 如果完全无法下载，可以使用vendor模式

---

更新日期: 2026-01-28
