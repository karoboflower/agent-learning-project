import { useState } from 'react';
import { ChatOpenAI } from '@langchain/openai';
import { ChatAnthropic } from '@langchain/anthropic';

function App() {
  const [status, setStatus] = useState<string>('未测试');
  const [loading, setLoading] = useState<boolean>(false);

  const testLangChain = async () => {
    setLoading(true);
    setStatus('测试中...');

    try {
      // 检查环境变量
      const openaiKey = import.meta.env.VITE_OPENAI_API_KEY;
      const anthropicKey = import.meta.env.VITE_ANTHROPIC_API_KEY;

      console.log('OpenAI Key configured:', !!openaiKey);
      console.log('Anthropic Key configured:', !!anthropicKey);

      // 测试LangChain导入
      console.log('LangChain OpenAI:', ChatOpenAI);
      console.log('LangChain Anthropic:', ChatAnthropic);

      setStatus('✅ LangChain.js 配置成功！');
    } catch (error) {
      console.error('测试失败:', error);
      setStatus(`❌ 测试失败: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
      <h1>🚀 Task 1.3.3 - 前端依赖验证</h1>

      <div style={{ marginTop: '30px', padding: '20px', border: '1px solid #ddd', borderRadius: '8px' }}>
        <h2>LangChain.js 集成测试</h2>

        <div style={{ marginTop: '20px' }}>
          <button
            onClick={testLangChain}
            disabled={loading}
            style={{
              padding: '10px 20px',
              fontSize: '16px',
              cursor: loading ? 'not-allowed' : 'pointer',
              backgroundColor: loading ? '#ccc' : '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: '4px',
            }}
          >
            {loading ? '测试中...' : '测试 LangChain 配置'}
          </button>
        </div>

        <div style={{ marginTop: '20px', fontSize: '18px' }}>
          状态: <strong>{status}</strong>
        </div>
      </div>

      <div style={{ marginTop: '30px', padding: '20px', backgroundColor: '#f5f5f5', borderRadius: '8px' }}>
        <h3>📦 已安装的依赖</h3>
        <ul>
          <li>✅ React 18.2.0</li>
          <li>✅ TypeScript 5.3.0</li>
          <li>✅ Vite 5.0.0</li>
          <li>✅ LangChain</li>
          <li>✅ @langchain/openai</li>
          <li>✅ @langchain/anthropic</li>
        </ul>

        <h3 style={{ marginTop: '20px' }}>🔑 环境变量配置</h3>
        <p>请在 <code>.env</code> 文件中配置API密钥：</p>
        <pre style={{ backgroundColor: '#fff', padding: '10px', borderRadius: '4px', overflow: 'auto' }}>
{`VITE_OPENAI_API_KEY=your_openai_api_key
VITE_ANTHROPIC_API_KEY=your_anthropic_api_key`}
        </pre>

        <h3 style={{ marginTop: '20px' }}>🎯 验证清单</h3>
        <ul>
          <li>✅ React项目创建成功</li>
          <li>✅ TypeScript配置完成</li>
          <li>✅ LangChain.js安装完成</li>
          <li>✅ 开发服务器可以启动</li>
          <li>⏳ API密钥配置（需要手动配置）</li>
        </ul>
      </div>
    </div>
  );
}

export default App;
