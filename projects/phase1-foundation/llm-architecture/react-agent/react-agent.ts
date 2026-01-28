/**
 * ReAct Agent 完整实现
 *
 * ReAct = Reasoning (推理) + Acting (行动)
 *
 * 这个Agent展示了如何结合思考和行动来解决任务：
 * 1. Thought - 分析当前状态，决定下一步
 * 2. Action - 执行工具调用
 * 3. Observation - 获取行动结果
 * 4. 循环直到完成任务
 *
 * 使用真实的Claude AI进行推理和决策
 */

import Anthropic from "@anthropic-ai/sdk";
import * as dotenv from "dotenv";

// 加载环境变量
dotenv.config();

// ==================== 接口定义 ====================

interface Tool {
  name: string;
  description: string;
  parameters: {
    type: 'object';
    properties: Record<string, { type: string; description: string }>;
    required?: string[];
  };
  execute(args: any): Promise<string>;
}

interface ReActStep {
  thought: string;
  action?: string;
  actionInput?: string;
  observation?: string;
  finalAnswer?: string;
}

// ==================== 工具实现 ====================

class CalculatorTool implements Tool {
  name = 'calculator';
  description = '计算数学表达式，支持基本的四则运算';
  parameters = {
    type: 'object' as const,
    properties: {
      expression: {
        type: 'string',
        description: '要计算的数学表达式，如 "2 + 2" 或 "15 * 3"'
      }
    },
    required: ['expression']
  };

  async execute(args: { expression: string }): Promise<string> {
    try {
      // 安全的数学计算（实际应用中应使用math.js等库）
      const sanitized = args.expression.replace(/[^0-9+\-*/().  ]/g, '');
      const result = eval(sanitized);
      return `计算结果: ${result}`;
    } catch (error) {
      return `计算错误: ${(error as Error).message}`;
    }
  }
}

class SearchTool implements Tool {
  name = 'search';
  description = '在知识库中搜索信息（模拟）';
  parameters = {
    type: 'object' as const,
    properties: {
      query: {
        type: 'string',
        description: '要搜索的内容'
      }
    },
    required: ['query']
  };

  async execute(args: { query: string }): Promise<string> {
    // 模拟搜索结果
    const knowledgeBase: Record<string, string> = {
      '巴黎': '巴黎是法国的首都，人口约212万（市区），位于法国北部。',
      '东京': '东京是日本的首都，是世界上人口最多的都市圈，约3700万人。',
      '纽约': '纽约是美国人口最多的城市，约850万人，是全球金融中心。',
      '中国人口': '截至2023年，中国人口约14.1亿。',
      '美国人口': '截至2023年，美国人口约3.3亿。'
    };

    const query = args.query.toLowerCase();
    for (const [key, value] of Object.entries(knowledgeBase)) {
      if (query.includes(key.toLowerCase())) {
        return `搜索结果: ${value}`;
      }
    }

    return `未找到关于"${args.query}"的信息`;
  }
}

class CurrentDateTool implements Tool {
  name = 'current_date';
  description = '获取当前日期和时间';
  parameters = {
    type: 'object' as const,
    properties: {},
    required: []
  };

  async execute(args: any): Promise<string> {
    const now = new Date();
    return `当前日期时间: ${now.toLocaleString('zh-CN', {
      timeZone: 'Asia/Shanghai',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })}`;
  }
}

// ==================== LLM服务 ====================

class LLMService {
  private client: Anthropic;
  private model: string;

  constructor() {
    const apiKey = process.env.ANTHROPIC_API_KEY || process.env.ANTHROPIC_AUTH_TOKEN;
    const baseURL = process.env.ANTHROPIC_BASE_URL;

    if (!apiKey) {
      throw new Error("❌ ANTHROPIC_API_KEY 环境变量未设置");
    }

    this.client = new Anthropic({
      apiKey: apiKey,
      baseURL: baseURL,
    });
    this.model = "claude-3-5-sonnet-20241022";
  }

  async generate(prompt: string): Promise<string> {
    try {
      const response = await this.client.messages.create({
        model: this.model,
        max_tokens: 2000,
        messages: [{ role: "user", content: prompt }],
        temperature: 0.0  // 使用确定性输出
      });

      const content = response.content[0];
      return content.type === 'text' ? content.text : '';
    } catch (error) {
      console.error("❌ LLM调用失败:", error);
      throw error;
    }
  }
}

// ==================== ReAct Agent ====================

class ReActAgent {
  private llm: LLMService;
  private tools: Map<string, Tool>;
  private maxIterations: number;
  private history: ReActStep[] = [];

  constructor(llm: LLMService, tools: Tool[], maxIterations: number = 10) {
    this.llm = llm;
    this.tools = new Map(tools.map(t => [t.name, t]));
    this.maxIterations = maxIterations;
  }

  async run(task: string): Promise<string> {
    console.log('\n🎯 开始执行任务:', task);
    console.log('─'.repeat(60));

    this.history = [];
    let iteration = 0;

    while (iteration < this.maxIterations) {
      iteration++;
      console.log(`\n🔄 迭代 ${iteration}/${this.maxIterations}`);

      // 构建当前prompt
      const prompt = this.buildPrompt(task);

      // 调用LLM
      console.log('\n💭 思考中...');
      const response = await this.llm.generate(prompt);

      // 解析响应
      const step = this.parseResponse(response);
      this.history.push(step);

      // 打印思考过程
      console.log(`\n💡 Thought: ${step.thought}`);

      // 检查是否有最终答案
      if (step.finalAnswer) {
        console.log(`\n✅ Final Answer: ${step.finalAnswer}`);
        console.log('─'.repeat(60));
        console.log(`\n📊 总共执行了 ${iteration} 次迭代`);
        return step.finalAnswer;
      }

      // 执行行动
      if (step.action && step.actionInput) {
        console.log(`\n🔧 Action: ${step.action}`);
        console.log(`📥 Action Input: ${step.actionInput}`);

        const tool = this.tools.get(step.action);
        if (tool) {
          try {
            const observation = await tool.execute(
              JSON.parse(step.actionInput)
            );
            step.observation = observation;
            console.log(`📤 Observation: ${observation}`);
          } catch (error) {
            step.observation = `错误: ${(error as Error).message}`;
            console.log(`❌ Observation: ${step.observation}`);
          }
        } else {
          step.observation = `错误: 工具 '${step.action}' 不存在`;
          console.log(`❌ ${step.observation}`);
        }
      }
    }

    throw new Error(`❌ 达到最大迭代次数 (${this.maxIterations})，任务未完成`);
  }

  private buildPrompt(task: string): string {
    // 构建工具描述
    const toolDescriptions = Array.from(this.tools.values())
      .map(t => {
        const params = Object.entries(t.parameters.properties)
          .map(([key, value]) => `${key}: ${value.description}`)
          .join(', ');
        return `- ${t.name}: ${t.description}\n  参数: {${params}}`;
      })
      .join('\n');

    // 构建历史记录
    let historyText = '';
    if (this.history.length > 0) {
      historyText = '\n\n之前的步骤:\n';
      this.history.forEach((step, i) => {
        historyText += `\n步骤 ${i + 1}:\n`;
        historyText += `Thought: ${step.thought}\n`;
        if (step.action) {
          historyText += `Action: ${step.action}\n`;
          historyText += `Action Input: ${step.actionInput}\n`;
          historyText += `Observation: ${step.observation}\n`;
        }
      });
    }

    return `
你是一个使用ReAct (Reasoning + Acting) 模式的智能助手。

你可以使用以下工具：
${toolDescriptions}

任务: ${task}
${historyText}

请按照以下格式回答：

Thought: [你对当前情况的思考，决定下一步做什么]
Action: [要使用的工具名称，或者不使用工具]
Action Input: [工具的输入参数，JSON格式，如 {"expression": "2+2"}]

如果你已经可以回答问题：
Thought: 我现在知道最终答案了
Final Answer: [你的最终答案]

重要规则：
1. 每次只能采取一个行动
2. 必须先思考(Thought)再行动(Action)
3. Action Input必须是有效的JSON格式
4. 如果观察到错误，重新思考并尝试其他方法
5. 当你有足够信息回答问题时，给出Final Answer

现在开始！请给出你的Thought和Action。
`.trim();
  }

  private parseResponse(response: string): ReActStep {
    const step: ReActStep = {
      thought: '',
    };

    // 提取Thought
    const thoughtMatch = response.match(/Thought:\s*(.+?)(?=\n(?:Action|Final Answer):|$)/s);
    if (thoughtMatch) {
      step.thought = thoughtMatch[1].trim();
    }

    // 提取Final Answer
    const answerMatch = response.match(/Final Answer:\s*(.+?)$/s);
    if (answerMatch) {
      step.finalAnswer = answerMatch[1].trim();
      return step;
    }

    // 提取Action
    const actionMatch = response.match(/Action:\s*(.+?)(?=\n|$)/);
    if (actionMatch) {
      step.action = actionMatch[1].trim();
    }

    // 提取Action Input
    const inputMatch = response.match(/Action Input:\s*(\{.+?\}|\[.+?\]|.+?)(?=\n|$)/s);
    if (inputMatch) {
      step.actionInput = inputMatch[1].trim();
    }

    return step;
  }

  getHistory(): ReActStep[] {
    return [...this.history];
  }
}

// ==================== 示例场景 ====================

async function example1() {
  console.log('\n' + '='.repeat(60));
  console.log('示例1: 数学计算');
  console.log('='.repeat(60));

  const llm = new LLMService();
  const tools = [new CalculatorTool(), new CurrentDateTool()];
  const agent = new ReActAgent(llm, tools);

  const answer = await agent.run(
    "如果一个商品原价100元，打8折后又降价10元，最终价格是多少？"
  );

  console.log('\n最终答案:', answer);
}

async function example2() {
  console.log('\n' + '='.repeat(60));
  console.log('示例2: 信息检索和计算');
  console.log('='.repeat(60));

  const llm = new LLMService();
  const tools = [
    new CalculatorTool(),
    new SearchTool(),
    new CurrentDateTool()
  ];
  const agent = new ReActAgent(llm, tools);

  const answer = await agent.run(
    "巴黎的人口大约是多少？如果每人平均占地50平方米，总共需要多少平方公里？"
  );

  console.log('\n最终答案:', answer);
}

async function example3() {
  console.log('\n' + '='.repeat(60));
  console.log('示例3: 多步推理');
  console.log('='.repeat(60));

  const llm = new LLMService();
  const tools = [
    new CalculatorTool(),
    new SearchTool(),
    new CurrentDateTool()
  ];
  const agent = new ReActAgent(llm, tools);

  const answer = await agent.run(
    "中国和美国的人口差距大约是多少倍？"
  );

  console.log('\n最终答案:', answer);
}

// ==================== 主函数 ====================

async function main() {
  try {
    // 运行所有示例
    await example1();
    await new Promise(resolve => setTimeout(resolve, 2000));

    await example2();
    await new Promise(resolve => setTimeout(resolve, 2000));

    await example3();

  } catch (error) {
    console.error('\n❌ 错误:', error);
    process.exit(1);
  }
}

main();
