import type { BaseChatModel } from '@langchain/core/language_models/chat_models';
import { HumanMessage, SystemMessage, AIMessage } from '@langchain/core/messages';
import type { LangChainConfig } from './config';
import { createLLM } from './config';

/**
 * 消息类型
 */
export interface Message {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

/**
 * Agent响应
 */
export interface AgentResponse {
  content: string;
  usage?: {
    promptTokens: number;
    completionTokens: number;
    totalTokens: number;
  };
}

/**
 * Agent基类
 */
export abstract class BaseAgent {
  protected llm: BaseChatModel;
  protected config: LangChainConfig;
  protected conversationHistory: Message[] = [];

  constructor(config: LangChainConfig) {
    this.config = config;
    this.llm = createLLM(config);
  }

  /**
   * 获取系统提示词（由子类实现）
   */
  protected abstract getSystemPrompt(): string;

  /**
   * 发送消息并获取响应
   */
  async chat(userMessage: string): Promise<AgentResponse> {
    try {
      // 添加用户消息到历史
      this.conversationHistory.push({
        role: 'user',
        content: userMessage,
      });

      // 构建消息列表
      const messages = this.buildMessages();

      // 调用LLM
      const response = await this.llm.invoke(messages);

      // 提取响应内容
      const content =
        typeof response.content === 'string'
          ? response.content
          : JSON.stringify(response.content);

      // 添加AI响应到历史
      this.conversationHistory.push({
        role: 'assistant',
        content,
      });

      // 返回响应
      return {
        content,
        usage: response.usage_metadata
          ? {
              promptTokens: response.usage_metadata.input_tokens || 0,
              completionTokens: response.usage_metadata.output_tokens || 0,
              totalTokens: response.usage_metadata.total_tokens || 0,
            }
          : undefined,
      };
    } catch (error) {
      console.error('Agent chat error:', error);
      throw new Error(
        `Agent执行失败: ${error instanceof Error ? error.message : '未知错误'}`
      );
    }
  }

  /**
   * 构建消息列表
   */
  protected buildMessages() {
    const messages = [];

    // 添加系统提示词
    const systemPrompt = this.getSystemPrompt();
    if (systemPrompt) {
      messages.push(new SystemMessage(systemPrompt));
    }

    // 添加对话历史
    for (const msg of this.conversationHistory) {
      if (msg.role === 'user') {
        messages.push(new HumanMessage(msg.content));
      } else if (msg.role === 'assistant') {
        messages.push(new AIMessage(msg.content));
      }
    }

    return messages;
  }

  /**
   * 清除对话历史
   */
  clearHistory(): void {
    this.conversationHistory = [];
  }

  /**
   * 获取对话历史
   */
  getHistory(): Message[] {
    return [...this.conversationHistory];
  }

  /**
   * 设置配置
   */
  setConfig(config: Partial<LangChainConfig>): void {
    this.config = { ...this.config, ...config };
    this.llm = createLLM(this.config);
  }
}
