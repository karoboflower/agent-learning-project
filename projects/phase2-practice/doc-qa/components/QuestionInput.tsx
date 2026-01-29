'use client';

import { useState } from 'react';

interface QuestionInputProps {
  onAsk: (question: string) => void;
  disabled?: boolean;
}

export default function QuestionInput({ onAsk, disabled }: QuestionInputProps) {
  const [question, setQuestion] = useState('');

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (question.trim() && !disabled) {
      onAsk(question.trim());
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <div className="relative">
        <textarea
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入你的问题... (Enter发送, Shift+Enter换行)"
          rows={3}
          disabled={disabled}
          className="w-full px-4 py-3 pr-24 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none disabled:bg-gray-100 disabled:cursor-not-allowed"
          maxLength={500}
        />
        <div className="absolute bottom-3 right-3 flex items-center gap-2">
          <span className="text-xs text-gray-400">
            {question.length}/500
          </span>
          <button
            type="submit"
            disabled={!question.trim() || disabled}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors font-medium text-sm"
          >
            {disabled ? '...' : '提问'}
          </button>
        </div>
      </div>

      <div className="mt-2 flex flex-wrap gap-2">
        <p className="text-xs text-gray-500">示例问题:</p>
        {[
          '文档的主要内容是什么？',
          '有哪些关键要点？',
          '如何理解这个概念？',
        ].map((example, index) => (
          <button
            key={index}
            type="button"
            onClick={() => setQuestion(example)}
            disabled={disabled}
            className="text-xs px-2 py-1 bg-gray-100 text-gray-600 rounded hover:bg-gray-200 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {example}
          </button>
        ))}
      </div>
    </form>
  );
}
