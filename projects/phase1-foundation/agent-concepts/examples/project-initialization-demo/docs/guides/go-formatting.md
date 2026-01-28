# Go项目配置

## gofmt配置

Go语言自带 `gofmt` 工具，无需额外配置。

### 使用方式

```bash
# 格式化单个文件
gofmt -w main.go

# 格式化整个目录
gofmt -w .

# 查看格式化diff（不修改文件）
gofmt -d main.go
```

## golangci-lint配置

推荐使用 `golangci-lint` 进行更全面的代码检查。

配置文件：`.golangci.yml`

```yaml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
```

## 编辑器集成

### VS Code
安装 Go 扩展后，会自动使用 gofmt。

### GoLand
默认集成 gofmt。
