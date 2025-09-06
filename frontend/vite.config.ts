import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const port = 8091
  const apiTarget = env.VITE_API_TARGET || 'http://localhost:8080'
  return {
    plugins: [vue()],
    server: {
      port,
      open: false,
      proxy: {
        // Proxy API requests during dev if base url points elsewhere
        '/api': {
          target: apiTarget,
          changeOrigin: true,
          rewrite: (path) => path,
        },
      },
    },
  }
})
