import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// 开发期把接口与文档请求转发到本地 Go 服务，保证浏览器视角下前后端同源。
const backend = process.env.OMP_BACKEND ?? 'http://127.0.0.1:9999'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': { target: backend, changeOrigin: true },
      '/swagger': { target: backend, changeOrigin: true },
    },
  },
})
