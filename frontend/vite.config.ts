import { fileURLToPath, URL } from 'node:url'
import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig(({ command }) => {
  // 后端地址与安全入口：开发期把接口请求转发给 Go 服务，并按入口前缀重写路径
  const backend = process.env.OMOP_BACKEND ?? 'http://127.0.0.1:9999'
  const entrance = (process.env.OMOP_ENTRANCE ?? '').replace(/^\/+|\/+$/g, '')
  const proxyPrefix = entrance ? `/${entrance}` : ''
  const withPrefix = (path: string) => `${proxyPrefix}${path}`

  return {
    // 产物用相对基址 + Go 服务注入的 <base href>，从而支持任意安全入口而无需重新构建
    base: command === 'build' ? './' : '/',
    plugins: [vue(), tailwindcss()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      port: 5173,
      proxy: {
        '/api': { target: backend, changeOrigin: true, rewrite: withPrefix },
        '/swagger': { target: backend, changeOrigin: true, rewrite: withPrefix },
      },
    },
  }
})
