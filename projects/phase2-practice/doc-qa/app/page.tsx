export default function Home() {
  return (
    <main className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="container mx-auto px-4 py-16">
        <div className="text-center mb-12">
          <h1 className="text-5xl font-bold text-gray-900 mb-4">
            📚 文档问答 Agent
          </h1>
          <p className="text-xl text-gray-600 mb-8">
            基于RAG的智能文档问答系统
          </p>
          <p className="text-lg text-gray-500">
            上传文档，AI帮你快速找到答案
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
          {/* 文档上传 */}
          <div className="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition-shadow">
            <div className="text-4xl mb-4">📤</div>
            <h2 className="text-2xl font-bold text-gray-900 mb-3">
              文档上传
            </h2>
            <p className="text-gray-600">
              支持PDF和Markdown格式文档上传，自动解析和向量化
            </p>
          </div>

          {/* 智能检索 */}
          <div className="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition-shadow">
            <div className="text-4xl mb-4">🔍</div>
            <h2 className="text-2xl font-bold text-gray-900 mb-3">
              智能检索
            </h2>
            <p className="text-gray-600">
              基于向量相似度的智能检索，快速找到相关内容
            </p>
          </div>

          {/* 智能问答 */}
          <div className="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition-shadow">
            <div className="text-4xl mb-4">💬</div>
            <h2 className="text-2xl font-bold text-gray-900 mb-3">
              智能问答
            </h2>
            <p className="text-gray-600">
              AI理解你的问题，从文档中提取准确答案
            </p>
          </div>
        </div>

        <div className="mt-12 text-center">
          <a
            href="/qa"
            className="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg text-lg font-semibold hover:bg-blue-700 transition-colors"
          >
            开始使用 →
          </a>
        </div>

        <div className="mt-16 text-center text-gray-500 text-sm">
          <p>Task 2.2 - 文档问答Agent (RAG系统)</p>
          <p className="mt-2">Next.js + LangChain.js + Vector Database</p>
        </div>
      </div>
    </main>
  );
}
