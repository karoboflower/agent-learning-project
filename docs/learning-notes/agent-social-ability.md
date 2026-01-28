# Agent社会性（Social Ability）详解

## 📚 目录

1. [什么是社会性](#什么是社会性)
2. [社会性的核心特征](#社会性的核心特征)
3. [社会性在Agent中的体现](#社会性在agent中的体现)
4. [社会性实现模式](#社会性实现模式)
5. [代码示例](#代码示例)
6. [最佳实践](#最佳实践)
7. [常见问题](#常见问题)

---

## 什么是社会性

### 定义

**社会性（Social Ability）**是指Agent能够与其他Agent或人类进行交互、通信、协商和协作的能力。社会性Agent不是孤立工作，而是能够在多Agent系统中有效地与他人合作。

### 核心要点

1. **通信能力**：能够发送和接收消息
2. **协作能力**：能够与其他Agent共同完成任务
3. **协商能力**：能够通过协商达成共识
4. **社交协议**：遵循特定的通信协议和规范
5. **角色认知**：理解自己和他人的角色与能力

### 与其他特征的区别

| 特征 | 定义 | 关键区别 |
|------|------|----------|
| **自主性** | 独立运行和决策 | 强调个体独立性 |
| **反应性** | 对环境变化做出响应 | 关注环境感知 |
| **主动性** | 主动采取行动实现目标 | 强调主动发起 |
| **社会性** | 与他人交互和协作 | 关注多方互动 |

---

## 社会性的核心特征

### 1. 通信机制

Agent能够通过消息传递与其他Agent通信：

```typescript
// TypeScript示例
interface Message {
  id: string;
  from: string;        // 发送者ID
  to: string;          // 接收者ID
  type: string;        // 消息类型
  content: any;        // 消息内容
  timestamp: Date;
  conversationId?: string;  // 会话ID
}

class SocialAgent {
  private agentId: string;
  private mailbox: Message[] = [];

  // 发送消息
  async sendMessage(to: string, type: string, content: any): Promise<void> {
    const message: Message = {
      id: `msg_${Date.now()}`,
      from: this.agentId,
      to: to,
      type: type,
      content: content,
      timestamp: new Date()
    };

    // 通过消息总线发送
    await this.messageBus.send(message);
    console.log(`📤 发送消息给 ${to}: ${type}`);
  }

  // 接收消息
  async receiveMessage(message: Message): Promise<void> {
    this.mailbox.push(message);
    console.log(`📬 收到来自 ${message.from} 的消息: ${message.type}`);

    // 处理消息
    await this.handleMessage(message);
  }

  // 处理消息
  private async handleMessage(message: Message): Promise<void> {
    switch (message.type) {
      case 'request':
        await this.handleRequest(message);
        break;
      case 'response':
        await this.handleResponse(message);
        break;
      case 'inform':
        await this.handleInform(message);
        break;
      default:
        console.log(`⚠️ 未知消息类型: ${message.type}`);
    }
  }
}
```

```go
// Go示例
type Message struct {
    ID             string
    From           string
    To             string
    Type           string
    Content        interface{}
    Timestamp      time.Time
    ConversationID string
}

type SocialAgent struct {
    agentID    string
    mailbox    []Message
    messageBus MessageBus
}

func (a *SocialAgent) SendMessage(to, msgType string, content interface{}) error {
    message := Message{
        ID:        fmt.Sprintf("msg_%d", time.Now().UnixNano()),
        From:      a.agentID,
        To:        to,
        Type:      msgType,
        Content:   content,
        Timestamp: time.Now(),
    }

    fmt.Printf("📤 发送消息给 %s: %s\n", to, msgType)
    return a.messageBus.Send(message)
}

func (a *SocialAgent) ReceiveMessage(message Message) error {
    a.mailbox = append(a.mailbox, message)
    fmt.Printf("📬 收到来自 %s 的消息: %s\n", message.From, message.Type)
    return a.handleMessage(message)
}
```

### 2. 协作机制

Agent能够与其他Agent协作完成任务：

```typescript
// 协作示例
interface Task {
  id: string;
  description: string;
  requiredCapabilities: string[];
  status: 'pending' | 'assigned' | 'in_progress' | 'completed';
}

class CollaborativeAgent extends SocialAgent {
  private capabilities: string[] = [];
  private assignedTasks: Task[] = [];

  // 请求协作
  async requestCollaboration(task: Task): Promise<void> {
    console.log(`🤝 请求协作完成任务: ${task.description}`);

    // 1. 分析任务需求
    const requiredCapabilities = task.requiredCapabilities;

    // 2. 查找具备相应能力的Agent
    const partners = await this.findCapableAgents(requiredCapabilities);

    // 3. 向候选Agent发送协作请求
    for (const partner of partners) {
      await this.sendMessage(partner.id, 'collaboration_request', {
        task: task,
        requiredCapability: this.getRequiredCapability(partner, task)
      });
    }
  }

  // 响应协作请求
  async handleCollaborationRequest(message: Message): Promise<void> {
    const { task, requiredCapability } = message.content;

    // 1. 检查自己是否具备所需能力
    if (this.hasCapability(requiredCapability)) {
      // 2. 评估当前工作负载
      if (this.canAcceptTask(task)) {
        // 3. 接受协作请求
        await this.sendMessage(message.from, 'collaboration_accept', {
          taskId: task.id,
          capability: requiredCapability
        });

        // 4. 开始执行任务
        await this.executeTask(task);
      } else {
        // 拒绝：工作负载过高
        await this.sendMessage(message.from, 'collaboration_reject', {
          taskId: task.id,
          reason: 'workload_high'
        });
      }
    } else {
      // 拒绝：不具备所需能力
      await this.sendMessage(message.from, 'collaboration_reject', {
        taskId: task.id,
        reason: 'capability_not_match'
      });
    }
  }
}
```

### 3. 协商机制

Agent能够通过协商达成共识：

```typescript
// 协商示例
interface Proposal {
  id: string;
  proposer: string;
  content: any;
  status: 'proposed' | 'accepted' | 'rejected' | 'countered';
}

class NegotiatingAgent extends SocialAgent {
  // 发起协商
  async initiateNegotiation(partners: string[], topic: string): Promise<void> {
    console.log(`💬 发起协商: ${topic}`);

    // 1. 生成提案
    const proposal = await this.generateProposal(topic);

    // 2. 向协商伙伴发送提案
    for (const partner of partners) {
      await this.sendMessage(partner, 'proposal', proposal);
    }

    // 3. 等待响应
    await this.waitForNegotiationResponses(proposal.id);
  }

  // 处理提案
  async handleProposal(message: Message): Promise<void> {
    const proposal: Proposal = message.content;

    // 1. 评估提案
    const evaluation = await this.evaluateProposal(proposal);

    if (evaluation.acceptable) {
      // 接受提案
      await this.sendMessage(message.from, 'proposal_accept', {
        proposalId: proposal.id
      });
    } else if (evaluation.negotiable) {
      // 提出反提案
      const counterProposal = await this.generateCounterProposal(proposal, evaluation);
      await this.sendMessage(message.from, 'proposal_counter', counterProposal);
    } else {
      // 拒绝提案
      await this.sendMessage(message.from, 'proposal_reject', {
        proposalId: proposal.id,
        reason: evaluation.reason
      });
    }
  }

  // 使用LLM评估提案
  private async evaluateProposal(proposal: Proposal): Promise<any> {
    const prompt = `
      你是一个智能协商Agent。请评估以下提案：

      提案内容: ${JSON.stringify(proposal.content)}

      请判断：
      1. 是否可以接受此提案？
      2. 如果不接受，是否可以协商？
      3. 拒绝的理由是什么？

      返回JSON格式：
      {
        "acceptable": true/false,
        "negotiable": true/false,
        "reason": "理由"
      }
    `;

    const result = await this.llm.analyze(prompt);
    return result;
  }
}
```

### 4. 角色与组织

Agent在社会系统中有明确的角色和组织结构：

```typescript
// 角色与组织示例
enum AgentRole {
  COORDINATOR = 'coordinator',    // 协调者
  WORKER = 'worker',              // 工作者
  EXPERT = 'expert',              // 专家
  MONITOR = 'monitor'             // 监控者
}

interface Organization {
  id: string;
  name: string;
  members: Map<string, AgentRole>;  // agentId -> role
  hierarchy: Map<string, string>;   // agentId -> supervisorId
}

class OrganizationalAgent extends SocialAgent {
  private role: AgentRole;
  private organization: Organization;
  private supervisor?: string;
  private subordinates: string[] = [];

  // 根据角色处理任务
  async handleTaskByRole(task: Task): Promise<void> {
    switch (this.role) {
      case AgentRole.COORDINATOR:
        // 协调者：分配任务给工作者
        await this.coordinateTask(task);
        break;

      case AgentRole.WORKER:
        // 工作者：执行任务
        await this.executeTask(task);
        break;

      case AgentRole.EXPERT:
        // 专家：提供专业建议
        await this.provideExpertise(task);
        break;

      case AgentRole.MONITOR:
        // 监控者：监控进度
        await this.monitorProgress(task);
        break;
    }
  }

  // 协调者：分配任务
  async coordinateTask(task: Task): Promise<void> {
    console.log(`🎯 [协调者] 协调任务: ${task.description}`);

    // 1. 分解任务
    const subtasks = await this.decomposeTask(task);

    // 2. 为每个子任务找到合适的工作者
    for (const subtask of subtasks) {
      const worker = await this.findBestWorker(subtask);

      // 3. 分配任务
      await this.sendMessage(worker, 'task_assignment', subtask);
    }

    // 4. 监控整体进度
    await this.monitorOverallProgress(task.id);
  }

  // 向上级报告
  async reportToSupervisor(report: any): Promise<void> {
    if (this.supervisor) {
      await this.sendMessage(this.supervisor, 'progress_report', report);
    }
  }

  // 向下级发出指令
  async instructSubordinates(instruction: any): Promise<void> {
    for (const subordinate of this.subordinates) {
      await this.sendMessage(subordinate, 'instruction', instruction);
    }
  }
}
```

---

## 社会性在Agent中的体现

### 1. 多Agent系统（MAS）

多个Agent协作形成系统：

```typescript
// 多Agent系统示例
class MultiAgentSystem {
  private agents: Map<string, SocialAgent> = new Map();
  private messageBus: MessageBus;

  // 添加Agent到系统
  addAgent(agent: SocialAgent): void {
    this.agents.set(agent.getId(), agent);
    agent.setMessageBus(this.messageBus);
  }

  // 启动系统
  async start(): Promise<void> {
    console.log(`🚀 启动多Agent系统，共 ${this.agents.size} 个Agent`);

    // 启动所有Agent
    const startPromises = Array.from(this.agents.values()).map(agent =>
      agent.start()
    );

    await Promise.all(startPromises);
  }

  // 广播消息
  async broadcast(from: string, type: string, content: any): Promise<void> {
    for (const [agentId, agent] of this.agents) {
      if (agentId !== from) {
        await agent.receiveMessage({
          id: `msg_${Date.now()}`,
          from: from,
          to: agentId,
          type: type,
          content: content,
          timestamp: new Date()
        });
      }
    }
  }
}
```

### 2. 通信协议

Agent使用标准化的通信协议：

```typescript
// FIPA ACL (Agent Communication Language) 样式的协议
enum PerformativeType {
  REQUEST = 'request',           // 请求
  INFORM = 'inform',             // 通知
  QUERY = 'query',               // 查询
  PROPOSE = 'propose',           // 提议
  ACCEPT = 'accept',             // 接受
  REJECT = 'reject',             // 拒绝
  CONFIRM = 'confirm',           // 确认
  AGREE = 'agree',               // 同意
  REFUSE = 'refuse'              // 拒绝
}

interface ACLMessage {
  performative: PerformativeType;
  sender: string;
  receiver: string;
  content: any;
  language?: string;
  ontology?: string;
  conversationId?: string;
  replyWith?: string;
  inReplyTo?: string;
}

class ACLAgent extends SocialAgent {
  // 发送ACL消息
  async sendACL(message: ACLMessage): Promise<void> {
    console.log(`📨 [${message.performative}] ${message.sender} -> ${message.receiver}`);

    await this.messageBus.send({
      id: `msg_${Date.now()}`,
      from: message.sender,
      to: message.receiver,
      type: message.performative,
      content: message.content,
      timestamp: new Date(),
      conversationId: message.conversationId
    });
  }

  // 请求-响应模式
  async requestResponse(receiver: string, request: any): Promise<any> {
    const conversationId = `conv_${Date.now()}`;

    // 发送请求
    await this.sendACL({
      performative: PerformativeType.REQUEST,
      sender: this.agentId,
      receiver: receiver,
      content: request,
      conversationId: conversationId
    });

    // 等待响应
    const response = await this.waitForResponse(conversationId);
    return response;
  }
}
```

### 3. 团队协作模式

多个Agent组成团队完成复杂任务：

```typescript
// 团队协作示例
interface Team {
  id: string;
  name: string;
  members: string[];
  leader: string;
  goal: string;
}

class TeamAgent extends SocialAgent {
  private team?: Team;
  private isLeader: boolean = false;

  // 组建团队
  async formTeam(goal: string, requiredCapabilities: string[]): Promise<Team> {
    console.log(`👥 组建团队，目标: ${goal}`);

    // 1. 寻找具备所需能力的Agent
    const candidates = await this.findCapableAgents(requiredCapabilities);

    // 2. 向候选Agent发送团队邀请
    const acceptedMembers: string[] = [this.agentId];

    for (const candidate of candidates) {
      const accepted = await this.inviteToTeam(candidate, goal);
      if (accepted) {
        acceptedMembers.push(candidate.id);
      }
    }

    // 3. 创建团队
    const team: Team = {
      id: `team_${Date.now()}`,
      name: `Team for ${goal}`,
      members: acceptedMembers,
      leader: this.agentId,
      goal: goal
    };

    this.team = team;
    this.isLeader = true;

    // 4. 通知所有成员
    await this.notifyTeamMembers(team);

    return team;
  }

  // 团队协作执行任务
  async collaborateOnTask(task: Task): Promise<void> {
    if (!this.team) {
      throw new Error("未加入任何团队");
    }

    if (this.isLeader) {
      // 领导者：协调任务
      await this.coordinateTeamTask(task);
    } else {
      // 成员：执行分配的任务
      await this.waitForTaskAssignment();
    }
  }

  // 协调团队任务
  private async coordinateTeamTask(task: Task): Promise<void> {
    console.log(`🎯 [团队领导] 协调团队任务: ${task.description}`);

    // 1. 使用LLM分解任务
    const subtasks = await this.decomposeTaskWithLLM(task);

    // 2. 分配给团队成员
    for (let i = 0; i < subtasks.length; i++) {
      const member = this.team!.members[i % this.team!.members.length];
      if (member !== this.agentId) {
        await this.sendMessage(member, 'task_assignment', subtasks[i]);
      }
    }

    // 3. 收集结果
    const results = await this.collectResults(subtasks.length);

    // 4. 整合结果
    const finalResult = await this.integrateResults(results);

    console.log(`✅ [团队领导] 任务完成: ${task.description}`);
  }

  // 使用LLM分解任务
  private async decomposeTaskWithLLM(task: Task): Promise<Task[]> {
    const prompt = `
      你是一个团队协调Agent。请将以下任务分解为子任务：

      任务: ${task.description}
      团队成员数: ${this.team!.members.length}

      请返回JSON格式的子任务列表：
      [
        {
          "id": "subtask_1",
          "description": "子任务描述",
          "assignedTo": "成员ID"
        }
      ]
    `;

    const result = await this.llm.analyze(prompt);
    return result;
  }
}
```

---

## 社会性实现模式

### 模式1：主从模式（Master-Slave）

```typescript
// 主从模式
class MasterAgent extends SocialAgent {
  private slaves: string[] = [];

  async assignTask(task: Task): Promise<void> {
    // 主Agent分配任务给从Agent
    const slave = this.selectSlave();
    await this.sendMessage(slave, 'task', task);
  }

  async collectResults(): Promise<any[]> {
    // 收集所有从Agent的结果
    const results = [];
    for (const slave of this.slaves) {
      const result = await this.requestResult(slave);
      results.push(result);
    }
    return results;
  }
}

class SlaveAgent extends SocialAgent {
  async handleTask(task: Task): Promise<void> {
    // 从Agent执行任务
    const result = await this.executeTask(task);

    // 向主Agent报告结果
    await this.sendMessage(this.masterId, 'result', result);
  }
}
```

### 模式2：合约网络（Contract Net）

```typescript
// 合约网络模式
class ContractNetAgent extends SocialAgent {
  // 管理者：发布任务公告
  async announceTask(task: Task): Promise<void> {
    console.log(`📢 发布任务公告: ${task.description}`);

    // 广播任务到所有Agent
    await this.broadcast('task_announcement', task);

    // 收集投标
    const bids = await this.collectBids(task.id);

    // 评估投标并选择最佳承包者
    const winner = await this.selectBestBidder(bids);

    // 授予合约
    await this.awardContract(winner, task);
  }

  // 承包者：投标
  async bid(task: Task): Promise<void> {
    // 1. 评估任务
    const canDo = await this.canExecuteTask(task);

    if (canDo) {
      // 2. 计算成本和时间
      const cost = await this.estimateCost(task);
      const time = await this.estimateTime(task);

      // 3. 提交投标
      await this.sendMessage(task.announcerId, 'bid', {
        taskId: task.id,
        bidderId: this.agentId,
        cost: cost,
        time: time,
        confidence: this.capabilities.match(task.requiredCapabilities)
      });
    }
  }
}
```

### 模式3：黑板模式（Blackboard）

```typescript
// 黑板模式
class BlackboardSystem {
  private blackboard: Map<string, any> = new Map();
  private agents: SocialAgent[] = [];

  // Agent向黑板写入信息
  async writeToBlackboard(key: string, value: any, author: string): Promise<void> {
    this.blackboard.set(key, {
      value: value,
      author: author,
      timestamp: new Date()
    });

    // 通知其他Agent
    await this.notifyAgents(key);
  }

  // Agent从黑板读取信息
  readFromBlackboard(key: string): any {
    return this.blackboard.get(key);
  }

  // 获取所有信息
  getAllInformation(): Map<string, any> {
    return new Map(this.blackboard);
  }
}

class BlackboardAgent extends SocialAgent {
  private blackboard: BlackboardSystem;

  // 贡献知识到黑板
  async contributeKnowledge(knowledge: any): Promise<void> {
    await this.blackboard.writeToBlackboard(
      `knowledge_${this.agentId}_${Date.now()}`,
      knowledge,
      this.agentId
    );
  }

  // 从黑板获取知识
  async getKnowledge(): Promise<any[]> {
    const allInfo = this.blackboard.getAllInformation();
    return Array.from(allInfo.values());
  }

  // 基于黑板信息做决策
  async makeDecisionBasedOnBlackboard(): Promise<void> {
    const knowledge = await this.getKnowledge();

    // 使用LLM分析黑板信息
    const decision = await this.llm.analyze({
      context: knowledge,
      goal: this.currentGoal
    });

    await this.executeDecision(decision);
  }
}
```

---

## 代码示例

### 示例1：代码审查团队

```typescript
// 代码审查多Agent系统
class CodeReviewTeam {
  private analyzer: CodeAnalyzerAgent;
  private reviewer: CodeReviewerAgent;
  private coordinator: CoordinatorAgent;

  async reviewCode(filePath: string): Promise<void> {
    // 1. 协调者启动审查流程
    await this.coordinator.initiateReview(filePath);

    // 2. 分析Agent分析代码
    const analysis = await this.analyzer.analyzeCode(filePath);

    // 3. 审查Agent基于分析提出建议
    const suggestions = await this.reviewer.generateSuggestions(analysis);

    // 4. 协调者整合结果
    const report = await this.coordinator.generateReport(analysis, suggestions);

    console.log("✅ 代码审查完成", report);
  }
}
```

### 示例2：协商系统

```typescript
// 价格协商Agent
class PriceNegotiationAgent extends SocialAgent {
  async negotiatePrice(sellerId: string, item: string, maxPrice: number): Promise<number> {
    let currentOffer = maxPrice * 0.7; // 从70%开始
    let round = 0;
    const maxRounds = 5;

    while (round < maxRounds) {
      // 发送报价
      await this.sendMessage(sellerId, 'price_offer', {
        item: item,
        price: currentOffer,
        round: round
      });

      // 等待卖家响应
      const response = await this.waitForResponse(sellerId);

      if (response.type === 'accept') {
        console.log(`✅ 协商成功，成交价: ${currentOffer}`);
        return currentOffer;
      } else if (response.type === 'counter_offer') {
        // 使用LLM评估反提案
        const shouldAccept = await this.evaluateCounterOffer(
          response.content.price,
          maxPrice,
          round
        );

        if (shouldAccept) {
          await this.sendMessage(sellerId, 'accept', {});
          return response.content.price;
        } else {
          // 提高报价
          currentOffer = (currentOffer + response.content.price) / 2;
        }
      } else if (response.type === 'reject') {
        break;
      }

      round++;
    }

    throw new Error("协商失败");
  }

  async evaluateCounterOffer(price: number, maxPrice: number, round: number): Promise<boolean> {
    // 使用LLM做决策
    const prompt = `
      你是一个协商Agent。卖家提出反报价 ${price}，你的最高预算是 ${maxPrice}。
      当前是第 ${round} 轮协商。

      请判断是否应该接受此报价？返回JSON：
      {
        "accept": true/false,
        "reasoning": "理由"
      }
    `;

    const result = await this.llm.analyze(prompt);
    return result.accept;
  }
}
```

---

## 最佳实践

### 1. 清晰的通信协议

```typescript
// ✅ 好的实践：标准化消息格式
interface StandardMessage {
  header: {
    id: string;
    from: string;
    to: string;
    type: string;
    timestamp: Date;
  };
  body: {
    content: any;
    metadata?: any;
  };
}

// ❌ 不好的实践：随意的消息格式
const badMessage = {
  sender: "agent1",
  msg: "some content",
  // 缺少关键字段
};
```

### 2. 异步消息处理

```typescript
// ✅ 好的实践：异步处理消息
class AsyncSocialAgent extends SocialAgent {
  private messageQueue: Message[] = [];

  async processMessages(): Promise<void> {
    while (this.isRunning) {
      if (this.messageQueue.length > 0) {
        const message = this.messageQueue.shift();
        await this.handleMessage(message!);
      }
      await this.sleep(100);
    }
  }

  async receiveMessage(message: Message): Promise<void> {
    // 立即返回，稍后处理
    this.messageQueue.push(message);
  }
}
```

### 3. 超时和重试机制

```typescript
// ✅ 好的实践：处理超时
class ReliableSocialAgent extends SocialAgent {
  async sendWithTimeout(
    to: string,
    type: string,
    content: any,
    timeout: number = 5000
  ): Promise<any> {
    const conversationId = `conv_${Date.now()}`;

    // 发送消息
    await this.sendMessage(to, type, content);

    // 等待响应，带超时
    return await Promise.race([
      this.waitForResponse(conversationId),
      this.timeout(timeout)
    ]);
  }

  private timeout(ms: number): Promise<never> {
    return new Promise((_, reject) =>
      setTimeout(() => reject(new Error("Timeout")), ms)
    );
  }
}
```

---

## 常见问题

### Q1: 社会性和协作有什么区别？

**A**: 社会性是更广泛的概念，包括：
- 通信
- 协作
- 协商
- 竞争
- 社交互动

协作只是社会性的一个方面。

### Q2: 如何处理Agent之间的冲突？

**A**: 通过协商和仲裁机制：

```typescript
class ConflictResolution {
  async resolveConflict(agent1: string, agent2: string, issue: any): Promise<any> {
    // 1. 收集双方意见
    const opinion1 = await this.getOpinion(agent1, issue);
    const opinion2 = await this.getOpinion(agent2, issue);

    // 2. 尝试协商
    const agreement = await this.negotiate(agent1, agent2, issue);

    if (agreement) {
      return agreement;
    }

    // 3. 仲裁
    return await this.arbitrate(opinion1, opinion2, issue);
  }
}
```

### Q3: 如何保证消息的可靠传递？

**A**: 使用确认和重传机制：

```typescript
class ReliableMessaging {
  async sendReliable(message: Message): Promise<void> {
    let attempts = 0;
    const maxAttempts = 3;

    while (attempts < maxAttempts) {
      try {
        await this.send(message);

        // 等待确认
        const ack = await this.waitForAck(message.id, 5000);

        if (ack) {
          return; // 成功
        }
      } catch (error) {
        attempts++;
        await this.sleep(1000 * attempts); // 退避
      }
    }

    throw new Error("消息发送失败");
  }
}
```

---

## 总结

Agent的社会性使其能够：

1. **通信交互**：与其他Agent和人类有效沟通
2. **协作共赢**：与他人合作完成复杂任务
3. **协商达成共识**：通过协商解决分歧
4. **组织协调**：在团队和组织中有效工作

社会性是构建多Agent系统的基础，让Agent从独立个体变成协作团队。

---

## 参考资料

- [FIPA ACL](http://www.fipa.org/specs/fipa00061/)
- [Multi-Agent Systems](https://en.wikipedia.org/wiki/Multi-agent_system)
- [Contract Net Protocol](https://en.wikipedia.org/wiki/Contract_Net_Protocol)
- [Blackboard System](https://en.wikipedia.org/wiki/Blackboard_system)

---

**下一步学习**：
- [x] 学习Agent的自主性（Autonomy） - [查看笔记](./agent-autonomy.md)
- [x] 学习Agent的反应性（Reactivity） - [查看笔记](./agent-reactivity.md)
- [x] 学习Agent的主动性（Proactiveness） - [查看笔记](./agent-proactiveness.md)
- [x] 学习Agent的社会性（Social Ability） - 当前文档
- [ ] 实践构建完整的多Agent系统
