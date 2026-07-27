# 字典颜色工具类统一改造总结

## 改造背景

项目中多个页面存在重复的字典颜色映射代码，导致：
1. 代码冗余，维护困难
2. 颜色定义不一致，可能出现显示差异
3. 新增颜色需要在多处修改，容易遗漏

## 改造方案

### 1. 创建统一工具类

**文件位置**: `/web/src/utils/dict-color.ts`

**提供的功能**:
- `getDictColorHex()` - 获取颜色的 Hex 值
- `getDictColorName()` - 获取颜色的中文名称
- `getDictColorTheme()` - 获取 TDesign 主题类型
- `isLightColor()` - 判断是否为浅色背景
- `getDictColorTextColor()` - 获取适配的文字颜色
- `getDictColorStyle()` - 获取完整样式对象
- `renderDictColorValue()` - 渲染颜色选择器显示内容

**支持的颜色**:
- 功能色: default, primary, warning, success, danger
- 中性色: gray-7, gray-9, gray-11, gray-14
- 扩展色: blue-8, cyan-7, purple-7, orange-5
- 兼容旧版本: gray-1/3/5, blue-1/4/7, cyan-1/4/5, purple-1/4, orange-4

### 2. 已更新的页面

#### ✅ dict/index.vue (主字典管理)
- 删除了本地定义的 `getColorHex`, `getColorName`, `isLightColor`, `renderColorValue` 函数
- 导入并使用统一工具类的函数
- 减少了 107 行重复代码

#### ✅ mold-dict/index.vue (模具字典管理)
- 删除了本地定义的 `getColorHex`, `getColorName`, `isLightColor` 函数
- 更新了所有模板中的颜色相关调用
- 减少了 49 行重复代码

### 3. 待更新的页面

以下页面仍使用本地定义，建议按优先级逐步迁移：

**高优先级**（字典管理相关）:
- `/web/src/pages/system/production-dict/index.vue`
- `/web/src/pages/system/status-dict/index.vue`
- `/web/src/pages/system/production-status-dict/index.vue`

**中优先级**（状态流转配置）:
- `/web/src/pages/system/status-flow-config.vue`

**低优先级**（业务页面）:
- `/web/src/pages/production/order-list/index.vue`
- `/web/src/pages/production/order-detail/index.vue`

## 使用方法

### 基本用法

```typescript
import { 
  getDictColorHex,
  getDictColorName,
  getDictColorStyle,
  renderDictColorValue 
} from '@/utils/dict-color';

// 获取 Hex 值
const hex = getDictColorHex('primary'); // '#0052d9'

// 获取颜色名称
const name = getDictColorName('primary'); // '品牌蓝'

// 获取完整样式
const style = getDictColorStyle('primary');
// { background: '#0052d9', color: '#ffffff', borderColor: '#0052d9' }
```

### Vue 模板中使用

```vue
<template>
  <!-- 方式1: 使用完整样式 -->
  <t-tag :style="getDictColorStyle(row.color)">
    {{ getDictColorName(row.color) }}
  </t-tag>
  
  <!-- 方式2: 使用主题类型 -->
  <t-tag :theme="getDictColorTheme(row.color)">
    {{ row.dict_label }}
  </t-tag>
  
  <!-- 方式3: 颜色选择器 -->
  <t-select 
    v-model="form.color" 
    :value-display="renderDictColorValue"
  >
    <!-- options -->
  </t-select>
</template>

<script setup lang="ts">
import { 
  getDictColorStyle, 
  getDictColorName,
  getDictColorTheme,
  renderDictColorValue 
} from '@/utils/dict-color';
</script>
```

## 改造收益

### 1. 代码量减少
- 已删除约 156 行重复代码
- 预计全部迁移后可再删除 300+ 行重复代码

### 2. 维护性提升
- 颜色定义集中在一处，修改只需改一个地方
- 新增颜色只需在工具类中添加映射
- 避免不同页面颜色定义不一致的问题

### 3. 一致性保证
- 所有页面使用相同的颜色映射
- 统一的文字颜色对比度处理
- 统一的用户体验

### 4. 可扩展性
- 易于添加新颜色
- 易于调整现有颜色的映射
- 提供了完整的工具函数集

## 后续工作

### 短期（1-2周）
1. 更新剩余的字典管理页面
2. 更新状态流转配置页面
3. 测试所有更新页面的功能

### 中期（1个月）
1. 更新业务页面中的颜色使用
2. 审查是否有其他地方使用了硬编码的颜色值
3. 完善工具类的文档和示例

### 长期
1. 考虑将颜色配置提取到配置文件
2. 支持动态配置颜色（从后端获取）
3. 提供颜色预览组件

## 注意事项

1. **不要在各组件中重新定义颜色映射** - 这是本次改造的核心原则
2. **如需新增颜色** - 在 `dict-color.ts` 中添加，并更新所有相关映射表
3. **保持向后兼容** - 修改时要确保不影响现有数据
4. **遵循 TypeScript 类型规范** - 工具类已提供完整的类型定义

## 相关文件

- 工具类: `/web/src/utils/dict-color.ts`
- 使用指南: `/web/src/utils/DICT_COLOR_USAGE.md`
- 数据库模型: `/server/internal/model/entity/dict.go`
- 初始化脚本: `/server/migrations/init_database_complete.sql`

## 技术细节

### 颜色映射策略

1. **Hex 值映射** (`DICT_COLOR_HEX_MAP`)
   - 定义了所有颜色标识到十六进制值的映射
   - 包含深色和浅色版本

2. **中文名称映射** (`DICT_COLOR_NAME_MAP`)
   - 提供用户友好的颜色名称
   - 用于界面显示

3. **TDesign 主题映射** (`DICT_COLOR_THEME_MAP`)
   - 将字典颜色映射到 TDesign 的主题类型
   - 用于 t-tag 等组件的 theme 属性

4. **浅色背景列表** (`LIGHT_COLOR_LIST`)
   - 定义哪些颜色是浅色背景
   - 用于自动调整文字颜色

### 文字颜色对比度

- 浅色背景: 使用 `rgba(0, 0, 0, 0.9)` (深灰色)
- 深色背景: 使用 `#ffffff` (白色)
- 确保良好的可读性和无障碍访问

## 总结

通过本次改造，我们建立了统一的字典颜色管理机制，大幅减少了代码冗余，提高了可维护性。后续应继续推进其他页面的迁移工作，最终实现整个项目的颜色管理统一化。
