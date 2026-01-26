# AgentGPT架构深度分析

## 项目概述

AgentGPT是一个自主AI Agent平台，允许用户在浏览器中配置和部署自主AI Agent。Agent会通过思考任务、执行任务并从结果中学习来尝试达成目标。

## 技术栈

### 前端
- **框架**: Next.js 13 + TypeScript
- **状态管理**: Zustand
- **样式**: TailwindCSS + HeadlessUI
- **认证**: Next-Auth.js
- **ORM**: Prisma
- **数据库**: Planetscale (MySQL)

### 后端
- **框架**: FastAPI (Python)
- **ORM**: SQLModel
- **LLM工具**: LangChain
- **数据库**: MySQL

## 核心架构

### 1. 前端架构

#### 1.1 AutonomousAgent类

`AutonomousAgent`是前端Agent执行的核心类，负责管理Agent的整个生命周期。

**核心属性**：
```typescript
class AutonomousAgent {
  model: AgentRunModel;              // Agent运行模型
  modelSettings: ModelSettings;       // 模型设置
  session?: Session;                 // 用户会话
  messageService: MessageService;    // 消息服务
  api: AgentApi;                     // API接口
  private readonly workLog: AgentWork[]; // 工作日志
}
```

**核心方法**：
- `run()`: 执行Agent主循环
- `stopAgent()`: 停止Agent
- `pauseAgent()`: 暂停Agent
- `runWork()`: 执行单个工作项

**执行流程**：
1. 初始化工作日志，添加`StartGoalWork`
2. 循环执行工作日志中的任务
3. 每个任务执行完成后，添加下一个任务
4. 当工作日志为空时，停止Agent

#### 1.2 AgentWork模式

AgentGPT使用Work模式来组织Agent的任务执行。每个Work代表一个工作单元。

**Work接口**：
```typescript
interface AgentWork {
  run(): Promise<void>;        // 执行工作
  conclude(): Promise<void>;   // 完成工作
  next(): AgentWork | null;    // 下一个工作
  onError?(error: Error): boolean; // 错误处理
}
```

**Work类型**：
1. **StartGoalWork**: 启动目标，生成初始任务列表
2. **AnalyzeTaskWork**: 分析任务，选择工具
3. **ExecuteTaskWork**: 执行任务，调用工具
4. **CreateTaskWork**: 创建新任务
5. **SummarizeWork**: 总结结果

**Work流程**：
```
StartGoalWork -> AnalyzeTaskWork -> ExecuteTaskWork -> CreateTaskWork -> ...
                                                           |
                                                           v
                                                    SummarizeWork
```

#### 1.3 AgentApi

`AgentApi`负责与后端API通信。

**主要方法**：
- `startGoal(goal: string)`: 启动目标
- `analyzeTask(goal: string, task: string)`: 分析任务
- `executeTask(...)`: 执行任务
- `createTasks(...)`: 创建任务
- `summarize(...)`: 总结结果

### 2. 后端架构

#### 2.1 AgentService接口

`AgentService`定义了Agent服务的核心接口。

**接口定义**：
```python
class AgentService(Protocol):
    async def start_goal_agent(self, *, goal: str) -> List[str]
    async def analyze_task_agent(self, *, goal: str, task: str, tool_names: List[str]) -> Analysis
    async def execute_task_agent(self, *, goal: str, task: str, analysis: Analysis) -> StreamingResponse
    async def create_tasks_agent(self, *, goal: str, tasks: List[str], last_task: str, result: str) -> List[str]
    async def summarize_task_agent(self, *, goal: str, results: List[str]) -> StreamingResponse
    async def chat(self, *, message: str, results: List[str]) -> StreamingResponse
```

#### 2.2 Prompt模板系统

AgentGPT使用LangChain的`PromptTemplate`来管理Prompt模板。

**核心Prompt**：
1. **start_goal_prompt**: 启动目标，生成初始任务列表
2. **analyze_task_prompt**: 分析任务，选择工具
3. **execute_task_prompt**: 执行任务
4. **create_tasks_prompt**: 创建新任务
5. **summarize_prompt**: 总结结果
6. **code_prompt**: 代码生成专用Prompt

**Prompt特点**：
- 支持多语言
- 使用变量插值
- 包含示例和指导

#### 2.3 工具系统

AgentGPT提供了多种工具供Agent使用。

**工具类型**：
1. **Search**: Google搜索工具
2. **Code**: 代码生成工具
3. **Image**: 图像生成工具
4. **Wikipedia**: Wikipedia搜索工具
5. **SID**: 私有信息搜索工具

**工具接口**：
```python
class Tool:
    description: str           # 工具描述
    public_description: str    # 公开描述
    arg_description: str       # 参数描述
    
    async def call(self, goal: str, task: str, input_str: str) -> str
```

## 执行流程

### 完整流程

```
用户输入目标
    |
    v
StartGoalWork (前端)
    |
    v
start_goal_agent (后端API)
    |
    v
start_goal_prompt -> LLM -> 初始任务列表
    |
    v
AnalyzeTaskWork (前端)
    |
    v
analyze_task_agent (后端API)
    |
    v
analyze_task_prompt -> LLM -> 工具选择
    |
    v
ExecuteTaskWork (前端)
    |
    v
execute_task_agent (后端API)
    |
    v
调用选定工具 -> 执行任务 -> 返回结果
    |
    v
CreateTaskWork (前端)
    |
    v
create_tasks_agent (后端API)
    |
    v
create_tasks_prompt -> LLM -> 新任务
    |
    v
循环执行...
    |
    v
SummarizeWork (前端)
    |
    v
summarize_task_agent (后端API)
    |
    v
summarize_prompt -> LLM -> 最终总结
```

### 关键设计模式

#### 1. Work模式
- **优点**: 职责清晰，易于扩展
- **实现**: 每个Work负责一个阶段的任务

#### 2. 流式响应
- **优点**: 实时反馈，用户体验好
- **实现**: 使用Server-Sent Events (SSE)

#### 3. 错误处理和重试
- **优点**: 提高系统可靠性
- **实现**: 前端实现重试机制，后端返回错误信息

## 学习要点

### 1. Agent设计模式
- Work模式的组织方式
- 任务分解和执行循环
- 状态管理和生命周期

### 2. Prompt Engineering
- Prompt模板的设计
- 多语言支持
- 示例和指导的使用

### 3. 工具集成
- 工具接口设计
- 工具调用机制
- 工具结果处理

### 4. 前后端协作
- API设计
- 流式响应处理
- 错误处理机制

## 可改进点

1. **多Agent协作**: 当前是单Agent系统，可以扩展为多Agent协作
2. **工具扩展**: 可以添加更多自定义工具
3. **性能优化**: 可以优化Token使用和响应速度
4. **可观测性**: 可以添加更完善的监控和日志系统

## 参考资源

- [AgentGPT GitHub](https://github.com/reworkd/AgentGPT)
- [LangChain文档](https://python.langchain.com/)
- [ReAct论文](https://arxiv.org/abs/2210.03629)
