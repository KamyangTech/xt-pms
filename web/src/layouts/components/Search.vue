<template>
  <t-popup
    v-model:visible="showResults"
    trigger="click"
    placement="bottom-left"
    :overlay-style="{ width: '565px', padding: 0, overflow: 'hidden', borderRadius: 'var(--td-radius-medium)', boxShadow: '0 8px 30px rgba(0,0,0,.12)' }"
    :hide-on-click="false"
    @visible-change="onPopupVisibleChange"
  >
    <!-- side layout: input 始终可见 -->
    <div v-if="layout === 'side'" class="header-menu-search">
      <t-input
        v-model="searchData"
        class="header-search"
        :class="[{ 'hover-active': isSearchFocus }]"
        :placeholder="t('layout.searchPlaceholder')"
        clearable
        @focus="onInputFocus"
        @blur="onInputBlur"
        @keydown="onKeyDown"
        @clear="onClear"
      >
        <template #prefix-icon>
          <t-icon class="icon" name="search" size="16" />
        </template>
      </t-input>
    </div>

    <!-- non-side layout: 点击图标展开 input -->
    <div v-else class="header-menu-search-left">
      <t-button
        :class="{ 'search-icon-hide': isSearchFocus }"
        theme="default"
        shape="square"
        variant="text"
        @click="changeSearchFocus(true)"
      >
        <t-icon name="search" />
      </t-button>
      <t-input
        v-model="searchData"
        class="header-search"
        :class="[{ 'width-zero': !isSearchFocus }]"
        placeholder="输入要搜索内容"
        :autofocus="isSearchFocus"
        clearable
        @focus="onInputFocus"
        @blur="onInputBlur"
        @keydown="onKeyDown"
        @clear="onClear"
      >
        <template #prefix-icon>
          <t-icon name="search" size="16" />
        </template>
      </t-input>
    </div>

    <!-- 搜索下拉结果 -->
    <template #content>
      <div class="search-panel">
        <!-- 加载中 -->
        <div v-if="isSearching" class="search-loading">
          <t-icon name="loading" size="24" class="loading-icon" />
          <div class="loading-text">正在搜索 "{{ searchData }}"...</div>
          <div class="loading-sub">稍等片刻</div>
        </div>

        <!-- 有搜索内容时显示结果面板 -->
        <template v-else-if="searchData && searchData.trim()">
          <div class="search-layout">
            <!-- 左侧：分类列表 -->
            <div class="search-sidebar">
              <div
                v-for="cat in categoryList"
                :key="cat.key"
                class="category-item"
                :class="{ 'is-active': selectedCategory === cat.key }"
                @click="selectedCategory = cat.key"
              >
                <t-icon :name="cat.icon" size="16" class="category-icon" />
                <span class="category-label">{{ cat.label }}</span>
                <span class="category-count">{{ cat.count }}</span>
              </div>
            </div>

            <!-- 右侧：选中分类的结果 -->
            <div class="search-content">
              <template v-if="selectedCategory === 'menu' && searchResults.length > 0">
                <div class="result-list">
                  <div
                    v-for="(item, index) in searchResults"
                    :key="'menu-' + item.path"
                    class="result-card"
                    :class="{ 'is-active': index === activeIndex }"
                    @mousedown.prevent="navigateTo(item)"
                    @mouseenter="activeIndex = index"
                  >
                    <div class="result-card-title">
                      {{ item.title }}
                    </div>
                    <div class="result-card-sub">{{ item.moduleLabel }}</div>
                  </div>
                </div>
              </template>

              <template v-else-if="selectedCategory !== 'menu' && groupedApiResults[selectedCategory]">
                <div class="result-list">
                  <div
                    v-for="entry in groupedApiResults[selectedCategory]"
                    :key="'api-' + entry.item.id"
                    class="result-card"
                    :class="{ 'is-active': entry.flatIndex === activeIndex }"
                    @mousedown.prevent="navigateToApi(entry.item)"
                    @mouseenter="activeIndex = entry.flatIndex"
                  >
                    <div class="result-card-title">
                      {{ entry.item.title }}
                    </div>
                    <div class="result-card-sub">{{ entry.item.description }}</div>
                  </div>
                </div>
              </template>

              <div v-else class="content-empty">
                <t-icon name="search" size="32" />
                <span>暂无结果</span>
              </div>
            </div>
          </div>
        </template>

        <!-- 搜索无结果 -->
        <div
          v-else-if="searchResults.length === 0 && !isSearching && hasSearched && apiResults.length === 0"
          class="search-empty"
        >
          <t-icon name="close-circle" size="24" />
          <span>未找到匹配的结果</span>
        </div>

        <!-- 初始提示 -->
        <div v-else-if="!searchData" class="search-empty">
          <t-icon name="search" size="24" />
          <span>输入关键词搜索菜单和业务数据</span>
        </div>
      </div>
    </template>
  </t-popup>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import { useDebounceFn } from '@vueuse/core';

import { t } from '@/locales';
import { usePermissionStore } from '@/store';
import { globalSearch, type SearchResultItem as ApiSearchResult } from '@/api/search';

defineProps({
  layout: {
    type: String,
    default: '',
  },
});

interface SearchResultItem {
  path: string;
  title: string;
  icon: string;
  moduleLabel: string;
}

const router = useRouter();
const permissionStore = usePermissionStore();

const isSearchFocus = ref(false);
const searchData = ref('');
const activeIndex = ref(0);
const showResults = ref(false);
const apiResults = ref<ApiSearchResult[]>([]);
const isSearching = ref(false);
const hasSearched = ref(false);
const selectedCategory = ref('');

const moduleIcons: Record<string, string> = {
  mold: 'chart-3d',
  mold_order: 'chart-3d',
  mold_work_order: 'chart-3d',
  product: 'layers',
  customer: 'user',
  production_order: 'tools',
  master_order: 'view-list',
};



/** 实际搜索逻辑 */
async function doSearch(keyword: string): Promise<void> {
  if (keyword.trim().length < 2) {
    apiResults.value = [];
    isSearching.value = false;
    hasSearched.value = false;
    return;
  }
  isSearching.value = true;
  hasSearched.value = true;
  try {
    const res = await globalSearch({ keyword: keyword.trim(), page_size: 10 });
    apiResults.value = res.items || [];
  } catch {
    apiResults.value = [];
  } finally {
    isSearching.value = false;
  }
}

/** 防抖搜索：输入后 300ms 自动调用 */
const debouncedSearch = useDebounceFn(doSearch, 300);

/** 手动触发搜索（按回车时调用，无视阈值直接搜索） */
const triggerSearch = () => {
  const keyword = searchData.value.trim();
  if (!keyword) return;
  doSearch(keyword);
};

/** 切换左边分类时重置选中索引 */
watch(selectedCategory, () => {
  activeIndex.value = 0;
});

/** 监听输入变化触发后端搜索（防抖 300ms） */
watch(searchData, (val) => {
  debouncedSearch(val.trim());
});

const changeSearchFocus = (value: boolean) => {
  isSearchFocus.value = value;
  if (!value) {
    searchData.value = '';
  }
};

const onInputFocus = () => {
  isSearchFocus.value = true;
  // popup 由 trigger="click" 的点击切换控制，focus 不干涉
};

const onInputBlur = () => {
  // 不在此处关闭 popup，由外部点击自动处理
};

const onPopupVisibleChange = (visible: boolean) => {
  if (!visible) {
    // 只收起搜索界面，不清空输入内容
    isSearchFocus.value = false;
  }
};

/** 清空输入时重置状态 */
const onClear = () => {
  apiResults.value = [];
  hasSearched.value = false;
  activeIndex.value = 0;
};

/** 将路由树扁平化为可搜索的列表（只保留非 hidden 的非目录叶子节点） */
function flattenRoutes(
  routes: RouteRecordRaw[],
  parentPath = '',
  parentLabel = '',
  parentIcon = '',
): SearchResultItem[] {
  const result: SearchResultItem[] = [];
  for (const route of routes) {
    const meta = route.meta as Record<string, unknown> | undefined;
    if (meta?.hidden) continue;

    const titleObj = meta?.title;
    const title = typeof titleObj === 'object' && titleObj !== null
      ? ((titleObj as Record<string, string>).zh_CN || (titleObj as Record<string, string>).en_US || String(titleObj))
      : String(titleObj || route.name || '');

    // 图标：子路由优先继承父级图标
    const icon = (meta?.icon as string) || parentIcon;

    const routePath = route.path || '';
    const fullPath = routePath.startsWith('/') ? routePath : `${parentPath}/${routePath}`;

    if (route.children && route.children.length > 0) {
      // 有子路由：递归处理子项，当前路由名作为父级标签
      const currentLabel = parentLabel ? `${parentLabel} / ${title}` : title;
      result.push(...flattenRoutes(route.children, fullPath, currentLabel, icon));
    } else if (routePath) {
      // 叶子路由（实际页面）
      result.push({
        path: fullPath,
        title,
        icon,
        moduleLabel: parentLabel,
      });
    }
  }
  return result;
}

/** 当前用户有权访问的扁平页面列表 */
const searchablePages = computed(() => flattenRoutes(permissionStore.routers as RouteRecordRaw[]));

/** 搜索结果（本地模糊匹配） */
const searchResults = computed(() => {
  const keyword = searchData.value.trim().toLowerCase();
  if (!keyword) return [];

  activeIndex.value = 0;

  return searchablePages.value.filter((page) => {
    return (
      page.title.toLowerCase().includes(keyword)
      || page.moduleLabel.toLowerCase().includes(keyword)
    );
  });
});

/** API 结果按模块分组，并预计算每组内每个条目的全局扁平索引 */
interface ApiGroupedEntry {
  flatIndex: number;
  item: ApiSearchResult;
}
const groupedApiResults = computed(() => {
  const groups: Record<string, ApiGroupedEntry[]> = {};
  let idx = searchResults.value.length; // 排在菜单结果之后
  for (const item of apiResults.value) {
    const moduleName = item.module_name;
    if (!groups[moduleName]) {
      groups[moduleName] = [];
    }
    groups[moduleName].push({ flatIndex: idx, item });
    idx++;
  }
  return groups;
});

/** 分类列表（左侧树形导航） */
interface CategoryItem {
  key: string;
  label: string;
  icon: string;
  count: number;
}
const categoryList = computed<CategoryItem[]>(() => {
  const cats: CategoryItem[] = [];
  // 页面导航
  if (searchResults.value.length > 0) {
    cats.push({ key: 'menu', label: '页面导航', icon: 'view-list', count: searchResults.value.length });
  }
  // API 分组
  const apiGroups = groupedApiResults.value;
  for (const moduleName of Object.keys(apiGroups)) {
    const entries = apiGroups[moduleName];
    cats.push({
      key: moduleName,
      label: moduleName,
      icon: moduleIcons[entries[0].item.module] || 'search',
      count: entries.length,
    });
  }
  return cats;
});

/** 当分类列表变化时，如果当前选中的分类不在列表中，自动选中第一个 */
watch(categoryList, (list) => {
  if (list.length === 0) return;
  const exists = list.some((cat) => cat.key === selectedCategory.value);
  if (!exists) {
    selectedCategory.value = list[0].key;
  }
});

/** 所有可选中的结果项（菜单 + API）的扁平列表，用于键盘导航 */
interface FlatSelectableItem {
  type: 'menu' | 'api';
  path: string;
}
const flatAllItems = computed<FlatSelectableItem[]>(() => {
  const items: FlatSelectableItem[] = [];
  for (const item of searchResults.value) {
    items.push({ type: 'menu', path: item.path });
  }
  for (const item of apiResults.value) {
    items.push({ type: 'api', path: item.path });
  }
  return items;
});

/** 导航到菜单页面 */
const navigateTo = (item: SearchResultItem) => {
  showResults.value = false;
  searchData.value = '';
  isSearchFocus.value = false;
  router.push(item.path);
};

/** 导航到 API 搜索结果页面 */
const navigateToApi = (item: ApiSearchResult) => {
  showResults.value = false;
  searchData.value = '';
  isSearchFocus.value = false;
  router.push(item.path);
};

/** 统一键盘事件处理（解决 IME 输入法下 .enter 修饰符不匹配的问题） */
const onKeyDown = (_value: unknown, context: { e: KeyboardEvent }) => {
  const e = context.e;
  if (e.key === 'Enter') {
    e.preventDefault();
    const flatItem = flatAllItems.value[activeIndex.value];
    if (flatItem) {
      // 有选中项 -> 导航
      showResults.value = false;
      searchData.value = '';
      isSearchFocus.value = false;
      router.push(flatItem.path);
    } else {
      // 无选中项 -> 手动搜索（无视字符数阈值）
      triggerSearch();
    }
  } else if (e.key === 'ArrowUp') {
    e.preventDefault();
    activeIndex.value = Math.max(0, activeIndex.value - 1);
  } else if (e.key === 'ArrowDown') {
    e.preventDefault();
    activeIndex.value = Math.min(flatAllItems.value.length - 1, activeIndex.value + 1);
  }
};
</script>

<style lang="less" scoped>
.header-menu-search {
  display: flex;
  margin-left: 16px;

  .hover-active {
    background: var(--td-bg-color-secondarycontainer);
  }

  .t-icon {
    color: var(--td-text-color-primary) !important;
  }

  .header-search {
    :deep(.t-input) {
      border: none;
      outline: none;
      box-shadow: none;
      transition: background @anim-duration-base linear;

      .t-input__inner {
        transition: background @anim-duration-base linear;
        background: none;
      }

      &:hover {
        background: var(--td-bg-color-secondarycontainer);

        .t-input__inner {
          background: var(--td-bg-color-secondarycontainer);
        }
      }
    }
  }
}

.t-button {
  margin: 0 8px;
  transition: opacity @anim-duration-base @anim-time-fn-easing;

  .t-icon {
    font-size: 20px;

    &.general {
      margin-right: 16px;
    }
  }
}

.search-icon-hide {
  opacity: 0;
}

.header-menu-search-left {
  display: flex;
  align-items: center;

  .header-search {
    width: 280px;
    transition: width @anim-duration-base @anim-time-fn-easing;

    :deep(.t-input) {
      border: 0;

      &:focus {
        box-shadow: none;
      }
    }

    &.width-zero {
      width: 0;
      opacity: 0;
    }
  }
}

/* 搜索结果面板 */
.search-panel {
  width: 550px;
  padding: 0;
  display: flex;
  flex-direction: column;
  border-radius: var(--td-radius-medium);
  overflow: hidden;
}

/* 左右分栏布局 */
.search-layout {
  display: flex;
  min-height: 200px;
  max-height: 480px;
  overflow: hidden;
}

/* 左侧分类栏 */
.search-sidebar {
  width: 160px;
  flex-shrink: 0;
  padding: 8px 0 8px 8px;
  overflow-y: auto;
}

.category-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  margin-right: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--td-text-color-primary);
  border-radius: var(--td-radius-default);
  transition: all @anim-duration-base linear;

  &:hover {
    background: var(--td-bg-color-secondarycontainer);
  }

  &.is-active {
    color: var(--td-brand-color);
    background: var(--td-bg-color-secondarycontainer);
    font-weight: 500;

    .category-count {
      background: var(--td-brand-color);
      color: #fff;
    }
  }

  .category-icon {
    flex-shrink: 0;
  }

  .category-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .category-count {
    font-size: 11px;
    color: var(--td-text-color-placeholder);
    background: var(--td-bg-color-component);
    border-radius: var(--td-radius-circle);
    padding: 0 5px;
    line-height: 16px;
    min-width: 16px;
    text-align: center;
  }
}

/* 右侧结果区域 */
.search-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  min-height: 0;
}

.result-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0;
}

.result-card {
  cursor: pointer;
  padding: 10px 12px;
  border-radius: var(--td-radius-medium);
  background: var(--td-bg-color-container);
  transition: all @anim-duration-base linear;

  &:hover {
    background: var(--td-bg-color-secondarycontainer);
  }

  &.is-active {
    background: var(--td-bg-color-secondarycontainer);
  }

  .result-card-title {
    font-size: 14px;
    color: var(--td-text-color-primary);
    font-weight: 500;
    line-height: 1.5;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .result-card-sub {
    font-size: 12px;
    color: var(--td-text-color-placeholder);
    line-height: 1.5;
    margin-top: 2px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

/* 内容区空状态 */
.content-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 48px 16px;
  color: var(--td-text-color-placeholder);
  font-size: 13px;
}

/* 搜索加载状态 */
.search-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 48px 16px;
  color: var(--td-text-color-primary);
  text-align: center;

  .loading-icon {
    animation: spin 1s linear infinite;
  }

  .loading-text {
    font-size: 14px;
    font-weight: 500;
  }

  .loading-sub {
    font-size: 12px;
    color: var(--td-text-color-placeholder);
  }
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 空状态 */
.search-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 40px 16px;
  color: var(--td-text-color-placeholder);
  font-size: 14px;
}
</style>
