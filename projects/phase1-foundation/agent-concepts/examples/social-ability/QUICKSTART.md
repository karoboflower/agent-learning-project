# 快速开始指南

## 运行Agent社会性示例的步骤

### 1. 进入项目目录

```bash
cd projects/phase1-foundation/agent-concepts/examples/social-ability
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

## 系统运行流程

### 观察多Agent协作

系统运行时，你将看到三个Agent协作完成代码审查：

```
1️⃣ 系统启动
   └─ 3个Agent同时启动

2️⃣ 协调Agent发起审查
   └─ 创建测试代码文件

3️⃣ 分析Agent分析代码
   ├─ 接收协调Agent的请求
   ├─ 使用LLM分析代码
   └─ 返回分析结果

4️⃣ 审查Agent提供建议
   ├─ 接收分析结果
   ├─ 使用LLM生成建议
   └─ 返回审查建议

5️⃣ 协调Agent整合报告
   ├─ 收集所有结果
   ├─ 使用LLM生成总结
   └─ 输出最终报告
```

### 查看生成的报告

```bash
# 查看JSON格式的完整报告
cat test_project/review_report.json

# 查看生成的测试代码
cat test_project/example.ts
```

## 测试自定义代码

### 修改测试代码

编辑 `typescript-social-agents.ts` 中的 `testCode` 变量：

```typescript
const testCode = `
// 在这里写你想审查的代码
function yourFunction() {
  // ...
}
`;
```

或者直接在 `test_project/` 目录下创建文件，然后修改审查路径。

### 运行自定义审查

```typescript
// 在main函数中修改要审查的文件
const report = await system.reviewCode("path/to/your/file.ts");
```

## 理解输出信息

### 消息通信日志

```
📨 [消息总线] sender -> receiver: message_type
  📬 [receiver] 收到消息: message_type
```

- `📨`: 消息被发送
- `📬`: 消息被接收
- `sender -> receiver`: 发送者到接收者
- `message_type`: 消息类型（如 analyze_request）

### Agent状态日志

```
🔍 [分析Agent] 开始分析代码
✅ [分析Agent] 分析完成

📝 [审查Agent] 开始审查代码
✅ [审查Agent] 审查完成

🎯 [协调Agent] 发起代码审查流程
✅ [协调Agent] 代码审查流程完成
```

### 审查报告

报告包含：
- **代码分析**：行数、复杂度、问题、优点
- **改进建议**：按优先级分类的具体建议
- **总体评分**：0-100分
- **总结**：简洁的审查总结

## 常见问题

### Q: 如何添加新的Agent？

A: 继承 `SocialAgent` 类并实现 `handleMessage` 方法：

```typescript
class MyCustomAgent extends SocialAgent {
  constructor(messageBus: MessageBus, llm: LLMService) {
    super('my_agent_01', 'My Role', ['my_capability'], messageBus, llm);
  }

  protected async handleMessage(message: Message): Promise<void> {
    // 处理特定类型的消息
  }
}
```

### Q: 如何修改Agent之间的协作流程？

A: 修改 `CoordinatorAgent` 的 `initiateCodeReview` 方法：

```typescript
async initiateCodeReview(filePath: string): Promise<ReviewReport> {
  // 1. 添加新的步骤
  // 2. 修改消息发送顺序
  // 3. 增加或删除Agent参与
}
```

### Q: 如何查看所有消息？

A: 在系统停止前添加：

```typescript
const messageLog = system.messageBus.getMessageLog();
console.log('所有消息:', messageLog);
```

### Q: Agent之间通信失败怎么办？

A: 检查几点：
1. Agent ID是否正确（接收者必须存在）
2. 消息类型是否被接收者处理
3. conversationId是否正确传递
4. 是否设置了合理的超时时间

### Q: 如何调整超时时间？

A: 修改 `waitForResponse` 调用：

```typescript
const response = await this.waitForResponse(
  conversationId,
  60000  // 改为60秒
);
```

## 扩展实验

### 实验1：添加安全检查Agent

```typescript
class SecurityAgent extends SocialAgent {
  constructor(messageBus: MessageBus, llm: LLMService) {
    super('security_01', 'Security Checker', ['security_analysis'], messageBus, llm);
  }

  protected async handleMessage(message: Message): Promise<void> {
    if (message.type === 'security_check') {
      const result = await this.checkSecurity(message.content.code);
      await this.sendMessage(message.from, 'security_result', result);
    }
  }

  private async checkSecurity(code: string) {
    // 使用LLM检查安全问题
    const prompt = `检查以下代码的安全问题：\n${code}`;
    return await this.llm.analyze(prompt);
  }
}
```

### 实验2：实现Agent投票机制

```typescript
class VotingCoordinator extends CoordinatorAgent {
  async collectVotes(proposal: any): Promise<any> {
    // 向所有Agent发送提案
    const votes = [];
    for (const agent of this.agents) {
      const vote = await this.requestVote(agent, proposal);
      votes.push(vote);
    }

    // 统计投票结果
    return this.tallyVotes(votes);
  }
}
```

### 实验3：添加知识共享

```typescript
class KnowledgeSharingAgent extends SocialAgent {
  private knowledgeBase: Map<string, any> = new Map();

  async shareKnowledge(key: string, value: any) {
    this.knowledgeBase.set(key, value);

    // 广播给其他Agent
    await this.broadcast('knowledge_update', { key, value });
  }

  async queryKnowledge(key: string): Promise<any> {
    return this.knowledgeBase.get(key);
  }
}
```

## 性能优化建议

### 1. 并行处理

如果有多个独立的Agent可以并行工作：

```typescript
const [analysis, review] = await Promise.all([
  this.requestAnalysis(code),
  this.requestReview(code)
]);
```

### 2. 消息批处理

对于大量消息：

```typescript
const messages = [...]; // 多条消息
await this.messageBus.sendBatch(messages);
```

### 3. 缓存LLM结果

对于相似的请求：

```typescript
private cache: Map<string, any> = new Map();

async analyzeWithCache(code: string) {
  const hash = this.hashCode(code);
  if (this.cache.has(hash)) {
    return this.cache.get(hash);
  }

  const result = await this.llm.analyze(code);
  this.cache.set(hash, result);
  return result;
}
```

## 下一步

- 阅读 [README.md](./README.md) 了解详细实现
- 查看 [agent-social-ability.md](../../../docs/learning-notes/agent-social-ability.md) 学习社会性理论
- 尝试实现多Agent协商系统
- 探索更复杂的组织结构
