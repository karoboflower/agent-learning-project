<template>
  <div class="max-w-7xl mx-auto px-4 py-8">
    <div class="bg-white rounded-lg shadow-sm p-6">
      <h2 class="text-2xl font-bold text-gray-900 mb-4">🔧 代码重构</h2>
      <p class="text-gray-600 mb-6">
        获取智能的代码重构建议，提升代码质量和可维护性。
      </p>

      <!-- 输入区域 -->
      <div class="space-y-4 mb-6">
        <!-- 语言选择和重构类型 -->
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
              重构类型
            </label>
            <select
              v-model="refactorType"
              class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="general">通用重构</option>
              <option value="extract-method">提取方法</option>
              <option value="rename">重命名</option>
              <option value="simplify-conditional">简化条件</option>
              <option value="remove-duplication">移除重复</option>
              <option value="design-pattern">应用设计模式</option>
              <option value="performance">性能优化</option>
            </select>
          </div>
        </div>

        <!-- 代码输入 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            待重构代码
          </label>
          <textarea
            v-model="code"
            placeholder="请粘贴需要重构的代码..."
            rows="15"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
          ></textarea>
        </div>

        <!-- 重构目标 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            重构目标
          </label>
          <input
            v-model="goal"
            type="text"
            placeholder="例如：提高代码可读性和可维护性"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <!-- 约束条件 -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">
            约束条件（可选，用逗号分隔）
          </label>
          <input
            v-model="constraintsInput"
            type="text"
            placeholder="例如：不能改变API接口, 保持向后兼容"
            class="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <!-- 保持行为不变 -->
        <div class="flex items-center">
          <input
            v-model="preserveBehavior"
            type="checkbox"
            id="preserveBehavior"
            class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          <label for="preserveBehavior" class="ml-2 text-sm text-gray-700">
            必须保持代码原有行为和功能不变
          </label>
        </div>

        <!-- 操作按钮 -->
        <div class="flex gap-4">
          <button
            @click="handleRefactor"
            :disabled="!code || !goal || loading"
            class="flex-1 bg-blue-600 text-white py-3 px-6 rounded-md hover:bg-blue-700 disabled:bg-gray-300 disabled:cursor-not-allowed font-medium transition-colors"
          >
            {{ loading ? '重构中...' : '开始重构' }}
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
          <h3 class="text-xl font-bold text-gray-900 mb-4">📊 重构建议</h3>

          <!-- 代码对比视图切换 -->
          <div class="mb-4 flex gap-2">
            <button
              @click="viewMode = 'markdown'"
              :class="[
                'px-4 py-2 rounded-md font-medium transition-colors',
                viewMode === 'markdown'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-200 text-gray-700 hover:bg-gray-300',
              ]"
            >
              详细说明
            </button>
            <button
              @click="viewMode = 'comparison'"
              :class="[
                'px-4 py-2 rounded-md font-medium transition-colors',
                viewMode === 'comparison'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-200 text-gray-700 hover:bg-gray-300',
              ]"
            >
              代码对比
            </button>
          </div>

          <!-- Markdown视图 -->
          <div
            v-if="viewMode === 'markdown'"
            class="prose prose-sm max-w-none bg-gray-50 p-6 rounded-lg"
            v-html="renderedResult"
          ></div>

          <!-- 代码对比视图 -->
          <div v-if="viewMode === 'comparison'" class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <h4 class="text-sm font-semibold text-gray-700 mb-2">原始代码</h4>
              <pre
                class="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm"
              ><code>{{ code }}</code></pre>
            </div>
            <div>
              <h4 class="text-sm font-semibold text-gray-700 mb-2">重构后代码</h4>
              <pre
                class="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm"
              ><code>{{ extractRefactoredCode(result.content) }}</code></pre>
            </div>
          </div>

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
import { refactorCodeService } from '@/services/agentService';
import { marked } from 'marked';
import type { AgentResponse } from '@/agent/BaseAgent';

// 状态
const code = ref('');
const language = ref('javascript');
const refactorType = ref('general');
const goal = ref('');
const constraintsInput = ref('');
const preserveBehavior = ref(true);
const loading = ref(false);
const error = ref('');
const result = ref<AgentResponse | null>(null);
const viewMode = ref<'markdown' | 'comparison'>('markdown');

// Agent实例已移除，使用服务层

// 计算约束条件数组
const constraints = computed(() => {
  if (!constraintsInput.value.trim()) return undefined;
  return constraintsInput.value.split(',').map((s) => s.trim()).filter(Boolean);
});

// 渲染Markdown结果
const renderedResult = computed(() => {
  if (!result.value) return '';
  return marked(result.value.content);
});

// 从Markdown中提取重构后的代码
function extractRefactoredCode(markdown: string): string {
  // 尝试提取代码块
  const codeBlockRegex = /```[\w]*\n([\s\S]*?)```/g;
  const matches = Array.from(markdown.matchAll(codeBlockRegex));

  if (matches.length > 0) {
    // 返回第一个代码块（通常是重构后的代码）
    return matches[0][1].trim();
  }

  return '无法提取重构后的代码';
}

// 处理重构
async function handleRefactor() {
  if (!code.value.trim()) {
    error.value = '请输入代码';
    return;
  }

  if (!goal.value.trim()) {
    error.value = '请输入重构目标';
    return;
  }

  loading.value = true;
  error.value = '';
  result.value = null;

  try {
    // 使用服务层（包含缓存和重试）
    const response = await refactorCodeService({
      code: code.value,
      language: language.value,
      goal: goal.value,
      constraints: constraints.value,
    });

    result.value = response;
  } catch (e: any) {
    error.value = e.message || '重构失败，请稍后重试';
  } finally {
    loading.value = false;
  }
}

// 清空表单
function handleClear() {
  code.value = '';
  goal.value = '';
  constraintsInput.value = '';
  preserveBehavior.value = true;
  result.value = null;
  error.value = '';
  viewMode.value = 'markdown';
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
