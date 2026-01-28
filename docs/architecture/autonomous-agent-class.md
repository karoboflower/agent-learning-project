# AutonomousAgent类深度分析

## 📚 目录

1. [类概述](#类概述)
2. [核心属性](#核心属性)
3. [生命周期管理](#生命周期管理)
4. [主运行循环](#主运行循环)
5. [工作队列管理](#工作队列管理)
6. [错误处理机制](#错误处理机制)
7. [消息系统](#消息系统)
8. [完整实现](#完整实现)

---

## 类概述

`AutonomousAgent`是AgentGPT前端的核心类，负责Agent的整个生命周期管理和任务执行。

### 设计目标

1. **自主运行**: Agent能够独立运行，不需要持续的用户输入
2. **可控性**: 提供暂停、恢复、停止等控制功能
3. **可观察性**: 通过消息系统让用户了解Agent的思考和行动
4. **容错性**: 完善的错误处理和重试机制

### 类图

```
┌─────────────────────────────────────────┐
│         AutonomousAgent                 │
├─────────────────────────────────────────┤
│ - model: AgentModel                     │
│ - workLog: AgentWork[]                  │
│ - api: AgentApi                         │
│ - messageCallback: Function             │
│ - isRunning: boolean                    │
│ - errorHandler: ErrorHandler            │
├─────────────────────────────────────────┤
│ + constructor(goal, api, callback)      │
│ + async run(): Promise<void>            │
│ + async runWork(work): Promise<void>    │
│ + stop(): void                          │
│ + pause(): void                         │
│ + resume(): void                        │
│ - sendMessage(message): void            │
│ - addTasksIfWorklogEmpty(): void        │
│ - stopAgent(): void                     │
└─────────────────────────────────────────┘
```

---

## 核心属性

### 1. model: AgentModel

存储Agent的所有状态信息。

```typescript
private model: AgentModel;

// 包含：
// - goal: 目标
// - lifecycle: 生命周期状态
// - tasks: 任务列表
// - completedTasks: 已完成任务
// - messages: 消息历史
// - currentAnalysis: 当前分析结果
// - summary: 最终总结
```

**作用**:
- 集中管理状态
- 提供查询接口
- 支持状态持久化

### 2. workLog: AgentWork[]

工作队列，存储待执行的Work。

```typescript
private workLog: AgentWork[];

// 初始化时添加第一个Work
this.workLog = [new StartGoalWork(this, goal)];

// 队列操作
workLog.shift();    // 移除已完成的Work
workLog.push(next); // 添加新的Work
workLog.unshift(work); // 重试时重新加入队列首部
```

**特点**:
- 先进先出(FIFO)队列
- 支持优先插入(unshift)用于重试
- 动态增长的队列

### 3. api: AgentApi

与后端通信的API接口。

```typescript
private api: AgentApi;

// 提供的方法：
// - startGoal(): 开始目标
// - analyzeTask(): 分析任务
// - executeTask(): 执行任务
// - createTasks(): 创建新任务
// - summarize(): 生成总结
```

**职责**:
- HTTP请求封装
- 错误处理
- 重试逻辑

### 4. messageCallback: Function

向UI发送消息的回调函数。

```typescript
private messageCallback: (message: Message) => void;

// 调用示例
this.messageCallback({
  type: MessageType.THINKING,
  content: "正在思考下一步..."
});
```

**作用**:
- 实时更新UI
- 展示Agent的思考过程
- 提供用户反馈

### 5. isRunning: boolean

控制主循环的运行状态。

```typescript
private isRunning: boolean = false;

// 控制循环
while (this.isRunning && this.workLog[0]) {
  // ...
}
```

---

## 生命周期管理

### 生命周期状态

```typescript
enum AgentLifecycle {
  IDLE = "idle",           // 空闲 - 初始状态
  RUNNING = "running",     // 运行中 - 正在执行任务
  PAUSING = "pausing",     // 暂停中 - 正在暂停
  PAUSED = "paused",       // 已暂停 - 暂停完成
  STOPPING = "stopping",   // 停止中 - 正在停止
  STOPPED = "stopped"      // 已停止 - 停止完成
}
```

### 状态转换图

```
     ┌──────┐
     │ IDLE │  初始状态
     └───┬──┘
         │ start()
         ▼
     ┌─────────┐
  ┌─→│ RUNNING │◄─┐
  │  └────┬────┘  │
  │       │       │
  │   pause()  resume()
  │       │       │
  │       ▼       │
  │  ┌─────────┐ │
  │  │PAUSING  │ │
  │  └────┬────┘ │
  │       │      │
  │       ▼      │
  │  ┌─────────┐ │
  └──│ PAUSED  │─┘
     └─────────┘
         │
      stop()
         │
         ▼
     ┌─────────┐
     │STOPPING │
     └────┬────┘
          │
          ▼
     ┌─────────┐
     │ STOPPED │  最终状态
     └─────────┘
```

### 状态转换实现

```typescript
class AutonomousAgent {
  // 1. 启动
  async start(): Promise<void> {
    if (this.isRunning) {
      throw new Error("Agent is already running");
    }

    this.isRunning = true;
    this.model.setLifecycle(AgentLifecycle.RUNNING);

    await this.run();
  }

  // 2. 暂停
  pause(): void {
    if (this.model.getLifecycle() === AgentLifecycle.RUNNING) {
      this.model.setLifecycle(AgentLifecycle.PAUSING);

      this.sendMessage({
        type: MessageType.SYSTEM,
        content: "正在暂停..."
      });
    }
  }

  // 3. 恢复
  resume(): void {
    if (this.model.getLifecycle() === AgentLifecycle.PAUSED) {
      this.model.setLifecycle(AgentLifecycle.RUNNING);

      this.sendMessage({
        type: MessageType.SYSTEM,
        content: "继续执行..."
      });

      // 继续执行
      this.run();
    }
  }

  // 4. 停止
  stop(): void {
    this.model.setLifecycle(AgentLifecycle.STOPPING);

    this.sendMessage({
      type: MessageType.SYSTEM,
      content: "正在停止..."
    });

    this.stopAgent();
  }

  private stopAgent(): void {
    this.isRunning = false;
    this.model.setLifecycle(AgentLifecycle.STOPPED);

    this.sendMessage({
      type: MessageType.SYSTEM,
      content: "✅ Agent已停止"
    });
  }
}
```

---

## 主运行循环

### run()方法完整实现

```typescript
async run(): Promise<void> {
  // === 阶段1: 初始化 ===
  this.model.setLifecycle(AgentLifecycle.RUNNING);
  this.isRunning = true;

  this.sendMessage({
    type: MessageType.SYSTEM,
    content: "🚀 Agent开始运行"
  });

  try {
    // === 阶段2: 主循环 ===
    while (this.workLog[0]) {
      // 2.1 检查暂停状态
      if (this.model.getLifecycle() === AgentLifecycle.PAUSING) {
        this.model.setLifecycle(AgentLifecycle.PAUSED);

        this.sendMessage({
          type: MessageType.SYSTEM,
          content: "⏸️ Agent已暂停"
        });
      }

      // 2.2 检查是否应该继续运行
      if (this.model.getLifecycle() !== AgentLifecycle.RUNNING) {
        // 暂停或停止，退出循环
        return;
      }

      // 2.3 检查运行标志
      if (!this.isRunning) {
        return;
      }

      // 2.4 获取当前工作
      const work = this.workLog[0];

      // 2.5 发送状态消息
      this.sendMessage({
        type: MessageType.STATUS,
        content: work.getStatusMessage()
      });

      // 2.6 执行工作
      try {
        await this.runWork(work);

        // 2.7 成功执行，移除工作
        this.workLog.shift();

        // 2.8 获取下一个工作
        const next = work.next();
        if (next) {
          this.workLog.push(next);
        }

      } catch (error) {
        // 2.9 错误处理
        await this.handleWorkError(work, error as Error);

        // 如果错误无法恢复，退出循环
        if (!this.isRunning) {
          return;
        }
      }

      // 2.10 检查工作队列
      this.addTasksIfWorklogEmpty();

      // 2.11 短暂延迟，避免过快执行
      await this.sleep(100);
    }

    // === 阶段3: 完成 ===
    this.sendMessage({
      type: MessageType.SYSTEM,
      content: "✅ 所有任务已完成"
    });

    this.stopAgent();

  } catch (error) {
    // === 阶段4: 全局错误处理 ===
    console.error("Agent run failed:", error);

    this.sendMessage({
      type: MessageType.ERROR,
      content: `Agent运行失败: ${(error as Error).message}`
    });

    this.stopAgent();
  }
}

private sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms));
}
```

### 执行流程图

```
┌──────────────┐
│  start()     │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│ 初始化状态   │
└──────┬───────┘
       │
       ▼
┌──────────────────────────────┐
│ while (workLog[0])           │
│                              │
│  ┌────────────────────┐     │
│  │ 1. 检查暂停/停止   │     │
│  └──────┬─────────────┘     │
│         │                   │
│         ▼                   │
│  ┌────────────────────┐     │
│  │ 2. 获取当前Work    │     │
│  └──────┬─────────────┘     │
│         │                   │
│         ▼                   │
│  ┌────────────────────┐     │
│  │ 3. 执行Work        │     │
│  └──────┬─────────────┘     │
│         │                   │
│         ▼                   │
│  ┌────────────────────┐     │
│  │ 4. 处理结果        │     │
│  └──────┬─────────────┘     │
│         │                   │
│         ▼                   │
│  ┌────────────────────┐     │
│  │ 5. 获取下一个Work  │     │
│  └──────┬─────────────┘     │
│         │                   │
│         └──────┐            │
│                │            │
└────────────────┼────────────┘
                 │
                 ▼
         ┌──────────────┐
         │  stopAgent() │
         └──────────────┘
```

---

## 工作队列管理

### addTasksIfWorklogEmpty()

```typescript
private addTasksIfWorklogEmpty(): void {
  // 1. 检查工作队列是否为空
  if (this.workLog.length > 0) {
    return; // 队列不为空，无需操作
  }

  // 2. 检查是否还有待处理任务
  const nextTask = this.model.getNextTask();
  if (!nextTask) {
    return; // 没有待处理任务
  }

  // 3. 创建新的AnalyzeTaskWork
  const work = new AnalyzeTaskWork(this, nextTask);

  // 4. 添加到队列
  this.workLog.push(work);

  // 5. 发送消息
  this.sendMessage({
    type: MessageType.SYSTEM,
    content: `开始新任务: ${nextTask}`
  });
}
```

### 队列操作场景

#### 场景1: 正常执行流程

```typescript
// 初始状态
workLog = [StartGoalWork]

// 执行StartGoalWork
await runWork(StartGoalWork)

// 移除并获取下一个
workLog.shift()  // workLog = []
next = StartGoalWork.next()  // AnalyzeTaskWork
workLog.push(next)  // workLog = [AnalyzeTaskWork]

// 循环继续...
```

#### 场景2: 重试场景

```typescript
// 当前状态
workLog = [ExecuteTaskWork]

// 执行失败
await runWork(ExecuteTaskWork)  // 抛出NetworkError

// 错误处理决定重试
if (shouldRetry(ExecuteTaskWork, error)) {
  workLog.unshift(ExecuteTaskWork)  // 重新加入队列首部
}

// 下次循环会重新执行ExecuteTaskWork
```

#### 场景3: 任务补充

```typescript
// 当前状态
workLog = [CreateTaskWork]

// 执行CreateTaskWork
await runWork(CreateTaskWork)

// 移除
workLog.shift()  // workLog = []

// 获取下一个 - CreateTaskWork决定是否有下一个任务
next = CreateTaskWork.next()  // 可能是AnalyzeTaskWork或SummarizeWork
workLog.push(next)

// 如果next返回undefined
if (!next) {
  // 队列为空，检查是否有待处理任务
  addTasksIfWorklogEmpty()
}
```

---

## 错误处理机制

### 错误分类

```typescript
// 1. 网络错误
class NetworkError extends Error {
  retryable = true;
  backoffMs = 1000;
}

// 2. API错误
class ApiError extends Error {
  constructor(
    message: string,
    public statusCode: number
  ) {
    super(message);
    this.retryable = statusCode >= 500; // 服务器错误可重试
  }
}

// 3. 速率限制错误
class RateLimitError extends Error {
  retryable = true;
  backoffMs = 30000; // 30秒后重试
}

// 4. 验证错误
class ValidationError extends Error {
  retryable = false; // 验证错误不可重试
}
```

### handleWorkError()实现

```typescript
private async handleWorkError(
  work: AgentWork,
  error: Error
): Promise<void> {
  console.error(`Work ${work.getType()} failed:`, error);

  // 1. 发送错误消息
  this.sendMessage({
    type: MessageType.ERROR,
    content: `❌ ${work.getType()}: ${error.message}`
  });

  // 2. 网络错误 - 重试
  if (error instanceof NetworkError) {
    if (await this.retryWork(work, error)) {
      this.sendMessage({
        type: MessageType.SYSTEM,
        content: "网络错误，正在重试..."
      });
      return;
    } else {
      this.sendMessage({
        type: MessageType.ERROR,
        content: "网络错误，重试失败，停止Agent"
      });
      this.stopAgent();
      return;
    }
  }

  // 3. 速率限制 - 延迟重试
  if (error instanceof RateLimitError) {
    this.sendMessage({
      type: MessageType.SYSTEM,
      content: "API速率限制，30秒后重试..."
    });

    // 暂停Agent
    this.pause();

    // 30秒后重试
    setTimeout(() => {
      this.workLog.unshift(work);
      this.resume();
    }, 30000);

    return;
  }

  // 4. API错误 - 根据状态码决定
  if (error instanceof ApiError) {
    if (error.statusCode >= 500) {
      // 服务器错误 - 重试
      if (await this.retryWork(work, error)) {
        return;
      }
    }

    // 客户端错误或重试失败 - 停止
    this.sendMessage({
      type: MessageType.ERROR,
      content: `API错误 (${error.statusCode}): ${error.message}`
    });
    this.stopAgent();
    return;
  }

  // 5. 其他错误 - 停止
  this.sendMessage({
    type: MessageType.ERROR,
    content: `未知错误: ${error.message}`
  });
  this.stopAgent();
}
```

### 重试策略

```typescript
private retryAttempts = new Map<AgentWork, number>();
private readonly MAX_RETRIES = 3;

private async retryWork(
  work: AgentWork,
  error: Error
): Promise<boolean> {
  // 1. 获取当前重试次数
  const attempts = this.retryAttempts.get(work) || 0;

  // 2. 检查是否超过最大重试次数
  if (attempts >= this.MAX_RETRIES) {
    this.retryAttempts.delete(work);
    return false;
  }

  // 3. 更新重试次数
  this.retryAttempts.set(work, attempts + 1);

  // 4. 计算退避时间
  const backoffMs = this.calculateBackoff(attempts, error);

  // 5. 发送重试消息
  this.sendMessage({
    type: MessageType.SYSTEM,
    content: `重试 ${attempts + 1}/${this.MAX_RETRIES}，${backoffMs / 1000}秒后重试...`
  });

  // 6. 延迟后重新加入队列
  await this.sleep(backoffMs);
  this.workLog.unshift(work);

  return true;
}

private calculateBackoff(attempts: number, error: Error): number {
  // 指数退避: 2^attempts * 1000ms
  // attempts=0: 1s, attempts=1: 2s, attempts=2: 4s
  let backoff = Math.pow(2, attempts) * 1000;

  // 如果是速率限制错误，使用更长的退避时间
  if (error instanceof RateLimitError) {
    backoff = 30000; // 30秒
  }

  // 最大不超过60秒
  return Math.min(backoff, 60000);
}
```

---

## 消息系统

### Message接口

```typescript
interface Message {
  type: MessageType;
  content: string;
  timestamp?: Date;
  metadata?: Record<string, any>;
}

enum MessageType {
  GOAL = "goal",           // 目标消息
  TASK = "task",           // 任务消息
  THINKING = "thinking",   // 思考消息
  ACTION = "action",       // 行动消息
  STATUS = "status",       // 状态消息
  ERROR = "error",         // 错误消息
  SYSTEM = "system"        // 系统消息
}
```

### 消息发送方法

```typescript
class AutonomousAgent {
  // 基础发送方法
  private sendMessage(message: Message): void {
    // 1. 添加时间戳
    message.timestamp = new Date();

    // 2. 保存到模型
    this.model.addMessage(message);

    // 3. 回调UI
    this.messageCallback(message);

    // 4. 记录日志
    console.log(`[${message.type}]`, message.content);
  }

  // 便捷方法
  sendGoalMessage(goal: string): void {
    this.sendMessage({
      type: MessageType.GOAL,
      content: `🎯 目标: ${goal}`
    });
  }

  sendTaskMessage(task: string): void {
    this.sendMessage({
      type: MessageType.TASK,
      content: `📋 任务: ${task}`
    });
  }

  sendThinkingMessage(thought: string): void {
    this.sendMessage({
      type: MessageType.THINKING,
      content: `💭 ${thought}`
    });
  }

  sendActionMessage(action: string): void {
    this.sendMessage({
      type: MessageType.ACTION,
      content: `🔧 ${action}`
    });
  }

  sendStatusMessage(status: string): void {
    this.sendMessage({
      type: MessageType.STATUS,
      content: status
    });
  }

  sendErrorMessage(error: Error): void {
    this.sendMessage({
      type: MessageType.ERROR,
      content: `❌ ${error.message}`
    });
  }

  sendSystemMessage(content: string): void {
    this.sendMessage({
      type: MessageType.SYSTEM,
      content: content
    });
  }
}
```

### 消息流程

```
Agent内部 ──┐
            │
            ▼
    ┌───────────────┐
    │ sendMessage() │
    └───────┬───────┘
            │
        ┌───┴────┬──────────────┐
        │        │              │
        ▼        ▼              ▼
   ┌───────┐  ┌──────┐    ┌──────────┐
   │ Model │  │ Log  │    │ Callback │
   └───────┘  └──────┘    └────┬─────┘
                                │
                                ▼
                          ┌──────────┐
                          │ React UI │
                          └──────────┘
```

---

## 完整实现

### 完整的AutonomousAgent类

```typescript
class AutonomousAgent {
  // ========== 属性 ==========
  private model: AgentModel;
  private workLog: AgentWork[];
  private api: AgentApi;
  private messageCallback: (message: Message) => void;
  private isRunning: boolean = false;
  private retryAttempts: Map<AgentWork, number> = new Map();
  private readonly MAX_RETRIES = 3;

  // ========== 构造函数 ==========
  constructor(
    goal: string,
    api: AgentApi,
    messageCallback: (message: Message) => void
  ) {
    this.model = new AgentModel(goal);
    this.api = api;
    this.messageCallback = messageCallback;

    // 初始化工作队列
    this.workLog = [new StartGoalWork(this, goal)];
  }

  // ========== 生命周期控制 ==========
  async start(): Promise<void> {
    if (this.isRunning) {
      throw new Error("Agent is already running");
    }

    this.isRunning = true;
    await this.run();
  }

  pause(): void {
    if (this.model.getLifecycle() === AgentLifecycle.RUNNING) {
      this.model.setLifecycle(AgentLifecycle.PAUSING);
      this.sendSystemMessage("正在暂停...");
    }
  }

  resume(): void {
    if (this.model.getLifecycle() === AgentLifecycle.PAUSED) {
      this.model.setLifecycle(AgentLifecycle.RUNNING);
      this.sendSystemMessage("继续执行...");
      this.run();
    }
  }

  stop(): void {
    this.model.setLifecycle(AgentLifecycle.STOPPING);
    this.sendSystemMessage("正在停止...");
    this.stopAgent();
  }

  // ========== 主运行循环 ==========
  async run(): Promise<void> {
    this.model.setLifecycle(AgentLifecycle.RUNNING);
    this.sendSystemMessage("🚀 Agent开始运行");

    try {
      while (this.workLog[0]) {
        // 检查状态
        if (this.model.getLifecycle() === AgentLifecycle.PAUSING) {
          this.model.setLifecycle(AgentLifecycle.PAUSED);
          this.sendSystemMessage("⏸️ Agent已暂停");
          return;
        }

        if (this.model.getLifecycle() !== AgentLifecycle.RUNNING) {
          return;
        }

        if (!this.isRunning) {
          return;
        }

        // 执行工作
        const work = this.workLog[0];
        this.sendStatusMessage(work.getStatusMessage());

        try {
          await this.runWork(work);
          this.workLog.shift();

          const next = work.next();
          if (next) {
            this.workLog.push(next);
          }

        } catch (error) {
          await this.handleWorkError(work, error as Error);
          if (!this.isRunning) {
            return;
          }
        }

        this.addTasksIfWorklogEmpty();
        await this.sleep(100);
      }

      this.sendSystemMessage("✅ 所有任务已完成");
      this.stopAgent();

    } catch (error) {
      console.error("Agent run failed:", error);
      this.sendErrorMessage(error as Error);
      this.stopAgent();
    }
  }

  // ========== 工作执行 ==========
  async runWork(work: AgentWork): Promise<void> {
    await work.run();
  }

  // ========== 错误处理 ==========
  private async handleWorkError(work: AgentWork, error: Error): Promise<void> {
    console.error(`Work ${work.getType()} failed:`, error);
    this.sendErrorMessage(error);

    if (error instanceof NetworkError) {
      if (await this.retryWork(work, error)) {
        this.sendSystemMessage("网络错误，正在重试...");
        return;
      }
    } else if (error instanceof RateLimitError) {
      this.sendSystemMessage("API速率限制，30秒后重试...");
      this.pause();
      setTimeout(() => {
        this.workLog.unshift(work);
        this.resume();
      }, 30000);
      return;
    }

    this.stopAgent();
  }

  private async retryWork(work: AgentWork, error: Error): Promise<boolean> {
    const attempts = this.retryAttempts.get(work) || 0;

    if (attempts >= this.MAX_RETRIES) {
      this.retryAttempts.delete(work);
      return false;
    }

    this.retryAttempts.set(work, attempts + 1);
    const backoffMs = this.calculateBackoff(attempts, error);

    this.sendSystemMessage(
      `重试 ${attempts + 1}/${this.MAX_RETRIES}，${backoffMs / 1000}秒后重试...`
    );

    await this.sleep(backoffMs);
    this.workLog.unshift(work);

    return true;
  }

  private calculateBackoff(attempts: number, error: Error): number {
    let backoff = Math.pow(2, attempts) * 1000;
    if (error instanceof RateLimitError) {
      backoff = 30000;
    }
    return Math.min(backoff, 60000);
  }

  // ========== 队列管理 ==========
  private addTasksIfWorklogEmpty(): void {
    if (this.workLog.length > 0) {
      return;
    }

    const nextTask = this.model.getNextTask();
    if (nextTask) {
      const work = new AnalyzeTaskWork(this, nextTask);
      this.workLog.push(work);
      this.sendSystemMessage(`开始新任务: ${nextTask}`);
    }
  }

  private stopAgent(): void {
    this.isRunning = false;
    this.model.setLifecycle(AgentLifecycle.STOPPED);
    this.sendSystemMessage("✅ Agent已停止");
  }

  // ========== 消息系统 ==========
  private sendMessage(message: Message): void {
    message.timestamp = new Date();
    this.model.addMessage(message);
    this.messageCallback(message);
  }

  sendGoalMessage(goal: string): void {
    this.sendMessage({ type: MessageType.GOAL, content: `🎯 目标: ${goal}` });
  }

  sendTaskMessage(task: string): void {
    this.sendMessage({ type: MessageType.TASK, content: `📋 任务: ${task}` });
  }

  sendThinkingMessage(thought: string): void {
    this.sendMessage({ type: MessageType.THINKING, content: `💭 ${thought}` });
  }

  sendActionMessage(action: string): void {
    this.sendMessage({ type: MessageType.ACTION, content: `🔧 ${action}` });
  }

  sendStatusMessage(status: string): void {
    this.sendMessage({ type: MessageType.STATUS, content: status });
  }

  sendErrorMessage(error: Error): void {
    this.sendMessage({ type: MessageType.ERROR, content: `❌ ${error.message}` });
  }

  sendSystemMessage(content: string): void {
    this.sendMessage({ type: MessageType.SYSTEM, content });
  }

  // ========== 工具方法 ==========
  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // ========== 访问器 ==========
  getModel(): AgentModel {
    return this.model;
  }

  getApi(): AgentApi {
    return this.api;
  }
}
```

---

## 总结

### 设计亮点

1. **清晰的职责分离**: Agent负责流程控制，Work负责具体任务
2. **完善的生命周期管理**: 支持启动、暂停、恢复、停止
3. **健壮的错误处理**: 分类处理不同错误，智能重试
4. **灵活的消息系统**: 统一的消息格式，易于扩展
5. **可维护的代码结构**: 清晰的方法命名和注释

### 关键方法

- `run()`: 主运行循环，控制整个执行流程
- `runWork()`: 执行单个Work
- `handleWorkError()`: 错误处理和恢复
- `sendMessage()`: 消息发送和UI更新

### 使用示例

```typescript
// 创建Agent
const agent = new AutonomousAgent(
  "创建一个Todo应用",
  new AgentApiImpl(),
  (message) => {
    // 更新UI
    console.log(message);
  }
);

// 启动
await agent.start();

// 暂停
agent.pause();

// 恢复
agent.resume();

// 停止
agent.stop();
```

---

**相关文档**:
- [前端架构总览](./agentgpt-frontend.md)
- [Work模式详解](./work-pattern.md)
