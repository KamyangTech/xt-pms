-- ============================================
-- 数据种子: 模具状态字典 & 状态流转配置
-- 导出时间: 2026-05-14
-- 说明：恢复模具状态、工单状态、订单状态的字典数据
-- 以及状态流转配置和生产订单状态流转规则
-- ============================================

-- 1. 先清理原有数据（按依赖顺序删除）
DELETE FROM production_status_flow;
DELETE FROM status_flow_config;
DELETE FROM sys_dict_data
WHERE dict_type_id IN (
  SELECT id FROM sys_dict_type
  WHERE dict_code IN ('mold-status', 'work-order-status', 'mold-order-status')
);

-- ============================================
-- 2. 字典类型（若无变化则跳过）
-- ============================================
INSERT INTO sys_dict_type (id, dict_name, dict_code, remark)
VALUES
  (1, '模具状态管理', 'mold-status', '模具的各种状态'),
  (7, '工单状态', 'work-order-status', '模具工单的状态'),
  (8, '订单状态', 'mold-order-status', '模具订单的状态')
ON DUPLICATE KEY UPDATE
  dict_name = VALUES(dict_name),
  dict_code = VALUES(dict_code),
  remark = VALUES(remark);

-- ============================================
-- 3. 模具档案状态（mold-status）
-- ============================================
INSERT INTO sys_dict_data (id, dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark, ext) VALUES
(27, 1, '未开始',   'DESIGNING',                'gray-7',   1,  1, 2, 0, 1, 0, '模具设计阶段', NULL),
(28, 1, '加工中',   'PROCESSING',               'orange-5', 2,  1, 2, 0, 1, 0, '模具正在加工', NULL),
(29, 1, '待试模',   'TESTING',                  'primary',  3,  1, 0, 0, 0, 0, '工序完成，等待试模', NULL),
(30, 1, '已试模',   'TESTED',                   'cyan-7',   4,  1, 1, 1, 1, 0, '试模进行中', NULL),
(32, 1, '仓库存放', 'IN_STOCK',                 'gray-11',  6,  1, 1, 1, 1, 0, '在仓库，未生产状态', NULL),
(33, 1, '已交付',   'DELIVERED',                'blue-8',   7,  1, 1, 0, 1, 0, '交付给客户', NULL),
(34, 1, '维修',     'REPAIRING',                'warning',  8,  1, 2, 0, 1, 0, '维修状态', NULL),
(35, 1, '报废',     'SCRAPPED',                 'danger',   9,  1, 2, 0, 1, 0, '模具报废', NULL),
(53, 1, '制造中断', 'MANUFACTURING_INTERRUPTED', 'danger',   10, 1, 1, 0, 1, 0, '订单取消，流程终止', NULL),
(160, 1, '生产中',  'PRODUCING',                'success',  10, 1, 0, 0, 1, 0, '模具投入生产中', NULL)
ON DUPLICATE KEY UPDATE
  dict_label = VALUES(dict_label),
  dict_value = VALUES(dict_value),
  color = VALUES(color),
  sort = VALUES(sort),
  status = VALUES(status),
  is_locked = VALUES(is_locked),
  allow_production = VALUES(allow_production),
  allow_work_report = VALUES(allow_work_report),
  needs_delivery_unlock = VALUES(needs_delivery_unlock),
  remark = VALUES(remark),
  ext = VALUES(ext);

-- ============================================
-- 4. 模具工单状态（work-order-status）
-- ============================================
INSERT INTO sys_dict_data (id, dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark, ext) VALUES
(36, 7, '未开始', 'NOT_STARTED', 'gray-7',  1, 1, 2, 0, 1, 0, '工单未开始', NULL),
(37, 7, '加工中', 'PROCESSING',  'warning', 2, 1, 2, 0, 1, 0, '工单加工中', NULL),
(38, 7, '待试模', 'COMPLETED',   'primary', 3, 1, 0, 0, 0, 0, '所有工序完成，等待试模', NULL),
(39, 7, '已试模', 'TESTED',      'cyan-7',  4, 1, 1, 0, 1, 0, '已试模', NULL),
(41, 7, '已入库', 'STOCKED_IN',  'gray-11', 6, 1, 1, 0, 1, 0, '已入库', NULL),
(42, 7, '已交付', 'DELIVERED',   'blue-8',  7, 1, 1, 0, 1, 0, '已交付客户', NULL),
(52, 7, '已取消', 'CANCELLED',   'danger',  8, 1, 2, 0, 1, 0, '工单取消', NULL)
ON DUPLICATE KEY UPDATE
  dict_label = VALUES(dict_label),
  dict_value = VALUES(dict_value),
  color = VALUES(color),
  sort = VALUES(sort),
  status = VALUES(status),
  is_locked = VALUES(is_locked),
  allow_production = VALUES(allow_production),
  allow_work_report = VALUES(allow_work_report),
  needs_delivery_unlock = VALUES(needs_delivery_unlock),
  remark = VALUES(remark),
  ext = VALUES(ext);

-- ============================================
-- 5. 模具订单状态（mold-order-status）
-- ============================================
INSERT INTO sys_dict_data (id, dict_type_id, dict_label, dict_value, color, sort, status, is_locked, allow_production, allow_work_report, needs_delivery_unlock, remark, ext) VALUES
(43, 8, '未开始', 'NOT_STARTED',  'gray-7',  1, 1, 2, 0, 1, 0, '订单未开始', NULL),
(44, 8, '加工中', 'PROCESSING',   'warning', 2, 1, 2, 0, 1, 0, '订单加工中', NULL),
(45, 8, '待试模', 'PENDING_TEST', 'primary', 3, 1, 0, 0, 0, 0, '待试模', NULL),
(46, 8, '已试模', 'TESTED',       'cyan-7',  4, 1, 1, 0, 1, 0, '已试模', NULL),
(47, 8, '已完成', 'COMPLETED',    'success', 5, 1, 2, 0, 1, 0, '已完成', NULL),
(48, 8, '已取消', 'CANCELLED',    'danger',  6, 1, 2, 0, 1, 0, '订单取消，终止', NULL),
(49, 8, '已交付', 'DELIVERED',    'blue-8',  7, 1, 1, 0, 1, 0, '已交付客户', NULL)
ON DUPLICATE KEY UPDATE
  dict_label = VALUES(dict_label),
  dict_value = VALUES(dict_value),
  color = VALUES(color),
  sort = VALUES(sort),
  status = VALUES(status),
  is_locked = VALUES(is_locked),
  allow_production = VALUES(allow_production),
  allow_work_report = VALUES(allow_work_report),
  needs_delivery_unlock = VALUES(needs_delivery_unlock),
  remark = VALUES(remark),
  ext = VALUES(ext);

-- ============================================
-- 6. 状态流转配置（status_flow_config）
-- ============================================
DELETE FROM status_flow_config;
INSERT INTO status_flow_config (id, work_order_status, mold_order_status, mold_status, description, priority, is_active, status_order) VALUES
(11, 'NOT_STARTED', 'NOT_STARTED', 'DESIGNING',                '未开始状态', 0, 1, 0),
(12, 'PROCESSING',  'PROCESSING',  'PROCESSING',               '加工中',    0, 1, 0),
(13, 'COMPLETED',   'PENDING_TEST', 'TESTING',                 '待试模',    0, 1, 0),
(14, 'TESTED',      'TESTED',      'TESTED',                   '已试模',    0, 1, 0),
(15, 'STOCKED_IN',  'COMPLETED',   'IN_STOCK',                 '仓库存放',  0, 1, 0),
(16, 'DELIVERED',   'DELIVERED',   'DELIVERED',                '交付客户',  0, 1, 0),
(17, 'CANCELLED',   'CANCELLED',   'MANUFACTURING_INTERRUPTED', '取消制造',  0, 1, 0)
ON DUPLICATE KEY UPDATE
  work_order_status = VALUES(work_order_status),
  mold_order_status = VALUES(mold_order_status),
  mold_status = VALUES(mold_status),
  description = VALUES(description),
  priority = VALUES(priority),
  is_active = VALUES(is_active),
  status_order = VALUES(status_order);

-- ============================================
-- 7. 生产订单状态流转规则（production_status_flow）
-- ============================================
DELETE FROM production_status_flow;
INSERT INTO production_status_flow (id, from_status_value, to_status_value, action_name, action_code, is_enabled, sort, remark) VALUES
(1,  'PENDING_SCHEDULE', 'PRODUCING',  '开始生产', 'start_production', 1, 1,  '待排产/已排产 → 生产中'),
(6,  'STOCKED_IN',       'SHIPPED',    '出货',     'shipment',         1, 6,  '已入库 → 已出货'),
(7,  'SHIPPED',          'COMPLETED',  '完成订单', 'finish_order',     1, 7,  '已出货 → 已完成'),
(8,  'PENDING_SCHEDULE', 'CANCELLED',  '取消订单', 'cancel_order',     1, 8,  '待排产 → 已取消'),
(10, 'PRODUCING',        'CANCELLED',  '取消订单', 'cancel_order',     1, 10, '生产中 → 已取消'),
(11, 'PRODUCING',        'STOCKED_IN', '结束生产', 'complete_production', 1, 5, '生产中 → 已入库（报工完成后标记明细项入库）'),
(14, 'STOCKED_IN',       'COMPLETED',  '完成订单', 'finish_order',     1, 0,  NULL),
(17, 'STOCKED_IN',       'STOCKED_IN', '继续入库', 'stock_in',         1, 0,  NULL)
ON DUPLICATE KEY UPDATE
  from_status_value = VALUES(from_status_value),
  to_status_value = VALUES(to_status_value),
  action_name = VALUES(action_name),
  action_code = VALUES(action_code),
  is_enabled = VALUES(is_enabled),
  sort = VALUES(sort),
  remark = VALUES(remark);
