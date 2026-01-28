# Agent社会性（Social Ability）示例

这个项目展示了如何实现一个具有社会性的多Agent系统，通过Agent之间的协作完成代码审查任务。

## 🎯 核心特性

### 1. 社会性（Social Ability）
- ✅ **通信机制**：Agent通过消息总线进行通信
- ✅ **协作能力**：多个Agent协作完成复杂任务
- ✅ **角色分工**：每个Agent有明确的专业角色
- ✅ **协调机制**：协调者统筹整个流程

### 2. 多Agent系统架构
- **分析Agent（AnalyzerAgent）**：分析代码结构和复杂度
- **审查Agent（ReviewerAgent）**：基于分析提出改进建议
- **协调Agent（CoordinatorAgent）**：协调任务流程和整合结果

### 3. 智能协作
- **消息传递**：异步消息通信
- **任务协调**：协调者分配任务并收集结果
- **角色配合**：不同专长的Agent协同工作
- **LLM驱动**：每个Agent使用Claude AI进行智能分析

## 📋 项目结构

```
social-ability/
├── typescript-social-agents.ts    # 多Agent系统实现
├── package.json                   # 项目配置
├── tsconfig.json                  # TypeScript配置
├── .gitignore                     # Git忽略文件
├── .env.example                   # 环境变量示例
└── README.md                      # 本文件
```

## 🚀 快速开始

### 前置条件

1. 安装Node.js (v16+)
2. 获取Anthropic API密钥

### 安装依赖

```bash
cd social-ability
npm install
```

### 配置环境变量

创建 `.env` 文件：

```bash
cp .env.example .env
# 编辑.env填入ANTHROPIC_API_KEY
```

### 运行示例

```bash
npm run dev
```

### 构建项目

```bash
npm run build
npm start
```

## 💡 工作原理

### 1. 多Agent协作流程

```
用户请求
   ↓
协调Agent (发起审查)
   ↓
   ├─→ 分析Agent (分析代码)
   │      ↓
   │   返回分析结果
   │      ↓
   └─→ 审查Agent (提供建议)
          ↓
       返回审查建议
          ↓
协调Agent (整合报告)
   ↓
生成最终报告
```

### 2. 消息通信机制

```typescript
// 发送消息
coordinator.sendMessage('analyzer_01', 'analyze_request', {
  filePath: 'example.ts',
  content: '...'
});

// 接收消息
analyzer.on('message', (message) => {
  if (message.type === 'analyze_request') {
    // 处理分析请求
  }
});
```

### 3. Agent角色定义

```typescript
// 分析Agent
class AnalyzerAgent {
  role = 'Code Analyzer';
  capabilities = ['code_analysis', 'complexity_analysis'];

  async analyzeCode(code) {
    // 使用LLM分析代码
  }
}

// 审查Agent
class ReviewerAgent {
  role = 'Code Reviewer';
  capabilities = ['code_review', 'best_practices'];

  async generateSuggestions(analysis) {
    // 基于分析生成建议
  }
}

// 协调Agent
class CoordinatorAgent {
  role = 'Coordinator';
  capabilities = ['task_coordination', 'result_integration'];

  async initiateCodeReview(filePath) {
    // 协调整个审查流程
  }
}
```

## 📊 示例输出

```
🚀 多Agent代码审查系统启动
========================================
👥 系统包含 3 个Agent:
✅ [analyzer_01] Code Analyzer 已启动
✅ [reviewer_01] Code Reviewer 已启动
✅ [coordinator_01] Coordinator 已启动
========================================

🎯 [协调Agent] 发起代码审查流程: example.ts

📤 [协调Agent] 请求分析Agent分析代码...

📨 [消息总线] coordinator_01 -> analyzer_01: analyze_request
  📬 [analyzer_01] 收到消息: analyze_request

🔍 [分析Agent] 开始分析代码: example.ts
✅ [分析Agent] 分析完成

📨 [消息总线] analyzer_01 -> coordinator_01: analyze_result
  📬 [coordinator_01] 收到消息: analyze_result

📤 [协调Agent] 请求审查Agent提供建议...

📨 [消息总线] coordinator_01 -> reviewer_01: review_request
  📬 [reviewer_01] 收到消息: review_request

📝 [审查Agent] 开始审查代码，基于分析结果
✅ [审查Agent] 审查完成，提出 4 条建议

📨 [消息总线] reviewer_01 -> coordinator_01: review_result
  📬 [coordinator_01] 收到消息: review_result

📊 [协调Agent] 整合结果生成报告...

✅ [协调Agent] 代码审查流程完成

============================================================
📋 代码审查报告
============================================================
📁 文件: example.ts
📊 评分: 68/100
⏰ 时间: 2026-01-26 23:15:30

--- 代码分析 ---
📏 代码行数: 24
🔢 复杂度: 中等
❌ 问题 (3):
   1. 缺少类型注解
   2. 使用 == 而非 ===
   3. 未处理边界情况

✅ 优点 (2):
   1. 代码结构清晰
   2. 命名规范

--- 改进建议 ---

1. [类型安全] 🔴 HIGH
   问题: addUser和getUser方法缺少类型注解
   建议: 为所有方法参数和返回值添加TypeScript类型

2. [代码质量] 🟡 MEDIUM
   问题: 使用 == 进行比较
   建议: 使用 === 进行严格相等比较

3. [性能优化] 🟡 MEDIUM
   问题: 使用传统for循环
   建议: 使用数组方法如reduce、find等提升可读性

4. [错误处理] 🟢 LOW
   问题: 未处理空数组情况
   建议: 添加边界条件检查

--- 总结 ---
代码基本功能完整，但需要加强类型安全和代码质量。
建议优先处理类型注解和比较运算符问题。
============================================================

💾 报告已保存到: test_project/review_report.json
```

## 🎓 学习要点

### 1. 社会性核心概念

- **通信**：Agent通过消息传递进行通信
- **协作**：多个Agent分工合作完成任务
- **协调**：协调者统筹任务分配和结果整合
- **角色**：每个Agent有明确的专业角色和能力

### 2. 社会性 vs 其他特征

| 特征 | 驱动方式 | 关键特点 | 本示例 |
|------|---------|---------|--------|
| 自主性 | 内部目标 | 独立决策 | 每个Agent独立分析 |
| 反应性 | 外部事件 | 即时响应 | 接收消息立即处理 |
| 主动性 | 目标+机会 | 主动发起 | 协调者主动发起审查 |
| 社会性 | 多方交互 | 协作通信 | Agent之间消息协作 |

### 3. 实现关键点

```typescript
// 1. 消息总线：Agent通信的基础设施
class MessageBus extends EventEmitter {
  send(message: Message) {
    this.emit(`message:${message.to}`, message);
  }
}

// 2. 异步消息处理
protected async handleMessage(message: Message) {
  switch (message.type) {
    case 'request':
      await this.handleRequest(message);
      break;
  }
}

// 3. 等待响应模式
const response = await this.waitForResponse(conversationId);

// 4. 角色专业化
class SpecializedAgent extends SocialAgent {
  capabilities = ['specific_skill'];

  async performSpecializedTask() {
    // 专业任务处理
  }
}
```

## 🔧 配置参数

### Agent配置

```typescript
// 分析Agent
new AnalyzerAgent(messageBus, llm);

// 审查Agent
new ReviewerAgent(messageBus, llm);

// 协调Agent
new CoordinatorAgent(messageBus, llm);
```

### 消息超时

```typescript
const response = await this.waitForResponse(
  conversationId,
  30000  // 30秒超时
);
```

## 🔍 扩展建议

### 1. 添加更多专业Agent

```typescript
class SecurityAgent extends SocialAgent {
  // 专门检查安全问题
  async checkSecurity(code: string) {
    // 使用LLM检查安全漏洞
  }
}

class PerformanceAgent extends SocialAgent {
  // 专门分析性能
  async analyzePerformance(code: string) {
    // 使用LLM分析性能瓶颈
  }
}
```

### 2. 实现协商机制

```typescript
class NegotiatingAgent extends SocialAgent {
  async negotiate(proposal: Proposal) {
    // 评估提案
    const evaluation = await this.evaluateProposal(proposal);

    if (evaluation.acceptable) {
      return this.accept();
    } else {
      return this.counterPropose();
    }
  }
}
```

### 3. 添加团队学习

```typescript
class LearningTeam extends MultiAgentSystem {
  async learnFromReviews() {
    // 从历史审查中学习
    const history = this.getReviewHistory();

    // 更新Agent的知识库
    for (const agent of this.agents.values()) {
      await agent.updateKnowledge(history);
    }
  }
}
```

### 4. 实现动态任务分配

```typescript
class DynamicCoordinator extends CoordinatorAgent {
  async assignTaskDynamically(task: Task) {
    // 查询所有Agent的当前负载
    const loads = await this.queryAgentLoads();

    // 选择负载最低且能力匹配的Agent
    const bestAgent = this.selectBestAgent(task, loads);

    await this.assignTask(bestAgent, task);
  }
}
```

## 📚 参考资料

- [Multi-Agent Systems](https://en.wikipedia.org/wiki/Multi-agent_system)
- [Agent Communication Languages](http://www.fipa.org/repository/aclspecs.html)
- [Cooperative Problem Solving](https://en.wikipedia.org/wiki/Cooperative_problem_solving)
- [Message Passing](https://en.wikipedia.org/wiki/Message_passing)

## 🤝 相关示例

- [自主性示例](../autonomy) - 学习Agent的自主性
- [反应性示例](../reactivity) - 学习Agent的反应性
- [主动性示例](../proactiveness) - 学习Agent的主动性

## 📝 总结

这个示例展示了如何实现一个具有社会性的多Agent系统：

1. ✅ **消息通信**：通过消息总线实现Agent间通信
2. ✅ **角色分工**：三个专业Agent各司其职
3. ✅ **协作流程**：协调者统筹整个审查流程
4. ✅ **智能决策**：每个Agent使用LLM进行分析和决策
5. ✅ **结果整合**：协调者整合多个Agent的结果

社会性使Agent能够通过协作完成单个Agent难以完成的复杂任务，是构建强大AI系统的关键特征。

---

**下一步学习**：
- [ ] 探索更复杂的协商机制
- [ ] 实现Agent之间的知识共享
- [ ] 构建大规模多Agent系统
- [ ] 研究Agent组织结构和层级
