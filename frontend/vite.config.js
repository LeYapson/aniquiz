import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      '/ws': {
        target: 'ws://192.168.27.74:8080',
        ws: true,
        changeOrigin: true
      },
      '/api': {
        target: 'http://192.168.27.74:8080',
        changeOrigin: true
      },
      '/animes': {
        target: 'http://192.168.27.74:8080',
        changeOrigin: true
      },
      '/rooms': {
        target: 'http://192.168.27.74:8080',
        changeOrigin: true
      },
      '/avatars': {
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