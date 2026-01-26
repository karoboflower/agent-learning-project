# 自主Agent示例运行指南

## 快速开始

### 方法1：使用 ts-node（推荐）

```bash
# 安装依赖
npm install

# 运行示例
npm start
```

### 方法2：编译后运行

```bash
# 安装依赖
npm install

# 编译TypeScript
npm run build

# 运行编译后的JavaScript
npm run run
```

### 方法3：使用Node.js直接运行（需要全局安装ts-node）

```bash
# 全局安装ts-node
npm install -g ts-node

# 直接运行
ts-node typescript-autonomous-agent.ts
```

## 示例说明

这个示例展示了：

1. **自主状态管理**：Agent维护自己的内部状态
2. **自主决策**：Agent自主选择要执行的任务
3. **自主执行循环**：Agent独立运行，不需要外部持续输入
4. **错误恢复**：Agent能够自主处理错误并重试

## 运行结果

运行后会看到：
- Agent的初始状态
- 任务执行过程
- 最终的执行结果统计
- 已完成和失败的任务列表

## 自定义示例

你可以修改 `example()` 函数中的参数来测试不同的场景：

```typescript
const agent = new AutonomousAgent('你的目标', {
  maxIterations: 100,    // 最大迭代次数
  maxCost: 1000,         // 最大成本
  minPriority: 0.3       // 最小优先级
});
```


## 环境变量配置

### 1. 安装依赖

```bash
npm install
```

### 2. 配置API密钥

**方式1：使用 .env 文件（推荐）**

```bash
# 复制示例文件
cp .env.example .env

# 编辑 .env 文件，填入你的API密钥
# OPENAI_API_KEY=your-api-key-here
```

**方式2：使用环境变量**

```bash
export OPENAI_API_KEY="your-api-key-here"
```

### 3. 测试API连接

```bash
# 使用环境变量测试curl
curl -X POST "https://api.openai.com/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}],"temperature":0.7,"max_tokens":1000}'
```

## ⚠️ 安全提示

- **不要**将 `.env` 文件提交到代码仓库
- **不要**在代码中硬编码API密钥
- 如果API密钥已暴露，请立即在OpenAI控制台撤销并重新生成

## 常见问题

### SSL连接错误

如果遇到 `SSL_ERROR_SYSCALL` 错误，可能是网络问题：

```bash
# 检查网络连接
ping api.openai.com

# 如果使用代理，设置代理环境变量
export https_proxy=http://your-proxy:port
```

### API配额不足

如果遇到 `insufficient_quota` 错误：
1. 检查OpenAI账户余额
2. 在OpenAI控制台查看使用情况
3. 考虑升级账户或等待配额重置