/**
 * 操作列宽度计算工具
 * 
 * 方案：固定宽度容器 + t-space 自动换行
 * - 每行固定显示 4 个按钮
 * - 超过 4 个按钮自动换行到第二行
 * - 使用固定宽度容器控制换行位置
 */

/**
 * 单个按钮的平均宽度（包含间距）
 * - 按钮文字宽度：约 35px（2-4个汉字）
 * - TDesign t-link 默认 padding：约 8px
 * - 平均每个按钮占位：约 45px
 */
const BUTTON_AVG_WIDTH = 45;

/**
 * 每行按钮之间的间距
 */
const BUTTON_GAP = 8;

/**
 * 容器左右内边距
 */
const CONTAINER_PADDING = 16;

/**
 * 每行最大按钮数量
 */
const MAX_BUTTONS_PER_ROW = 4;

/**
 * 操作列固定宽度配置
 * 根据最大可能按钮数量设置合适的固定宽度
 * 
 * 计算公式：
 * - 容器宽度 = 按钮数 × 按钮平均宽度 + (按钮数-1) × 间距 + 容器内边距
 * - 例如：4个按钮 = 4×45 + 3×8 + 16 = 180 + 24 + 16 = 220px
 */
export const OPERATION_COLUMN_WIDTH = {
  SMALL: 130,   // 2个按钮：2×45 + 1×8 + 16 = 114px ≈ 130px
  MEDIUM: 180,  // 3个按钮：3×45 + 2×8 + 16 = 167px ≈ 180px
  LARGE: 220,   // 4个按钮：4×45 + 3×8 + 16 = 220px
  XLARGE: 220,  // 5-6个按钮：第一行4个(220px)，第二行自动换行
} as const;

/**
 * 获取操作列宽度
 * 
 * @param buttonCount 最大可能按钮数量
 * @returns 操作列宽度（px）
 */
export const getOperationColumnWidth = (buttonCount: number): number => {
  if (buttonCount <= 2) return OPERATION_COLUMN_WIDTH.SMALL;
  if (buttonCount <= 3) return OPERATION_COLUMN_WIDTH.MEDIUM;
  if (buttonCount <= 4) return OPERATION_COLUMN_WIDTH.LARGE;
  return OPERATION_COLUMN_WIDTH.XLARGE; // 5-6个按钮，使用相同宽度，自动换行
};

/**
 * 兼容旧版本的简化估算函数（已废弃，建议使用 getOperationColumnWidth）
 * 
 * @deprecated 使用 getOperationColumnWidth 代替
 * @param buttonCount 按钮数量
 * @returns 操作列宽度（px）
 */
export const estimateOperationWidth = (buttonCount: number): number => {
  return getOperationColumnWidth(buttonCount);
};
