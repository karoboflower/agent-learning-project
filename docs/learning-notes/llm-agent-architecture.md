# LLM Agent架构详解

## 📚 目录

1. [Prompt Engineering基础](#prompt-engineering基础)
2. [Chain-of-Thought (CoT)](#chain-of-thought-cot)
3. [ReAct模式](#react模式)
4. [Tool Use / Function Calling](#tool-use--function-calling)
5. [Memory机制](#memory机制)
6. [LLM Agent完整架构](#llm-agent完整架构)

---

## Prompt Engineering基础

### 什么是Prompt Engineering

**Prompt Engineering（提示工程）**是设计和优化输入提示词的技术，以引导大语言模型生成期望的输出。

### 核心原则

#### 1. 清晰明确

```typescript
// ❌ 不好的Prompt
"分析这段代码"

// ✅ 好的Prompt
"请分析以下TypeScript代码，识别潜在的性能问题和安全漏洞，并提供具体的改进建议。"
```

#### 2. 提供上下文

```typescript
const prompt = `
你是一个专业的代码审查专家，拥有10年的TypeScript开发经验。

当前任务：审查以下代码
代码语言：TypeScript
项目类型：Web应用后端API

代码：
\`\`\`typescript
${code}
\`\`\`

请提供：
1. 代码质量评分（0-100）
2. 发现的问题列表
3. 改进建议
`;
```

#### 3. 指定输出格式

```typescript
const prompt = `
分析以下代码并以JSON格式返回结果：

{
  "score": 数字,
  "issues": [
    {
      "type": "类型",
      "severity": "严重程度",
      "description": "描述"
    }
  ],
  "suggestions": ["建议1", "建议2"]
}
`;
```

#### 4. Few-Shot Learning

```typescript
const fewShotPrompt = `
我需要你将用户输入分类为不同的意图。

示例1:
输入: "今天天气怎么样？"
输出: {"intent": "weather_query", "confidence": 0.95}

示例2:
输入: "帮我预定明天的餐厅"
输出: {"intent": "reservation", "confidence": 0.9}

示例3:
输入: "给我讲个笑话"
输出: {"intent": "entertainment", "confidence": 0.85}

现在请分类：
输入: "${userInput}"
输出:
`;
```

### Prompt模板系统

```typescript
// Prompt模板类
class PromptTemplate {
  private template: string;
  private variables: string[];

  constructor(template: string) {
    this.template = template;
    this.variables = this.extractVariables(template);
  }

  private extractVariables(template: string): string[] {
    const matches = template.match(/\{(\w+)\}/g) || [];
    return matches.map(m => m.slice(1, -1));
  }

  format(values: Record<string, string>): string {
    let result = this.template;
    for (const [key, value] of Object.entries(values)) {
      result = result.replace(new RegExp(`\\{${key}\\}`, 'g'), value);
    }
    return result;
  }
}

// 使用示例
const template = new PromptTemplate(`
你是一个{role}。

任务：{task}
上下文：{context}

请提供详细的{output_type}。
`);

const prompt = template.format({
  role: "代码审查专家",
  task: "审查以下代码",
  context: "这是一个Web API项目",
  output_type: "审查报告"
});
```

---

## Chain-of-Thought (CoT)

### 什么是CoT

**Chain-of-Thought（思维链）**是一种提示技术，引导模型展示其推理过程，通过逐步思考来得出答案。

### 基本原理

```typescript
// 不使用CoT
const simplePrompt = "计算：25 * 4 + 15 / 3 - 8";

// 使用CoT
const cotPrompt = `
计算：25 * 4 + 15 / 3 - 8

让我们一步步思考：
1. 首先计算乘法：25 * 4 = ?
2. 然后计算除法：15 / 3 = ?
3. 最后进行加减：? + ? - 8 = ?

请按照这个步骤解决问题。
`;
```

### CoT的类型

#### 1. Zero-Shot CoT

```typescript
const zeroShotCoT = `
问题：${question}

让我们一步步思考这个问题。
`;
```

#### 2. Few-Shot CoT

```typescript
const fewShotCoT = `
问题：Roger有5个网球。他又买了2罐网球。每罐有3个网球。他现在有多少个网球？
思考：Roger开始有5个网球。2罐网球，每罐3个，所以是2 * 3 = 6个网球。5 + 6 = 11。
答案：11个网球。

问题：食堂开始有23个苹果。如果他们用20个做午餐，又买了6个，他们现在有多少个苹果？
思考：食堂有23个苹果。用了20个，剩下23 - 20 = 3个。又买了6个，所以3 + 6 = 9。
答案：9个苹果。

问题：${newQuestion}
思考：
`;
```

### CoT在Agent中的应用

```typescript
class CoTAgent {
  private llm: LLM;

  async solve(problem: string): Promise<string> {
    const cotPrompt = `
问题：${problem}

请按照以下步骤解决：
1. 理解问题 - 问题在问什么？
2. 识别信息 - 有哪些已知信息？
3. 制定计划 - 需要哪些步骤？
4. 执行计划 - 逐步计算
5. 验证答案 - 检查是否合理

让我们开始：
`;

    const response = await this.llm.generate(cotPrompt);
    return response;
  }
}
```

### 自洽性（Self-Consistency）

```typescript
class SelfConsistentCoTAgent {
  private llm: LLM;

  async solveWithConsistency(problem: string, numSamples: number = 5): Promise<string> {
    const answers: string[] = [];

    // 生成多个推理路径
    for (let i = 0; i < numSamples; i++) {
      const response = await this.llm.generate(`
问题：${problem}

让我们一步步思考（方法${i + 1}）：
`, { temperature: 0.7 });

      const answer = this.extractAnswer(response);
      answers.push(answer);
    }

    // 选择最常见的答案
    return this.majorityVote(answers);
  }

  private majorityVote(answers: string[]): string {
    const counts = new Map<string, number>();
    for (const answer of answers) {
      counts.set(answer, (counts.get(answer) || 0) + 1);
    }

    let maxCount = 0;
    let bestAnswer = '';
    for (const [answer, count] of counts) {
      if (count > maxCount) {
        maxCount = count;
        bestAnswer = answer;
      }
    }

    return bestAnswer;
  }
}
```

---

## ReAct模式

### 什么是ReAct

**ReAct = Reasoning + Acting**

ReAct是一种结合推理（Reasoning）和行动（Acting）的Agent范式，通过交替进行思考和执行来解决任务。

### ReAct循环

```
1. Thought（思考）→ 分析当前状态，决定下一步
2. Action（行动）→ 执行工具调用
3. Observation（观察）→ 获取行动结果
4. 回到步骤1，直到任务完成
```

### ReAct Prompt模板

```typescript
const reactPrompt = `
你是一个智能助手，可以使用以下工具：

工具列表：
- search(query: string): 在互联网上搜索信息
- calculate(expression: string): 计算数学表达式
- read_file(path: string): 读取文件内容
- write_file(path: string, content: string): 写入文件

你的任务：${task}

请按照以下格式回答：
Thought: [你的思考过程]
Action: [工具名称]
Action Input: [工具输入]

观察到结果后，继续：
Observation: [工具返回的结果]
Thought: [基于观察的新思考]
...

当你有了最终答案：
Thought: 我现在知道最终答案了
Final Answer: [你的答案]

开始！
`;
```

### ReAct Agent实现

```typescript
interface Tool {
  name: string;
  description: string;
  execute(input: string): Promise<string>;
}

class ReActAgent {
  private llm: LLM;
  private tools: Map<string, Tool>;
  private maxIterations: number = 10;

  constructor(llm: LLM, tools: Tool[]) {
    this.llm = llm;
    this.tools = new Map(tools.map(t => [t.name, t]));
  }

  async run(task: string): Promise<string> {
    const history: string[] = [];
    let iteration = 0;

    // 构建初始prompt
    const systemPrompt = this.buildSystemPrompt();
    history.push(`Task: ${task}\n`);

    while (iteration < this.maxIterations) {
      iteration++;

      // 1. 让LLM思考和决定行动
      const prompt = systemPrompt + '\n' + history.join('\n') + '\n';
      const response = await this.llm.generate(prompt);

      // 2. 解析响应
      const parsed = this.parseResponse(response);

      if (parsed.finalAnswer) {
        // 任务完成
        return parsed.finalAnswer;
      }

      // 3. 记录思考
      history.push(`Thought: ${parsed.thought}`);

      // 4. 执行行动
      if (parsed.action && parsed.actionInput) {
        history.push(`Action: ${parsed.action}`);
        history.push(`Action Input: ${parsed.actionInput}`);

        const tool = this.tools.get(parsed.action);
        if (tool) {
          try {
            const observation = await tool.execute(parsed.actionInput);
            history.push(`Observation: ${observation}\n`);
          } catch (error) {
            history.push(`Observation: Error - ${error.message}\n`);
          }
        } else {
          history.push(`Observation: Tool '${parsed.action}' not found\n`);
        }
      }
    }

    throw new Error("达到最大迭代次数，任务未完成");
  }

  private buildSystemPrompt(): string {
    const toolDescriptions = Array.from(this.tools.values())
      .map(t => `- ${t.name}: ${t.description}`)
      .join('\n');

    return `
你是一个智能助手，可以使用以下工具：

${toolDescriptions}

请按照ReAct格式回答：
Thought: [思考]
Action: [工具名称]
Action Input: [输入]

当有最终答案时：
Thought: 我现在知道最终答案了
Final Answer: [答案]
`;
  }

  private parseResponse(response: string): {
    thought?: string;
    action?: string;
    actionInput?: string;
    finalAnswer?: string;
  } {
    const result: any = {};

    // 提取Thought
    const thoughtMatch = response.match(/Thought:\s*(.+?)(?=\n|$)/);
    if (thoughtMatch) result.thought = thoughtMatch[1].trim();

    // 提取Action
    const actionMatch = response.match(/Action:\s*(.+?)(?=\n|$)/);
    if (actionMatch) result.action = actionMatch[1].trim();

    // 提取Action Input
    const inputMatch = response.match(/Action Input:\s*(.+?)(?=\n|$)/);
    if (inputMatch) result.actionInput = inputMatch[1].trim();

    // 提取Final Answer
    const answerMatch = response.match(/Final Answer:\s*(.+?)(?=\n|$)/);
    if (answerMatch) result.finalAnswer = answerMatch[1].trim();

    return result;
  }
}
```

### ReAct示例

```typescript
// 定义工具
class SearchTool implements Tool {
  name = 'search';
  description = '在互联网上搜索信息';

  async execute(query: string): Promise<string> {
    // 实际实现会调用搜索API
    return `搜索结果：关于"${query}"的信息...`;
  }
}

class CalculatorTool implements Tool {
  name = 'calculate';
  description = '计算数学表达式';

  async execute(expression: string): Promise<string> {
    try {
      const result = eval(expression);
      return `计算结果：${result}`;
    } catch (error) {
      return `计算错误：${error.message}`;
    }
  }
}

// 使用ReAct Agent
async function example() {
  const llm = new ClaudeLLM();
  const tools = [new SearchTool(), new CalculatorTool()];
  const agent = new ReActAgent(llm, tools);

  const answer = await agent.run(
    "2024年奥运会在哪里举办？该城市的人口是多少？请计算人口除以100万。"
  );

  console.log("最终答案：", answer);
}
```

---

## Tool Use / Function Calling

### 什么是Tool Use

**Tool Use（工具使用）**是让LLM能够调用外部工具和函数的能力，扩展LLM的功能边界。

### Function Calling格式

```typescript
interface FunctionDefinition {
  name: string;
  description: string;
  parameters: {
    type: 'object';
    properties: Record<string, {
      type: string;
      description: string;
      enum?: string[];
    }>;
    required: string[];
  };
}

// 示例：定义一个天气查询函数
const weatherFunction: FunctionDefinition = {
  name: 'get_weather',
  description: '获取指定城市的天气信息',
  parameters: {
    type: 'object',
    properties: {
      city: {
        type: 'string',
        description: '城市名称，如"北京"、"上海"'
      },
      unit: {
        type: 'string',
        description: '温度单位',
        enum: ['celsius', 'fahrenheit']
      }
    },
    required: ['city']
  }
};
```

### Tool Use Agent实现

```typescript
class ToolUseAgent {
  private llm: LLM;
  private tools: Map<string, Function>;
  private toolDefinitions: FunctionDefinition[];

  constructor(llm: LLM) {
    this.llm = llm;
    this.tools = new Map();
    this.toolDefinitions = [];
  }

  registerTool(definition: FunctionDefinition, implementation: Function): void {
    this.toolDefinitions.push(definition);
    this.tools.set(definition.name, implementation);
  }

  async chat(message: string): Promise<string> {
    // 1. 调用LLM，提供工具定义
    const response = await this.llm.chat(message, {
      tools: this.toolDefinitions
    });

    // 2. 检查是否需要调用工具
    if (response.tool_calls) {
      const results = [];

      // 3. 执行所有工具调用
      for (const toolCall of response.tool_calls) {
        const tool = this.tools.get(toolCall.function.name);
        if (tool) {
          const args = JSON.parse(toolCall.function.arguments);
          const result = await tool(args);
          results.push({
            tool_call_id: toolCall.id,
            result: result
          });
        }
      }

      // 4. 将工具结果返回给LLM
      const finalResponse = await this.llm.chat(message, {
        tools: this.toolDefinitions,
        tool_results: results
      });

      return finalResponse.content;
    }

    return response.content;
  }
}

// 使用示例
async function toolUseExample() {
  const agent = new ToolUseAgent(new ClaudeLLM());

  // 注册工具
  agent.registerTool(
    {
      name: 'get_weather',
      description: '获取天气信息',
      parameters: {
        type: 'object',
        properties: {
          city: { type: 'string', description: '城市名称' }
        },
        required: ['city']
      }
    },
    async (args: { city: string }) => {
      // 实际实现会调用天气API
      return {
        city: args.city,
        temperature: 25,
        condition: '晴朗'
      };
    }
  );

  const response = await agent.chat("北京今天天气怎么样？");
  console.log(response);
}
```

### 工具组合

```typescript
class ComposableTool {
  private tools: Map<string, Tool> = new Map();

  addTool(tool: Tool): void {
    this.tools.set(tool.name, tool);
  }

  async executePipeline(steps: string[]): Promise<any> {
    let result: any = null;

    for (const step of steps) {
      const [toolName, input] = this.parseStep(step);
      const tool = this.tools.get(toolName);

      if (tool) {
        result = await tool.execute(input || result);
      }
    }

    return result;
  }

  private parseStep(step: string): [string, string?] {
    const match = step.match(/(\w+)\((.+)\)/);
    if (match) {
      return [match[1], match[2]];
    }
    return [step];
  }
}
```

---

## Memory机制

### 什么是Memory

**Memory（记忆）**是Agent记录和利用历史信息的能力，包括短期记忆和长期记忆。

### Memory类型

#### 1. 短期记忆（Short-term Memory）

保存当前会话的上下文：

```typescript
class ShortTermMemory {
  private messages: Message[] = [];
  private maxMessages: number = 10;

  add(role: 'user' | 'assistant', content: string): void {
    this.messages.push({ role, content, timestamp: new Date() });

    // 保持固定长度
    if (this.messages.length > this.maxMessages) {
      this.messages.shift();
    }
  }

  getContext(): Message[] {
    return [...this.messages];
  }

  clear(): void {
    this.messages = [];
  }
}
```

#### 2. 长期记忆（Long-term Memory）

持久化存储重要信息：

```typescript
class LongTermMemory {
  private storage: Map<string, any> = new Map();

  async save(key: string, value: any, metadata?: any): Promise<void> {
    this.storage.set(key, {
      value,
      metadata,
      timestamp: new Date()
    });

    // 实际实现会存储到数据库
    await this.persist(key, value, metadata);
  }

  async retrieve(key: string): Promise<any> {
    const item = this.storage.get(key);
    return item?.value;
  }

  async search(query: string): Promise<any[]> {
    // 实际实现会使用向量搜索
    const results = [];
    for (const [key, item] of this.storage) {
      if (this.relevanceScore(query, item.value) > 0.7) {
        results.push(item.value);
      }
    }
    return results;
  }

  private async persist(key: string, value: any, metadata?: any): Promise<void> {
    // 存储到数据库
  }

  private relevanceScore(query: string, value: any): number {
    // 计算相关性得分
    return 0.8;
  }
}
```

#### 3. 工作记忆（Working Memory）

临时存储当前任务的中间结果：

```typescript
class WorkingMemory {
  private workspace: Map<string, any> = new Map();

  set(key: string, value: any): void {
    this.workspace.set(key, value);
  }

  get(key: string): any {
    return this.workspace.get(key);
  }

  clear(): void {
    this.workspace.clear();
  }

  snapshot(): Record<string, any> {
    return Object.fromEntries(this.workspace);
  }
}
```

### 完整的Memory系统

```typescript
class MemorySystem {
  private shortTerm: ShortTermMemory;
  private longTerm: LongTermMemory;
  private working: WorkingMemory;

  constructor() {
    this.shortTerm = new ShortTermMemory();
    this.longTerm = new LongTermMemory();
    this.working = new WorkingMemory();
  }

  // 添加对话消息
  addMessage(role: 'user' | 'assistant', content: string): void {
    this.shortTerm.add(role, content);
  }

  // 获取对话上下文
  getConversationContext(): Message[] {
    return this.shortTerm.getContext();
  }

  // 保存重要信息到长期记忆
  async rememberImportant(key: string, value: any): Promise<void> {
    await this.longTerm.save(key, value);
  }

  // 回忆相关信息
  async recall(query: string): Promise<any[]> {
    return await this.longTerm.search(query);
  }

  // 临时存储工作数据
  setWorkingData(key: string, value: any): void {
    this.working.set(key, value);
  }

  getWorkingData(key: string): any {
    return this.working.get(key);
  }

  // 清理工作记忆
  clearWorkingMemory(): void {
    this.working.clear();
  }
}
```

### 带Memory的Agent

```typescript
class MemoryAgent {
  private llm: LLM;
  private memory: MemorySystem;

  constructor(llm: LLM) {
    this.llm = llm;
    this.memory = new MemorySystem();
  }

  async chat(userMessage: string): Promise<string> {
    // 1. 添加用户消息到短期记忆
    this.memory.addMessage('user', userMessage);

    // 2. 从长期记忆中回忆相关信息
    const relevantMemories = await this.memory.recall(userMessage);

    // 3. 构建带上下文的prompt
    const context = this.memory.getConversationContext();
    const prompt = this.buildPrompt(userMessage, context, relevantMemories);

    // 4. 调用LLM
    const response = await this.llm.generate(prompt);

    // 5. 添加助手响应到短期记忆
    this.memory.addMessage('assistant', response);

    // 6. 识别并保存重要信息到长期记忆
    await this.extractAndSaveImportant(userMessage, response);

    return response;
  }

  private buildPrompt(
    message: string,
    context: Message[],
    memories: any[]
  ): string {
    let prompt = '';

    // 添加相关记忆
    if (memories.length > 0) {
      prompt += '相关背景信息：\n';
      memories.forEach(m => {
        prompt += `- ${JSON.stringify(m)}\n`;
      });
      prompt += '\n';
    }

    // 添加对话历史
    prompt += '对话历史：\n';
    context.forEach(msg => {
      prompt += `${msg.role}: ${msg.content}\n`;
    });

    prompt += `\n当前消息：${message}\n\n请回复：`;

    return prompt;
  }

  private async extractAndSaveImportant(
    userMessage: string,
    response: string
  ): Promise<void> {
    // 使用LLM提取重要信息
    const extractPrompt = `
从以下对话中提取需要记住的重要信息（如用户偏好、事实、决策等）：

用户：${userMessage}
助手：${response}

如果有重要信息，以JSON格式返回：
{
  "key": "信息类别",
  "value": "具体内容"
}

如果没有重要信息，返回null。
`;

    const extraction = await this.llm.generate(extractPrompt);

    try {
      const info = JSON.parse(extraction);
      if (info && info.key && info.value) {
        await this.memory.rememberImportant(info.key, info.value);
      }
    } catch {
      // 解析失败，忽略
    }
  }
}
```

---

## LLM Agent完整架构

### 架构图

```
┌─────────────────────────────────────────────────────┐
│                   User Interface                    │
└─────────────────┬───────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────────────┐
│                  Agent Controller                   │
│  - 任务规划                                          │
│  - 流程控制                                          │
│  - 错误处理                                          │
└─────────────────┬───────────────────────────────────┘
                  │
        ┌─────────┼──────────┬──────────┐
        │         │          │          │
        ▼         ▼          ▼          ▼
┌───────────┐ ┌────────┐ ┌──────┐ ┌─────────┐
│  Memory   │ │  LLM   │ │Tools │ │Planning │
│  System   │ │  Core  │ │System│ │ Module  │
└───────────┘ └────────┘ └──────┘ └─────────┘
```

### 完整实现

```typescript
class LLMAgent {
  private llm: LLM;
  private memory: MemorySystem;
  private tools: Map<string, Tool>;
  private planner: TaskPlanner;

  constructor(config: AgentConfig) {
    this.llm = new LLM(config.apiKey);
    this.memory = new MemorySystem();
    this.tools = new Map();
    this.planner = new TaskPlanner();
  }

  async execute(task: string): Promise<string> {
    console.log(`执行任务: ${task}`);

    try {
      // 1. 任务规划
      const plan = await this.planner.createPlan(task, this.llm);
      console.log('任务计划:', plan);

      // 2. 执行计划
      const results = [];
      for (const step of plan.steps) {
        const result = await this.executeStep(step);
        results.push(result);

        // 保存中间结果到工作记忆
        this.memory.setWorkingData(`step_${step.id}`, result);
      }

      // 3. 整合结果
      const finalResult = await this.integrateResults(task, results);

      // 4. 保存重要信息
      await this.memory.rememberImportant(task, finalResult);

      return finalResult;

    } catch (error) {
      console.error('任务执行失败:', error);
      throw error;
    }
  }

  private async executeStep(step: PlanStep): Promise<any> {
    console.log(`执行步骤: ${step.description}`);

    // 使用ReAct模式执行
    const reactAgent = new ReActAgent(this.llm, Array.from(this.tools.values()));
    return await reactAgent.run(step.description);
  }

  private async integrateResults(task: string, results: any[]): Promise<string> {
    const prompt = `
任务: ${task}

各步骤结果:
${results.map((r, i) => `步骤${i + 1}: ${r}`).join('\n')}

请整合以上结果，给出最终答案。
`;

    return await this.llm.generate(prompt);
  }

  registerTool(tool: Tool): void {
    this.tools.set(tool.name, tool);
  }
}
```

---

## 总结

LLM Agent架构的核心要素：

1. **Prompt Engineering** - 精心设计输入提示
2. **CoT** - 引导逐步推理
3. **ReAct** - 结合思考和行动
4. **Tool Use** - 扩展功能边界
5. **Memory** - 保持上下文和知识

这些技术相互配合，构成了强大的LLM Agent系统。

---

## 参考资料

- [ReAct论文](https://arxiv.org/abs/2210.03629)
- [Chain-of-Thought论文](https://arxiv.org/abs/2201.11903)
- [Prompt Engineering指南](https://www.promptingguide.ai/)
- [LangChain文档](https://python.langchain.com/docs/get_started/introduction)

---

**学习检查清单**：
- [ ] 理解Prompt Engineering原则
- [ ] 掌握CoT推理技术
- [ ] 理解ReAct循环模式
- [ ] 实现Tool Use机制
- [ ] 构建Memory系统
- [ ] 整合完整Agent架构
