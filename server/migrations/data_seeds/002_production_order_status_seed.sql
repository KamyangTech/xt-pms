-- ============================================
-- 数据种子: 生产订单状态字典 & 流转规则
-- 导出时间: 2026-05-14
-- 说明：恢复生产订单状态的字典数据及流转规则
-- ============================================

-- 1. 先清理原有数据
DELETE FROM production_status_flow
WHERE from_status_value IN (
  SELECT dict_value FROM sys_dict_data
  WHERE dict_type_id = (SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status')
);

DELETE FROM sys_dict_data
WHERE dict_type_id = (SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status');

-- ============================================
-- 2. 生产订单状态字典数据
-- ============================================
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark)
SELECT dt.id, '待排产', 'PENDING_SCHEDULE', 'gray-7',   1, 1, 0, 0, 1, 0, '订单已创建，等待排产'
FROM sys_dict_type dt WHERE dt.dict_code = 'production-order-status'
AND NOT EXISTS (SELECT 1 FROM sys_dict_data dd WHERE dd.dict_type_id = dt.id AND dd.dict_value = 'PENDING_SCHEDULE');

INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark)
SELECT dt.id, '生产中', 'PRODUCING', 'orange-5', 2, 1, 0, 0, 1, 0, '正在生产中'
FROM sys_dict_type dt WHERE dt.dict_code = 'production-order-status'
AND NOT EXISTS (SELECT 1 FROM sys_dict_data dd WHERE dd.dict_type_id = dt.id AND dd.dict_value = 'PRODUCING');

INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark)
SELECT dt.id, '已入库', 'STOCKED_IN', 'gray-11',  3, 1, 0, 0, 0, 0, '产品已入库'
FROM sys_dict_type dt WHERE dt.dict_code = 'production-order-status'
AND NOT EXISTS (SELECT 1 FROM sys_dict_data dd WHERE dd.dict_type_id = dt.id AND dd.dict_value = 'STOCKED_IN');

INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark)
SELECT dt.id, '已出货', 'SHIPPED', 'success', 4, 1, 0, 0, 0, 0, '已经发货给客户'
FROM sys_dict_type dt WHERE dt.dict_code = 'production-order-status'
AND NOT EXISTS (SELECT 1 FROM sys_dict_data dd WHERE dd.dict_type_id = dt.id AND dd.dict_value = 'SHIPPED');

INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark)
SELECT dt.id, '已完成', 'COMPLETED', 'success', 5, 1, 0, 0, 0, 0, '订单全部完成'
FROM sys_dict_type dt WHERE dt.dict_code = 'production-order-status'
AND NOT EXISTS (SELECT 1 FROM sys_dict_data dd WHERE dd.dict_type_id = dt.id AND dd.dict_value = 'COMPLETED');

INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark)
SELECT dt.id, '已取消', 'CANCELLED', 'danger', 6, 1, 0, 0, 0, 0, '订单已取消'
FROM sys_dict_type dt WHERE dt.dict_code = 'production-order-status'
AND NOT EXISTS (SELECT 1 FROM sys_dict_data dd WHERE dd.dict_type_id = dt.id AND dd.dict_value = 'CANCELLED');

-- ============================================
-- 3. 生产订单状态流转规则
-- ============================================
DELETE FROM production_status_flow;

INSERT INTO production_status_flow (from_status_value, to_status_value, action_name, action_code, is_enabled, sort, remark) VALUES
('PENDING_SCHEDULE', 'PRODUCING', '开始生产', 'start_production', 1, 1, '待排产 → 生产中'),
('STOCKED_IN',       'SHIPPED',   '出货',     'shipment',         1, 6, '已入库 → 已出货'),
('SHIPPED',          'COMPLETED', '完成订单', 'finish_order',     1, 7, '已出货 → 已完成'),
('PENDING_SCHEDULE', 'CANCELLED', '取消订单', 'cancel_order',     1, 8, '待排产 → 已取消'),
('PRODUCING',        'CANCELLED', '取消订单', 'cancel_order',     1, 10, '生产中 → 已取消'),
('PRODUCING',        'STOCKED_IN', '结束生产', 'complete_production', 1, 5, '生产中 → 已入库'),
('STOCKED_IN',       'COMPLETED', '完成订单', 'finish_order',     1, 0,  NULL),
('STOCKED_IN',       'STOCKED_IN', '继续入库', 'stock_in',         1, 0,  NULL);
