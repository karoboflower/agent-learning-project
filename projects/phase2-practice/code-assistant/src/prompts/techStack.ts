/**
 * 技术栈选择Prompt模板
 */

export interface TechStackInput {
  projectDescription: string;
  projectType?: string;
  requirements: string[];
  constraints?: string[];
  teamSkills?: string[];
  scale?: 'small' | 'medium' | 'large' | 'enterprise';
}

/**
 * 技术栈选择系统Prompt
 */
export const TECH_STACK_SYSTEM_PROMPT = `你是一位经验丰富的技术架构师，在技术选型和系统架构设计方面有深厚的经验。

你的选型原则：
1. **适合场景**：技术选择必须符合项目需求和规模
2. **团队能力**：考虑团队的技术栈和学习曲线
3. **生态成熟度**：优先选择生态完善、社区活跃的技术
4. **长期维护**：考虑技术的可维护性和未来发展
5. **成本效益**：平衡开发成本、运维成本和时间成本

技术选型考虑因素：
- 项目需求和业务场景
- 性能和可扩展性要求
- 团队技术储备
- 开发效率
- 社区支持和文档
- 学习曲线
- 部署和运维难度
- 许可证和成本

回复格式要求：
1. 系统化地分析项目需求
2. 提供完整的技术栈方案
3. 说明每个技术选择的理由
4. 提供替代���案
5. 给出架构建议`;

/**
 * 构建技术栈选择Prompt
 */
export function buildTechStackPrompt(input: TechStackInput): string {
  const {
    projectDescription,
    projectType,
    requirements,
    constraints,
    teamSkills,
    scale = 'medium',
  } = input;

  let prompt = `请为以下项目推荐完整的技术栈方案：\n\n`;

  prompt += `## 项目信息\n\n`;
  prompt += `**项目描述：**\n${projectDescription}\n\n`;

  if (projectType) {
    prompt += `**项目类型：** ${projectType}\n\n`;
  }

  prompt += `**项目规模：** ${scale}\n\n`;

  prompt += `**功能需求：**\n${requirements.map((r, i) => `${i + 1}. ${r}`).join('\n')}\n\n`;

  if (constraints && constraints.length > 0) {
    prompt += `**约束条件：**\n${constraints.map((c, i) => `${i + 1}. ${c}`).join('\n')}\n\n`;
  }

  if (teamSkills && teamSkills.length > 0) {
    prompt += `**团队技术栈：**\n${teamSkills.join(', ')}\n\n`;
  }

  prompt += `请提供详细的技术选型方案：

## 1. 需求分析
- 核心功能分析
- 技术挑战识别
- 性能要求评估

## 2. 推荐技术栈

### 前端技术栈
- 框架/库：
- 状态管理：
- UI组件库：
- 构建工具：
- 其他工具：

### 后端技术栈
- 编程语言：
- Web框架：
- 数据库：
- 缓存：
- 消息队列：

### 基础设施
- 服务器/云服务：
- 容器化：
- CI/CD：
- 监控日志：

### 开发工具
- 版本控制：
- 项目管理：
- 测试工具：

## 3. 技术选择理由
对每个主要技术选择，说明：
- 为什么选择这个技术
- 它如何满足项目需求
- 与其他技术的对比优势

## 4. 替代方案
对关键技术提供1-2个替代选项，并说明差异

## 5. 架构建议
- 整体架构模式（如微服务、单体、Serverless等）
- 关键的架构决策
- 可扩展性考虑

## 6. 风险和注意事项
- 潜在的技术风险
- 学习曲线评估
- 避坑指南

## 7. 实施路线图
建议的技术实施顺序和里程碑`;

  return prompt;
}

/**
 * 前端技术栈选择Prompt
 */
export function buildFrontendTechStackPrompt(
  projectDescription: string,
  requirements: string[]
): string {
  return `请为以下前端项目推荐技术栈：

**项目描述：**
${projectDescription}

**功能需求：**
${requirements.map((r, i) => `${i + 1}. ${r}`).join('\n')}

请重点推荐：
1. **前端框架**（React/Vue/Angular/Svelte等）
2. **状态管理方案**
3. **路由方案**
4. **UI组件库**
5. **样式方案**（CSS-in-JS/Tailwind/SCSS等）
6. **构建工具**（Vite/Webpack等）
7. **类型系统**（TypeScript等）
8. **测试框架**

对每个选择，请说明：
- 推荐理由
- 适用场景
- 学习成本
- 替代方案`;
}

/**
 * 后端技术栈选择Prompt
 */
export function buildBackendTechStackPrompt(
  projectDescription: string,
  requirements: string[]
): string {
  return `请为以下后端项目推荐技术栈：

**项目描述：**
${projectDescription}

**功能需求：**
${requirements.map((r, i) => `${i + 1}. ${r}`).join('\n')}

请重点推荐：
1. **编程语言**（Node.js/Python/Java/Go等）
2. **Web框架**
3. **数据库**（SQL/NoSQL）
4. **ORM/ODM**
5. **认证方案**
6. **API设计**（REST/GraphQL/gRPC）
7. **缓存策略**
8. **消息队列**（如需要）
9. **文件存储**
10. **日志和监控**

对每个选择，请说明：
- 推荐理由
- 性能特点
- 可扩展性
- 运维成本`;
}

/**
 * 数据库选择Prompt
 */
export function buildDatabaseSelectionPrompt(
  dataRequirements: string[],
  scale: string,
  queryPatterns?: string[]
): string {
  return `请推荐合适的数据库方案：

**数据需求：**
${dataRequirements.map((r, i) => `${i + 1}. ${r}`).join('\n')}

**数据规模：** ${scale}

${queryPatterns ? `**查询模式：**\n${queryPatterns.map((p, i) => `${i + 1}. ${p}`).join('\n')}\n` : ''}

请分析并推荐：

## 1. 主数据库推荐
- 数据库类型（SQL/NoSQL/NewSQL）
- 具体产品推荐
- 选择理由

## 2. 缓存方案
- 缓存类型
- 缓存策略
- 具体产品

## 3. 数据架构
- 分库分表策略（如需要）
- 读写分离方案
- 数据备份策略

## 4. 替代方案
其他可考虑的数据库选项

## 5. 技术对比
主要候选数据库的对比分析`;
}

/**
 * 微服务架构技术栈Prompt
 */
export function buildMicroservicesTechStackPrompt(
  projectDescription: string,
  services: string[]
): string {
  return `请为以下微服务项目推荐技术栈：

**项目描述：**
${projectDescription}

**规划的服务：**
${services.map((s, i) => `${i + 1}. ${s}`).join('\n')}

请推荐：

## 1. 服务开发
- 编程语言和框架
- API网关
- 服务注册与发现

## 2. 服务间通信
- 同步通信（REST/gRPC）
- 异步通信（消息队列）
- 服务网格（Service Mesh）

## 3. 数据管理
- 每个服务的数据存储
- 分布式事务方案
- 数据一致性策略

## 4. 基础设施
- 容器化（Docker/Kubernetes）
- 配置中心
- 日志聚合
- 分布式追踪
- 监控告警

## 5. DevOps
- CI/CD流程
- 自动化测试
- 部署策略

## 6. 架构建议
- 服务拆分原则
- 接口设计规范
- 版本管理策略`;
}
