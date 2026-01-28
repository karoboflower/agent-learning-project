# ReAct Agent 示例

这个项目实现了一个完整的 ReAct (Reasoning + Acting) Agent，展示了如何结合推理和行动来解决复杂任务。

## 🎯 什么是ReAct

**ReAct** = **Rea**soning (推理) + **Act**ing (行动)

ReAct是一种Agent范式，通过交替进行思考和执行来解决任务：

```
循环流程：
Thought (思考) → Action (行动) → Observation (观察) → 回到思考...
```

### ReAct vs 其他模式

| 模式 | 特点 | 适用场景 |
|------|------|----------|
| **直接回答** | LLM直接给出答案 | 简单问答 |
| **CoT** | 逐步推理但不执行 | 数学推理、逻辑问题 |
| **ReAct** | 推理+执行工具 | 需要外部信息/计算的任务 |

## 🏗️ 项目结构

```
react-agent/
├── react-agent.ts      # ReAct Agent完整实现
├── package.json        # 项目配置
├── tsconfig.json       # TypeScript配置
├── .env.example        # 环境变量示例
├── .gitignore         # Git忽略文件
└── README.md          # 本文件
```

## 🚀 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 配置API密钥

```bash
cp .env.example .env
# 编辑 .env 文件，填入 ANTHROPIC_API_KEY
```

### 3. 运行示例

```bash
npm run dev
```

## 💡 核心概念

### 1. ReAct循环

```typescript
while (!完成) {
  // 1. 思考 (Thought)
  thought = await llm.think(task, history);

  // 2. 决定行动 (Action)
  if (需要工具) {
    action = await llm.decideAction(thought);
    result = await executeTool(action);

    // 3. 观察结果 (Observation)
    observation = result;
    history.push({thought, action, observation});
  } else {
    // 给出最终答案
    finalAnswer = await llm.answer(thought, history);
    break;
  }
}
```

### 2. Prompt格式

```
你是一个使用ReAct模式的助手。

可用工具：
- calculator: 计算数学表达式
- search: 搜索信息

任务: 计算2的10次方是多少？

请按以下格式回答：
Thought: [思考]
Action: [工具名]
Action Input: [输入参数JSON]

或者给出最终答案：
Thought: 我现在知道最终答案了
Final Answer: [答案]
```

### 3. 工具定义

```typescript
interface Tool {
  name: string;                // 工具名称
  description: string;         // 工具描述
  parameters: {...};           // 参数定义
  execute(args: any): Promise<string>;  // 执行函数
}
```

## 📊 示例场景

### 示例1: 数学计算

```
任务: "如果一个商品原价100元，打8折后又降价10元，最终价格是多少？"

执行过程:
1. Thought: 需要先计算打8折的价格
   Action: calculator
   Input: {"expression": "100 * 0.8"}
   Observation: 80

2. Thought: 然后从80元中减去10元
   Action: calculator
   Input: {"expression": "80 - 10"}
   Observation: 70

3. Thought: 我现在知道最终答案了
   Final Answer: 最终价格是70元
```

### 示例2: 信息检索+计算

```
任务: "巴黎的人口大约是多少？如果每人平均占地50平方米，总共需要多少平方公里？"

执行过程:
1. Thought: 首先需要查询巴黎人口
   Action: search
   Input: {"query": "巴黎"}
   Observation: 巴黎人口约212万

2. Thought: 然后计算总面积
   Action: calculator
   Input: {"expression": "2120000 * 50"}
   Observation: 106000000平方米

3. Thought: 转换为平方公里
   Action: calculator
   Input: {"expression": "106000000 / 1000000"}
   Observation: 106

4. Thought: 我现在知道最终答案了
   Final Answer: 约需要106平方公里
```

## 🔧 实现细节

### LLM服务

```typescript
class LLMService {
  async generate(prompt: string): Promise<string> {
    const response = await this.client.messages.create({
      model: "claude-3-5-sonnet-20241022",
      max_tokens: 2000,
      messages: [{ role: "user", content: prompt }],
      temperature: 0.0  // 确定性输出
    });
    return response.content[0].text;
  }
}
```

### 工具执行

```typescript
class CalculatorTool implements Tool {
  name = 'calculator';
  description = '计算数学表达式';

  async execute(args: { expression: string }): Promise<string> {
    const result = eval(args.expression);
    return `计算结果: ${result}`;
  }
}
```

### 响应解析

```typescript
private parseResponse(response: string): ReActStep {
  // 提取Thought
  const thoughtMatch = response.match(/Thought:\s*(.+?)(?=\n)/);

  // 提取Action
  const actionMatch = response.match(/Action:\s*(.+?)(?=\n)/);

  // 提取Action Input
  const inputMatch = response.match(/Action Input:\s*(\{.+?\})/);

  // 提取Final Answer
  const answerMatch = response.match(/Final Answer:\s*(.+?)$/);

  return { thought, action, actionInput, finalAnswer };
}
```

## 🎓 学习要点

### 1. ReAct的优势

- ✅ **可解释性**: 每步思考都可见
- ✅ **可纠错**: 可以根据观察调整策略
- ✅ **可扩展**: 易于添加新工具
- ✅ **通用性**: 适用于各种任务

### 2. 关键设计决策

#### Prompt设计
- 明确的格式要求
- 提供工具描述
- 包含历史记录
- 给出清晰示例

#### 解析策略
- 正则表达式提取
- 容错处理
- JSON参数验证

#### 循环控制
- 最大迭代次数
- 终止条件判断
- 错误处理

### 3. 常见问题

#### Q: 为什么需要ReAct？

A: LLM本身无法执行计算、搜索等操作，ReAct通过工具调用扩展了LLM的能力。

#### Q: 与Function Calling有什么区别？

A:
- **Function Calling**: API原生支持，LLM直接输出JSON
- **ReAct**: Prompt工程实现，通过文本格式

#### Q: 如何添加新工具？

A: 实现`Tool`接口并注册：

```typescript
class MyTool implements Tool {
  name = 'my_tool';
  description = '工具描述';
  parameters = {...};

  async execute(args: any): Promise<string> {
    // 实现逻辑
    return result;
  }
}

const agent = new ReActAgent(llm, [new MyTool()]);
```

## 🔍 扩展方向

### 1. 添加更多工具

```typescript
// 天气查询工具
class WeatherTool implements Tool {
  name = 'weather';
  description = '查询天气信息';

  async execute(args: { city: string }): Promise<string> {
    // 调用天气API
    const weather = await fetchWeather(args.city);
    return `${args.city}的天气: ${weather}`;
  }
}

// 文件操作工具
class FileOperationTool implements Tool {
  name = 'file_operation';
  description = '读写文件';

  async execute(args: { operation: string; path: string; content?: string }): Promise<string> {
    if (args.operation === 'read') {
      return await fs.promises.readFile(args.path, 'utf-8');
    } else if (args.operation === 'write') {
      await fs.promises.writeFile(args.path, args.content || '');
      return '文件写入成功';
    }
  }
}
```

### 2. 增强推理能力

```typescript
// Self-Consistency: 生成多个推理路径并投票
class SelfConsistentReActAgent extends ReActAgent {
  async run(task: string, samples: number = 3): Promise<string> {
    const answers = [];

    for (let i = 0; i < samples; i++) {
      const answer = await super.run(task);
      answers.push(answer);
    }

    return this.majorityVote(answers);
  }
}
```

### 3. 添加Memory

```typescript
class MemoryReActAgent extends ReActAgent {
  private memory: Map<string, any> = new Map();

  async run(task: string): Promise<string> {
    // 从memory获取相关信息
    const context = this.memory.get('context') || '';

    const result = await super.run(task + '\n' + context);

    // 保存重要信息到memory
    this.memory.set('last_result', result);

    return result;
  }
}
```

## 📚 参考资料

- [ReAct论文](https://arxiv.org/abs/2210.03629)
- [LangChain ReAct实现](https://python.langchain.com/docs/modules/agents/agent_types/react)
- [Anthropic工具使用文档](https://docs.anthropic.com/claude/docs/tool-use)

## 📝 总结

ReAct Agent的核心特点：

1. **思考与行动的结合**: 不仅推理，还能执行
2. **工具扩展能力**: 通过工具调用扩展LLM能力
3. **可解释的过程**: 每步思考都清晰可见
4. **灵活的架构**: 易于扩展和定制

这是构建实用AI Agent的重要范式！

---

**下一步学习**:
- [ ] 理解CoT和ReAct的区别
- [ ] 实现自己的工具
- [ ] 探索更复杂的任务场景
- [ ] 学习Memory和Planning机制
