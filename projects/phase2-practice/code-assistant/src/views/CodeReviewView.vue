<template>
  <div class="max-w-7xl mx-auto px-4 py-8">
    <div class="bg-white rounded-lg shadow-sm p-6">
      <h2 class="text-2xl font-bold text-gray-900 mb-4">🔍 代码审查</h2>
      <p class="text-gray-600 mb-6">
        上传或粘贴您的代码，AI将进行全面的代码审查并提供改进建议。
      </p>

      <!-- 输入区域 -->
      <div class="space-y-4 mb-6">
        <!-- 语言选择和选项 -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              编程语言
            </label>
            <select
              v-model="language"
              class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="javascript">JavaScript</option>
              <option value="typescript">TypeScript</option>
              <option value="python">Python</option>
              <option value="java">Java</option>
              <option value="go">Go</option>
              <option value="rust">Rust</option>
              <option value="cpp">C++</option>
            </select>
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">
              审查类型
            </label>
            <select
              v-model="reviewType"
              class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="full">全面审查</option>
              <option value="quick">快速审查</option>
              <option value="security">安全审查</option>
              <option value="performance">性能审查</option>
              <option value="accessibility">可访问性审查</option>
              <option value="testing">测试审查</option>
            </select>
          </div>
        </div>

        <!-- 代码输入 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            代码
          </label>
          <textarea
            v-model="code"
            placeholder="请粘贴您的代码..."
            rows="15"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
          ></textarea>
        </div>

        <!-- 代码背景（可选） -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            代码背景（可选）
          </label>
          <input
            v-model="context"
            type="text"
            placeholder="例如：这是一个API处理函数"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <!-- 关注点（可选） -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            重点关注（可选，用逗号分隔）
          </label>
          <input
            v-model="focusAreasInput"
            type="text"
            placeholder="例如：错误处理, 性能优化, 代码规范"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <!-- 操作按钮 -->
        <div class="flex gap-4">
          <button
            @click="handleReview"
            :disabled="!code || loading"
            class="flex-1 bg-blue-600 text-white py-3 px-6 rounded-md hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium transition-colors"
          >
            {{ loading ? '审查中...' : '开始审查' }}
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
          <h3 class="text-xl font-bold text-gray-900 mb-4">📊 审查结果</h3>

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
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { reviewCodeService } from '@/services/agentService';
//import { marked } from 'marked';
import type { AgentResponse } from '@/agent/BaseAgent';

// 状态
const code = ref('');
const language = ref('javascript');
const reviewType = ref('full');
const context = ref('');
const focusAreasInput = ref('');
const loading = ref(false);
const error = ref('');
const result = ref<AgentResponse | null>(null);

// 计算关注点数组
const focusAreas = computed(() => {
  if (!focusAreasInput.value.trim()) return undefined;
  return focusAreasInput.value.split(',').map((s) => s.trim()).filter(Boolean);
});

// 渲染Markdown结果
const renderedResult = computed(() => {
  if (!result.value) return '';
  return result.value.content;
});

// 处理审查
async function handleReview() {
  if (!code.value.trim()) {
    error.value = '请输入代码';
    return;
  }

  loading.value = true;
  error.value = '';
  result.value = null;

  try {
    // 根据审查类型设置上下文和关注点
    let reviewContext = context.value;
    let reviewFocusAreas = focusAreas.value;

    switch (reviewType.value) {
      case 'quick':
        reviewContext = '快速审查';
        reviewFocusAreas = ['明显问题'];
        break;
      case 'security':
      case 'performance':
      case 'accessibility':
      case 'testing':
        reviewContext = `${reviewType.value}专项审查`;
        reviewFocusAreas = [reviewType.value];
        break;
    }

    // 使用服务层（包含缓存和重试）
    const response = await reviewCodeService({
      code: code.value,
      language: language.value,
      context: reviewContext || undefined,
      focusAreas: reviewFocusAreas,
    });

    result.value = response;
  } catch (e: any) {
    error.value = e.message || '审查失败，请稍后重试';
  } finally {
    loading.value = false;
  }
}

// 清空表单
function handleClear() {
  code.value = '';
  context.value = '';
  focusAreasInput.value = '';
  result.value = null;
  error.value = '';
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
