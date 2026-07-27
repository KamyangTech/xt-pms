import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';

// 扩展 dayjs 插件
if (dayjs.extend) {
  dayjs.extend(utc);
  dayjs.extend(timezone);
}

// 全局配置时区为中国东八区
dayjs.tz.setDefault('Asia/Shanghai');

export const LAST_7_DAYS = [
  dayjs().subtract(7, 'day').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD'),
];

export const LAST_30_DAYS = [
  dayjs().subtract(30, 'day').format('YYYY-MM-DD'),
  dayjs().format('YYYY-MM-DD'),
];

/**
 * 格式化日期时间为 YYYY-MM-DD HH:mm:ss（使用东八区时区）
 * @param dateStr 日期字符串或时间戳
 * @returns 格式化后的日期时间字符串，如果输入无效则返回 '-'
 */
export function formatDateTime(dateStr: string | number | Date | null | undefined): string {
  if (!dateStr) return '-';
  try {
    // 使用 dayjs.tz 确保时区正确
    return dayjs(dateStr).tz('Asia/Shanghai').format('YYYY-MM-DD HH:mm:ss');
  } catch (error) {
    console.error('日期时间格式化错误:', error, '原始值:', dateStr);
    return String(dateStr);
  }
}

/**
 * 格式化日期为 YYYY-MM-DD（使用东八区时区）
 * @param dateStr 日期字符串或时间戳
 * @returns 格式化后的日期字符串，如果输入无效则返回 '-'
 */
export function formatDate(dateStr: string | number | Date | null | undefined): string {
  if (!dateStr) return '-';
  try {
    // 使用 dayjs.tz 确保时区正确
    return dayjs(dateStr).tz('Asia/Shanghai').format('YYYY-MM-DD');
  } catch (error) {
    console.error('日期格式化错误:', error, '原始值:', dateStr);
    return String(dateStr);
  }
}

/**
 * 获取当前时间的格式化字符串（YYYY-MM-DD HH:mm:ss）
 * @returns 格式化后的当前时间
 */
export function getCurrentDateTime(): string {
  return dayjs().tz('Asia/Shanghai').format('YYYY-MM-DD HH:mm:ss');
}

/**
 * 获取当前日期（YYYY-MM-DD）
 * @returns 格式化后的当前日期
 */
export function getCurrentDate(): string {
  return dayjs().tz('Asia/Shanghai').format('YYYY-MM-DD');
}
