/**
 * 统一颜色配置
 * 集中管理所有字典颜色选项，避免硬编码和重复定义
 */

/**
 * 颜色分类枚举
 */
export enum ColorCategory {
  FUNCTIONAL = 'functional',    // 功能色
  NEUTRAL = 'neutral',          // 中性色
  EXTENDED = 'extended',        // 扩展色
}

/**
 * 颜色选项接口
 */
export interface ColorOption {
  value: string;
  label: string;
  hex: string;
  category: ColorCategory;
}

/**
 * 颜色配置映射表
 * 标识 -> Hex值
 */
export const COLOR_HEX_MAP: Record<string, string> = {
  // 功能色（深色版本，确保白色文字对比度好）
  'default': '#999999',
  'primary': '#0052d9',      // 品牌色-7
  'warning': '#e37318',      // 警告色-5（深橙）
  'success': '#008858',      // 成功色-6（深绿）
  'danger': '#ad352f',       // 错误色-7（深红）
  
  // 中性色（深色版本）
  'gray-7': '#8b8b8b',
  'gray-9': '#5e5e5e',
  'gray-11': '#383838',
  'gray-14': '#181818',
  
  // 扩展色（深色版本）
  'blue-8': '#003cab',       // 品牌色-8（更深蓝）
  'cyan-7': '#00a38a',       // 青色-7（深青）
  'purple-7': '#7a34d1',     // 紫色-7（深紫）
  'orange-5': '#e37318',     // 橙色-5（深橙）
  
  // 新增扩展色（用于更多分类场景）
  'pink-5': '#e91e63',       // 粉色
  'red-7': '#d32f2f',        // 红色
  'green-6': '#43a047',      // 绿色
  'yellow-5': '#fbc02d',     // 黄色
  'indigo-7': '#3949ab',     // 靛蓝
  'teal-6': '#00897b',       // 青色（蓝绿）
  'lime-6': '#c0ca33',       // 黄绿
  'amber-5': '#ffc107',      // 琥珀色
};

/**
 * 颜色名称映射表
 * 标识 -> 显示名称
 */
export const COLOR_NAME_MAP: Record<string, string> = {
  // 功能色
  'default': '灰色',
  'primary': '品牌蓝',
  'warning': '深橙色',
  'success': '深绿色',
  'danger': '深红色',
  
  // 中性色
  'gray-7': '深灰',
  'gray-9': '极深灰',
  'gray-11': '深黑',
  'gray-14': '纯黑',
  
  // 扩展色
  'blue-8': '更深蓝',
  'cyan-7': '深青',
  'purple-7': '深紫',
  'orange-5': '深橙',
  
  // 新增扩展色
  'pink-5': '粉色',
  'red-7': '红色',
  'green-6': '绿色',
  'yellow-5': '黄色',
  'indigo-7': '靛蓝',
  'teal-6': '青绿',
  'lime-6': '黄绿',
  'amber-5': '琥珀色',
};

/**
 * 颜色分类配置
 */
export const COLOR_CATEGORIES: Record<ColorCategory, {
  title: string;
  colors: string[];
}> = {
  [ColorCategory.FUNCTIONAL]: {
    title: '功能色',
    colors: ['default', 'primary', 'warning', 'success', 'danger'],
  },
  [ColorCategory.NEUTRAL]: {
    title: '中性色',
    colors: ['gray-7', 'gray-9', 'gray-11', 'gray-14'],
  },
  [ColorCategory.EXTENDED]: {
    title: '扩展色',
    colors: [
      'blue-8', 'cyan-7', 'purple-7', 'orange-5',  // 原有
      'pink-5', 'red-7', 'green-6', 'yellow-5',    // 新增
      'indigo-7', 'teal-6', 'lime-6', 'amber-5',   // 新增
    ],
  },
};

/**
 * 获取所有颜色选项（按分类组织）
 */
export function getAllColorOptions(): Array<{
  group: string;
  options: ColorOption[];
}> {
  return Object.entries(COLOR_CATEGORIES).map(([categoryKey, category]) => ({
    group: category.title,
    options: category.colors.map((value) => ({
      value,
      label: `${COLOR_NAME_MAP[value]}（${COLOR_HEX_MAP[value]}）`,
      hex: COLOR_HEX_MAP[value],
      category: categoryKey as ColorCategory,
    })),
  }));
}

/**
 * 获取扁平化的颜色选项列表（用于下拉选择器）
 */
export function getFlatColorOptions(): ColorOption[] {
  const allOptions: ColorOption[] = [];
  
  Object.entries(COLOR_CATEGORIES).forEach(([categoryKey, category]) => {
    category.colors.forEach((value) => {
      allOptions.push({
        value,
        label: `${COLOR_NAME_MAP[value]}（${COLOR_HEX_MAP[value]}）`,
        hex: COLOR_HEX_MAP[value],
        category: categoryKey as ColorCategory,
      });
    });
  });
  
  return allOptions;
}

/**
 * 根据颜色标识获取Hex值
 */
export function getColorHex(colorValue: string): string {
  return COLOR_HEX_MAP[colorValue] || COLOR_HEX_MAP['default'];
}

/**
 * 根据颜色标识获取显示名称
 */
export function getColorName(colorValue: string): string {
  return COLOR_NAME_MAP[colorValue] || COLOR_NAME_MAP['default'];
}

/**
 * 判断是否为浅色（用于决定文字颜色）
 */
export function isLightColor(colorValue: string): boolean {
  const hex = getColorHex(colorValue);
  if (!hex) return false;
  
  // 移除 # 号
  const color = hex.replace('#', '');
  
  const r = parseInt(color.substr(0, 2), 16);
  const g = parseInt(color.substr(2, 2), 16);
  const b = parseInt(color.substr(4, 2), 16);
  
  const brightness = (r * 299 + g * 587 + b * 114) / 1000;
  
  // 亮度大于150认为是浅色（提高阈值，避免中等亮度颜色误判）
  return brightness > 150;
}

/**
 * 获取适合的文字颜色（黑色或白色）
 */
export function getTextColor(colorValue: string): string {
  return isLightColor(colorValue) ? 'rgba(0, 0, 0, 0.9)' : '#fff';
}
