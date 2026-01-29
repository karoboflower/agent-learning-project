/**
 * 文档类型
 */
export interface Document {
  id: string;
  name: string;
  type: 'pdf' | 'markdown' | 'txt';
  size: number;
  uploadedAt: Date;
  status: 'processing' | 'ready' | 'error';
  chunks?: number;
  error?: string;
}

/**
 * 文档块
 */
export interface DocumentChunk {
  id: string;
  documentId: string;
  content: string;
  metadata: {
    pageNumber?: number;
    chunkIndex: number;
    startChar: number;
    endChar: number;
  };
  embedding?: number[];
}

/**
 * 问答记录
 */
export interface QARecord {
  id: string;
  question: string;
  answer: string;
  sources: Array<{
    documentId: string;
    documentName: string;
    chunkId: string;
    content: string;
    score: number;
  }>;
  timestamp: Date;
}

/**
 * 检索结果
 */
export interface SearchResult {
  documentId: string;
  documentName: string;
  chunkId: string;
  content: string;
  score: number;
  metadata?: Record<string, any>;
}

/**
 * RAG配置
 */
export interface RAGConfig {
  chunkSize: number;
  chunkOverlap: number;
  topK: number;
  similarityThreshold: number;
  embeddingModel: string;
  llmModel: string;
}
