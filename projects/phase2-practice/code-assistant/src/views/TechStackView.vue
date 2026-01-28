<template>
  <div class="max-w-7xl mx-auto px-4 py-8">
    <div class="bg-white rounded-lg shadow-sm p-6">
      <h2 class="text-2xl font-bold text-gray-900 mb-4">📚 技术栈选择</h2>
      <p class="text-gray-600 mb-6">
        描述您的项目需求，AI将推荐最合适的技术栈和架构方案。
      </p>

      <!-- 输入区域 -->
      <div class="space-y-4 mb-6">
        <!-- 项目信息 -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              项目类型
            </label>
            <select
              v-model="projectType"
              class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="web">Web应用</option>
              <option value="mobile">移动应用</option>
              <option value="desktop">桌面应用</option>
              <option value="api">API服务</option>
              <option value="microservices">微服务</option>
              <option value="data-pipeline">数据管道</option>
              <option value="ml">机器学习</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              项目规模
            </label>
            <select
              v-model="scale"
              class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="small">小型（个人/小团队）</option>
              <option value="medium">中型（10-50人）</option>
              <option value="large">大型（50-200人）</option>
              <option value="enterprise">企业级（200+人）</option>
            </select>
          </div>
        </div>

        <!-- 项目描述 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            项目描述
          </label>
          <textarea
            v-model="projectDescription"
            placeholder="请详细描述您的项目..."
            rows="4"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          ></textarea>
        </div>

        <!-- 功能需求 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            功能需求（每行一个）
          </label>
          <textarea
            v-model="requirementsInput"
            placeholder="例如：&#10;用户认证&#10;数据可视化&#10;实时通信"
            rows="5"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          ></textarea>
        </div>

        <!-- 约束条件 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            约束条件（可选，每行一个）
          </label>
          <textarea
            v-model="constraintsInput"
            placeholder="例如：&#10;预算有限&#10;3个月内上线&#10;必须支持IE11"
            rows="3"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          ></textarea>
        </div>

        <!-- 团队技术栈 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            团队技术栈（可选，用逗号分隔）
          </label>
          <input
            v-model="teamSkillsInput"
            type="text"
            placeholder="例如：JavaScript, Python, React, Django"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <!-- 操作按钮 -->
        <div class="flex gap-4">
          <button
            @click="handleAnalyze"
            :disabled="!projectDescription || requirements.length === 0 || loading"
            class="flex-1 bg-blue-600 text-white py-3 px-6 rounded-md hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium transition-colors"
          >
            {{ loading ? '分析中...' : '生成技术方案' }}
          </button>
          <button
            @click="handleClear"
            :disabled="loading"
            class="px-6 py-3 border border-gray-300 rounded-md hover:bg-gray-50 disabled:cursor-not-allowed transition-colors"
          >
            清空
          </button>
        </div>
      </div>

      <!-- 错误提示 -->
      <div
        v-if="error"
        class="mb-6 p-4 bg-red-50 border border-red-200 rounded-md"
      >
        <p class="text-red-800 font-medium">错误</p>
        <p class="text-red-600 text-sm mt-1">{{ error }}</p>
      </div>

      <!-- 结果展示 -->
      <div v-if="result" class="mt-8 space-y-6">
        <div class="border-t pt-6">
          <h3 class="text-xl font-bold text-gray-900 mb-4">📊 技术方案建议</h3>

          <!-- Markdown渲染结果 -->
          <div
            class="prose prose-sm max-w-none bg-gray-50 p-6 rounded-lg"
            v-html="renderedResult"
          ></div>

          <!-- Token使用信息 -->
          <div
            v-if="result.usage"
            class="mt-4 p-3 bg-blue-50 border border-blue-200 rounded-md text-sm text-blue-800"
          >
            <span class="font-medium">Token使用：</span>
            输入 {{ result.usage.promptTokens }} | 输出
            {{ result.usage.completionTokens }} | 总计
            {{ result.usage.totalTokens }}
          </div>

          <!-- 快速预设按钮 -->
          <div class="mt-6">
            <h4 class="text-sm font-semibold text-gray-700 mb-3">快速预设</h4>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
              <button
                @click="applyPreset('ecommerce')"
                class="p-3 border border-gray-300 rounded-md hover:bg-gray-50 text-sm text-left"
              >
                <div class="font-medium">电商平台</div>
                <div class="text-xs text-gray-500 mt-1">Web应用</div>
              </button>
              <button
                @click="applyPreset('social')"
                class="p-3 border border-gray-300 rounded-md hover:bg-gray-50 text-sm text-left"
              >
                <div class="font-medium">社交网络</div>
                <div class="text-xs text-gray-500 mt-1">Web + Mobile</div>
              </button>
              <button
                @click="applyPreset('dashboard')"
                class="p-3 border border-gray-300 rounded-md hover:bg-gray-50 text-sm text-left"
              >
                <div class="font-medium">数据看板</div>
                <div class="text-xs text-gray-500 mt-1">Web应用</div>
              </button>
              <button
                @click="applyPreset('api')"
                class="p-3 border border-gray-300 rounded-md hover:bg-gray-50 text-sm text-left"
              >
                <div class="font-medium">REST API</div>
                <div class="text-xs text-gray-500 mt-1">后端服务</div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { techStackService } from '@/services/agentService';
import { marked } from 'marked';
import type { AgentResponse } from '@/agent/BaseAgent';

// 状态
const projectDescription = ref('');
const projectType = ref('web');
const scale = ref<'small' | 'medium' | 'large' | 'enterprise'>('medium');
const requirementsInput = ref('');
const constraintsInput = ref('');
const teamSkillsInput = ref('');
const loading = ref(false);
const error = ref('');
const result = ref<AgentResponse | null>(null);

// Agent实例已移除，使用服务层

// 计算数组
const requirements = computed(() => {
  if (!requirementsInput.value.trim()) return [];
  return requirementsInput.value
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean);
});

const constraints = computed(() => {
  if (!constraintsInput.value.trim()) return undefined;
  return constraintsInput.value
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean);
});

const teamSkills = computed(() => {
  if (!teamSkillsInput.value.trim()) return undefined;
  return teamSkillsInput.value.split(',').map((s) => s.trim()).filter(Boolean);
});

// 渲染Markdown结果
const renderedResult = computed(() => {
  if (!result.value) return '';
  return marked(result.value.content);
});

// 处理分析
async function handleAnalyze() {
  if (!projectDescription.value.trim()) {
    error.value = '请输入项目描述';
    return;
  }

  if (requirements.value.length === 0) {
    error.value = '请至少输入一个功能需求';
    return;
  }

  loading.value = true;
  error.value = '';
  result.value = null;

  try {
    // 使用服务层（包含缓存和重试）
    const response = await techStackService({
      projectDescription: projectDescription.value,
      requirements: requirements.value,
      constraints: constraints.value,
      teamSkills: teamSkills.value,
      projectType: projectType.value,
      scale: scale.value,
    });

    result.value = response;
  } catch (e: any) {
    error.value = e.message || '分析失败，请稍后重试';
  } finally {
    loading.value = false;
  }
}

// 清空表单
function handleClear() {
  projectDescription.value = '';
  requirementsInput.value = '';
  constraintsInput.value = '';
  teamSkillsInput.value = '';
  result.value = null;
  error.value = '';
}

// 应用预设
function applyPreset(preset: string) {
  switch (preset) {
    case 'ecommerce':
      projectType.value = 'web';
      projectDescription.value = '一个现代化的电商平台';
      requirementsInput.value =
        '用户注册登录\n商品浏览和搜索\n购物车\n订单管理\n支付集成\n商品评论\n后台管理';
      constraintsInput.value = '需要高可用性\n支持高并发\n移动端友好';
      teamSkillsInput.value = 'JavaScript, TypeScript, React, Node.js';
      scale.value = 'medium';
      break;

    case 'social':
      projectType.value = 'web';
      projectDescription.value = '社交网络平台';
      requirementsInput.value =
        '用户注册登录\n发布动态\n关注好友\n点赞评论\n私信功能\n消息通知\n内容推荐';
      constraintsInput.value = '需要实时更新\n支持大量用户';
      teamSkillsInput.value = 'JavaScript, React, Python';
      scale.value = 'large';
      break;

    case 'dashboard':
      projectType.value = 'web';
      projectDescription.value = '数据可视化看板';
      requirementsInput.value =
        '数据统计\n图表展示\n实时更新\n数据导出\n权限管理\n自定义报表';
      constraintsInput.value = '需要高性能\n复杂的数据可视化';
      teamSkillsInput.value = 'TypeScript, Vue, D3.js';
      scale.value = 'medium';
      break;

    case 'api':
      projectType.value = 'api';
      projectDescription.value = 'RESTful API服务';
      requirementsInput.value =
        'CRUD操作\n身份认证\n数据验证\nAPI文档\n速率限制\n日志记录';
      constraintsInput.value = '需要高性能\n良好的可扩展性';
      teamSkillsInput.value = 'Node.js, Express, MongoDB';
      scale.value = 'medium';
      break;
  }
}
</script>

<style scoped>
/* Markdown渲染样式 */
:deep(.prose) {
  color: #374151;
}

:deep(.prose h2) {
  font-size: 1.5em;
  font-weight: 700;
  margin-top: 1.5em;
  margin-bottom: 0.75em;
  color: #111827;
}

:deep(.prose h3) {
  font-size: 1.25em;
  font-weight: 600;
  margin-top: 1.25em;
  margin-bottom: 0.5em;
  color: #1f2937;
}

:deep(.prose ul) {
  list-style-type: disc;
  padding-left: 1.5em;
}

:deep(.prose ol) {
  list-style-type: decimal;
  padding-left: 1.5em;
}

:deep(.prose code) {
  background-color: #f3f4f6;
  padding: 0.125em 0.25em;
  border-radius: 0.25em;
  font-size: 0.875em;
}

:deep(.prose pre) {
  background-color: #1f2937;
  color: #f9fafb;
  padding: 1em;
  border-radius: 0.5em;
  overflow-x: auto;
}

:deep(.prose pre code) {
  background-color: transparent;
  padding: 0;
  color: inherit;
}
</style>
