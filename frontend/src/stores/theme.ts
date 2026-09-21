import { defineStore } from 'pinia'
import { ref } from 'vue'

type Theme = 'light' | 'dark'

const STORAGE_KEY = 'omop-theme'

/** 主题：默认跟随面板设置，用户在界面上手选后以本地选择为准。 */
export const useThemeStore = defineStore('theme', () => {
  const theme = ref<Theme>('light')

  function apply(next: Theme) {
    theme.value = next
    document.documentElement.classList.toggle('dark', next === 'dark')
  }

  /** 首屏初始化：优先本地选择，其次系统偏好。 */
  function init() {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved === 'light' || saved === 'dark') {
      apply(saved)
      return
    }
    apply(window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
  }

  /** 与面板主题联动，仅在用户未手动选择时生效。 */
  function syncWithPanel(panelTheme: string) {
    if (localStorage.getItem(STORAGE_KEY)) {
      return
    }
    apply(panelTheme === 'dark' ? 'dark' : 'light')
  }

  function toggle() {
    const next: Theme = theme.value === 'dark' ? 'light' : 'dark'
    localStorage.setItem(STORAGE_KEY, next)
    apply(next)
  }

  return { theme, init, syncWithPanel, toggle }
})

export type { Theme }
