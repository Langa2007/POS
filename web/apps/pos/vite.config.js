import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
export default defineConfig({ plugins: [vue()], test: { environment: 'jsdom' }, server: { port: 5173, proxy: { '/api': 'http://localhost:8080', '/health': 'http://localhost:8080' } }, build: { outDir: 'dist' } })
