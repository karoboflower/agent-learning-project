# Enterprise Agent Platform - Makefile

.PHONY: all proto-gen build test docker-build docker-push k8s-deploy dev-up dev-down clean help

# 变量定义
PROJECT_NAME := enterprise-platform
VERSION := v1.0.0
REGISTRY := your-registry.com
ENV ?= dev

# Proto文件路径
PROTO_DIR := architecture/proto
PROTO_FILES := $(wildcard $(PROTO_DIR)/*.proto)

# 服务列表
SERVICES := agent task tool user tenant cost monitoring

# 默认目标
all: proto-gen build test

# 帮助信息
help:
	@echo "Enterprise Agent Platform - Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make proto-gen        - 生成gRPC代码"
	@echo "  make build            - 编译所有服务"
	@echo "  make test             - 运行测试"
	@echo "  make docker-build     - 构建Docker镜像"
	@echo "  make docker-push      - 推送Docker镜像"
	@echo "  make k8s-deploy       - 部署到Kubernetes"
	@echo "  make dev-up           - 启动开发环境"
	@echo "  make dev-down         - 停止开发环境"
	@echo "  make clean            - 清理构建文件"
	@echo ""
	@echo "Example:"
	@echo "  make proto-gen"
	@echo "  make build"
	@echo "  make k8s-deploy ENV=prod"

# 生成gRPC代码
proto-gen:
	@echo "Generating gRPC code..."
	@for service in $(SERVICES); do \
		echo "Generating $$service..."; \
		protoc --go_out=services/$$service/proto \
		       --go_opt=paths=source_relative \
		       --go-grpc_out=services/$$service/proto \
		       --go-grpc_opt=paths=source_relative \
		       $(PROTO_DIR)/$$service.proto; \
	done
	@echo "gRPC code generation completed!"

# 编译所有服务
build:
	@echo "Building all services..."
	@for service in $(SERVICES); do \
		echo "Building $$service-service..."; \
		cd services/$$service && go build -o bin/$$service-service ./cmd; \
		cd ../..; \
	done
	@echo "Build completed!"

# 运行测试
test:
	@echo "Running tests..."
	@for service in $(SERVICES); do \
		echo "Testing $$service-service..."; \
		cd services/$$service && go test -v ./...; \
		cd ../..; \
	done
	@echo "Tests completed!"

# 运行集成测试
test-integration:
	@echo "Running integration tests..."
	@cd tests/integration && go test -v ./...
	@echo "Integration tests completed!"

# 运行性能测试
test-performance:
	@echo "Running performance tests..."
	@cd tests/performance && go test -v -bench=. ./...
	@echo "Performance tests completed!"

# 构建Docker镜像
docker-build:
	@echo "Building Docker images..."
	@for service in $(SERVICES); do \
		echo "Building $$service-service:$(VERSION)..."; \
		docker build -t $(REGISTRY)/$$service-service:$(VERSION) \
		             -f services/$$service/Dockerfile \
		             services/$$service; \
	done
	@echo "Docker images build completed!"

# 推送Docker镜像
docker-push:
	@echo "Pushing Docker images..."
	@for service in $(SERVICES); do \
		echo "Pushing $$service-service:$(VERSION)..."; \
		docker push $(REGISTRY)/$$service-service:$(VERSION); \
	done
	@echo "Docker images push completed!"

# 部署到Kubernetes
k8s-deploy:
	@echo "Deploying to Kubernetes ($(ENV))..."
	@kubectl apply -k deploy/kubernetes/overlays/$(ENV)
	@echo "Deployment completed!"

# 启动开发环境
dev-up:
	@echo "Starting development environment..."
	@docker-compose -f deploy/docker-compose/docker-compose.dev.yml up -d
	@echo "Development environment started!"
	@echo "Services:"
	@echo "  - Agent Service:      http://localhost:8081"
	@echo "  - Task Service:       http://localhost:8082"
	@echo "  - Tool Service:       http://localhost:8083"
	@echo "  - User Service:       http://localhost:8084"
	@echo "  - Tenant Service:     http://localhost:8085"
	@echo "  - Cost Service:       http://localhost:8086"
	@echo "  - Monitoring Service: http://localhost:8087"
	@echo "  - Prometheus:         http://localhost:9090"
	@echo "  - Grafana:            http://localhost:3000"
	@echo "  - Jaeger:             http://localhost:16686"

# 停止开发环境
dev-down:
	@echo "Stopping development environment..."
	@docker-compose -f deploy/docker-compose/docker-compose.dev.yml down
	@echo "Development environment stopped!"

# 查看日志
logs:
	@docker-compose -f deploy/docker-compose/docker-compose.dev.yml logs -f

# 数据库迁移
migrate-up:
	@echo "Running database migrations..."
	@for service in $(SERVICES); do \
		if [ -d "services/$$service/migrations" ]; then \
			echo "Migrating $$service database..."; \
			migrate -path services/$$service/migrations \
			        -database "postgresql://localhost:5432/$$service?sslmode=disable" \
			        up; \
		fi; \
	done
	@echo "Migrations completed!"

# 回滚数据库迁移
migrate-down:
	@echo "Rolling back database migrations..."
	@for service in $(SERVICES); do \
		if [ -d "services/$$service/migrations" ]; then \
			echo "Rolling back $$service database..."; \
			migrate -path services/$$service/migrations \
			        -database "postgresql://localhost:5432/$$service?sslmode=disable" \
			        down; \
		fi; \
	done
	@echo "Rollback completed!"

# 代码格式化
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatting completed!"

# 代码检查
lint:
	@echo "Running linters..."
	@golangci-lint run ./...
	@echo "Linting completed!"

# 依赖更新
deps-update:
	@echo "Updating dependencies..."
	@for service in $(SERVICES); do \
		echo "Updating $$service dependencies..."; \
		cd services/$$service && go get -u ./... && go mod tidy; \
		cd ../..; \
	done
	@echo "Dependencies updated!"

# 生成API文档
docs-gen:
	@echo "Generating API documentation..."
	@protoc --doc_out=docs/api --doc_opt=html,index.html $(PROTO_FILES)
	@echo "API documentation generated!"

# 清理构建文件
clean:
	@echo "Cleaning build artifacts..."
	@for service in $(SERVICES); do \
		rm -rf services/$$service/bin; \
		rm -rf services/$$service/proto/*.pb.go; \
	done
	@docker-compose -f deploy/docker-compose/docker-compose.dev.yml down -v
	@echo "Clean completed!"

# 健康检查
health-check:
	@echo "Running health checks..."
	@./scripts/health-check.sh

# 监控服务状态
status:
	@echo "Service Status:"
	@kubectl get pods -n agent-platform-$(ENV)

# 查看服务日志
k8s-logs:
	@kubectl logs -f -l app=$(SERVICE) -n agent-platform-$(ENV)

# 进入容器
k8s-exec:
	@kubectl exec -it $$(kubectl get pods -l app=$(SERVICE) -n agent-platform-$(ENV) -o jsonpath='{.items[0].metadata.name}') -n agent-platform-$(ENV) -- /bin/sh

# 性能分析
profile:
	@echo "Starting profiling..."
	@go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 基准测试
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...
