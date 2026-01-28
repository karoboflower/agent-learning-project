import { createRouter, createWebHistory } from 'vue-router';
import HomeView from '@/views/HomeView.vue';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/agent-test',
      name: 'agent-test',
      component: () => import('@/views/AgentTestView.vue'),
    },
    {
      path: '/code-review',
      name: 'code-review',
      component: () => import('@/views/CodeReviewView.vue'),
    },
    {
      path: '/refactor',
      name: 'refactor',
      component: () => import('@/views/RefactorView.vue'),
    },
    {
      path: '/tech-stack',
      name: 'tech-stack',
      component: () => import('@/views/TechStackView.vue'),
    },
  ],
});

export default router;
