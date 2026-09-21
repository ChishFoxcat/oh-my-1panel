import { createRouter, createWebHistory } from 'vue-router'
import { useSystemStore } from '@/stores/system'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/login/index.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      name: 'overview',
      component: () => import('@/views/overview/index.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

router.beforeEach(async (to) => {
  const system = useSystemStore()
  if (!system.info) {
    try {
      await system.refresh()
    }
    catch {
      // 面板不可达时仍进入登录页，由登录页展示具体错误
      return to.meta.public ? true : { name: 'login' }
    }
  }
  if (!to.meta.public && !system.logged) {
    return { name: 'login' }
  }
  if (to.name === 'login' && system.logged) {
    return { name: 'overview' }
  }
  return true
})

export default router
