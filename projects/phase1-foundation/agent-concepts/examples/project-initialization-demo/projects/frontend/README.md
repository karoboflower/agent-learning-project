# 前端项目 - LangChain.js 集成

Task 1.3.3 前端依赖安装和验证示例。

## ✅ 已完成

- [x] 创建React项目
- [x] 安装LangChain.js
- [x] 安装其他必要依赖
- [x] 配置TypeScript
- [x] 配置Vite

## 📦 依赖列表

### 核心依赖
- React 18.2.0
- React DOM 18.2.0
- **LangChain** - LangChain核心库
- **@langchain/openai** - OpenAI集成
- **@langchain/anthropic** - Anthropic集成

### 开发依赖
- TypeScript 5.3.0
- Vite 5.0.0
- ESLint + Prettier
- @vitejs/plugin-react

## 🚀 快速开始

### 1. 安装依赖

```bash
pnpm install
```

### 2. 配置环境变量

```bash
cp .env.example .env
```

编辑 `.env` 文件，填入API密钥：
```env
VITE_OPENAI_API_KEY=your_openai_api_key_here
VITE_ANTHROPIC_API_KEY=your_anthropic_api_key_here
```

### 3. 启动开发服务器

```bash
pnpm dev
```

访问 http://localhost:3000

### 4. 验证安装

点击页面上的"测试 LangChain 配置"按钮，验证LangChain.js是否正确安装。

## 📁 项目结构

```
frontend/
├── index.html              # HTML入口
├── package.json            # 项目配置
├── tsconfig.json           # TypeScript配置
├── vite.config.ts          # Vite配置
├── .env.example            # 环��变量示例
└── src/
    ├── main.tsx            # 应用入口
    ├── App.tsx             # 主组件（验证测试）
    ├── index.css           # 全局样式
    └── vite-env.d.ts       # TypeScript类型定义
```

## 🧪 验证清单

运行以下命令验证安装：

```bash
# TypeScript类型检查
pnpm exec tsc --noEmit

# 启动开发服务器
pnpm dev

# 构建生产版本
pnpm build
```

所有命令应该成功执行！

## 📚 LangChain.js 使用示例

### 基础使用

```typescript
import { ChatOpenAI } from '@langchain/openai';

const model = new ChatOpenAI({
  openAIApiKey: import.meta.env.VITE_OPENAI_API_KEY,
  modelName: 'gpt-3.5-turbo',
});

const response = await model.invoke('Hello, LangChain!');
console.log(response);
```

### Anthropic Claude

```typescript
import { ChatAnthropic } from '@langchain/anthropic';

const model = new ChatAnthropic({
  anthropicApiKey: import.meta.env.VITE_ANTHROPIC_API_KEY,
  modelName: 'claude-3-sonnet-20240229',
});

const response = await model.invoke('Hello, Claude!');
console.log(response);
```

## 🔑 API密钥获取

### OpenAI API密钥
1. 访问 https://platform.openai.com/
2. 注册/登录账号
3. 进入 API Keys 页面
4. 创建新的API密钥

### Anthropic API密钥
1. 访问 https://console.anthropic.com/
2. 注册/登录账号
3. 进入 API Keys 页面
4. 创建新的API密钥

## 📖 相关文档

- [LangChain.js 文档](https://js.langchain.com/)
- [React 文档](https://react.dev/)
- [Vite 文档](https://vitejs.dev/)
- [TypeScript 文档](https://www.typescriptlang.org/)

---

**创建日期**: 2026-01-28
**任务来源**: phase1-tasks.md - 1.3.3 依赖安装和验证
