# Claude API 配置说明

本项目已更新为使用 Claude API，支持两种配置方式：

## 方式1：使用官方 Anthropic API

1. 在项目根目录创建 `.env` 文件：
```bash
ANTHROPIC_API_KEY=your-anthropic-api-key-here
```

2. 运行程序：
```bash
npm run start:real
```

## 方式2：使用自定义 API 端点（例如中转服务）

如果您使用的是自定义 API 端点（如 `.claude/settings.json` 中配置的），需要：

1. 在项目根目录创建 `.env` 文件：
```bash
ANTHROPIC_AUTH_TOKEN=xxxx
ANTHROPIC_BASE_URL=https://cn.aihezu.dev/api
```

2. 运行程序：
```bash
npm run start:real
```

## 代码修改说明

### 1. 支持的环境变量
- `ANTHROPIC_API_KEY` - 官方 Anthropic API 密钥
- `ANTHROPIC_AUTH_TOKEN` - 自定义端点的认证令牌
- `ANTHROPIC_BASE_URL` - 自定义 API 端点 URL

### 2. ClaudeLLMService 构造函数
```typescript
constructor(
  apiKey: string = "",
  model: string = "claude-3-5-sonnet-20241022",
  baseURL?: string,
)
```

现在支持传入 `baseURL` 参数，会自动从环境变量读取。

### 3. 可用的 Claude 模型
- `claude-3-5-sonnet-20241022` (推荐，最新的Claude 3.5 Sonnet)
- `claude-3-opus-20240229` (最强大的推理能力)
- `claude-3-sonnet-20240229` (平衡性能和速度)
- `claude-3-haiku-20240307` (最快速且经济)

## 快速开始

1. 复制配置文件：
```bash
cp .env.example .env
```

2. 编辑 `.env` 文件，根据您的情况选择方式1或方式2配置

3. 安装依赖：
```bash
npm install
```

4. 运行程序：
```bash
npm run start:real
```

## 注意事项

- `.env` 文件包含敏感信息，已添加到 `.gitignore`，不会被提交到版本控制
- `ANTHROPIC_API_KEY` 和 `ANTHROPIC_AUTH_TOKEN` 二选一即可
- 使用自定义端点时，必须同时设置 `ANTHROPIC_AUTH_TOKEN` 和 `ANTHROPIC_BASE_URL`
