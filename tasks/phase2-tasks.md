# 阶段二：实践入门（3-4周）

## 任务2.1：前端Agent应用 - 智能代码助手（Week 1-2）

### 项目目标
构建一个基于React/Vue的智能代码助手Agent，能够进行代码审查、重构建议和技术栈选择。

### 功能需求
- [ ] 代码审查和建议
- [ ] 代码重构建议
- [ ] 技术栈选择建议
- [ ] 代码质量评估

### 技术栈
- React/Vue 3 + TypeScript
- LangChain.js
- OpenAI API / Claude API
- Web Workers（后台处理）

### 实现步骤

#### 2.1.1 项目初始化（Day 1）

**创建项目**
- [ ] 使用Vite创建React/Vue项目
- [ ] 配置TypeScript
- [ ] 配置TailwindCSS（可选）
- [ ] 配置项目结构

**安装依赖**
- [ ] 安装LangChain.js
- [ ] 安装OpenAI SDK
- [ ] 安装其他必要依赖

**输出**：
- `projects/phase2-practice/code-assistant/` - 项目目录

#### 2.1.2 集成LangChain.js（Day 2）

**基础配置**
- [ ] 配置LangChain.js
- [ ] 创建LLM实例
- [ ] 配置API密钥管理

**实现基础Agent**
- [ ] 创建简单的Agent类
- [ ] 实现基础的对话功能
- [ ] 测试Agent运行

**输出**：
- `projects/phase2-practice/code-assistant/src/agent/` - Agent代码

#### 2.1.3 实现代码分析Prompt（Day 3）

**代码审查Prompt**
- [ ] 设计代码审查Prompt模板
- [ ] 实现代码审查功能
- [ ] 测试代码审查效果

**代码重构Prompt**
- [ ] 设计代码重构Prompt模板
- [ ] 实现代码重构建议功能
- [ ] 测试重构建议效果

**技术栈选择Prompt**
- [ ] 设计技术栈选择Prompt模板
- [ ] 实现技术栈建议功能
- [ ] 测试技术栈建议效果

**输出**：
- `projects/phase2-practice/code-assistant/src/prompts/` - Prompt模板
- `docs/learning-notes/code-assistant-prompts.md` - Prompt设计文档

#### 2.1.4 实现核心功能（Day 4-5）

**代码审查功能**
- [ ] 实现代码输入界面
- [ ] 实现代码审查逻辑
- [ ] 实现结果展示界面
- [ ] 添加错误处理

**代码重构功能**
- [ ] 实现代码重构逻辑
- [ ] 实现重构建议展示
- [ ] 添加代码对比功能

**技术栈选择功能**
- [ ] 实现需求输入界面
- [ ] 实现技术栈分析逻辑
- [ ] 实现建议展示界面

**输出**：
- `projects/phase2-practice/code-assistant/src/components/` - UI组件
- `projects/phase2-practice/code-assistant/src/services/` - 业务逻辑

#### 2.1.5 优化和测试（Day 6-7）

**性能优化**
- [ ] 使用Web Workers处理后台任务
- [ ] 实现请求缓存
- [ ] 优化UI响应速度

**用户体验优化**
- [ ] 添加加载状态
- [ ] 添加错误提示
- [ ] 优化界面设计

**测试**
- [ ] 编写单元测试
- [ ] 编写集成测试
- [ ] 进行用户测试

**输出**：
- 完整的智能代码助手项目
- `docs/projects/code-assistant.md` - 项目文档

---

## 任务2.2：前端Agent应用 - 文档问答Agent（Week 2-3）

### 项目目标
构建一个RAG（检索增强生成）文档问答系统，支持文档上传、向量化存储和智能问答。

### 功能需求
- [ ] 文档上传（PDF、Markdown）
- [ ] 文档解析和分块
- [ ] 向量化存储
- [ ] RAG检索
- [ ] 智能问答

### 技术栈
- Next.js / Nuxt.js
- LangChain.js
- 向量数据库（Pinecone / Weaviate / Qdrant）
- Embedding模型（OpenAI / 本地模型）

### 实现步骤

#### 2.2.1 项目初始化（Day 1）

**创建项目**
- [ ] 使用Next.js/Nuxt.js创建项目
- [ ] 配置TypeScript
- [ ] 配置项目结构

**安装依赖**
- [ ] 安装LangChain.js
- [ ] 安装向量数据库SDK
- [ ] 安装文档解析库（pdf-parse等）
- [ ] 安装其他必要依赖

**输出**：
- `projects/phase2-practice/doc-qa/` - 项目目录

#### 2.2.2 集成向量数据库（Day 2）

**选择向量数据库**
- [ ] 对比Pinecone、Weaviate、Qdrant
- [ ] 选择适合的向量数据库
- [ ] 配置向量数据库连接

**实现向量存储**
- [ ] 实现文档向量化
- [ ] 实现向量存储功能
- [ ] 实现向量检索功能

**输出**：
- `projects/phase2-practice/doc-qa/src/services/vector-db.ts` - 向量数据库服务

#### 2.2.3 实现文档处理（Day 3）

**文档上传**
- [ ] 实现文件上传界面
- [ ] 实现文件类型验证
- [ ] 实现文件大小限制

**文档解析**
- [ ] 实现PDF解析
- [ ] 实现Markdown解析
- [ ] 实现文本提取

**文档分块**
- [ ] 实现文本分块策略
- [ ] 实现分块大小控制
- [ ] 实现分块重叠处理

**输出**：
- `projects/phase2-practice/doc-qa/src/services/document-processor.ts` - 文档处理服务

#### 2.2.4 实现RAG功能（Day 4-5）

**Embedding生成**
- [ ] 配置Embedding模型
- [ ] 实现文档Embedding生成
- [ ] 实现查询Embedding生成

**检索实现**
- [ ] 实现相似度检索
- [ ] 实现Top-K检索
- [ ] 实现检索结果排序

**问答实现**
- [ ] 设计RAG Prompt模板
- [ ] 实现上下文构建
- [ ] 实现问答生成

**输出**：
- `projects/phase2-practice/doc-qa/src/services/rag.ts` - RAG服务
- `projects/phase2-practice/doc-qa/src/prompts/rag-prompt.ts` - RAG Prompt

#### 2.2.5 实现UI界面（Day 6）

**文档管理界面**
- [ ] 实现文档列表展示
- [ ] 实现文档删除功能
- [ ] 实现文档状态显示

**问答界面**
- [ ] 实现问答输入框
- [ ] 实现问答结果展示
- [ ] 实现引用来源显示
- [ ] 实现历史记录

**输出**：
- `projects/phase2-practice/doc-qa/src/components/` - UI组件

#### 2.2.6 优化和测试（Day 7）

**性能优化**
- [ ] 优化向量检索速度
- [ ] 实现结果缓存
- [ ] 优化文档处理速度

**用户体验优化**
- [ ] 添加加载状态
- [ ] 添加进度显示
- [ ] 优化界面设计

**测试**
- [ ] 编写单元测试
- [ ] 编写集成测试
- [ ] 进行用户测试

**输出**：
- 完整的文档问答Agent项目
- `docs/projects/doc-qa.md` - 项目文档

---

## 任务2.3：Go后端Agent服务 - API Agent服务（Week 3-4）

### 项目目标
构建一个Go后端的Agent API服务，支持多Agent协作、任务调度和工具调用。

### 功能需求
- [ ] Agent服务接口
- [ ] 任务调度和分发
- [ ] Agent状态管理
- [ ] 工具调用机制
- [ ] RESTful API接口

### 技术栈
- Go + Gin / Fiber
- OpenAI Go SDK
- Redis（状态存储）
- PostgreSQL（任务历史）

### 实现步骤

#### 2.3.1 项目初始化（Day 1）

**创建Go项目**
- [ ] 初始化Go模块
- [ ] 配置项目结构
- [ ] 配置Go开发工具

**安装依赖**
- [ ] 安装Gin/Fiber框架
- [ ] 安装OpenAI Go SDK
- [ ] 安装Redis客户端
- [ ] 安装PostgreSQL驱动
- [ ] 安装其他必要依赖

**输出**：
- `projects/phase2-practice/go-agent-api/` - 项目目录

#### 2.3.2 实现Agent服务接口（Day 2）

**定义Agent接口**
- [ ] 设计Agent服务接口
- [ ] 定义Agent结构体
- [ ] 定义任务结构体

**实现基础Agent**
- [ ] 实现Agent初始化
- [ ] 实现Agent运行方法
- [ ] 实现Agent停止方法

**输出**：
- `projects/phase2-practice/go-agent-api/internal/agent/` - Agent代码

#### 2.3.3 实现任务调度器（Day 3）

**任务队列**
- [ ] 设计任务队列结构
- [ ] 实现任务入队
- [ ] 实现任务出队

**任务调度**
- [ ] 实现任务调度算法
- [ ] 实现任务优先级
- [ ] 实现任务分发

**输出**：
- `projects/phase2-practice/go-agent-api/internal/scheduler/` - 调度器代码

#### 2.3.4 实现Agent状态管理（Day 4）

**状态存储**
- [ ] 设计状态数据结构
- [ ] 实现Redis状态存储
- [ ] 实现状态读取和更新

**状态同步**
- [ ] 实现状态同步机制
- [ ] 实现状态持久化
- [ ] 实现状态恢复

**输出**：
- `projects/phase2-practice/go-agent-api/internal/state/` - 状态管理代码

#### 2.3.5 实现工具调用机制（Day 5）

**工具接口**
- [ ] 设计工具接口规范
- [ ] 定义工具注册机制
- [ ] 实现工具发现机制

**工具实现**
- [ ] 实现搜索工具
- [ ] 实现代码工具
- [ ] 实现文件操作工具

**工具调用**
- [ ] 实现工具选择逻辑
- [ ] 实现工具执行
- [ ] 实现结果处理

**输出**：
- `projects/phase2-practice/go-agent-api/internal/tools/` - 工具代码

#### 2.3.6 实现API接口（Day 6）

**RESTful API**
- [ ] 实现Agent创建接口
- [ ] 实现任务提交接口
- [ ] 实现状态查询接口
- [ ] 实现结果获取接口

**API文档**
- [ ] 使用Swagger生成API文档
- [ ] 编写API使用示例
- [ ] 编写API测试用例

**输出**：
- `projects/phase2-practice/go-agent-api/internal/api/` - API代码
- `docs/api/go-agent-api.md` - API文档

#### 2.3.7 集成数据库和优化（Day 7）

**数据库集成**
- [ ] 设计数据库表结构
- [ ] 实现任务历史存储
- [ ] 实现结果存储

**性能优化**
- [ ] 优化并发处理
- [ ] 实现连接池
- [ ] 优化数据库查询

**测试**
- [ ] 编写单元测试
- [ ] 编写集成测试
- [ ] 进行压力测试

**输出**：
- 完整的Go Agent API服务
- `docs/projects/go-agent-api.md` - 项目文档

---

## 阶段二总结

### 完成标准
- [ ] 完成智能代码助手项目
- [ ] 完成文档问答Agent项目
- [ ] 完成Go后端Agent API服务
- [ ] 所有项目都有完整文档
- [ ] 所有项目都经过测试

### 下一步
进入阶段三：进阶应用，开始构建多Agent系统和工具生态。
