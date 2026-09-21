import { createRouter, createWebHistory } from 'vue-router'
import { routeMount } from '@/lib/app-base'
import { useSystemStore } from '@/stores/system'

// 安全入口由路由路径承载（见 lib/app-base 的 routeMount），history 基址必须是空串：
// vue-router 在基址为"位于 base 根"时会把地址补成 base + '/'，而 createWebHistory()
// 不传参时还会自动读取注入的 <base href>，把 /chish 又变回 /chish/。
// 显式传 '/' 可避免读取 <base>，规范化后得到空基址，地址栏保持原样。
const home = routeMount() || '/'

const router = createRouter({
  history: createWebHistory('/'),
  routes: [
    {
      path: home,
      name: 'home',
      component: () => import('@/views/home/index.vue'),
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: home,
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
