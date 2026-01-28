<template>
  <div class="min-h-screen bg-gray-50">
    <header class="bg-white shadow-sm">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <h1 class="text-2xl font-bold text-gray-900">
          �� 智能代码助手
        </h1>
        <p class="mt-1 text-sm text-gray-500">
          AI驱动的代码审查、重构和技术栈选择助手
        </p>
      </div>
    </header>

    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <router-view />
    </main>

    <footer class="bg-white border-t mt-auto">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <p class="text-center text-sm text-gray-500">
          Task 2.1: 前端Agent应用 - 智能代码助手
        </p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
// App root component
</script>
