'use client';

import { useState, useRef } from 'react';
import type { Document } from '@/lib/types';

interface DocumentUploadProps {
  onUploadSuccess?: (document: Document) => void;
}

export default function DocumentUpload({ onUploadSuccess }: DocumentUploadProps) {
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');
  const [dragActive, setDragActive] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleUpload = async (file: File) => {
    setUploading(true);
    setError('');

    try {
      const formData = new FormData();
      formData.append('file', file);

      const response = await fetch('/api/upload', {
        method: 'POST',
        body: formData,
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || '上传失败');
      }

      if (onUploadSuccess) {
        onUploadSuccess(data.document);
      }
    } catch (err: any) {
      setError(err.message || '上传失败，请重试');
    } finally {
      setUploading(false);
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      handleUpload(file);
    }
  };

  const handleDrag = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    const file = e.dataTransfer.files?.[0];
    if (file) {
      handleUpload(file);
    }
  };

  return (
    <div className="w-full">
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          dragActive
            ? 'border-blue-500 bg-blue-50'
            : 'border-gray-300 hover:border-gray-400'
        } ${uploading ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
        onDragEnter={handleDrag}
        onDragLeave={handleDrag}
        onDragOver={handleDrag}
        onDrop={handleDrop}
        onClick={() => !uploading && fileInputRef.current?.click()}
      >
        <input
          ref={fileInputRef}
          type="file"
          className="hidden"
          accept=".pdf,.md,.txt"
          onChange={handleFileChange}
          disabled={uploading}
        />

        <div className="text-6xl mb-4">📤</div>

        {uploading ? (
          <div>
            <div className="text-lg font-semibold text-gray-700 mb-2">
              上传中...
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2 mb-4">
              <div className="bg-blue-600 h-2 rounded-full animate-pulse w-1/2"></div>
            </div>
            <p className="text-sm text-gray-500">
              正在处理文档，请稍候...
            </p>
          </div>
        ) : (
          <div>
            <div className="text-lg font-semibold text-gray-700 mb-2">
              点击上传或拖拽文件到这里
            </div>
            <p className="text-sm text-gray-500 mb-4">
              支持 PDF、Markdown (.md)、纯文本 (.txt) 格式
            </p>
            <p className="text-xs text-gray-400">
              最大文件大小: 10MB
            </p>
          </div>
        )}
      </div>

      {error && (
        <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
          <p className="text-red-800 text-sm font-medium">上传失败</p>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      )}
    </div>
  );
}
