# LLM（大语言模型）详解

## 📚 目录

1. [什么是LLM](#什么是llm)
2. [LLM的工作原理](#llm的工作原理)
3. [主流LLM模型](#主流llm模型)
4. [LLM在Agent中的应用](#llm在agent中的应用)
5. [如何集成LLM](#如何集成llm)
6. [最佳实践](#最佳实践)
7. [常见问题](#常见问题)

---

## 什么是LLM

### 定义

**LLM（Large Language Model，大语言模型）**是一种基于深度学习的AI模型，能够理解和生成人类语言。它们通过在海量文本数据上训练，学会了语言的模式、语法、语义和知识。

### 核心特点

1. **大规模**：参数量通常达到数十亿甚至数千亿
2. **通用性**：能够处理各种语言任务
3. **上下文理解**：能够理解上下文和语境
4. **生成能力**：能够生成连贯、有意义的文本

### LLM vs 传统NLP模型

| 特性 | 传统NLP模型 | LLM |
|------|------------|-----|
| **训练数据** | 特定领域数据 | 大规模通用文本 |
| **参数量** | 百万到千万级 | 十亿到千亿级 |
| **能力** | 单一任务 | 多任务通用 |
| **微调** | 需要大量标注数据 | 少量样本即可（Few-shot） |
| **泛化能力** | 较弱 | 很强 |

---

## LLM的工作原理

### 1. 预训练（Pre-training）

LLM首先在大规模文本数据上进行无监督预训练：

```
训练数据（互联网文本、书籍、文章等）
    ↓
Token化（将文本转换为数字）
    ↓
Transformer架构处理
    ↓
学习语言模式和知识
    ↓
预训练模型
```

### 2. Transformer架构

LLM基于Transformer架构，核心组件包括：

- **Self-Attention机制**：理解词与词之间的关系
- **Feed-Forward Networks**：处理信息
- **Layer Normalization**：稳定训练
- **Positional Encoding**：理解词序

### 3. 生成过程

LLM生成文本的过程：

```
输入提示（Prompt）
    ↓
Token化
    ↓
模型处理（多层Transformer）
    ↓
输出概率分布
    ↓
采样下一个Token
    ↓
重复直到生成完整文本
```

### 4. 上下文窗口

LLM能够处理的上下文长度有限：

- **GPT-3.5**: 4K tokens
- **GPT-4**: 8K/32K tokens
- **Claude 3**: 200K tokens
- **GPT-4 Turbo**: 128K tokens

---

## 主流LLM模型

### OpenAI系列

#### GPT-3.5
- **特点**：性价比高，速度快
- **适用场景**：日常对话、简单任务
- **API**: `gpt-3.5-turbo`

#### GPT-4
- **特点**：能力强，推理能力好
- **适用场景**：复杂任务、代码生成、分析
- **API**: `gpt-4`, `gpt-4-turbo-preview`

### Anthropic系列

#### Claude 3
- **特点**：安全性好，上下文长
- **适用场景**：长文档处理、安全敏感应用
- **API**: `claude-3-opus`, `claude-3-sonnet`, `claude-3-haiku`

### 开源模型

#### Llama 2/3 (Meta)
- **特点**：开源，可本地部署
- **适用场景**：私有部署、成本控制

#### Mistral
- **特点**：性能好，开源
- **适用场景**：商业应用

---

## LLM在Agent中的应用

### 1. 任务理解

Agent使用LLM理解用户的目标：

```typescript
const prompt = `
用户目标：${goal}
请将这个目标分解为具体的任务列表。
`;

const tasks = await llm.generate(prompt);
```

### 2. 任务规划

Agent使用LLM规划执行步骤：

```typescript
const prompt = `
目标：${goal}
已完成任务：${completedTasks}
当前任务：${currentTask}

基于以上信息，规划下一步行动。
`;

const plan = await llm.generate(prompt);
```

### 3. 工具选择

Agent使用LLM选择合适的工具：

```typescript
const prompt = `
任务：${task}
可用工具：${availableTools}

选择最适合的工具来完成任务。
`;

const selectedTool = await llm.selectTool(prompt, availableTools);
```

### 4. 结果分析

Agent使用LLM分析执行结果：

```typescript
const prompt = `
任务：${task}
执行结果：${result}

分析这个结果，判断任务是否成功完成。
`;

const analysis = await llm.analyze(prompt);
```

### 5. 新任务生成

Agent使用LLM生成新任务：

```typescript
const prompt = `
目标：${goal}
已完成：${completedTasks}
最后结果：${lastResult}

生成下一步需要执行的任务。
`;

const newTasks = await llm.generateTasks(prompt);
```

---

## 如何集成LLM

### 方法1：使用OpenAI API

#### 安装依赖

```bash
npm install openai
```

#### 基本使用

```typescript
import OpenAI from 'openai';

const openai = new OpenAI({
  apiKey: process.env.OPENAI_API_KEY
});

async function callLLM(prompt: string): Promise<string> {
  const response = await openai.chat.completions.create({
    model: 'gpt-3.5-turbo',
    messages: [
      { role: 'user', content: prompt }
    ],
    temperature: 0.7,
    max_tokens: 1000
  });
  
  return response.choices[0].message.content || '';
}
```

### 方法2：使用Anthropic API

#### 安装依赖

```bash
npm install @anthropic-ai/sdk
```

#### 基本使用

```typescript
import Anthropic from '@anthropic-ai/sdk';

const anthropic = new Anthropic({
  apiKey: process.env.ANTHROPIC_API_KEY
});

async function callLLM(prompt: string): Promise<string> {
  const message = await anthropic.messages.create({
    model: 'claude-3-sonnet-20240229',
    max_tokens: 1000,
    messages: [
      { role: 'user', content: prompt }
    ]
  });
  
  return message.content[0].type === 'text' 
    ? message.content[0].text 
    : '';
}
```

### 方法3：使用LangChain

#### 安装依赖

```bash
npm install langchain @langchain/openai
```

#### 基本使用

```typescript
import { ChatOpenAI } from '@langchain/openai';
import { HumanMessage } from '@langchain/core/messages';

const llm = new ChatOpenAI({
  modelName: 'gpt-3.5-turbo',
  temperature: 0.7
});

async function callLLM(prompt: string): Promise<string> {
  const response = await llm.invoke([
    new HumanMessage(prompt)
  ]);
  
  return response.content as string;
}
```

---

## 在自主Agent中集成LLM

### 1. 创建LLM服务类

```typescript
interface LLMService {
  generate(prompt: string): Promise<string>;
  generateTasks(goal: string, context: string): Promise<string[]>;
  analyze(result: string): Promise<string>;
}

class OpenAILLMService implements LLMService {
  private client: OpenAI;
  
  constructor(apiKey: string) {
    this.client = new OpenAI({ apiKey });
  }
  
  async generate(prompt: string): Promise<string> {
    const response = await this.client.chat.completions.create({
      model: 'gpt-3.5-turbo',
      messages: [{ role: 'user', content: prompt }],
      temperature: 0.7
    });
    
    return response.choices[0].message.content || '';
  }
  
  async generateTasks(goal: string, context: string): Promise<string[]> {
    const prompt = `
目标：${goal}
上下文：${context}

请生成3-5个具体的任务来完成这个目标。每个任务一行。
    `;
    
    const response = await this.generate(prompt);
    return response.split('\n').filter(line => line.trim().length > 0);
  }
  
  async analyze(result: string): Promise<string> {
    const prompt = `
分析以下执行结果，判断任务是否成功完成：
${result}
    `;
    
    return await this.generate(prompt);
  }
}
```

### 2. 在Agent中使用LLM

```typescript
class AutonomousAgent {
  private llm: LLMService;
  
  constructor(goal: string, llm: LLMService) {
    this.goal = goal;
    this.llm = llm;
  }
  
  async createInitialTasks(): Promise<void> {
    const prompt = `
目标：${this.state.goal}

请将这个目标分解为3-5个具体的任务。每个任务一行，格式：任务描述|优先级(0-1)
    `;
    
    const response = await this.llm.generate(prompt);
    const taskLines = response.split('\n').filter(line => line.trim());
    
    const tasks = taskLines.map((line, index) => {
      const [description, priorityStr] = line.split('|');
      return {
        id: `task_${index + 1}`,
        description: description.trim(),
        priority: parseFloat(priorityStr?.trim() || '0.5'),
        dependencies: [],
        status: 'pending' as const
      };
    });
    
    this.state.tasks.push(...tasks);
  }
  
  async createNewTasks(): Promise<void> {
    const lastTask = this.state.completedTasks[this.state.completedTasks.length - 1];
    const lastResult = this.state.knowledge.get(`task_${lastTask.id}`);
    
    const context = `
已完成任务：${this.state.completedTasks.map(t => t.description).join(', ')}
最后任务结果：${lastResult}
    `;
    
    const newTaskDescriptions = await this.llm.generateTasks(this.state.goal, context);
    
    const newTasks = newTaskDescriptions.map((desc, index) => ({
      id: `task_${Date.now()}_${index}`,
      description: desc,
      priority: 0.7,
      dependencies: [lastTask.id],
      status: 'pending' as const
    }));
    
    this.state.tasks.push(...newTasks);
  }
}
```

---

## 最佳实践

### 1. Prompt设计

#### ✅ 好的Prompt

```
目标：构建一个Web应用
已完成：需求分析、技术选型
当前任务：设计数据库架构

请基于以上信息，生成下一步需要执行的任务。
要求：
1. 任务要具体可执行
2. 考虑依赖关系
3. 优先级要合理
```

#### ❌ 不好的Prompt

```
生成任务
```

### 2. 温度（Temperature）设置

- **低温度（0.1-0.3）**：确定性高，适合需要准确性的任务
- **中温度（0.5-0.7）**：平衡，适合大多数任务
- **高温度（0.8-1.0）**：创造性高，适合需要多样性的任务

### 3. Token管理

```typescript
class TokenManager {
  private maxTokens = 4000;
  private usedTokens = 0;
  
  canUse(tokens: number): boolean {
    return this.usedTokens + tokens <= this.maxTokens;
  }
  
  use(tokens: number): void {
    this.usedTokens += tokens;
  }
  
  reset(): void {
    this.usedTokens = 0;
  }
}
```

### 4. 错误处理

```typescript
async function callLLMWithRetry(
  prompt: string,
  maxRetries = 3
): Promise<string> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await llm.generate(prompt);
    } catch (error) {
      if (i === maxRetries - 1) throw error;
      await sleep(1000 * (i + 1)); // 指数退避
    }
  }
  throw new Error('Max retries exceeded');
}
```

### 5. 成本控制

```typescript
class CostTracker {
  private totalCost = 0;
  private costPer1KTokens = 0.002; // GPT-3.5价格示例
  
  track(tokens: number): void {
    this.totalCost += (tokens / 1000) * this.costPer1KTokens;
  }
  
  getTotalCost(): number {
    return this.totalCost;
  }
  
  reset(): void {
    this.totalCost = 0;
  }
}
```

---

## 常见问题

### Q1: LLM调用失败怎么办？

**A**: 实现重试机制和降级策略：

```typescript
async function callLLMWithFallback(prompt: string): Promise<string> {
  try {
    return await openaiLLM.generate(prompt);
  } catch (error) {
    console.warn('OpenAI failed, trying Claude...');
    try {
      return await claudeLLM.generate(prompt);
    } catch (error2) {
      // 最后的降级方案
      return 'LLM服务暂时不可用';
    }
  }
}
```

### Q2: 如何控制LLM的输出格式？

**A**: 在Prompt中明确指定格式：

```typescript
const prompt = `
请生成任务列表，格式为JSON：
{
  "tasks": [
    {"description": "任务描述", "priority": 0.8}
  ]
}
`;
```

### Q3: LLM响应太慢怎么办？

**A**: 使用流式响应或异步处理：

```typescript
// 流式响应
async function* streamLLM(prompt: string) {
  const stream = await openai.chat.completions.create({
    model: 'gpt-3.5-turbo',
    messages: [{ role: 'user', content: prompt }],
    stream: true
  });
  
  for await (const chunk of stream) {
    yield chunk.choices[0]?.delta?.content || '';
  }
}
```

### Q4: 如何减少Token消耗？

**A**: 
1. 压缩Prompt
2. 使用缓存
3. 批量处理
4. 选择合适的模型

```typescript
class PromptCache {
  private cache = new Map<string, string>();
  
  async get(prompt: string): Promise<string> {
    const key = this.hash(prompt);
    if (this.cache.has(key)) {
      return this.cache.get(key)!;
    }
    
    const result = await llm.generate(prompt);
    this.cache.set(key, result);
    return result;
  }
}
```

---

## 总结

LLM是Agent的核心组件，它赋予Agent：

1. **理解能力**：理解用户意图和目标
2. **规划能力**：制定执行计划
3. **决策能力**：选择合适的行动
4. **生成能力**：生成新任务和内容

通过合理集成LLM，Agent能够实现真正的自主性和智能性。

---

## 参考资料

- [OpenAI API文档](https://platform.openai.com/docs)
- [Anthropic API文档](https://docs.anthropic.com/)
- [LangChain文档](https://js.langchain.com/)
- [Prompt Engineering指南](https://www.promptingguide.ai/)

---

**下一步学习**：
- [ ] 学习Prompt Engineering技巧
- [ ] 学习如何优化LLM调用
- [ ] 学习如何在Agent中有效使用LLM
