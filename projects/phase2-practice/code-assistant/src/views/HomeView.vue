<template>
  <div class="space-y-8">
    <!-- Welcome Section -->
    <div class="bg-white rounded-lg shadow-sm p-6">
      <h2 class="text-3xl font-bold text-gray-900 mb-4">
        欢迎使用智能代码助手 🚀
      </h2>
      <p class="text-lg text-gray-600 mb-6">
        基于AI的代码审查、重构和技术栈选择工具
      </p>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- 功能卡片 -->
        <div
          v-for="feature in features"
          :key="feature.path"
          class="bg-gray-50 rounded-lg p-6 hover:shadow-md transition-shadow cursor-pointer"
          @click="$router.push(feature.path)"
        >
          <div class="text-4xl mb-4">{{ feature.icon }}</div>
          <h3 class="text-xl font-semibold text-gray-900 mb-2">
            {{ feature.title }}
          </h3>
          <p class="text-gray-600 mb-4">{{ feature.description }}</p>
          <button
            class="w-full bg-primary-600 text-white py-2 px-4 rounded-lg hover:bg-primary-700 transition-colors"
          >
            开始使用
          </button>
        </div>
      </div>
    </div>

    <!-- Status Section -->
    <div class="bg-white rounded-lg shadow-sm p-6">
      <h3 class="text-xl font-semibold text-gray-900 mb-4">项目状态</h3>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="text-center p-4 bg-green-50 rounded-lg">
          <div class="text-3xl font-bold text-green-600">✓</div>
          <div class="mt-2 text-sm text-gray-600">项目初始化</div>
        </div>
        <div class="text-center p-4 bg-green-50 rounded-lg">
          <div class="text-3xl font-bold text-green-600">✓</div>
          <div class="mt-2 text-sm text-gray-600">LangChain集成</div>
        </div>
        <div class="text-center p-4 bg-gray-50 rounded-lg">
          <div class="text-3xl font-bold text-gray-400">⏳</div>
          <div class="mt-2 text-sm text-gray-600">Prompt设计</div>
        </div>
        <div class="text-center p-4 bg-gray-50 rounded-lg">
          <div class="text-3xl font-bold text-gray-400">⏳</div>
          <div class="mt-2 text-sm text-gray-600">核心功能</div>
        </div>
      </div>
    </div>

    <!-- Tech Stack Section -->
    <div class="bg-white rounded-lg shadow-sm p-6">
      <h3 class="text-xl font-semibold text-gray-900 mb-4">技术栈</h3>
      <div class="flex flex-wrap gap-2">
        <span
          v-for="tech in techStack"
          :key="tech"
          class="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm"
        >
          {{ tech }}
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const features = ref([
  {
    icon: '🧪',
    title: 'Agent测试',
    description: '测试LangChain.js集成和Agent基础功能',
    path: '/agent-test',
  },
  {
    icon: '🔍',
    title: '代码审查',
    description: 'AI驱动的代码审查，发现潜在问题和改进建议',
    path: '/code-review',
  },
  {
    icon: '🔧',
    title: '代码重构',
    description: '智能重构建议，提高代码质量和可维护性',
    path: '/refactor',
  },
  {
    icon: '📚',
    title: '技术栈选择',
    description: '基于需求推荐最合适的技术栈和架构方案',
    path: '/tech-stack',
  },
]);

const techStack = ref([
  'Vue 3',
  'TypeScript',
  'Vite',
  'TailwindCSS',
  'LangChain.js',
  'OpenAI API',
  'Web Workers',
]);
</script>
