# 依赖问题修复指南

## 问题原因

LangChain包的版本不兼容或未正确安装。

## 一次性解决方案

### 方案1：使用修复脚本（推荐）

```bash
# 进入项目目录
cd /Users/wangchunhua/Desktop/work/project/agent-learning-project/projects/phase2-practice/code-assistant

# 给脚本执行权限
chmod +x fix-deps.sh

# 运行修复脚本
./fix-deps.sh

# 启动开发服务器
pnpm dev
```

### 方案2：手动修复

```bash
# 1. 进入项目目录
cd /Users/wangchunhua/Desktop/work/project/agent-learning-project/projects/phase2-practice/code-assistant

# 2. 清理旧依赖
rm -rf node_modules pnpm-lock.yaml

# 3. 安装依赖
pnpm install

# 4. 启动开发服务器
pnpm dev
```

### 方案3：使用npm替代pnpm

如果pnpm持续有问题，可以使用npm：

```bash
# 1. 清理
rm -rf node_modules package-lock.json

# 2. 使用npm安装
npm install

# 3. 启动
npm run dev
```

## 已修复的问题

✅ 更新package.json中的LangChain版本：
- `langchain`: ^0.1.0 → ^0.3.0
- `@langchain/core`: 新增 ^0.3.0
- `@langchain/openai`: ^0.0.19 → ^0.3.0
- `@langchain/anthropic`: ^0.1.0 → ^0.3.0
- `marked`: 新增 ^11.0.0

✅ 确保所有LangChain包版本一致

## 验证安装

安装完成后，运行以下命令验证：

```bash
# 检查langchain相关包
pnpm list | grep langchain

# 应该看到类似输出：
# @langchain/anthropic 0.3.x
# @langchain/core 0.3.x
# @langchain/openai 0.3.x
# langchain 0.3.x
```

## 如果仍然有问题

### 检查Node版本

```bash
node --version  # 应该是 v18+ 或 v20+
```

### 清除pnpm缓存

```bash
pnpm store prune
pnpm install
```

### 检查是否有代理/网络问题

```bash
# 如果在国内，可能需要设置npm镜像
npm config set registry https://registry.npmmirror.com
pnpm config set registry https://registry.npmmirror.com
```

## 环境变量配置

安装成功后，还需要配置环境变量：

```bash
# 创建.env文件
cp .env.example .env

# 编辑.env，添加API密钥
# VITE_LLM_PROVIDER=openai
# VITE_OPENAI_API_KEY=your_api_key_here
```

## 常见问题

### Q: pnpm install卡住不动？
A: 可能是网络问题，尝试：
```bash
pnpm install --no-frozen-lockfile
# 或使用npm
npm install
```

### Q: 权限错误？
A: 使用sudo（不推荐）或修复npm权限：
```bash
sudo chown -R $USER ~/.pnpm-store
```

### Q: 版本冲突？
A: 删除所有依赖重新安装：
```bash
rm -rf node_modules pnpm-lock.yaml ~/.pnpm-store
pnpm install
```

---

**最后更新**: 2026-01-27
**状态**: 已修复
