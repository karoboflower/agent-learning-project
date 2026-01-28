# Task 1.3.2 项目初始化演示

本项目严格按照 `phase1-tasks.md` 中 **1.3.2 项目初始化** 的要求创建。

## ✅ 任务完成清单

### 创建项目结构
- [x] 创建项目根目录
- [x] 创建docs目录结构
- [x] 创建projects目录结构
- [x] 创建tasks目录结构

### 配置开发工具
- [x] 配置Git仓库
- [x] 配置代码格式化工具（Prettier/ESLint）
- [x] 配置Go代码格式化工具（gofmt）
- [x] 配置编辑器（VS Code）

## 📁 项目结构

```
project-initialization-demo/
├── README.md                 # 本文件
├── .gitignore                # Git忽略文件配置
├── .prettierrc               # Prettier配置
├── .prettierignore           # Prettier忽略文件
├── .eslintrc.json            # ESLint配置
├── .eslintignore             # ESLint忽略文件
├── .golangci.yml             # Go代码检查配置
│
├── .vscode/                  # VS Code编辑器配置
│   ├── settings.json         # 编辑器设置
│   └── extensions.json       # 推荐扩展
│
├── docs/                     # 文档目录
│   ├── architecture/         # 架构文档
│   │   └── README.md
│   ├── api/                  # API文档
│   │   └── README.md
│   └── guides/               # 指南文档
│       ├── README.md
│       └── go-formatting.md  # Go格式化指南
│
├── projects/                 # 项目代��目录
│   ├── frontend/             # 前端项目
│   │   └── README.md
│   └── backend/              # 后端项目
│       └── README.md
│
└── tasks/                    # 任务管理目录
    └── README.md
```

## 🔧 配置文件说明

### 1. Git配置
**文件**: `.gitignore`
- 忽略node_modules、dist等构建产物
- 忽略.env等敏感文件
- 忽略编辑器配置（部分保留）

### 2. Prettier配置
**文件**: `.prettierrc`, `.prettierignore`
- 使用单引号
- 添加分号
- 每行最大80字符
- 使用2空格缩进

### 3. ESLint配置
**文件**: `.eslintrc.json`, `.eslintignore`
- 继承推荐配置
- 集成TypeScript支持
- 集成Prettier
- 自定义规则（允许console，警告any类型等）

### 4. Go格式化配置
**文件**: `.golangci.yml`, `docs/guides/go-formatting.md`
- 启用gofmt、goimports
- 启用常用linters
- 配置超时和测试检查

### 5. VS Code配置
**文件**: `.vscode/settings.json`, `.vscode/extensions.json`
- 保存时自动格式化
- ESLint自动修复
- TypeScript工作区SDK
- Go格式化工具配置
- 推荐扩展列表

## 📚 目录说明

### docs/ - 文档目录
- **architecture/**: 存放系统架构、技术选型等文档
- **api/**: 存放API接口文档
- **guides/**: 存放开发指南、部署指南等

### projects/ - 项目代码目录
- **frontend/**: 前端项目代码（React/Vue等）
- **backend/**: 后端项目代码（Node.js/Go等）

### tasks/ - 任务管理目录
- 存放项目任务列表、迭代计划等

## 🎯 使用说明

### 1. TypeScript/JavaScript项目

```bash
# 安装依赖
npm install prettier eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin eslint-config-prettier eslint-plugin-prettier --save-dev

# 格式化代码
npx prettier --write .

# 检查代码
npx eslint .

# 修复代码
npx eslint . --fix
```

### 2. Go项目

```bash
# 安装golangci-lint
brew install golangci-lint  # macOS
# 或
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 格式化代码
gofmt -w .

# 代码检查
golangci-lint run
```

### 3. VS Code设置

1. 安装推荐的扩展（打开项目时会提示）
2. 配置会自动生效
3. 保存时会自动格式化代码

## ✨ 最佳实践

### 代码提交前检查
```bash
# 1. 格式化代码
prettier --write .

# 2. 检查代码质量
eslint . --fix

# 3. TypeScript类型检查（如果有）
tsc --noEmit

# 4. 提交代码
git add .
git commit -m "feat: your commit message"
```

### Git提交信息规范
```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
test: 测试相关
chore: 构建或辅助工具变动
```

## 📖 相关文档

- [Go格式化指南](docs/guides/go-formatting.md)
- [Prettier文档](https://prettier.io/docs/en/)
- [ESLint文档](https://eslint.org/docs/user-guide/)
- [golangci-lint文档](https://golangci-lint.run/)

## 🎓 学习要点

1. **项目结构**: 理解docs、projects、tasks的组织方式
2. **Git配置**: 知道哪些文件应该被忽略
3. **代码格式化**: 理解Prettier和ESLint的区别和配合
4. **Go工具链**: 熟悉gofmt和golangci-lint的使用
5. **编辑器配置**: 利用VS Code提高开发效率

---

**创建日期**: 2026-01-28
**任务来源**: phase1-tasks.md - 1.3.2 项目初始化
