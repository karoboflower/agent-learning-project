/**
 * 真正具有反应性的Agent - 实际项目版本
 * 
 * 参照 autonomy/typescript-autonomous-agent-real-project.ts 改造
 * 
 * 这个Agent具备真正的反应性特征：
 * 1. 真实感知：监听文件系统变动（而非模拟数据）
 * 2. 智能决策：使用LLM分析事件并决定行动（而非硬编码规则）
 * 3. 实时响应：通过事件驱动架构处理变化
 * 4. 工具执行：具备实际操作文件系统的能力
 * 
 * 场景：持续监控项目目录，当检测到文件变化时，自动进行代码审查或辅助开发
 */

import * as fs from "fs";
import * as path from "path";
import { EventEmitter } from "events";
import Anthropic from "@anthropic-ai/sdk";
import * as dotenv from "dotenv";

// 加载环境变量
dotenv.config();

// ==================== 基础接口定义 ====================

interface EnvironmentState {
  monitoredPath: string;
  files: Set<string>;
  lastActivity: Date;
  status: "idle" | "processing" | "error";
}

// 传感器接口
interface Sensor {
  id: string;
  type: string;
  start(): Promise<void>;
  stop(): Promise<void>;
  on(event: string, listener: (...args: any[]) => void): this;
}

// 反应式事件
interface ReactiveEvent {
  id: string;
  type: string; // e.g., "file:change", "file:create"
  source: string; // sensor id
  data: any;
  priority: number; // 0-1, 1 is highest
  timestamp: Date;
}

// 响应动作
interface AgentAction {
  type: string; // e.g., "write_file", "log", "ignore"
  parameters: Record<string, any>;
  reasoning: string;
}

// ==================== LLM服务实现 ====================

class LLMService {
  private client: Anthropic;
  private model: string;

  private useMock: boolean = false;

  constructor() {
    const apiKey = process.env.ANTHROPIC_API_KEY || process.env.ANTHROPIC_AUTH_TOKEN;
    const baseURL = process.env.ANTHROPIC_BASE_URL;

    if (!apiKey) {
      console.warn("⚠️ 未检测到 ANTHROPIC_API_KEY，切换到模拟模式");
      this.useMock = true;
    }

    this.client = new Anthropic({
      apiKey: apiKey || "dummy-key",
      baseURL: baseURL,
    });
    this.model = "claude-3-5-sonnet-20241022";
  }

  async analyzeEvent(event: ReactiveEvent, context: string): Promise<AgentAction> {
    if (this.useMock) {
      console.log("🤖 [MOCK] 模拟LLM思考...");
      await new Promise(resolve => setTimeout(resolve, 1000));

      if (event.type === "file:delete") {
        return {
          type: "log",
          parameters: { message: `检测到文件被删除: ${event.data.filename}` },
          reasoning: "模拟模式：文件删除事件"
        };
      }

      const content = event.data.content || "";
      if (content.includes("hello")) {
        return {
          type: "write_file",
          parameters: {
            filePath: "reply.txt",
            content: "Hi there! I see you said hello."
          },
          reasoning: "模拟模式：检测到 hello，自动回复"
        };
      }

      return {
        type: "log",
        parameters: { message: `已处理变更: ${event.data.filename}` },
        reasoning: "模拟模式：默认日志记录"
      };
    }

    const prompt = `
你是一个具有反应性的智能代码助手 Agent。
检测到一个新的事件，请分析并决定如何响应。

事件类型: ${event.type}
事件数据: ${JSON.stringify(event.data, null, 2)}
当前上下文: ${context}

你的任务是：
1. 分析文件变更的内容或含义
2. 判断是否需要采取行动（例如：发现明显的代码错误需要修复、需要添加注释、或者只是记录日志）
3. 如果是无关紧要的变更（如自动生成的日志、临时文件），请选择 "ignore"

可用的行动类型(type)：
- "write_file": 修改或创建文件 (参数: filePath, content)
- "log": 记录重要信息 (参数: message)
- "ignore": 忽略此次变更 (无参数)

请以 JSON 格式返回你的决定：
{
  "type": "...",
  "parameters": { ... },
  "reasoning": "..."
}
`;

    try {
      const response = await this.client.messages.create({
        model: this.model,
        max_tokens: 1000,
        messages: [{ role: "user", content: prompt }]
      });

      const content = response.content[0].type === 'text' ? response.content[0].text : '';
      const jsonMatch = content.match(/\{[\s\S]*\}/);

      if (jsonMatch) {
        return JSON.parse(jsonMatch[0]);
      }
      return { type: "log", parameters: { message: "无法解析LLM响应" }, reasoning: "Parse Error" };
    } catch (error) {
      console.error("LLM 调用失败:", error);
      return { type: "ignore", parameters: {}, reasoning: "LLM Error" };
    }
  }
}

// ==================== 传感器实现 ====================

class FileSystemSensor extends EventEmitter implements Sensor {
  id: string;
  type = "file_system";
  private watcher: fs.FSWatcher | null = null;
  private monitoredPath: string;
  private ignorePatterns: RegExp[] = [/node_modules/, /\.git/, /\.log$/, /dist/];
  private processingFiles: Set<string> = new Set();

  constructor(id: string, path: string) {
    super();
    this.id = id;
    this.monitoredPath = path;
  }

  async start(): Promise<void> {
    if (!fs.existsSync(this.monitoredPath)) {
      fs.mkdirSync(this.monitoredPath, { recursive: true });
    }

    console.log(`👁️  启动文件系统传感器，监控路径: ${this.monitoredPath}`);

    // 使用简单的 fs.watch，实际生产中推荐 chokidar
    this.watcher = fs.watch(this.monitoredPath, { recursive: true }, (eventType, filename) => {
      if (!filename) return;
      if (this.shouldIgnore(filename)) return;

      // 简单的防抖：如果在处理中则忽略（避免响应自己产生的变更）
      if (this.processingFiles.has(filename)) return;

      const fullPath = path.join(this.monitoredPath, filename);

      // 检测文件是否存在以区分 删除 vs 修改/新增
      let eventTypeExplicit = "file:change";
      let fileContent = null;

      try {
        if (fs.existsSync(fullPath)) {
          const stat = fs.statSync(fullPath);
          if (stat.isFile()) {
            fileContent = fs.readFileSync(fullPath, 'utf-8');
          } else {
            return; // 忽略目录变更
          }
        } else {
          eventTypeExplicit = "file:delete";
        }
      } catch (e) {
        // 文件可能在读取时被锁定或再次删除
        return;
      }

      this.emit("event", {
        id: `evt_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`,
        type: eventTypeExplicit,
        source: this.id,
        data: {
          filename,
          eventType: eventTypeExplicit,
          content: fileContent ? fileContent.slice(0, 500) + (fileContent.length > 500 ? "..." : "") : null, // 只发送前500字符避免token溢出
          timestamp: new Date()
        },
        priority: 0.8, // 文件变更高优先级
        timestamp: new Date()
      } as ReactiveEvent);
    });
  }

  async stop(): Promise<void> {
    if (this.watcher) {
      this.watcher.close();
      this.watcher = null;
    }
  }

  private shouldIgnore(filename: string): boolean {
    return this.ignorePatterns.some(regex => regex.test(filename));
  }

  // 标记文件正在被Agent处理，避免循环触发
  markProcessing(filename: string) {
    this.processingFiles.add(filename);
    setTimeout(() => this.processingFiles.delete(filename), 2000);
  }
}

// ==================== 反应性Agent实现 ====================

class ReactiveAgent {
  private sensors: Sensor[] = [];
  private llm: LLMService;
  private isRunning = false;
  private eventQueue: ReactiveEvent[] = [];
  private processing: boolean = false;
  private monitoredPath: string;

  constructor(monitoredPath: string) {
    this.monitoredPath = monitoredPath;
    this.llm = new LLMService();
    this.setupSensors();
  }

  private setupSensors() {
    const fsSensor = new FileSystemSensor("fs_sensor_main", this.monitoredPath);

    // 监听传感器产生的事件
    fsSensor.on("event", (event: ReactiveEvent) => {
      this.handleIncomingEvent(event);
    });

    this.sensors.push(fsSensor);
  }

  private handleIncomingEvent(event: ReactiveEvent) {
    console.log(`\n📨 收到事件 [${event.type}]: ${event.data.filename}`);
    this.eventQueue.push(event);
    // 按照优先级排序
    this.eventQueue.sort((a, b) => b.priority - a.priority);

    this.processQueue();
  }

  private async processQueue() {
    if (this.processing) return;
    this.processing = true;

    while (this.eventQueue.length > 0) {
      const event = this.eventQueue.shift();
      if (event) {
        await this.react(event);
      }
    }

    this.processing = false;
  }

  private async react(event: ReactiveEvent) {
    console.log(`🤔 正在分析事件...`);

    // 上下文：可以是项目状态、最近的操作等
    const context = `监控目录: ${this.monitoredPath}`;

    try {
      // 1. 认知：使用LLM分析
      const action = await this.llm.analyzeEvent(event, context);
      console.log(`💡 决策: ${action.type} - 原因: ${action.reasoning}`);

      // 2. 行动：执行决策
      await this.executeAction(action, event);

    } catch (error) {
      console.error("❌ 反应过程出错:", error);
    }
  }

  private async executeAction(action: AgentAction, triggerEvent: ReactiveEvent) {
    switch (action.type) {
      case "write_file":
        await this.handleWriteFile(action.parameters, triggerEvent);
        break;
      case "log":
        console.log(`📝 记录: ${action.parameters.message}`);
        break;
      case "ignore":
        console.log(`IGNORE: 忽略此事`);
        break;
      default:
        console.warn(`⚠️ 未知行动类型: ${action.type}`);
    }
  }

  private async handleWriteFile(params: any, triggerEvent: ReactiveEvent) {
    const { filePath, content } = params;
    if (!filePath || !content) return;

    const targetPath = path.join(this.monitoredPath, filePath);

    // 标记传感器忽略此文件，防止循环
    const fsSensor = this.sensors.find(s => s.id === triggerEvent.source) as FileSystemSensor;
    if (fsSensor) {
      fsSensor.markProcessing(filePath);
    }

    // 确保目录存在
    const dir = path.dirname(targetPath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }

    fs.writeFileSync(targetPath, content, "utf-8");
    console.log(`✅ 已写入文件: ${filePath}`);
  }

  async start() {
    if (this.isRunning) return;
    this.isRunning = true;
    console.log("🚀 反应性 Agent (Real Project) 已启动");
    console.log("----------------------------------------");

    // 启动所有传感器
    await Promise.all(this.sensors.map(s => s.start()));

    // 保持进程运行
    process.on('SIGINT', async () => {
      await this.stop();
      process.exit(0);
    });
  }

  async stop() {
    this.isRunning = false;
    console.log("\n🛑 Agent 正在停止...");
    await Promise.all(this.sensors.map(s => s.stop()));
    console.log("✅ Agent 已停止");
  }
}

// ==================== 主入口 ====================

async function main() {
  // 定义监控的目录（默认为当前目录下的 monitored_project 文件夹，避免污染根目录）
  const targetDir = path.join(__dirname, "monitored_project");

  const agent = new ReactiveAgent(targetDir);
  await agent.start();

  console.log(`\n你可以尝试在 ${targetDir} 目录下创建或修改文件。`);
  console.log("示例：创建一个名为 'hello.txt' 的文件，内容为 'hello world'");
  console.log("Agent 将会检测到变化并做出反应。\n");
}

main().catch(console.error);
