import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@pos/ui': fileURLToPath(new URL('../../packages/ui/src', import.meta.url)),
      '@pos/module-sdk': fileURLToPath(new URL('../../packages/module-sdk/src', import.meta.url)),
      '@pos/offline': fileURLToPath(new URL('../../packages/offline/src', import.meta.url)),
      'vue': fileURLToPath(new URL('./node_modules/vue', import.meta.url)),
      'vue-router': fileURLToPath(new URL('./node_modules/vue-router', import.meta.url)),
      'pinia': fileURLToPath(new URL('./node_modules/pinia', import.meta.url)),
      'vue-i18n': fileURLToPath(new URL('./node_modules/vue-i18n', import.meta.url)),
    },
  },
  test: { environment: 'jsdom' },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080',
      '/health': 'http://localhost:8080',
    },
  },
  build: { outDir: 'dist' },
})
