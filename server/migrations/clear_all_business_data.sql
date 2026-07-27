-- ============================================
-- 清空所有业务数据 SQL 脚本
-- 用途：清空项目中的所有业务数据，保留系统设置和基础配置数据
-- 执行顺序：按照外键依赖关系的逆序删除（先删子表，再删主表）
-- 注意：执行前请确认已备份需要保留的数据
-- ============================================

-- 禁用外键检查（避免外键约束导致的删除失败）
SET FOREIGN_KEY_CHECKS = 0;

-- ============================================
-- 一、删除模具相关数据
-- ============================================

-- 1. 模具试模记录（mold_try_test）
DELETE FROM mold_try_test;

-- 2. 模具交付记录（mold_delivery）
DELETE FROM mold_delivery;

-- 3. 模具库存记录（mold_stock）
DELETE FROM mold_stock;

-- 4. 模具使用记录（mold_usage_record）
DELETE FROM mold_usage_record;

-- 5. 模具 BOM（mold_bom）
DELETE FROM mold_bom;

-- 6. 模具订单款项（mold_order_payment）
DELETE FROM mold_order_payment;

-- 7. 工单工序关联（work_order_process）
DELETE FROM work_order_process;

-- 8. 模具订单明细（mold_order_item）
DELETE FROM mold_order_item;

-- 9. 模具工单（mold_work_order）
DELETE FROM mold_work_order;

-- 10. 模具订单（mold_order）
DELETE FROM mold_order;

-- 11. 模具产品关联（mold_product_relation）
DELETE FROM mold_product_relation;

-- 12. 模具档案（mold）
DELETE FROM mold;

-- ============================================
-- 二、删除产品相关数据
-- ============================================

-- 13. 产品 BOM（product_bom）
DELETE FROM product_bom;

-- 14. 产品表（product）
DELETE FROM product;

-- ============================================
-- 三、删除生产管理相关数据
-- ============================================

-- 15. 生产订单明细（production_order_item）
DELETE FROM production_order_item;

-- 16. 生产排程（production_schedule）
DELETE FROM production_schedule;

-- 17. 生产报工（production_report）
DELETE FROM production_report;

-- 18. 外发报工（outsourcing_report）
DELETE FROM outsourcing_report;

-- 19. 外发记录（outsourcing_record）
DELETE FROM outsourcing_record;

-- 20. 生产出库（production_stock_out）
DELETE FROM production_stock_out;

-- 21. 生产入库（production_stock_in）
DELETE FROM production_stock_in;

-- 22. 生产订单（production_order）
DELETE FROM production_order;

-- ============================================
-- 四、删除库存相关数据
-- ============================================

-- 23. 库存表（stock）
DELETE FROM stock;

-- ============================================
-- 五、删除综合订单相关数据
-- ============================================

-- 24. 综合订单（master_orders）
DELETE FROM master_orders;

-- ============================================
-- 六、删除外来模具档案
-- ============================================

-- 25. 外来模具档案（external_mold_archive）
DELETE FROM external_mold_archive;

-- ============================================
-- 七、删除财务管理相关数据
-- ============================================

-- 26. 费用支付记录（expense_payments）
DELETE FROM expense_payments;

-- 27. 收款记录（receivable_payments）
DELETE FROM receivable_payments;

-- 28. 付款记录（payable_payments）
DELETE FROM payable_payments;

-- 29. 财务费用（finance_expense）
DELETE FROM finance_expense;

-- 30. 应收账款（accounts_receivable）
DELETE FROM accounts_receivable;

-- 31. 应付账款（accounts_payable）
DELETE FROM accounts_payable;

-- 32. 收款账户（receiving_accounts）
DELETE FROM receiving_accounts;

-- ============================================
-- 八、删除系统通知和日志
-- ============================================

-- 33. 系统通知（sys_notification）
DELETE FROM sys_notification;

-- 34. 通知配置（sys_notification_config）
DELETE FROM sys_notification_config;

-- 35. 操作日志（sys_oper_log）
DELETE FROM sys_oper_log;

-- ============================================
-- 九、删除备份表业务数据
-- ============================================

-- 36. 备份模具验收记录（_backup_mold_acceptance）
DELETE FROM _backup_mold_acceptance;

-- 37. 备份质量检验记录（_backup_quality_inspection）
DELETE FROM _backup_quality_inspection;

-- ============================================
-- 十、删除所有业务文件附件
-- ============================================

-- 38. 文件附件表（files）— 清空所有业务关联的文件记录
DELETE FROM files;

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 清空完成
-- ============================================
-- 保留的数据表（系统设置 + 基础数据 + 配置）：
-- --- 系统管理 ---
-- - sys_user（系统用户）
-- - sys_role（角色）
-- - sys_user_role（用户角色关联）
-- - sys_role_menu（角色菜单关联）
-- - sys_menu（菜单权限）
-- - sys_dept（部门）
-- - sys_company（公司信息）
-- - sys_system_info（系统信息配置）
-- - sys_security_config（安全配置）
-- --- 字典数据 ---
-- - sys_dict_type（字典类型）
-- - sys_dict_data（字典数据）
-- --- 基础资料 ---
-- - customer（客户）
-- - material（物料）
-- - warehouse（仓库）
-- - supplier（供应商）
-- - processor（加工商）
-- - process（工序模板）
-- - process_template（工序模板分组）
-- --- 配置数据 ---
-- - status_flow_config（状态同步配置）
-- - production_status_flow（生产订单状态流转规则）
-- - print_notice_config（打印通知单配置）
-- ============================================
