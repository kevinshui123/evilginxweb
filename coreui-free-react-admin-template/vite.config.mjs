import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import autoprefixer from 'autoprefixer'

export default defineConfig(() => {
  return {
    // ① base 设置为 './'，让打包产物中的静态资源走相对路径
    base: './',

    // ② 打包输出到 build/ 目录
    build: {
      outDir: 'build',
    },

    // ③ PostCSS 配置
    css: {
      postcss: {
        plugins: [
          autoprefixer({}),
        ],
      },
    },

    // ④ ESBuild Loader 配置
    esbuild: {
      loader: 'jsx',
      include: /src\/.*\.jsx?$/,
      exclude: [],
    },

    // ✅ ⑤ 添加 Ace 编辑器模块进依赖优化
    optimizeDeps: {
      include: [
        'ace-builds/src-noconflict/ace',
        'ace-builds/src-noconflict/mode-yaml',
        'ace-builds/src-noconflict/theme-github',
      ],
      force: true,
      esbuildOptions: {
        loader: {
          '.js': 'jsx',
        },
      },
    },

    // ⑥ 插件
    plugins: [react()],

    // ⑦ 路径别名
    resolve: {
      alias: [
        {
          find: 'src/',
          replacement: `${path.resolve(__dirname, 'src')}/`,
        },
      ],
      extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.scss'],
    },

    // ⑧ 开发服务器设置
    server: {
      host: '0.0.0.0', // 允许局域网访问
      port: 3001,
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          secure: false,
        },
      },
    },
  }
})
