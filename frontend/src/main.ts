import { createApp } from 'vue'
import { createPinia } from 'pinia'

import '@fontsource-variable/geist'
import './assets/index.css'

import App from './App.vue'
import router from './router'
import { useThemeStore } from './stores/theme'

const app = createApp(App)

app.use(createPinia())
app.use(router)

// 主题在挂载前初始化，避免首屏闪白
useThemeStore().init()

app.mount('#app')
