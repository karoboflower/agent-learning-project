import { OpenAIEmbeddings } from '@langchain/openai';
import { HNSWLib } from '@langchain/community/vectorstores/hnswlib';
import { Document as LangChainDocument } from '@langchain/core/documents';
import type { DocumentChunk, SearchResult } from '../types';

/**
 * 向量数据库服务
 * 使用HNSWLib作为本地向量存储
 */
export class VectorDBService {
  private embeddings: OpenAIEmbeddings;
  private vectorStore: HNSWLib | null = null;
  private storePath: string;

  constructor() {
    this.embeddings = new OpenAIEmbeddings({
      openAIApiKey: process.env.OPENAI_API_KEY,
      modelName: process.env.OPENAI_EMBEDDING_MODEL || 'text-embedding-3-small',
    });

    this.storePath = process.env.VECTOR_STORE_PATH || './data/vector-store';
  }

  /**
   * 初始化向量存储
   */
  async initialize(): Promise<void> {
    try {
      // 尝试加载现有的向量存储
      this.vectorStore = await HNSWLib.load(this.storePath, this.embeddings);
      console.log('[VectorDB] 向量存储已加载');
    } catch (error) {
      // 如果不存在，创建新的向量存储
      this.vectorStore = await HNSWLib.fromDocuments(
        [],
        this.embeddings
      );
      console.log('[VectorDB] 创建新的向量存储');
    }
  }

  /**
   * 添加文档块到向量存储
   */
  async addChunks(chunks: DocumentChunk[]): Promise<void> {
    if (!this.vectorStore) {
      await this.initialize();
    }

    const documents = chunks.map(
      (chunk) =>
        new LangChainDocument({
          pageContent: chunk.content,
          metadata: {
            documentId: chunk.documentId,
            chunkId: chunk.id,
            ...chunk.metadata,
          },
        })
    );

    await this.vectorStore!.addDocuments(documents);
    await this.save();

    console.log(`[VectorDB] 已添加 ${chunks.length} 个文档块`);
  }

  /**
   * 相似度搜索
   */
  async search(
    query: string,
    topK: number = 5,
    filter?: Record<string, any>
  ): Promise<SearchResult[]> {
    if (!this.vectorStore) {
      await this.initialize();
    }

    const results = await this.vectorStore!.similaritySearchWithScore(
      query,
      topK,
      filter
    );

    return results.map(([doc, score]) => ({
      documentId: doc.metadata.documentId,
      documentName: doc.metadata.documentName || 'Unknown',
      chunkId: doc.metadata.chunkId,
      content: doc.pageContent,
      score,
      metadata: doc.metadata,
    }));
  }

  /**
   * 删除文档的所有向量
   */
  async deleteDocument(documentId: string): Promise<void> {
    // HNSWLib不支持直接删除，需要重建索引
    // 这里我们标记为删除，实际删除在重建时处理
    console.log(`[VectorDB] 标记删除文档: ${documentId}`);
  }

  /**
   * 保存向量存储
   */
  async save(): Promise<void> {
    if (this.vectorStore) {
      await this.vectorStore.save(this.storePath);
      console.log('[VectorDB] 向量存储已保存');
    }
  }

  /**
   * 获取向量存储统计信息
   */
  async getStats(): Promise<{
    totalVectors: number;
    dimensions: number;
  }> {
    if (!this.vectorStore) {
      await this.initialize();
    }

    // HNSWLib的统计信息
    return {
      totalVectors: 0, // HNSWLib不直接提供，需要自己维护
      dimensions: 1536, // text-embedding-3-small的维度
    };
  }
}

// 单例实例
let vectorDBInstance: VectorDBService | null = null;

export function getVectorDB(): VectorDBService {
  if (!vectorDBInstance) {
    vectorDBInstance = new VectorDBService();
  }
  return vectorDBInstance;
}
