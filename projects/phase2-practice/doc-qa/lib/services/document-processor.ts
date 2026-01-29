import { RecursiveCharacterTextSplitter } from 'langchain/text_splitter';
import type { Document, DocumentChunk } from '../types';
import { v4 as uuidv4 } from 'uuid';
import pdf from 'pdf-parse/lib/pdf-parse';
import { marked } from 'marked';

/**
 * 文档处理服务
 * 负责文档解析、分块和向量化准备
 */
export class DocumentProcessor {
  private textSplitter: RecursiveCharacterTextSplitter;

  constructor() {
    this.textSplitter = new RecursiveCharacterTextSplitter({
      chunkSize: parseInt(process.env.CHUNK_SIZE || '1000'),
      chunkOverlap: parseInt(process.env.CHUNK_OVERLAP || '200'),
      separators: ['\n\n', '\n', '。', '！', '？', '；', '，', ' ', ''],
    });
  }

  /**
   * 处理上传的文件
   */
  async processFile(
    file: File,
    documentId: string
  ): Promise<{ chunks: DocumentChunk[]; totalChunks: number }> {
    const fileType = this.getFileType(file.name);
    let text: string;

    switch (fileType) {
      case 'pdf':
        text = await this.parsePDF(file);
        break;
      case 'markdown':
        text = await this.parseMarkdown(file);
        break;
      case 'txt':
        text = await this.parseText(file);
        break;
      default:
        throw new Error(`不支持的文件类型: ${fileType}`);
    }

    // 分块
    const chunks = await this.splitText(text, documentId);

    return {
      chunks,
      totalChunks: chunks.length,
    };
  }

  /**
   * 解析PDF文件
   */
  private async parsePDF(file: File): Promise<string> {
    const arrayBuffer = await file.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);

    try {
      const data = await pdf(buffer);
      return data.text;
    } catch (error) {
      console.error('[DocumentProcessor] PDF解析失败:', error);
      throw new Error('PDF文件解析失败');
    }
  }

  /**
   * 解析Markdown文件
   */
  private async parseMarkdown(file: File): Promise<string> {
    const text = await file.text();
    // 保留Markdown文本，不转换为HTML
    return text;
  }

  /**
   * 解析纯文本文件
   */
  private async parseText(file: File): Promise<string> {
    return await file.text();
  }

  /**
   * 文本分块
   */
  private async splitText(
    text: string,
    documentId: string
  ): Promise<DocumentChunk[]> {
    const splits = await this.textSplitter.splitText(text);

    let currentChar = 0;
    const chunks: DocumentChunk[] = splits.map((content, index) => {
      const chunk: DocumentChunk = {
        id: uuidv4(),
        documentId,
        content: content.trim(),
        metadata: {
          chunkIndex: index,
          startChar: currentChar,
          endChar: currentChar + content.length,
        },
      };

      currentChar += content.length;
      return chunk;
    });

    return chunks;
  }

  /**
   * 获取文件类型
   */
  private getFileType(filename: string): 'pdf' | 'markdown' | 'txt' {
    const ext = filename.split('.').pop()?.toLowerCase();

    switch (ext) {
      case 'pdf':
        return 'pdf';
      case 'md':
      case 'markdown':
        return 'markdown';
      case 'txt':
        return 'txt';
      default:
        throw new Error(`不支持的文件扩展名: ${ext}`);
    }
  }

  /**
   * 验证文件
   */
  validateFile(file: File): {
    valid: boolean;
    error?: string;
  } {
    // 检查文件大小
    const maxSize = parseInt(process.env.MAX_FILE_SIZE || '10485760'); // 10MB
    if (file.size > maxSize) {
      return {
        valid: false,
        error: `文件大小超过限制 (${(maxSize / 1024 / 1024).toFixed(0)}MB)`,
      };
    }

    // 检查文件类型
    const allowedTypes = (
      process.env.ALLOWED_FILE_TYPES || 'pdf,md,txt'
    ).split(',');
    const ext = file.name.split('.').pop()?.toLowerCase();

    if (!ext || !allowedTypes.includes(ext)) {
      return {
        valid: false,
        error: `不支持的文件类型，仅支持: ${allowedTypes.join(', ')}`,
      };
    }

    return { valid: true };
  }

  /**
   * 提取文本摘要
   */
  extractSummary(text: string, maxLength: number = 200): string {
    if (text.length <= maxLength) {
      return text;
    }
    return text.substring(0, maxLength) + '...';
  }
}

// 单例实例
let processorInstance: DocumentProcessor | null = null;

export function getDocumentProcessor(): DocumentProcessor {
  if (!processorInstance) {
    processorInstance = new DocumentProcessor();
  }
  return processorInstance;
}
