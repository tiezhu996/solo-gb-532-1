import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: { port: 5173, proxy: { '/api': { target: 'http://127.0.0.1:20532', changeOrigin: true, rewrite: path => path === '/api/healthz' ? '/healthz' : path } } },
  build: { sourcemap: false, chunkSizeWarningLimit: 850 }
})
