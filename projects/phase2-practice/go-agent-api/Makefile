.PHONY: help build run test clean docker-up docker-down install

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the application
	go build -o bin/server cmd/server/main.go

run: ## Run the application
	go run cmd/server/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

docker-up: ## Start Docker services (Redis + PostgreSQL)
	docker-compose up -d

docker-down: ## Stop Docker services
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

mod-verify: ## Verify dependencies
	go mod verify

dev: docker-up run ## Start dev environment

.DEFAULT_GOAL := help
