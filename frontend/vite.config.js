import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      // Redirige les appels WebSocket
      '/ws': {
        target: 'ws://192.168.27.74:8080',
        ws: true,
        changeOrigin: true
      },
      // Redirige les appels API classiques (si tu en as)
      '/api': {
        target: 'http://192.168.27.74:8080',
        changeOrigin: true
      }
    }
  },
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['src/**/*.{test,spec}.{js,ts}'],
  }
})