import type { Document } from '../types';
import { promises as fs } from 'fs';
import path from 'path';

/**
 * ��档存储服务
 * 使用JSON文件存储文档元数据（生产环境应使用数据库）
 */
export class DocumentStore {
  private storePath: string;
  private documents: Map<string, Document>;

  constructor() {
    this.storePath = path.join(process.cwd(), 'data', 'documents.json');
    this.documents = new Map();
  }

  /**
   * 初始化存储
   */
  async initialize(): Promise<void> {
    try {
      // 确保data目录存在
      const dataDir = path.join(process.cwd(), 'data');
      await fs.mkdir(dataDir, { recursive: true });

      // 尝试加载现有数据
      try {
        const data = await fs.readFile(this.storePath, 'utf-8');
        const docs = JSON.parse(data) as Document[];
        docs.forEach((doc) => {
          // 转换日期字符串为Date对象
          doc.uploadedAt = new Date(doc.uploadedAt);
          this.documents.set(doc.id, doc);
        });
        console.log(`[DocumentStore] 已加载 ${docs.length} 个文档`);
      } catch {
        // 文件不存在，创建新的
        await this.save();
        console.log('[DocumentStore] 创建新的文档存储');
      }
    } catch (error) {
      console.error('[DocumentStore] 初始化失败:', error);
      throw error;
    }
  }

  /**
   * 添加文档
   */
  async addDocument(document: Document): Promise<void> {
    this.documents.set(document.id, document);
    await this.save();
  }

  /**
   * 获取文档
   */
  async getDocument(id: string): Promise<Document | null> {
    return this.documents.get(id) || null;
  }

  /**
   * 获取所有文档
   */
  async getAllDocuments(): Promise<Document[]> {
    return Array.from(this.documents.values()).sort(
      (a, b) => b.uploadedAt.getTime() - a.uploadedAt.getTime()
    );
  }

  /**
   * 更新文档
   */
  async updateDocument(id: string, updates: Partial<Document>): Promise<void> {
    const doc = this.documents.get(id);
    if (doc) {
      this.documents.set(id, { ...doc, ...updates });
      await this.save();
    }
  }

  /**
   * 删除文档
   */
  async deleteDocument(id: string): Promise<void> {
    this.documents.delete(id);
    await this.save();
  }

  /**
   * 保存到文件
   */
  private async save(): Promise<void> {
    const docs = Array.from(this.documents.values());
    await fs.writeFile(this.storePath, JSON.stringify(docs, null, 2));
  }

  /**
   * 获取统计信息
   */
  async getStats(): Promise<{
    total: number;
    ready: number;
    processing: number;
    error: number;
  }> {
    const docs = Array.from(this.documents.values());
    return {
      total: docs.length,
      ready: docs.filter((d) => d.status === 'ready').length,
      processing: docs.filter((d) => d.status === 'processing').length,
      error: docs.filter((d) => d.status === 'error').length,
    };
  }
}

// 单例实例
let storeInstance: DocumentStore | null = null;

export async function getDocumentStore(): Promise<DocumentStore> {
  if (!storeInstance) {
    storeInstance = new DocumentStore();
    await storeInstance.initialize();
  }
  return storeInstance;
}
