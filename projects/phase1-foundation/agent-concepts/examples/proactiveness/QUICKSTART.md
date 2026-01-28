# 快速开始指南

## 运行Agent主动性示例的步骤

### 1. 进入项目目录

```bash
cd projects/phase1-foundation/agent-concepts/examples/proactiveness
```

### 2. 安装依赖

```bash
npm install
```

### 3. 配置API密钥

复制环境变量示例文件并配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入你的API密钥：

```
ANTHROPIC_API_KEY=your_actual_api_key_here
```

### 4. 运行示例

```bash
npm run dev
```

## 测试Agent行为

### 创建测试文件

Agent启动后，你可以在 `monitored_project` 目录下创建一些测试文件：

```bash
# 创建一个带有TODO的TypeScript文件
cat > monitored_project/example.ts << 'EOF'
function calculate(a: number, b: number) {
  // TODO: 添加输入验证
  return a + b;
}

function processData(data: any) {
  // FIXME: 需要处理空值情况
  return data.value * 2;
}
EOF
```

### 观察Agent行为

Agent将会：
1. 🔍 主动扫描发现这个文件
2. 💡 识别TODO和FIXME为改进机会
3. 🤖 使用LLM分析代码质量
4. ✨ 主动生成改进报告或TODO列表
5. 🔮 预测项目未来需求

### 查看生成的报告

```bash
# 查看改进报告
cat monitored_project/IMPROVEMENT_REPORT.md

# 查看TODO列表
cat monitored_project/TODO.md

# 查看预测报告
cat monitored_project/PREDICTIONS.md
```

## 常见问题

### Q: Agent没有反应？
A: 确保 `monitored_project` 目录下有代码文件（.ts, .js, .py等）

### Q: 如何停止Agent？
A: 按 `Ctrl+C` 或等待2分钟后自动停止

### Q: 如何调整扫描频率？
A: 编辑 `typescript-proactive-agent.ts` 中的 `config` 对象

```typescript
private config = {
  opportunityScanInterval: 30000,  // 改为更长或更短的间隔
  predictionInterval: 60000,
  opportunityThreshold: 0.3,
  maxActionsPerCycle: 3
};
```

## 下一步

- 阅读 [README.md](./README.md) 了解详细的实现原理
- 查看 [agent-proactiveness.md](../../../docs/learning-notes/agent-proactiveness.md) 学习主动性理论
- 尝试修改代码，添加新的目标和行为模式
