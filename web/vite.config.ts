import path from 'node:path';

import vue from '@vitejs/plugin-vue';
import vueJsx from '@vitejs/plugin-vue-jsx';
import type { ConfigEnv, UserConfig } from 'vite';
import { loadEnv } from 'vite';

import svgLoader from 'vite-svg-loader';

const CWD = process.cwd();

// 直接从 process.env 读取，避免 loadEnv 可能漏掉 shell 环境变量
const APP_VERSION = process.env.VITE_APP_VERSION || 'dev';

// https://vitejs.dev/config/
export default ({ mode }: ConfigEnv): UserConfig => {
  const { VITE_BASE_URL, VITE_API_URL_PREFIX } = loadEnv(mode, CWD);
  return {
    base: VITE_BASE_URL,
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },

    css: {
      preprocessorOptions: {
        less: {
          modifyVars: {
            hack: `true; @import (reference) "${path.resolve('src/style/variables.less')}";`,
          },
          math: 'strict',
          javascriptEnabled: true,
        },
      },
    },

    plugins: [
      vue(),
      vueJsx(),
      svgLoader(),
    ],

    define: {
      __APP_VERSION__: JSON.stringify(APP_VERSION),
    },

    server: {
      port: 9800,
      host: '0.0.0.0',
      proxy: {
        [VITE_API_URL_PREFIX]: {
          target: 'http://127.0.0.1:9898',
          changeOrigin: true,
          cookieDomainRewrite: 'localhost',
          ws: true,
        },
        '/uploads': {
          target: 'http://127.0.0.1:9898',
          changeOrigin: true,
        },
        '/system-assets': {
          target: 'http://127.0.0.1:9898',
          changeOrigin: true,
        },
        '/health': {
          target: 'http://127.0.0.1:9898',
          changeOrigin: true,
        },
      },
    },

    build: {
      minify: 'terser',
      terserOptions: {
        compress: {
          drop_console: true,
          drop_debugger: true,
        },
      },
    },
  };
};
