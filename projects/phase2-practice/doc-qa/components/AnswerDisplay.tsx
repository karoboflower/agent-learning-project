'use client';

import type { QARecord } from '@/lib/types';

interface AnswerDisplayProps {
  qaRecord: QARecord | null;
  loading?: boolean;
}

export default function AnswerDisplay({ qaRecord, loading }: AnswerDisplayProps) {
  if (loading) {
    return (
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <div className="flex items-center gap-3">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
          <p className="text-gray-600">AI正在思考...</p>
        </div>
      </div>
    );
  }

  if (!qaRecord) {
    return (
      <div className="text-center py-12 bg-gray-50 rounded-lg">
        <div className="text-6xl mb-4">💬</div>
        <p className="text-gray-600 text-lg">在上方输入问题开始提问</p>
        <p className="text-gray-500 text-sm mt-2">
          AI会根据你上传的文档来回答问题
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* 问题 */}
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div className="flex items-start gap-3">
          <div className="text-2xl">❓</div>
          <div className="flex-1">
            <p className="text-sm font-medium text-blue-900 mb-1">你的问题</p>
            <p className="text-gray-800">{qaRecord.question}</p>
          </div>
        </div>
      </div>

      {/* 答案 */}
      <div className="bg-white border border-gray-200 rounded-lg p-6">
        <div className="flex items-start gap-3 mb-4">
          <div className="text-2xl">🤖</div>
          <div className="flex-1">
            <p className="text-sm font-medium text-gray-900 mb-2">AI的回答</p>
            <div className="prose prose-sm max-w-none text-gray-800 whitespace-pre-wrap">
              {qaRecord.answer}
            </div>
          </div>
        </div>

        {/* 来源引用 */}
        {qaRecord.sources && qaRecord.sources.length > 0 && (
          <div className="mt-6 pt-6 border-t border-gray-200">
            <p className="text-sm font-medium text-gray-900 mb-3">
              📚 参考来源 ({qaRecord.sources.length})
            </p>
            <div className="space-y-3">
              {qaRecord.sources.map((source, index) => (
                <div
                  key={index}
                  className="bg-gray-50 rounded-lg p-4 border border-gray-200"
                >
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <span className="text-xs font-semibold px-2 py-1 bg-blue-100 text-blue-800 rounded">
                        来源 {index + 1}
                      </span>
                      <span className="text-sm text-gray-700 font-medium">
                        {source.documentName}
                      </span>
                    </div>
                    <span className="text-xs text-gray-500">
                      相关度: {(source.score * 100).toFixed(0)}%
                    </span>
                  </div>
                  <p className="text-sm text-gray-600 leading-relaxed">
                    {source.content}
                  </p>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* 时间戳 */}
        <div className="mt-4 text-xs text-gray-400">
          {new Date(qaRecord.timestamp).toLocaleString('zh-CN')}
        </div>
      </div>
    </div>
  );
}
