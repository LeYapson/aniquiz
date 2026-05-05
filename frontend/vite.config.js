import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      // Redirige les appels WebSocket
      '/ws': {
        target: 'ws://127.0.0.1:8080',
        ws: true,
        changeOrigin: true
      },
      // Redirige les appels API classiques (si tu en as)
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  test: {
    // Cette ligne est la clé :
    environment: 'jsdom', 
    globals: true // Permet d'utiliser 'describe', 'it', 'expect' sans les importer
  }
})