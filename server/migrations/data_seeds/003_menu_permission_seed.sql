-- ============================================
-- 系统菜单与权限种子数据
-- 生成依据：web/src/router/modules/*.ts 路由配置
-- 用法：复制全文一次性执行即可
-- ============================================

-- 清空旧数据（外键约束：先删关联表，再删主表）
DELETE FROM sys_role_menu;
DELETE FROM sys_menu;

-- ============================================
-- 1. 一级目录 (menu_type=1)
--    注意：permission 为 NULL 表示公开路由，不参与权限校验
--          空字符串 '' 会违反 UNIQUE 约束，所以统一用 NULL
-- ============================================
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, icon, sort, status) VALUES
(1, 0, '仪表盘', 1, NULL, '/dashboard', 'dashboard', 0, 1),
(2, 0, '订单中心', 1, NULL, '/order-management', 'view-list', 1, 1),
(3, 0, '模具管理', 1, NULL, '/mold', 'chart-3d', 2, 1),
(4, 0, '生产管理', 1, NULL, '/production', 'tools', 3, 1),
(5, 0, '产品管理', 1, NULL, '/product', 'layers', 4, 1),
(6, 0, '外发管理', 1, NULL, '/outsourcing', 'send', 5, 1),
(7, 0, '财务管理', 1, NULL, '/finance', 'wallet', 6, 1),
(8, 0, '系统设置', 1, NULL, '/system', 'setting', 7, 1),
(9, 0, '系统字典', 1, NULL, '/dict', 'list', 8, 1);

-- ============================================
-- 2. 二级菜单 (menu_type=2)
-- ============================================

-- 仪表盘 (parent_id=1)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(10, 1, '概览仪表盘', 2, 'dashboard:base', '/dashboard/base', 'dashboard/base/index', 0, 1),
(11, 1, '统计报表', 2, 'dashboard:detail', '/dashboard/detail', 'dashboard/detail/index', 1, 1);

-- 订单中心 (parent_id=2)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(12, 2, '订单中心', 2, 'order-management:list', '/order-management/list', 'order-management/index', 0, 1);

-- 模具管理 (parent_id=3)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(13, 3, '模具档案', 2, 'mold:mold:list', '/mold/index', 'mold/index', 0, 1),
(14, 3, '模具订单', 2, 'mold:order:list', '/mold/order', 'mold/order/index', 1, 1),
(15, 3, '模具工单', 2, 'mold:work-order:list', '/mold/work-order', 'mold/work-order/index', 2, 1),
(16, 3, '使用历史', 2, 'mold:usage-history:list', '/mold/usage-history', 'mold/usage-history/index', 3, 1);

-- 生产管理 (parent_id=4)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(17, 4, '生产订单', 2, 'production:order:list', '/production/order/list', 'production/order-list/index', 0, 1),
(18, 4, '生产报工', 2, 'production:work-report:list', '/production/work-report', 'production/work-report/index', 1, 1),
(19, 4, '库存管理', 2, 'production:stock:list', '/production/stock', 'production/stock/index', 2, 1),
(20, 4, '发货管理', 2, 'production:stock-out:list', '/production/stock-out', 'production/stock-out/index', 3, 1);

-- 产品管理 (parent_id=5)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(21, 5, '产品列表', 2, 'product:list', '/product/index', 'basic/product/index', 0, 1);

-- 外发管理 (parent_id=6)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(22, 6, '模具外发', 2, 'outsourcing:mold:list', '/outsourcing/mold', 'outsourcing/mold/index', 0, 1),
(23, 6, '生产外发', 2, 'outsourcing:production:list', '/outsourcing/production', 'outsourcing/production/index', 1, 1);

-- 财务管理 (parent_id=7)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(24, 7, '应收账款', 2, 'finance:receivable:list', '/finance/receivable', 'finance/receivable/index', 0, 1),
(25, 7, '应付账款', 2, 'finance:payable:list', '/finance/payable', 'finance/payable/index', 1, 1),
(26, 7, '费用管理', 2, 'finance:expense:list', '/finance/expense', 'finance/expense/index', 2, 1),
(27, 7, '收款账户管理', 2, 'system:receiving-account:list', '/finance/receiving-account', 'system/receiving-account/index', 3, 1);

-- 系统设置 (parent_id=8)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(28, 8, '组织与权限', 2, 'system:organization:list', '/system/organization', 'system/organization/index', 0, 1),
(29, 8, '菜单管理', 2, 'system:menu:list', '/system/menu', 'system/menu/index', 1, 1),
(30, 8, '基础资料', 2, 'system:basic-data:list', '/system/basic-data', 'system/basic-data/index', 2, 1),
(31, 8, '打印配置', 2, 'system:print-config:list', '/system/print-config', 'system/print-config/index', 3, 1),
(32, 8, '系统设置', 2, 'system:settings:list', '/system/settings', 'system/settings/index', 4, 1),
(33, 8, '操作日志', 2, 'system:oper-log:list', '/system/oper-log', 'system/oper-log/index', 5, 1),
(34, 8, '业务配置', 2, 'system:business-config:list', '/system/business-config', 'system/business-config/index', 6, 1),
(49, 8, '系统监控', 2, 'system:monitor:view', '/system/monitor', 'system/monitor/index', 7, 1);

-- 系统字典 (parent_id=9)
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, path, component, sort, status) VALUES
(35, 9, '模具状态字典', 2, 'dict:mold-status:list', '/dict/mold-status', 'system/status-dict/index', 0, 1),
(36, 9, '生产状态字典', 2, 'dict:production-status:list', '/dict/production-status-dict', 'system/production-status-dict/index', 1, 1),
(37, 9, '综合订单字典', 2, 'dict:master-order:list', '/dict/master-order-dict', 'system/master-order-dict/index', 2, 1),
(38, 9, '外发状态字典', 2, 'dict:outsourcing-status:list', '/dict/outsourcing-status', 'system/outsourcing-status-dict/index', 3, 1);

-- ============================================
-- 3. 按钮级权限 (menu_type=3)
-- 挂载在对应的父菜单下，不在侧边栏显示
-- ============================================

-- 组织与权限 (parent_id=28) 下的按钮权限
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, sort, status) VALUES
(39, 28, '创建用户', 3, 'system:user:create', 0, 1),
(40, 28, '编辑用户', 3, 'system:user:update', 1, 1),
(41, 28, '删除用户', 3, 'system:user:delete', 2, 1),
(42, 28, '创建角色', 3, 'system:role:create', 3, 1),
(43, 28, '编辑角色', 3, 'system:role:update', 4, 1),
(44, 28, '分配权限', 3, 'system:role:assign-perm', 5, 1),
(45, 28, '删除角色', 3, 'system:role:delete', 6, 1);

-- 菜单管理 (parent_id=29) 下的按钮权限
INSERT INTO sys_menu (id, parent_id, menu_name, menu_type, permission, sort, status) VALUES
(46, 29, '创建菜单', 3, 'system:menu:create', 0, 1),
(47, 29, '编辑菜单', 3, 'system:menu:update', 1, 1),
(48, 29, '删除菜单', 3, 'system:menu:delete', 2, 1);
