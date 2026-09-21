import type { LoginResponse, SystemInfo, UserInfo } from '@/api/types'
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { fetchSystemInfo } from '@/api/auth'
import { useThemeStore } from './theme'

/** 面板信息与登录态，供路由守卫与登录页共用。 */
export const useSystemStore = defineStore('system', () => {
  const info = ref<SystemInfo | null>(null)
  const error = ref('')
  const loading = ref(false)
  const expiresAt = ref('')

  const logged = computed(() => info.value?.logged === true)
  const user = computed<UserInfo | undefined>(() => info.value?.user)
  const needCaptcha = computed(() => info.value?.needCaptcha === true)
  const panelName = computed(() => info.value?.panelName || '1Panel')

  /** 拉取面板信息并同步主题。 */
  async function refresh() {
    loading.value = true
    try {
      const result = await fetchSystemInfo()
      info.value = result
      error.value = ''
      useThemeStore().syncWithPanel(result.theme)
      return result
    }
    catch (thrown) {
      error.value = thrown instanceof Error ? thrown.message : '无法连接面板'
      info.value = null
      throw thrown
    }
    finally {
      loading.value = false
    }
  }

  /** 首次进入时确保已加载，已加载则直接复用。 */
  async function ensure() {
    if (info.value) {
      return info.value
    }
    return refresh()
  }

  /** 登录成功后更新本地登录态与会话到期时间。 */
  async function markLoggedIn(result: LoginResponse) {
    expiresAt.value = result.expiresAt ?? ''
    await refresh()
  }

  /** 退出登录后清除本地状态。 */
  function clear() {
    info.value = null
    expiresAt.value = ''
  }

  return { info, error, loading, expiresAt, logged, user, needCaptcha, panelName, refresh, ensure, markLoggedIn, clear }
})
