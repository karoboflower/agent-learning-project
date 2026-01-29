'use client';

import { useEffect, useState } from 'react';
import type { Document } from '@/lib/types';

interface DocumentListProps {
  refreshTrigger?: number;
}

export default function DocumentList({ refreshTrigger }: DocumentListProps) {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [deleting, setDeleting] = useState<string | null>(null);

  const fetchDocuments = async () => {
    try {
      const response = await fetch('/api/documents');
      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || '获取文档列表失败');
      }

      setDocuments(data.documents);
      setError('');
    } catch (err: any) {
      setError(err.message || '获取文档列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDocuments();
  }, [refreshTrigger]);

  const handleDelete = async (documentId: string) => {
    if (!confirm('确定要删除这个文档吗？')) {
      return;
    }

    setDeleting(documentId);

    try {
      const response = await fetch(`/api/documents?id=${documentId}`, {
        method: 'DELETE',
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || '删除失败');
      }

      // 刷新列表
      await fetchDocuments();
    } catch (err: any) {
      alert(err.message || '删除失败');
    } finally {
      setDeleting(null);
    }
  };

  const getStatusBadge = (status: Document['status']) => {
    switch (status) {
      case 'ready':
        return (
          <span className="px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
            ✓ 就绪
          </span>
        );
      case 'processing':
        return (
          <span className="px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800 animate-pulse">
            ⟳ 处理中
          </span>
        );
      case 'error':
        return (
          <span className="px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
            ✗ 错误
          </span>
        );
    }
  };

  const formatFileSize = (bytes: number) => {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  };

  const formatDate = (date: Date) => {
    return new Date(date).toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="text-center py-8">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p className="mt-2 text-gray-600">加载中...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
        <p className="text-red-800 font-medium">加载失败</p>
        <p className="text-red-600 text-sm mt-1">{error}</p>
        <button
          onClick={fetchDocuments}
          className="mt-2 text-sm text-red-600 underline hover:text-red-700"
        >
          重试
        </button>
      </div>
    );
  }

  if (documents.length === 0) {
    return (
      <div className="text-center py-12 bg-gray-50 rounded-lg">
        <div className="text-6xl mb-4">📄</div>
        <p className="text-gray-600 text-lg">还没有上传任何文档</p>
        <p className="text-gray-500 text-sm mt-2">
          上传文档后开始提问
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-900">
          文档列表 ({documents.length})
        </h3>
        <button
          onClick={fetchDocuments}
          className="text-sm text-blue-600 hover:text-blue-700 underline"
        >
          刷新
        </button>
      </div>

      <div className="space-y-3">
        {documents.map((doc) => (
          <div
            key={doc.id}
            className="bg-white border border-gray-200 rounded-lg p-4 hover:shadow-md transition-shadow"
          >
            <div className="flex items-start justify-between">
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-2xl">
                    {doc.type === 'pdf' ? '📕' : doc.type === 'markdown' ? '📘' : '📄'}
                  </span>
                  <h4 className="text-sm font-semibold text-gray-900 truncate">
                    {doc.name}
                  </h4>
                </div>

                <div className="flex flex-wrap items-center gap-3 text-xs text-gray-500">
                  <span>{formatFileSize(doc.size)}</span>
                  <span>•</span>
                  <span>{formatDate(doc.uploadedAt)}</span>
                  {doc.chunks && (
                    <>
                      <span>•</span>
                      <span>{doc.chunks} 个文档块</span>
                    </>
                  )}
                </div>

                {doc.error && (
                  <p className="mt-2 text-xs text-red-600">
                    错误: {doc.error}
                  </p>
                )}
              </div>

              <div className="flex items-center gap-2 ml-4">
                {getStatusBadge(doc.status)}
                <button
                  onClick={() => handleDelete(doc.id)}
                  disabled={deleting === doc.id}
                  className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors disabled:opacity-50"
                  title="删除"
                >
                  {deleting === doc.id ? (
                    <span className="animate-spin">⟳</span>
                  ) : (
                    '🗑️'
                  )}
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
