# Agent反应性（Reactivity）详解

## 📚 目录

1. [什么是反应性](#什么是反应性)
2. [反应性的核心特征](#反应性的核心特征)
3. [反应性在Agent中的体现](#反应性在agent中的体现)
4. [反应性实现模式](#反应性实现模式)
5. [代码示例](#代码示例)
6. [反应性与其他特征的关系](#反应性与其他特征的关系)
7. [最佳实践](#最佳实践)
8. [常见问题](#常见问题)

---

## 什么是反应性

### 定义

**反应性（Reactivity）**是指Agent能够**感知环境变化**并**及时做出响应**的能力。反应性Agent能够实时监控环境状态，并根据环境的变化自动调整行为。

### 核心要点

1. **环境感知**：持续监控环境状态
2. **快速响应**：对环境变化做出及时反应
3. **适应性**：根据不同环境变化采取不同响应策略
4. **实时性**：强调响应的及时性和准确性

### 与其他特征的区别

| 特征 | 定义 | 关键区别 |
|------|------|----------|
| **反应性** | 对环境变化做出响应 | **被动的**，由外部事件触发 |
| **自主性** | 独立运行和决策 | 主动的，内部驱动 |
| **主动性** | 主动采取行动 | 主动的，目标驱动 |
| **社会性** | 与其他Agent协作 | 交互性，关注沟通 |

### 示意图

```
环境变化 → 感知 → 分析 → 决策 → 响应动作
   ↑                              ↓
   └──────────── 反馈循环 ─────────┘
```

---

## 反应性的核心特征

### 1. 环境监控

Agent需要持续监控环境的各种变化：

```typescript
// TypeScript示例：环境监控
interface EnvironmentState {
  temperature: number;
  resources: Resource[];
  obstacles: Obstacle[];
  threats: Threat[];
  opportunities: Opportunity[];
  timestamp: Date;
}

class ReactiveAgent {
  private environmentState: EnvironmentState;
  private sensors: Sensor[] = [];

  // 持续监控环境
  async monitorEnvironment(): Promise<void> {
    while (this.isRunning) {
      // 1. 从传感器获取数据
      const sensorData = await this.readSensors();

      // 2. 更新环境状态
      const newState = this.updateEnvironmentState(sensorData);

      // 3. 检测变化
      const changes = this.detectChanges(this.environmentState, newState);

      // 4. 如果有变化，触发响应
      if (changes.length > 0) {
        await this.reactToChanges(changes);
      }

      // 5. 更新状态
      this.environmentState = newState;

      // 6. 短暂等待，避免过度占用CPU
      await this.sleep(100); // 100ms监控周期
    }
  }

  private async readSensors(): Promise<SensorData[]> {
    return await Promise.all(
      this.sensors.map(sensor => sensor.read())
    );
  }

  private detectChanges(oldState: EnvironmentState, newState: EnvironmentState): Change[] {
    const changes: Change[] = [];

    // 检测温度变化
    if (Math.abs(newState.temperature - oldState.temperature) > 5) {
      changes.push({
        type: 'temperature',
        oldValue: oldState.temperature,
        newValue: newState.temperature
      });
    }

    // 检测新的威胁
    const newThreats = newState.threats.filter(
      t => !oldState.threats.some(ot => ot.id === t.id)
    );
    if (newThreats.length > 0) {
      changes.push({
        type: 'new_threats',
        value: newThreats
      });
    }

    // 检测资源变化
    // ...

    return changes;
  }
}
```

```go
// Go示例：环境监控
type EnvironmentState struct {
    Temperature  float64
    Resources    []Resource
    Obstacles    []Obstacle
    Threats      []Threat
    Opportunities []Opportunity
    Timestamp    time.Time
}

type ReactiveAgent struct {
    environmentState EnvironmentState
    sensors         []Sensor
    isRunning       bool
}

func (a *ReactiveAgent) MonitorEnvironment() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for a.isRunning {
        select {
        case <-ticker.C:
            // 1. 读取传感器
            sensorData := a.readSensors()

            // 2. 更新环境状态
            newState := a.updateEnvironmentState(sensorData)

            // 3. 检测变化
            changes := a.detectChanges(a.environmentState, newState)

            // 4. 响应变化
            if len(changes) > 0 {
                a.reactToChanges(changes)
            }

            // 5. 更新状态
            a.environmentState = newState
        }
    }
}

func (a *ReactiveAgent) detectChanges(oldState, newState EnvironmentState) []Change {
    changes := []Change{}

    // 检测温度变化
    if math.Abs(newState.Temperature - oldState.Temperature) > 5 {
        changes = append(changes, Change{
            Type:     "temperature",
            OldValue: oldState.Temperature,
            NewValue: newState.Temperature,
        })
    }

    // 检测新威胁
    // ...

    return changes
}
```

### 2. 事件驱动响应

基于事件的响应机制：

```typescript
// 事件驱动响应示例
import { EventEmitter } from 'events';

class EventDrivenReactiveAgent extends EventEmitter {
  private handlers: Map<string, EventHandler[]> = new Map();

  constructor() {
    super();
    this.setupEventHandlers();
  }

  private setupEventHandlers() {
    // 注册温度变化处理器
    this.on('temperature:high', async (temp: number) => {
      console.log(`🔥 高温警告: ${temp}°C`);
      await this.activateCooling();
    });

    this.on('temperature:low', async (temp: number) => {
      console.log(`❄️ 低温警告: ${temp}°C`);
      await this.activateHeating();
    });

    // 注册威胁响应处理器
    this.on('threat:detected', async (threat: Threat) => {
      console.log(`⚠️ 检测到威胁: ${threat.type}`);
      await this.handleThreat(threat);
    });

    // 注册机会响应处理器
    this.on('opportunity:detected', async (opportunity: Opportunity) => {
      console.log(`💡 发现机会: ${opportunity.type}`);
      await this.seizeOpportunity(opportunity);
    });

    // 注册资源变化处理器
    this.on('resource:low', async (resource: Resource) => {
      console.log(`📦 资源不足: ${resource.name}`);
      await this.replenishResource(resource);
    });

    this.on('resource:available', async (resource: Resource) => {
      console.log(`✅ 资源可用: ${resource.name}`);
      await this.utilizéResource(resource);
    });
  }

  // 分析环境变化并触发相应事件
  async analyzeAndReact(change: Change) {
    switch (change.type) {
      case 'temperature':
        if (change.newValue > 30) {
          this.emit('temperature:high', change.newValue);
        } else if (change.newValue < 10) {
          this.emit('temperature:low', change.newValue);
        }
        break;

      case 'new_threats':
        for (const threat of change.value) {
          this.emit('threat:detected', threat);
        }
        break;

      case 'new_opportunities':
        for (const opportunity of change.value) {
          this.emit('opportunity:detected', opportunity);
        }
        break;

      case 'resource_change':
        if (change.value.level < 20) {
          this.emit('resource:low', change.value);
        }
        break;
    }
  }
}
```

### 3. 优先级响应

根据事件的紧急程度调整响应优先级：

```typescript
// 优先级响应示例
enum Priority {
  CRITICAL = 0,  // 立即响应
  HIGH = 1,      // 高优先级
  MEDIUM = 2,    // 中优先级
  LOW = 3        // 低优先级
}

interface ReactiveEvent {
  type: string;
  data: any;
  priority: Priority;
  timestamp: Date;
}

class PriorityReactiveAgent {
  private eventQueue: ReactiveEvent[] = [];
  private isProcessing = false;

  // 添加事件到队列
  addEvent(event: ReactiveEvent) {
    // 插入到正确的位置（按优先级排序）
    const insertIndex = this.eventQueue.findIndex(
      e => e.priority > event.priority
    );

    if (insertIndex === -1) {
      this.eventQueue.push(event);
    } else {
      this.eventQueue.splice(insertIndex, 0, event);
    }

    // 如果是关键事件，立即中断当前任务
    if (event.priority === Priority.CRITICAL) {
      this.interruptAndProcess(event);
    } else if (!this.isProcessing) {
      this.processQueue();
    }
  }

  // 处理事件队列
  private async processQueue() {
    if (this.isProcessing || this.eventQueue.length === 0) {
      return;
    }

    this.isProcessing = true;

    while (this.eventQueue.length > 0) {
      const event = this.eventQueue.shift()!;

      try {
        await this.handleEvent(event);
      } catch (error) {
        console.error(`处理事件失败: ${event.type}`, error);
      }
    }

    this.isProcessing = false;
  }

  // 中断当前任务处理关键事件
  private async interruptAndProcess(event: ReactiveEvent) {
    console.log(`🚨 关键事件，中断当前任务: ${event.type}`);

    // 暂停当前任务
    await this.pauseCurrentTask();

    // 立即处理关键事件
    await this.handleEvent(event);

    // 恢复之前的任务
    await this.resumeCurrentTask();
  }

  private async handleEvent(event: ReactiveEvent): Promise<void> {
    console.log(`处理事件 [${Priority[event.priority]}]: ${event.type}`);

    // 根据事件类型执行相应操作
    switch (event.type) {
      case 'system:shutdown':
        await this.emergencyShutdown();
        break;
      case 'threat:critical':
        await this.handleCriticalThreat(event.data);
        break;
      case 'resource:depleted':
        await this.handleResourceDepletion(event.data);
        break;
      // ... 其他事件类型
    }
  }
}
```

---

## 反应性在Agent中的体现

### 1. 条件-动作规则（Condition-Action Rules）

最简单的反应性实现：

```typescript
// 条件-动作规则实现
interface Rule {
  condition: (state: EnvironmentState) => boolean;
  action: (state: EnvironmentState) => Promise<void>;
  priority: number;
}

class RuleBasedReactiveAgent {
  private rules: Rule[] = [];

  constructor() {
    this.defineRules();
  }

  private defineRules() {
    // 规则1: 温度过高 → 启动冷却
    this.rules.push({
      condition: (state) => state.temperature > 30,
      action: async (state) => {
        console.log(`温度过高(${state.temperature}°C)，启动冷却系统`);
        await this.activateCooling();
      },
      priority: 1
    });

    // 规则2: 检测到威胁 → 采取防御措施
    this.rules.push({
      condition: (state) => state.threats.length > 0,
      action: async (state) => {
        console.log(`检测到 ${state.threats.length} 个威胁`);
        for (const threat of state.threats) {
          await this.defendAgainst(threat);
        }
      },
      priority: 0 // 最高优先级
    });

    // 规则3: 资源不足 → 寻找资源
    this.rules.push({
      condition: (state) => state.resources.some(r => r.level < 20),
      action: async (state) => {
        const lowResources = state.resources.filter(r => r.level < 20);
        console.log(`资源不足: ${lowResources.map(r => r.name).join(', ')}`);
        await this.seekResources(lowResources);
      },
      priority: 2
    });

    // 规则4: 发现机会 → 抓住机会
    this.rules.push({
      condition: (state) => state.opportunities.length > 0,
      action: async (state) => {
        console.log(`发现 ${state.opportunities.length} 个机会`);
        for (const opportunity of state.opportunities) {
          await this.seizeOpportunity(opportunity);
        }
      },
      priority: 3
    });

    // 按优先级排序
    this.rules.sort((a, b) => a.priority - b.priority);
  }

  // 评估规则并执行响应
  async evaluate(state: EnvironmentState) {
    for (const rule of this.rules) {
      if (rule.condition(state)) {
        await rule.action(state);
      }
    }
  }
}
```

### 2. 反应式架构（Reactive Architecture）

使用观察者模式实现反应性：

```typescript
// 观察者模式实现反应性
interface Observer {
  update(subject: Subject, event: any): void;
}

interface Subject {
  attach(observer: Observer): void;
  detach(observer: Observer): void;
  notify(event: any): void;
}

class Environment implements Subject {
  private observers: Observer[] = [];
  private state: EnvironmentState;

  attach(observer: Observer): void {
    if (!this.observers.includes(observer)) {
      this.observers.push(observer);
    }
  }

  detach(observer: Observer): void {
    const index = this.observers.indexOf(observer);
    if (index > -1) {
      this.observers.splice(index, 1);
    }
  }

  notify(event: any): void {
    for (const observer of this.observers) {
      observer.update(this, event);
    }
  }

  // 状态变化时自动通知观察者
  setState(newState: EnvironmentState) {
    const changes = this.detectChanges(this.state, newState);
    this.state = newState;

    if (changes.length > 0) {
      this.notify({ type: 'state:changed', changes });
    }
  }

  private detectChanges(oldState: EnvironmentState, newState: EnvironmentState): Change[] {
    // 检测变化逻辑
    return [];
  }
}

class ReactiveAgentObserver implements Observer {
  private agent: ReactiveAgent;

  constructor(agent: ReactiveAgent) {
    this.agent = agent;
  }

  update(subject: Subject, event: any): void {
    console.log('环境发生变化，Agent做出响应...');

    // 根据事件类型做出不同响应
    if (event.type === 'state:changed') {
      this.agent.reactToChanges(event.changes);
    }
  }
}

// 使用示例
const environment = new Environment();
const agent = new ReactiveAgent();
const observer = new ReactiveAgentObserver(agent);

environment.attach(observer);
```

### 3. 反应式传感器系统

```typescript
// 传感器系统实现
interface Sensor {
  id: string;
  type: string;
  read(): Promise<SensorReading>;
}

interface SensorReading {
  sensorId: string;
  value: any;
  timestamp: Date;
  metadata?: any;
}

class TemperatureSensor implements Sensor {
  id: string;
  type = 'temperature';

  constructor(id: string) {
    this.id = id;
  }

  async read(): Promise<SensorReading> {
    // 模拟读取温度
    const value = 20 + Math.random() * 15;

    return {
      sensorId: this.id,
      value,
      timestamp: new Date()
    };
  }
}

class ThreatDetectionSensor implements Sensor {
  id: string;
  type = 'threat_detection';

  constructor(id: string) {
    this.id = id;
  }

  async read(): Promise<SensorReading> {
    // 模拟威胁检测
    const threats = [];

    if (Math.random() > 0.9) {
      threats.push({
        id: `threat_${Date.now()}`,
        type: 'obstacle',
        severity: Math.random(),
        location: { x: Math.random() * 100, y: Math.random() * 100 }
      });
    }

    return {
      sensorId: this.id,
      value: threats,
      timestamp: new Date()
    };
  }
}

class SensorBasedReactiveAgent {
  private sensors: Sensor[] = [];
  private readings: Map<string, SensorReading[]> = new Map();

  addSensor(sensor: Sensor) {
    this.sensors.push(sensor);
    this.readings.set(sensor.id, []);
  }

  async monitorSensors() {
    while (this.isRunning) {
      // 并行读取所有传感器
      const readings = await Promise.all(
        this.sensors.map(sensor => sensor.read())
      );

      // 处理每个传感器读数
      for (const reading of readings) {
        await this.processSensorReading(reading);
      }

      await this.sleep(100);
    }
  }

  private async processSensorReading(reading: SensorReading) {
    // 存储读数
    const history = this.readings.get(reading.sensorId)!;
    history.push(reading);

    // 只保留最近100个读数
    if (history.length > 100) {
      history.shift();
    }

    // 分析读数并做出响应
    const sensor = this.sensors.find(s => s.id === reading.sensorId)!;

    switch (sensor.type) {
      case 'temperature':
        await this.reactToTemperature(reading.value);
        break;
      case 'threat_detection':
        if (reading.value.length > 0) {
          await this.reactToThreats(reading.value);
        }
        break;
    }
  }

  private async reactToTemperature(temp: number) {
    if (temp > 30) {
      console.log(`🔥 高温响应: ${temp.toFixed(1)}°C`);
      await this.activateCooling();
    } else if (temp < 15) {
      console.log(`❄️ 低温响应: ${temp.toFixed(1)}°C`);
      await this.activateHeating();
    }
  }

  private async reactToThreats(threats: any[]) {
    console.log(`⚠️ 威胁响应: 检测到 ${threats.length} 个威胁`);

    for (const threat of threats) {
      if (threat.severity > 0.7) {
        await this.avoidThreat(threat);
      } else {
        await this.monitorThreat(threat);
      }
    }
  }
}
```

---

## 反应性实现模式

### 模式1：轮询模式（Polling Pattern）

定期检查环境状态：

```typescript
// 轮询模式实现
class PollingReactiveAgent {
  private pollingInterval = 1000; // 1秒
  private isRunning = false;

  async start() {
    this.isRunning = true;

    while (this.isRunning) {
      // 1. 读取当前环境状态
      const state = await this.readEnvironment();

      // 2. 评估是否需要响应
      const needsReaction = this.evaluateState(state);

      // 3. 如果需要，执行响应
      if (needsReaction) {
        await this.react(state);
      }

      // 4. 等待下一次轮询
      await this.sleep(this.pollingInterval);
    }
  }

  private async readEnvironment(): Promise<EnvironmentState> {
    // 读取环境状态
    return {
      temperature: await this.readTemperature(),
      resources: await this.scanResources(),
      threats: await this.detectThreats(),
      opportunities: await this.findOpportunities(),
      timestamp: new Date()
    };
  }

  private evaluateState(state: EnvironmentState): boolean {
    // 评估是否需要响应
    return (
      state.temperature > 30 ||
      state.temperature < 10 ||
      state.threats.length > 0 ||
      state.resources.some(r => r.level < 20)
    );
  }
}
```

### 模式2：发布-订阅模式（Pub-Sub Pattern）

```typescript
// 发布-订阅模式实现
interface Message {
  topic: string;
  data: any;
  timestamp: Date;
}

class MessageBroker {
  private subscribers: Map<string, Set<Subscriber>> = new Map();

  subscribe(topic: string, subscriber: Subscriber) {
    if (!this.subscribers.has(topic)) {
      this.subscribers.set(topic, new Set());
    }
    this.subscribers.get(topic)!.add(subscriber);
  }

  unsubscribe(topic: string, subscriber: Subscriber) {
    this.subscribers.get(topic)?.delete(subscriber);
  }

  publish(message: Message) {
    const subscribers = this.subscribers.get(message.topic);

    if (subscribers) {
      for (const subscriber of subscribers) {
        subscriber.onMessage(message);
      }
    }
  }
}

interface Subscriber {
  onMessage(message: Message): void;
}

class PubSubReactiveAgent implements Subscriber {
  private broker: MessageBroker;

  constructor(broker: MessageBroker) {
    this.broker = broker;

    // 订阅感兴趣的主题
    this.broker.subscribe('environment:temperature', this);
    this.broker.subscribe('environment:threat', this);
    this.broker.subscribe('environment:resource', this);
  }

  onMessage(message: Message): void {
    console.log(`收到消息 [${message.topic}]:`, message.data);

    // 根据主题做出响应
    switch (message.topic) {
      case 'environment:temperature':
        this.reactToTemperatureChange(message.data);
        break;
      case 'environment:threat':
        this.reactToThreat(message.data);
        break;
      case 'environment:resource':
        this.reactToResourceChange(message.data);
        break;
    }
  }

  private reactToTemperatureChange(data: any) {
    if (data.value > 30) {
      console.log('🔥 启动冷却系统');
    }
  }

  private reactToThreat(data: any) {
    console.log('⚠️ 采取防御措施');
  }

  private reactToResourceChange(data: any) {
    if (data.level < 20) {
      console.log('📦 开始寻找资源');
    }
  }
}
```

### 模式3：反应式流（Reactive Streams）

使用RxJS等库实现反应式编程：

```typescript
// 反应式流实现（使用RxJS）
import { Subject, interval, merge } from 'rxjs';
import { filter, map, debounceTime, distinctUntilChanged } from 'rxjs/operators';

class StreamReactiveAgent {
  private temperatureStream = new Subject<number>();
  private threatStream = new Subject<Threat>();
  private resourceStream = new Subject<Resource>();

  constructor() {
    this.setupReactions();
  }

  private setupReactions() {
    // 响应温度变化（去抖动，避免频繁触发）
    this.temperatureStream.pipe(
      debounceTime(500),
      distinctUntilChanged(),
      filter(temp => temp > 30 || temp < 10)
    ).subscribe(temp => {
      if (temp > 30) {
        console.log(`🔥 高温: ${temp}°C`);
        this.activateCooling();
      } else {
        console.log(`❄️ 低温: ${temp}°C`);
        this.activateHeating();
      }
    });

    // 响应威胁（立即响应）
    this.threatStream.subscribe(threat => {
      console.log(`⚠️ 威胁: ${threat.type}`);
      this.handleThreat(threat);
    });

    // 响应资源变化（按资源等级过滤）
    this.resourceStream.pipe(
      filter(resource => resource.level < 20)
    ).subscribe(resource => {
      console.log(`📦 资源不足: ${resource.name}`);
      this.replenishResource(resource);
    });

    // 合并多个流
    merge(
      this.temperatureStream.pipe(map(t => ({ type: 'temperature', value: t }))),
      this.threatStream.pipe(map(t => ({ type: 'threat', value: t }))),
      this.resourceStream.pipe(map(r => ({ type: 'resource', value: r })))
    ).subscribe(event => {
      console.log('环境事件:', event);
      this.logEvent(event);
    });
  }

  // 模拟传感器数据流
  startSensors() {
    // 每秒读取温度
    interval(1000).subscribe(() => {
      const temp = 20 + Math.random() * 15;
      this.temperatureStream.next(temp);
    });

    // 随机生成威胁
    interval(5000).subscribe(() => {
      if (Math.random() > 0.7) {
        this.threatStream.next({
          id: `threat_${Date.now()}`,
          type: 'obstacle',
          severity: Math.random()
        });
      }
    });

    // 定期检查资源
    interval(3000).subscribe(() => {
      const resources = this.scanResources();
      resources.forEach(r => this.resourceStream.next(r));
    });
  }
}
```

---

## 反应性与其他特征的关系

### 1. 反应性 + 自主性

反应性Agent在自主运行时响应环境变化：

```typescript
// 结合反应性和自主性
class AutonomousReactiveAgent {
  private isRunning = false;
  private currentTask?: Task;

  async run() {
    this.isRunning = true;

    // 启动环境监控（反应性）
    this.startEnvironmentMonitoring();

    // 自主执行任务（自主性）
    while (this.isRunning && this.hasTasks()) {
      this.currentTask = await this.selectNextTask();

      try {
        await this.executeTask(this.currentTask);
      } catch (error) {
        // 遇到错误时响应
        await this.reactToError(error);
      }

      this.currentTask = undefined;
    }
  }

  private startEnvironmentMonitoring() {
    // 持续监控环境（反应性）
    setInterval(async () => {
      const state = await this.readEnvironment();
      const urgentEvents = this.detectUrgentEvents(state);

      if (urgentEvents.length > 0) {
        // 暂停当前任务，响应紧急事件
        await this.pauseCurrentTask();
        await this.handleUrgentEvents(urgentEvents);
        await this.resumeCurrentTask();
      }
    }, 500);
  }
}
```

### 2. 反应性 + 主动性

主动规划同时保持对环境的响应：

```typescript
// 结合反应性和主动性
class ProactiveReactiveAgent {
  private plan: Plan;

  async execute() {
    // 主动制定计划（主动性）
    this.plan = await this.createPlan();

    // 执行计划时保持反应性
    for (const step of this.plan.steps) {
      // 执行前检查环境（反应性）
      const state = await this.assessEnvironment();

      // 如果环境变化，调整计划（反应性）
      if (this.shouldAdjustPlan(state)) {
        console.log('环境变化，调整计划...');
        this.plan = await this.adjustPlan(this.plan, state);
      }

      // 执行计划步骤（主动性）
      await this.executeStep(step);
    }
  }
}
```

---

## 最佳实践

### 1. 设计合理的响应阈值

```typescript
// ✅ 好的实践：使用阈值避免过度响应
class ThresholdReactiveAgent {
  private config = {
    temperatureThresholds: {
      high: 30,
      low: 10,
      criticalHigh: 35,
      criticalLow: 5
    },
    resourceThresholds: {
      low: 20,
      critical: 10
    },
    threatThresholds: {
      high: 0.7,
      critical: 0.9
    }
  };

  async reactToTemperature(temp: number) {
    if (temp > this.config.temperatureThresholds.criticalHigh) {
      await this.emergencyCooling();
    } else if (temp > this.config.temperatureThresholds.high) {
      await this.activateCooling();
    }
    // 不响应小的温度波动
  }
}

// ❌ 不好的实践：对任何变化都响应
class OverReactiveAgent {
  async reactToTemperature(temp: number) {
    // 温度有任何变化就响应，导致系统不稳定
    if (temp > 25) {
      await this.activateCooling();
    }
  }
}
```

### 2. 实现响应防抖和节流

```typescript
// ✅ 好的实践：使用防抖避免频繁响应
class DebouncedReactiveAgent {
  private lastReactionTime: Map<string, number> = new Map();
  private debounceDelay = 1000; // 1秒

  async reactToChange(eventType: string, handler: () => Promise<void>) {
    const now = Date.now();
    const lastTime = this.lastReactionTime.get(eventType) || 0;

    // 如果距离上次响应不足1秒，跳过
    if (now - lastTime < this.debounceDelay) {
      return;
    }

    this.lastReactionTime.set(eventType, now);
    await handler();
  }
}

// 使用示例
agent.reactToChange('temperature', async () => {
  await agent.activateCooling();
});
```

### 3. 实现响应优先级队列

```typescript
// ✅ 好的实践：按优先级处理响应
class PriorityQueueReactiveAgent {
  private responseQueue: PriorityQueue<Response> = new PriorityQueue();

  async addResponse(response: Response) {
    this.responseQueue.enqueue(response, response.priority);

    if (!this.isProcessing) {
      await this.processResponses();
    }
  }

  private async processResponses() {
    while (!this.responseQueue.isEmpty()) {
      const response = this.responseQueue.dequeue();
      await this.executeResponse(response);
    }
  }
}
```

### 4. 记录响应历史

```typescript
// ✅ 好的实践：记录响应历史用于分析
class LoggingReactiveAgent {
  private responseHistory: ResponseLog[] = [];

  async react(event: Event) {
    const startTime = Date.now();

    try {
      await this.executeReaction(event);

      this.responseHistory.push({
        event,
        timestamp: new Date(),
        duration: Date.now() - startTime,
        success: true
      });
    } catch (error) {
      this.responseHistory.push({
        event,
        timestamp: new Date(),
        duration: Date.now() - startTime,
        success: false,
        error
      });
    }
  }

  getResponseStatistics() {
    return {
      totalResponses: this.responseHistory.length,
      successRate: this.responseHistory.filter(r => r.success).length / this.responseHistory.length,
      averageDuration: this.responseHistory.reduce((sum, r) => sum + r.duration, 0) / this.responseHistory.length
    };
  }
}
```

---

## 常见问题

### Q1: 反应性和自主性有什么区别？

**A**:
- **反应性**：被动的，由外部事件触发，"环境变化 → Agent响应"
- **自主性**：主动的，内部驱动，"Agent自己决策 → 执行动作"

```typescript
// 反应性：环境驱动
environment.on('change', () => {
  agent.react(); // 被动响应
});

// 自主性：内部驱动
agent.run(); // 主动执行
```

### Q2: 如何避免过度响应？

**A**: 使用以下策略：

1. **设置阈值**：只响应显著变化
2. **防抖**：短时间内只响应一次
3. **节流**：限制响应频率
4. **优先级**：只响应重要事件

```typescript
// 综合策略
class WellBehavedReactiveAgent {
  async reactToTemperature(temp: number) {
    // 1. 阈值检查
    if (Math.abs(temp - this.lastTemp) < 2) {
      return; // 变化太小，不响应
    }

    // 2. 防抖检查
    if (Date.now() - this.lastReactionTime < 1000) {
      return; // 响应太频繁，不响应
    }

    // 3. 优先级检查
    if (temp < 35) {
      return; // 不够紧急，不响应
    }

    // 4. 执行响应
    await this.activateCooling();
    this.lastReactionTime = Date.now();
    this.lastTemp = temp;
  }
}
```

### Q3: 如何测试反应性？

**A**: 通过模拟环境变化：

```typescript
describe('Agent Reactivity', () => {
  it('should react to temperature changes', async () => {
    const agent = new ReactiveAgent();
    const environment = new MockEnvironment();

    agent.attachTo(environment);

    // 模拟温度变化
    environment.setTemperature(35);
    await wait(100);

    // 验证Agent做出了响应
    expect(agent.isCoolingActive()).toBe(true);
  });

  it('should prioritize critical threats', async () => {
    const agent = new ReactiveAgent();

    // 同时发生多个事件
    agent.addEvent({ type: 'resource:low', priority: Priority.MEDIUM });
    agent.addEvent({ type: 'threat:critical', priority: Priority.CRITICAL });
    agent.addEvent({ type: 'temperature:high', priority: Priority.HIGH });

    // 验证优先处理关键威胁
    const firstResponse = await agent.getFirstResponse();
    expect(firstResponse.type).toBe('threat:critical');
  });
});
```

### Q4: 反应性Agent适合什么场景？

**A**:
- ✅ 实时监控系统
- ✅ 游戏AI（响应玩家动作）
- ✅ 机器人控制（响应传感器数据）
- ✅ 自动驾驶（响应路况变化）
- ✅ 网络安全（响应入侵检测）

---

## 总结

Agent的反应性是其核心特征之一，它使得Agent能够：

1. **感知环境**：持续监控环境状态
2. **及时响应**：对环境变化做出快速反应
3. **优先处理**：根据紧急程度调整响应顺序
4. **自适应**：根据不同变化采取不同策略

通过合理设计传感器系统、事件处理机制和响应策略，可以构建出高效的反应性Agent系统。

---

## 参考资料

- [Reactive Programming](https://en.wikipedia.org/wiki/Reactive_programming)
- [Observer Pattern](https://refactoring.guru/design-patterns/observer)
- [RxJS Documentation](https://rxjs.dev/)
- [Event-Driven Architecture](https://en.wikipedia.org/wiki/Event-driven_architecture)

---

**下一步学习**：
- [ ] 学习Agent的主动性（Proactiveness）
- [ ] 学习Agent的社会性（Social Ability）
- [ ] 实践构建反应式Agent
- [ ] 学习Agent的学习能力（Learning）
