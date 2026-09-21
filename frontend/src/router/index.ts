import { createRouter, createWebHistory } from 'vue-router'
import { appBase } from '@/lib/app-base'
import { useSystemStore } from '@/stores/system'

const router = createRouter({
  history: createWebHistory(appBase()),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/home/index.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/',
    },
  ],
})

// 进入首页前先取回面板信息与登录态，避免已登录时闪出登录表单。
// 面板不可达时不阻塞渲染，由登录表单展示具体错误。
router.beforeEach(async () => {
  const system = useSystemStore()
  if (!system.info) {
    try {
      await system.refresh()
    }
    catch {
      // 忽略：错误状态已写入 system.error
    }
  }
  return true
})

export default router
