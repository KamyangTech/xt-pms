import { request } from '@/utils/request';

// 生产通知单打印配置
export interface PrintNoticeConfig {
  basic_info: {
    order_code: boolean;
    customer_name: boolean;
    production_type: boolean;
    mold_source: boolean;
  };
  production_info: {
    plan_qty: boolean;
    worker: boolean;
    plan_start_date: boolean;
    plan_end_date: boolean;
    completed_qty: boolean;
  };
  detail_columns: {
    product_name: boolean;
    mold_code: boolean;
    product_image: boolean;
    plan_qty: boolean;
    completed_qty: boolean;
  };
}

// 送货单打印配置
export interface DeliveryNoteConfig {
  basic_info: {
    stock_code: boolean;
    customer_name: boolean;
    mold_code: boolean;
    product_name: boolean;
  };
  delivery_info: {
    delivery_date: boolean;
    out_date: boolean;
    delivery_method: boolean;
    quantity: boolean;
    express_company: boolean;
    express_no: boolean;
    driver_phone: boolean;
    driver_plate: boolean;
  };
  detail_columns: {
    product_image: boolean;
    mold_code: boolean;
    product_name: boolean;
    quantity: boolean;
    batch_no: boolean;
    remark: boolean;
  };
}

/**
 * 获取打印配置
 * @param configType 配置类型：production_order | delivery_note
 */
export function getPrintNoticeConfig(configType: string = 'production_order') {
  return request.get({
    url: '/system/print-notice-config',
    params: { config_type: configType },
  });
}

/**
 * 保存打印配置
 * @param configType 配置类型：production_order | delivery_note
 */
export function savePrintNoticeConfig(
  config: Record<string, unknown>,
  configType: string = 'production_order'
) {
  return request.post({
    url: '/system/print-notice-config',
    data: config,
    params: { config_type: configType },
  });
}

/**
 * 重置打印配置为默认值
 * @param configType 配置类型：production_order | delivery_note
 */
export function resetPrintNoticeConfig(configType: string = 'production_order') {
  return request.post({
    url: '/system/print-notice-config/reset',
    params: { config_type: configType },
  });
}
