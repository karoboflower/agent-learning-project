# Agent核心逻辑

Task 2.1.2 - LangChain.js集成完成。

## ✅ 已完成

- [x] 配置LangChain.js
- [x] 创建LLM实例
- [x] 配置API密钥管理
- [x] 创建Agent基类
- [x] 实现代码助手Agent
- [x] 测试Agent运行

## 📁 文件结构

```
agent/
├── config.ts                  # LangChain配置和LLM实例创建
├── BaseAgent.ts               # Agent基类
├── CodeAssistantAgent.ts      # 代码助手Agent实现
├── index.ts                   # 导出
└── README.md                  # 本文件
```

## 🔧 核心组件

### config.ts
- LangChain配置类型定义
- 创建LLM实例（支持OpenAI和Anthropic）
- 从环境变量获取配置
- 模型列表常量

### BaseAgent.ts
- Agent基类，提供通用功能
- 对话历史管理
- 消息构建
- LLM调用封装

### CodeAssistantAgent.ts
- 继承BaseAgent
- 专门用于代码助手功能
- 实现方��：
  - `reviewCode()` - 代码审查
  - `suggestRefactor()` - 重构建议
  - `suggestTechStack()` - 技术栈选择
  - `ask()` - 通用对话

## 🚀 使用示例

### 基础使用

```typescript
import { createCodeAssistant } from '@/agent';

// 创建Agent实例
const agent = createCodeAssistant();

// 进行对话
const response = await agent.ask('什么是SOLID原则？');
console.log(response.content);
```

### 代码审查

```typescript
const response = await agent.reviewCode(
  `
  function add(a, b) {
    return a + b;
  }
  `,
  'javascript',
  '这是一个简单的加法函数'
);
```

### 代码重构

```typescript
const response = await agent.suggestRefactor(
  `
  const x = 1;
  const y = 2;
  const result = x + y;
  console.log(result);
  `,
  'javascript',
  '提高代码可读性'
);
```

### 技术栈选择

```typescript
const response = await agent.suggestTechStack(
  '需要构建一个电商网站',
  ['用户认证', '商品管理', '订单处理', '支付集成'],
  ['预算有限', '团队熟悉JavaScript']
);
```

## 🧪 测试

访问 `/agent-test` 页面进行交互式测试。

测试功能：
- ✅ 基础对话
- ✅ 代码审查
- ✅ 代码重构
- ✅ 技术栈选择

## 📖 API文档

### BaseAgent

#### 方法

- `chat(userMessage: string): Promise<AgentResponse>` - 发送消息
- `clearHistory(): void` - 清除对话历史
- `getHistory(): Message[]` - 获取对话历史
- `setConfig(config: Partial<LangChainConfig>): void` - 更新配置

### CodeAssistantAgent

#### 方法

- `reviewCode(code: string, language: string, context?: string)` - 代码审查
- `suggestRefactor(code: string, language: string, goal: string)` - 重构建议
- `suggestTechStack(projectDescription: string, requirements: string[], constraints?: string[])` - 技术栈选择
- `ask(question: string)` - 通用对话

## 🎯 下一步

Task 2.1.3 - 实现代码分析Prompt模板。

---

**完成日期**: 2026-01-28
**任务来源**: phase2-tasks.md - Task 2.1.2
