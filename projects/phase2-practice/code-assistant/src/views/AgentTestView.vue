<template>
  <div class="bg-white rounded-lg shadow-sm p-6">
    <h2 class="text-2xl font-bold text-gray-900 mb-4">
      🧪 Agent测试
    </h2>
    <p class="text-gray-600 mb-6">
      测试LangChain.js集成和基础Agent功能
    </p>

    <!-- 测试表单 -->
    <div class="space-y-4 mb-6">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          选择测试功能
        </label>
        <select
          v-model="selectedTest"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        >
          <option value="chat">基础对话</option>
          <option value="codeReview">代码审查</option>
          <option value="refactor">代码重构</option>
          <option value="techStack">技术栈选择</option>
        </select>
      </div>

      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          输入��容
        </label>
        <textarea
          v-model="userInput"
          rows="6"
          class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-transparent font-mono text-sm"
          :placeholder="getPlaceholder()"
        ></textarea>
      </div>

      <button
        @click="runTest"
        :disabled="loading || !userInput"
        class="w-full bg-primary-600 text-white py-2 px-4 rounded-lg hover:bg-primary-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
      >
        {{ loading ? '处理中...' : '运行测试' }}
      </button>
    </div>

    <!-- 结果显示 -->
    <div v-if="response" class="mt-6">
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Agent响应：</h3>
      <div class="bg-gray-50 rounded-lg p-4 whitespace-pre-wrap">
        {{ response.content }}
      </div>

      <div v-if="response.usage" class="mt-4 text-sm text-gray-600">
        <p>Token使用: {{ response.usage.totalTokens }} (输入: {{ response.usage.promptTokens }}, 输出: {{ response.usage.completionTokens }})</p>
      </div>
    </div>

    <!-- 错误显示 -->
    <div v-if="error" class="mt-6 bg-red-50 border border-red-200 rounded-lg p-4">
      <p class="text-red-800">❌ {{ error }}</p>
    </div>

    <!-- 状态显示 -->
    <div class="mt-6 pt-6 border-t">
      <h3 class="text-lg font-semibold text-gray-900 mb-2">集成状态：</h3>
      <div class="space-y-2">
        <div class="flex items-center">
          <span :class="apiKeyConfigured ? 'text-green-600' : 'text-red-600'">
            {{ apiKeyConfigured ? '✓' : '✗' }}
          </span>
          <span class="ml-2 text-gray-700">API密钥配置</span>
        </div>
        <div class="flex items-center">
          <span class="text-green-600">✓</span>
          <span class="ml-2 text-gray-700">LangChain.js已集成</span>
        </div>
        <div class="flex items-center">
          <span class="text-green-600">✓</span>
          <span class="ml-2 text-gray-700">Agent基类已创建</span>
        </div>
        <div class="flex items-center">
          <span class="text-green-600">✓</span>
          <span class="ml-2 text-gray-700">CodeAssistantAgent已实现</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { createCodeAssistant } from '@/agent';
import type { AgentResponse } from '@/agent/BaseAgent';

const selectedTest = ref<string>('chat');
const userInput = ref<string>('');
const response = ref<AgentResponse | null>(null);
const error = ref<string>('');
const loading = ref<boolean>(false);

const apiKeyConfigured = computed(() => {
  return !!import.meta.env.VITE_OPENAI_API_KEY;
});

function getPlaceholder(): string {
  switch (selectedTest.value) {
    case 'chat':
      return '输入你的问题，例如：什么是SOLID原则？';
    case 'codeReview':
      return '粘贴需要审查的代码...';
    case 'refactor':
      return '粘贴需要重构的代码...';
    case 'techStack':
      return '描述你的项目需求...';
    default:
      return '输入内容...';
  }
}

async function runTest() {
  if (!userInput.value) return;

  loading.value = true;
  error.value = '';
  response.value = null;

  try {
    // 创建Agent实例
    const agent = createCodeAssistant();

    // 根据选择的测试类型执行
    let result: AgentResponse;

    switch (selectedTest.value) {
      case 'chat':
        result = await agent.ask(userInput.value);
        break;
      case 'codeReview':
        result = await agent.reviewCode(userInput.value, 'typescript');
        break;
      case 'refactor':
        result = await agent.suggestRefactor(
          userInput.value,
          'typescript',
          '提高代码质量'
        );
        break;
      case 'techStack':
        result = await agent.suggestTechStack(userInput.value, [
          '用户认证',
          '数据存储',
          'API接口',
        ]);
        break;
      default:
        throw new Error('未知的测试类型');
    }

    response.value = result;
  } catch (err) {
    error.value =
      err instanceof Error ? err.message : '未知错误';
    console.error('Test error:', err);
  } finally {
    loading.value = false;
  }
}
</script>
