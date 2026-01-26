# 集成LLM的自主Agent示例

## 概述

这个示例展示了如何在自主Agent中集成LLM（大语言模型），使Agent能够：

1. **智能任务生成**：使用LLM根据目标自动生成任务列表
2. **优先级评估**：使用LLM评估任务的优先级
3. **结果分析**：使用LLM分析任务执行结果
4. **动态任务创建**：基于执行结果使用LLM生成新任务

## 快速开始

### 1. 配置API密钥（可选）

如果要使用真实的LLM API，需要设置环境变量：

```bash
# OpenAI
export OPENAI_API_KEY=your-openai-api-key

# 或者使用Anthropic Claude
export ANTHROPIC_API_KEY=your-anthropic-api-key
```

**注意**：如果没有配置API密钥，示例会使用模拟响应，仍然可以运行和演示。

### 2. 运行示例

```bash
# 运行集成LLM的版本
npm run start:llm

# 或者运行基础版本（不使用LLM）
npm start
```

## LLM集成说明

### LLM服务接口

```typescript
interface LLMService {
  generate(prompt: string): Promise<string>;
  generateTasks(goal: string, context: string): Promise<Task[]>;
  analyzeResult(task: Task, result: TaskResult): Promise<string>;
  evaluateTaskPriority(task: Task, goal: string): Promise<number>;
}
```

### 主要功能

#### 1. 任务生成

Agent使用LLM将用户目标分解为具体任务：

```typescript
const tasks = await llm.generateTasks(
  '构建一个待办事项管理Web应用',
  '已完成：需求分析'
);
```

#### 2. 优先级评估

Agent使用LLM评估任务的优先级：

```typescript
const priority = await llm.evaluateTaskPriority(
  task,
  '构建一个待办事项管理Web应用'
);
```

#### 3. 结果分析

Agent使用LLM分析任务执行结果：

```typescript
const analysis = await llm.analyzeResult(task, result);
```

## 代码结构

```
typescript-autonomous-agent-with-llm.ts
├── LLMService接口          # LLM服务抽象
├── OpenAILLMService        # OpenAI实现
├── AutonomousAgentWithLLM  # 集成LLM的Agent
└── example()               # 使用示例
```

## 自定义LLM服务

你可以实现自己的LLM服务：

```typescript
class CustomLLMService implements LLMService {
  async generate(prompt: string): Promise<string> {
    // 实现你的LLM调用逻辑
  }
  
  async generateTasks(goal: string, context: string): Promise<Task[]> {
    // 实现任务生成逻辑
  }
  
  // ... 其他方法
}

// 使用自定义LLM服务
const agent = new AutonomousAgentWithLLM(
  '你的目标',
  new CustomLLMService()
);
```

## 成本控制

示例中包含了成本跟踪机制：

```typescript
class CostTracker {
  track(tokens: number): void {
    // 跟踪Token使用和成本
  }
}
```

## 错误处理

示例实现了LLM调用的错误处理和降级策略：

- API调用失败时，自动降级到模拟响应
- 重试机制
- 错误日志记录

## 最佳实践

1. **Prompt设计**：设计清晰、具体的Prompt
2. **温度设置**：根据任务类型调整temperature
3. **Token管理**：监控和限制Token使用
4. **错误处理**：实现降级和重试机制
5. **成本控制**：跟踪和限制API调用成本

## 下一步

- 阅读 [LLM详解文档](../../../../docs/learning-notes/what-is-llm.md)
- 学习Prompt Engineering技巧
- 探索其他LLM提供商（Anthropic、Google等）
