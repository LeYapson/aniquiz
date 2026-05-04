import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      // Redirige les appels WebSocket
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      },
      // Redirige les appels API classiques (si tu en as)
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  }
})