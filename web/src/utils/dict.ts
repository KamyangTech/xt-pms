/**
 * 字典管理工具类
 * 提供统一的字典数据加载、缓存和查询功能
 * 
 * 使用示例：
 * ```typescript
 * // 1. 获取字典选项（用于表单下拉框）
 * const options = await dictUtils.getDictOptions('production-order-status')
 * // 返回: [{ label: '待排产', value: 'PENDING_SCHEDULE', color: '#1890ff' }, ...]
 * 
 * // 2. 获取字典标签（用于列表显示）
 * const label = dictUtils.getDictLabel('production-order-status', 'PENDING_SCHEDULE')
 * // 返回: '待排产'
 * 
 * // 3. 获取字典颜色（用于状态标签）
 * const color = dictUtils.getDictColor('production-order-status', 'PENDING_SCHEDULE')
 * // 返回: '#1890ff'
 * ```
 */

import { getDictDataByType } from '@/api/reference';

// 字典数据类型定义
export interface DictItem {
  dict_label: string;      // 字典标签（显示文本）
  dict_value: string;      // 字典值（语义化字符串）
  css_class?: string;      // CSS 类名/颜色
  list_class?: string;     // 列表样式类
  color?: string;          // 颜色标识（TDesign主题色名或扩展色名）
  ext?: Record<string, any> | string;  // 扩展配置（JSON字符串或对象）
  status?: number;         // 状态：1=启用，0=禁用
  is_default?: number;     // 是否默认
  sort?: number;           // 排序
  remark?: string;         // 备注
}

// 字典选项类型（用于表单）
export interface DictOption {
  label: string;
  value: string;
  color?: string;
}

// 字典缓存
class DictCache {
  private cache: Map<string, DictItem[]> = new Map();
  private loading: Map<string, Promise<DictItem[]>> = new Map();

  /**
   * 同步获取字典数据（仅返回已缓存的数据）
   */
  getDictDataSync(dictType: string): DictItem[] {
    return this.cache.get(dictType) || [];
  }

  /**
   * 获取字典数据（带缓存）
   * @param dictType 字典类型
   * @param forceRefresh 是否强制刷新
   */
  async getDictData(dictType: string, forceRefresh = false): Promise<DictItem[]> {
    // 如果缓存中存在且不强制刷新，直接返回
    if (!forceRefresh && this.cache.has(dictType)) {
      return this.cache.get(dictType)!;
    }

    // 如果正在加载中，返回同一个 Promise（避免重复请求）
    if (this.loading.has(dictType)) {
      return this.loading.get(dictType)!;
    }

    const loadPromise = this.loadDictData(dictType);
    this.loading.set(dictType, loadPromise);

    try {
      const data = await loadPromise;
      this.cache.set(dictType, data);
      return data;
    } finally {
      this.loading.delete(dictType);
    }
  }

  /**
   * 从 API 加载字典数据
   */
  private async loadDictData(dictType: string): Promise<DictItem[]> {
    try {
      const res = await getDictDataByType({ dict_type: dictType });
      
      let dataList: DictItem[] = [];
      
      // 情况1: res.data 直接是 {list: [...], total: N}
      if (res.data?.list) {
        dataList = res.data.list;
      }
      // 情况2: res.data 是数组
      else if (Array.isArray(res.data)) {
        dataList = res.data;
      }
      // 情况3: res 本身包含 list 字段（可能 request 拦截器已经解包）
      else if (res.list) {
        dataList = res.list;
      }
      // 情况4: res 本身是数组
      else if (Array.isArray(res)) {
        dataList = res;
      }

      // 按 sort 排序
      return dataList.sort((a, b) => (a.sort || 0) - (b.sort || 0));
    } catch (error) {
      console.error(`加载字典 ${dictType} 失败:`, error);
      return [];
    }
  }

  /**
   * 清除指定字典的缓存
   */
  clearCache(dictType?: string) {
    if (dictType) {
      this.cache.delete(dictType);
    } else {
      this.cache.clear();
    }
  }
}

const dictCache = new DictCache();

/**
 * 字典工具类
 */
export class DictUtils {
  /**
   * 获取字典选项列表（用于表单下拉框）
   * @param dictType 字典类型
   * @param forceRefresh 是否强制刷新
   * @returns 字典选项数组
   * 
   * @example
   * const options = await dictUtils.getDictOptions('production-order-status')
   * // [{ label: '待排产', value: 'PENDING_SCHEDULE', color: '#1890ff' }, ...]
   */
  async getDictOptions(dictType: string, forceRefresh = false): Promise<DictOption[]> {
    const data = await dictCache.getDictData(dictType, forceRefresh);
    
    return data.map(item => ({
      label: item.dict_label,
      value: item.dict_value,
      color: item.css_class || undefined,
    }));
  }

  /**
   * 获取字典原始数据
   * @param dictType 字典类型
   * @param forceRefresh 是否强制刷新
   * @returns 字典数据数组
   */
  async getDictItems(dictType: string, forceRefresh = false): Promise<DictItem[]> {
    return dictCache.getDictData(dictType, forceRefresh);
  }

  /**
   * 根据字典值获取标签
   * @param dictType 字典类型
   * @param value 字典值
   * @param defaultValue 默认值（找不到时返回）
   * @returns 字典标签
   * 
   * @example
   * const label = dictUtils.getDictLabel('production-order-status', 'PENDING_SCHEDULE')
   * // '待排产'
   */
  getDictLabel(dictType: string, value: string | undefined | null, defaultValue = ''): string {
    if (!value) return defaultValue;
    
    const items = dictCache.getDictDataSync(dictType);
    const item = items.find((d: DictItem) => d.dict_value === value);
    return item?.dict_label || defaultValue;
  }

  /**
   * 根据字典值获取颜色/CSS类
   * @param dictType 字典类型
   * @param value 字典值
   * @returns 颜色/CSS类
   * 
   * @example
   * const color = dictUtils.getDictColor('production-order-status', 'PENDING_SCHEDULE')
   * // '#1890ff'
   */
  getDictColor(dictType: string, value: string | undefined | null): string | undefined {
    if (!value) return undefined;
    
    const items = dictCache.getDictDataSync(dictType);
    const item = items.find((d: DictItem) => d.dict_value === value);
    return item?.css_class || undefined;
  }

  /**
   * 根据字典值获取完整信息
   * @param dictType 字典类型
   * @param value 字典值
   * @returns 字典项或 undefined
   */
  getDictItem(dictType: string, value: string | undefined | null): DictItem | undefined {
    if (!value) return undefined;
    
    const items = dictCache.getDictDataSync(dictType);
    return items.find((d: DictItem) => d.dict_value === value);
  }

  /**
   * 预加载多个字典类型（用于页面初始化）
   * @param dictTypes 字典类型数组
   * 
   * @example
   * await dictUtils.preload(['production-order-status', 'production-type', 'priority'])
   */
  async preload(dictTypes: string[]): Promise<void> {
    await Promise.all(
      dictTypes.map(type => dictCache.getDictData(type))
    );
  }

  /**
   * 清除缓存
   * @param dictType 可选，指定字典类型，不传则清除所有
   */
  clearCache(dictType?: string): void {
    dictCache.clearCache(dictType);
  }
}

// 导出单例
export const dictUtils = new DictUtils();

// 默认导出
export default dictUtils;
