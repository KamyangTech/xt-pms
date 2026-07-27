/**
 * 字典 composable - 基于 dictUtils 封装的 Vue 响应式 Hook
 *
 * 使用示例：
 * ```typescript
 * // 单个字典（返回 {label, value} 格式，用于下拉框）
 * const { options: statusOptions } = useDict('production-order-status')
 *
 * // 单个字典（返回原始 DictItem，需要 ext/color 等额外字段时使用）
 * const { items: methodItems } = useDictRaw('payment-method')
 *
 * // 多个字典（批量预加载）
 * const {
 *   'production-order-status': statusOptions,
 *   'production-type': typeOptions,
 * } = useDicts(['production-order-status', 'production-type'])
 * ```
 */

import type { Ref } from 'vue';
import { onMounted, ref } from 'vue';
import { dictUtils } from '@/utils/dict';
import type { DictItem, DictOption } from '@/utils/dict';

/**
 * 获取单个字典的响应式选项（{ label, value } 格式）
 */
export function useDict(dictType: string): { options: Ref<DictOption[]> } {
  const options = ref<DictOption[]>([]);

  const load = async () => {
    try {
      options.value = await dictUtils.getDictOptions(dictType);
    } catch {
      options.value = [];
    }
  };

  onMounted(load);
  return { options };
}

/**
 * 获取单个字典的响应式原始数据（支持泛型，默认返回 DictItem 完整字段）
 *
 * @template T 输出类型，默认为 DictItem
 * @param dictType 字典类型
 * @param transform 可选转换函数，用于对原始数据进行过滤/映射/排序（可改变输出类型）
 * @example
 * // 默认：返回 DictItem[]（包含 dict_value/dict_label 等原始字段）
 * const { items: methodItems } = useDictRaw('payment-method')
 *
 * // 自定义类型：转换为扁平化的 { value, label, color } 格式
 * const { items: statusList } = useDictRaw('work-order-status', (items) =>
 *   items.filter(i => i.status === 1).map(i => ({
 *     value: i.dict_value,
 *     label: i.dict_label,
 *     color: i.css_class || 'default',
 *   }))
 * )
 */
export function useDictRaw(
  dictType: string,
  transform?: (items: DictItem[]) => any[],
): { items: Ref<any[]> } {
  const items = ref<any[]>([]);

  const load = async () => {
    try {
      let data = await dictUtils.getDictItems(dictType);
      if (transform) {
        data = transform(data);
      }
      items.value = data;
    } catch {
      items.value = [];
    }
  };

  onMounted(load);
  return { items };
}

/**
 * 获取多个字典的响应式选项（批量预加载，自动去重请求）
 * @example
 * const refs = useDicts(['production-order-status', 'production-type'])
 * refs['production-order-status'].value  // => DictOption[]
 * refs['production-type'].value         // => DictOption[]
 */
export function useDicts(dictTypes: string[]): Record<string, Ref<DictOption[]>> {
  const refs: Record<string, Ref<DictOption[]>> = {};
  dictTypes.forEach((type) => {
    refs[type] = ref<DictOption[]>([]);
  });

  const loadAll = async () => {
    await dictUtils.preload(dictTypes);
    for (const type of dictTypes) {
      try {
        refs[type].value = await dictUtils.getDictOptions(type);
      } catch {
        refs[type].value = [];
      }
    }
  };

  onMounted(loadAll);
  return refs;
}
