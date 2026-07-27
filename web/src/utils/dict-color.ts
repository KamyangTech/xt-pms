/**
 * 字典颜色工具类（已废弃）
 * @deprecated 请使用 @/config/dict-color-config 替代
 * 此文件保留仅为向后兼容
 */

// 导入新配置中的所有函数和常量
import {
  COLOR_HEX_MAP,
  COLOR_NAME_MAP,
  COLOR_CATEGORIES,
  getColorHex,
  getTextColor,
  getColorName,
  isLightColor,
  ColorCategory,
} from '@/config/dict-color-config';

// 重新导出新配置中的所有函数和常量，保持向后兼容
export {
  COLOR_HEX_MAP as DICT_COLOR_HEX_MAP,
  COLOR_NAME_MAP as DICT_COLOR_NAME_MAP,
  COLOR_CATEGORIES,
  ColorCategory,
  type ColorOption,
  getAllColorOptions,
  getFlatColorOptions,
  getColorHex as getDictColorHex,
  getColorName as getDictColorName,
  isLightColor,
  getTextColor as getDictColorTextColor,
} from '@/config/dict-color-config';

/**
 * 获取 TDesign Tag 主题
 * @deprecated 请直接使用颜色值设置样式
 */
export function getDictColorTheme(colorValue: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' {
  // 功能色映射到 TDesign 主题
  const themeMap: Record<string, 'default' | 'primary' | 'success' | 'warning' | 'danger'> = {
    'default': 'default',
    'primary': 'primary',
    'warning': 'warning',
    'success': 'success',
    'danger': 'danger',
  };
  
  return themeMap[colorValue] || 'default';
}

/**
 * 获取字典颜色样式对象
 * @deprecated 建议直接在模板中使用内联样式或CSS变量
 */
export function getDictColorStyle(colorValue: string) {
  return {
    background: getColorHex(colorValue),
    color: getTextColor(colorValue),
    borderColor: getColorHex(colorValue),
  };
}

/**
 * 获取字典颜色的渲染配置（用于 outline 变体）
 * 专门用于 variant="outline" 的情况，返回透明背景+边框+文字颜色
 * @param color 颜色标识
 * @returns { mode: 'theme' | 'style', theme?: string, style?: object }
 */
export function getDictColorRenderConfigForOutline(color: string): {
  mode: 'theme' | 'style';
  theme?: 'default' | 'primary' | 'success' | 'warning' | 'danger';
  style?: {
    background: string;
    color: string;
    borderColor: string;
  };
} {
  // 如果是 TDesign 标准主题色，使用 theme 属性
  const TDESIGN_THEME_COLORS = ['default', 'primary', 'success', 'warning', 'danger'];
  if (TDESIGN_THEME_COLORS.includes(color)) {
    return {
      mode: 'theme',
      theme: color as 'default' | 'primary' | 'success' | 'warning' | 'danger',
    };
  }
  
  // 其他颜色使用自定义 style（outline 模式：透明背景 + 边框 + 文字颜色）
  const borderColor = getColorHex(color);
  const textColor = getColorHex(color); // 文字颜色与边框颜色一致
  
  return {
    mode: 'style',
    style: {
      background: 'transparent',
      color: textColor,
      borderColor: borderColor,
    },
  };
}

/**
 * 获取字典颜色的渲染配置
 * 智能判断：功能色返回 theme，扩展色返回 style
 * @param color 颜色标识
 * @returns { mode: 'theme' | 'style', theme?: string, style?: object }
 *   - mode === 'theme' 时同时返回 style（基于主题色计算的实际 hex 样式），
 *     保证通过 `config.style` 取样式也能拿到正确值（兼容旧调用）
 *   - mode === 'style' 时仅返回 style（无对应主题）
 */
export function getDictColorRenderConfig(color: string): {
  mode: 'theme' | 'style';
  theme?: 'default' | 'primary' | 'success' | 'warning' | 'danger';
  style?: {
    background: string;
    color: string;
    borderColor: string;
  };
} {
  // 如果是 TDesign 标准主题色，使用 theme 属性
  const TDESIGN_THEME_COLORS = ['default', 'primary', 'success', 'warning', 'danger'];
  if (TDESIGN_THEME_COLORS.includes(color)) {
    const bgColor = getColorHex(color);
    const textColor = getTextColor(color);
    return {
      mode: 'theme',
      theme: color as 'default' | 'primary' | 'success' | 'warning' | 'danger',
      // 同时返回基于 hex 的 style，兼容只读 .style 的旧调用
      style: {
        background: bgColor,
        color: textColor,
        borderColor: bgColor,
      },
    };
  }

  // 其他颜色使用自定义 style
  const bgColor = getColorHex(color);
  const textColor = getTextColor(color);

  return {
    mode: 'style',
    style: {
      background: bgColor,
      color: textColor,
      borderColor: bgColor,
    },
  };
}

/**
 * 获取字典颜色的 CSS 样式字符串（用于 :style 绑定）
 * @param color 颜色标识
 * @returns CSS 样式字符串
 */
export function getDictColorStyleString(color: string): string {
  const bgColor = getColorHex(color);
  const textColor = getTextColor(color);
  
  return `background: ${bgColor}; color: ${textColor}; border-color: ${bgColor};`;
}
