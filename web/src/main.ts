/* eslint-disable simple-import-sort/imports */
import { createApp, defineComponent, h } from 'vue';
import TDesign from 'tdesign-vue-next';

import { iconMap } from './utils/icon-map';

import App from './App.vue';
import router from './router';
import { store, useUserStore } from './store';
import i18n from './locales';
import { initGlobalWebSocket } from './utils/websocket';
import { useNotificationStore } from './store';

import 'tdesign-vue-next/es/style/index.css';
import '@/style/index.less';
import './permission';

const app = createApp(App);

app.directive('permission', {
  mounted(el, binding) {
    const userStore = useUserStore();
    const permValue = binding.value as string;
    if (permValue && userStore.perms && !userStore.perms.includes(permValue)) {
      el.parentNode?.removeChild(el);
    }
  },
});

app.use(TDesign);
// 自定义 TIcon 组件：使用本地 SVG 图标组件替代 TDesign 默认实现
// 先移除已有注册（TDesign 已在 use 中注册 TIcon），避免 "already registered" 警告
const ctx = (app as any)._context;
if (ctx?.components) {
  delete ctx.components['TIcon'];
}
app.component('TIcon', defineComponent({
  name: 'TIcon',
  props: {
    name: { type: String, default: '' },
    size: { type: [String, Number], default: undefined },
    onClick: { type: Function, default: undefined },
  },
  setup(props, { attrs }) {
    return () => {
      const IconComponent = iconMap[props.name];
      if (IconComponent) {
        return h(IconComponent, { ...attrs, size: props.size, onClick: props.onClick });
      }
      // 图标不存在时渲染空占位
      const size = props.size;
      const sizeStyle = size
        ? { width: typeof size === 'number' ? `${size}px` : size, height: typeof size === 'number' ? `${size}px` : size }
        : {};
      return h('i', { class: `t-icon t-icon-${props.name}`, style: sizeStyle });
    };
  },
}));
app.use(store);
app.use(router);
app.use(i18n);

app.mount('#app');

initGlobalWebSocket();

const notificationStore = useNotificationStore();
notificationStore.initWebSocketListener();
