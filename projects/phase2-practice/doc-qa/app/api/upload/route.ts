import { NextRequest, NextResponse } from 'next/server';
import { getDocumentProcessor } from '@/lib/services/document-processor';
import { getVectorDB } from '@/lib/services/vector-db';
import { getDocumentStore } from '@/lib/services/document-store';
import type { Document } from '@/lib/types';
import { v4 as uuidv4 } from 'uuid';

export async function POST(request: NextRequest) {
  try {
    const formData = await request.formData();
    const file = formData.get('file') as File;

    if (!file) {
      return NextResponse.json(
        { error: '请选择要上传的文件' },
        { status: 400 }
      );
    }

    // 1. 验证文件
    const processor = getDocumentProcessor();
    const validation = processor.validateFile(file);

    if (!validation.valid) {
      return NextResponse.json(
        { error: validation.error },
        { status: 400 }
      );
    }

    // 2. 创建文档记录
    const documentId = uuidv4();
    const document: Document = {
      id: documentId,
      name: file.name,
      type: file.name.endsWith('.pdf')
        ? 'pdf'
        : file.name.endsWith('.md')
        ? 'markdown'
        : 'txt',
      size: file.size,
      uploadedAt: new Date(),
      status: 'processing',
    };

    const store = await getDocumentStore();
    await store.addDocument(document);

    // 3. 处理文档（异步）
    processDocumentAsync(file, documentId, document.name).catch((error) => {
      console.error('[Upload API] 文档处理失败:', error);
      store.updateDocument(documentId, {
        status: 'error',
        error: error.message,
      });
    });

    // 4. 立即返回
    return NextResponse.json({
      success: true,
      document: {
        ...document,
        status: 'processing',
      },
    });
  } catch (error: any) {
    console.error('[Upload API] 错误:', error);
    return NextResponse.json(
      { error: '文件上传失败' },
      { status: 500 }
    );
  }
}

/**
 * 异步处理文档
 */
async function processDocumentAsync(
  file: File,
  documentId: string,
  documentName: string
) {
  const processor = getDocumentProcessor();
  const vectorDB = getVectorDB();
  const store = await getDocumentStore();

  try {
    // 1. 初始化向量数据库
    await vectorDB.initialize();

    // 2. 处理文件并分块
    const { chunks } = await processor.processFile(file, documentId);

    // 3. 添加文档名到每个chunk的metadata
    const chunksWithName = chunks.map((chunk) => ({
      ...chunk,
      metadata: {
        ...chunk.metadata,
        documentName,
      },
    }));

    // 4. 向量化并存储
    await vectorDB.addChunks(chunksWithName);

    // 5. 更新文档状态
    await store.updateDocument(documentId, {
      status: 'ready',
      chunks: chunks.length,
    });

    console.log(`[Upload API] 文档处理完成: ${documentName}`);
  } catch (error: any) {
    console.error('[Upload API] 文档处理失败:', error);
    await store.updateDocument(documentId, {
      status: 'error',
      error: error.message,
    });
  }
}
