# 代码助手Prompt设计文档

Task 2.1.3 - 代码分析Prompt模板设计。

## 📋 概述

本文档详细说明了代码助手中使用的Prompt模板设计，包括代码审查、代码重构和技术栈选择三个核心功能的Prompt设计思路和实现。

## 🎯 设计原则

### 1. 结构化输出
- 使用Markdown格式组织输出
- 明确的章节划分
- 便于解析和展示

### 2. 上下文丰富
- 提供充足的背景信息
- 明确任务目标和约束
- 包含示例和参考标准

### 3. 可定制性
- 支持自定义关注点
- 灵活的参数配置
- 多种使用场景

### 4. 专业性
- 遵循行业最佳实践
- 引用专业术语和标准
- 提供深度分析

## 📝 Prompt模板详解

### 一、代码审查Prompt

#### 设计思路

代码审查是代码助手的核心功能之一，需要全面、深入地分析代码质量。

#### 系统Prompt

```typescript
CODE_REVIEW_SYSTEM_PROMPT = `你是一位资深的代码审查专家...`
```

**关键要素：**
- 角色定位：资深代码审查专家
- 审查重点：质量、问题、最佳实践、设计模式、测试覆盖
- 审查标准：SOLID原则、代码复用、错误处理等
- 输出格式：结构化、具体位置、示例代码

#### 用户Prompt构建

```typescript
buildCodeReviewPrompt(input: CodeReviewInput): string
```

**Prompt结构：**

1. **基础信息**
   - 语言类型
   - 代码背景（可选）
   - 重点关注领域（可选）

2. **代码内容**
   - 代码块展示
   - 语法高亮标记

3. **输出要求**
   - 整体评估（评分+优点）
   - 问题分类（严重/警告/建议）
   - 改进建议（带示例）
   - 最佳实践
   - 测试建议

#### 变体Prompt

**1. 快速审查**
```typescript
buildQuickCodeReviewPrompt()
```
- 简化版本，快速反馈
- 只关注明显问题
- 适合快速迭代

**2. 领域专项审查**
```typescript
buildDomainSpecificReviewPrompt()
```
支持的领域：
- `security`: 安全性审查
- `performance`: 性能审查
- `accessibility`: 可访问性审查
- `testing`: 测试相关审查

#### 示例

```typescript
const input = {
  code: '...',
  language: 'typescript',
  context: '这是一个API处理函数',
  focusAreas: ['错误处理', '性能优化']
};

const prompt = buildCodeReviewPrompt(input);
```

---

### 二、代码重构Prompt

#### 设计思路

重构Prompt需要平衡代码质量提升和行为保持不变，提供清晰的重构路径。

#### 系统Prompt

```typescript
CODE_REFACTOR_SYSTEM_PROMPT = `你是一位经验丰富的代码重构专家...`
```

**关键要素：**
- 角色定位：代码重构专家
- 重构原则：保持行为、小步快跑、提高可读性等
- 常用技术：提取方法、重命名、简化条件、移除重复等
- 输出格式：目标说明、对比展示、步骤解释

#### 用户Prompt构建

```typescript
buildRefactorPrompt(input: RefactorInput): string
```

**Prompt结构：**

1. **重构信息**
   - 重构目标
   - 约束条件
   - 是否保持行为

2. **原始代码**
   - 完整代码展示

3. **输出要求**
   - 重构分析（问题+必要性+预期效果）
   - 重构后代码（完整）
   - 改动说明（技术+描述+改进点）
   - 代码对比
   - 测试建议

#### 变体Prompt

**1. 特定重构技术**
```typescript
buildSpecificRefactorPrompt()
```
支持的技术：
- `extract-method`: 提取方法
- `rename`: 重命名
- `simplify-conditional`: 简化条件
- `remove-duplication`: 移除重复
- `introduce-parameter-object`: 引入参数对象

**2. 设计模式应用**
```typescript
buildDesignPatternRefactorPrompt()
```
- 建议或应用特定设计模式
- 提供模式说明和适用场景

**3. 性能优化重构**
```typescript
buildPerformanceRefactorPrompt()
```
- 专注性能提升
- 包含Big-O分析
- 权衡可读性和性能

#### 示例

```typescript
const input = {
  code: '...',
  language: 'javascript',
  goal: '提高代码可读性和可维护性',
  constraints: ['不能改变API接口'],
  preserveBehavior: true
};

const prompt = buildRefactorPrompt(input);
```

---

### 三、技术栈选择Prompt

#### 设计思路

技术选型需要综合考虑多个因素，提供全面的分析和多个候选方案。

#### 系统Prompt

```typescript
TECH_STACK_SYSTEM_PROMPT = `你是一位经验丰富的技术架构师...`
```

**关键要素：**
- 角色定位：技术架构师
- 选型原则：适合场景、团队能力、生态成熟度等
- 考虑因素：需求、性能、团队、成本、维护等
- 输出格式：系统分析、完整方案、理由说明、替代方案

#### 用户Prompt构建

```typescript
buildTechStackPrompt(input: TechStackInput): string
```

**Prompt结构：**

1. **项目信息**
   - 项目描述
   - 项目类型
   - 项目规模
   - 功能需求
   - 约束条件
   - 团队技术栈

2. **输出要求**
   - 需求分析
   - 推荐技术栈（前端/后端/基础设施/工具）
   - 技术选择理由
   - 替代方案
   - 架构建议
   - 风险和注意事项
   - 实施路线图

#### 变体Prompt

**1. 前端技术栈**
```typescript
buildFrontendTechStackPrompt()
```
- 专注前端技术选型
- 框架、状态管理、构建工具等

**2. 后端技术栈**
```typescript
buildBackendTechStackPrompt()
```
- 专注后端技术选型
- 语言、框架、数据库、API设计等

**3. 数据库选择**
```typescript
buildDatabaseSelectionPrompt()
```
- 专注数据库选型
- SQL vs NoSQL
- 具体产品推荐
- 数据架构建议

**4. 微服务架构**
```typescript
buildMicroservicesTechStackPrompt()
```
- 微服务全栈方案
- 服务通信、数据管理、基础设施

#### 示例

```typescript
const input = {
  projectDescription: '电商平台',
  projectType: 'web应用',
  requirements: [
    '用户认证',
    '商品管理',
    '订单处理',
    '支付集成'
  ],
  constraints: ['预算有限', '3个月上线'],
  teamSkills: ['JavaScript', 'Python'],
  scale: 'medium'
};

const prompt = buildTechStackPrompt(input);
```

---

## 🎨 Prompt工程最佳实践

### 1. 明确角色定位
```
你是一位资深的XX专家，拥有X年的X经验
```
- 建立专业性
- 设定期望水平

### 2. 提供清晰的任务描述
```
请对以下代码进行全面审查...
```
- 明确任务类型
- 说明输出期望

### 3. 使用结构化输出
```
## 1. 整体评估
## 2. 发现的问题
## 3. 改进建议
```
- 便于解析
- 提高可读性

### 4. 包含示例和参考
```
对每个问题，请提供：
- 问题描述
- 代码位置
- 修复建议
```
- 明确输出格式
- 保证一致性

### 5. 考虑上下文
```
代码背景：这是一个API处理函数
重点关注：错误处理、性能优化
```
- 提供必要信息
- 聚焦关键点

### 6. 设置约束条件
```
重构必须保持代码的原有行为和功能不变
```
- 明确边界
- 避免越界

## 🧪 测试和优化

### Prompt测试方法

1. **功能测试**
   - 验证输出是否符合要求
   - 检查格式是否正确
   - 确认内容完整性

2. **边界测试**
   - 测试极端情况
   - 验证错误处理
   - 检查约束遵守

3. **对比测试**
   - 比较不同Prompt版本
   - 评估输出质量
   - 收集用户反馈

### Prompt优化策略

1. **迭代优化**
   - 基于测试结果调整
   - 收集真实使用数据
   - 持续改进

2. **A/B测试**
   - 对比不同Prompt方案
   - 量化评估效果
   - 选择最优方案

3. **用户反馈**
   - 收集用户建议
   - 分析常见问题
   - 针对性改进

## 📊 Prompt效果评估

### 评估维度

1. **准确性**
   - 输出是否符合预期
   - 建议是否正确
   - 分析是否深入

2. **完整性**
   - 是否覆盖所有要点
   - 信息是否充分
   - 格式是否完整

3. **可用性**
   - 输出是否易于理解
   - 建议是否可操作
   - 格式是否便于使用

4. **一致性**
   - 多次调用结果是否稳定
   - 格式是否统一
   - 风格是否一致

### 质量指标

- **响应质量**: 输出的专业性和准确性
- **格式规范**: 输出格式的一致性
- **可操作性**: 建议的实用性
- **用户满意度**: 用户反馈评分

## 🔄 持续改进

### 改进流程

1. **收集反馈**
   - ���户评价
   - 使用数据
   - 错误日志

2. **分析问题**
   - 识别常见问题
   - 分析根本原因
   - 优先级排序

3. **设计改进**
   - 提出改进方案
   - 设计新Prompt
   - 准备测试

4. **测试验证**
   - 功能测试
   - 对比测试
   - 用户验证

5. **部署上线**
   - 灰度发布
   - 监控效果
   - 全量上线

## 📚 参考资源

### Prompt工程资源
- [OpenAI Prompt Engineering Guide](https://platform.openai.com/docs/guides/prompt-engineering)
- [Anthropic Prompt Engineering](https://docs.anthropic.com/claude/docs/introduction-to-prompt-design)
- [LangChain Prompt Templates](https://js.langchain.com/docs/modules/prompts/)

### 代码质量标准
- SOLID原则
- Clean Code
- 设计模式
- 重构手法

### 技术选型参考
- ThoughtWorks技术雷达
- Stack Overflow技术趋势
- GitHub项目热度

---

**创建日期**: 2026-01-28
**任务来源**: phase2-tasks.md - Task 2.1.3
**维护者**: Code Assistant Team
