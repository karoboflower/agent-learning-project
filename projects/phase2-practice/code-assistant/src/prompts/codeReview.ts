/**
 * 代码审查Prompt模板
 */

export interface CodeReviewInput {
  code: string;
  language: string;
  context?: string;
  focusAreas?: string[];
}

/**
 * 代码审查系统Prompt
 */
export const CODE_REVIEW_SYSTEM_PROMPT = `你是一位资深的代码审查专家，拥有多年的软件开发和代码审查经验。

你的审查重点：
1. **代码质量**：评估代码的整体质量、可读性和可维护性
2. **潜在问题**：发现bug、性能问题、安全漏洞
3. **最佳实践**：检查是否遵循语言和框架的最佳实践
4. **设计模式**：评估代码结构和设计模式的使用
5. **测试覆盖**：建议需要添加的测试用例

审查标准：
- 遵循SOLID原则
- 代码复用性和模块化
- 错误处理和边界条件
- 性能和可扩展性
- 安全性考虑

回复格式要求：
1. 使用清晰的结构化格式
2. 对每个问题提供具体的代码位置
3. 给出改进建议和示例代码
4. 评分理由要充分`;

/**
 * 构建代码审查Prompt
 */
export function buildCodeReviewPrompt(input: CodeReviewInput): string {
  const { code, language, context, focusAreas } = input;

  let prompt = `请对以下${language}代码进行全面审查：\n\n`;

  if (context) {
    prompt += `**代码背景：**\n${context}\n\n`;
  }

  if (focusAreas && focusAreas.length > 0) {
    prompt += `**重点关注：**\n${focusAreas.map((area, i) => `${i + 1}. ${area}`).join('\n')}\n\n`;
  }

  prompt += `**代码：**\n\`\`\`${language}\n${code}\n\`\`\`\n\n`;

  prompt += `请提供详细的审查报告，包括：

## 1. 整体评估
- 代码质量评分（1-10分）
- 总体印象和主要优点

## 2. 发现的问题
列出所有发现的问题，按严重程度分类：
- 🔴 严重问题（Critical）
- 🟡 警告（Warning）
- 🔵 建议（Suggestion）

对每个问题，请提供：
- 问题描述
- 代码位置（行号）
- 影响分析
- 修复建议

## 3. 代码改进建议
提供具体的改进建议，并附上改进后的代码示例。

## 4. 最佳实践建议
针对${language}和相关框架，提供最佳实践建议。

## 5. 测试建议
建议需要添加的测试用例。`;

  return prompt;
}

/**
 * 快速代码审查Prompt（简化版）
 */
export function buildQuickCodeReviewPrompt(
  code: string,
  language: string
): string {
  return `请快速审查以下${language}代码，重点关注：
1. 明显的bug或错误
2. 性能问题
3. 安全漏洞

代码：
\`\`\`${language}
${code}
\`\`\`

请简明扼要地列出发现的问题和改进建议。`;
}

/**
 * 特定领域代码审查Prompt
 */
export function buildDomainSpecificReviewPrompt(
  code: string,
  language: string,
  domain: 'security' | 'performance' | 'accessibility' | 'testing'
): string {
  const domainFocus = {
    security: {
      title: '安全性审查',
      focus: [
        'SQL注入、XSS等安全漏洞',
        '输入验证和数据清理',
        '身份认证和授权',
        '敏感数据处理',
        '依赖安全',
      ],
    },
    performance: {
      title: '性能审查',
      focus: [
        '算法复杂度',
        '内存使用',
        '数据库查询优化',
        '缓存策略',
        '异步处理',
      ],
    },
    accessibility: {
      title: '可访问性审查',
      focus: [
        'ARIA标签',
        '键盘导航',
        '屏幕阅读器支持',
        '颜色对比度',
        '语义化HTML',
      ],
    },
    testing: {
      title: '测试相关审查',
      focus: [
        '测试覆盖率',
        '测试用例质量',
        '边界条件测试',
        'Mock和Stub使用',
        '测试可维护性',
      ],
    },
  };

  const { title, focus } = domainFocus[domain];

  return `请对以下${language}代码进行${title}：

代码：
\`\`\`${language}
${code}
\`\`\`

重点关注：
${focus.map((item, i) => `${i + 1}. ${item}`).join('\n')}

请提供：
1. 发现的问题
2. 具体的改进建议
3. 示例代码（如适用）`;
}
