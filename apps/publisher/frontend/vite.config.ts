import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
export default defineConfig({
  plugins: [vue()], base: './',
  build: { rollupOptions: { input: { app: 'index.html', 'design-system': 'design-system.html' } } },
});
