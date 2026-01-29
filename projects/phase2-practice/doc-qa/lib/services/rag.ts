import { ChatOpenAI } from '@langchain/openai';
import type { QARecord, SearchResult } from '../types';
import { getVectorDB } from './vector-db';
import { v4 as uuidv4 } from 'uuid';

/**
 * RAG提示词模板
 */
const RAG_SYSTEM_PROMPT = `你是一个专业的文档问答助手。你的任务是根据提供的文档上下文来回答用户的问题。

回答要求：
1. 只基于提供的文档上下文回答问题
2. 如果上下文中没有足够信息，明确告知用户
3. 回答要准确、简洁、易懂
4. 可以引用具体的文档内容来支持你的回答
5. 如果问题不清楚，可以要求用户澄清

请始终保持专业和有帮助的态度。`;

function buildRAGPrompt(question: string, context: SearchResult[]): string {
  const contextText = context
    .map((result, index) => {
      return `[文档 ${index + 1}] (相关度: ${result.score.toFixed(2)})\n${result.content}\n`;
    })
    .join('\n---\n\n');

  return `基于以下文档上下文回答问题：

## 文档上下文

${contextText}

## 用户问题

${question}

## 你的回答

请基于上述文档上下文回答用户的问题。如果上下文中没有足够信息，请明确说明。`;
}

/**
 * RAG服务
 * 实现检索增强生成
 */
export class RAGService {
  private llm: ChatOpenAI;
  private vectorDB: ReturnType<typeof getVectorDB>;
  private topK: number;
  private similarityThreshold: number;

  constructor() {
    this.llm = new ChatOpenAI({
      openAIApiKey: process.env.OPENAI_API_KEY,
      modelName: process.env.OPENAI_CHAT_MODEL || 'gpt-4',
      temperature: 0.7,
    });

    this.vectorDB = getVectorDB();
    this.topK = parseInt(process.env.TOP_K || '5');
    this.similarityThreshold = parseFloat(
      process.env.SIMILARITY_THRESHOLD || '0.7'
    );
  }

  /**
   * 问答接口
   */
  async ask(question: string): Promise<QARecord> {
    console.log(`[RAG] 处理问题: ${question}`);

    // 1. 检索相关文档
    const searchResults = await this.retrieveContext(question);

    if (searchResults.length === 0) {
      return {
        id: uuidv4(),
        question,
        answer: '抱歉，我在文档中没有找到与你的问题相关的信息。请尝试换一个问题或上传更多相关文档。',
        sources: [],
        timestamp: new Date(),
      };
    }

    // 2. 构建prompt
    const prompt = buildRAGPrompt(question, searchResults);

    // 3. 生成回答
    try {
      const response = await this.llm.invoke([
        { role: 'system', content: RAG_SYSTEM_PROMPT },
        { role: 'user', content: prompt },
      ]);

      const answer = response.content as string;

      // 4. 构建QA记录
      const qaRecord: QARecord = {
        id: uuidv4(),
        question,
        answer,
        sources: searchResults,
        timestamp: new Date(),
      };

      console.log(`[RAG] 回答生成成功`);
      return qaRecord;
    } catch (error) {
      console.error('[RAG] 生成回答失败:', error);
      throw new Error('生成回答失败，请稍后重试');
    }
  }

  /**
   * 检索相关文档
   */
  private async retrieveContext(query: string): Promise<SearchResult[]> {
    try {
      // 使用向量数据库进行相似度搜索
      const results = await this.vectorDB.search(query, this.topK);

      // 过滤低相关度的结果
      const filteredResults = results.filter(
        (result) => result.score >= this.similarityThreshold
      );

      console.log(
        `[RAG] 检索到 ${filteredResults.length}/${results.length} 个相关文档块`
      );

      return filteredResults;
    } catch (error) {
      console.error('[RAG] 检索失败:', error);
      return [];
    }
  }

  /**
   * 生成Embedding
   */
  async generateEmbedding(text: string): Promise<number[]> {
    // 这个方法由VectorDB内部使用
    // 这里提供一个接口供测试使用
    return [];
  }

  /**
   * 重新排序检索结果
   * 可以基于多种因素进行重排序
   */
  private rerank(results: SearchResult[]): SearchResult[] {
    // 简单实现：按相似度排序
    return results.sort((a, b) => b.score - a.score);
  }
}

// 单例实例
let ragInstance: RAGService | null = null;

export function getRAGService(): RAGService {
  if (!ragInstance) {
    ragInstance = new RAGService();
  }
  return ragInstance;
}
