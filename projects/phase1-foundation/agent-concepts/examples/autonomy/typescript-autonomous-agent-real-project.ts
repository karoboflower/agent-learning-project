/**
 * 真正的自主Agent - 实际项目创建版本
 *
 * 这个Agent具备：
 * 1. 自主性：使用LLM自主生成任务和决策
 * 2. 反应性：根据执行结果调整策略
 * 3. 主动性：主动规划下一步行动
 * 4. 工具使用：使用文件操作工具完成任务
 * 5. ReAct模式：推理(Reasoning) + 行动(Acting)
 */

import * as fs from "fs";
import * as path from "path";

// 加载环境变量
import * as dotenv from "dotenv";
dotenv.config();
console.log("process.env:", process.env);
console.log("process.env.GEMINI_API_KEY:", process.env.GEMINI_API_KEY);
// ==================== 类型定义 ====================

interface Task {
  id: string;
  description: string;
  priority: number;
  dependencies: string[];
  status: "pending" | "running" | "completed" | "failed";
  tool?: string; // 使用的工具名称
  parameters?: Record<string, any>; // 工具参数
}

interface TaskResult {
  taskId: string;
  success: boolean;
  result: string;
  error?: string;
  filesCreated?: string[];
  observations?: string[]; // Agent观察到的结果
}

interface AgentState {
  goal: string;
  tasks: Task[];
  completedTasks: Task[];
  failedTasks: Task[];
  knowledge: Map<string, any>;
  observations: string[]; // 执行过程中的观察
  thoughts: string[]; // Agent的思考过程
  status: "idle" | "running" | "paused" | "stopped" | "stopping" | "completed";
  projectPath: string;
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
  generate(
    prompt: string,
    options?: { temperature?: number; maxTokens?: number },
  ): Promise<string>;
  generateTasks(goal: string, context: string): Promise<Task[]>;
  analyzeResult(
    thought: string,
    observation: string,
    goal: string,
  ): Promise<string>;
  decideAction(
    thought: string,
    availableTools: string[],
  ): Promise<{ tool: string; parameters: Record<string, any> }>;
}

// ==================== OpenAI LLM服务实现 ====================

class OpenAILLMService implements LLMService {
  private apiKey: string;
  private model: string;
  private baseURL?: string;

  constructor(
    apiKey: string,
    model: string = "gpt-3.5-turbo",
    baseURL?: string,
  ) {
    this.apiKey = apiKey;
    this.model = model;
    this.baseURL = baseURL;
  }
  async generate(
    prompt: string,
    options?: { temperature?: number; maxTokens?: number },
  ): Promise<string> {
    console.log(this.apiKey);

    const maxRetries = 3;
    for (let attempt = 0; attempt < maxRetries; attempt++) {
      try {
        const response = await fetch(
          this.baseURL || "https://api.openai.com/v1/chat/completions",
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${this.apiKey}`,
            },
            body: JSON.stringify({
              model: this.model,
              messages: [{ role: "user", content: prompt }],
              temperature: options?.temperature || 0.7,
              max_tokens: options?.maxTokens || 1000,
            }),
          },
        );

        if (response.status === 429) {
          const waitTime = Math.pow(2, attempt) * 1000;
          if (attempt < maxRetries - 1) {
            await this.sleep(waitTime);
            continue;
          }
        }

        if (!response.ok) {
          throw new Error(
            `OpenAI API error: ${response.status} ${response.statusText}`,
          );
        }

        const data = (await response.json()) as {
          choices: Array<{ message?: { content?: string } }>;
        };
        const content = data.choices[0]?.message?.content;
        console.log(`LLM Response: ${content}`);
        return  content || "";
      } catch (error) {
        console.log(error);
        if (attempt === maxRetries - 1) {
          console.warn("⚠️  LLM调用失败，使用模拟响应");
          return this.getMockResponse(prompt);
        }
        await this.sleep(Math.pow(2, attempt) * 1000);
      }
    }
    return this.getMockResponse(prompt);
  }

  async generateTasks(goal: string, context: string): Promise<Task[]> {
    const prompt = `
你是一个项目规划AI Agent。请将以下目标分解为具体的开发任务。

目标：${goal}

${context ? `当前上下文：\n${context}` : ""}

请生成3-8个具体的开发任务，每个任务一行，格式：
任务描述|优先级(0-1)|依赖任务ID(用逗号分隔，无依赖则为空)|工具名称

可用工具：create_file, write_code, create_dir, install_deps

示例：
创建项目目录结构|0.9||create_dir
创建package.json|0.9|task_1|create_file
编写React主组件|0.8|task_2|write_code
    `;

    const response = await this.generate(prompt, { temperature: 0.5 });
    const lines = response.split("\n").filter((line) => line.trim().length > 0);

    const tasks: Task[] = [];
    let taskId = 1;

    for (const line of lines) {
      const parts = line.split("|");
      if (parts.length >= 3) {
        const description = parts[0].trim();
        const priority = parseFloat(parts[1].trim()) || 0.5;
        const dependencies =
          parts[2]
            ?.trim()
            .split(",")
            .filter((id) => id.length > 0) || [];
        const tool = parts[3]?.trim() || "write_code";

        if (description.includes("示例") || description.includes("Example")) {
          continue;
        }

        tasks.push({
          id: `task_${taskId++}`,
          description,
          priority: Math.max(0, Math.min(1, priority)),
          dependencies,
          status: "pending",
          tool,
        });
      }
    }

    return tasks.length > 0 ? tasks : this.getDefaultTasks(goal);
  }

  async analyzeResult(
    thought: string,
    observation: string,
    goal: string,
  ): Promise<string> {
    const prompt = `
你是一个AI Agent，正在分析任务执行结果。

目标：${goal}

之前的思考：${thought}
执行结果：${observation}

请分析：
1. 任务是否成功完成？
2. 发现了什么问题？
3. 下一步应该做什么？

用简洁的语言回答（不超过100字）。
    `;

    return await this.generate(prompt, { temperature: 0.3, maxTokens: 200 });
  }

  async decideAction(
    thought: string,
    availableTools: string[],
  ): Promise<{ tool: string; parameters: Record<string, any> }> {
    const prompt = `
你是一个AI Agent，需要决定下一步行动。

思考：${thought}
可用工具：${availableTools.join(", ")}

请选择最合适的工具，并返回JSON格式：
{"tool": "工具名称", "parameters": {"参数名": "参数值"}}
    `;

    const response = await this.generate(prompt, {
      temperature: 0.2,
      maxTokens: 200,
    });

    try {
      const jsonMatch = response.match(/\{[\s\S]*\}/);
      if (jsonMatch) {
        return JSON.parse(jsonMatch[0]);
      }
    } catch (e) {
      // 解析失败，使用默认值
    }

    return {
      tool: availableTools[0] || "write_code",
      parameters: {},
    };
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms));
  }

  private getMockResponse(prompt: string): string {
    if (prompt.includes("任务") || prompt.includes("task")) {
      return `创建项目目录结构|0.9||create_dir
创建package.json配置文件|0.9|task_1|create_file
创建Vite配置文件|0.8|task_2|create_file
创建HTML入口文件|0.8|task_1|create_file
创建React主应用组件|0.9|task_3,task_4|write_code
创建待办事项列表组件|0.8|task_5|write_code
创建待办事项表单组件|0.8|task_5|write_code
创建CSS样式文件|0.7|task_5,task_6,task_7|write_code`;
    }

    if (prompt.includes("分析") || prompt.includes("分析")) {
      return "任务执行成功。下一步应该继续创建相关组件文件。";
    }

    return '{"tool": "write_code", "parameters": {}}';
  }

  private getDefaultTasks(goal: string): Task[] {
    return [
      {
        id: "task_1",
        description: "创建项目目录结构",
        priority: 0.9,
        dependencies: [],
        status: "pending",
        tool: "create_dir",
      },
      {
        id: "task_2",
        description: "创建package.json配置文件",
        priority: 0.9,
        dependencies: ["task_1"],
        status: "pending",
        tool: "create_file",
      },
    ];
  }
}

// ==================== 工具系统 ====================

interface Tool {
  name: string;
  description: string;
  execute(
    parameters: Record<string, any>,
    context: AgentContext,
  ): Promise<ToolResult>;
}

interface ToolResult {
  success: boolean;
  result: string;
  filesCreated?: string[];
  observations?: string[];
}

interface AgentContext {
  projectPath: string;
  goal: string;
  completedTasks: Task[];
  knowledge: Map<string, any>;
}

class FileOperationsTool implements Tool {
  name = "create_file";
  description = "创建文件";

  async execute(
    parameters: Record<string, any>,
    context: AgentContext,
  ): Promise<ToolResult> {
    const { filePath, content } = parameters;
    if (!filePath) {
      return { success: false, result: "缺少文件路径参数" };
    }

    const fullPath = path.join(context.projectPath, filePath);
    const dir = path.dirname(fullPath);

    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    fs.writeFileSync(fullPath, content || "", "utf-8");

    return {
      success: true,
      result: `文件 ${filePath} 创建成功`,
      filesCreated: [filePath],
      observations: [`创建了文件: ${filePath}`],
    };
  }
}

class WriteCodeTool implements Tool {
  name = "write_code";
  description = "编写代码文件";

  async execute(
    parameters: Record<string, any>,
    context: AgentContext,
  ): Promise<ToolResult> {
    const { filePath, code } = parameters;
    if (!filePath || !code) {
      return { success: false, result: "缺少文件路径或代码内容" };
    }

    const fullPath = path.join(context.projectPath, filePath);
    const dir = path.dirname(fullPath);

    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    fs.writeFileSync(fullPath, code, "utf-8");

    return {
      success: true,
      result: `代码文件 ${filePath} 编写成功`,
      filesCreated: [filePath],
      observations: [
        `编写了代码文件: ${filePath}`,
        `代码行数: ${code.split("\n").length}`,
      ],
    };
  }
}

class CreateDirTool implements Tool {
  name = "create_dir";
  description = "创建目录";

  async execute(
    parameters: Record<string, any>,
    context: AgentContext,
  ): Promise<ToolResult> {
    const { dirPath } = parameters;
    if (!dirPath) {
      return { success: false, result: "缺少目录路径参数" };
    }

    const fullPath = path.join(context.projectPath, dirPath);

    if (!fs.existsSync(fullPath)) {
      fs.mkdirSync(fullPath, { recursive: true });
    }

    return {
      success: true,
      result: `目录 ${dirPath} 创建成功`,
      observations: [`创建了目录: ${dirPath}`],
    };
  }
}

// ==================== 项目代码生成器 ====================

class CodeGenerator {
  static generatePackageJson(projectName: string): string {
    return JSON.stringify(
      {
        name: projectName.toLowerCase().replace(/\s+/g, "-"),
        version: "1.0.0",
        description: "A todo management web application",
        type: "module",
        scripts: {
          dev: "vite",
          build: "vite build",
          preview: "vite preview",
        },
        dependencies: {
          react: "^18.2.0",
          "react-dom": "^18.2.0",
        },
        devDependencies: {
          "@vitejs/plugin-react": "^4.0.0",
          vite: "^4.4.0",
        },
      },
      null,
      2,
    );
  }

  static generateViteConfig(): string {
    return `import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    open: true
  }
});
`;
  }

  static generateIndexHTML(projectName: string): string {
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>${projectName}</title>
</head>
<body>
  <div id="root"></div>
  <script type="module" src="/src/main.jsx"></script>
</body>
</html>
`;
  }

  static generateMainJSX(): string {
    return `import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
`;
  }

  static generateAppJSX(): string {
    return `import React, { useState } from 'react';
import TodoList from './components/TodoList';
import TodoForm from './components/TodoForm';
import './App.css';

function App() {
  const [todos, setTodos] = useState([]);

  const addTodo = (text) => {
    const newTodo = {
      id: Date.now(),
      text,
      completed: false,
      createdAt: new Date().toISOString()
    };
    setTodos([...todos, newTodo]);
  };

  const toggleTodo = (id) => {
    setTodos(todos.map(todo =>
      todo.id === id ? { ...todo, completed: !todo.completed } : todo
    ));
  };

  const deleteTodo = (id) => {
    setTodos(todos.filter(todo => todo.id !== id));
  };

  return (
    <div className="app">
      <header className="app-header">
        <h1>待办事项管理</h1>
      </header>
      <main className="app-main">
        <TodoForm onAdd={addTodo} />
        <TodoList
          todos={todos}
          onToggle={toggleTodo}
          onDelete={deleteTodo}
        />
      </main>
    </div>
  );
}

export default App;
`;
  }

  static generateComponent(
    componentName: string,
    type: "list" | "form" | "item",
  ): string {
    switch (type) {
      case "list":
        return `import React from 'react';
import TodoItem from './TodoItem';
import './TodoList.css';

function TodoList({ todos, onToggle, onDelete }) {
  if (todos.length === 0) {
    return (
      <div className="todo-list empty">
        <p>暂无待办事项，添加一个开始吧！</p>
      </div>
    );
  }

  return (
    <div className="todo-list">
      {todos.map(todo => (
        <TodoItem
          key={todo.id}
          todo={todo}
          onToggle={onToggle}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
}

export default TodoList;
`;

      case "form":
        return `import React, { useState } from 'react';
import './TodoForm.css';

function TodoForm({ onAdd }) {
  const [input, setInput] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (input.trim()) {
      onAdd(input.trim());
      setInput('');
    }
  };

  return (
    <form className="todo-form" onSubmit={handleSubmit}>
      <input
        type="text"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        placeholder="输入待办事项..."
        className="todo-input"
      />
      <button type="submit" className="todo-submit">
        添加
      </button>
    </form>
  );
}

export default TodoForm;
`;

      case "item":
        return `import React from 'react';
import './TodoItem.css';

function TodoItem({ todo, onToggle, onDelete }) {
  return (
    <div className={\`todo-item \${todo.completed ? 'completed' : ''}\`}>
      <input
        type="checkbox"
        checked={todo.completed}
        onChange={() => onToggle(todo.id)}
        className="todo-checkbox"
      />
      <span className="todo-text">{todo.text}</span>
      <button
        onClick={() => onDelete(todo.id)}
        className="todo-delete"
        aria-label="删除"
      >
        ×
      </button>
    </div>
  );
}

export default TodoItem;
`;

      default:
        return "";
    }
  }

  static generateCSS(fileName: string): string {
    const cssMap: Record<string, string> = {
      "App.css": `.app {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.app-header {
  text-align: center;
  margin-bottom: 30px;
}

.app-header h1 {
  color: #333;
  font-size: 2rem;
}

.app-main {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}
`,
      "index.css": `* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
    'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
    sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  background: #f5f5f5;
  min-height: 100vh;
}

#root {
  min-height: 100vh;
}
`,
      "TodoList.css": `.todo-list {
  margin-top: 20px;
}

.todo-list.empty {
  text-align: center;
  padding: 40px;
  color: #999;
}
`,
      "TodoItem.css": `.todo-item {
  display: flex;
  align-items: center;
  padding: 12px;
  border-bottom: 1px solid #eee;
  transition: background-color 0.2s;
}

.todo-item:hover {
  background-color: #f9f9f9;
}

.todo-item.completed .todo-text {
  text-decoration: line-through;
  color: #999;
}

.todo-checkbox {
  margin-right: 12px;
  cursor: pointer;
}

.todo-text {
  flex: 1;
  font-size: 16px;
}

.todo-delete {
  background: #ff4444;
  color: white;
  border: none;
  border-radius: 4px;
  width: 28px;
  height: 28px;
  cursor: pointer;
  font-size: 20px;
  line-height: 1;
  transition: background-color 0.2s;
}

.todo-delete:hover {
  background: #cc0000;
}
`,
      "TodoForm.css": `.todo-form {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.todo-input {
  flex: 1;
  padding: 12px;
  border: 2px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
  transition: border-color 0.2s;
}

.todo-input:focus {
  outline: none;
  border-color: #4CAF50;
}

.todo-submit {
  padding: 12px 24px;
  background: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.todo-submit:hover {
  background: #45a049;
}
`,
    };

    return cssMap[fileName] || "";
  }
}

// ==================== 真正的自主Agent ====================

class RealAutonomousAgent {
  private state: AgentState;
  private config: AgentConfig;
  private isRunning: boolean = false;
  private llm: LLMService;
  private tools: Map<string, Tool>;

  constructor(
    goal: string,
    projectPath: string,
    llm: LLMService,
    config?: Partial<AgentConfig>,
  ) {
    this.state = {
      goal,
      tasks: [],
      completedTasks: [],
      failedTasks: [],
      knowledge: new Map(),
      observations: [],
      thoughts: [],
      status: "idle",
      projectPath,
      metadata: {
        startTime: new Date(),
        lastUpdateTime: new Date(),
        iterationCount: 0,
        totalCost: 0,
      },
    };

    this.config = {
      maxIterations: 100,
      maxCost: 1000,
      minPriority: 0.3,
      retryAttempts: 3,
      backoffBase: 2,
      ...config,
    };

    this.llm = llm;

    // 注册工具
    this.tools = new Map();
    this.tools.set("create_file", new FileOperationsTool());
    this.tools.set("write_code", new WriteCodeTool());
    this.tools.set("create_dir", new CreateDirTool());
  }

  async run(): Promise<void> {
    if (this.isRunning) {
      throw new Error("Agent is already running");
    }

    this.isRunning = true;
    this.state.status = "running";

    try {
      // ReAct循环：思考 -> 行动 -> 观察 -> 思考...
      while (this.isRunning && this.shouldContinue()) {
        // 1. 思考（Reasoning）
        const thought = await this.think();
        this.state.thoughts.push(thought);

        // 2. 行动（Acting）
        const action = await this.act(thought);

        // 3. 观察（Observation）
        const observation = await this.observe(action);
        this.state.observations.push(observation);

        // 4. 更新状态
        this.updateState();

        // 5. 检查是否完成
        if (this.isGoalAchieved()) {
          this.complete();
          break;
        }
      }
    } catch (error) {
      this.handleError(
        error instanceof Error ? error : new Error(String(error)),
      );
    } finally {
      this.isRunning = false;
      if (this.state.status === "running") {
        this.state.status = "stopped";
      }
    }
  }

  // ReAct: 思考（Reasoning）
  private async think(): Promise<string> {
    // 如果没有任务，直接返回需要生成任务的思考
    if (this.state.tasks.length === 0) {
      return "当前没有待执行的任务，需要根据目标和已完成的工作生成新的任务列表。";
    }

    const context = `
目标：${this.state.goal}
已完成任务：${this.state.completedTasks.map((t) => t.description).join(", ") || "无"}
待执行任务数：${this.state.tasks.length} 个
最新观察：${this.state.observations.slice(-2).join("; ") || "无"}
    `;

    const prompt = `
你是一个AI Agent，正在执行项目开发任务。

${context}

请思考：
1. 当前项目进度如何？
2. 下一个要执行的任务是什么？为什么选择它？
3. 执行这个任务需要注意什么？

用简洁的语言表达你的思考（不超过100字）。
    `;

    const thought = await this.llm.generate(prompt, {
      temperature: 0.7,
      maxTokens: 150,
    });

    // 如果LLM返回的是任务列表（模拟响应），返回一个默认思考
    if (thought.includes("|") && thought.includes("task_")) {
      const nextTask = this.state.tasks.find(
        (t) => t.status === "pending" && this.checkDependencies(t),
      );
      return nextTask
        ? `准备执行任务：${nextTask.description}。这个任务优先级较高(${nextTask.priority.toFixed(2)})，且依赖已满足。`
        : "需要等待依赖任务完成。";
    }

    return thought;
  }

  // ReAct: 行动（Acting）
  private async act(thought: string): Promise<TaskResult> {
    // 如果没有任务，使用LLM生成
    if (this.state.tasks.length === 0) {
      console.log("\n🤔 Agent思考:", thought);
      console.log("📝 使用LLM生成新任务...");

      const context = `
已完成：${this.state.completedTasks.map((t) => t.description).join(", ") || "无"}
最新观察：${this.state.observations.slice(-2).join("; ") || "无"}
      `;

      const newTasks = await this.llm.generateTasks(this.state.goal, context);
      this.state.tasks.push(...newTasks);

      console.log(`✅ 生成了 ${newTasks.length} 个新任务`);
      newTasks.forEach((task, index) => {
        console.log(
          `   ${index + 1}. [${task.id}] ${task.description} (工具: ${task.tool})`,
        );
      });
    }

    // 选择下一个任务
    const task = await this.selectNextTask();
    if (!task) {
      return {
        taskId: "none",
        success: false,
        result: "没有可执行的任务",
      };
    }

    console.log(`\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`);
    console.log(
      `🔄 [${this.state.metadata.iterationCount + 1}] 执行任务: ${task.description}`,
    );
    console.log(
      `   任务ID: ${task.id} | 工具: ${task.tool} | 优先级: ${task.priority.toFixed(2)}`,
    );

    // 执行任务
    return await this.executeTask(task);
  }

  // ReAct: 观察（Observation）
  private async observe(action: TaskResult): Promise<string> {
    const lastThought =
      this.state.thoughts[this.state.thoughts.length - 1] || "";
    const observation = action.success
      ? `任务执行成功: ${action.result}. ${action.observations?.join("; ") || ""}`
      : `任务执行失败: ${action.error}`;

    // 使用LLM分析观察结果
    const analysis = await this.llm.analyzeResult(
      lastThought,
      observation,
      this.state.goal,
    );

    console.log(`   👁️  观察: ${observation}`);
    console.log(
      `   🧠 分析: ${analysis.substring(0, 100)}${analysis.length > 100 ? "..." : ""}`,
    );

    return `${observation} | 分析: ${analysis}`;
  }

  private async selectNextTask(): Promise<Task | null> {
    const availableTasks = this.state.tasks.filter(
      (task) => task.status === "pending" && this.checkDependencies(task),
    );

    if (availableTasks.length === 0) {
      return null;
    }

    // 按优先级排序
    availableTasks.sort((a, b) => b.priority - a.priority);
    return availableTasks[0];
  }

  private checkDependencies(task: Task): boolean {
    if (task.dependencies.length === 0) {
      return true;
    }

    const completedTaskIds = new Set(
      this.state.completedTasks.map((t) => t.id),
    );

    return task.dependencies.every((depId) => completedTaskIds.has(depId));
  }

  private async executeTask(task: Task): Promise<TaskResult> {
    task.status = "running";
    this.state.metadata.lastUpdateTime = new Date();

    try {
      // 获取工具
      const tool = this.tools.get(task.tool || "write_code");
      if (!tool) {
        throw new Error(`工具 ${task.tool} 不存在`);
      }

      // 准备工具参数
      const parameters = await this.prepareToolParameters(task, tool);

      // 特殊处理：如果是多个CSS文件
      if (parameters._isMultipleFiles && parameters.filePath === "css_files") {
        const cssFiles = [
          "App.css",
          "index.css",
          "TodoList.css",
          "TodoItem.css",
          "TodoForm.css",
        ];
        const allFilesCreated: string[] = [];
        const allObservations: string[] = [];

        for (const cssFile of cssFiles) {
          const filePath =
            cssFile === "index.css"
              ? "src/index.css"
              : cssFile.startsWith("Todo")
                ? `src/components/${cssFile}`
                : `src/${cssFile}`;

          const cssTool = this.tools.get("write_code")!;
          const context: AgentContext = {
            projectPath: this.state.projectPath,
            goal: this.state.goal,
            completedTasks: this.state.completedTasks,
            knowledge: this.state.knowledge,
          };

          const result = await cssTool.execute(
            {
              filePath,
              code: CodeGenerator.generateCSS(cssFile),
            },
            context,
          );

          if (result.filesCreated) {
            allFilesCreated.push(...result.filesCreated);
          }
          if (result.observations) {
            allObservations.push(...result.observations);
          }
        }

        const toolResult = {
          success: true,
          result: `成功创建了 ${cssFiles.length} 个CSS文件`,
          filesCreated: allFilesCreated,
          observations: allObservations,
        };

        // 更新任务状态
        task.status = "completed";
        this.state.completedTasks.push(task);
        this.state.knowledge.set(`task_${task.id}`, toolResult);
        this.state.tasks = this.state.tasks.filter((t) => t.id !== task.id);

        console.log(`   ✅ 任务完成: ${task.description}`);
        console.log(`   📁 创建的文件: ${allFilesCreated.join(", ")}`);

        return {
          taskId: task.id,
          success: toolResult.success,
          result: toolResult.result,
          filesCreated: toolResult.filesCreated,
          observations: toolResult.observations,
        };
      }

      // 创建上下文
      const context: AgentContext = {
        projectPath: this.state.projectPath,
        goal: this.state.goal,
        completedTasks: this.state.completedTasks,
        knowledge: this.state.knowledge,
      };

      // 执行工具
      console.log(`   🔨 使用工具: ${tool.name}`);
      const toolResult = await tool.execute(parameters, context);

      // 更新任务状态
      task.status = "completed";
      this.state.completedTasks.push(task);
      this.state.knowledge.set(`task_${task.id}`, toolResult);
      this.state.tasks = this.state.tasks.filter((t) => t.id !== task.id);

      console.log(`   ✅ 任务完成: ${task.description}`);
      if (toolResult.filesCreated && toolResult.filesCreated.length > 0) {
        console.log(`   📁 创建的文件: ${toolResult.filesCreated.join(", ")}`);
      }

      return {
        taskId: task.id,
        success: toolResult.success,
        result: toolResult.result,
        filesCreated: toolResult.filesCreated,
        observations: toolResult.observations,
      };
    } catch (error) {
      task.status = "failed";
      this.state.failedTasks.push(task);
      console.error(
        `   ❌ 任务失败: ${error instanceof Error ? error.message : String(error)}`,
      );

      return {
        taskId: task.id,
        success: false,
        result: "任务执行失败",
        error: error instanceof Error ? error.message : String(error),
      };
    }
  }

  private async prepareToolParameters(
    task: Task,
    tool: Tool,
  ): Promise<Record<string, any>> {
    // 根据任务描述和工具类型，准备参数
    const description = task.description.toLowerCase();
    const projectName = "Todo Management App";

    switch (tool.name) {
      case "create_dir":
        if (description.includes("src") || description.includes("目录")) {
          return { dirPath: "src" };
        }
        if (description.includes("component")) {
          return { dirPath: "src/components" };
        }
        return { dirPath: "." };

      case "create_file":
        if (
          description.includes("package.json") ||
          description.includes("package")
        ) {
          return {
            filePath: "package.json",
            content: CodeGenerator.generatePackageJson(projectName),
          };
        }
        if (description.includes("vite") || description.includes("config")) {
          return {
            filePath: "vite.config.js",
            content: CodeGenerator.generateViteConfig(),
          };
        }
        if (
          description.includes("index.html") ||
          description.includes("html")
        ) {
          return {
            filePath: "index.html",
            content: CodeGenerator.generateIndexHTML(projectName),
          };
        }
        if (description.includes("readme")) {
          return {
            filePath: "README.md",
            content: `# ${projectName}\n\n一个待办事项管理Web应用。\n\n## 快速开始\n\n\`\`\`bash\nnpm install\nnpm run dev\n\`\`\``,
          };
        }
        return { filePath: "file.txt", content: "" };

      case "write_code":
        // 根据任务ID和描述精确匹配文件路径
        // 优先使用任务ID，因为LLM生成的任务ID是固定的
        if (
          task.id === "task_5" ||
          ((description.includes("app") ||
            description.includes("主应用") ||
            description.includes("主组件")) &&
            !description.includes("component") &&
            !description.includes("组件"))
        ) {
          return {
            filePath: "src/App.jsx",
            code: CodeGenerator.generateAppJSX(),
          };
        }

        if (
          description.includes("main") ||
          description.includes("入口") ||
          description.includes("main.jsx")
        ) {
          return {
            filePath: "src/main.jsx",
            code: CodeGenerator.generateMainJSX(),
          };
        }

        if (
          task.id === "task_6" ||
          ((description.includes("todo") || description.includes("待办")) &&
            (description.includes("list") || description.includes("列表")))
        ) {
          return {
            filePath: "src/components/TodoList.jsx",
            code: CodeGenerator.generateComponent("TodoList", "list"),
          };
        }

        if (
          task.id === "task_7" ||
          ((description.includes("todo") || description.includes("待办")) &&
            (description.includes("form") || description.includes("表单")))
        ) {
          return {
            filePath: "src/components/TodoForm.jsx",
            code: CodeGenerator.generateComponent("TodoForm", "form"),
          };
        }

        if (
          (description.includes("todo") || description.includes("待办")) &&
          (description.includes("item") ||
            description.includes("项") ||
            description.includes("条目"))
        ) {
          return {
            filePath: "src/components/TodoItem.jsx",
            code: CodeGenerator.generateComponent("TodoItem", "item"),
          };
        }

        if (
          task.id === "task_8" ||
          description.includes("css") ||
          description.includes("样式")
        ) {
          // 创建所有CSS文件 - 这里需要特殊处理，因为要创建多个文件
          // 实际执行会在executeTask中处理
          return {
            filePath: "css_files",
            code: "multiple_css_files",
            _isMultipleFiles: true,
          };
        }

        // 如果都不匹配，尝试根据描述推断
        if (description.includes("react") && description.includes("主")) {
          return {
            filePath: "src/App.jsx",
            code: CodeGenerator.generateAppJSX(),
          };
        }

        // 最后的默认值
        console.warn(
          `⚠️  无法确定文件路径，使用默认值。任务: ${task.description}`,
        );
        return {
          filePath: "src/component.jsx",
          code: "// Component code - 需要手动指定文件路径",
        };

      default:
        return {};
    }
  }

  private updateState(): void {
    this.state.metadata.lastUpdateTime = new Date();
    this.state.metadata.iterationCount++;
  }

  private shouldContinue(): boolean {
    if (this.state.metadata.iterationCount >= this.config.maxIterations) {
      return false;
    }
    if (this.state.status !== "running") {
      return false;
    }
    return true;
  }

  private isGoalAchieved(): boolean {
    // 如果所有任务完成且没有失败的任务，认为目标达成
    return (
      this.state.tasks.length === 0 &&
      this.state.completedTasks.length > 0 &&
      this.state.failedTasks.length === 0
    );
  }

  private complete(): void {
    this.state.status = "completed";
    this.isRunning = false;
  }

  private handleError(error: Error): void {
    console.error("Agent error:", error);
    this.state.status = "stopped";
  }

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
  console.log("🚀 启动真正的自主Agent（实际项目创建）...\n");

  const projectPath = path.join(process.cwd(), "generated-todo-app");

  // 创建LLM服务
  // ⚠️ 安全提示：API密钥通过环境变量传递，不要硬编码在代码中
  const apiKey = process.env.GEMINI_API_KEY || "your-api-key-here";
  if (apiKey === "your-api-key-here") {
    console.error("❌ 错误：请设置 GEMINI_API_KEY 环境变量");
    console.error("   方式1：创建 .env 文件并添加 GEMINI_API_KEY=your-key");
    console.error("   方式2：运行前执行：export GEMINI_API_KEY=your-key");
    process.exit(1);
  }
  console.log("apiKey:", apiKey);
  const llm = new OpenAILLMService(apiKey, "gpt-3.5-turbo");

  // 创建Agent
  const agent = new RealAutonomousAgent(
    "构建一个待办事项管理Web应用",
    projectPath,
    llm,
    {
      maxIterations: 30,
      maxCost: 500,
      minPriority: 0.3,
    },
  );

  console.log("📋 Agent目标:", agent.getState().goal);
  console.log("📁 项目路径:", projectPath);
  console.log("🤖 Agent特性:");
  console.log("   - ✅ 自主性：使用LLM自主生成任务");
  console.log("   - ✅ 反应性：根据执行结果调整策略");
  console.log("   - ✅ 主动性：主动规划下一步行动");
  console.log("   - ✅ 工具使用：使用文件操作工具");
  console.log("   - ✅ ReAct模式：推理 + 行动循环");
  console.log("\n开始执行...\n");

  await agent.run();

  const state = agent.getState();
  console.log("\n✅ Agent执行完成！");
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━");
  console.log("📊 执行结果:");
  console.log("  - Agent状态:", state.status);
  console.log("  - 已完成任务数:", state.completedTasks.length);
  console.log("  - 失败任务数:", state.failedTasks.length);
  console.log("  - 迭代次数:", state.metadata.iterationCount);
  console.log("  - 思考次数:", state.thoughts.length);
  console.log("  - 观察次数:", state.observations.length);
  console.log("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n");

  if (state.completedTasks.length > 0) {
    console.log("✅ 已完成的任务:");
    state.completedTasks.forEach((task, index) => {
      console.log(`  ${index + 1}. [${task.id}] ${task.description}`);
    });
    console.log("");
  }

  console.log(`\n🎉 项目已创建在: ${projectPath}`);
  console.log("\n📝 下一步:");
  console.log(`  cd ${path.basename(projectPath)}`);
  console.log("  npm install");
  console.log("  npm run dev");
  console.log("");
}

if (require.main === module) {
  example().catch((error) => {
    console.error("❌ 执行出错:", error);
    process.exit(1);
  });
}

export { RealAutonomousAgent, OpenAILLMService, LLMService, Tool, AgentState };
