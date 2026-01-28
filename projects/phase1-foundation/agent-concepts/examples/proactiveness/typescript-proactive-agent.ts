/**
 * 真正具有主动性的Agent - 实际项目版本
 *
 * 这个Agent具备真正的主动性特征：
 * 1. 目标驱动：主动追求设定的目标
 * 2. 机会识别：主动扫描环境寻找机会
 * 3. 预测性行为：预测未来需求并提前准备
 * 4. 主动学习：不断改进自身策略
 *
 * 场景：主动监控项目目录，发现代码质量问题、优化机会，并主动采取改进措施
 */

import * as fs from "fs";
import * as path from "path";
import Anthropic from "@anthropic-ai/sdk";
import * as dotenv from "dotenv";

// 加载环境变量
dotenv.config();

// ==================== 基础接口定义 ====================

interface Goal {
  id: string;
  description: string;
  priority: number; // 0-1
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  createdAt: Date;
  deadline?: Date;
}

interface Opportunity {
  id: string;
  type: string; // e.g., "code_improvement", "documentation_gap"
  description: string;
  value: number; // 0-1, expected benefit
  cost: number; // 0-1, estimated effort
  discoveredAt: Date;
}

interface Prediction {
  id: string;
  type: string;
  description: string;
  confidence: number; // 0-1
  expectedTime: Date;
  requiredActions: string[];
}

interface ProactiveAction {
  id: string;
  type: string;
  description: string;
  reasoning: string;
  parameters: Record<string, any>;
  expectedBenefit: number;
}

// ==================== LLM服务实现 ====================

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

  /**
   * 主动分析代码文件并识别改进机会
   */
  async analyzeCodeForOpportunities(filePath: string, content: string): Promise<Opportunity[]> {
    const prompt = `
你是一个主动的代码质量Agent。请分析以下代码文件并识别改进机会。

文件路径: ${filePath}
代码内容:
\`\`\`
${content.slice(0, 2000)}${content.length > 2000 ? '\n...(truncated)' : ''}
\`\`\`

请识别以下类型的机会：
1. 代码质量改进（重复代码、复杂度过高等）
2. 文档缺失或不完整
3. 潜在的bug或安全问题
4. 性能优化机会
5. TODO/FIXME项

返回JSON格式的机会列表：
[
  {
    "type": "机会类型",
    "description": "详细描述",
    "value": 0.7,  // 预期收益 (0-1)
    "cost": 0.3    // 预估工作量 (0-1)
  }
]

如果没有发现明显的机会，返回空数组 []
`;

    try {
      const response = await this.client.messages.create({
        model: this.model,
        max_tokens: 2000,
        messages: [{ role: "user", content: prompt }]
      });

      const content = response.content[0].type === 'text' ? response.content[0].text : '';
      const jsonMatch = content.match(/\[[\s\S]*?\]/);

      if (jsonMatch) {
        const opportunities = JSON.parse(jsonMatch[0]);
        return opportunities.map((opp: any, index: number) => ({
          id: `opp_${Date.now()}_${index}`,
          type: opp.type || 'unknown',
          description: opp.description || '',
          value: opp.value || 0.5,
          cost: opp.cost || 0.5,
          discoveredAt: new Date()
        }));
      }

      return [];
    } catch (error) {
      console.error("❌ LLM分析失败:", error);
      return [];
    }
  }

  /**
   * 主动生成改进代码的建议
   */
  async generateImprovementSuggestion(opportunity: Opportunity, context: string): Promise<ProactiveAction | null> {
    const prompt = `
你是一个主动的代码改进Agent。发现了一个改进机会，请制定行动计划。

机会类型: ${opportunity.type}
机会描述: ${opportunity.description}
预期收益: ${opportunity.value}
预估成本: ${opportunity.cost}

上下文: ${context}

请分析这个机会并决定采取什么行动。可选行动：
1. "write_improvement_report" - 生成改进报告
2. "create_todo" - 创建TODO任务文件
3. "improve_code" - 直接改进代码（仅限简单改进）
4. "skip" - 跳过这个机会

返回JSON格式的行动计划：
{
  "type": "行动类型",
  "description": "行动描述",
  "reasoning": "为什么采取这个行动",
  "parameters": { "具体参数" }
}
`;

    try {
      const response = await this.client.messages.create({
        model: this.model,
        max_tokens: 1500,
        messages: [{ role: "user", content: prompt }]
      });

      const content = response.content[0].type === 'text' ? response.content[0].text : '';
      const jsonMatch = content.match(/\{[\s\S]*?\}/);

      if (jsonMatch) {
        const action = JSON.parse(jsonMatch[0]);
        return {
          id: `action_${Date.now()}`,
          type: action.type || 'skip',
          description: action.description || '',
          reasoning: action.reasoning || '',
          parameters: action.parameters || {},
          expectedBenefit: opportunity.value
        };
      }

      return null;
    } catch (error) {
      console.error("❌ LLM生成建议失败:", error);
      return null;
    }
  }

  /**
   * 主动预测未来需求
   */
  async predictFutureNeeds(projectState: any): Promise<Prediction[]> {
    const prompt = `
你是一个具有预测能力的主动Agent。请根据当前项目状态预测未来可能的需求。

项目状态:
${JSON.stringify(projectState, null, 2)}

请预测可能出现的情况，例如：
1. 代码库增长带来的维护需求
2. 可能出现的问题
3. 优化机会

返回JSON格式的预测列表：
[
  {
    "type": "预测类型",
    "description": "详细描述",
    "confidence": 0.7,  // 信心度 (0-1)
    "requiredActions": ["需要采取的行动"]
  }
]
`;

    try {
      const response = await this.client.messages.create({
        model: this.model,
        max_tokens: 1500,
        messages: [{ role: "user", content: prompt }]
      });

      const content = response.content[0].type === 'text' ? response.content[0].text : '';
      const jsonMatch = content.match(/\[[\s\S]*?\]/);

      if (jsonMatch) {
        const predictions = JSON.parse(jsonMatch[0]);
        return predictions.map((pred: any, index: number) => ({
          id: `pred_${Date.now()}_${index}`,
          type: pred.type || 'unknown',
          description: pred.description || '',
          confidence: pred.confidence || 0.5,
          expectedTime: new Date(Date.now() + 3600000), // 默认1小时后
          requiredActions: pred.requiredActions || []
        }));
      }

      return [];
    } catch (error) {
      console.error("❌ LLM预测失败:", error);
      return [];
    }
  }
}

// ==================== 主动性Agent实现 ====================

class ProactiveAgent {
  private goals: Goal[] = [];
  private opportunities: Opportunity[] = [];
  private predictions: Prediction[] = [];
  private llm: LLMService;
  private isRunning = false;
  private projectPath: string;

  // 主动性配置
  private config = {
    opportunityScanInterval: 30000, // 30秒扫描一次机会
    predictionInterval: 60000, // 60秒预测一次
    opportunityThreshold: 0.3, // 机会价值阈值
    maxActionsPerCycle: 3 // 每个周期最多执行3个行动
  };

  private statistics = {
    opportunitiesFound: 0,
    actionsExecuted: 0,
    goalsCompleted: 0,
    startTime: new Date()
  };

  constructor(projectPath: string) {
    this.projectPath = projectPath;
    this.llm = new LLMService();
    this.initializeGoals();
  }

  private initializeGoals() {
    // 设置初始目标
    this.goals.push({
      id: 'goal_1',
      description: '提升项目代码质量',
      priority: 0.9,
      status: 'pending',
      createdAt: new Date()
    });

    this.goals.push({
      id: 'goal_2',
      description: '完善项目文档',
      priority: 0.7,
      status: 'pending',
      createdAt: new Date()
    });
  }

  /**
   * 启动Agent
   */
  async start() {
    if (this.isRunning) return;
    this.isRunning = true;

    console.log("🚀 主动性 Agent 已启动");
    console.log("========================================");
    console.log(`📂 监控路径: ${this.projectPath}`);
    console.log(`🎯 当前目标: ${this.goals.length} 个`);
    console.log("========================================\n");

    // 确保项目目录存在
    if (!fs.existsSync(this.projectPath)) {
      fs.mkdirSync(this.projectPath, { recursive: true });
      console.log(`📁 创建项目目录: ${this.projectPath}\n`);
    }

    // 启动多个主动行为循环
    await Promise.all([
      this.goalPursuitLoop(),
      this.opportunityScanLoop(),
      this.predictionLoop()
    ]);
  }

  /**
   * 目标追求循环
   */
  private async goalPursuitLoop() {
    while (this.isRunning) {
      try {
        // 主动选择并追求目标
        const activeGoals = this.goals.filter(g => g.status === 'pending' || g.status === 'in_progress');

        if (activeGoals.length > 0) {
          const goal = this.selectNextGoal(activeGoals);
          await this.pursueGoal(goal);
        }

        await this.sleep(5000); // 5秒
      } catch (error) {
        console.error("❌ 目标追求错误:", error);
      }
    }
  }

  /**
   * 机会扫描循环
   */
  private async opportunityScanLoop() {
    while (this.isRunning) {
      try {
        console.log("\n🔍 [主动] 扫描项目寻找改进机会...");

        // 主动扫描文件
        const files = await this.scanProjectFiles();

        for (const file of files) {
          if (!this.isRunning) break;

          // 主动分析文件
          const content = fs.readFileSync(file, 'utf-8');
          const opportunities = await this.llm.analyzeCodeForOpportunities(file, content);

          // 记录发现的机会
          for (const opp of opportunities) {
            if (opp.value >= this.config.opportunityThreshold) {
              this.opportunities.push(opp);
              this.statistics.opportunitiesFound++;
              console.log(`💡 [发现机会] ${opp.description} (价值: ${opp.value.toFixed(2)})`);
            }
          }
        }

        // 主动利用最佳机会
        await this.seizeTopOpportunities();

        await this.sleep(this.config.opportunityScanInterval);
      } catch (error) {
        console.error("❌ 机会扫描错误:", error);
        await this.sleep(this.config.opportunityScanInterval);
      }
    }
  }

  /**
   * 预测循环
   */
  private async predictionLoop() {
    while (this.isRunning) {
      try {
        console.log("\n🔮 [主动] 预测未来需求...");

        // 收集项目状态
        const projectState = await this.gatherProjectState();

        // 主动预测
        const predictions = await this.llm.predictFutureNeeds(projectState);

        for (const prediction of predictions) {
          if (prediction.confidence > 0.6) {
            this.predictions.push(prediction);
            console.log(`📊 [预测] ${prediction.description} (信心度: ${prediction.confidence.toFixed(2)})`);

            // 主动为高可信度的预测做准备
            await this.prepareForPrediction(prediction);
          }
        }

        await this.sleep(this.config.predictionInterval);
      } catch (error) {
        console.error("❌ 预测错误:", error);
        await this.sleep(this.config.predictionInterval);
      }
    }
  }

  /**
   * 主动追求目标
   */
  private async pursueGoal(goal: Goal) {
    if (goal.status === 'pending') {
      goal.status = 'in_progress';
      console.log(`\n🎯 [主动] 开始追求目标: ${goal.description}`);
    }

    // 根据目标类型采取不同行动
    if (goal.description.includes('代码质量')) {
      // 主动寻找代码质量改进机会
      await this.opportunityScanLoop();
    } else if (goal.description.includes('文档')) {
      // 主动寻找文档缺失
      await this.scanForDocumentationGaps();
    }
  }

  /**
   * 选择下一个目标
   */
  private selectNextGoal(goals: Goal[]): Goal {
    // 按优先级排序
    goals.sort((a, b) => b.priority - a.priority);
    return goals[0];
  }

  /**
   * 主动利用最佳机会
   */
  private async seizeTopOpportunities() {
    // 按价值排序
    const sortedOpps = [...this.opportunities].sort((a, b) => {
      const scoreA = a.value / (a.cost + 0.1);
      const scoreB = b.value / (b.cost + 0.1);
      return scoreB - scoreA;
    });

    const topOpps = sortedOpps.slice(0, this.config.maxActionsPerCycle);

    for (const opp of topOpps) {
      await this.seizeOpportunity(opp);
    }

    // 清空已处理的机会
    this.opportunities = [];
  }

  /**
   * 主动利用单个机会
   */
  private async seizeOpportunity(opportunity: Opportunity) {
    console.log(`\n✨ [主动行动] 利用机会: ${opportunity.description}`);

    // 生成行动建议
    const context = `项目路径: ${this.projectPath}`;
    const action = await this.llm.generateImprovementSuggestion(opportunity, context);

    if (!action || action.type === 'skip') {
      console.log("⏭️  跳过此机会");
      return;
    }

    // 执行行动
    await this.executeAction(action);
    this.statistics.actionsExecuted++;
  }

  /**
   * 执行行动
   */
  private async executeAction(action: ProactiveAction) {
    console.log(`🎬 执行行动: ${action.description}`);
    console.log(`💭 原因: ${action.reasoning}`);

    switch (action.type) {
      case 'write_improvement_report':
        await this.writeImprovementReport(action);
        break;

      case 'create_todo':
        await this.createTodoFile(action);
        break;

      case 'improve_code':
        await this.improveCode(action);
        break;

      default:
        console.log(`⚠️ 未知行动类型: ${action.type}`);
    }
  }

  /**
   * 写入改进报告
   */
  private async writeImprovementReport(action: ProactiveAction) {
    const reportPath = path.join(this.projectPath, 'IMPROVEMENT_REPORT.md');
    const timestamp = new Date().toISOString();

    let content = '';
    if (fs.existsSync(reportPath)) {
      content = fs.readFileSync(reportPath, 'utf-8');
    } else {
      content = '# 代码改进报告\n\n';
    }

    content += `## ${timestamp}\n\n`;
    content += `**机会**: ${action.parameters.opportunity?.description || '未知'}\n\n`;
    content += `**建议**:\n`;

    const suggestions = action.parameters.suggestions || [];
    for (const suggestion of suggestions) {
      content += `- ${suggestion}\n`;
    }
    content += '\n---\n\n';

    fs.writeFileSync(reportPath, content, 'utf-8');
    console.log(`✅ 已写入改进报告: ${reportPath}`);
  }

  /**
   * 创建TODO文件
   */
  private async createTodoFile(action: ProactiveAction) {
    const todoPath = path.join(this.projectPath, 'TODO.md');

    let content = '';
    if (fs.existsSync(todoPath)) {
      content = fs.readFileSync(todoPath, 'utf-8');
    } else {
      content = '# TODO 列表\n\n';
    }

    content += `- [ ] ${action.description}\n`;

    fs.writeFileSync(todoPath, content, 'utf-8');
    console.log(`✅ 已添加TODO项: ${todoPath}`);
  }

  /**
   * 改进代码（示例）
   */
  private async improveCode(action: ProactiveAction) {
    console.log(`📝 代码改进: ${action.description}`);
    // 这里可以实现实际的代码改进逻辑
    // 为了安全起见，这里只记录改进建议
    await this.writeImprovementReport(action);
  }

  /**
   * 为预测做准备
   */
  private async prepareForPrediction(prediction: Prediction) {
    console.log(`🎯 为预测做准备: ${prediction.description}`);

    // 创建准备报告
    const reportPath = path.join(this.projectPath, 'PREDICTIONS.md');

    let content = '';
    if (fs.existsSync(reportPath)) {
      content = fs.readFileSync(reportPath, 'utf-8');
    } else {
      content = '# 预测与准备\n\n';
    }

    content += `## ${new Date().toISOString()}\n\n`;
    content += `**预测**: ${prediction.description}\n`;
    content += `**信心度**: ${prediction.confidence.toFixed(2)}\n`;
    content += `**建议行动**:\n`;

    for (const action of prediction.requiredActions) {
      content += `- ${action}\n`;
    }
    content += '\n---\n\n';

    fs.writeFileSync(reportPath, content, 'utf-8');
    console.log(`✅ 已记录预测准备`);
  }

  /**
   * 扫描项目文件
   */
  private async scanProjectFiles(): Promise<string[]> {
    if (!fs.existsSync(this.projectPath)) {
      return [];
    }

    const files: string[] = [];
    const items = fs.readdirSync(this.projectPath);

    for (const item of items) {
      const fullPath = path.join(this.projectPath, item);
      const stat = fs.statSync(fullPath);

      if (stat.isFile() && this.shouldAnalyzeFile(item)) {
        files.push(fullPath);
      }
    }

    return files;
  }

  /**
   * 判断是否应该分析文件
   */
  private shouldAnalyzeFile(filename: string): boolean {
    const analyzed = ['.ts', '.js', '.py', '.java', '.go'];
    const ignored = ['.md', '.json', '.log', '.tmp'];

    const ext = path.extname(filename);
    return analyzed.includes(ext) && !ignored.includes(ext);
  }

  /**
   * 扫描文档缺失
   */
  private async scanForDocumentationGaps() {
    console.log("\n📚 [主动] 扫描文档缺失...");

    const files = await this.scanProjectFiles();
    let gapsFound = 0;

    for (const file of files) {
      const content = fs.readFileSync(file, 'utf-8');

      // 简单检查：文件是否有文档注释
      if (!content.includes('/**') && content.split('\n').length > 20) {
        gapsFound++;

        this.opportunities.push({
          id: `opp_doc_${Date.now()}`,
          type: 'documentation_gap',
          description: `${path.basename(file)} 缺少文档`,
          value: 0.6,
          cost: 0.3,
          discoveredAt: new Date()
        });
      }
    }

    if (gapsFound > 0) {
      console.log(`📋 发现 ${gapsFound} 个文档缺失`);
    }
  }

  /**
   * 收集项目状态
   */
  private async gatherProjectState() {
    const files = await this.scanProjectFiles();

    return {
      projectPath: this.projectPath,
      fileCount: files.length,
      totalLines: files.reduce((sum, file) => {
        const content = fs.readFileSync(file, 'utf-8');
        return sum + content.split('\n').length;
      }, 0),
      lastScanTime: new Date()
    };
  }

  /**
   * 打印统计信息
   */
  printStatistics() {
    const runtime = Date.now() - this.statistics.startTime.getTime();
    const runtimeMinutes = Math.floor(runtime / 60000);

    console.log("\n");
    console.log("============================================================");
    console.log("📊 Agent运行统计");
    console.log("============================================================");
    console.log(`⏱️  运行时长: ${runtimeMinutes} 分钟`);
    console.log(`💡 发现机会: ${this.statistics.opportunitiesFound} 个`);
    console.log(`🎬 执行行动: ${this.statistics.actionsExecuted} 次`);
    console.log(`🎯 完成目标: ${this.statistics.goalsCompleted} 个`);
    console.log(`📋 活跃目标: ${this.goals.filter(g => g.status === 'in_progress').length} 个`);
    console.log("============================================================\n");
  }

  /**
   * 停止Agent
   */
  async stop() {
    this.isRunning = false;
    console.log("\n🛑 Agent 正在停止...");
    this.printStatistics();
    console.log("✅ Agent 已停止");
  }

  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// ==================== 主入口 ====================

async function main() {
  // 定义监控的项目目录
  const targetDir = path.join(__dirname, "monitored_project");

  const agent = new ProactiveAgent(targetDir);

  // 设置退出处理
  process.on('SIGINT', async () => {
    await agent.stop();
    process.exit(0);
  });

  await agent.start();

  console.log(`\n💡 提示: 你可以在 ${targetDir} 目录下创建一些代码文件`);
  console.log("Agent 将主动扫描、分析并提出改进建议。\n");
  console.log("示例：创建 'example.ts' 文件，添加一些有TODO的代码\n");

  // 运行2分钟后显示统计并停止（演示用）
  setTimeout(async () => {
    await agent.stop();
    process.exit(0);
  }, 120000); // 2分钟
}

main().catch(console.error);
