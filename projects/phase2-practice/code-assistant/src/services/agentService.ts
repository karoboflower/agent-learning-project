/**
 * Agent服务层
 * 封装Agent调用，添加缓存和错误处理
 */

import { createCodeAssistant } from '@/agent/CodeAssistantAgent';
import { agentCache } from './cache';
import { retry } from '@/utils/helpers';
import type { AgentResponse } from '@/agent/BaseAgent';

// 创建Agent实例
const agent = createCodeAssistant();

/**
 * 代码审查服务
 */
export async function reviewCodeService(params: {
  code: string;
  language: string;
  context?: string;
  focusAreas?: string[];
}): Promise<AgentResponse> {
  const cacheKey = `review:${params.language}:${params.code.slice(0, 100)}`;

  return agentCache.withCache(cacheKey, () =>
    retry(
      () =>
        agent.reviewCode(
          params.code,
          params.language,
          params.context,
          params.focusAreas
        ),
      {
        retries: 2,
        delay: 1000,
        onRetry: (error, attempt) => {
          console.log(`[Retry] Review code attempt ${attempt}:`, error.message);
        },
      }
    )
  );
}

/**
 * 代码重构服务
 */
export async function refactorCodeService(params: {
  code: string;
  language: string;
  goal: string;
  constraints?: string[];
}): Promise<AgentResponse> {
  const cacheKey = `refactor:${params.language}:${params.goal}:${params.code.slice(0, 100)}`;

  return agentCache.withCache(cacheKey, () =>
    retry(
      () =>
        agent.suggestRefactor(
          params.code,
          params.language,
          params.goal,
          params.constraints
        ),
      {
        retries: 2,
        delay: 1000,
        onRetry: (error, attempt) => {
          console.log(
            `[Retry] Refactor code attempt ${attempt}:`,
            error.message
          );
        },
      }
    )
  );
}

/**
 * 技术栈选择服务
 */
export async function techStackService(params: {
  projectDescription: string;
  requirements: string[];
  constraints?: string[];
  teamSkills?: string[];
  projectType?: string;
  scale?: 'small' | 'medium' | 'large' | 'enterprise';
}): Promise<AgentResponse> {
  const cacheKey = `techstack:${params.projectType}:${params.scale}:${params.projectDescription.slice(0, 100)}`;

  return agentCache.withCache(cacheKey, () =>
    retry(
      () =>
        agent.suggestTechStack(
          params.projectDescription,
          params.requirements,
          params.constraints,
          params.teamSkills,
          params.projectType,
          params.scale
        ),
      {
        retries: 2,
        delay: 1000,
        onRetry: (error, attempt) => {
          console.log(
            `[Retry] Tech stack analysis attempt ${attempt}:`,
            error.message
          );
        },
      }
    )
  );
}

/**
 * 清除缓存
 */
export function clearAgentCache(): void {
  agentCache.clear();
  console.log('[Cache] Cleared all agent cache');
}

/**
 * 获取缓存统计
 */
export function getCacheStats(): {
  size: number;
  maxSize: number;
} {
  return {
    size: agentCache.size(),
    maxSize: 50, // 从cache.ts中的配置
  };
}
