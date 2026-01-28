# Prompt模板

Task 2.1.3 - 代码分析Prompt模板完成。

## ✅ 已完成

- [x] 设计代码审查Prompt模板
- [x] 实现代码审查功能
- [x] 设计代码重构Prompt模板
- [x] 实现代码重构建议功能
- [x] 设计技术栈选择Prompt模板
- [x] 实现技术栈建议功能
- [x] 编写Prompt设计文档

## 📁 文件结构

```
prompts/
├── codeReview.ts          # 代码审查Prompt模板
├── refactor.ts            # 代码重构Prompt模板
├── techStack.ts           # 技术栈选择Prompt模板
├── index.ts               # 导出
└── README.md              # 本文件
```

## 🎯 Prompt模板功能

### 1. 代码审查（codeReview.ts）

#### 主要功能
- **全面代码审查**: `buildCodeReviewPrompt()`
- **快速审查**: `buildQuickCodeReviewPrompt()`
- **领域专项审查**: `buildDomainSpecificReviewPrompt()`

#### 系统Prompt
- `CODE_REVIEW_SYSTEM_PROMPT` - 代码审查专家系统提示词

#### 审查重点
- 代码质量评估
- 潜在问题发现
- 最佳实践检查
- 设计模式评估
- 测试覆盖建议

#### 领域专项
- Security（安全性）
- Performance（性能）
- Accessibility（可访问性）
- Testing（测试）

### 2. 代码重构（refactor.ts）

#### 主要功能
- **通用重构**: `buildRefactorPrompt()`
- **特定技术重构**: `buildSpecificRefactorPrompt()`
- **设计模式应用**: `buildDesignPatternRefactorPrompt()`
- **性能优化**: `buildPerformanceRefactorPrompt()`

#### 系统Prompt
- `CODE_REFACTOR_SYSTEM_PROMPT` - 代码重构专家系统提示词

#### 重构技术
- Extract Method（提取方法）
- Rename（重命名）
- Simplify Conditional（简化条件）
- Remove Duplication（移除重复）
- Introduce Parameter Object（引入参数对象）

#### 重构原则
- 保持行为不变
- 小步快跑
- 提高可读性
- 增强可维护性
- 改进性能

### 3. 技术栈选择（techStack.ts）

#### 主要功能
- **完整技术栈方案**: `buildTechStackPrompt()`
- **前端技术栈**: `buildFrontendTechStackPrompt()`
- **后端技术栈**: `buildBackendTechStackPrompt()`
- **数据库选择**: `buildDatabaseSelectionPrompt()`
- **微服务架构**: `buildMicroservicesTechStackPrompt()`

#### 系统Prompt
- `TECH_STACK_SYSTEM_PROMPT` - 技术架构师系统提示词

#### 选型考虑
- 项目需求和场景
- 性能和可扩展性
- 团队技术能力
- 生态成熟度
- 成本和维护

#### 输出内容
- 需求分析
- 完整技术栈
- 选择理由
- 替代方案
- 架构建议
- 风险评估
- 实施路线图

## 🚀 使用示例

### 代码审查

```typescript
import { buildCodeReviewPrompt } from '@/prompts';

const input = {
  code: `
    function add(a, b) {
      return a + b;
    }
  `,
  language: 'javascript',
  context: '简单的加法函数',
  focusAreas: ['类型安全', '错误处理']
};

const prompt = buildCodeReviewPrompt(input);
// 使用prompt调用LLM
```

### 代码重构

```typescript
import { buildRefactorPrompt } from '@/prompts';

const input = {
  code: `
    const x = 1;
    const y = 2;
    console.log(x + y);
  `,
  language: 'javascript',
  goal: '提高代码可读性',
  constraints: ['保持输出格式'],
  preserveBehavior: true
};

const prompt = buildRefactorPrompt(input);
```

### 技术栈选择

```typescript
import { buildTechStackPrompt } from '@/prompts';

const input = {
  projectDescription: '在线教育平台',
  projectType: 'web应用',
  requirements: [
    '视频直播',
    '在线作业',
    '学习进度跟踪'
  ],
  constraints: ['预算有限'],
  teamSkills: ['JavaScript', 'Python'],
  scale: 'medium'
};

const prompt = buildTechStackPrompt(input);
```

## 🎨 Prompt设计原则

1. **结构化输出**
   - 使用Markdown格式
   - 明确的章节划分
   - 便于解析和展示

2. **上下文丰富**
   - 提供背景信息
   - 明确任务目标
   - 包含约束条件

3. **可定制性**
   - 灵活的参数
   - 支持多种场景
   - 可选的配置项

4. **专业性**
   - 遵循最佳实践
   - 引用行业标准
   - 提供深度分析

## 📖 详细文档

完整的Prompt设计文档请查看：
- [code-assistant-prompts.md](../../docs/learning-notes/code-assistant-prompts.md)

## 🎯 下一步

Task 2.1.4 - 实现核心功能（UI组件和业务逻辑）

---

**完成日期**: 2026-01-28
**任务来源**: phase2-tasks.md - Task 2.1.3
