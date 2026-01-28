#!/bin/bash

echo "🔧 修复依赖问题..."

# 1. 删除旧的依赖
echo "📦 清理旧依赖..."
rm -rf node_modules
rm -f pnpm-lock.yaml

# 2. 安装所有依赖
echo "📥 安装依赖..."
pnpm install

echo "✅ 依赖安装完成！"
echo ""
echo "现在可以运行: pnpm dev"
