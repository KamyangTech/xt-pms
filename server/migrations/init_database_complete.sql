-- ============================================
-- 产品生产管理系统 - 完整数据库初始化脚本
-- 版本：2026-05-16（以运行中的 production_pms_bk 为基准对齐结构）
-- 数据库名：production_pms
-- 字符集：utf8mb4
-- 排序规则：utf8mb4_unicode_ci
-- ============================================

-- 禁用外键检查（允许删除有外键约束的表）
SET FOREIGN_KEY_CHECKS = 0;

-- 创建数据库
DROP DATABASE IF EXISTS production_pms;
CREATE DATABASE production_pms DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE production_pms;

-- 恢复外键检查（表创建完成后）
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- 一、系统基础表（6 张）
-- ============================================

-- 1.1 系统用户表（sys_user）
DROP TABLE IF EXISTS sys_user;
CREATE TABLE sys_user (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  username varchar(50) NOT NULL COMMENT '登录账号',
  password varchar(100) NOT NULL COMMENT '密码（加密存储）',
  real_name varchar(50) DEFAULT NULL COMMENT '真实姓名',
  phone varchar(20) DEFAULT NULL COMMENT '手机号',
  email varchar(100) DEFAULT NULL COMMENT '邮箱',
  position varchar(50) DEFAULT NULL COMMENT '职位',
  hire_date DATE DEFAULT NULL COMMENT '入职日期',
  dept_id bigint DEFAULT NULL COMMENT '关联部门 ID',
  status tinyint NOT NULL DEFAULT '1' COMMENT '用户状态：0=禁用, 1=正常',
  login_attempts int NOT NULL DEFAULT '0' COMMENT '连续登录失败次数',
  locked_until datetime DEFAULT NULL COMMENT '账号锁定截止时间，NULL表示未锁定',
  password_version int NOT NULL DEFAULT '1' COMMENT '密码版本号，修改密码时自增，旧 Token 自动失效',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username) COMMENT '用户名唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统用户表';

-- 1.2 角色表（sys_role）
DROP TABLE IF EXISTS sys_role;
CREATE TABLE sys_role (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '角色 ID',
  role_name varchar(50) NOT NULL COMMENT '角色名称',
  remark varchar(255) DEFAULT NULL COMMENT '角色备注',
  role_type tinyint DEFAULT 0 COMMENT '0=自定义角色 1=系统内置',
  data_scope tinyint DEFAULT 1 COMMENT '1=全部数据 2=本部门及以下 3=本部门 4=仅本人',
  status tinyint DEFAULT 1 COMMENT '0=禁用 1=启用',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_role_name (role_name) COMMENT '角色名唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 1.3 用户角色关联表（sys_user_role）
DROP TABLE IF EXISTS sys_user_role;
CREATE TABLE sys_user_role (
  user_id bigint NOT NULL COMMENT '用户 ID',
  role_id bigint NOT NULL COMMENT '角色 ID',
  PRIMARY KEY (user_id, role_id) COMMENT '联合主键',
  CONSTRAINT fk_user_role_user FOREIGN KEY (user_id) REFERENCES sys_user(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_role_role FOREIGN KEY (role_id) REFERENCES sys_role(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- 1.4 部门表（sys_dept）
DROP TABLE IF EXISTS sys_dept;
CREATE TABLE sys_dept (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '部门 ID',
  dept_name varchar(50) NOT NULL COMMENT '部门名称',
  parent_id bigint DEFAULT 0 COMMENT '上级部门 ID，0 表示顶级部门',
  sort int DEFAULT 0 COMMENT '部门排序',
  leader varchar(50) DEFAULT NULL COMMENT '部门负责人',
  phone varchar(20) DEFAULT NULL COMMENT '联系电话',
  status tinyint DEFAULT '1' COMMENT '部门状态：0=停用，1=正常',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';

-- 1.5 公司信息表（sys_company）
DROP TABLE IF EXISTS sys_company;
CREATE TABLE sys_company (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  company_name varchar(100) NOT NULL COMMENT '公司全称',
  short_name varchar(50) DEFAULT NULL COMMENT '公司简称',
  english_name varchar(200) DEFAULT NULL COMMENT '英文全称',
  english_short_name varchar(100) DEFAULT NULL COMMENT '英文简称',
  logo varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '公司 logo URL（建议尺寸 336*44px）',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公司信息表';

-- 1.6 安全配置表（sys_security_config）
DROP TABLE IF EXISTS sys_security_config;
CREATE TABLE sys_security_config (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `token_expire` int NOT NULL DEFAULT '7200' COMMENT 'Token有效期(秒)',
  `max_login_attempts` int NOT NULL DEFAULT '5' COMMENT '最大登录失败次数',
  `lock_duration` int NOT NULL DEFAULT '900' COMMENT '账号锁定时长(秒),0表示需管理员解锁',
  `password_min_len` int NOT NULL DEFAULT '6' COMMENT '密码最小长度',
  `password_require_upper` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否要求大写字母',
  `password_require_number` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否要求数字',
  `password_require_special` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否要求特殊字符',
  `cookie_secure` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'Cookie Secure标志：1=仅HTTPS发送，0=HTTP/HTTPS均可',
  `websocket_secure` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'WebSocket Secure标志：1=仅WSS连接，0=WS/WSS均可',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='安全配置表';

-- 1.7 文件表（files）
DROP TABLE IF EXISTS files;
CREATE TABLE files (
  `id` varchar(36) NOT NULL COMMENT 'UUID',
  `url` varchar(500) NOT NULL COMMENT '存储访问路径（如 /uploads/ab/cd/uuid_filename.pdf）',
  `file_name` varchar(255) NOT NULL COMMENT '原始文件名',
  `file_size` bigint NOT NULL DEFAULT '0' COMMENT '文件大小（字节）',
  `file_type` varchar(20) NOT NULL DEFAULT '' COMMENT '文件扩展名（如 .pdf/.jpg）',
  `ref_count` int NOT NULL DEFAULT '0' COMMENT '引用计数（被业务引用的次数）',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间（ref_count=0 时设置）',
  `business_type` varchar(50) DEFAULT NULL COMMENT '业务类型（expense/receivable/mold/product 等）',
  `business_id` bigint DEFAULT NULL COMMENT '业务记录 ID',
  `is_primary` tinyint NOT NULL DEFAULT '0' COMMENT '是否主图(0否 1是)',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序（数字越小越靠前）',
  `created_by` bigint DEFAULT NULL COMMENT '上传用户 ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_business` (`business_type`,`business_id`),
  KEY `idx_created_by` (`created_by`),
  KEY `idx_create_time` (`create_time`),
  KEY `idx_cleanup` (`ref_count`,`deleted_at`),
  KEY `idx_business_lookup` (`business_type`,`business_id`,`is_primary`,`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件表 - 统一文件存储记录';

-- 1.8 数据字典表（双表模式：sys_dict_type + sys_dict_data）
DROP TABLE IF EXISTS sys_dict_data;
DROP TABLE IF EXISTS sys_dict_type;

CREATE TABLE sys_dict_type (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '字典类型 ID' PRIMARY KEY,
  dict_name varchar(100) NOT NULL COMMENT '字典名称',
  dict_code varchar(50) NOT NULL COMMENT '字典编码（如：mold-status）',
  remark varchar(255) DEFAULT '' COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_dict_code (dict_code) COMMENT '字典编码唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='字典类型表';

CREATE TABLE sys_dict_data (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '字典数据 ID' PRIMARY KEY,
  dict_type_id bigint NOT NULL COMMENT '字典类型 ID（外键）',
  dict_label varchar(100) NOT NULL COMMENT '字典显示标签',
  dict_value varchar(100) NOT NULL COMMENT '字典存储值',
  color varchar(20) DEFAULT 'default' COMMENT '颜色标签',
  sort int DEFAULT 0 COMMENT '排序',
  status tinyint DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
  is_locked TINYINT DEFAULT 0 COMMENT '锁定级别：0=未锁定，1=可解锁锁定，2=永久锁定（不可删除）',
  allow_production TINYINT DEFAULT 0 COMMENT '是否可用于生产：1=是，0=否（仅模具状态字典有效）',
  allow_work_report INT8 NOT NULL DEFAULT 1 COMMENT '是否在生产报工页面展示：1=是，0=否',
  needs_delivery_unlock TINYINT DEFAULT 0 COMMENT '是否需要交付后解锁：1=是，0=否（仅款项类型字典有效）',
  remark varchar(255) DEFAULT '' COMMENT '备注',
  ext text COMMENT '扩展配置（JSON格式，如付款方式对应的可用账户类型）',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  KEY idx_dict_type_id (dict_type_id) COMMENT '字典类型 ID 索引',
  KEY idx_sort (sort) COMMENT '排序索引',
  KEY idx_sys_dict_data_type_sort_status (dict_type_id, sort, status) COMMENT '字典类型+排序+状态组合索引，优化查询',
  KEY idx_status (status) COMMENT '状态索引',
  UNIQUE KEY uk_dict_type_label (dict_type_id, dict_label) COMMENT '字典类型和标签唯一约束',
  UNIQUE KEY uk_dict_type_value (dict_type_id, dict_value) COMMENT '字典类型和值唯一约束',
  CONSTRAINT fk_dict_data_type FOREIGN KEY (dict_type_id) REFERENCES sys_dict_type(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='字典数据表';

-- ============================================
-- 二、公共基础资料表（6 张）
-- ============================================

-- 2.1 客户表（customer）
DROP TABLE IF EXISTS customer;
CREATE TABLE customer (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '客户 ID',
  cus_code varchar(50) NOT NULL COMMENT '客户编号（唯一）',
  cus_name varchar(100) NOT NULL COMMENT '客户名称',
  contact varchar(50) DEFAULT NULL COMMENT '联系人',
  phone varchar(20) DEFAULT NULL COMMENT '联系电话',
  email varchar(100) DEFAULT NULL COMMENT '电子邮箱',
  address varchar(255) DEFAULT NULL COMMENT '客户地址',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  status tinyint DEFAULT '1' COMMENT '客户状态：0=停用，1=正常',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_cus_code (cus_code) COMMENT '客户编号唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户表';

-- 2.2 物料表（material）
DROP TABLE IF EXISTS material;
CREATE TABLE material (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '物料 ID',
  mat_code varchar(50) NOT NULL COMMENT '物料编码（唯一）',
  mat_name varchar(100) NOT NULL COMMENT '物料名称',
  spec varchar(255) DEFAULT NULL COMMENT '物料规格型号',
  unit varchar(20) DEFAULT NULL COMMENT '物料单位（如：个、kg）',
  mat_type tinyint DEFAULT 1 COMMENT '物料类型：1=原料，2=模具配件，3=产品，4=辅料',
  price decimal(10,2) DEFAULT 0.00 COMMENT '物料单价',
  stock_warn int DEFAULT 0 COMMENT '库存预警数量',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mat_code (mat_code) COMMENT '物料编码唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物料表';

-- 2.3 仓库表（warehouse）
DROP TABLE IF EXISTS warehouse;
CREATE TABLE warehouse (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '仓库 ID',
  wh_code varchar(50) NOT NULL COMMENT '仓库编码（唯一）',
  wh_name varchar(100) NOT NULL COMMENT '仓库名称',
  address varchar(255) DEFAULT NULL COMMENT '仓库地址',
  manager varchar(50) DEFAULT NULL COMMENT '仓库管理员',
  phone varchar(20) DEFAULT NULL COMMENT '联系电话',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_wh_code (wh_code) COMMENT '仓库编码唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仓库表';

-- 2.5 工序模板表（process）
DROP TABLE IF EXISTS process;
CREATE TABLE process (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '工序 ID',
  process_code varchar(50) NOT NULL COMMENT '工序编码（唯一）',
  process_name varchar(50) NOT NULL COMMENT '工序名称（如：CNC、线割、装配）',
  process_type tinyint DEFAULT 1 COMMENT '工序类型：1=模具工序，2=产品工序',
  work_hour decimal(8,2) DEFAULT 0.00 COMMENT '标准工时',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_process_code (process_code) COMMENT '工序编码唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工序模板表';

-- 2.6 加工商表（processor）
DROP TABLE IF EXISTS processor;
CREATE TABLE processor (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '加工商 ID',
  processor_code varchar(50) NOT NULL COMMENT '加工商编号（唯一）',
  processor_name varchar(100) NOT NULL COMMENT '加工商名称',
  contact varchar(50) DEFAULT NULL COMMENT '联系人',
  phone varchar(20) DEFAULT NULL COMMENT '联系电话',
  email varchar(100) DEFAULT NULL COMMENT '电子邮箱',
  address varchar(255) DEFAULT NULL COMMENT '加工商地址',
  process_type VARCHAR(100) DEFAULT NULL COMMENT '加工类型（用户自定义填写）',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_processor_code (processor_code) COMMENT '加工商编号唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='加工商表';

-- 2.7 供应商表（supplier）
DROP TABLE IF EXISTS supplier;
CREATE TABLE supplier (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '供应商 ID',
  sup_code varchar(50) NOT NULL COMMENT '供应商编号（唯一）',
  sup_name varchar(100) NOT NULL COMMENT '供应商名称',
  contact varchar(50) DEFAULT NULL COMMENT '联系人',
  phone varchar(20) DEFAULT NULL COMMENT '联系电话',
  email varchar(100) DEFAULT NULL COMMENT '电子邮箱',
  address varchar(255) DEFAULT NULL COMMENT '供应商地址',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_sup_code (sup_code) COMMENT '供应商编号唯一索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='供应商表';

-- ============================================
-- 三、模具制造管理表（7 张）【核心业务】
-- ============================================

-- 3.1 模具档案表（mold）
DROP TABLE IF EXISTS `mold`;
CREATE TABLE `mold` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模具 ID',
  `mold_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '模具编号（唯一）',
  `mold_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '模具名称（用户自定义，如：塑料多产品模具、手机壳专用模具等）',
  `customer_id` bigint NOT NULL COMMENT '客户 ID（customer.id）',
  `latest_order_id` bigint DEFAULT NULL COMMENT '最新关联订单 ID',
  `latest_work_order_id` bigint DEFAULT NULL COMMENT '最新关联工单 ID',
  `total_orders` int DEFAULT '0' COMMENT '累计订单数量',
  `total_work_orders` int DEFAULT '0' COMMENT '累计工单数量',
  `last_delivery_time` datetime DEFAULT NULL COMMENT '最后交付时间',
  `total_shots` int DEFAULT '0' COMMENT '累计使用模次',
  `production_date` datetime DEFAULT NULL COMMENT '开始生产日期',
  `process_status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'PROCESSING' COMMENT '工序状态：PROCESSING=制造中，PRODUCING=啤货，IN_STOCK=在仓，SCRAPPED=废弃',
  `life_cycle_status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'DESIGNING' COMMENT '生命周期状态：DESIGNING=设计中, PROCESSING=加工中, TESTING=试模, TESTED=已试模, MASS_PRODUCTION=量产, REPAIRING=维修, IN_STOCK=库存, DELIVERED=已交付, SCRAPPED=报废',
  `mold_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'NEW' COMMENT '模胚情况：NEW=全新，USED=二手',
  `mold_process` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'NORMAL' COMMENT '模具工艺：NORMAL=普通模具，HOT_RUNNER=热流道',
  `material_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'STEEL' COMMENT '材质类型：STEEL=钢材，HARDENED_STEEL=加硬钢材',
  `cavity_count` int DEFAULT '1' COMMENT '模腔数（一模几腔）',
  `production_capacity` decimal(10,2) DEFAULT NULL COMMENT '日产能（件/天）',
  `product_types` json DEFAULT NULL COMMENT '可生产产品类型列表（JSON数组）',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `current_location` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'IN_FACTORY' COMMENT '当前位置：IN_FACTORY=厂内，OUTSOURCING=外发中，AT_CUSTOMER=客户处',
  `last_usage_date` datetime DEFAULT NULL COMMENT '最后一次使用日期',
  `ownership_type` tinyint DEFAULT '1' COMMENT '权属类型: 1=自有, 2=客户资产, 3=第三方',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mold_code` (`mold_code`) COMMENT '模具编号唯一索引',
  KEY `idx_latest_order_id` (`latest_order_id`),
  KEY `idx_latest_work_order_id` (`latest_work_order_id`),
  KEY `idx_current_location` (`current_location`),
  KEY `idx_ownership_type` (`ownership_type`),
  KEY `idx_mold_lifecycle_status` (`life_cycle_status`),
  KEY `idx_mold_lifecycle_id` (`life_cycle_status`,`id`),
  KEY `idx_mold_customer_id` (`customer_id`),
  CONSTRAINT `fk_mold_customer` FOREIGN KEY (`customer_id`) REFERENCES `customer` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=54 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具档案表';

-- 3.2 模具订单表（mold_order）
DROP TABLE IF EXISTS `mold_order`;
CREATE TABLE `mold_order` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模具订单 ID',
  `order_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单编号（唯一）',
  `customer_id` bigint NOT NULL COMMENT '关联客户 ID（customer.id）',
  `order_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'NEW_MOLD' COMMENT '订单类型：NEW_MOLD=新模订单，REPAIR=修模订单，MODIFICATION=改造订单',
  `priority` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'NORMAL' COMMENT '优先级（字典：order_priority）',
  `mold_id` bigint DEFAULT NULL COMMENT '[已废弃] 请使用 mold_order_item 表关联模具',
  `plan_start_date` date DEFAULT NULL COMMENT '计划开始日期',
  `plan_end_date` date DEFAULT NULL COMMENT '订单交期',
  `total_amount` decimal(12,2) DEFAULT '0.00' COMMENT '订单总金额',
  `additional_amount` decimal(10,2) DEFAULT '0.00' COMMENT '附加金额',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'NOT_STARTED' COMMENT '订单状态：NOT_STARTED=未开始, PROCESSING=加工中, PENDING_TEST=待试模, TESTED=已试模, COMPLETED=已完成, CANCELLED=已取消, DELIVERED=已交付',
  `deliver_time` datetime DEFAULT NULL COMMENT '交付时间',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `master_order_id` bigint DEFAULT NULL COMMENT '关联综合订单 ID (master_orders.id)',
  `master_cancelled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '父综合订单已取消标记：0=否，1=是',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_code` (`order_code`) COMMENT '订单编号唯一索引',
  KEY `idx_customer_id` (`customer_id`) COMMENT '客户 ID 索引',
  KEY `idx_mold_id` (`mold_id`) COMMENT '模具 ID 索引',
  KEY `idx_order_type` (`order_type`) COMMENT '订单类型索引',
  KEY `idx_status` (`status`),
  KEY `idx_create_time` (`create_time` DESC),
  KEY `idx_master_order_id` (`master_order_id`),
  CONSTRAINT `fk_mold_order_customer` FOREIGN KEY (`customer_id`) REFERENCES `customer` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_mold_order_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具订单表';

-- 3.2.1 模具订单明细表（mold_order_item）【一单多款】
DROP TABLE IF EXISTS mold_order_item;
CREATE TABLE mold_order_item (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '明细ID',
  order_id BIGINT NOT NULL COMMENT '订单ID（mold_order.id）',
  mold_id BIGINT NOT NULL COMMENT '模具ID（mold.id）',
  amount DECIMAL(12,2) DEFAULT 0.00 COMMENT '该模具的合同金额',
  sort INT DEFAULT 0 COMMENT '排序号',
  remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_mold (order_id, mold_id) COMMENT '同一订单下模具唯一',
  KEY idx_order_id (order_id) COMMENT '订单ID索引',
  KEY idx_mold_id (mold_id) COMMENT '模具ID索引',
  CONSTRAINT fk_order_item_order FOREIGN KEY (order_id) REFERENCES mold_order(id) ON DELETE CASCADE,
  CONSTRAINT fk_order_item_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具订单明细表（一单多款）';

-- 3.3 模具 BOM 表（mold_bom）
DROP TABLE IF EXISTS mold_bom;
CREATE TABLE mold_bom (
  id bigint NOT NULL AUTO_INCREMENT COMMENT 'BOM ID',
  mold_id bigint NOT NULL COMMENT '关联模具 ID（mold.id）',
  material_id bigint NOT NULL COMMENT '关联物料 ID（material.id）',
  qty decimal(10,2) DEFAULT 1.00 COMMENT '物料用量',
  unit_price decimal(10,2) DEFAULT 0.00 COMMENT '物料单价',
  total_price decimal(10,2) DEFAULT 0.00 COMMENT '总价',
  remark varchar(255) DEFAULT NULL COMMENT '备注（如：模具配件规格）',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_mold_id (mold_id) COMMENT '模具 ID 索引',
  KEY idx_material_id (material_id) COMMENT '物料 ID 索引',
  CONSTRAINT fk_mold_bom_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE CASCADE,
  CONSTRAINT fk_mold_bom_material FOREIGN KEY (material_id) REFERENCES material(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具 BOM 表（模具用料清单）';

-- 3.4 模具生产工单表（mold_work_order）
DROP TABLE IF EXISTS `mold_work_order`;
CREATE TABLE `mold_work_order` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模具工单 ID',
  `work_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工单号（唯一）',
  `mold_order_id` bigint DEFAULT NULL COMMENT '关联模具订单 ID（mold_order.id）',
  `order_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'NEW_MOLD' COMMENT '订单类型：NEW_MOLD=新模，REPAIR=修模，MODIFICATION=改造',
  `work_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'IN_HOUSE' COMMENT '工单类型（字典：work-type）：IN_HOUSE=内部生产，OUTSOURCED=外发加工',
  `processor_id` bigint DEFAULT NULL COMMENT '加工商ID（外发时填写）',
  `outsource_price` decimal(10,2) DEFAULT '0.00' COMMENT '外发单价',
  `outsource_total_amount` decimal(12,2) DEFAULT '0.00' COMMENT '外发总金额',
  `mold_id` bigint DEFAULT NULL COMMENT '关联模具 ID（mold.id）',
  `plan_start` datetime DEFAULT NULL COMMENT '计划开始时间',
  `plan_end` datetime DEFAULT NULL COMMENT '计划结束时间',
  `actual_start` datetime DEFAULT NULL COMMENT '实际开始时间',
  `actual_end` datetime DEFAULT NULL COMMENT '实际结束时间',
  `progress` int DEFAULT '0' COMMENT '工单总进度（0-100）',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'NOT_STARTED' COMMENT '工单状态：NOT_STARTED=未开始, PROCESSING=加工中, COMPLETED=已完成, TESTED=已试模, STOCKED_IN=已入库, DELIVERED=已交付, CANCELLED=已取消',
  `warehouse_id` bigint DEFAULT NULL COMMENT '入库仓库 ID（warehouse.id）',
  `deliver_time` datetime DEFAULT NULL COMMENT '交付时间',
  `actual_deliver_time` datetime DEFAULT NULL COMMENT '实际交付时间',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `is_cancelled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否已取消：0=否，1=是',
  `master_cancelled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '父综合订单已取消标记：0=否，1=是',
  `cancel_time` datetime DEFAULT NULL COMMENT '取消时间',
  `cancel_reason` text COLLATE utf8mb4_unicode_ci COMMENT '取消原因',
  `cancel_user_id` bigint DEFAULT NULL COMMENT '取消操作人 ID（sys_user.id）',
  `original_mold_status` tinyint DEFAULT NULL COMMENT '修模/改造前模具的原始生命周期状态',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_work_code` (`work_code`) COMMENT '工单号唯一索引',
  KEY `idx_mold_order_id` (`mold_order_id`) COMMENT '模具订单 ID 索引',
  KEY `idx_mold_id` (`mold_id`) COMMENT '模具 ID 索引',
  KEY `idx_warehouse_id` (`warehouse_id`) COMMENT '仓库 ID 索引',
  KEY `idx_is_cancelled` (`is_cancelled`),
  KEY `idx_mold_work_order_mold_id_id` (`mold_id`,`id`),
  KEY `idx_status` (`status`),
  KEY `idx_order_type` (`order_type`),
  KEY `idx_work_type` (`work_type`),
  KEY `idx_processor_id` (`processor_id`),
  CONSTRAINT `fk_mold_work_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_mold_work_order` FOREIGN KEY (`mold_order_id`) REFERENCES `mold_order` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具生产工单表';

-- 3.5 工单工序关联表（work_order_process）
-- 已见 5.6 节

-- 3.6 模具试模记录表（mold_try_test）
DROP TABLE IF EXISTS mold_try_test;
CREATE TABLE mold_try_test (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '试模 ID',
  mold_work_id bigint NOT NULL COMMENT '关联模具工单 ID（mold_work_order.id）',
  test_time datetime DEFAULT NULL COMMENT '试模时间',
  ok_qty int DEFAULT 0 COMMENT '试模合格数量',
  bad_qty int DEFAULT 0 COMMENT '试模不良数量',
  result text DEFAULT NULL COMMENT '试模结果（如：合格、需返修）',
  test_mode varchar(20) DEFAULT NULL COMMENT '试模方式（自家试模、外协试模、免试）',
  user_id bigint DEFAULT NULL COMMENT '操作人 ID（sys_user.id）',
  remark text DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_mold_work_id (mold_work_id) COMMENT '模具工单 ID 索引',
  KEY idx_user_id (user_id) COMMENT '操作人 ID 索引',
  CONSTRAINT fk_mold_try_work FOREIGN KEY (mold_work_id) REFERENCES mold_work_order(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具试模记录表';

-- 3.7 模具库存记录表（mold_stock）
DROP TABLE IF EXISTS mold_stock;
CREATE TABLE mold_stock (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '库存记录 ID',
  mold_id bigint NOT NULL COMMENT '关联模具 ID（mold.id）',
  warehouse_id bigint DEFAULT NULL COMMENT '关联仓库 ID（warehouse.id）',
  stock_type tinyint NOT NULL COMMENT '库存类型：1=入库，2=出库，3=交付客户',
  operate_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  operator_id bigint DEFAULT NULL COMMENT '操作人 ID（sys_user.id）',
  customer_id bigint DEFAULT NULL COMMENT '交付客户 ID（stock_type=3 时必填，customer.id）',
  quantity int DEFAULT 1 COMMENT '数量',
  remark varchar(255) DEFAULT NULL COMMENT '备注（如：入库原因、交付单号）',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_mold_id (mold_id) COMMENT '模具 ID 索引',
  KEY idx_warehouse_id (warehouse_id) COMMENT '仓库 ID 索引',
  KEY idx_operator_id (operator_id) COMMENT '操作人 ID 索引',
  KEY idx_customer_id (customer_id) COMMENT '客户 ID 索引',
  CONSTRAINT fk_mold_stock_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具库存记录表（跟踪模具库存/交付状态）';

-- 3.8 模具产品关联表（mold_product_relation）
DROP TABLE IF EXISTS mold_product_relation;
CREATE TABLE mold_product_relation (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `mold_id` bigint NOT NULL COMMENT '模具ID (mold.id)',
  `product_id` bigint DEFAULT NULL COMMENT '产品ID (product.id)，可选',
  `production_order_item_id` bigint DEFAULT NULL COMMENT '生产订单明细ID (production_order_item.id)',
  `product_name` varchar(200) DEFAULT NULL COMMENT '产品名称快照（生产时填写）',
  `product_image_url` varchar(500) DEFAULT NULL COMMENT '产品图片URL（生产时上传）',
  `plan_qty` int DEFAULT '0' COMMENT '计划生产数量',
  `relation_type` varchar(50) DEFAULT 'PRODUCTION' COMMENT '关联类型: PRODUCTION=生产关联',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mold_item` (`mold_id`,`production_order_item_id`) COMMENT '模具-生产明细唯一索引',
  KEY `idx_mold_id` (`mold_id`) COMMENT '模具ID索引',
  KEY `idx_product_id` (`product_id`) COMMENT '产品ID索引',
  KEY `idx_production_order_item_id` (`production_order_item_id`) COMMENT '生产订单明细ID索引',
  CONSTRAINT `fk_mpr_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_mpr_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_mpr_production_item` FOREIGN KEY (`production_order_item_id`) REFERENCES `production_order_item` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='模具产品关联表（记录模具与产品的生产关系）';

-- ============================================
-- 四、产品生产管理表（6 张）
-- ============================================

-- 4.1 产品表（product）
DROP TABLE IF EXISTS product;
CREATE TABLE product (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '产品 ID',
  pro_code varchar(50) NOT NULL COMMENT '产品编码（唯一）',
  pro_name varchar(100) NOT NULL COMMENT '产品名称',
  spec varchar(255) DEFAULT NULL COMMENT '产品规格',
  spec_detail TEXT DEFAULT NULL COMMENT '产品详细规格参数(JSON格式)',
  mold_id bigint DEFAULT NULL COMMENT '默认使用模具 ID（mold.id）',
  unit varchar(20) DEFAULT NULL COMMENT '产品单位',
  price decimal(10,2) DEFAULT 0.00 COMMENT '产品销售单价',
  weight decimal(10,2) DEFAULT 0.00 COMMENT '产品重量',
  status tinyint NOT NULL DEFAULT '1' COMMENT '产品状态：0=停用, 1=启用',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  image_url varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品图片URL（手动上传，优先展示）',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_pro_code (pro_code) COMMENT '产品编码唯一索引',
  KEY idx_mold_id (mold_id) COMMENT '模具 ID 索引',
  CONSTRAINT fk_product_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品表';

-- 4.2 产品 BOM 表（product_bom）
DROP TABLE IF EXISTS product_bom;
CREATE TABLE product_bom (
  id bigint NOT NULL AUTO_INCREMENT COMMENT 'BOM ID',
  product_id bigint NOT NULL COMMENT '关联产品 ID（product.id）',
  material_id bigint NOT NULL COMMENT '关联物料 ID（material.id）',
  qty decimal(10,2) DEFAULT 1.00 COMMENT '单台产品物料用量',
  loss_rate decimal(5,2) DEFAULT 0.00 COMMENT '物料损耗率（%）',
  unit_price decimal(10,2) DEFAULT 0.00 COMMENT '物料单价',
  total_price decimal(10,2) DEFAULT 0.00 COMMENT '总价',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_product_id (product_id) COMMENT '产品 ID 索引',
  KEY idx_material_id (material_id) COMMENT '物料 ID 索引',
  CONSTRAINT fk_product_bom_product FOREIGN KEY (product_id) REFERENCES product(id) ON DELETE CASCADE,
  CONSTRAINT fk_product_bom_material FOREIGN KEY (material_id) REFERENCES material(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品 BOM 表（产品用料清单）';

-- 4.4 库存表（stock）
DROP TABLE IF EXISTS stock;
CREATE TABLE stock (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '库存 ID',
  material_id bigint NOT NULL COMMENT '关联物料 ID（material.id）',
  product_id bigint DEFAULT NULL COMMENT '产品ID（product.id），用于区分同一物料不同产品的库存',
  wh_id bigint NOT NULL COMMENT '关联仓库 ID（warehouse.id）',
  stock_qty decimal(12,2) DEFAULT 0.00 COMMENT '当前库存数量',
  safe_stock decimal(12,2) DEFAULT 0.00 COMMENT '安全库存数量',
  remark varchar(255) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mat_wh (material_id, wh_id) COMMENT '唯一索引，避免同一物料同一仓库重复记录',
  KEY idx_product_id (product_id) COMMENT '产品 ID 索引',
  KEY idx_wh_id (wh_id) COMMENT '仓库 ID 索引',
  CONSTRAINT fk_stock_material FOREIGN KEY (material_id) REFERENCES material(id) ON DELETE CASCADE,
  CONSTRAINT fk_stock_warehouse FOREIGN KEY (wh_id) REFERENCES warehouse(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存表';

-- ============================================
-- 五、扩展功能表（新增）
-- ============================================

-- 5.1 模具交付记录表
DROP TABLE IF EXISTS mold_delivery;
CREATE TABLE mold_delivery (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '交付 ID',
  mold_work_id bigint NOT NULL COMMENT '关联模具工单 ID',
  delivery_no varchar(50) DEFAULT NULL COMMENT '交付单号',
  customer_id bigint NOT NULL COMMENT '客户 ID',
  delivery_time datetime DEFAULT NULL COMMENT '交付时间',
  deliverer_id bigint DEFAULT NULL COMMENT '交付人 ID',
  receiver_name varchar(50) DEFAULT NULL COMMENT '收货人姓名',
  receiver_phone varchar(20) DEFAULT NULL COMMENT '收货人电话',
  delivery_address text DEFAULT NULL COMMENT '交付地址',
  logistics_company varchar(100) DEFAULT NULL COMMENT '物流公司',
  tracking_no varchar(50) DEFAULT NULL COMMENT '物流单号',
  delivery_method varchar(20) DEFAULT NULL COMMENT '交付方式：自提/快递/物流',
  documents_attached text DEFAULT NULL COMMENT '随附文档（JSON格式）',
  confirmation_status tinyint DEFAULT 1 COMMENT '确认状态：1=待确认，2=已确认，3=拒收',
  confirmation_time datetime DEFAULT NULL COMMENT '确认时间',
  rejection_reason text DEFAULT NULL COMMENT '拒收原因',
  remark text DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_delivery_no (delivery_no) COMMENT '交付单号唯一索引',
  KEY idx_mold_work_id (mold_work_id) COMMENT '模具工单 ID 索引',
  KEY idx_customer_id (customer_id) COMMENT '客户 ID 索引',
  CONSTRAINT fk_mold_delivery_work FOREIGN KEY (mold_work_id) REFERENCES mold_work_order(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具交付记录表';

-- 5.3 模具图片（已废弃，统一由 files 表管理）
DROP TABLE IF EXISTS `mold_image`;

-- 5.4 模具订单款项表
DROP TABLE IF EXISTS mold_order_payment;
CREATE TABLE mold_order_payment (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '款项 ID',
  order_id bigint NOT NULL COMMENT '关联模具订单 ID',
  work_order_id BIGINT COMMENT '关联工单 ID（mold_work_order.id），用于按模具核算',
  payment_type varchar(50) NOT NULL COMMENT '款项类型（自定义，如：首款、二期款、尾款等）',
  contract_amount decimal(12,2) DEFAULT 0.00 COMMENT '合同约定金额',
  actual_amount decimal(12,2) DEFAULT 0.00 COMMENT '实际到账金额',
  should_amount decimal(12,2) DEFAULT 0.00 COMMENT '应收尾款金额',
  payment_date datetime DEFAULT NULL COMMENT '收款日期',
  status tinyint DEFAULT 1 COMMENT '状态：1=未收，2=部分收款，3=已收款',
  remark varchar(500) DEFAULT NULL COMMENT '备注',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_order_id (order_id),
  KEY idx_payment_work_order_id (work_order_id),
  CONSTRAINT fk_payment_order FOREIGN KEY (order_id) REFERENCES mold_order(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具订单款项表';

-- 5.5 工序模板分组表
DROP TABLE IF EXISTS process_template;
CREATE TABLE process_template (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  group_id int DEFAULT 0 COMMENT '分组ID（同组工序可并行）',
  group_name varchar(50) DEFAULT NULL COMMENT '分组名称',
  process_name varchar(100) NOT NULL COMMENT '工序名称（如：CNC、磨床、放电等）',
  sort int NOT NULL DEFAULT 0 COMMENT '排序号',
  status tinyint NOT NULL DEFAULT 1 COMMENT '状态：1=启用，2=禁用',
  remark varchar(500) DEFAULT NULL COMMENT '备注',
  created_at datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_sort (sort),
  KEY idx_status (status),
  KEY idx_group_id (group_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工序模板表';

-- 5.6 工单工序关联表
DROP TABLE IF EXISTS work_order_process;
CREATE TABLE work_order_process (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  work_order_id bigint NOT NULL COMMENT '工单ID',
  process_template_id bigint NOT NULL COMMENT '工序模板ID',
  is_completed TINYINT NOT NULL DEFAULT 0 COMMENT '工序状态：0=未完成，1=已完成，2=跳过/不需要',
  completed_at datetime DEFAULT NULL COMMENT '完成时间',
  remark varchar(500) DEFAULT NULL COMMENT '备注',
  created_at datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_work_process (work_order_id, process_template_id),
  KEY idx_work_order (work_order_id),
  KEY idx_template (process_template_id),
  KEY idx_completed (is_completed),
  CONSTRAINT fk_wop_work_order FOREIGN KEY (work_order_id) REFERENCES mold_work_order(id) ON DELETE CASCADE,
  CONSTRAINT fk_wop_template FOREIGN KEY (process_template_id) REFERENCES process_template(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='工单工序关联表';

-- 5.7 状态流转配置表（新版-语义化字符串模型）
DROP TABLE IF EXISTS status_flow_config;
CREATE TABLE status_flow_config (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '配置ID',

  -- 三个状态字段（至少填写一个）- 使用语义化字符串
  work_order_status VARCHAR(50) COMMENT '工单状态值（语义化字符串，如 PROCESSING）',
  mold_order_status VARCHAR(50) COMMENT '订单状态值（语义化字符串，如 PROCESSING）',
  mold_status VARCHAR(50) COMMENT '模具档案状态值（语义化字符串，如 DESIGNING）',

  -- 配置说明
  description VARCHAR(500) COMMENT '配置说明',
  priority INT DEFAULT 0 COMMENT '优先级（数字越小优先级越高）',
  is_active TINYINT DEFAULT 1 COMMENT '是否启用：1=启用，0=禁用',
  status_order INT DEFAULT 0 COMMENT '状态顺序号（用于状态比较）',

  -- 审计字段
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

  INDEX idx_work_order_status (work_order_status),
  INDEX idx_mold_order_status (mold_order_status),
  INDEX idx_mold_status (mold_status),
  INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='状态流转配置表（新版-语义化字符串）';

-- ============================================
-- 六、生产管理模块表（新增）
-- ============================================

-- 6.1 生产订单主表
DROP TABLE IF EXISTS production_order;
CREATE TABLE production_order (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '生产订单ID',
  order_code VARCHAR(50) NOT NULL COMMENT '订单编号（唯一）',
  customer_id BIGINT NOT NULL COMMENT '客户ID（customer.id）',
  product_name VARCHAR(200) NOT NULL DEFAULT '' COMMENT '产品名称（手动输入）',
  
  -- 生产配置
  production_type VARCHAR(50) NOT NULL DEFAULT 'IN_HOUSE' COMMENT '生产方式：IN_HOUSE=内部生产，OUTSOURCED=外发生产',
  mold_source_type VARCHAR(50) NOT NULL DEFAULT 'SELF_MADE' COMMENT '模具来源：SELF_MADE=自有模具，CUSTOMER_PROVIDED=外来模具',
  mold_id BIGINT COMMENT '模具ID（mold.id），自有模具时必填',
  external_mold_code VARCHAR(100) DEFAULT NULL COMMENT '外来模具编号（外来模具时填写）',
  external_mold_info JSON COMMENT '外来模具信息JSON（模具来源=外来时必填）',
  
  -- 生产计划
  plan_qty INT NOT NULL DEFAULT 0 COMMENT '计划生产数量',
  completed_qty INT DEFAULT 0 COMMENT '已完成数量',
  ok_qty INT DEFAULT 0 COMMENT '良品数量',
  bad_qty INT DEFAULT 0 COMMENT '不良数量',
  mold_shots INT DEFAULT 0 COMMENT '模具使用次数(模次)',
  plan_start_date DATE COMMENT '计划开始日期',
  plan_end_date DATE COMMENT '计划交期',
  actual_start_date DATE COMMENT '实际开始日期',
  actual_end_date DATE COMMENT '实际完成日期',
  
  -- 资源配置（内部生产）
  device_id VARCHAR(50) COMMENT '生产设备字典值',
  worker_id VARCHAR(50) COMMENT '负责人字典值',
  
  -- 工艺信息
  process_type VARCHAR(50) COMMENT '生产工艺类型：injection=注塑, blow_molding=吹塑, stamping=冲压',
  process_remark TEXT COMMENT '工艺备注/详细说明',
  material_type VARCHAR(100) COMMENT '原料类型：PP, PE, TPE等',
  product_weight DECIMAL(10,2) COMMENT '产品重量（克）',
  has_metal_parts BOOLEAN DEFAULT FALSE COMMENT '是否包含金属配件',
  manual_process TEXT COMMENT '手工处理流程说明',
  assembly_process TEXT COMMENT '组装流程说明',
  
  -- 外发配置（外发生产）
  processor_id BIGINT COMMENT '加工商ID（processor.id）',
  outsource_price DECIMAL(10,2) DEFAULT 0.00 COMMENT '外发单价',
  outsource_total_amount DECIMAL(10,2) DEFAULT 0.00 COMMENT '外发总金额',
  
  -- 成本核算
  material_cost DECIMAL(10,2) DEFAULT 0.00 COMMENT '材料成本',
  processing_cost DECIMAL(10,2) DEFAULT 0.00 COMMENT '加工成本',
  total_cost DECIMAL(10,2) DEFAULT 0.00 COMMENT '总成本',
  unit_price DECIMAL(10,2) DEFAULT 0.00 COMMENT '产品单价',
  total_amount DECIMAL(10,2) DEFAULT 0.00 COMMENT '订单总金额',
  
  -- 状态管理
  status VARCHAR(50) NOT NULL DEFAULT 'PENDING_SCHEDULE' COMMENT '订单状态：PENDING_SCHEDULE=待排产，SCHEDULED=已排产，PRODUCING=生产中，STOCKED_IN=已入库，SHIPPED=已出货，COMPLETED=已完成，CANCELLED=已取消',
  priority VARCHAR(50) DEFAULT 'NORMAL' COMMENT '优先级：NORMAL=普通，IMPORTANT=重要，URGENT=紧急，VERY_URGENT=非常紧急',
  
  -- 其他
  remark TEXT COMMENT '备注',
  create_user_id BIGINT COMMENT '创建人ID',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  -- 关联综合订单 ID
  master_order_id BIGINT COMMENT '关联综合订单 ID (master_orders.id)',
  master_cancelled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '父综合订单已取消标记：0=否，1=是',
  dependency_status VARCHAR(50) DEFAULT 'NONE' COMMENT '依赖状态: NONE=无依赖, WAITING_MOLD=等待模具, READY=就绪',
  
  PRIMARY KEY (id),
  UNIQUE KEY uk_order_code (order_code),
  KEY idx_customer_id (customer_id),
  KEY idx_mold_id (mold_id),
  KEY idx_production_type (production_type),
  KEY idx_mold_source_type (mold_source_type),
  KEY idx_status (status),
  KEY idx_plan_end_date (plan_end_date),
  KEY idx_master_order_id (master_order_id),
  KEY idx_mold_status (mold_id, status),
  KEY idx_customer_status (customer_id, status),
  KEY idx_processor_id (processor_id),
  
  CONSTRAINT fk_po_customer FOREIGN KEY (customer_id) REFERENCES customer(id) ON DELETE CASCADE,
  CONSTRAINT fk_po_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE SET NULL,
  CONSTRAINT fk_po_processor FOREIGN KEY (processor_id) REFERENCES processor(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='生产订单表';

-- 6.2 生产订单明细表（支持一单多产品）
DROP TABLE IF EXISTS production_order_item;
CREATE TABLE production_order_item (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '明细ID',
  `order_id` bigint NOT NULL COMMENT '订单ID',
  `product_id` bigint DEFAULT NULL COMMENT '产品ID（可选，手动输入产品名称时为空）',
  `product_name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '产品名称（独立于模具，如：手机壳A）',
  `mold_id` bigint DEFAULT NULL COMMENT '模具ID',
  `mold_ids` json DEFAULT NULL COMMENT '关联模具ID列表（JSON数组，支持多模具）',
  `mold_relation_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'SINGLE' COMMENT '模具关联类型: SINGLE=单模具, MULTI=多模具',
  `external_mold_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '外来模具编号（外来模具时填写）',
  `device_id` bigint DEFAULT NULL COMMENT '设备ID',
  `plan_qty` int NOT NULL DEFAULT '0' COMMENT '计划数量',
  `process_remark` text COLLATE utf8mb4_unicode_ci COMMENT '工艺参数/备注',
  `completed_qty` int DEFAULT '0' COMMENT '完成数量',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'producing' COMMENT '明细项状态：PENDING_SCHEDULE=待排产, SCHEDULED=已排产, PRODUCING=生产中, PAUSED=已暂停, STOCKED_IN=已入库, COMPLETED=已完成',
  `actual_completion_time` datetime DEFAULT NULL COMMENT '实际完成时间（该明细项完成生产的时刻）',
  `ok_qty` int DEFAULT '0' COMMENT '良品数',
  `bad_qty` int DEFAULT '0' COMMENT '不良数',
  `sort` int DEFAULT '0' COMMENT '排序',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `unit_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '单价（元）',
  `packaging_fee` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '包装费（元）',
  `freight_fee` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '运费（元）',
  `outsource_price` decimal(10,2) NOT NULL DEFAULT '0.00' COMMENT '外发单价（元）',
  `dependency_status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'NONE' COMMENT '依赖状态: NONE=无依赖, WAITING_MOLD=等待模具, READY=就绪',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `product_image_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品图片URL',
  `plastic_material_id` int DEFAULT NULL COMMENT '塑胶原料字典值（关联sys_dict_data.id）',
  `weight` decimal(10,2) DEFAULT NULL COMMENT '产品重量（克）',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_product_id` (`product_id`),
  KEY `fk_poi_mold` (`mold_id`),
  KEY `idx_order_mold` (`order_id`,`mold_id`),
  CONSTRAINT `fk_poi_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_poi_order` FOREIGN KEY (`order_id`) REFERENCES `production_order` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_poi_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产订单明细表';

-- 6.3 生产排程表
DROP TABLE IF EXISTS production_schedule;
CREATE TABLE production_schedule (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '排程ID',
  `order_id` bigint NOT NULL COMMENT '订单ID',
  `schedule_date` date NOT NULL COMMENT '排产日期',
  `time_slot` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '时间段：MORNING=上午，AFTERNOON=下午，ALL_DAY=全天',
  `device_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分配设备',
  `worker_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分配人员',
  `plan_qty` int DEFAULT '0' COMMENT '计划产量',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'PENDING' COMMENT '状态：PENDING=待执行，RUNNING=执行中，COMPLETED=已完成，DELAYED=已延期',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_schedule_date` (`schedule_date`),
  KEY `idx_device_id` (`device_id`),
  CONSTRAINT `fk_ps_order` FOREIGN KEY (`order_id`) REFERENCES `production_order` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产排程表';

-- 6.4 生产报工表
DROP TABLE IF EXISTS production_report;
CREATE TABLE production_report (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '报工ID',
  order_id BIGINT NOT NULL COMMENT '订单ID',
  schedule_id BIGINT COMMENT '排程ID',
  report_date DATETIME NOT NULL COMMENT '报工时间',
  item_id BIGINT COMMENT '订单明细项ID（支持按明细报工）',
  shift VARCHAR(50) COMMENT '班次（字典：shift-type）',
  ok_qty INT DEFAULT 0 COMMENT '良品数量',
  bad_qty INT DEFAULT 0 COMMENT '不良数量',
  scrap_qty INT DEFAULT 0 COMMENT '报废数量',
  device_id VARCHAR(50) COMMENT '使用设备',
  worker_id VARCHAR(50) COMMENT '操作工',
  work_hours DECIMAL(5,2) DEFAULT 0.00 COMMENT '工时（小时）',
  defect_reason TEXT COMMENT '不良原因',
  remark TEXT COMMENT '备注',
  create_user_id BIGINT COMMENT '报工人ID',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (id),
  KEY idx_order_id (order_id),
  KEY idx_schedule_id (schedule_id),
  KEY idx_report_date (report_date),
  KEY fk_pr_device (device_id),
  KEY idx_item_id (item_id),
  
  CONSTRAINT fk_pr_order FOREIGN KEY (order_id) REFERENCES production_order(id) ON DELETE CASCADE,
  CONSTRAINT fk_pr_schedule FOREIGN KEY (schedule_id) REFERENCES production_schedule(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='生产报工表';

-- 6.5 外发记录表
DROP TABLE IF EXISTS outsourcing_record;
CREATE TABLE outsourcing_record (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '外发记录ID',
  `record_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '外发单号（唯一）',
  `order_id` bigint DEFAULT NULL COMMENT '关联生产订单ID(production_order.id),生产外发时填写',
  `outsource_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'PRODUCTION' COMMENT '外发类型：PRODUCTION=生产外发，MOLD=模具外发',
  `work_order_id` bigint DEFAULT NULL COMMENT '关联模具工单ID(mold_work_order.id),模具外发时填写',
  `processor_id` bigint NOT NULL COMMENT '加工商ID',
  `mold_id` bigint DEFAULT NULL COMMENT '外发模具ID（自有模具）',
  `external_mold_info` json DEFAULT NULL COMMENT '外来模具信息',
  `outsource_date` datetime DEFAULT NULL COMMENT '外发日期',
  `expected_return_date` datetime DEFAULT NULL COMMENT '预计返回日期',
  `actual_return_date` datetime DEFAULT NULL COMMENT '实际返回日期',
  `quantity` int DEFAULT '0' COMMENT '外发数量',
  `unit_price` decimal(10,2) DEFAULT '0.00' COMMENT '单价',
  `total_amount` decimal(10,2) DEFAULT '0.00' COMMENT '总金额',
  `paid_amount` decimal(10,2) DEFAULT '0.00' COMMENT '已付金额',
  `logistics_company` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '物流公司',
  `tracking_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '物流单号',
  `delivery_address` text COLLATE utf8mb4_unicode_ci COMMENT '交付地址',
  `receiver_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收货人',
  `receiver_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收货电话',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'PENDING' COMMENT '状态：PENDING=待外发，SENT=已外发，PRODUCING=生产中，RETURNED=已返回，CANCELLED=已取消',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `share_token` varchar(32) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '分享token（8位随机字符，用于加工商免登录报工）',
  `create_user_id` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_record_code` (`record_code`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_processor_id` (`processor_id`),
  KEY `idx_status` (`status`),
  KEY `fk_or_mold` (`mold_id`),
  KEY `idx_work_order_id` (`work_order_id`),
  KEY `idx_outsource_type` (`outsource_type`),
  UNIQUE KEY `idx_share_token` (`share_token`),
  CONSTRAINT `fk_or_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_or_order` FOREIGN KEY (`order_id`) REFERENCES `production_order` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_or_processor` FOREIGN KEY (`processor_id`) REFERENCES `processor` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外发记录表';

-- 6.7 外发报工表
DROP TABLE IF EXISTS outsourcing_report;
CREATE TABLE outsourcing_report (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '外发报工ID',
  outsourcing_id BIGINT NOT NULL COMMENT '外发记录ID',
  report_date DATETIME COMMENT '报工日期',
  ok_qty INT DEFAULT 0 COMMENT '良品数量',
  bad_qty INT DEFAULT 0 COMMENT '不良数量',
  defect_reason TEXT COMMENT '不良原因',
  photos JSON COMMENT '现场照片URL列表',
  remark TEXT COMMENT '备注',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (id),
  KEY idx_outsourcing_id (outsourcing_id),
  
  CONSTRAINT fk_orsr_outsourcing FOREIGN KEY (outsourcing_id) REFERENCES outsourcing_record(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='外发报工表';

-- 6.8 生产入库表
DROP TABLE IF EXISTS production_stock_in;
CREATE TABLE production_stock_in (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '入库ID',
  stock_code VARCHAR(50) NOT NULL COMMENT '入库单号（唯一）',
  order_id BIGINT NULL COMMENT '订单ID（直接入库时为空）',
  warehouse_id BIGINT NOT NULL COMMENT '仓库ID',
  mold_id BIGINT DEFAULT NULL COMMENT '模具ID（自有模具）',
  external_mold_code VARCHAR(100) DEFAULT NULL COMMENT '外来模具编号（外来模具）',
  product_id BIGINT DEFAULT NULL COMMENT '产品ID（product.id），用于区分同一模具不同产品的入库',
  product_name VARCHAR(200) DEFAULT '' COMMENT '产品名称（入库时的产品名称）',
  quantity INT NOT NULL COMMENT '入库数量',
  batch_no VARCHAR(50) COMMENT '批次号',
  stock_in_date DATETIME COMMENT '入库日期',
  operator_id BIGINT COMMENT '操作人ID',
  remark TEXT COMMENT '备注',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_code (stock_code),
  KEY idx_order_id (order_id),
  KEY idx_warehouse_id (warehouse_id),
  KEY idx_stock_in_date (stock_in_date),
  KEY idx_mold_id (mold_id),
  KEY idx_external_mold_code (external_mold_code),
  KEY idx_product_name (product_name),
  KEY idx_stock_in_product_id (product_id),
  KEY idx_stock_code (stock_code),
  KEY idx_mold_product (mold_id, product_id),
  KEY idx_mold_date (mold_id, stock_in_date DESC),
  KEY idx_external_product (external_mold_code, product_id),

  CONSTRAINT fk_psi_order FOREIGN KEY (order_id) REFERENCES production_order(id) ON DELETE SET NULL,
  CONSTRAINT fk_psi_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouse(id) ON DELETE CASCADE,
  CONSTRAINT fk_production_stock_in_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci 
COMMENT='生产入库表';

-- 6.9 生产出库表
DROP TABLE IF EXISTS production_stock_out;
CREATE TABLE production_stock_out (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '出库ID',
  stock_code VARCHAR(50) DEFAULT NULL COMMENT '出库单号（正常发货时有值，其他出库类型为NULL）',
  order_id BIGINT DEFAULT NULL COMMENT '订单ID（直接出库时为空）',
  warehouse_id BIGINT NOT NULL COMMENT '仓库ID',
  mold_id BIGINT DEFAULT NULL COMMENT '模具ID（自有模具）',
  external_mold_code VARCHAR(100) DEFAULT NULL COMMENT '外来模具编号（外来模具）',
  product_id BIGINT DEFAULT NULL COMMENT '产品ID（product.id），用于区分同一模具不同产品的出库',
  product_name VARCHAR(200) DEFAULT NULL COMMENT '产品名称（出库时的产品名称）',
  customer_name VARCHAR(200) DEFAULT NULL COMMENT '接收客户名称',
  quantity INT NOT NULL DEFAULT 0 COMMENT '出库数量',
  batch_no VARCHAR(50) DEFAULT NULL COMMENT '批次号',
  stock_out_date DATETIME DEFAULT NULL COMMENT '出库日期',
  out_type VARCHAR(20) DEFAULT 'normal' COMMENT '出库类型(normal=正常发货, other=其他原因出库)',
  out_reason VARCHAR(50) DEFAULT NULL COMMENT '出库原因(expired=过期丢弃, damaged=损坏, unqualified=不达标, other=其他原因)',
  out_reason_remark VARCHAR(500) DEFAULT NULL COMMENT '出库原因备注',
  delivery_date DATE DEFAULT NULL COMMENT '发货日期',
  delivery_method VARCHAR(50) DEFAULT NULL COMMENT '发货方式(字典值: express/lalamove/pickup/logistics)',
  express_company VARCHAR(100) DEFAULT NULL COMMENT '快递公司(字典值)',
  express_no VARCHAR(100) DEFAULT NULL COMMENT '快递单号',
  photos JSON DEFAULT NULL COMMENT '发货照片(JSON数组)',
  total_weight DECIMAL(10,2) DEFAULT NULL COMMENT '总重量(kg)',
  driver_phone VARCHAR(50) DEFAULT NULL COMMENT '司机电话(货拉拉用)',
  driver_plate VARCHAR(50) DEFAULT NULL COMMENT '车牌号(货拉拉用)',
  operator_id BIGINT DEFAULT NULL COMMENT '操作人ID',
  remark TEXT COMMENT '备注',
  create_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  
  PRIMARY KEY (id),
  KEY idx_order_id (order_id),
  KEY idx_warehouse_id (warehouse_id),
  KEY idx_mold_id (mold_id),
  KEY idx_external_mold_code (external_mold_code),
  KEY idx_stock_out_date (stock_out_date),
  KEY idx_product_name (product_name),
  KEY idx_stock_code (stock_code),
  KEY idx_out_type (out_type),
  KEY idx_product_id (product_id),
  KEY idx_mold_product (mold_id, product_id),
  KEY idx_mold_date (mold_id, stock_out_date DESC),
  KEY idx_external_product (external_mold_code, product_id),

  CONSTRAINT fk_stock_out_order FOREIGN KEY (order_id) REFERENCES production_order(id) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT fk_stock_out_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouse(id) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_stock_out_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产出库表';

-- 6.10 模具使用记录表
DROP TABLE IF EXISTS mold_usage_record;
CREATE TABLE mold_usage_record (
  id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '使用记录ID',
  mold_id BIGINT NOT NULL COMMENT '模具ID',
  production_order_id BIGINT NOT NULL COMMENT '生产订单ID',
  production_order_item_id BIGINT DEFAULT NULL COMMENT '关联生产订单明细ID（用于细分到具体产品）',
  product_id BIGINT DEFAULT NULL COMMENT '产品ID（用于区分同一模具不同产品）',
  product_name VARCHAR(200) DEFAULT NULL COMMENT '产品名称快照',
  usage_date DATE COMMENT '使用日期',
  total_shots INT DEFAULT 0 COMMENT '累计模次(本次使用)',
  ok_qty INT DEFAULT 0 COMMENT '良品数量',
  bad_qty INT DEFAULT 0 COMMENT '不良数量',
  start_date DATE COMMENT '开始使用日期',
  end_date DATE COMMENT '结束使用日期',
  remark TEXT COMMENT '备注',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP,
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  INDEX idx_mold_id (mold_id),
  INDEX idx_production_order_id (production_order_id),
  INDEX idx_production_order_item_id (production_order_item_id),
  INDEX idx_product_id (product_id),
  INDEX idx_mur_mold_product (mold_id, product_id),
  INDEX idx_mur_item_product (production_order_item_id, product_id),
  UNIQUE KEY uk_mold_order (mold_id, production_order_id),

  CONSTRAINT fk_usage_mold FOREIGN KEY (mold_id) REFERENCES mold(id) ON DELETE CASCADE,
  CONSTRAINT fk_usage_order FOREIGN KEY (production_order_id) REFERENCES production_order(id) ON DELETE CASCADE,
  CONSTRAINT fk_mur_production_item FOREIGN KEY (production_order_item_id) REFERENCES production_order_item(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具使用记录表';

-- 6.11 外来模具档案表
DROP TABLE IF EXISTS external_mold_archive;
CREATE TABLE external_mold_archive (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '档案ID',
  mold_code VARCHAR(100) NOT NULL COMMENT '外来模具编号（唯一）',
  product_name VARCHAR(200) NOT NULL COMMENT '产品名称',
  customer_id BIGINT UNSIGNED DEFAULT NULL COMMENT '客户ID（可选）',
  customer_name VARCHAR(100) DEFAULT NULL COMMENT '客户名称（冗余字段，便于查询）',
  first_order_id BIGINT UNSIGNED DEFAULT NULL COMMENT '首次使用的订单ID',
  first_use_date DATETIME DEFAULT NULL COMMENT '首次使用日期',
  last_order_id BIGINT UNSIGNED DEFAULT NULL COMMENT '最近使用的订单ID',
  last_use_date DATETIME DEFAULT NULL COMMENT '最近使用日期',
  use_count INT UNSIGNED DEFAULT 0 COMMENT '使用次数',
  status TINYINT DEFAULT 1 COMMENT '状态：1=启用，0=禁用',
  remark VARCHAR(500) DEFAULT NULL COMMENT '备注',
  create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_mold_code (mold_code),
  KEY idx_customer_id (customer_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='外来模具档案表';

-- 6.14 打印通知单配置表
DROP TABLE IF EXISTS print_notice_config;
CREATE TABLE print_notice_config (
  id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID',
  user_id BIGINT NOT NULL COMMENT '用户ID，0表示全局配置',
  config_type VARCHAR(50) NOT NULL DEFAULT 'production_order' COMMENT '配置类型：production_order=生产订单',
  config_data JSON NOT NULL COMMENT '配置数据（JSON格式）',
  is_default TINYINT(1) DEFAULT 0 COMMENT '是否为默认配置：1=是，0=否',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE KEY uk_user_type (user_id, config_type),
  INDEX idx_user_id (user_id),
  INDEX idx_config_type (config_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='打印通知单配置表';

-- 6.15 综合订单表 (master_orders)
DROP TABLE IF EXISTS master_orders;
CREATE TABLE master_orders (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '综合订单 ID',
  `order_code` varchar(50) NOT NULL COMMENT '综合订单编号（唯一）',
  `customer_id` bigint NOT NULL COMMENT '客户 ID',
  `order_type` varchar(50) NOT NULL DEFAULT 'COMBINED' COMMENT '下单类型: COMBINED=模具+产品, MOLD_ONLY=仅模具, PROD_ONLY=仅产品',
  `status` varchar(50) NOT NULL DEFAULT 'PENDING' COMMENT '整体状态',
  `total_amount` decimal(10,2) DEFAULT '0.00' COMMENT '订单总金额',
  `additional_amount` decimal(10,2) DEFAULT '0.00' COMMENT '附加金额（税点、杂费等）',
  `remark` text COMMENT '备注',
  `cancel_reason` text COMMENT '取消原因',
  `create_user_id` bigint DEFAULT NULL COMMENT '创建人 ID',
  `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deposit_ratio` int NOT NULL DEFAULT '30' COMMENT '首付款比例（5%一档，默认30%）',
  `quick_plan_end` date DEFAULT NULL COMMENT '订单交期（综合级别，快捷编辑用）',
  PRIMARY KEY (`id`),
  UNIQUE KEY `order_code` (`order_code`),
  KEY `idx_customer_id` (`customer_id`),
  KEY `idx_order_code` (`order_code`),
  KEY `idx_create_time` (`create_time` DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='综合订单表';

-- ============================================
-- 七、财务管理模块表
-- ============================================

-- 7.1 应收账款表（accounts_receivable）
DROP TABLE IF EXISTS accounts_receivable;
CREATE TABLE accounts_receivable (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '应收账款ID',
  `customer_id` bigint NOT NULL COMMENT '客户ID（customer.id）',
  `order_id` bigint DEFAULT NULL COMMENT '关联订单ID（mold_order.id 或 sale_order.id）',
  `order_type` varchar(50) DEFAULT NULL COMMENT '订单类型：mold_order=模具订单, sale_order=销售订单',
  `payment_type` varchar(50) DEFAULT NULL COMMENT '款项类型：total=总应收, deposit=首付款, final=尾款, production=生产款, freight=运费, packaging=包装费',
  `invoice_no` varchar(50) DEFAULT NULL COMMENT '发票号码',
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '应收金额',
  `received_amount` decimal(12,2) DEFAULT '0.00' COMMENT '已收金额',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '优惠减免金额',
  `balance` decimal(12,2) GENERATED ALWAYS AS (((`amount` - `received_amount`) - `discount_amount`)) STORED COMMENT '余额',
  `due_date` date DEFAULT NULL COMMENT '到期日期',
  `aging_days` int DEFAULT '0' COMMENT '逾期天数（由应用层计算更新）',
  `aging_category` varchar(20) DEFAULT NULL COMMENT '账龄分类：current=当期, within_30=30天内, within_60=60天内, within_90=90天内, overdue=逾期',
  `status` varchar(20) DEFAULT 'pending' COMMENT '状态：pending=待收款, partial=部分收款, paid=已收款, overdue=逾期',
  `is_locked` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否锁定：0=否, 1=是',
  `unlock_rule` varchar(100) DEFAULT NULL COMMENT '解锁规则，例如：delivery_completed',
  `source_type` varchar(20) DEFAULT '' COMMENT '款项来源：mold=模具款, production=生产款',
  `remark` text COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_customer_id` (`customer_id`) COMMENT '客户ID索引',
  KEY `idx_order_id` (`order_id`) COMMENT '订单ID索引',
  KEY `idx_due_date` (`due_date`) COMMENT '到期日期索引',
  KEY `idx_status` (`status`) COMMENT '状态索引',
  KEY `idx_aging_category` (`aging_category`) COMMENT '账龄分类索引',
  KEY `idx_payment_type` (`payment_type`) COMMENT '款项类型索引',
  CONSTRAINT `fk_ar_customer` FOREIGN KEY (`customer_id`) REFERENCES `customer` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应收账款表';

-- 7.2 应收付款记录表（receivable_payments）
DROP TABLE IF EXISTS receivable_payments;
CREATE TABLE receivable_payments (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '收款记录ID',
  `receivable_id` bigint NOT NULL COMMENT '应收账款ID（accounts_receivable.id）',
  `payment_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '收款金额（实际到账）',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '本次优惠金额',
  `payment_date` date NOT NULL COMMENT '收款日期',
  `payment_method` varchar(50) DEFAULT NULL COMMENT '收款方式：bank_transfer=银行转账, cash=现金, check=支票, wechat=微信, alipay=支付宝',
  `transaction_no` varchar(100) DEFAULT NULL COMMENT '交易流水号/回单号',
  `receiving_account` varchar(100) DEFAULT NULL COMMENT '收款账户',
  `remark` text COMMENT '备注',
  `receipt_file_id` varchar(500) DEFAULT NULL COMMENT '回单文件ID（多个ID用逗号分隔）',
  `invoice_file_id` varchar(500) DEFAULT NULL COMMENT '发票文件ID（多个ID用逗号分隔）',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_receivable_id` (`receivable_id`),
  KEY `idx_payment_date` (`payment_date`),
  KEY `idx_payment_method` (`payment_method`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收款记录表';

-- 7.3 应付账款表（accounts_payable）
DROP TABLE IF EXISTS accounts_payable;
CREATE TABLE accounts_payable (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '应付账款ID',
  `supplier_id` bigint NOT NULL COMMENT '供应商/加工商ID（supplier.id 或 processor.id）',
  `order_id` bigint DEFAULT NULL COMMENT '关联订单ID（outsourcing_record.id 等）',
  `order_type` varchar(50) DEFAULT NULL COMMENT '订单类型：outsourcing=外发订单, purchase=采购订单',
  `payment_type` varchar(50) NOT NULL DEFAULT '' COMMENT '款项类型：total=全款, first=首期款, final=尾款, installment=分期款',
  `source_type` varchar(20) NOT NULL DEFAULT '' COMMENT '款项来源：outsource=外发, purchase=采购',
  `order_code` varchar(50) NOT NULL DEFAULT '' COMMENT '关联订单编号（外发单号等）',
  `invoice_no` varchar(50) DEFAULT NULL COMMENT '发票号码',
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '应付金额',
  `paid_amount` decimal(12,2) DEFAULT '0.00' COMMENT '已付金额',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '优惠减免金额',
  `balance` decimal(12,2) GENERATED ALWAYS AS (((`amount` - `paid_amount`) - `discount_amount`)) STORED COMMENT '余额',
  `due_date` date DEFAULT NULL COMMENT '到期日期',
  `status` varchar(20) DEFAULT 'pending' COMMENT '状态：pending=待付款, partial=部分付款, paid=已付款, overdue=逾期',
  `remark` text COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_supplier_id` (`supplier_id`) COMMENT '供应商ID索引',
  KEY `idx_order_id` (`order_id`) COMMENT '订单ID索引',
  KEY `idx_due_date` (`due_date`) COMMENT '到期日期索引',
  KEY `idx_status` (`status`) COMMENT '状态索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应付账款表';

-- 7.4 付款记录表（payable_payments）
DROP TABLE IF EXISTS payable_payments;
CREATE TABLE payable_payments (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '付款记录ID',
  `payable_id` bigint NOT NULL COMMENT '应付账款ID（accounts_payable.id）',
  `payment_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '付款金额',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '本次优惠金额',
  `payment_date` date NOT NULL COMMENT '付款日期',
  `payment_method` varchar(50) DEFAULT NULL COMMENT '付款方式：bank_transfer=银行转账, cash=现金, check=支票, wechat=微信, alipay=支付宝',
  `transaction_no` varchar(100) DEFAULT NULL COMMENT '交易流水号/回单号',
  `payment_account` varchar(100) DEFAULT NULL COMMENT '付款账户',
  `receiver_account` varchar(100) DEFAULT NULL COMMENT '收款方账户',
  `remark` text COMMENT '备注',
  `receipt_file_id` varchar(500) DEFAULT NULL COMMENT '回单文件ID（多个ID用逗号分隔）',
  `invoice_file_id` varchar(500) DEFAULT NULL COMMENT '发票文件ID（多个ID用逗号分隔）',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_payable_id` (`payable_id`),
  KEY `idx_payment_date` (`payment_date`),
  KEY `idx_payment_method` (`payment_method`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='付款记录表';

-- 7.5 财务费用表（finance_expense）
DROP TABLE IF EXISTS finance_expense;
CREATE TABLE finance_expense (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '费用ID',
  `expense_type` varchar(50) NOT NULL COMMENT '费用类型：freight=运费, packaging=包装费, insurance=保险费, other=其他',
  `related_order_id` bigint DEFAULT NULL COMMENT '关联订单ID（mold_order.id 或 sale_order.id）',
  `related_order_type` varchar(50) DEFAULT NULL COMMENT '订单类型：mold_order=模具订单, sale_order=销售订单',
  `delivery_id` bigint DEFAULT NULL COMMENT '关联交付记录ID（mold_delivery.id）',
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '费用金额',
  `paid_amount` decimal(12,2) DEFAULT '0.00' COMMENT '已付金额',
  `payment_date` date DEFAULT NULL COMMENT '支付日期',
  `payer` varchar(100) DEFAULT NULL COMMENT '支付方',
  `receiver` varchar(100) DEFAULT NULL COMMENT '收款方',
  `remark` text COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_related_order` (`related_order_id`,`related_order_type`),
  KEY `idx_delivery_id` (`delivery_id`),
  KEY `idx_expense_type` (`expense_type`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='财务费用表';

-- 7.6 费用支付记录表（expense_payments）
DROP TABLE IF EXISTS expense_payments;
CREATE TABLE expense_payments (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '支付记录ID',
  `expense_id` bigint NOT NULL COMMENT '费用ID（finance_expense.id）',
  `payment_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '支付金额',
  `payment_date` date DEFAULT NULL COMMENT '支付日期',
  `payment_method` varchar(50) DEFAULT NULL COMMENT '支付方式：bank_transfer=银行转账, cash=现金, check=支票, wechat=微信, alipay=支付宝',
  `transaction_no` varchar(100) DEFAULT NULL COMMENT '交易流水号/回单号',
  `payment_account` varchar(100) DEFAULT NULL COMMENT '付款账户',
  `receiver_account` varchar(100) DEFAULT NULL COMMENT '收款方账户',
  `remark` text COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_expense_id` (`expense_id`),
  KEY `idx_payment_date` (`payment_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='费用支付记录表';

-- 7.7 收款账户表（receiving_accounts）
DROP TABLE IF EXISTS receiving_accounts;
CREATE TABLE receiving_accounts (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '账户ID',
  `account_name` varchar(100) NOT NULL COMMENT '账户名称（如：公司公账、法人银行卡）',
  `account_no` varchar(100) NOT NULL COMMENT '账号',
  `bank_name` varchar(200) DEFAULT NULL COMMENT '开户行',
  `account_type` varchar(50) DEFAULT NULL COMMENT '账户类型：company=对公, personal=个人, wechat=微信, alipay=支付宝',
  `sort` int DEFAULT '0' COMMENT '排序',
  `status` tinyint DEFAULT '1' COMMENT '状态：1=启用, 0=禁用',
  `remark` text COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_account_name_no` (`account_name`,`account_no`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_status` (`status`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收款账户表';

-- ============================================
-- 八、视图和索引优化
-- ============================================

-- 创建模具状态统计视图
CREATE OR REPLACE VIEW v_mold_status_stat AS
SELECT 
  process_status,
  COUNT(*) as mold_count
FROM mold
GROUP BY process_status;

-- 创建工单进度视图
CREATE OR REPLACE VIEW v_work_order_progress AS
SELECT 
  wo.id,
  wo.work_code,
  wo.mold_order_id,
  mo.order_code as order_code,
  wo.mold_id,
  m.mold_code,
  wo.order_type,
  wo.status,
  wo.progress,
  COUNT(DISTINCT wop.id) as total_process,
  SUM(CASE WHEN wop.is_completed = 1 THEN 1 ELSE 0 END) as completed_process
FROM mold_work_order wo
LEFT JOIN mold_order mo ON wo.mold_order_id = mo.id
LEFT JOIN mold m ON wo.mold_id = m.id
LEFT JOIN work_order_process wop ON wo.id = wop.work_order_id

GROUP BY wo.id, wo.work_code, wo.mold_order_id, mo.order_code, wo.mold_id, m.mold_code, wo.order_type, wo.status, wo.progress;

-- 创建模具全貌视图（包含订单和工单统计）
CREATE OR REPLACE VIEW v_mold_full_view AS
SELECT 
  m.id,
  m.mold_code,
  m.mold_name,
  m.customer_id,
  c.cus_name as customer_name,
  m.process_status,
  m.mold_type,
  m.mold_process,
  m.material_type,
  m.latest_order_id,
  m.latest_work_order_id,
  m.total_orders,
  m.total_work_orders,
  m.last_delivery_time,
  -- 最新订单信息
  lo.order_code as latest_order_code,
  lo.status as latest_order_status,
  lo.plan_end_date as latest_order_deadline,
  -- 最新工单信息
  lwo.work_code as latest_work_order_code,
  lwo.status as latest_work_order_status,
  lwo.progress as latest_work_order_progress,
  lwo.actual_start as work_actual_start,
  lwo.actual_end as work_actual_end,
  -- 统计信息
  (SELECT COUNT(*) FROM mold_order WHERE mold_id = m.id) as actual_total_orders,
  (SELECT COUNT(*) FROM mold_work_order WHERE mold_id = m.id) as actual_total_work_orders,
  (SELECT MAX(deliver_time) FROM mold_order WHERE mold_id = m.id AND status = 'DELIVERED') as actual_last_delivery
FROM mold m
LEFT JOIN customer c ON m.customer_id = c.id
LEFT JOIN mold_order lo ON m.latest_order_id = lo.id
LEFT JOIN mold_work_order lwo ON m.latest_work_order_id = lwo.id;

-- ============================================
-- 九、新增复合索引（高频查询组合）
-- ============================================

-- 模具订单：按客户查询订单状态（高频）
CREATE INDEX idx_mo_customer_status ON mold_order (customer_id, status);
-- 模具工单：按模具查工单状态
CREATE INDEX idx_mwo_mold_status ON mold_work_order (mold_id, status);
-- 生产订单明细：按订单查明细状态
CREATE INDEX idx_poi_order_status ON production_order_item (order_id, status);
-- 应收账款：按客户查应收状态（财务模块高频）
CREATE INDEX idx_ar_customer_status ON accounts_receivable (customer_id, status);
-- 应付账款：按供应商查应付状态
CREATE INDEX idx_ap_supplier_status ON accounts_payable (supplier_id, status);
-- 生产订单：按计划结束日期排序查询
CREATE INDEX idx_po_plan_end_date ON production_order (plan_end_date);
-- 操作日志：按业务类型+ID 查询
CREATE INDEX idx_oper_log_business ON sys_oper_log (business_type, business_id);

-- ============================================
-- 十、初始化基础数据
-- ============================================

-- 7.1 插入默认系统用户（密码：admin，使用 bcrypt + Pepper 加密）
-- ⚠️ 部署前请务必修改 config.yaml 中的 security.pepper 值，并重新生成下方密码哈希
INSERT INTO sys_user (username, password, real_name, status) VALUES
('admin', '$2b$10$tUw1hx54nBTTMfDDs6xdAO5kvnbtw2TyX1KMyPgG6vcwcJyXB7Nqq', '超级管理员', 1),
('user', '$2b$10$tUw1hx54nBTTMfDDs6xdAO5kvnbtw2TyX1KMyPgG6vcwcJyXB7Nqq', '普通用户', 1);

-- 7.2 插入默认角色
INSERT INTO sys_role (role_name, remark) VALUES
('超级管理员', '拥有所有权限'),
('生产经理', '负责生产管理'),
('生产人员', '负责生产报工');

-- 7.2.1 插入用户角色关联数据
INSERT INTO sys_user_role (user_id, role_id)
SELECT u.id, r.id
FROM sys_user u, sys_role r
WHERE u.username = 'admin' AND r.role_name = '超级管理员';

INSERT INTO sys_user_role (user_id, role_id)
SELECT u.id, r.id
FROM sys_user u, sys_role r
WHERE u.username = 'user' AND r.role_name = '生产人员';

-- 7.3 插入默认部门
INSERT INTO sys_dept (dept_name, parent_id, sort) VALUES
('总公司', 0, 0),
('生产部', 1, 1),
('技术部', 1, 2),
('销售部', 1, 3),
('仓库', 1, 4);

-- 7.4 插入字典类型
INSERT INTO sys_dict_type (dict_name, dict_code, remark) VALUES
('模具状态管理', 'mold-status', '模具的各种状态'),
('模具工单状态', 'work-order-status', '模具工单的状态管理'),
('模具订单状态', 'mold-order-status', '模具订单的状态管理'),
('模具工艺管理', 'mold-process', '模具加工工艺类型'),
('材质类型管理', 'material-type', '材质分类管理'),
('模胚情况管理', 'mold-type', '模胚新旧情况管理'),
('字典类型管理', 'dict-type', '系统字典类型管理'),
('款项类型', 'payment-type', '模具订单款项类型（首款、中期款、尾款等）'),
('试模方式', 'test-mode', '试模方式（自家试模、外协试模、免试）'),
('订单优先级', 'order-priority', '订单优先级管理（普通、重要、紧急、非常紧急）'),
('生产类型', 'production-type', '生产类型管理（本厂生产、外发生产）'),
('模具来源', 'mold-source', '模具来源管理（本厂制造模具、客户提供模具）'),
('生产设备', 'production-equipment', '生产设备管理（960吨机、600吨机、300吨机）'),
('负责员工', 'responsible-employee', '负责员工管理'),
('加工商管理', 'processor', '加工商信息管理'),
('供应商管理', 'supplier', '供应商信息管理'),
('生产订单状态', 'production-order-status', '生产订单的状态管理（待排产、生产中、已完成等）'),
('订单/工单类型', 'mold_order_type', '模具工单和模具订单的订单类型（新模、修模、改造）'),
('塑胶原料', 'plastic-material', '常见塑胶原料字典（PP、PE、TPE、TPU、尼龙等）'),
('班次管理', 'shift-type', '报工班次管理（白班、夜班等）'),
('模具位置', 'mold-location', '模具的当前位置'),
('检验类型', 'inspection-type', '质量检验类型（首件检验、过程检验、成品检验）'),
('检验结果', 'inspection-result', '质量检验结果（合格、不合格、让步接收）'),
('综合订单状态', 'master-order-status', '综合订单的状态管理（处理中、已完成等）'),
('综合订单类型', 'master-order-type', '综合订单的类型（模具+产品、仅模具、仅产品）'),
('依赖状态', 'dependency-status', '生产订单对模具订单的依赖状态'),
('外发状态', 'outsourcing-status', '模具外发和生产外发的状态字典'),
('应收账款状态', 'receivable-status', '应收账款的收款状态'),
('应付账款状态', 'payable-status', '应付账款的付款状态'),
('账龄分类', 'aging-category', '应收账款的账龄区间分类'),
('费用类型', 'expense-type', '财务费用类型字典'),
('收款付款方式', 'payment-method', '收款和付款的方式类型');

-- 7.5 插入字典数据

-- 模具状态 (mold-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, allow_production, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '设计中', 'DESIGNING', 'cyan-4', 1, 2, 0, '模具设计阶段'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '加工中', 'PROCESSING', 'orange-4', 2, 2, 0, '模具正在加工'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '试模', 'TESTING', 'gray-7', 3, 2, 0, '试模阶段'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '生产中', 'PRODUCING', 'primary', 4, 2, 0, '模具投入生产中'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '维修', 'REPAIRING', 'warning', 5, 2, 0, '模具维修中'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '仓库库存', 'IN_STOCK', 'gray-14', 6, 2, 1, '在库状态'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '已交付客户', 'DELIVERED', 'purple-4', 7, 2, 1, '已交付给客户'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '报废', 'SCRAPPED', 'danger', 8, 2, 0, '模具报废'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-status'), '已试模', 'TESTED', 'green-4', 9, 2, 0, '模具试模完成，等待后续处理');

-- 模具工单状态 (work-order-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'work-order-status'), '未开始', 'NOT_STARTED', 'gray-7', 1, 2, '工单尚未开始'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'work-order-status'), '加工中', 'PROCESSING', 'orange-5', 2, 2, '工单正在加工'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'work-order-status'), '已完成', 'COMPLETED', 'success', 3, 2, '所有工序完成'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'work-order-status'), '已试模', 'TESTED', 'cyan-7', 4, 2, '已试模'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'work-order-status'), '已入库', 'STOCKED_IN', 'gray-11', 5, 2, '已入库'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'work-order-status'), '已交付', 'DELIVERED', 'purple-7', 6, 2, '已交付客户'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'work-order-status'), '已取消', 'CANCELLED', 'danger', 7, 2, '工单已取消');

-- 模具订单状态 (mold-order-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-order-status'), '未开始', 'NOT_STARTED', 'gray-7', 1, 2, '订单尚未开始'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-order-status'), '加工中', 'PROCESSING', 'orange-5', 2, 2, '订单正在加工'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-order-status'), '待试模', 'PENDING_TEST', 'blue-8', 3, 2, '等待试模'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-order-status'), '已试模', 'TESTED', 'cyan-7', 4, 2, '已试模'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-order-status'), '已完成', 'COMPLETED', 'success', 5, 2, '订单已完成'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-order-status'), '已取消', 'CANCELLED', 'danger', 6, 2, '订单已取消'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-order-status'), '已交付', 'DELIVERED', 'purple-7', 7, 2, '已交付客户');

-- 模具工艺 (mold-process)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-process'), '普通模具', 'NORMAL', 1, '普通工艺'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-process'), '热流道', 'HOT_RUNNER', 2, '热流道工艺');

-- 材质类型 (material-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'material-type'), '钢材', 'STEEL', 1, '普通钢材'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'material-type'), '加硬钢材', 'HARDENED_STEEL', 2, '加硬处理钢材');

-- 模胚情况 (mold-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-type'), '全新', 'NEW', 1, '全新模胚'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-type'), '二手', 'USED', 2, '二手模胚');

-- 款项类型 (payment-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'payment-type'), '全款', 'TOTAL', 0, '一次性收/付清全款'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'payment-type'), '首款', 'first-payment', 1, ''),
((SELECT id FROM sys_dict_type WHERE dict_code = 'payment-type'), '中期款', 'middle-payment', 2, ''),
((SELECT id FROM sys_dict_type WHERE dict_code = 'payment-type'), '尾款', 'final-payment', 3, '');

-- 试模方式 (test-mode)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'test-mode'), '本厂试模', 'SELF_TEST', 1, '在自家工厂试模'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'test-mode'), '外发试模', 'EXTERNAL_TEST', 2, '发出去给其他工厂试模'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'test-mode'), '免试模', 'EXEMPTED', 3, '免试直接进入下一步');

-- 订单优先级 (order-priority)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'order-priority'), '普通', 'NORMAL', 'default', 1, '普通优先级'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'order-priority'), '重要', 'IMPORTANT', 'warning', 2, '重要优先级'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'order-priority'), '紧急', 'URGENT', 'danger', 3, '紧急优先级'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'order-priority'), '非常紧急', 'VERY_URGENT', 'danger', 4, '非常紧急优先级');

-- 生产类型 (production-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-type'), '本厂生产', 'IN_HOUSE', 'success', 1, '在本厂进行生产'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-type'), '外发生产', 'OUTSOURCED', 'primary', 2, '外发给其他工厂生产');

-- 模具来源 (mold-source)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-source'), '本厂制造模具', 'SELF_MADE', 'success', 1, '本厂自行制造的模具'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-source'), '客户提供模具', 'CUSTOMER_PROVIDED', 'primary', 2, '客户提供的模具');

-- 生产设备 (production-equipment)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-equipment'), '960吨机', '960-ton', 'primary', 1, '960吨注塑机'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-equipment'), '600吨机', '600-ton', 'warning', 2, '600吨注塑机'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-equipment'), '300吨机', '300-ton', 'success', 3, '300吨注塑机');

-- 负责员工 (responsible-employee)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'responsible-employee'), '张三', 'zhangsan', 'primary', 1, '员工张三'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'responsible-employee'), '王五', 'wangwu', 'success', 2, '员工王五');

-- 生产订单状态 (production-order-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, allow_work_report, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status'), '待排产', 'PENDING_SCHEDULE', 'gray-7', 1, 0, 1, '订单已创建，等待排产'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status'), '生产中', 'PRODUCING', 'orange-5', 2, 0, 1, '正在生产中'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status'), '已入库', 'STOCKED_IN', 'success', 3, 0, 0, '产品已入库'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status'), '已出货', 'SHIPPED', 'blue-8', 4, 0, 0, '货已交付客户'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status'), '已完成', 'COMPLETED', 'primary', 5, 2, 0, '订单全部完成'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status'), '已取消', 'CANCELLED', 'danger', 6, 2, 0, '订单已取消');

-- 订单/工单类型 (mold_order_type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold_order_type'), '新模', 'NEW_MOLD', 'primary', 1, 1, '新模具工单/订单'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold_order_type'), '修模', 'REPAIR', 'warning', 2, 1, '修模工单/订单'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold_order_type'), '改造', 'MODIFICATION', 'success', 3, 1, '模具改造工单/订单');

-- 塑胶原料 (plastic-material)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PP（聚丙烯）', 'PP', 'primary', 1, 1, '聚丙烯，具有良好的耐化学性和耐热性'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PE（聚乙烯）', 'PE', 'success', 2, 1, '聚乙烯，分为高密度HDPE和低密度LDPE'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'ABS', 'ABS', 'warning', 3, 1, '丙烯腈-丁二烯-苯乙烯共聚物，具有良好的韧性和加工性'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PS（聚苯乙烯）', 'PS', 'danger', 4, 1, '聚苯乙烯，透明度高，易加工'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PC（聚碳酸酯）', 'PC', 'cyan-7', 5, 1, '聚碳酸酯，具有优异的抗冲击性和透明度'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PVC（聚氯乙烯）', 'PVC', 'purple-7', 6, 1, '聚氯乙烯，分为硬质和软质两种'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PMMA（亚克力）', 'PMMA', 'blue-8', 7, 1, '聚甲基丙烯酸甲酯，透明度高，耐候性好'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PA（尼龙）', 'PA', 'orange-5', 8, 1, '聚酰胺，具有优异的耐磨性和机械强度'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'POM（聚甲醛）', 'POM', 'gray-7', 9, 1, '聚甲醛，具有高硬度和良好的耐磨性'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'PET（聚酯）', 'PET', 'gray-9', 10, 1, '聚对苯二甲酸乙二醇酯，常用于包装和纤维'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'TPE（热塑性弹性体）', 'TPE', 'gray-11', 11, 1, '热塑性弹性体，兼具橡胶和塑料特性'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'plastic-material'), 'TPU（热塑性聚氨酯）', 'TPU', 'gray-14', 12, 1, '热塑性聚氨酯，具有优异的耐磨性和弹性');

-- 班次管理 (shift-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'shift-type'), '白班', 'DAY_SHIFT', 'warning', 1, 1, '白班（通常为白天工作时间）'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'shift-type'), '夜班', 'NIGHT_SHIFT', 'primary', 2, 1, '夜班（通常为夜间工作时间）');

-- 模具位置 (mold-location)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, allow_production, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-location'), '厂内', 'IN_FACTORY', 'success', 1, 2, 1, '模具在厂内'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-location'), '外发中', 'OUTSOURCING', 'warning', 2, 2, 0, '模具外发加工中'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'mold-location'), '客户处', 'AT_CUSTOMER', 'primary', 3, 2, 0, '模具在客户处');

-- 检验类型 (inspection-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'inspection-type'), '首件检验', 'FIRST_ARTICLE', 'blue-7', 1, 2, '生产开始时的首件产品检验'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'inspection-type'), '过程检验', 'IN_PROCESS', 'orange-7', 2, 2, '生产过程中的巡检'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'inspection-type'), '成品检验', 'FINAL', 'green-7', 3, 2, '生产完成后的成品检验');

-- 检验结果 (inspection-result)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'inspection-result'), '合格', 'PASS', 'success', 1, 2, '检验合格，可以入库'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'inspection-result'), '不合格', 'FAIL', 'danger', 2, 2, '检验不合格，需要返工或报废'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'inspection-result'), '让步接收', 'CONCESSION', 'warning', 3, 2, '虽然不合格但可以接受');

-- 综合订单状态 (master-order-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-status'), '待开始', 'PENDING', 'primary', 1, 0, '订单待开始'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-status'), '模具制造中', 'MOLD_PRODUCING', 'warning', 2, 0, '模具制造中'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-status'), '生产中', 'PRODUCING', 'processing', 3, 0, '产品生产中'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-status'), '待发货', 'PENDING_SHIP', 'cyan-7', 4, 0, '等待发货'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-status'), '已完成', 'COMPLETED', 'success', 5, 0, '订单已完成'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-status'), '已取消', 'CANCELLED', 'danger', 6, 0, '订单已取消');

-- 综合订单类型 (master-order-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-type'), '模具+产品', 'COMBINED', 'primary', 1, 0, '同时包含模具和生产'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-type'), '仅模具', 'MOLD_ONLY', 'warning', 2, 0, '只包含模具订单'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'master-order-type'), '仅产品', 'PROD_ONLY', 'success', 3, 0, '只包含生产订单');

-- 依赖状态 (dependency-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'dependency-status'), '等待模具', 'WAITING_MOLD', 'warning', 1, 0, '等待模具订单完成'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'dependency-status'), '无依赖', 'NONE', 'success', 2, 0, '无依赖关系'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'dependency-status'), '就绪', 'READY', 'success', 3, 0, '依赖已完成');

-- 外发状态 (outsourcing-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'outsourcing-status'), '待发出', 'PENDING', 'warning', 1, 1, '外发记录已创建，等待发出'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'outsourcing-status'), '已发出', 'SENT', 'primary', 2, 1, '外发记录已发出，等待加工'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'outsourcing-status'), '生产中', 'PRODUCING', 'success', 3, 1, '加工商正在生产中'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'outsourcing-status'), '已返回', 'RETURNED', 'cyan-7', 4, 1, '外发物品已返回工厂'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'outsourcing-status'), '已撤回', 'CANCELLED', 'danger', 5, 1, '外发记录已撤回/取消');

-- 发货方式 (delivery_method)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'delivery_method'), '快递', 'EXPRESS', 'primary', 1, 1, 0, '快递发货'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'delivery_method'), '货拉拉', 'LALAMOVE', 'warning', 2, 1, 0, '货拉拉配送'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'delivery_method'), '自提', 'PICKUP', 'success', 3, 1, 0, '客户自提'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'delivery_method'), '物流配送', 'LOGISTICS', 'default', 4, 1, 0, '物流配送');

-- 快递公司 (express_company)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, is_locked, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '顺丰速运', 'SF', 'danger', 1, 1, 0, '顺丰'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '圆通速递', 'YT', 'warning', 2, 1, 0, '圆通'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '中通快递', 'ZT', 'success', 3, 1, 0, '中通'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '申通快递', 'ST', 'primary', 4, 1, 0, '申通'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '韵达快递', 'YD', 'cyan-7', 5, 1, 0, '韵达'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '邮政快递', 'YZ', 'purple-7', 6, 1, 0, '邮政'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '极兔速递', 'JT', 'orange-7', 7, 1, 0, '极兔速递(J&T Express)'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '德邦快递', 'DB', 'arcoblue-7', 8, 1, 0, '德邦快递'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '百世快递', 'BS', 'cyan-9', 9, 1, 0, '百世快递(BEST Express)'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '京东物流', 'JD', 'red-7', 10, 1, 0, '京东物流'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'express_company'), '天天快递', 'TT', 'green-7', 11, 1, 0, '天天快递(TTKDEX)');

-- 出库类型 (out_type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'out_type'), '发货', 'normal', 'warning', 1, 1, '正常发货出库'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'out_type'), '其他', 'other', 'default', 2, 1, '其他原因出库');

-- 出库原因 (out_reason)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, color, sort, status, remark) VALUES
((SELECT id FROM sys_dict_type WHERE dict_code = 'out_reason'), '过期丢弃', 'EXPIRED', 'danger', 1, 1, '产品过期需要丢弃'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'out_reason'), '损坏', 'DAMAGED', 'warning', 2, 1, '产品损坏无法使用'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'out_reason'), '不达标', 'UNQUALIFIED', 'danger', 3, 1, '产品质量不达标'),
((SELECT id FROM sys_dict_type WHERE dict_code = 'out_reason'), '其他原因', 'OTHER', 'default', 4, 1, '其他原因');

-- 应收账款状态 (receivable-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, status, remark)
SELECT dt.id, item.dict_label, item.dict_value, item.sort, 1, item.remark
FROM sys_dict_type dt
CROSS JOIN (
  SELECT '待收款' AS dict_label, 'pending' AS dict_value, 1 AS sort, '' AS remark
  UNION ALL SELECT '部分收款', 'partial', 2, ''
  UNION ALL SELECT '已收款', 'paid', 3, ''
  UNION ALL SELECT '逾期', 'overdue', 4, ''
) item
WHERE dt.dict_code = 'receivable-status'
ON DUPLICATE KEY UPDATE dict_label=VALUES(dict_label);

-- 应付账款状态 (payable-status)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, status, remark)
SELECT dt.id, item.dict_label, item.dict_value, item.sort, 1, item.remark
FROM sys_dict_type dt
CROSS JOIN (
  SELECT '待付款' AS dict_label, 'pending' AS dict_value, 1 AS sort, '' AS remark
  UNION ALL SELECT '部分付款', 'partial', 3, ''
  UNION ALL SELECT '已付款', 'paid', 4, ''
  UNION ALL SELECT '逾期', 'overdue', 5, ''
) item
WHERE dt.dict_code = 'payable-status'
ON DUPLICATE KEY UPDATE dict_label=VALUES(dict_label);

-- 账龄分类 (aging-category)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, status, remark)
SELECT dt.id, item.dict_label, item.dict_value, item.sort, 1, item.remark
FROM sys_dict_type dt
CROSS JOIN (
  SELECT '当期' AS dict_label, 'current' AS dict_value, 1 AS sort, '' AS remark
  UNION ALL SELECT '30天内', 'within_30', 2, ''
  UNION ALL SELECT '60天内', 'within_60', 3, ''
  UNION ALL SELECT '90天内', 'within_90', 4, ''
  UNION ALL SELECT '逾期', 'overdue', 5, ''
) item
WHERE dt.dict_code = 'aging-category'
ON DUPLICATE KEY UPDATE dict_label=VALUES(dict_label);

-- 费用类型 (expense-type)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, status, remark)
SELECT dt.id, item.dict_label, item.dict_value, item.sort, 1, item.remark
FROM sys_dict_type dt
CROSS JOIN (
  SELECT '运费' AS dict_label, 'freight' AS dict_value, 1 AS sort, '' AS remark
  UNION ALL SELECT '包装费', 'packaging', 2, ''
  UNION ALL SELECT '保险费', 'insurance', 3, ''
  UNION ALL SELECT '其他', 'other', 4, ''
) item
WHERE dt.dict_code = 'expense-type'
ON DUPLICATE KEY UPDATE dict_label=VALUES(dict_label);

-- 收款/付款方式 (payment-method)
INSERT INTO sys_dict_data (dict_type_id, dict_label, dict_value, sort, status, remark)
SELECT dt.id, item.dict_label, item.dict_value, item.sort, 1, item.remark
FROM sys_dict_type dt
CROSS JOIN (
  SELECT '银行转账' AS dict_label, 'bank_transfer' AS dict_value, 1 AS sort, '' AS remark
  UNION ALL SELECT '现金', 'cash', 2, ''
  UNION ALL SELECT '支票', 'check', 3, ''
  UNION ALL SELECT '微信', 'wechat', 4, ''
  UNION ALL SELECT '支付宝', 'alipay', 5, ''
) item
WHERE dt.dict_code = 'payment-method'
ON DUPLICATE KEY UPDATE dict_label=VALUES(dict_label);

-- 配置 payment-method 字典 ext 扩展字段（映射各付款方式对应的可用账户类型）
UPDATE sys_dict_data
SET ext = CASE dict_value
  WHEN 'bank_transfer' THEN '{"account_types":["company","personal"]}'
  WHEN 'cash' THEN '{"account_types":[]}'
  WHEN 'check' THEN '{"account_types":["company","personal"]}'
  WHEN 'wechat' THEN '{"account_types":["wechat"]}'
  WHEN 'alipay' THEN '{"account_types":["alipay"]}'
  ELSE NULL
END
WHERE dict_type_id = (SELECT id FROM sys_dict_type WHERE dict_code = 'payment-method');

-- 7.6 插入客户数据（5 个）
INSERT INTO customer (cus_code, cus_name, contact, phone, email, address) VALUES
('恩齐', '客户 A', '张三', '13800138001', 'zhanga@example.com', '北京市朝阳区'),
('CUST002', '锦杨科技', '李四', '13800138002', 'lisi@example.com', '上海市浦东新区'),
('CUST003', '客户 C', '王五', '13800138003', 'wangwu@example.com', '广州市天河区'),
('CUST004', '客户 D', '赵六', '13800138004', 'zhaoliu@example.com', '深圳市南山区'),
('CUST005', '客户 E', '钱七', '13800138005', 'qianqi@example.com', '杭州市西湖区');

-- 7.7 插入供应商数据（2 个）
INSERT INTO supplier (sup_code, sup_name, contact, phone, email, address) VALUES
('SUP001', '供应商 A', '供应联系人 1', '13900139001', 'supply1@example.com', '江苏省苏州市'),
('SUP002', '供应商 B', '供应联系人 2', '13900139002', 'supply2@example.com', '浙江省宁波市');

-- 7.8 插入仓库数据（2 个）
INSERT INTO warehouse (wh_code, wh_name, address, manager, phone) VALUES
('WH001', '成品仓库', '厂区 A 栋', '张三', '010-88888888'),
('WH002', '原料仓库', '厂区 B 栋', '李四', '010-88888889');

-- 7.9 插入默认工序模板
INSERT INTO process (process_code, process_name, process_type, work_hour) VALUES
('PROC001', '设计', 1, 8.00),
('PROC002', 'CNC 加工', 1, 4.00),
('PROC003', '线切割', 1, 3.00),
('PROC004', '磨床加工', 1, 2.00),
('PROC005', '放电加工', 1, 3.00),
('PROC006', '抛光', 1, 2.00),
('PROC007', '装配', 1, 4.00),
('PROC008', '试模', 1, 2.00),
('PROC009', '注塑', 2, 1.00),
('PROC010', '包装', 2, 0.50);

-- 7.10 插入模具数据（5 套）
INSERT INTO mold (mold_code, mold_name, customer_id, production_date, process_status, life_cycle_status, mold_type, mold_process, material_type, remark) VALUES
('MOLD001', NULL, 1, '2026-01-15', 'PROCESSING', 'DESIGNING', 'NEW', 'NORMAL', 'STEEL', '模具 A - 设计中'),
('MOLD002', NULL, 2, '2026-02-01', 'PRODUCING', 'MASS_PRODUCTION', 'NEW', 'HOT_RUNNER', 'HARDENED_STEEL', '模具 B-加工中'),
('MOLD003', NULL, 1, '2026-02-15', 'PRODUCING', 'MASS_PRODUCTION', 'USED', 'NORMAL', 'STEEL', '模具 C-量产'),
('MOLD004', NULL, 3, '2026-03-01', 'IN_STOCK', 'IN_STOCK', 'NEW', 'NORMAL', 'STEEL', '模具 D-仓库库存'),
('MOLD009', '拍痧板', 4, '2026-03-15', 'IN_STOCK', 'DELIVERED', 'NEW', 'HOT_RUNNER', 'HARDENED_STEEL', '模具 E-已交付客户');

-- 7.11 插入产品数据（3 个）
INSERT INTO product (pro_code, pro_name, spec, mold_id, unit, price, status) VALUES
('PROD001', '产品 A', '规格 A-标准型', 1, '个', 150.00, 1),
('PROD002', '产品 B', '规格 B-增强型', 2, '个', 280.00, 1),
('PROD003', '产品 C', '规格 C-专业型', 3, '个', 350.00, 1);

-- 7.12 插入模具订单数据（2 个）
INSERT INTO mold_order (order_code, customer_id, order_type, mold_id, plan_start_date, plan_end_date, total_amount, status) VALUES
('MO001', 1, 'NEW_MOLD', 1, '2026-01-20', '2026-03-20', 50000.00, 'PROCESSING'),
('MO002', 2, 'NEW_MOLD', 2, '2026-02-05', '2026-04-05', 80000.00, 'PROCESSING');

-- 7.13 插入默认物料
INSERT INTO material (mat_code, mat_name, spec, unit, mat_type, price, stock_warn) VALUES
('MAT001', '钢材 A', '规格 A', 'kg', 1, 50.00, 100),
('MAT002', '钢材 B', '规格 B', 'kg', 1, 80.00, 100),
('MAT003', '模具配件 A', '型号 A', '个', 2, 200.00, 50),
('MAT004', '原料 A', '工业级', 'kg', 1, 30.00, 500);

-- 7.14 插入工序模板分组数据
INSERT INTO process_template (group_id, group_name, process_name, sort, status, remark) VALUES
(2, '加工阶段', 'CNC加工', 1, 1, 'CNC数控加工工序'),
(2, '加工阶段', '磨床', 2, 1, '磨床加工工序'),
(2, '加工阶段', '放电加工', 3, 1, '电火花加工工序'),
(2, '加工阶段', '线切割', 4, 1, '线切割加工工序'),
(2, '加工阶段', '钳工', 5, 1, '钳工装配工序'),
(3, '后处理阶段', '抛光', 6, 1, '抛光处理工序'),
(4, '装配阶段', '装配', 7, 1, '模具装配工序'),
(4, '装配阶段', '检验', 8, 2, '质量检验工序'),
(1, '准备阶段', '定料', 1, 1, ''),
(4, '装配阶段', '合模', 9, 1, '');

-- 7.15 插入状态流转配置数据（新版-语义化字符串模型）
-- 工单 → 模具状态同步
INSERT INTO status_flow_config (work_order_status, mold_status, description, priority, status_order) VALUES
('NOT_STARTED', 'PROCESSING', '工单未开始 → 模具加工中', 1, 1),
('PROCESSING', 'PROCESSING', '工单加工中 → 模具加工中', 1, 2),
('COMPLETED', 'TESTING', '工单全部工序完成 → 模具试模', 1, 3),
('TESTED', 'TESTED', '已试模 → 模具已试模', 1, 4),
('STOCKED_IN', 'IN_STOCK', '已入库 → 仓库库存', 1, 5),
('DELIVERED', 'DELIVERED', '已交付 → 已交付客户', 1, 6);

-- 工单 → 订单状态同步
INSERT INTO status_flow_config (work_order_status, mold_order_status, description, priority, status_order) VALUES
('NOT_STARTED', 'NOT_STARTED', '工单未开始 → 订单未开始', 1, 1),
('PROCESSING', 'PROCESSING', '工单加工中 → 订单加工中', 1, 2),
('COMPLETED', 'PENDING_TEST', '工单全部工序完成 → 订单待试模', 1, 3),
('TESTED', 'TESTED', '已试模 → 订单已试模', 1, 4),
('STOCKED_IN', 'DELIVERED', '已入库 → 订单已交付', 1, 5),
('DELIVERED', 'DELIVERED', '已交付 → 订单已交付', 1, 6);

-- 订单 → 模具状态同步
INSERT INTO status_flow_config (mold_order_status, mold_status, description, priority, status_order) VALUES
('NOT_STARTED', 'PROCESSING', '订单未开始 → 模具加工中', 1, 1),
('PROCESSING', 'PROCESSING', '订单加工中 → 模具加工中', 1, 2),
('PENDING_TEST', 'TESTING', '订单待试模 → 模具试模', 1, 3),
('TESTED', 'TESTED', '订单已试模 → 模具已试模', 1, 4),
('COMPLETED', 'IN_STOCK', '订单已完成 → 仓库库存', 1, 5),
('DELIVERED', 'DELIVERED', '订单已交付 → 模具已交付客户', 1, 6);

-- 7.16 创建生产订单状态流转规则表
DROP TABLE IF EXISTS production_status_flow;
CREATE TABLE production_status_flow (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '主键ID',
    from_status_value VARCHAR(50) NOT NULL COMMENT '起始状态值（语义化字符串）',
    to_status_value VARCHAR(50) NOT NULL COMMENT '目标状态值（语义化字符串）',
    action_name VARCHAR(50) NOT NULL COMMENT '操作名称（如：入库、出货、完成）',
    action_code VARCHAR(50) NOT NULL COMMENT '操作编码（前端按钮标识）',
    is_enabled TINYINT DEFAULT 1 COMMENT '是否启用：1=启用，0=禁用',
    sort INT DEFAULT 0 COMMENT '排序',
    remark VARCHAR(255) COMMENT '备注',
    create_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    update_time DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    UNIQUE KEY uk_from_to (from_status_value, to_status_value),
    KEY idx_action_code (action_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='生产订单状态流转规则表';

-- 7.17 插入生产订单状态流转规则数据
INSERT INTO production_status_flow 
(from_status_value, to_status_value, action_name, action_code, is_enabled, sort, remark)
VALUES
('PENDING_SCHEDULE', 'PRODUCING', '开始生产', 'start_production', 1, 1, '待排产 → 生产中'),
('SCHEDULED',        'PRODUCING', '开始生产', 'start_production', 1, 2, '已排产 → 生产中'),
('PRODUCING',        'STOCKED_IN', '结束生产', 'complete_production', 1, 5, '生产中 → 已入库'),
('STOCKED_IN',       'SHIPPED',   '出货',     'shipment',         1, 6, '已入库 → 已出货'),
('SHIPPED',          'COMPLETED', '完成订单', 'finish_order',     1, 7, '已出货 → 已完成'),
('PENDING_SCHEDULE', 'CANCELLED', '取消订单', 'cancel_order',     1, 8, '待排产 → 已取消'),
('SCHEDULED',        'CANCELLED', '取消订单', 'cancel_order',     1, 9, '已排产 → 已取消'),
('PRODUCING',        'CANCELLED', '取消订单', 'cancel_order',     1, 10, '生产中 → 已取消'),
('STOCKED_IN',       'COMPLETED', '完成订单', 'finish_order',     1, 0,  NULL),
('STOCKED_IN',       'STOCKED_IN', '继续入库', 'stock_in',         1, 0,  NULL);

-- 7.18 插入收款账户初始数据
INSERT IGNORE INTO receiving_accounts (account_name, account_no, bank_name, account_type, sort, status, create_time, update_time) VALUES
('公司公账', '449299282882828', '中国工商银行深圳分行', 'company', 1, 1, NOW(), NOW()),
('法人银行卡', '6222024000123456', '中国工商银行', 'personal', 2, 1, NOW(), NOW()),
('财务银行卡', '6222024000654321', '中国建设银行', 'personal', 3, 1, NOW(), NOW()),
('微信收款', '13800138000', NULL, 'wechat', 4, 1, NOW(), NOW()),
('支付宝收款', 'company@example.com', NULL, 'alipay', 5, 1, NOW(), NOW());

-- ============================================
-- 十一、验证数据
-- ============================================

SELECT '========================================' as '';
SELECT '数据库初始化完成！' as info;
SELECT '========================================' as '';
SELECT '基础数据统计：' as info;

SELECT '系统用户数：' as item, COUNT(*) as count FROM sys_user
UNION ALL
SELECT '客户数：', COUNT(*) FROM customer
UNION ALL
SELECT '供应商数：', COUNT(*) FROM supplier
UNION ALL
SELECT '仓库数：', COUNT(*) FROM warehouse
UNION ALL
SELECT '模具数：', COUNT(*) FROM mold
UNION ALL
SELECT '产品数：', COUNT(*) FROM product
UNION ALL
SELECT '模具订单数：', COUNT(*) FROM mold_order
UNION ALL
SELECT '工序模板数：', COUNT(*) FROM process
UNION ALL
SELECT '应收账款数：', COUNT(*) FROM accounts_receivable
UNION ALL
SELECT '应付账款数：', COUNT(*) FROM accounts_payable
UNION ALL
SELECT '收款账户数：', COUNT(*) FROM receiving_accounts;

SELECT '========================================' as '';
SELECT '默认账号：' as info;
SELECT '用户名：admin / 密码：admin' as account;
SELECT '用户名：user / 密码：user' as account;
SELECT '========================================' as '';

-- 1.19 系统信息配置表（sys_system_info）
DROP TABLE IF EXISTS sys_system_info;
CREATE TABLE sys_system_info (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  system_name varchar(100) NOT NULL DEFAULT '' COMMENT '系统名称',
  system_version varchar(50) NOT NULL DEFAULT '' COMMENT '系统版本',
  system_description varchar(500) NOT NULL DEFAULT '' COMMENT '系统简介',
  tech_stack varchar(500) NOT NULL DEFAULT '' COMMENT '技术栈',
  favicon_url varchar(500) NOT NULL DEFAULT '' COMMENT 'Favicon图标URL',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统信息配置表';

-- 插入默认系统信息
INSERT INTO sys_system_info (system_name, system_version, system_description, tech_stack, favicon_url)
VALUES ('xt-pms', '1.0.0', '基于 Vue 3 + TDesign + Go 的生产管理解决方案', 'Vue 3,Go,MySQL,TDesign', '');

-- 如果表已存在（初次建表后忽略此条），执行以下语句添加 favicon_url 字段：
-- ALTER TABLE sys_system_info ADD COLUMN favicon_url varchar(500) NOT NULL DEFAULT '' COMMENT 'Favicon图标URL' AFTER tech_stack;

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;

-- =============================================
-- 2.0 操作日志表（sys_oper_log）
-- =============================================
DROP TABLE IF EXISTS sys_oper_log;
CREATE TABLE sys_oper_log (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  title varchar(128) NOT NULL DEFAULT '' COMMENT '操作模块标题',
  oper_type tinyint NOT NULL DEFAULT 0 COMMENT '操作类型：1=增 2=删 3=改 4=查 5=导入 6=导出 7=登录 8=登出',
  business_type varchar(64) NOT NULL DEFAULT '' COMMENT '业务类型标识',
  business_id bigint NOT NULL DEFAULT 0 COMMENT '业务记录ID',
  method varchar(255) NOT NULL DEFAULT '' COMMENT '请求方法（路由路径）',
  request_url varchar(255) NOT NULL DEFAULT '' COMMENT '请求URL',
  request_method varchar(10) NOT NULL DEFAULT '' COMMENT '请求方式 GET/POST/PUT/DELETE',
  oper_name varchar(64) NOT NULL DEFAULT '' COMMENT '操作人员',
  dept_name varchar(64) NOT NULL DEFAULT '' COMMENT '部门名称',
  oper_ip varchar(128) NOT NULL DEFAULT '' COMMENT '操作IP',
  request_params text COMMENT '请求参数',
  json_result text COMMENT '返回结果',
  status tinyint NOT NULL DEFAULT 0 COMMENT '操作状态：0=正常 1=异常',
  error_msg varchar(2048) NOT NULL DEFAULT '' COMMENT '错误消息',
  cost_time bigint NOT NULL DEFAULT 0 COMMENT '耗时(ms)',
  create_time datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_oper_name (oper_name) COMMENT '按操作人查询',
  KEY idx_oper_type (oper_type) COMMENT '按操作类型查询',
  KEY idx_title (title) COMMENT '按模块查询',
  KEY idx_business_type (business_type) COMMENT '按业务类型查询',
  KEY idx_status (status) COMMENT '按状态查询',
  KEY idx_create_time (create_time) COMMENT '按时间排序查询'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志记录';

-- ============================================
-- 系统通知模块
-- ============================================

-- 系统通知表
DROP TABLE IF EXISTS sys_notification;
CREATE TABLE sys_notification (
  id          bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  user_id     bigint NOT NULL COMMENT '接收用户 ID',
  title       varchar(200) NOT NULL DEFAULT '' COMMENT '通知标题',
  content     varchar(500) NOT NULL DEFAULT '' COMMENT '通知内容',
  type        varchar(50) NOT NULL DEFAULT '' COMMENT '通知类型（如 mold:work_order_status、production:order_status 等）',
  type_label  varchar(50) NOT NULL DEFAULT '' COMMENT '类型显示名称（如 模具工单、生产订单）',
  quality     varchar(20) NOT NULL DEFAULT 'low' COMMENT '优先级：low=低 middle=中 high=高',
  ref_id      bigint DEFAULT 0 COMMENT '关联业务记录 ID',
  ref_url     varchar(255) DEFAULT '' COMMENT '关联跳转链接',
  is_read     tinyint NOT NULL DEFAULT 0 COMMENT '是否已读：0=未读 1=已读',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (id),
  KEY idx_user_id (user_id) COMMENT '用户索引',
  KEY idx_user_read (user_id, is_read) COMMENT '用户未读查询索引',
  KEY idx_create_time (create_time) COMMENT '时间排序索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统通知表';

-- 通知配置表
DROP TABLE IF EXISTS sys_notification_config;
CREATE TABLE sys_notification_config (
  id           bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  type_pattern varchar(100) NOT NULL COMMENT '通知类型匹配模式，支持 LIKE 通配符，如 mold:% 匹配所有模具通知',
  target_type  varchar(20) NOT NULL COMMENT '目标类型：role=角色 user=用户',
  target_id    bigint NOT NULL COMMENT '目标 ID（角色 ID 或用户 ID）',
  enabled      tinyint NOT NULL DEFAULT 1 COMMENT '是否启用：1=启用 0=禁用',
  create_time  datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time  datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  KEY idx_type_pattern (type_pattern),
  KEY idx_target (target_type, target_id),
  UNIQUE KEY uk_type_target (type_pattern, target_type, target_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知配置表';

-- ============================================
-- 菜单权限模块
-- ============================================

-- 菜单权限表
DROP TABLE IF EXISTS sys_menu;
CREATE TABLE sys_menu (
  id bigint NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
  parent_id bigint DEFAULT 0 COMMENT '父菜单ID，0=顶级',
  menu_name varchar(50) NOT NULL COMMENT '菜单名称',
  menu_type tinyint DEFAULT 1 COMMENT '1=目录 2=菜单 3=按钮 4=页面元素',
  permission varchar(100) DEFAULT NULL COMMENT '权限标识，如 system:user:list',
  path varchar(200) DEFAULT NULL COMMENT '路由路径',
  component varchar(255) DEFAULT NULL COMMENT '组件路径',
  icon varchar(50) DEFAULT NULL COMMENT '图标',
  sort int DEFAULT 0 COMMENT '排序',
  status tinyint DEFAULT 1 COMMENT '0=禁用 1=启用',
  create_time datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (id),
  UNIQUE KEY uk_permission (permission) COMMENT '权限标识唯一',
  INDEX idx_parent_id (parent_id) COMMENT '父菜单索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单权限表';

-- 角色菜单关联表
DROP TABLE IF EXISTS sys_role_menu;
CREATE TABLE sys_role_menu (
  role_id bigint NOT NULL COMMENT '角色ID',
  menu_id bigint NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (role_id, menu_id),
  KEY idx_role_menu_menu (menu_id),
  CONSTRAINT fk_role_menu_role FOREIGN KEY (role_id) REFERENCES sys_role(id) ON DELETE CASCADE,
  CONSTRAINT fk_role_menu_menu FOREIGN KEY (menu_id) REFERENCES sys_menu(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';

-- 恢复外键检查
SET FOREIGN_KEY_CHECKS = 1;