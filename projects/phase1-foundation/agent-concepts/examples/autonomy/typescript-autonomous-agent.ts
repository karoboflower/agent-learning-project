/**
 * 自主Agent示例 - TypeScript版本
 * 
 * 这个示例展示了如何实现一个具有自主性的Agent
 * 包括：内部状态管理、自主决策、自主执行循环
 */

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

// ==================== 自主Agent类 ====================

class AutonomousAgent {
  private state: AgentState;
  private config: AgentConfig;
  private isRunning: boolean = false;

  constructor(goal: string, config?: Partial<AgentConfig>) {
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
  }

  // ==================== 自主运行 ====================

  /**
   * 启动Agent的自主运行循环
   */
  async run(): Promise<void> {
    if (this.isRunning) {
      throw new Error('Agent is already running');
    }

    this.isRunning = true;
    this.state.status = 'running';

    try {
      // 1. 生成初始任务
      if (this.state.tasks.length === 0) {
        await this.createInitialTasks();
      }

      // 2. 自主执行循环
      while (this.isRunning && this.shouldContinue()) {
        // 自主选择下一个任务
        const task = await this.selectNextTask();
        
        if (!task) {
          // 没有更多任务，尝试生成新任务
          await this.createNewTasks();
          continue;
        }

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

  /**
   * 停止Agent
   */
  stop(): void {
    this.isRunning = false;
    this.state.status = 'stopping';
  }

  /**
   * 暂停Agent
   */
  pause(): void {
    this.isRunning = false;
    this.state.status = 'paused';
  }

  /**
   * 恢复Agent
   */
  resume(): void {
    if (this.state.status === 'paused') {
      this.run();
    }
  }

  // ==================== 自主决策 ====================

  /**
   * 自主选择下一个要执行的任务
   */
  private async selectNextTask(): Promise<Task | null> {
    // 1. 过滤可用任务（依赖已满足、优先级足够）
    const availableTasks = this.state.tasks.filter(task => 
      task.status === 'pending' && 
      this.checkDependencies(task) &&
      this.checkPriority(task)
    );

    if (availableTasks.length === 0) {
      return null;
    }

    // 2. 自主评估任务优先级
    const scoredTasks = await Promise.all(
      availableTasks.map(async task => ({
        task,
        score: await this.evaluateTaskPriority(task)
      }))
    );

    // 3. 选择得分最高的任务
    scoredTasks.sort((a, b) => b.score - a.score);
    return scoredTasks[0].task;
  }

  /**
   * 评估任务优先级（自主决策）
   */
  private async evaluateTaskPriority(task: Task): Promise<number> {
    let score = task.priority;

    // 考虑任务依赖的完成情况
    const dependencyProgress = this.calculateDependencyProgress(task);
    score += dependencyProgress * 0.2;

    // 考虑任务对目标的贡献度
    const contribution = await this.estimateContribution(task);
    score += contribution * 0.3;

    // 考虑资源消耗
    const resourceCost = this.estimateResourceCost(task);
    score -= resourceCost * 0.1;

    return Math.max(0, Math.min(1, score));
  }

  /**
   * 检查任务依赖是否满足
   */
  private checkDependencies(task: Task): boolean {
    if (task.dependencies.length === 0) {
      return true;
    }

    const completedTaskIds = new Set(
      this.state.completedTasks.map(t => t.id)
    );

    return task.dependencies.every(depId => completedTaskIds.has(depId));
  }

  /**
   * 检查任务优先级是否足够
   */
  private checkPriority(task: Task): boolean {
    return task.priority >= this.config.minPriority;
  }

  /**
   * 计算依赖完成进度
   */
  private calculateDependencyProgress(task: Task): number {
    if (task.dependencies.length === 0) {
      return 1;
    }

    const completedTaskIds = new Set(
      this.state.completedTasks.map(t => t.id)
    );

    const completedDeps = task.dependencies.filter(depId => 
      completedTaskIds.has(depId)
    ).length;

    return completedDeps / task.dependencies.length;
  }

  /**
   * 估算任务对目标的贡献度
   */
  private async estimateContribution(task: Task): Promise<number> {
    // 这里可以使用LLM来评估任务对目标的贡献
    // 简化实现：基于任务描述的关键词匹配
    const goalKeywords = this.state.goal.toLowerCase().split(/\s+/);
    const taskKeywords = task.description.toLowerCase().split(/\s+/);
    
    const commonKeywords = goalKeywords.filter(kw => 
      taskKeywords.includes(kw)
    ).length;

    return Math.min(1, commonKeywords / Math.max(goalKeywords.length, 1));
  }

  /**
   * 估算资源消耗
   */
  private estimateResourceCost(task: Task): number {
    // 简化实现：基于任务描述长度
    return Math.min(1, task.description.length / 100);
  }

  // ==================== 任务执行 ====================

  /**
   * 执行任务（带错误恢复）
   */
  private async executeTask(task: Task): Promise<void> {
    task.status = 'running';
    this.state.metadata.lastUpdateTime = new Date();

    try {
      const result = await this.executeWithRecovery(task);
      
      task.status = 'completed';
      this.state.completedTasks.push(task);
      this.state.knowledge.set(`task_${task.id}`, result);

      // 从待执行列表中移除
      this.state.tasks = this.state.tasks.filter(t => t.id !== task.id);
    } catch (error) {
      task.status = 'failed';
      this.state.failedTasks.push(task);
      console.error(`Task ${task.id} failed:`, error);
    }
  }

  /**
   * 带错误恢复的执行
   */
  private async executeWithRecovery(
    task: Task, 
    attempt: number = 0
  ): Promise<TaskResult> {
    try {
      // 执行任务
      const result = await this.doExecuteTask(task);
      return result;
    } catch (error) {
      // 分析错误
      const errorAnalysis = this.analyzeError(error as Error);

      // 决定是否重试
      if (errorAnalysis.isRetryable && attempt < this.config.retryAttempts) {
        // 等待后重试（指数退避）
        const backoff = this.config.backoffBase ** attempt * 1000;
        await this.sleep(backoff);

        // 调整任务后重试
        const adjustedTask = this.adjustTaskForRetry(task, errorAnalysis);
        return await this.executeWithRecovery(adjustedTask, attempt + 1);
      } else {
        // 无法恢复，抛出错误
        throw error;
      }
    }
  }

  /**
   * 实际执行任务（需要子类实现）
   */
  protected async doExecuteTask(task: Task): Promise<TaskResult> {
    // 这里是示例实现，实际应该调用具体的工具或服务
    await this.sleep(100); // 模拟执行时间

    return {
      taskId: task.id,
      success: true,
      result: `Task ${task.id} completed successfully`
    };
  }

  /**
   * 分析错误
   */
  private analyzeError(error: Error): {
    type: string;
    isRetryable: boolean;
    suggestedFix: string;
  } {
    // 简化实现：基于错误消息判断
    const message = error.message.toLowerCase();
    
    let isRetryable = false;
    let type = 'unknown';
    let suggestedFix = '';

    if (message.includes('timeout') || message.includes('network')) {
      type = 'network_error';
      isRetryable = true;
      suggestedFix = 'Retry with backoff';
    } else if (message.includes('rate limit')) {
      type = 'rate_limit';
      isRetryable = true;
      suggestedFix = 'Wait and retry';
    } else if (message.includes('invalid')) {
      type = 'validation_error';
      isRetryable = false;
      suggestedFix = 'Fix input data';
    }

    return { type, isRetryable, suggestedFix };
  }

  /**
   * 为重试调整任务
   */
  private adjustTaskForRetry(
    task: Task, 
    errorAnalysis: ReturnType<typeof this.analyzeError>
  ): Task {
    // 可以根据错误类型调整任务
    // 例如：添加重试标记、调整参数等
    return {
      ...task,
      description: `${task.description} [retry after ${errorAnalysis.type}]`
    };
  }

  // ==================== 任务创建 ====================

  /**
   * 创建初始任务
   */
  private async createInitialTasks(): Promise<void> {
    // 这里应该使用LLM来生成初始任务
    // 简化实现：基于目标创建示例任务
    const tasks: Task[] = [
      {
        id: 'task_1',
        description: `Analyze the goal: ${this.state.goal}`,
        priority: 0.9,
        dependencies: [],
        status: 'pending'
      },
      {
        id: 'task_2',
        description: `Break down the goal into subtasks`,
        priority: 0.8,
        dependencies: ['task_1'],
        status: 'pending'
      }
    ];

    this.state.tasks.push(...tasks);
  }

  /**
   * 创建新任务（基于执行结果）
   */
  private async createNewTasks(): Promise<void> {
    if (this.state.completedTasks.length === 0) {
      return;
    }

    const lastTask = this.state.completedTasks[this.state.completedTasks.length - 1];
    const lastResult = this.state.knowledge.get(`task_${lastTask.id}`);

    // 这里应该使用LLM来生成新任务
    // 简化实现：基于最后完成的任务创建新任务
    if (lastResult && this.shouldCreateNewTasks(lastResult)) {
      const newTask: Task = {
        id: `task_${Date.now()}`,
        description: `Continue work based on: ${lastTask.description}`,
        priority: 0.7,
        dependencies: [lastTask.id],
        status: 'pending'
      };

      this.state.tasks.push(newTask);
    }
  }

  /**
   * 判断是否应该创建新任务
   */
  private shouldCreateNewTasks(result: any): boolean {
    // 简化实现：如果还有未完成的任务，就不创建新任务
    return this.state.tasks.length === 0;
  }

  // ==================== 状态管理 ====================

  /**
   * 更新状态
   */
  private updateState(): void {
    this.state.metadata.lastUpdateTime = new Date();
    this.state.metadata.iterationCount++;
  }

  /**
   * 检查是否应该继续
   */
  private shouldContinue(): boolean {
    // 检查迭代次数
    if (this.state.metadata.iterationCount >= this.config.maxIterations) {
      return false;
    }

    // 检查成本
    if (this.state.metadata.totalCost >= this.config.maxCost) {
      return false;
    }

    // 检查状态
    if (this.state.status !== 'running') {
      return false;
    }

    return true;
  }

  /**
   * 检查目标是否达成
   */
  private isGoalAchieved(): boolean {
    // 这里应该使用LLM来评估目标是否达成
    // 简化实现：如果所有任务都完成了
    return this.state.tasks.length === 0 && 
           this.state.completedTasks.length > 0;
  }

  /**
   * 完成Agent
   */
  private complete(): void {
    this.state.status = 'completed';
    this.isRunning = false;
  }

  /**
   * 处理错误
   */
  private handleError(error: Error): void {
    console.error('Agent error:', error);
    this.state.status = 'stopped';
  }

  // ==================== 工具方法 ====================

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // ==================== 公共方法 ====================

  /**
   * 获取当前状态
   */
  getState(): Readonly<AgentState> {
    return { ...this.state };
  }

  /**
   * 获取任务列表
   */
  getTasks(): Readonly<Task[]> {
    return [...this.state.tasks];
  }

  /**
   * 获取已完成任务
   */
  getCompletedTasks(): Readonly<Task[]> {
    return [...this.state.completedTasks];
  }
}

// ==================== 使用示例 ====================

async function example() {
  console.log('🚀 启动自主Agent示例...\n');
  
  // 创建Agent
  const agent = new AutonomousAgent('Build a web application', {
    maxIterations: 50,
    maxCost: 500,
    minPriority: 0.5
  });

  console.log('📋 Agent目标:', agent.getState().goal);
  console.log('⚙️  Agent配置:', {
    maxIterations: 50,
    maxCost: 500,
    minPriority: 0.5
  });
  console.log('\n开始执行...\n');

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
  console.log('  - 总成本:', state.metadata.totalCost);
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n');
  
  // 显示已完成的任务详情
  if (state.completedTasks.length > 0) {
    console.log('✅ 已完成的任务:');
    state.completedTasks.forEach((task, index) => {
      console.log(`  ${index + 1}. [${task.id}] ${task.description}`);
    });
    console.log('');
  }
  
  // 显示失败的任务
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
export { AutonomousAgent, Task, TaskResult, AgentState, AgentConfig };
