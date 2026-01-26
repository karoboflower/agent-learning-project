/**
 * 自主Agent示例 - 集成LLM版本
 * 
 * 这个示例展示了如何在自主Agent中集成LLM
 * 包括：LLM服务封装、任务生成、结果分析等
 */

// 加载环境变量
import * as dotenv from "dotenv";
dotenv.config();

// ==================== 类型定义 ====================

interface Task {
  id: string;
  description: string;
  priority: number;
  dependencies: string[];
  status: 'pending' | 'running' | 'completed' | 'failed';
}

interface TaskResult {
  taskId: string;
  success: boolean;
  result: string;
  error?: string;
}

interface AgentState {
  goal: string;
  tasks: Task[];
  completedTasks: Task[];
  failedTasks: Task[];
  knowledge: Map<string, any>;
  status: 'idle' | 'running' | 'paused' | 'stopped' | 'stopping' | 'completed';
  metadata: {
    startTime: Date;
    lastUpdateTime: Date;
    iterationCount: number;
    totalCost: number;
  };
}

interface AgentConfig {
  maxIterations: number;
  maxCost: number;
  minPriority: number;
  retryAttempts: number;
  backoffBase: number;
}

// ==================== LLM服务接口 ====================

interface LLMService {
  generate(prompt: string, options?: { temperature?: number; maxTokens?: number }): Promise<string>;
  generateTasks(goal: string, context: string): Promise<Task[]>;
  analyzeResult(task: Task, result: TaskResult): Promise<string>;
  evaluateTaskPriority(task: Task, goal: string): Promise<number>;
}

// ==================== Gemini LLM服务实现 ====================

class GeminiLLMService implements LLMService {
  private apiKey: string;
  private model: string;

  constructor(apiKey: string, model: string = 'gemini-pro') {
    // 安全提示：API密钥应该通过环境变量传递，不要硬编码在代码中
    this.apiKey = apiKey;
    this.model = model;
  }

  async generate(prompt: string, options?: { temperature?: number; maxTokens?: number }): Promise<string> {
    // 如果没有API密钥，返回模拟响应
    if (!this.apiKey || this.apiKey === 'your-api-key-here') {
      console.warn('⚠️  LLM API密钥未配置，使用模拟响应');
      return this.getMockResponse(prompt);
    }

    // 重试机制
    const maxRetries = 3;
    let lastError: Error | null = null;

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        // 使用fetch调用Gemini API
        const url = `https://generativelanguage.googleapis.com/v1beta/models/${this.model}:generateContent?key=${this.apiKey}`;

        const response = await fetch(url, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            contents: [{
              parts: [{
                text: prompt
              }]
            }],
            generationConfig: {
              temperature: options?.temperature || 0.7,
              maxOutputTokens: options?.maxTokens || 1000
            }
          })
        });

        // 处理速率限制
        if (response.status === 429) {
          const retryAfter = response.headers.get('retry-after');
          const waitTime = retryAfter ? parseInt(retryAfter) * 1000 : Math.pow(2, attempt) * 1000;

          if (attempt < maxRetries - 1) {
            console.warn(`⏳ 遇到速率限制，等待 ${waitTime / 1000} 秒后重试... (尝试 ${attempt + 1}/${maxRetries})`);
            await this.sleep(waitTime);
            continue;
          } else {
            throw new Error(`Gemini API error: Too Many Requests (已重试 ${maxRetries} 次)`);
          }
        }

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(`Gemini API error: ${response.status} ${response.statusText} - ${JSON.stringify(errorData)}`);
        }

        const data = await response.json() as {
          candidates?: Array<{
            content?: {
              parts?: Array<{
                text?: string;
              }>;
            };
          }>;
        };

        return data.candidates?.[0]?.content?.parts?.[0]?.text || '';
      } catch (error) {
        lastError = error instanceof Error ? error : new Error(String(error));

        // 如果是速率限制错误且还有重试机会，继续重试
        if (lastError.message.includes('429') || lastError.message.includes('Too Many Requests')) {
          if (attempt < maxRetries - 1) {
            const waitTime = Math.pow(2, attempt) * 1000;
            console.warn(`⏳ 遇到速率限制，等待 ${waitTime / 1000} 秒后重试... (尝试 ${attempt + 1}/${maxRetries})`);
            await this.sleep(waitTime);
            continue;
          }
        }

        // 其他错误或重试次数用完，记录错误
        if (attempt === maxRetries - 1) {
          console.error('LLM调用失败:', lastError.message);
          // 降级到模拟响应
          return this.getMockResponse(prompt);
        }
      }
    }

    // 如果所有重试都失败，返回模拟响应
    console.warn('⚠️  LLM调用失败，使用模拟响应');
    return this.getMockResponse(prompt);
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  async generateTasks(goal: string, context: string): Promise<Task[]> {
    const prompt = `
你是一个任务规划AI。请将以下目标分解为具体的任务。

目标：${goal}

${context ? `上下文信息：\n${context}` : ''}

请生成3-5个具体的任务，每个任务一行，格式：
任务描述|优先级(0-1之间的数字)

示例：
分析需求文档|0.9
设计系统架构|0.8
编写核心代码|0.7
    `;

    const response = await this.generate(prompt, { temperature: 0.5 });
    const lines = response.split('\n').filter(line => line.trim().length > 0);

    const tasks: Task[] = [];
    let taskId = 1;

    for (const line of lines) {
      const parts = line.split('|');
      if (parts.length >= 2) {
        const description = parts[0].trim();
        const priority = parseFloat(parts[1].trim()) || 0.5;
        
        // 跳过示例行
        if (description.includes('示例') || description.includes('示例')) {
          continue;
        }

        tasks.push({
          id: `task_${taskId++}`,
          description,
          priority: Math.max(0, Math.min(1, priority)),
          dependencies: [],
          status: 'pending'
        });
      }
    }

    return tasks.length > 0 ? tasks : this.getDefaultTasks(goal);
  }

  async analyzeResult(task: Task, result: TaskResult): Promise<string> {
    const prompt = `
分析以下任务执行结果：

任务：${task.description}
成功：${result.success}
结果：${result.result}
${result.error ? `错误：${result.error}` : ''}

请简要分析这个结果，判断任务是否成功完成，以及下一步应该做什么。
    `;

    return await this.generate(prompt, { temperature: 0.3, maxTokens: 200 });
  }

  async evaluateTaskPriority(task: Task, goal: string): Promise<number> {
    const prompt = `
评估以下任务对目标的优先级：

目标：${goal}
任务：${task.description}

请给出一个0-1之间的优先级分数，0.9-1.0表示非常重要，0.5-0.8表示重要，0-0.4表示不太重要。
只返回数字，不要其他文字。
    `;

    const response = await this.generate(prompt, { temperature: 0.2, maxTokens: 10 });
    const priority = parseFloat(response.trim());
    
    return isNaN(priority) ? task.priority : Math.max(0, Math.min(1, priority));
  }

  // 模拟响应（用于演示，不需要API密钥）
  private getMockResponse(prompt: string): string {
    if (prompt.includes('任务') || prompt.includes('任务')) {
      return `分析需求文档|0.9
设计系统架构|0.8
编写核心代码|0.7
编写测试用例|0.6
部署应用|0.5`;
    }
    
    if (prompt.includes('分析') || prompt.includes('分析')) {
      return '任务执行成功，可以继续下一步。';
    }
    
    return '0.7';
  }

  private getDefaultTasks(goal: string): Task[] {
    return [
      {
        id: 'task_1',
        description: `分析目标：${goal}`,
        priority: 0.9,
        dependencies: [],
        status: 'pending'
      },
      {
        id: 'task_2',
        description: `制定执行计划`,
        priority: 0.8,
        dependencies: ['task_1'],
        status: 'pending'
      }
    ];
  }
}

// ==================== 自主Agent类（集成LLM） ====================

class AutonomousAgentWithLLM {
  private state: AgentState;
  private config: AgentConfig;
  private isRunning: boolean = false;
  private llm: LLMService;

  constructor(goal: string, llm: LLMService, config?: Partial<AgentConfig>) {
    this.state = {
      goal,
      tasks: [],
      completedTasks: [],
      failedTasks: [],
      knowledge: new Map(),
      status: 'idle',
      metadata: {
        startTime: new Date(),
        lastUpdateTime: new Date(),
        iterationCount: 0,
        totalCost: 0
      }
    };

    this.config = {
      maxIterations: 100,
      maxCost: 1000,
      minPriority: 0.5,
      retryAttempts: 3,
      backoffBase: 2,
      ...config
    };

    this.llm = llm;
  }

  // ==================== 自主运行 ====================

  async run(): Promise<void> {
    if (this.isRunning) {
      throw new Error('Agent is already running');
    }

    this.isRunning = true;
    this.state.status = 'running';

    try {
      // 1. 使用LLM生成初始任务
      if (this.state.tasks.length === 0) {
        console.log('📝 使用LLM生成初始任务...');
        await this.createInitialTasks();
        console.log(`✅ 生成了 ${this.state.tasks.length} 个初始任务`);
      }

      // 2. 自主执行循环
      while (this.isRunning && this.shouldContinue()) {
        // 自主选择下一个任务
        const task = await this.selectNextTask();
        
        if (!task) {
          // 没有更多任务，使用LLM生成新任务
          console.log('📝 使用LLM生成新任务...');
          await this.createNewTasks();
          continue;
        }

        console.log(`\n🔄 执行任务: ${task.description}`);

        // 自主执行任务
        await this.executeTask(task);

        // 更新状态
        this.updateState();

        // 检查是否完成目标
        if (this.isGoalAchieved()) {
          this.complete();
          break;
        }
      }
    } catch (error) {
      this.handleError(error instanceof Error ? error : new Error(String(error)));
    } finally {
      this.isRunning = false;
      if (this.state.status === 'running') {
        this.state.status = 'stopped';
      }
    }
  }

  stop(): void {
    this.isRunning = false;
    this.state.status = 'stopping';
  }

  pause(): void {
    this.isRunning = false;
    this.state.status = 'paused';
  }

  resume(): void {
    if (this.state.status === 'paused') {
      this.run();
    }
  }

  // ==================== 自主决策（使用LLM） ====================

  private async selectNextTask(): Promise<Task | null> {
    const availableTasks = this.state.tasks.filter(task => 
      task.status === 'pending' && 
      this.checkDependencies(task) &&
      this.checkPriority(task)
    );

    if (availableTasks.length === 0) {
      return null;
    }

    // 使用LLM评估任务优先级
    const scoredTasks = await Promise.all(
      availableTasks.map(async task => ({
        task,
        score: await this.llm.evaluateTaskPriority(task, this.state.goal)
      }))
    );

    scoredTasks.sort((a, b) => b.score - a.score);
    return scoredTasks[0].task;
  }

  private checkDependencies(task: Task): boolean {
    if (task.dependencies.length === 0) {
      return true;
    }

    const completedTaskIds = new Set(
      this.state.completedTasks.map(t => t.id)
    );

    return task.dependencies.every(depId => completedTaskIds.has(depId));
  }

  private checkPriority(task: Task): boolean {
    return task.priority >= this.config.minPriority;
  }

  // ==================== 任务执行 ====================

  private async executeTask(task: Task): Promise<void> {
    task.status = 'running';
    this.state.metadata.lastUpdateTime = new Date();

    try {
      const result = await this.executeWithRecovery(task);
      
      task.status = 'completed';
      this.state.completedTasks.push(task);
      this.state.knowledge.set(`task_${task.id}`, result);

      // 使用LLM分析结果
      const analysis = await this.llm.analyzeResult(task, result);
      console.log(`✅ 任务完成: ${task.description}`);
      console.log(`📊 分析: ${analysis}`);

      // 从待执行列表中移除
      this.state.tasks = this.state.tasks.filter(t => t.id !== task.id);
    } catch (error) {
      task.status = 'failed';
      this.state.failedTasks.push(task);
      console.error(`❌ 任务失败: ${task.description}`, error);
    }
  }

  private async executeWithRecovery(
    task: Task, 
    attempt: number = 0
  ): Promise<TaskResult> {
    try {
      const result = await this.doExecuteTask(task);
      return result;
    } catch (error) {
      if (attempt < this.config.retryAttempts) {
        const backoff = this.config.backoffBase ** attempt * 1000;
        await this.sleep(backoff);
        return await this.executeWithRecovery(task, attempt + 1);
      } else {
        throw error;
      }
    }
  }

  protected async doExecuteTask(task: Task): Promise<TaskResult> {
    // 模拟任务执行
    await this.sleep(100);

    return {
      taskId: task.id,
      success: true,
      result: `任务 "${task.description}" 执行成功`
    };
  }

  // ==================== 任务创建（使用LLM） ====================

  private async createInitialTasks(): Promise<void> {
    const tasks = await this.llm.generateTasks(this.state.goal, '');
    this.state.tasks.push(...tasks);
  }

  private async createNewTasks(): Promise<void> {
    if (this.state.completedTasks.length === 0) {
      return;
    }

    const lastTask = this.state.completedTasks[this.state.completedTasks.length - 1];
    const lastResult = this.state.knowledge.get(`task_${lastTask.id}`);

    const context = `
已完成任务：
${this.state.completedTasks.map(t => `- ${t.description}`).join('\n')}

最后任务结果：
${lastResult ? JSON.stringify(lastResult) : '无'}
    `;

    const newTasks = await this.llm.generateTasks(this.state.goal, context);
    
    // 设置依赖关系
    newTasks.forEach(task => {
      task.dependencies = [lastTask.id];
    });

    this.state.tasks.push(...newTasks);
    console.log(`✅ 生成了 ${newTasks.length} 个新任务`);
  }

  // ==================== 状态管理 ====================

  private updateState(): void {
    this.state.metadata.lastUpdateTime = new Date();
    this.state.metadata.iterationCount++;
  }

  private shouldContinue(): boolean {
    if (this.state.metadata.iterationCount >= this.config.maxIterations) {
      return false;
    }
    if (this.state.metadata.totalCost >= this.config.maxCost) {
      return false;
    }
    if (this.state.status !== 'running') {
      return false;
    }
    return true;
  }

  private isGoalAchieved(): boolean {
    return this.state.tasks.length === 0 && 
           this.state.completedTasks.length > 0;
  }

  private complete(): void {
    this.state.status = 'completed';
    this.isRunning = false;
  }

  private handleError(error: Error): void {
    console.error('Agent error:', error);
    this.state.status = 'stopped';
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // ==================== 公共方法 ====================

  getState(): Readonly<AgentState> {
    return { ...this.state };
  }

  getTasks(): Readonly<Task[]> {
    return [...this.state.tasks];
  }

  getCompletedTasks(): Readonly<Task[]> {
    return [...this.state.completedTasks];
  }
}

// ==================== 使用示例 ====================

async function example() {
  console.log('🚀 启动自主Agent示例（集成LLM）...\n');

  // 创建LLM服务
  // ⚠️ 安全提示：API密钥通过环境变量传递，不要硬编码在代码中
  // 使用方式：创建 .env 文件并添加 GEMINI_API_KEY=your-key
  const apiKey = process.env.GEMINI_API_KEY || "your-api-key-here";
  if (apiKey === "your-api-key-here") {
    console.error("❌ 错误：请设置 GEMINI_API_KEY 环境变量");
    console.error("   方式1：创建 .env 文件并添加 GEMINI_API_KEY=your-key");
    console.error("   方式2：运行前执行：export GEMINI_API_KEY=your-key");
    process.exit(1);
  }
  const llm = new GeminiLLMService(
    apiKey,
    'gemini-pro'
  );

  // 创建Agent
  const agent = new AutonomousAgentWithLLM(
    '构建一个待办事项管理Web应用',
    llm,
    {
      maxIterations: 20,
      maxCost: 500,
      minPriority: 0.3
    }
  );

  console.log('📋 Agent目标:', agent.getState().goal);
  console.log('⚙️  Agent配置:', {
    maxIterations: 20,
    maxCost: 500,
    minPriority: 0.3
  });

  // if (!process.env.GEMINI_API_KEY || process.env.GEMINI_API_KEY === 'your-api-key-here') {
  //   console.log('\n⚠️  提示：未配置GEMINI_API_KEY，将使用模拟LLM响应');
  //   console.log('   要使用真实LLM，请设置环境变量：export GEMINI_API_KEY=your-key\n');
  // }
  
  console.log('开始执行...\n');

  // 启动Agent
  await agent.run();

  // 获取结果
  const state = agent.getState();
  console.log('\n✅ Agent执行完成！');
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('📊 执行结果:');
  console.log('  - Agent状态:', state.status);
  console.log('  - 已完成任务数:', state.completedTasks.length);
  console.log('  - 失败任务数:', state.failedTasks.length);
  console.log('  - 迭代次数:', state.metadata.iterationCount);
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');
  
  if (state.completedTasks.length > 0) {
    console.log('✅ 已完成的任务:');
    state.completedTasks.forEach((task, index) => {
      console.log(`  ${index + 1}. [${task.id}] ${task.description} (优先级: ${task.priority.toFixed(2)})`);
    });
    console.log('');
  }
  
  if (state.failedTasks.length > 0) {
    console.log('❌ 失败的任务:');
    state.failedTasks.forEach((task, index) => {
      console.log(`  ${index + 1}. [${task.id}] ${task.description}`);
    });
    console.log('');
  }
}

// 如果直接运行此文件，执行示例
if (require.main === module) {
  example().catch(error => {
    console.error('❌ 执行出错:', error);
    process.exit(1);
  });
}

// 导出
export { AutonomousAgentWithLLM, GeminiLLMService, LLMService, Task, TaskResult, AgentState };
