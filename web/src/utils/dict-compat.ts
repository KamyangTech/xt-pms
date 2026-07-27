/**
 * 字典值兼容转换工具
 * 用于在迁移期间支持新旧两种格式的互转
 */

import {
  MoldStatus,
  WorkOrderStatus,
  MoldOrderStatus,
  ProductionOrderStatus,
  MoldProcess,
  MaterialType,
  MoldType,
  OrderPriority,
  ProductionType,
  MoldSource,
  MoldLocation,
  MoldOrderType,
  PlasticMaterial,
  ShiftType,
} from '@/types/dict-enum';

// ============================================
// 旧数字ID到新语义化字符串的映射
// ============================================

const MOLD_STATUS_OLD_TO_NEW: Record<string, string> = {
  '1': MoldStatus.DESIGNING,
  '2': MoldStatus.PROCESSING,
  '3': MoldStatus.TESTING,
  '4': MoldStatus.PRODUCING,
  '5': MoldStatus.REPAIRING,
  '6': MoldStatus.IN_STOCK,
  '7': MoldStatus.DELIVERED,
  '9': MoldStatus.SCRAPPED,
};

const WORK_ORDER_STATUS_OLD_TO_NEW: Record<string, string> = {
  '1': WorkOrderStatus.NOT_STARTED,
  '2': WorkOrderStatus.PROCESSING,
  '3': WorkOrderStatus.COMPLETED,
  '4': WorkOrderStatus.TESTED,
  // '5': WorkOrderStatus.ACCEPTED, (已废弃 - 取消验收步骤)
  '5': WorkOrderStatus.STOCKED_IN,
  '6': WorkOrderStatus.DELIVERED,
  '7': WorkOrderStatus.DELIVERED,
};

const MOLD_ORDER_STATUS_OLD_TO_NEW: Record<string, string> = {
  '1': MoldOrderStatus.NOT_STARTED,
  '2': MoldOrderStatus.PROCESSING,
  '3': MoldOrderStatus.PENDING_TEST,
  '4': MoldOrderStatus.TESTED,
  '5': MoldOrderStatus.COMPLETED,
  '6': MoldOrderStatus.CANCELLED,
  '7': MoldOrderStatus.DELIVERED,
};

const PRODUCTION_ORDER_STATUS_OLD_TO_NEW: Record<string, string> = {
  '1': ProductionOrderStatus.PENDING_SCHEDULE,
  '2': ProductionOrderStatus.PRODUCING,
  '3': ProductionOrderStatus.STOCKED_IN,
  '4': ProductionOrderStatus.COMPLETED,
  '5': ProductionOrderStatus.CANCELLED,
};

const MOLD_PROCESS_OLD_TO_NEW: Record<string, string> = {
  '1': MoldProcess.NORMAL,
  '2': MoldProcess.HOT_RUNNER,
};

const MATERIAL_TYPE_OLD_TO_NEW: Record<string, string> = {
  '1': MaterialType.STEEL,
  '2': MaterialType.HARDENED_STEEL,
};

const MOLD_TYPE_OLD_TO_NEW: Record<string, string> = {
  '1': MoldType.NEW,
  '2': MoldType.USED,
};

const ORDER_PRIORITY_OLD_TO_NEW: Record<string, string> = {
  '1': OrderPriority.NORMAL,
  '2': OrderPriority.IMPORTANT,
  '3': OrderPriority.URGENT,
  '4': OrderPriority.VERY_URGENT,
};

const PRODUCTION_TYPE_OLD_TO_NEW: Record<string, string> = {
  '1': ProductionType.IN_HOUSE,
  '2': ProductionType.OUTSOURCED,
};

const MOLD_SOURCE_OLD_TO_NEW: Record<string, string> = {
  '1': MoldSource.SELF_MADE,
  '2': MoldSource.CUSTOMER_PROVIDED,
};

const MOLD_LOCATION_OLD_TO_NEW: Record<string, string> = {
  '1': MoldLocation.IN_FACTORY,
  '2': MoldLocation.OUTSOURCING,
  '3': MoldLocation.AT_CUSTOMER,
};

const MOLD_ORDER_TYPE_OLD_TO_NEW: Record<string, string> = {
  '1': MoldOrderType.NEW_MOLD,
  '2': MoldOrderType.REPAIR,
  '3': MoldOrderType.MODIFICATION,
};

const PLASTIC_MATERIAL_OLD_TO_NEW: Record<string, string> = {
  '1': PlasticMaterial.PP,
  '2': PlasticMaterial.PE,
  '3': PlasticMaterial.ABS,
  '4': PlasticMaterial.PS,
  '5': PlasticMaterial.PC,
  '6': PlasticMaterial.PVC,
  '7': PlasticMaterial.PMMA,
  '8': PlasticMaterial.PA,
  '9': PlasticMaterial.POM,
  '10': PlasticMaterial.PET,
  '11': PlasticMaterial.TPE,
  '12': PlasticMaterial.TPU,
};

const SHIFT_TYPE_OLD_TO_NEW: Record<string, string> = {
  '1': ShiftType.DAY_SHIFT,
  '2': ShiftType.NIGHT_SHIFT,
};

// ============================================
// ============================================

/**
 * 将旧的数字ID转换为新的语义化字符串
 * @param dictType 字典类型
 * @param oldValue 旧的值（数字字符串）
 * @returns 新的语义化字符串，如果找不到映射则返回原值
 */
export function convertOldDictValue(dictType: string, oldValue: string): string {
  const converters: Record<string, Record<string, string>> = {
    'mold-status': MOLD_STATUS_OLD_TO_NEW,
    'work-order-status': WORK_ORDER_STATUS_OLD_TO_NEW,
    'mold-order-status': MOLD_ORDER_STATUS_OLD_TO_NEW,
    'production-order-status': PRODUCTION_ORDER_STATUS_OLD_TO_NEW,
    'mold-process': MOLD_PROCESS_OLD_TO_NEW,
    'material-type': MATERIAL_TYPE_OLD_TO_NEW,
    'mold-type': MOLD_TYPE_OLD_TO_NEW,
    'order-priority': ORDER_PRIORITY_OLD_TO_NEW,
    'production-type': PRODUCTION_TYPE_OLD_TO_NEW,
    'mold-source': MOLD_SOURCE_OLD_TO_NEW,
    'mold-location': MOLD_LOCATION_OLD_TO_NEW,
    'mold_order_type': MOLD_ORDER_TYPE_OLD_TO_NEW,
    'plastic-material': PLASTIC_MATERIAL_OLD_TO_NEW,
    'shift-type': SHIFT_TYPE_OLD_TO_NEW,
  };

  const converter = converters[dictType];
  if (!converter) {
    return oldValue;
  }

  const newValue = converter[oldValue];
  if (!newValue) {
    return oldValue;
  }

  return newValue;
}

/**
 * 批量转换字典数据列表
 * @param dictType 字典类型
 * @param dataList 数据列表
 * @param valueField 值字段名，默认为 'dict_value'
 * @returns 转换后的数据列表
 */
export function convertDictDataList(
  dictType: string,
  dataList: any[],
  valueField: string = 'dict_value'
): any[] {
  return dataList.map((item) => ({
    ...item,
    [valueField]: convertOldDictValue(dictType, String(item[valueField])),
  }));
}

/**
 * 检查值是否已经是新的语义化格式
 * @param value 要检查的值
 * @returns 是否为语义化字符串
 */
export function isSemanticValue(value: string): boolean {
  // 如果不是纯数字，则认为已经是语义化字符串
  return !/^\d+$/.test(value);
}

/**
 * 安全地获取字典值（兼容新旧格式）
 * @param dictType 字典类型
 * @param value 字典值（可能是数字或语义化字符串）
 * @returns 语义化字符串
 */
export function safeGetDictValue(dictType: string, value: string | number): string {
  const strValue = String(value);
  
  // 如果已经是语义化字符串，直接返回
  if (isSemanticValue(strValue)) {
    return strValue;
  }
  
  // 否则进行转换
  return convertOldDictValue(dictType, strValue);
}
