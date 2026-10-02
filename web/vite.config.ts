import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: {
    // lands inside cmd/server/ so the Go binary can embed it
    outDir: '../cmd/server/web/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: { '/v1': 'http://localhost:8419' },
  },
})
