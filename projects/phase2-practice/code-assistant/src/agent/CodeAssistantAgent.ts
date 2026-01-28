import { BaseAgent } from './BaseAgent';
import type { LangChainConfig } from './config';
import { getConfigFromEnv } from './config';
import {
  buildCodeReviewPrompt,
  buildRefactorPrompt,
  buildTechStackPrompt,
  CODE_REVIEW_SYSTEM_PROMPT,
  CODE_REFACTOR_SYSTEM_PROMPT,
  TECH_STACK_SYSTEM_PROMPT,
} from '@/prompts';
import type {
  CodeReviewInput,
  RefactorInput,
  TechStackInput,
} from '@/prompts';

/**
 * 代码助手Agent
 * 专门用于代码审查、重构建议和技术栈选择
 */
export class CodeAssistantAgent extends BaseAgent {
  constructor(config?: LangChainConfig) {
    // 如果没有提供配置，从环境变量获取
    super(config || getConfigFromEnv());
  }

  /**
   * 系统提示词
   */
  protected getSystemPrompt(): string {
    return `你是一个专业的代码助手，擅长代码审查、重构建议和技术栈选择。

你的职责：
1. 代码审查：分析代码质量，发现潜在问题，提供改进建议
2. 代码重构：提供重构建议，提高代码的可读性、可维护性和性能
3. 技术栈选择：根据项目需求，推荐合适的技术栈和架构方案

回复要求：
- 清晰、专业、有条理
- 提供具体的代码示例
- 解释建议的原因
- 考虑最佳实践和行业标准

请始终保持友好和专业的态度。`;
  }

  /**
   * 代码审查（使用Prompt模板）
   */
  async reviewCode(
    code: string,
    language: string,
    context?: string,
    focusAreas?: string[]
  ) {
    const input: CodeReviewInput = {
      code,
      language,
      context,
      focusAreas,
    };

    const prompt = buildCodeReviewPrompt(input);
    return this.chat(prompt);
  }

  /**
   * 代码重构建议（使用Prompt模板）
   */
  async suggestRefactor(
    code: string,
    language: string,
    goal: string,
    constraints?: string[]
  ) {
    const input: RefactorInput = {
      code,
      language,
      goal,
      constraints,
      preserveBehavior: true,
    };

    const prompt = buildRefactorPrompt(input);
    return this.chat(prompt);
  }

  /**
   * 技术栈选择建议（使用Prompt模板）
   */
  async suggestTechStack(
    projectDescription: string,
    requirements: string[],
    constraints?: string[],
    teamSkills?: string[],
    projectType?: string,
    scale?: 'small' | 'medium' | 'large' | 'enterprise'
  ) {
    const input: TechStackInput = {
      projectDescription,
      projectType,
      requirements,
      constraints,
      teamSkills,
      scale,
    };

    const prompt = buildTechStackPrompt(input);
    return this.chat(prompt);
  }

  /**
   * 通用对话
   */
  async ask(question: string) {
    return this.chat(question);
  }

  /**
   * 使用特定系统Prompt进行代码审查
   */
  async reviewCodeWithSystemPrompt(code: string, language: string) {
    // 临时切换系统提示词
    const originalPrompt = this.getSystemPrompt;
    (this as any).getSystemPrompt = () => CODE_REVIEW_SYSTEM_PROMPT;

    const result = await this.reviewCode(code, language);

    // 恢复原始系统提示词
    (this as any).getSystemPrompt = originalPrompt;

    return result;
  }

  /**
   * 使用特定系统Prompt进行代码重构
   */
  async refactorWithSystemPrompt(
    code: string,
    language: string,
    goal: string
  ) {
    const originalPrompt = this.getSystemPrompt;
    (this as any).getSystemPrompt = () => CODE_REFACTOR_SYSTEM_PROMPT;

    const result = await this.suggestRefactor(code, language, goal);

    (this as any).getSystemPrompt = originalPrompt;

    return result;
  }

  /**
   * 使用特定系统Prompt进行技术栈选择
   */
  async selectTechStackWithSystemPrompt(
    projectDescription: string,
    requirements: string[]
  ) {
    const originalPrompt = this.getSystemPrompt;
    (this as any).getSystemPrompt = () => TECH_STACK_SYSTEM_PROMPT;

    const result = await this.suggestTechStack(
      projectDescription,
      requirements
    );

    (this as any).getSystemPrompt = originalPrompt;

    return result;
  }
}

/**
 * 创建代码助手Agent实例
 */
export function createCodeAssistant(
  config?: LangChainConfig
): CodeAssistantAgent {
  return new CodeAssistantAgent(config);
}
