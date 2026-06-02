import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  base: '/',
  build: {
    outDir: '../internal/web/assets/app',
    emptyOutDir: true,
    assetsDir: 'assets',
  },
})
