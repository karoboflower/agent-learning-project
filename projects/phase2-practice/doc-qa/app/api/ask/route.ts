import { NextRequest, NextResponse } from 'next/server';
import { getRAGService } from '@/lib/services/rag';
import { getVectorDB } from '@/lib/services/vector-db';

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { question } = body;

    if (!question || typeof question !== 'string') {
      return NextResponse.json(
        { error: '请输入问题' },
        { status: 400 }
      );
    }

    if (question.length > 500) {
      return NextResponse.json(
        { error: '问题长度不能超过500字符' },
        { status: 400 }
      );
    }

    // 初始化向量数据库
    const vectorDB = getVectorDB();
    await vectorDB.initialize();

    // 获取RAG服务并提问
    const ragService = getRAGService();
    const result = await ragService.ask(question.trim());

    return NextResponse.json(result);
  } catch (error: any) {
    console.error('[Ask API] 问答失败:', error);

    // 根据错误类型返回不同的错误消息
    if (error.message.includes('API key')) {
      return NextResponse.json(
        { error: 'OpenAI API配置错误，请检查环境变量' },
        { status: 500 }
      );
    }

    return NextResponse.json(
      { error: '问答失败，请稍后重试' },
      { status: 500 }
    );
  }
}
