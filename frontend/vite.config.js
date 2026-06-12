import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/demo': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/radar': {
        target: 'http://localhost:3001',
        changeOrigin: true,
      },
      '/sensor': {
        target: 'http://localhost:3002',
        changeOrigin: true,
      },
      '/nav': {
        target: 'http://localhost:3003',
        changeOrigin: true,
      },
    },
  },
})
