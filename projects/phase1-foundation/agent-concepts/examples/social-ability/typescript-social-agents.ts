/**
 * 具有社会性的Multi-Agent系统 - 代码审查团队
 *
 * 这个系统展示了Agent的社会性特征：
 * 1. 通信能力：Agent之间通过消息传递进行通信
 * 2. 协作能力：多个Agent协作完成代码审查任务
 * 3. 角色分工：不同Agent有不同的专业角色
 * 4. 协调机制：协调者Agent负责任务分配和结果整合
 *
 * 场景：三个Agent协作审查代码
 * - AnalyzerAgent: 分析代码结构和复杂度
 * - ReviewerAgent: 提出改进建议
 * - CoordinatorAgent: 协调整个审查流程
 */

import * as fs from "fs";
import * as path from "path";
import Anthropic from "@anthropic-ai/sdk";
import * as dotenv from "dotenv";
import { EventEmitter } from "events";

// 加载环境变量
dotenv.config();

// ==================== 基础接口定义 ====================

interface Message {
  id: string;
  from: string;
  to: string;
  type: string;
  content: any;
  timestamp: Date;
  conversationId?: string;
}

interface Agent {
  id: string;
  role: string;
  capabilities: string[];
}

interface CodeAnalysisResult {
  filePath: string;
  linesOfCode: number;
  complexity: string;
  issues: string[];
  strengths: string[];
  analyzer: string;
}

interface ReviewSuggestion {
  category: string;
  priority: 'high' | 'medium' | 'low';
  description: string;
  recommendation: string;
  reviewer: string;
}

interface ReviewReport {
  filePath: string;
  analysis: CodeAnalysisResult;
  suggestions: ReviewSuggestion[];
  overallScore: number;
  summary: string;
  timestamp: Date;
}

// ==================== 消息总线 ====================

class MessageBus extends EventEmitter {
  private messageLog: Message[] = [];

  send(message: Message): void {
    this.messageLog.push(message);
    console.log(`\n📨 [消息总线] ${message.from} -> ${message.to}: ${message.type}`);

    // 触发消息事件
    this.emit('message', message);
    this.emit(`message:${message.to}`, message);
  }

  getMessageLog(): Message[] {
    return [...this.messageLog];
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
      throw new Error("❌ ANTHROPIC_API_KEY 环境变量未设置。请在 .env 文件中配置 API 密钥。");
    }

    this.client = new Anthropic({
      apiKey: apiKey,
      baseURL: baseURL,
    });
    this.model = "claude-3-5-sonnet-20241022";
  }

  async analyze(prompt: string): Promise<any> {
    try {
      const response = await this.client.messages.create({
        model: this.model,
        max_tokens: 2000,
        messages: [{ role: "user", content: prompt }]
      });

      const content = response.content[0].type === 'text' ? response.content[0].text : '';

      // 尝试解析JSON
      const jsonMatch = content.match(/\{[\s\S]*\}|\[[\s\S]*\]/);
      if (jsonMatch) {
        try {
          return JSON.parse(jsonMatch[0]);
        } catch {
          return { rawText: content };
        }
      }

      return { rawText: content };
    } catch (error) {
      console.error("❌ LLM调用失败:", error);
      throw error;
    }
  }
}

// ==================== 基础Agent类 ====================

abstract class SocialAgent {
  protected agentId: string;
  protected role: string;
  protected capabilities: string[];
  protected messageBus: MessageBus;
  protected llm: LLMService;
  protected mailbox: Message[] = [];
  protected isRunning: boolean = false;

  constructor(id: string, role: string, capabilities: string[], messageBus: MessageBus, llm: LLMService) {
    this.agentId = id;
    this.role = role;
    this.capabilities = capabilities;
    this.messageBus = messageBus;
    this.llm = llm;

    // 监听发给自己的消息
    this.messageBus.on(`message:${this.agentId}`, (message: Message) => {
      this.receiveMessage(message);
    });
  }

  // 发送消息
  protected async sendMessage(to: string, type: string, content: any, conversationId?: string): Promise<void> {
    const message: Message = {
      id: `msg_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`,
      from: this.agentId,
      to: to,
      type: type,
      content: content,
      timestamp: new Date(),
      conversationId: conversationId
    };

    this.messageBus.send(message);
  }

  // 接收消息
  protected receiveMessage(message: Message): void {
    this.mailbox.push(message);
    console.log(`  📬 [${this.agentId}] 收到消息: ${message.type}`);

    // 异步处理消息
    this.handleMessage(message).catch(error => {
      console.error(`❌ [${this.agentId}] 处理消息失败:`, error);
    });
  }

  // 处理消息（子类实现）
  protected abstract handleMessage(message: Message): Promise<void>;

  // 等待特定类型的响应
  protected async waitForResponse(conversationId: string, timeout: number = 30000): Promise<Message> {
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        reject(new Error(`等待响应超时: ${conversationId}`));
      }, timeout);

      const checkMailbox = setInterval(() => {
        const response = this.mailbox.find(
          msg => msg.conversationId === conversationId && msg.from !== this.agentId
        );

        if (response) {
          clearTimeout(timer);
          clearInterval(checkMailbox);
          resolve(response);
        }
      }, 100);
    });
  }

  // 启动Agent
  async start(): Promise<void> {
    this.isRunning = true;
    console.log(`✅ [${this.agentId}] ${this.role} 已启动`);
  }

  // 停止Agent
  async stop(): Promise<void> {
    this.isRunning = false;
    console.log(`🛑 [${this.agentId}] ${this.role} 已停止`);
  }

  getId(): string {
    return this.agentId;
  }

  getRole(): string {
    return this.role;
  }
}

// ==================== 分析Agent ====================

class AnalyzerAgent extends SocialAgent {
  constructor(messageBus: MessageBus, llm: LLMService) {
    super('analyzer_01', 'Code Analyzer', ['code_analysis', 'complexity_analysis'], messageBus, llm);
  }

  protected async handleMessage(message: Message): Promise<void> {
    switch (message.type) {
      case 'analyze_request':
        await this.handleAnalyzeRequest(message);
        break;
      default:
        console.log(`  ⚠️ [${this.agentId}] 未知消息类型: ${message.type}`);
    }
  }

  private async handleAnalyzeRequest(message: Message): Promise<void> {
    const { filePath, content } = message.content;

    console.log(`\n🔍 [分析Agent] 开始分析代码: ${filePath}`);

    try {
      // 使用LLM分析代码
      const analysis = await this.analyzeCode(filePath, content);

      // 发送分析结果
      await this.sendMessage(
        message.from,
        'analyze_result',
        analysis,
        message.conversationId
      );

      console.log(`✅ [分析Agent] 分析完成`);
    } catch (error) {
      console.error(`❌ [分析Agent] 分析失败:`, error);
      await this.sendMessage(
        message.from,
        'analyze_error',
        { error: (error as Error).message },
        message.conversationId
      );
    }
  }

  private async analyzeCode(filePath: string, content: string): Promise<CodeAnalysisResult> {
    const prompt = `
你是一个专业的代码分析Agent。请分析以下代码：

文件路径: ${filePath}
代码内容:
\`\`\`
${content.slice(0, 2000)}${content.length > 2000 ? '\n...(truncated)' : ''}
\`\`\`

请提供以下分析：
1. 代码行数估计
2. 复杂度评估（简单/中等/复杂）
3. 发现的问题（数组）
4. 代码优点（数组）

返回JSON格式：
{
  "linesOfCode": 数字,
  "complexity": "简单/中等/复杂",
  "issues": ["问题1", "问题2"],
  "strengths": ["优点1", "优点2"]
}
`;

    const result = await this.llm.analyze(prompt);

    return {
      filePath: filePath,
      linesOfCode: result.linesOfCode || content.split('\n').length,
      complexity: result.complexity || '中等',
      issues: result.issues || [],
      strengths: result.strengths || [],
      analyzer: this.agentId
    };
  }
}

// ==================== 审查Agent ====================

class ReviewerAgent extends SocialAgent {
  constructor(messageBus: MessageBus, llm: LLMService) {
    super('reviewer_01', 'Code Reviewer', ['code_review', 'best_practices'], messageBus, llm);
  }

  protected async handleMessage(message: Message): Promise<void> {
    switch (message.type) {
      case 'review_request':
        await this.handleReviewRequest(message);
        break;
      default:
        console.log(`  ⚠️ [${this.agentId}] 未知消息类型: ${message.type}`);
    }
  }

  private async handleReviewRequest(message: Message): Promise<void> {
    const { analysis, content } = message.content;

    console.log(`\n📝 [审查Agent] 开始审查代码，基于分析结果`);

    try {
      // 基于分析结果生成审查建议
      const suggestions = await this.generateSuggestions(analysis, content);

      // 发送审查建议
      await this.sendMessage(
        message.from,
        'review_result',
        suggestions,
        message.conversationId
      );

      console.log(`✅ [审查Agent] 审查完成，提出 ${suggestions.length} 条建议`);
    } catch (error) {
      console.error(`❌ [审查Agent] 审查失败:`, error);
      await this.sendMessage(
        message.from,
        'review_error',
        { error: (error as Error).message },
        message.conversationId
      );
    }
  }

  private async generateSuggestions(analysis: CodeAnalysisResult, content: string): Promise<ReviewSuggestion[]> {
    const prompt = `
你是一个专业的代码审查Agent。基于以下代码分析结果，请提出改进建议：

代码分析:
- 复杂度: ${analysis.complexity}
- 发现的问题: ${analysis.issues.join(', ')}
- 代码优点: ${analysis.strengths.join(', ')}

代码内容（部分）:
\`\`\`
${content.slice(0, 1000)}
\`\`\`

请提供3-5条具体的改进建议，每条建议包括：
- category: 类别（如"代码质量"、"性能"、"可维护性"）
- priority: 优先级（high/medium/low）
- description: 问题描述
- recommendation: 具体建议

返回JSON数组格式：
[
  {
    "category": "类别",
    "priority": "high/medium/low",
    "description": "问题描述",
    "recommendation": "具体建议"
  }
]
`;

    const result = await this.llm.analyze(prompt);

    // 确保result是数组
    const suggestions = Array.isArray(result) ? result : (result.suggestions || []);

    return suggestions.map((s: any) => ({
      category: s.category || '代码质量',
      priority: s.priority || 'medium',
      description: s.description || '',
      recommendation: s.recommendation || '',
      reviewer: this.agentId
    }));
  }
}

// ==================== 协调Agent ====================

class CoordinatorAgent extends SocialAgent {
  constructor(messageBus: MessageBus, llm: LLMService) {
    super('coordinator_01', 'Coordinator', ['task_coordination', 'result_integration'], messageBus, llm);
  }

  protected async handleMessage(message: Message): Promise<void> {
    // 协调者主要是发��者，不太处理消息
    console.log(`  📩 [协调Agent] 收到消息: ${message.type}`);
  }

  // 发起代码审查流程
  async initiateCodeReview(filePath: string): Promise<ReviewReport> {
    console.log(`\n🎯 [协调Agent] 发起代码审查流程: ${filePath}`);

    try {
      // 1. 读取文件
      const content = fs.readFileSync(filePath, 'utf-8');
      const conversationId = `review_${Date.now()}`;

      // 2. 请求分析Agent分析代码
      console.log(`\n📤 [协调Agent] 请求分析Agent分析代码...`);
      await this.sendMessage(
        'analyzer_01',
        'analyze_request',
        { filePath, content },
        conversationId
      );

      // 3. 等待分析结果
      const analysisMsg = await this.waitForResponse(conversationId);
      if (analysisMsg.type !== 'analyze_result') {
        throw new Error('分析失败');
      }
      const analysis: CodeAnalysisResult = analysisMsg.content;

      // 4. 请求审查Agent提供建议
      console.log(`\n📤 [协调Agent] 请求审查Agent提供建议...`);
      await this.sendMessage(
        'reviewer_01',
        'review_request',
        { analysis, content },
        conversationId
      );

      // 5. 等待审查建议
      const reviewMsg = await this.waitForResponse(conversationId);
      if (reviewMsg.type !== 'review_result') {
        throw new Error('审查失败');
      }
      const suggestions: ReviewSuggestion[] = reviewMsg.content;

      // 6. 整合结果生成报告
      console.log(`\n📊 [协调Agent] 整合结果生成报告...`);
      const report = await this.generateReport(filePath, analysis, suggestions);

      console.log(`\n✅ [协调Agent] 代码审查流程完成`);

      return report;
    } catch (error) {
      console.error(`\n❌ [协调Agent] 代码审查流程失败:`, error);
      throw error;
    }
  }

  private async generateReport(
    filePath: string,
    analysis: CodeAnalysisResult,
    suggestions: ReviewSuggestion[]
  ): Promise<ReviewReport> {
    // 使用LLM生成总结
    const prompt = `
你是一个代码审查协调Agent。请基于以下信息生成审查总结：

代码分析:
- 文件: ${filePath}
- 行数: ${analysis.linesOfCode}
- 复杂度: ${analysis.complexity}
- 问题: ${analysis.issues.join(', ')}
- 优点: ${analysis.strengths.join(', ')}

审查建议数量: ${suggestions.length}
高优先级建议: ${suggestions.filter(s => s.priority === 'high').length}

请生成一个简洁的总结（2-3句话）和整体评分（0-100分）。

返回JSON格式：
{
  "summary": "总结内容",
  "overallScore": 分数
}
`;

    const result = await this.llm.analyze(prompt);

    return {
      filePath: filePath,
      analysis: analysis,
      suggestions: suggestions,
      overallScore: result.overallScore || 75,
      summary: result.summary || '代码审查完成',
      timestamp: new Date()
    };
  }
}

// ==================== 多Agent系统 ====================

class MultiAgentCodeReviewSystem {
  private messageBus: MessageBus;
  private llm: LLMService;
  private agents: Map<string, SocialAgent>;
  private analyzer: AnalyzerAgent;
  private reviewer: ReviewerAgent;
  private coordinator: CoordinatorAgent;

  constructor() {
    this.messageBus = new MessageBus();
    this.llm = new LLMService();
    this.agents = new Map();

    // 创建Agent
    this.analyzer = new AnalyzerAgent(this.messageBus, this.llm);
    this.reviewer = new ReviewerAgent(this.messageBus, this.llm);
    this.coordinator = new CoordinatorAgent(this.messageBus, this.llm);

    this.agents.set(this.analyzer.getId(), this.analyzer);
    this.agents.set(this.reviewer.getId(), this.reviewer);
    this.agents.set(this.coordinator.getId(), this.coordinator);
  }

  async start(): Promise<void> {
    console.log("\n🚀 多Agent代码审查系统启动");
    console.log("========================================");
    console.log(`👥 系统包含 ${this.agents.size} 个Agent:`);

    for (const agent of this.agents.values()) {
      await agent.start();
    }

    console.log("========================================\n");
  }

  async reviewCode(filePath: string): Promise<ReviewReport> {
    const report = await this.coordinator.initiateCodeReview(filePath);
    return report;
  }

  async stop(): Promise<void> {
    console.log("\n🛑 系统正在停止...");

    for (const agent of this.agents.values()) {
      await agent.stop();
    }

    console.log("✅ 系统已停止\n");
  }

  printReport(report: ReviewReport): void {
    console.log("\n");
    console.log("============================================================");
    console.log("📋 代码审查报告");
    console.log("============================================================");
    console.log(`📁 文件: ${report.filePath}`);
    console.log(`📊 评分: ${report.overallScore}/100`);
    console.log(`⏰ 时间: ${report.timestamp.toLocaleString()}`);
    console.log("\n--- 代码分析 ---");
    console.log(`📏 代码行数: ${report.analysis.linesOfCode}`);
    console.log(`🔢 复杂度: ${report.analysis.complexity}`);
    console.log(`❌ 问题 (${report.analysis.issues.length}):`);
    report.analysis.issues.forEach((issue, i) => {
      console.log(`   ${i + 1}. ${issue}`);
    });
    console.log(`✅ 优点 (${report.analysis.strengths.length}):`);
    report.analysis.strengths.forEach((strength, i) => {
      console.log(`   ${i + 1}. ${strength}`);
    });
    console.log("\n--- 改进建议 ---");
    report.suggestions.forEach((suggestion, i) => {
      const priorityEmoji = suggestion.priority === 'high' ? '🔴' : suggestion.priority === 'medium' ? '🟡' : '🟢';
      console.log(`\n${i + 1}. [${suggestion.category}] ${priorityEmoji} ${suggestion.priority.toUpperCase()}`);
      console.log(`   问题: ${suggestion.description}`);
      console.log(`   建议: ${suggestion.recommendation}`);
    });
    console.log("\n--- 总结 ---");
    console.log(report.summary);
    console.log("============================================================\n");
  }
}

// ==================== 主入口 ====================

async function main() {
  // 创建多Agent系统
  const system = new MultiAgentCodeReviewSystem();

  try {
    // 启动系统
    await system.start();

    // 创建测试代码文件
    const testDir = path.join(__dirname, "test_project");
    if (!fs.existsSync(testDir)) {
      fs.mkdirSync(testDir, { recursive: true });
    }

    const testFile = path.join(testDir, "example.ts");
    const testCode = `
function calculateTotal(items: any[]) {
  let total = 0;
  for (let i = 0; i < items.length; i++) {
    total += items[i].price;
  }
  return total;
}

class UserManager {
  private users = [];

  addUser(user) {
    this.users.push(user);
  }

  getUser(id) {
    for (let i = 0; i < this.users.length; i++) {
      if (this.users[i].id == id) {
        return this.users[i];
      }
    }
  }
}
`;

    fs.writeFileSync(testFile, testCode, 'utf-8');
    console.log(`📝 已创建测试文件: ${testFile}\n`);

    // 执行代码审查
    const report = await system.reviewCode(testFile);

    // 打印报告
    system.printReport(report);

    // 保存报告到文件
    const reportFile = path.join(testDir, "review_report.json");
    fs.writeFileSync(reportFile, JSON.stringify(report, null, 2), 'utf-8');
    console.log(`💾 报告已保存到: ${reportFile}\n`);

    // 停止系统
    await system.stop();

  } catch (error) {
    console.error("\n❌ 系统错误:", error);
    await system.stop();
    process.exit(1);
  }
}

main().catch(console.error);
