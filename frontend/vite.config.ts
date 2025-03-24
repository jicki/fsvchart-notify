import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],

  // 将打包后文件输出到 dist
  build: {
    outDir: 'dist',
    emptyOutDir: true
  },

  // 开发服务器配置
  server: {
    host: '0.0.0.0',  // 设置为 0.0.0.0 以便接受所有外部请求
    port: 5173, // dev时本地端口，随意
    proxy: {
      '/api': 'http://127.0.0.1:8080'  // 代理后端 API
    }
  }
})
