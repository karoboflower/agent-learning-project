import { NextResponse } from 'next/server';
import { getDocumentStore } from '@/lib/services/document-store';
import { getVectorDB } from '@/lib/services/vector-db';

/**
 * GET /api/documents
 * 获取所有文档列表
 */
export async function GET() {
  try {
    const store = await getDocumentStore();
    const documents = await store.getAllDocuments();

    return NextResponse.json({
      documents,
      total: documents.length,
    });
  } catch (error: any) {
    console.error('[Documents API] 获取文档列表失败:', error);
    return NextResponse.json(
      { error: '获取文档列表失败' },
      { status: 500 }
    );
  }
}

/**
 * DELETE /api/documents?id=xxx
 * 删除文档
 */
export async function DELETE(request: Request) {
  try {
    const { searchParams } = new URL(request.url);
    const documentId = searchParams.get('id');

    if (!documentId) {
      return NextResponse.json(
        { error: '缺少文档ID' },
        { status: 400 }
      );
    }

    const store = await getDocumentStore();
    const vectorDB = getVectorDB();

    // 删除文档元数据
    await store.deleteDocument(documentId);

    // 删除向量数据（标记删除）
    await vectorDB.deleteDocument(documentId);

    return NextResponse.json({
      success: true,
      message: '文档已删除',
    });
  } catch (error: any) {
    console.error('[Documents API] 删除文档失败:', error);
    return NextResponse.json(
      { error: '删除文档失败' },
      { status: 500 }
    );
  }
}
