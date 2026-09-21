import type { RouteRecordRaw } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'
import { entranceMount } from '@/lib/app-base'
import { useSystemStore } from '@/stores/system'

// 资源与路由都在根路径；安全入口仅作为进入凭证（见 lib/app-base 的 entranceMount）。
// history 基址固定为空串：createWebHistory 不传参时会读取 <base href>，
// 且 vue-router 会把地址补成 base + '/'，两者都会污染入口地址。
const mount = entranceMount()
const home = () => import('@/views/home/index.vue')

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: home,
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

// 入口路径同样渲染首页，保证直接访问 /chish 时地址不被改写
if (mount && mount !== '/') {
  routes.splice(1, 0, {
    path: mount,
    name: 'entrance',
    component: home,
  })
}

const router = createRouter({
  history: createWebHistory('/'),
  routes,
})

// 进入页面前先取回面板信息与登录态，避免已登录时闪出登录表单；
// 已登录时把安全入口从地址栏去掉（登录后不再需要入口）。
// 面板不可达时不阻塞渲染，由登录表单展示具体错误。
router.beforeEach(async (to) => {
  const system = useSystemStore()
  if (!system.info) {
    try {
      await system.refresh()
    }
    catch {
      // 忽略：错误状态已写入 system.error
    }
  }
  if (system.logged && mount !== '' && to.path === mount) {
    return '/'
  }
  return true
})

export default router
