# 环境配置指南

Task 1.3.3 - 依赖安装和验证的环境配置文档。

## 📋 任务完成清单

### 前端依赖
- [x] 创建React项目
- [x] 安装LangChain.js
- [x] 安装其他必要依赖
- [x] 验证安装成功

### Go依赖
- [x] 创建Go模块
- [x] 安装OpenAI Go SDK
- [x] 安装其他必要依赖
- [x] 验证安装成功

### API密钥配置
- [x] 配置环境变量示例文件
- [ ] 获取OpenAI API密钥（需要用户操作）
- [ ] 获取Anthropic API密钥（可选，需要用户操作）
- [ ] 验证API连接（需要配置密钥后）

## 🛠️ 环境要求

### 系统要求
- **操作系统**: macOS, Linux, Windows
- **Node.js**: >= 18.0.0
- **Go**: >= 1.21
- **pnpm**: >= 8.0.0 (推荐) 或 npm >= 9.0.0

### 版本检查

```bash
# 检查Node.js版本
node --version

# 检查Go版本
go version

# 检查pnpm版本
pnpm --version
```

## 📦 前端环境配置

### 1. 进入前端目录

```bash
cd projects/frontend
```

### 2. 安装依赖

```bash
# 使用pnpm（推荐）
pnpm install

# 或使用npm
npm install
```

### 3. 配置环境变量

```bash
# 复制环境变量示例
cp .env.example .env

# 编辑.env文件
vim .env
```

配置内容：
```env
VITE_OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
VITE_ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxxxxxxxxxxxxxx
```

### 4. 验证安装

```bash
# TypeScript类型检查
pnpm exec tsc --noEmit

# 启动开发服务器
pnpm dev
```

访问 http://localhost:3000，点击"测试 LangChain 配置"按钮验证。

### 前端依赖列表

| 依赖 | 版本 | 说明 |
|------|------|------|
| react | 18.2.0 | React框架 |
| react-dom | 18.2.0 | React DOM |
| langchain | 0.1.0 | LangChain核心库 |
| @langchain/openai | 0.0.19 | OpenAI集成 |
| @langchain/anthropic | 0.1.0 | Anthropic集成 |
| typescript | 5.3.0 | TypeScript |
| vite | 5.0.0 | 构建工具 |

## 🔧 Go环境配置

### 1. 进入后端目录

```bash
cd projects/backend
```

### 2. 下载Go依赖

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy

# 验证依赖
go mod verify
```

### 3. 配置环境变量

```bash
# 复制环境变量示例
cp .env.example .env

# 编辑.env文件
vim .env
```

配置内容：
```env
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
ANTHROPIC_API_KEY=sk-ant-xxxxxxxxxxxxxxxxxxxxxxxx
PORT=8080
```

### 4. 验证安装

```bash
# 基础验证（不需要API密钥）
go run main.go

# API连接测试（需要API密钥）
go run test_api.go
```

### Go依赖列表

| 依赖 | 版本 | 说明 |
|------|------|------|
| github.com/sashabaranov/go-openai | v1.17.0 | OpenAI Go SDK |
| github.com/joho/godotenv | v1.5.1 | 环境变量加载 |

## 🔑 API密钥获取

### OpenAI API密钥

1. **访问官网**: https://platform.openai.com/
2. **注册/登录**: 使用邮箱或Google账号
3. **进入API Keys页面**: https://platform.openai.com/api-keys
4. **创建新密钥**:
   - 点击 "Create new secret key"
   - 输入密钥名称（如: "agent-learning-dev"）
   - 复制生成的密钥（只显示一次！）
5. **配置到.env文件**

**注意事项**:
- API密钥以 `sk-` 开头
- 密钥只显示一次，务必保存
- 不要将密钥提交到Git
- 建议设置使用限额

### Anthropic API密钥（可选）

1. **访问官网**: https://console.anthropic.com/
2. **注册/登录**: 使用邮箱
3. **进入API Keys页面**: https://console.anthropic.com/settings/keys
4. **创建新密钥**:
   - 点击 "Create Key"
   - 输入密钥名称
   - 复制生成的密钥
5. **配置到.env文件**

**注意事项**:
- API密钥以 `sk-ant-` 开头
- Claude模型需要单独的API密钥
- 可能需要加入waitlist

## ✅ 验证API连接

### 前端验证

1. 启动开发服务器：
```bash
cd projects/frontend
pnpm dev
```

2. 访问 http://localhost:3000
3. 点击"测试 LangChain 配置"按钮
4. 查看控制台输出

**预期结果**:
```
OpenAI Key configured: true
Anthropic Key configured: true
LangChain OpenAI: [Function]
LangChain Anthropic: [Function]
✅ LangChain.js 配置成功！
```

### 后端验证

1. 运行验证脚本：
```bash
cd projects/backend
go run main.go
```

**预期输出**:
```
🚀 Task 1.3.3 - Go后端依赖验证
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ 验证清单:
   [✓] Go模块初始化成功
   [✓] go.mod文件创建完成
   [✓] OpenAI API密钥已配置
   [✓] OpenAI客户端初始化成功
   [✓] godotenv包安装成功
```

2. 运行API测试：
```bash
go run test_api.go
```

**预期输出**:
```
🧪 测试OpenAI API连接
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ API密钥已加载

测试1: 列出可用模型...
✓ 成功! 找到 XX 个模型

测试2: 测试Chat Completion API...
✓ 成功!
回复: Hello from Go!

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✨ 所有测试通过！OpenAI Go SDK工作正常
```

## 🐛 常见问题

### 1. Node.js依赖安装失败

**问题**: `npm install` 报错

**解决方案**:
```bash
# 清理缓存
npm cache clean --force

# 删除node_modules和lock文件
rm -rf node_modules package-lock.json

# 重新安装
npm install

# 或使用pnpm
pnpm install
```

### 2. Go依赖下载失败

**问题**: `go mod download` 超时

**解决方案**:
```bash
# 设置GOPROXY代理
export GOPROXY=https://goproxy.cn,direct

# 或永久设置
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.bashrc
source ~/.bashrc

# 重新下载
go mod download
```

### 3. API密钥配置错误

**问题**: "API key not found"

**检查步骤**:
1. 确认 `.env` 文件存在
2. 确认密钥格式正确（以`sk-`开头）
3. 确认没有多余空格
4. 前端：确认使用 `VITE_` 前缀
5. 重启开发服务器

### 4. TypeScript类型错误

**问题**: 类型检查失败

**解决方案**:
```bash
# 清理缓存
rm -rf node_modules/.cache

# 重新安装类型定义
pnpm install

# 检查tsconfig.json配置
```

### 5. Vite启动失败

**问题**: 端口被占用

**解决方案**:
```bash
# 查找占用端口的进程
lsof -i :3000

# 杀死进程
kill -9 <PID>

# 或修改端口
# 编辑 vite.config.ts，更改port配置
```

## 📊 环境检查脚本

创建检查脚本 `check-env.sh`:

```bash
#!/bin/bash

echo "🔍 检查开发环境..."
echo

# 检查Node.js
if command -v node &> /dev/null; then
    echo "✅ Node.js: $(node --version)"
else
    echo "❌ Node.js 未安装"
fi

# 检查Go
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
else
    echo "❌ Go 未安装"
fi

# 检查pnpm
if command -v pnpm &> /dev/null; then
    echo "✅ pnpm: $(pnpm --version)"
else
    echo "⚠️  pnpm 未安装（可选）"
fi

echo
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "检查完成！"
```

运行：
```bash
chmod +x check-env.sh
./check-env.sh
```

## 🎯 下一步

配置完成后，可以：

1. **前端开发**:
   ```bash
   cd projects/frontend
   pnpm dev
   ```

2. **后端开发**:
   ```bash
   cd projects/backend
   go run main.go
   ```

3. **运行测试**:
   - 前端: 访问 http://localhost:3000 测试
   - 后端: `go run test_api.go`

4. **开始开发Agent应用**！

## 📚 相关文档

- [前端项目README](../projects/frontend/README.md)
- [后端项目README](../projects/backend/README.md)
- [LangChain.js文档](https://js.langchain.com/)
- [OpenAI Go SDK文档](https://github.com/sashabaranov/go-openai)

---

**创建日期**: 2026-01-28
**最后更新**: 2026-01-28
**任务来源**: phase1-tasks.md - 1.3.3 依赖安装和验证
