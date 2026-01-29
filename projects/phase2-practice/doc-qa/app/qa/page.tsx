'use client';

import { useState } from 'react';
import DocumentUpload from '@/components/DocumentUpload';
import DocumentList from '@/components/DocumentList';
import QuestionInput from '@/components/QuestionInput';
import AnswerDisplay from '@/components/AnswerDisplay';
import type { QARecord } from '@/lib/types';

export default function QAPage() {
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [currentQA, setCurrentQA] = useState<QARecord | null>(null);
  const [loading, setLoading] = useState(false);
  const [history, setHistory] = useState<QARecord[]>([]);
  const [showHistory, setShowHistory] = useState(false);

  const handleUploadSuccess = () => {
    // 刷新文档列表
    setRefreshTrigger((prev) => prev + 1);
  };

  const handleAsk = async (question: string) => {
    setLoading(true);
    setCurrentQA(null);

    try {
      const response = await fetch('/api/ask', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ question }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || '提问失败');
      }

      setCurrentQA(data);
      setHistory((prev) => [data, ...prev]);
    } catch (error: any) {
      alert(error.message || '提问失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const loadHistoryItem = (qa: QARecord) => {
    setCurrentQA(qa);
    setShowHistory(false);
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      {/* Header */}
      <div className="bg-white border-b border-gray-200 shadow-sm">
        <div className="container mx-auto px-4 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900">
                📚 文档问答 Agent
              </h1>
              <p className="text-gray-600 mt-1">
                上传文档，智能问答
              </p>
            </div>
            <a
              href="/"
              className="px-4 py-2 text-gray-600 hover:text-gray-900 transition-colors"
            >
              ← 返回首页
            </a>
          </div>
        </div>
      </div>

      <div className="container mx-auto px-4 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* 左侧：文档管理 */}
          <div className="lg:col-span-1">
            <div className="sticky top-8 space-y-6">
              {/* 文档上传 */}
              <div className="bg-white rounded-lg shadow-lg p-6">
                <h2 className="text-xl font-bold text-gray-900 mb-4">
                  上传文档
                </h2>
                <DocumentUpload onUploadSuccess={handleUploadSuccess} />
              </div>

              {/* 文档列表 */}
              <div className="bg-white rounded-lg shadow-lg p-6">
                <DocumentList refreshTrigger={refreshTrigger} />
              </div>
            </div>
          </div>

          {/* 右侧：问答界面 */}
          <div className="lg:col-span-2 space-y-6">
            {/* 问答输入 */}
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-xl font-bold text-gray-900 mb-4">
                提出问题
              </h2>
              <QuestionInput onAsk={handleAsk} disabled={loading} />
            </div>

            {/* 答案显示 */}
            <div>
              <AnswerDisplay qaRecord={currentQA} loading={loading} />
            </div>

            {/* 历史记录 */}
            {history.length > 0 && (
              <div className="bg-white rounded-lg shadow-lg p-6">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-xl font-bold text-gray-900">
                    历史记录 ({history.length})
                  </h2>
                  <button
                    onClick={() => setShowHistory(!showHistory)}
                    className="text-sm text-blue-600 hover:text-blue-700"
                  >
                    {showHistory ? '收起' : '展开'}
                  </button>
                </div>

                {showHistory && (
                  <div className="space-y-3">
                    {history.map((qa) => (
                      <button
                        key={qa.id}
                        onClick={() => loadHistoryItem(qa)}
                        className="w-full text-left p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
                      >
                        <p className="text-sm font-medium text-gray-900 mb-1">
                          {qa.question}
                        </p>
                        <p className="text-xs text-gray-500 truncate">
                          {qa.answer}
                        </p>
                        <p className="text-xs text-gray-400 mt-2">
                          {new Date(qa.timestamp).toLocaleString('zh-CN')}
                        </p>
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Footer */}
      <div className="mt-16 py-8 bg-white border-t border-gray-200">
        <div className="container mx-auto px-4 text-center text-gray-500 text-sm">
          <p>Task 2.2 - 文档问答Agent (RAG系统)</p>
          <p className="mt-2">
            Next.js + LangChain.js + HNSWLib + OpenAI
          </p>
        </div>
      </div>
    </main>
  );
}
