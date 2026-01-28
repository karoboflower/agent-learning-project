/**
 * 代码重构Prompt模板
 */

export interface RefactorInput {
  code: string;
  language: string;
  goal: string;
  constraints?: string[];
  preserveBehavior?: boolean;
}

/**
 * 代码重构系统Prompt
 */
export const CODE_REFACTOR_SYSTEM_PROMPT = `你是一位经验丰富的代码重构专家，精通各种重构技术和设计模式。

你的重构原则：
1. **保持行为不变**：重构不改变代码的外部行为
2. **小步快跑**：进行小的、可验证的改进
3. **提高可读性**：让代码更容易理解
4. **增强可维护性**：让代码更容易修改和扩展
5. **改进性能**：在不牺牲可读性的前提下优化性能

常用重构技术：
- 提取方法/函数
- 重命名变量和函数
- 简化条件��达式
- 移除重复代码
- 引入参数对象
- 应用设计模式

回复格式要求：
1. 清晰说明重构的目标和原因
2. 提供重构前后的代���对比
3. 解释每个重构步骤
4. 指出重构带来的改进`;

/**
 * 构建代码重构Prompt
 */
export function buildRefactorPrompt(input: RefactorInput): string {
  const { code, language, goal, constraints, preserveBehavior = true } = input;

  let prompt = `请对以下${language}代码进行重构：\n\n`;

  prompt += `**重构目标：**\n${goal}\n\n`;

  if (constraints && constraints.length > 0) {
    prompt += `**约束条件：**\n${constraints.map((c, i) => `${i + 1}. ${c}`).join('\n')}\n\n`;
  }

  if (preserveBehavior) {
    prompt += `**重要：** 重构必须保持代码的原有行为和功能不变。\n\n`;
  }

  prompt += `**原始代码：**\n\`\`\`${language}\n${code}\n\`\`\`\n\n`;

  prompt += `请提供：

## 1. 重构分析
- 当前代码的问题
- 重构的必要性
- 预期的改进效果

## 2. 重构后的代码
\`\`\`${language}
// 在这里提供重构后的完整代码
\`\`\`

## 3. 主要改动说明
列出所有重要的改动，包括：
- 重构技术名称
- 改动描述
- 改进点

## 4. 代码对比
突出显示关键的改动点

## 5. 测试建议
建议如何验证重构是否成功`;

  return prompt;
}

/**
 * 特定重构技术Prompt
 */
export function buildSpecificRefactorPrompt(
  code: string,
  language: string,
  technique:
    | 'extract-method'
    | 'rename'
    | 'simplify-conditional'
    | 'remove-duplication'
    | 'introduce-parameter-object'
): string {
  const techniques = {
    'extract-method': {
      name: '提取方法',
      description: '将代码片段提取为独立的方法或函数',
      focus: '找出可以提取的代码块，提高代码的可读性和复用性',
    },
    rename: {
      name: '重命名',
      description: '为变量、函数、类等提供更有意义的名称',
      focus: '使用清晰、表意的名称，遵循命名规范',
    },
    'simplify-conditional': {
      name: '简化条件表达式',
      description: '简化复杂的if-else或switch语句',
      focus: '使用卫语句、策略模式等简化条件逻辑',
    },
    'remove-duplication': {
      name: '移除重复代码',
      description: '识别并消除代码重复',
      focus: '提取公共逻辑，应用DRY原则',
    },
    'introduce-parameter-object': {
      name: '引入参数对象',
      description: '将多个参数组合成对象',
      focus: '减少参数数量，提高函数签名的可读性',
    },
  };

  const tech = techniques[technique];

  return `请对以下${language}代码应用【${tech.name}】重构技术：

**重构技术：** ${tech.name}
**描述：** ${tech.description}
**重点：** ${tech.focus}

**代码：**
\`\`\`${language}
${code}
\`\`\`

请提供：
1. 识别需要重构的部分
2. 重构后的代码
3. 重构步骤说明
4. 改进效果分析`;
}

/**
 * 设计模式应用Prompt
 */
export function buildDesignPatternRefactorPrompt(
  code: string,
  language: string,
  pattern?: string
): string {
  let prompt = `请分析以下${language}代码，并建议应用合适的设计模式进行重构：\n\n`;

  if (pattern) {
    prompt += `**建议应用的设计模式：** ${pattern}\n\n`;
  }

  prompt += `**代码：**\n\`\`\`${language}\n${code}\n\`\`\`\n\n`;

  prompt += `请提供：

## 1. 设计模式建议
${pattern ? `说明如何应用${pattern}模式` : '建议最适合的设计模式'}

## 2. 重构后的代码
应用设计模式后的完整代码

## 3. 设计模式说明
- 模式的作用
- 为什么适合这个场景
- 带来的好处

## 4. 实现要点
关键的实现细节和注意事项`;

  return prompt;
}

/**
 * 性能优化重构Prompt
 */
export function buildPerformanceRefactorPrompt(
  code: string,
  language: string,
  performanceIssue?: string
): string {
  return `请对以下${language}代码进行性能优化重构：

${performanceIssue ? `**已知性能问题：** ${performanceIssue}\n\n` : ''}**代码：**
\`\`\`${language}
${code}
\`\`\`

请分析并优化：
1. 时间复杂度
2. 空间复杂度
3. 不必要的计算
4. 可以缓存的数据
5. 可以并行的操作

提供：
1. 性能瓶颈分析
2. 优化后的代码
3. 性能改进说明（Big-O分析）
4. 权衡考虑（如可读性vs性能）`;
}
