# Xt-PMS - 前端管理界面

基于 **Vue 3 + TDesign + Vite + TypeScript** 构建的生产管理系统前端。

## 技术栈

- **框架**: Vue 3.5.13
- **构建工具**: Vite 6.2.2
- **UI 组件库**: TDesign (腾讯) 1.13.2
- **状态管理**: Pinia 2.1.7
- **路由**: Vue Router 4.3
- **HTTP 客户端**: Axios 1.8.4
- **图表**: ECharts 5.4.3
- **国际化**: Vue i18n 9.9.1

## 开发

```bash
# 安装依赖
npm install

# 启动开发服务器（默认 http://localhost:3002）
npm run dev

# 使用 Mock 数据启动
npm run dev:mock
```

## 构建

```bash
# 生产构建
npm run build

# TypeScript 类型检查
npm run build:type

# 预览构建结果
npm run preview
```

## 代码规范

```bash
# ESLint 检查
npm run lint

# Stylelint 样式检查
npm run stylelint

# 自动修复
npm run lint:fix
npm run stylelint:fix
```

## 项目结构

```
web/
├── src/
│   ├── api/         # API 接口封装
│   ├── assets/      # 静态资源
│   ├── components/  # 公共组件
│   ├── config/      # 前端配置
│   ├── constants/   # 常量定义
│   ├── hooks/       # 组合式函数
│   ├── layouts/     # 布局组件
│   ├── locales/     # 国际化
│   ├── pages/       # 页面组件
│   ├── router/      # 路由配置
│   ├── store/       # Pinia 状态管理
│   ├── style/       # 样式文件
│   ├── types/       # TypeScript 类型定义
│   └── utils/       # 工具函数
├── public/          # 公共静态资源
└── mock/            # Mock 数据
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `VITE_IS_REQUEST_PROXY` | 是否使用 Vite 代理 | `true` |
| `VITE_API_URL` | API 服务器地址 | `http://127.0.0.1:8080` |
| `VITE_API_URL_PREFIX` | API 路径前缀 | `/api` |

更多信息请参考：[CONTRIBUTING.md](../CONTRIBUTING.md)
