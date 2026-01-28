# Agent反应性（Reactivity）示例

这个项目展示了如何实现一个具有反应性的Agent系统。

## 🎯 核心特性

### 1. 反应性（Reactivity）
- ✅ **环境感知**：通过多种传感器实时监控环境
- ✅ **快速响应**：对环境变化做出及时反应
- ✅ **事件驱动**：基于事件的响应机制
- ✅ **优先级处理**：根据紧急程度调整响应顺序

### 2. 传感器系统
- **温度传感器**：监控环境温度
- **资源传感器**：扫描可用资源
- **威胁检测器**：检测环境中的威胁
- **机会扫描器**：发现潜在机会

### 3. 智能响应
- **防抖机制**：避免对同一事件过度响应
- **优先级队列**：关键事件优先处理
- **阈值管理**：只响应显著变化

## 📋 项目结构

```
reactivity/
├── typescript-reactive-agent.ts  # 反应性Agent实现
├── package.json                  # 项目配置
├── tsconfig.json                 # TypeScript配置
├── .gitignore                    # Git忽略文件
└── README.md                     # 本文件
```

## 🚀 快速开始

### 安装依赖

```bash
npm install
```

### 运行示例

```bash
npm run dev
```

### 构建项目

```bash
npm run build
npm start
```

## 💡 工作原理

### 1. 传感器监控循环

```typescript
while (isRunning) {
  // 1. 读取所有传感器
  const readings = await readAllSensors();

  // 2. 处理每个读数
  for (const reading of readings) {
    processSensorReading(reading);
  }

  // 3. 短暂等待后继续
  await sleep(100);
}
```

### 2. 事件生成与处理

```
传感器读数 → 分析 → 生成事件 → 添加到优先级队列 → 按优先级处理
```

### 3. 响应机制

```typescript
// 根据事件类型触发不同响应
switch (event.type) {
  case 'temperature:critical':
    await emergencyCooling();
    break;
  case 'threat:detected':
    await evadeThreat();
    break;
  case 'resource:low':
    await seekResource();
    break;
}
```

## 📊 响应优先级

| 优先级 | 事件类型 | 示例 |
|--------|----------|------|
| CRITICAL (0) | 极端环境/严重威胁 | 温度>35°C, 威胁严重度>0.8 |
| HIGH (1) | 异常环境/一般威胁 | 温度>30°C, 威胁严重度>0.6 |
| MEDIUM (2) | 资源不足/机会 | 资源<30%, 发现机会 |
| LOW (3) | 一般信息 | 常规监控数据 |

## 🔧 配置参数

### 温度阈值

```typescript
temperature: {
  high: 30,           // 高温阈值
  low: 15,            // 低温阈值
  criticalHigh: 35,   // 极高温阈值
  criticalLow: 10     // 极低温阈值
}
```

### 资源阈值

```typescript
resource: {
  low: 30,      // 资源不足阈值
  critical: 15  // 资源严重不足阈值
}
```

### 威胁阈值

```typescript
threat: {
  high: 0.6,     // 高威胁阈值
  critical: 0.8  // 严重威胁阈值
}
```

## 📈 示例输出

```
🚀 反应性Agent启动
⏱️  运行时长: 30秒

👁️  开始监控传感器...

🔥 [HIGH] 高温警告: 31.2°C
   → 启动冷却系统
   ✅ 冷却系统已激活

⚠️⚠️⚠️ [CRITICAL] 严重威胁: predator (严重度: 0.85)
   → 执行紧急规避动作
   ✅ 已规避威胁，移动到安全位置

📦 [MEDIUM] 资源不足: Water (25%)
   → 寻找 Water 资源
   ✅ 已规划资源收集路线

💡 [MEDIUM] 发现机会: resource_cache (价值: 75)
   → 评估机会 resource_cache
   ✅ 机会价值高，立即行动

============================================================
📊 Agent运行统计
============================================================
总响应次数: 24
成功响应: 24
失败响应: 0
平均响应时间: 0.52ms

优先级分布:
  Critical: 5
  High: 8
  Medium: 11

事件类型分布:
  temperature: 15
  threat: 3
  resource: 4
  opportunity: 2

当前环境状态:
  温度: 26.3°C
  资源数: 2
  威胁数: 0
  机会数: 0
============================================================
```

## 🎓 学习要点

### 1. 反应性核心概念

- **感知-响应循环**：持续监控 → 检测变化 → 及时响应
- **事件驱动架构**：基于事件的异步响应机制
- **优先级管理**：根据紧急程度调整处理顺序

### 2. 反应性 vs 自主性

| 特征 | 反应性 | 自主性 |
|------|--------|--------|
| 驱动方式 | 外部事件驱动 | 内部目标驱动 |
| 行为特点 | 被动响应 | 主动执行 |
| 决策依据 | 环境变化 | 内部状态和目标 |
| 触发条件 | 环境事件 | 自主决策 |

### 3. 实现关键点

```typescript
// 1. 防抖：避免过度响应
if (now - lastReactionTime < debounceDelay) {
  return; // 跳过
}

// 2. 阈值：只响应显著变化
if (Math.abs(newValue - oldValue) < threshold) {
  return; // 变化太小，不响应
}

// 3. 优先级：关键事件优先
if (event.priority === CRITICAL) {
  handleImmediately(event);
} else {
  addToQueue(event);
}
```

## 🔍 扩展建议

### 1. 添加更多传感器类型

```typescript
class ProximitySensor implements Sensor {
  // 检测附近物体
}

class BatterySensor implements Sensor {
  // 监控电池状态
}
```

### 2. 实现复杂的响应策略

```typescript
class AdaptiveReactiveAgent extends ReactiveAgent {
  // 根据历史数据调整响应策略
  async adaptResponse(event: Event) {
    const history = this.getEventHistory(event.type);
    const effectiveness = this.evaluateEffectiveness(history);

    if (effectiveness < 0.5) {
      return this.adjustStrategy(event);
    }
  }
}
```

### 3. 添加学习能力

```typescript
class LearningReactiveAgent extends ReactiveAgent {
  // 从响应结果中学习
  async learn(event: Event, response: Response) {
    if (response.success) {
      this.reinforceStrategy(event.type);
    } else {
      this.exploreAlternatives(event.type);
    }
  }
}
```

## 📚 参考资料

- [Reactive Programming](https://en.wikipedia.org/wiki/Reactive_programming)
- [Event-Driven Architecture](https://martinfowler.com/articles/201701-event-driven.html)
- [Observer Pattern](https://refactoring.guru/design-patterns/observer)
- [Publish-Subscribe Pattern](https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern)

## 🤝 相关示例

- [自主性示例](../autonomy) - 学习Agent的自主性
- [主动性示例](../proactiveness) - 学习Agent的主动性（即将推出）

## 📝 总结

这个示例展示了如何实现一个具有反应性的Agent系统：

1. ✅ **实时监控**：通过传感器系统持续监控环境
2. ✅ **快速响应**：检测到变化后立即做出反应
3. ✅ **智能处理**：使用优先级队列和防抖机制
4. ✅ **事件驱动**：基于EventEmitter的事件驱动架构

反应性是Agent的核心特征之一，它使Agent能够适应动态变化的环境，做出及时准确的响应。

---

**下一步学习**：
- [ ] 学习Agent的主动性（Proactiveness）
- [ ] 学习Agent的社会性（Social Ability）
- [ ] 结合反应性和自主性构建更复杂的Agent
