import { ChatOpenAI } from '@langchain/openai';
import { ChatAnthropic } from '@langchain/anthropic';

/**
 * LangChain配置
 */
export interface LangChainConfig {
  provider: 'openai' | 'anthropic';
  apiKey: string;
  model: string;
  temperature?: number;
  maxTokens?: number;
}

/**
 * 默认配置
 */
export const defaultConfig: Partial<LangChainConfig> = {
  provider: 'openai',
  temperature: 0.7,
  maxTokens: 2000,
};

/**
 * OpenAI模型列表
 */
export const OPENAI_MODELS = {
  GPT4: 'gpt-4',
  GPT4_TURBO: 'gpt-4-turbo-preview',
  GPT35_TURBO: 'gpt-3.5-turbo',
} as const;

/**
 * Anthropic模型列表
 */
export const ANTHROPIC_MODELS = {
  CLAUDE_3_OPUS: 'claude-3-opus-20240229',
  CLAUDE_3_SONNET: 'claude-3-sonnet-20240229',
  CLAUDE_3_HAIKU: 'claude-3-haiku-20240307',
} as const;

/**
 * 创建LLM实例
 */
export function createLLM(config: LangChainConfig) {
  const { provider, apiKey, model, temperature, maxTokens } = {
    ...defaultConfig,
    ...config,
  };

  if (provider === 'openai') {
    return new ChatOpenAI({
      openAIApiKey: apiKey,
      modelName: model,
      temperature,
      maxTokens,
    });
  } else if (provider === 'anthropic') {
    return new ChatAnthropic({
      anthropicApiKey: apiKey,
      modelName: model,
      temperature,
      maxTokens,
    });
  }

  throw new Error(`Unsupported provider: ${provider}`);
}

/**
 * 从环境变量获取配置
 */
export function getConfigFromEnv(): LangChainConfig {
  const apiKey = import.meta.env.VITE_OPENAI_API_KEY;

  if (!apiKey) {
    throw new Error('API密钥未配置，请在.env文件中设置VITE_OPENAI_API_KEY');
  }

  return {
    provider: 'openai',
    apiKey,
    model: OPENAI_MODELS.GPT35_TURBO,
    temperature: 0.7,
    maxTokens: 2000,
  };
}
