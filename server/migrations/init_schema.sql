
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for _backup_mold_acceptance
-- ----------------------------
DROP TABLE IF EXISTS `_backup_mold_acceptance`;
CREATE TABLE `_backup_mold_acceptance` (
  `id` bigint NOT NULL DEFAULT '0' COMMENT '验收 ID',
  `mold_work_id` bigint NOT NULL COMMENT '关联模具工单 ID',
  `acceptance_time` datetime DEFAULT NULL COMMENT '验收时间',
  `acceptor_id` bigint DEFAULT NULL COMMENT '验收人 ID',
  `acceptance_result` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '验收结果：合格/不合格/让步接收',
  `quality_score` int DEFAULT NULL COMMENT '质量评分（1-100）',
  `dimension_check` text COLLATE utf8mb4_unicode_ci COMMENT '尺寸检测记录',
  `appearance_check` text COLLATE utf8mb4_unicode_ci COMMENT '外观检查记录',
  `function_check` text COLLATE utf8mb4_unicode_ci COMMENT '功能测试记录',
  `issues_found` text COLLATE utf8mb4_unicode_ci COMMENT '发现的问题',
  `rectification_required` tinyint DEFAULT '0' COMMENT '是否需要整改：0=否，1=是',
  `rectification_deadline` datetime DEFAULT NULL COMMENT '整改截止日期',
  `attachment_urls` text COLLATE utf8mb4_unicode_ci COMMENT '附件URL列表（JSON格式）',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- Table structure for _backup_quality_inspection
-- ----------------------------
DROP TABLE IF EXISTS `_backup_quality_inspection`;
CREATE TABLE `_backup_quality_inspection` (
  `id` bigint NOT NULL DEFAULT '0' COMMENT '检验ID',
  `order_id` bigint NOT NULL COMMENT '订单ID',
  `item_id` bigint DEFAULT NULL COMMENT '明细项ID（支持按明细项质检）',
  `inspection_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '检验类型（字典：inspection-type）',
  `inspection_date` datetime NOT NULL COMMENT '检验时间',
  `sample_qty` int DEFAULT '0' COMMENT '抽样数量',
  `ok_qty` int DEFAULT '0' COMMENT '合格数量',
  `bad_qty` int DEFAULT '0' COMMENT '不合格数量',
  `inspector_id` bigint DEFAULT NULL COMMENT '检验员ID',
  `result` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '检验结果（字典：inspection-result）',
  `dimension_check` text COLLATE utf8mb4_unicode_ci COMMENT '尺寸检测记录',
  `appearance_check` text COLLATE utf8mb4_unicode_ci COMMENT '外观检查记录',
  `function_check` text COLLATE utf8mb4_unicode_ci COMMENT '功能测试记录',
  `defect_description` text COLLATE utf8mb4_unicode_ci COMMENT '缺陷描述',
  `photos` json DEFAULT NULL COMMENT '检验照片URL列表',
  `conclusion` text COLLATE utf8mb4_unicode_ci COMMENT '检验结论',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----------------------------
-- Table structure for accounts_payable
-- ----------------------------
DROP TABLE IF EXISTS `accounts_payable`;
CREATE TABLE `accounts_payable` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '应付账款ID',
  `supplier_id` bigint NOT NULL COMMENT '供应商/加工商ID（supplier.id 或 processor.id）',
  `order_id` bigint DEFAULT NULL COMMENT '关联订单ID（outsourcing_record.id 等）',
  `order_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '订单类型：outsourcing=外发订单, purchase=采购订单',
  `payment_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '款项类型：total=全款, first=首期款, final=尾款, installment=分期款',
  `source_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '款项来源：outsource=外发, purchase=采购',
  `order_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '关联订单编号（外发单号等）',
  `invoice_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发票号码',
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '应付金额',
  `paid_amount` decimal(12,2) DEFAULT '0.00' COMMENT '已付金额',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '优惠减免金额',
  `balance` decimal(12,2) GENERATED ALWAYS AS (((`amount` - `paid_amount`) - `discount_amount`)) STORED COMMENT '余额',
  `due_date` date DEFAULT NULL COMMENT '到期日期',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '状态：pending=待付款, partial=部分付款, paid=已付款, overdue=逾期',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_supplier_id` (`supplier_id`) COMMENT '供应商ID索引',
  KEY `idx_order_id` (`order_id`) COMMENT '订单ID索引',
  KEY `idx_due_date` (`due_date`) COMMENT '到期日期索引',
  KEY `idx_status` (`status`) COMMENT '状态索引',
  KEY `idx_ap_supplier_status` (`supplier_id`,`status`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应付账款表';

-- ----------------------------
-- Table structure for accounts_receivable
-- ----------------------------
DROP TABLE IF EXISTS `accounts_receivable`;
CREATE TABLE `accounts_receivable` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '应收账款ID',
  `customer_id` bigint NOT NULL COMMENT '客户ID（customer.id）',
  `order_id` bigint DEFAULT NULL COMMENT '关联订单ID（mold_order.id 或 sale_order.id）',
  `order_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '订单类型：mold_order=模具订单, sale_order=销售订单',
  `payment_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '款项类型：total=总应收, deposit=首付款, final=尾款, production=生产款, freight=运费, packaging=包装费',
  `invoice_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发票号码',
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '应收金额',
  `received_amount` decimal(12,2) DEFAULT '0.00' COMMENT '已收金额',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '优惠减免金额',
  `balance` decimal(12,2) GENERATED ALWAYS AS (((`amount` - `received_amount`) - `discount_amount`)) STORED COMMENT '余额',
  `due_date` date DEFAULT NULL COMMENT '到期日期',
  `aging_days` int DEFAULT '0' COMMENT '逾期天数（由应用层计算更新）',
  `aging_category` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '账龄分类：current=当期, within_30=30天内, within_60=60天内, within_90=90天内, overdue=逾期',
  `status` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'pending' COMMENT '状态：pending=待收款, partial=部分收款, paid=已收款, overdue=逾期',
  `is_locked` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否锁定：0=否, 1=是',
  `unlock_rule` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '解锁规则，例如：delivery_completed',
  `source_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '款项来源：mold=模具款, production=生产款',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
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
  KEY `idx_ar_customer_status` (`customer_id`,`status`),
  CONSTRAINT `fk_ar_customer` FOREIGN KEY (`customer_id`) REFERENCES `customer` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=46 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='应收账款表';

-- ----------------------------
-- Table structure for customer
-- ----------------------------
DROP TABLE IF EXISTS `customer`;
CREATE TABLE `customer` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '客户 ID',
  `cus_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '客户编号（唯一）',
  `cus_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '客户名称',
  `contact` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系人',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '电子邮箱',
  `address` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '客户地址',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `status` tinyint DEFAULT '1' COMMENT '客户状态：0=停用，1=正常',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_cus_code` (`cus_code`) COMMENT '客户编号唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='客户表';

-- ----------------------------
-- Table structure for expense_payments
-- ----------------------------
DROP TABLE IF EXISTS `expense_payments`;
CREATE TABLE `expense_payments` (
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='费用支付记录表';

-- ----------------------------
-- Table structure for external_mold_archive
-- ----------------------------
DROP TABLE IF EXISTS `external_mold_archive`;
CREATE TABLE `external_mold_archive` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '档案ID',
  `mold_code` varchar(100) NOT NULL COMMENT '外来模具编号（唯一）',
  `product_name` varchar(200) NOT NULL COMMENT '产品名称',
  `customer_id` bigint unsigned DEFAULT NULL COMMENT '客户ID（可选）',
  `customer_name` varchar(100) DEFAULT NULL COMMENT '客户名称（冗余字段，便于查询）',
  `first_order_id` bigint unsigned DEFAULT NULL COMMENT '首次使用的订单ID',
  `first_use_date` datetime DEFAULT NULL COMMENT '首次使用日期',
  `last_order_id` bigint unsigned DEFAULT NULL COMMENT '最近使用的订单ID',
  `last_use_date` datetime DEFAULT NULL COMMENT '最近使用日期',
  `use_count` int unsigned DEFAULT '0' COMMENT '使用次数',
  `status` tinyint DEFAULT '1' COMMENT '状态：1=启用，0=禁用',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mold_code` (`mold_code`),
  KEY `idx_customer_id` (`customer_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='外来模具档案表';

-- ----------------------------
-- Table structure for files
-- ----------------------------
DROP TABLE IF EXISTS `files`;
CREATE TABLE `files` (
  `id` varchar(36) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'UUID',
  `url` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '存储访问路径（如 /uploads/ab/cd/uuid_filename.pdf）',
  `file_name` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '原始文件名',
  `file_size` bigint NOT NULL DEFAULT '0' COMMENT '文件大小（字节）',
  `file_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '文件扩展名（如 .pdf/.jpg）',
  `ref_count` int NOT NULL DEFAULT '0' COMMENT '引用计数（被业务引用的次数）',
  `deleted_at` datetime DEFAULT NULL COMMENT '软删除时间（ref_count=0 时设置）',
  `business_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '业务类型（expense/receivable/mold/product 等）',
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

-- ----------------------------
-- Table structure for finance_expense
-- ----------------------------
DROP TABLE IF EXISTS `finance_expense`;
CREATE TABLE `finance_expense` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '费用ID',
  `expense_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '费用类型：freight=运费, packaging=包装费, insurance=保险费, other=其他',
  `related_order_id` bigint DEFAULT NULL COMMENT '关联订单ID（mold_order.id 或 sale_order.id）',
  `related_order_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '订单类型：mold_order=模具订单, sale_order=销售订单',
  `delivery_id` bigint DEFAULT NULL COMMENT '关联交付记录ID（mold_delivery.id）',
  `amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '费用金额',
  `paid_amount` decimal(12,2) DEFAULT '0.00' COMMENT '已付金额',
  `payment_date` date DEFAULT NULL COMMENT '支付日期',
  `payer` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '支付方',
  `receiver` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收款方',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_related_order` (`related_order_id`,`related_order_type`),
  KEY `idx_delivery_id` (`delivery_id`),
  KEY `idx_expense_type` (`expense_type`),
  KEY `idx_create_time` (`create_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='财务费用表';

-- ----------------------------
-- Table structure for master_orders
-- ----------------------------
DROP TABLE IF EXISTS `master_orders`;
CREATE TABLE `master_orders` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '综合订单 ID',
  `order_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '综合订单编号（唯一）',
  `customer_id` bigint NOT NULL COMMENT '客户 ID',
  `order_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'COMBINED' COMMENT '下单类型: COMBINED=模具+产品, MOLD_ONLY=仅模具, PROD_ONLY=仅产品',
  `status` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'PROCESSING' COMMENT '整体状态',
  `total_amount` decimal(10,2) DEFAULT '0.00' COMMENT '订单总金额',
  `additional_amount` decimal(10,2) DEFAULT '0.00' COMMENT '附加金额（税点、杂费等）',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `cancel_reason` text COLLATE utf8mb4_unicode_ci COMMENT '取消原因',
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
) ENGINE=InnoDB AUTO_INCREMENT=85 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='综合订单表';

-- ----------------------------
-- Table structure for material
-- ----------------------------
DROP TABLE IF EXISTS `material`;
CREATE TABLE `material` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '物料 ID',
  `mat_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '物料编码（唯一）',
  `mat_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '物料名称',
  `spec` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '物料规格型号',
  `unit` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '物料单位（如：个、kg）',
  `mat_type` tinyint DEFAULT '1' COMMENT '物料类型：1=原料，2=模具配件，3=产品，4=辅料',
  `price` decimal(10,2) DEFAULT '0.00' COMMENT '物料单价',
  `stock_warn` int DEFAULT '0' COMMENT '库存预警数量',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mat_code` (`mat_code`) COMMENT '物料编码唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=24 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='物料表';

-- ----------------------------
-- Table structure for mold
-- ----------------------------
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

-- ----------------------------
-- Table structure for mold_bom
-- ----------------------------
DROP TABLE IF EXISTS `mold_bom`;
CREATE TABLE `mold_bom` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'BOM ID',
  `mold_id` bigint NOT NULL COMMENT '关联模具 ID（mold.id）',
  `material_id` bigint NOT NULL COMMENT '关联物料 ID（material.id）',
  `qty` decimal(10,2) DEFAULT '1.00' COMMENT '物料用量',
  `unit_price` decimal(10,2) DEFAULT '0.00' COMMENT '物料单价',
  `total_price` decimal(10,2) DEFAULT '0.00' COMMENT '总价',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注（如：模具配件规格）',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_mold_id` (`mold_id`) COMMENT '模具 ID 索引',
  KEY `idx_material_id` (`material_id`) COMMENT '物料 ID 索引',
  CONSTRAINT `fk_mold_bom_material` FOREIGN KEY (`material_id`) REFERENCES `material` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_mold_bom_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具 BOM 表（模具用料清单）';

-- ----------------------------
-- Table structure for mold_delivery
-- ----------------------------
DROP TABLE IF EXISTS `mold_delivery`;
CREATE TABLE `mold_delivery` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '交付 ID',
  `mold_work_id` bigint NOT NULL COMMENT '关联模具工单 ID',
  `delivery_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交付单号',
  `customer_id` bigint NOT NULL COMMENT '客户 ID',
  `delivery_time` datetime DEFAULT NULL COMMENT '交付时间',
  `deliverer_id` bigint DEFAULT NULL COMMENT '交付人 ID',
  `receiver_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收货人姓名',
  `receiver_phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收货人电话',
  `delivery_address` text COLLATE utf8mb4_unicode_ci COMMENT '交付地址',
  `logistics_company` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '物流公司',
  `tracking_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '物流单号',
  `delivery_method` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交付方式：自提/快递/物流',
  `documents_attached` text COLLATE utf8mb4_unicode_ci COMMENT '随附文档（JSON格式）',
  `confirmation_status` tinyint DEFAULT '1' COMMENT '确认状态：1=待确认，2=已确认，3=拒收',
  `confirmation_time` datetime DEFAULT NULL COMMENT '确认时间',
  `rejection_reason` text COLLATE utf8mb4_unicode_ci COMMENT '拒收原因',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_delivery_no` (`delivery_no`) COMMENT '交付单号唯一索引',
  KEY `idx_mold_work_id` (`mold_work_id`) COMMENT '模具工单 ID 索引',
  KEY `idx_customer_id` (`customer_id`) COMMENT '客户 ID 索引',
  CONSTRAINT `fk_mold_delivery_work` FOREIGN KEY (`mold_work_id`) REFERENCES `mold_work_order` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具交付记录表';

-- ----------------------------
-- Table structure for mold_order
-- ----------------------------
DROP TABLE IF EXISTS `mold_order`;
CREATE TABLE `mold_order` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模具订单 ID',
  `order_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单编号（唯一）',
  `customer_id` bigint NOT NULL COMMENT '关联客户 ID（customer.id）',
  `order_type` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT 'NEW_MOLD' COMMENT '订单类型：NEW_MOLD=新模订单，REPAIR=修模订单，MODIFICATION=改造订单',
  `create_user_id` bigint DEFAULT NULL COMMENT '创建人ID',
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
  KEY `idx_mo_customer_status` (`customer_id`,`status`),
  CONSTRAINT `fk_mold_order_customer` FOREIGN KEY (`customer_id`) REFERENCES `customer` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_mold_order_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=71 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具订单表';

-- ----------------------------
-- Table structure for mold_order_item
-- ----------------------------
DROP TABLE IF EXISTS `mold_order_item`;
CREATE TABLE `mold_order_item` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '明细ID',
  `order_id` bigint NOT NULL COMMENT '订单ID（mold_order.id）',
  `mold_id` bigint NOT NULL COMMENT '模具ID（mold.id）',
  `amount` decimal(12,2) DEFAULT '0.00' COMMENT '该模具的合同金额',
  `sort` int DEFAULT '0' COMMENT '排序号',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_mold` (`order_id`,`mold_id`) COMMENT '同一订单下模具唯一',
  KEY `idx_order_id` (`order_id`) COMMENT '订单ID索引',
  KEY `idx_mold_id` (`mold_id`) COMMENT '模具ID索引',
  KEY `idx_sort` (`sort`),
  CONSTRAINT `fk_order_item_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_order_item_order` FOREIGN KEY (`order_id`) REFERENCES `mold_order` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=50 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具订单明细表（一单多款）';

-- ----------------------------
-- Table structure for mold_order_payment
-- ----------------------------
DROP TABLE IF EXISTS `mold_order_payment`;
CREATE TABLE `mold_order_payment` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '款项 ID',
  `order_id` bigint NOT NULL COMMENT '关联模具订单 ID',
  `work_order_id` bigint DEFAULT NULL COMMENT '关联工单 ID（mold_work_order.id），用于按模具核算',
  `payment_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '款项类型（自定义，如：首款、二期款、尾款等）',
  `contract_amount` decimal(12,2) DEFAULT '0.00' COMMENT '合同约定金额',
  `actual_amount` decimal(12,2) DEFAULT '0.00' COMMENT '实际到账金额',
  `should_amount` decimal(12,2) DEFAULT '0.00' COMMENT '应收尾款金额',
  `payment_date` datetime DEFAULT NULL COMMENT '收款日期',
  `status` tinyint DEFAULT '1' COMMENT '状态：1=未收，2=部分收款，3=已收款',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_payment_work_order_id` (`work_order_id`),
  CONSTRAINT `fk_payment_order` FOREIGN KEY (`order_id`) REFERENCES `mold_order` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具订单款项表';

-- ----------------------------
-- Table structure for mold_product_relation
-- ----------------------------
DROP TABLE IF EXISTS `mold_product_relation`;
CREATE TABLE `mold_product_relation` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `mold_id` bigint NOT NULL COMMENT '模具ID (mold.id)',
  `product_id` bigint DEFAULT NULL COMMENT '产品ID (product.id)，可选',
  `production_order_item_id` bigint DEFAULT NULL COMMENT '生产订单明细ID (production_order_item.id)',
  `product_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品名称快照（生产时填写）',
  `product_image_url` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品图片URL（生产时上传）',
  `plan_qty` int DEFAULT '0' COMMENT '计划生产数量',
  `relation_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'PRODUCTION' COMMENT '关联类型: PRODUCTION=生产关联',
  `remark` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具产品关联表（记录模具与产品的生产关系）';

-- ----------------------------
-- Table structure for mold_stock
-- ----------------------------
DROP TABLE IF EXISTS `mold_stock`;
CREATE TABLE `mold_stock` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '库存记录 ID',
  `mold_id` bigint NOT NULL COMMENT '关联模具 ID（mold.id）',
  `warehouse_id` bigint DEFAULT NULL COMMENT '关联仓库 ID（warehouse.id）',
  `stock_type` tinyint NOT NULL COMMENT '库存类型：1=入库，2=出库，3=交付客户',
  `operate_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  `operator_id` bigint DEFAULT NULL COMMENT '操作人 ID（sys_user.id）',
  `customer_id` bigint DEFAULT NULL COMMENT '交付客户 ID（stock_type=3 时必填，customer.id）',
  `quantity` int DEFAULT '1' COMMENT '数量',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注（如：入库原因、交付单号）',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_mold_id` (`mold_id`) COMMENT '模具 ID 索引',
  KEY `idx_warehouse_id` (`warehouse_id`) COMMENT '仓库 ID 索引',
  KEY `idx_operator_id` (`operator_id`) COMMENT '操作人 ID 索引',
  KEY `idx_customer_id` (`customer_id`) COMMENT '客户 ID 索引',
  CONSTRAINT `fk_mold_stock_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具库存记录表（跟踪模具库存/交付状态）';

-- ----------------------------
-- Table structure for mold_try_test
-- ----------------------------
DROP TABLE IF EXISTS `mold_try_test`;
CREATE TABLE `mold_try_test` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '试模 ID',
  `mold_work_id` bigint NOT NULL COMMENT '关联模具工单 ID（mold_work_order.id）',
  `test_time` datetime DEFAULT NULL COMMENT '试模时间',
  `ok_qty` int DEFAULT '0' COMMENT '试模合格数量',
  `bad_qty` int DEFAULT '0' COMMENT '试模不良数量',
  `result` text COLLATE utf8mb4_unicode_ci COMMENT '试模结果（如：合格、需返修）',
  `test_mode` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '试模方式（自家试模、外协试模、免试）',
  `user_id` bigint DEFAULT NULL COMMENT '操作人 ID（sys_user.id）',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_mold_work_id` (`mold_work_id`) COMMENT '模具工单 ID 索引',
  KEY `idx_user_id` (`user_id`) COMMENT '操作人 ID 索引',
  CONSTRAINT `fk_mold_try_work` FOREIGN KEY (`mold_work_id`) REFERENCES `mold_work_order` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=27 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具试模记录表';

-- ----------------------------
-- Table structure for mold_usage_record
-- ----------------------------
DROP TABLE IF EXISTS `mold_usage_record`;
CREATE TABLE `mold_usage_record` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '使用记录ID',
  `mold_id` bigint NOT NULL COMMENT '模具ID',
  `production_order_id` bigint NOT NULL COMMENT '生产订单ID',
  `production_order_item_id` bigint DEFAULT NULL COMMENT '关联生产订单明细ID（用于细分到具体产品）',
  `product_id` bigint DEFAULT NULL COMMENT '产品ID（用于区分同一模具不同产品）',
  `product_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品名称快照',
  `usage_date` date DEFAULT NULL COMMENT '使用日期',
  `total_shots` int DEFAULT '0' COMMENT '累计模次(本次使用)',
  `ok_qty` int DEFAULT '0' COMMENT '良品数量',
  `bad_qty` int DEFAULT '0' COMMENT '不良数量',
  `start_date` date DEFAULT NULL COMMENT '开始使用日期',
  `end_date` date DEFAULT NULL COMMENT '结束使用日期',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mold_order` (`mold_id`,`production_order_id`),
  KEY `idx_mold_id` (`mold_id`),
  KEY `idx_production_order_id` (`production_order_id`),
  KEY `idx_production_order_item_id` (`production_order_item_id`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_mur_mold_product` (`mold_id`,`product_id`),
  KEY `idx_mur_item_product` (`production_order_item_id`,`product_id`),
  CONSTRAINT `fk_mur_production_item` FOREIGN KEY (`production_order_item_id`) REFERENCES `production_order_item` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_usage_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_usage_order` FOREIGN KEY (`production_order_id`) REFERENCES `production_order` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具使用记录表';

-- ----------------------------
-- Table structure for mold_work_order
-- ----------------------------
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
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `original_mold_status` VARCHAR(50) DEFAULT NULL COMMENT '修模/改造前模具的原始生命周期状态',
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
  KEY `idx_mwo_mold_status` (`mold_id`,`status`),
  CONSTRAINT `fk_mold_work_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_mold_work_order` FOREIGN KEY (`mold_order_id`) REFERENCES `mold_order` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=68 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='模具生产工单表';

-- ----------------------------
-- Table structure for outsourcing_record
-- ----------------------------
DROP TABLE IF EXISTS `outsourcing_record`;
CREATE TABLE `outsourcing_record` (
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
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外发记录表';

-- ----------------------------
-- Table structure for outsourcing_report
-- ----------------------------
DROP TABLE IF EXISTS `outsourcing_report`;
CREATE TABLE `outsourcing_report` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '外发报工ID',
  `outsourcing_id` bigint NOT NULL COMMENT '外发记录ID',
  `report_date` datetime DEFAULT NULL COMMENT '报工日期',
  `ok_qty` int DEFAULT '0' COMMENT '良品数量',
  `bad_qty` int DEFAULT '0' COMMENT '不良数量',
  `defect_reason` text COLLATE utf8mb4_unicode_ci COMMENT '不良原因',
  `photos` json DEFAULT NULL COMMENT '现场照片URL列表',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_outsourcing_id` (`outsourcing_id`),
  CONSTRAINT `fk_orsr_outsourcing` FOREIGN KEY (`outsourcing_id`) REFERENCES `outsourcing_record` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='外发报工表';

-- ----------------------------
-- Table structure for payable_payments
-- ----------------------------
DROP TABLE IF EXISTS `payable_payments`;
CREATE TABLE `payable_payments` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '付款记录ID',
  `payable_id` bigint NOT NULL COMMENT '应付账款ID（accounts_payable.id）',
  `payment_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '付款金额',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '本次优惠金额',
  `payment_date` date NOT NULL COMMENT '付款日期',
  `payment_method` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '付款方式：bank_transfer=银行转账, cash=现金, check=支票, wechat=微信, alipay=支付宝',
  `transaction_no` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交易流水号/回单号',
  `payment_account` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '付款账户',
  `receiver_account` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收款方账户',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `receipt_file_id` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '回单文件ID（多个ID用逗号分隔）',
  `invoice_file_id` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发票文件ID（多个ID用逗号分隔）',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_payable_id` (`payable_id`),
  KEY `idx_payment_date` (`payment_date`),
  KEY `idx_payment_method` (`payment_method`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='付款记录表';

-- ----------------------------
-- Table structure for print_notice_config
-- ----------------------------
DROP TABLE IF EXISTS `print_notice_config`;
CREATE TABLE `print_notice_config` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `user_id` bigint NOT NULL COMMENT '用户ID，0表示全局配置',
  `config_type` varchar(50) NOT NULL DEFAULT 'production_order' COMMENT '配置类型：production_order=生产订单',
  `config_data` json NOT NULL COMMENT '配置数据（JSON格式）',
  `is_default` tinyint(1) DEFAULT '0' COMMENT '是否为默认配置：1=是，0=否',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_type` (`user_id`,`config_type`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_config_type` (`config_type`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='打印通知单配置表';

-- ----------------------------
-- Table structure for process
-- ----------------------------
DROP TABLE IF EXISTS `process`;
CREATE TABLE `process` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '工序 ID',
  `process_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工序编码（唯一）',
  `process_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '工序名称（如：CNC、线割、装配）',
  `process_type` tinyint DEFAULT '1' COMMENT '工序类型：1=模具工序，2=产品工序',
  `work_hour` decimal(8,2) DEFAULT '0.00' COMMENT '标准工时',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_process_code` (`process_code`) COMMENT '工序编码唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='工序模板表';

-- ----------------------------
-- Table structure for process_template
-- ----------------------------
DROP TABLE IF EXISTS `process_template`;
CREATE TABLE `process_template` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `group_id` int DEFAULT '0' COMMENT '分组ID（同组工序可并行）',
  `group_name` varchar(50) DEFAULT NULL COMMENT '分组名称',
  `process_name` varchar(100) NOT NULL COMMENT '工序名称（如：CNC、磨床、放电等）',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序号',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1=启用，2=禁用',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_sort` (`sort`),
  KEY `idx_status` (`status`),
  KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='工序模板表';

-- ----------------------------
-- Table structure for processor
-- ----------------------------
DROP TABLE IF EXISTS `processor`;
CREATE TABLE `processor` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '加工商 ID',
  `processor_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '加工商编号（唯一）',
  `processor_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '加工商名称',
  `contact` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系人',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '电子邮箱',
  `address` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加工商地址',
  `process_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '加工类型（用户自定义填写）',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `status` tinyint DEFAULT '1' COMMENT '状态：0=停用，1=正常',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_processor_code` (`processor_code`) COMMENT '加工商编号唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='加工商表';

-- ----------------------------
-- Table structure for product
-- ----------------------------
DROP TABLE IF EXISTS `product`;
CREATE TABLE `product` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '产品 ID',
  `pro_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '产品编码（唯一）',
  `pro_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '产品名称',
  `spec` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品规格',
  `spec_detail` text COLLATE utf8mb4_unicode_ci COMMENT '产品详细规格参数(JSON格式)',
  `mold_id` bigint DEFAULT NULL COMMENT '默认使用模具 ID（mold.id）',
  `unit` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品单位',
  `price` decimal(10,2) DEFAULT '0.00' COMMENT '产品销售单价',
  `weight` decimal(10,2) DEFAULT '0.00' COMMENT '产品重量',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '产品状态：0=停用, 1=启用',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `image_url` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品图片URL（手动上传，优先展示）',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_pro_code` (`pro_code`) COMMENT '产品编码唯一索引',
  UNIQUE KEY `uk_pro_name` (`pro_name`) COMMENT '产品名称唯一索引，防止同名产品重复创建',
  KEY `idx_mold_id` (`mold_id`) COMMENT '模具 ID 索引',
  CONSTRAINT `fk_product_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=14 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品表';

-- ----------------------------
-- Table structure for product_bom
-- ----------------------------
DROP TABLE IF EXISTS `product_bom`;
CREATE TABLE `product_bom` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'BOM ID',
  `product_id` bigint NOT NULL COMMENT '关联产品 ID（product.id）',
  `material_id` bigint NOT NULL COMMENT '关联物料 ID（material.id）',
  `qty` decimal(10,2) DEFAULT '1.00' COMMENT '单台产品物料用量',
  `loss_rate` decimal(5,2) DEFAULT '0.00' COMMENT '物料损耗率（%）',
  `unit_price` decimal(10,2) DEFAULT '0.00' COMMENT '物料单价',
  `total_price` decimal(10,2) DEFAULT '0.00' COMMENT '总价',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_product_id` (`product_id`) COMMENT '产品 ID 索引',
  KEY `idx_material_id` (`material_id`) COMMENT '物料 ID 索引',
  CONSTRAINT `fk_product_bom_material` FOREIGN KEY (`material_id`) REFERENCES `material` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_product_bom_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='产品 BOM 表（产品用料清单）';

-- ----------------------------
-- Table structure for production_order
-- ----------------------------
DROP TABLE IF EXISTS `production_order`;
CREATE TABLE `production_order` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '生产订单ID',
  `order_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '订单编号（唯一）',
  `customer_id` bigint NOT NULL COMMENT '客户ID（customer.id）',
  `product_name` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '产品名称（手动输入）',
  `production_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'IN_HOUSE' COMMENT '生产方式：IN_HOUSE=内部生产，OUTSOURCED=外发生产',
  `mold_source_type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'SELF_MADE' COMMENT '模具来源：SELF_MADE=自有模具，CUSTOMER_PROVIDED=外来模具',
  `mold_id` bigint DEFAULT NULL COMMENT '模具ID（mold.id），自有模具时必填',
  `external_mold_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '外来模具编号（外来模具时填写）',
  `external_mold_info` json DEFAULT NULL COMMENT '外来模具信息JSON（模具来源=外来时必填）',
  `plan_qty` int NOT NULL DEFAULT '0' COMMENT '计划生产数量',
  `completed_qty` int DEFAULT '0' COMMENT '已完成数量',
  `ok_qty` int DEFAULT '0' COMMENT '良品数量',
  `bad_qty` int DEFAULT '0' COMMENT '不良数量',
  `plan_start_date` date DEFAULT NULL COMMENT '计划开始日期',
  `plan_end_date` date DEFAULT NULL COMMENT '计划交期',
  `actual_start_date` date DEFAULT NULL COMMENT '实际开始日期',
  `actual_end_date` date DEFAULT NULL COMMENT '实际完成日期',
  `device_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '生产设备字典值',
  `worker_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '负责人字典值',
  `processor_id` bigint DEFAULT NULL COMMENT '加工商ID（processor.id）',
  `outsource_price` decimal(10,2) DEFAULT '0.00' COMMENT '外发单价',
  `outsource_total_amount` decimal(10,2) DEFAULT '0.00' COMMENT '外发总金额',
  `material_cost` decimal(10,2) DEFAULT '0.00' COMMENT '材料成本',
  `processing_cost` decimal(10,2) DEFAULT '0.00' COMMENT '加工成本',
  `total_cost` decimal(10,2) DEFAULT '0.00' COMMENT '总成本',
  `unit_price` decimal(10,2) DEFAULT '0.00' COMMENT '产品单价',
  `total_amount` decimal(10,2) DEFAULT '0.00' COMMENT '订单总金额',
  `status` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'PENDING_SCHEDULE' COMMENT '生产订单状态：PENDING_SCHEDULE=待排产, SCHEDULED=已排产, PRODUCING=生产中, STOCKED_IN=已入库, SHIPPED=已出货, COMPLETED=已完成, CANCELLED=已取消',
  `priority` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'URGENT' COMMENT '优先级：NORMAL=普通，IMPORTANT=重要，URGENT=紧急，VERY_URGENT=非常紧急',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_user_id` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `mold_shots` int DEFAULT '0' COMMENT '模具使用次数(模次)',
  `process_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '生产工艺类型：injection=注塑, blow_molding=吹塑, stamping=冲压',
  `process_remark` text COLLATE utf8mb4_unicode_ci COMMENT '工艺备注/详细说明',
  `material_type` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '原料类型：PP, PE, TPE等',
  `product_weight` decimal(10,2) DEFAULT NULL COMMENT '产品重量（克）',
  `has_metal_parts` tinyint(1) DEFAULT '0' COMMENT '是否包含金属配件',
  `manual_process` text COLLATE utf8mb4_unicode_ci COMMENT '手工处理流程说明',
  `assembly_process` text COLLATE utf8mb4_unicode_ci COMMENT '组装流程说明',
  `master_order_id` bigint DEFAULT NULL COMMENT '关联综合订单 ID (master_orders.id)',
  `master_cancelled` tinyint(1) NOT NULL DEFAULT '0' COMMENT '父综合订单已取消标记：0=否，1=是',
  `dependency_status` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT 'NONE' COMMENT '依赖状态: NONE=无依赖, WAITING_MOLD=等待模具, READY=就绪',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_code` (`order_code`),
  KEY `idx_customer_id` (`customer_id`),
  KEY `idx_mold_id` (`mold_id`),
  KEY `idx_production_type` (`production_type`),
  KEY `idx_mold_source_type` (`mold_source_type`),
  KEY `idx_status` (`status`),
  KEY `idx_plan_end_date` (`plan_end_date`),
  KEY `fk_po_device` (`device_id`),
  KEY `idx_master_order_id` (`master_order_id`),
  KEY `idx_mold_status` (`mold_id`,`status`),
  KEY `idx_customer_status` (`customer_id`,`status`),
  KEY `idx_po_processor_id` (`processor_id`),
  KEY `idx_po_plan_end_date` (`plan_end_date`),
  CONSTRAINT `fk_po_customer` FOREIGN KEY (`customer_id`) REFERENCES `customer` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_po_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_po_processor` FOREIGN KEY (`processor_id`) REFERENCES `processor` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=60 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产订单表';

-- ----------------------------
-- Table structure for production_order_item
-- ----------------------------
DROP TABLE IF EXISTS `production_order_item`;
CREATE TABLE `production_order_item` (
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
  KEY `idx_poi_order_status` (`order_id`,`status`),
  CONSTRAINT `fk_poi_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_poi_order` FOREIGN KEY (`order_id`) REFERENCES `production_order` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_poi_product` FOREIGN KEY (`product_id`) REFERENCES `product` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=106 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产订单明细表';

-- ----------------------------
-- Table structure for production_report
-- ----------------------------
DROP TABLE IF EXISTS `production_report`;
CREATE TABLE `production_report` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '报工ID',
  `order_id` bigint NOT NULL COMMENT '订单ID',
  `item_id` bigint DEFAULT NULL COMMENT '订单明细项ID（支持按明细报工）',
  `schedule_id` bigint DEFAULT NULL COMMENT '排程ID',
  `report_date` datetime NOT NULL COMMENT '报工时间',
  `shift` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '班次（字典：shift-type）',
  `ok_qty` int DEFAULT '0' COMMENT '良品数量',
  `bad_qty` int DEFAULT '0' COMMENT '不良数量',
  `scrap_qty` int DEFAULT '0' COMMENT '报废数量',
  `device_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '使用设备',
  `worker_id` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '操作工',
  `work_hours` decimal(5,2) DEFAULT '0.00' COMMENT '工时（小时）',
  `defect_reason` text COLLATE utf8mb4_unicode_ci COMMENT '不良原因',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_user_id` bigint DEFAULT NULL COMMENT '报工人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_schedule_id` (`schedule_id`),
  KEY `idx_report_date` (`report_date`),
  KEY `fk_pr_device` (`device_id`),
  KEY `idx_item_id` (`item_id`),
  CONSTRAINT `fk_pr_order` FOREIGN KEY (`order_id`) REFERENCES `production_order` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_pr_schedule` FOREIGN KEY (`schedule_id`) REFERENCES `production_schedule` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB AUTO_INCREMENT=45 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产报工表';

-- ----------------------------
-- Table structure for production_schedule
-- ----------------------------
DROP TABLE IF EXISTS `production_schedule`;
CREATE TABLE `production_schedule` (
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

-- ----------------------------
-- Table structure for production_status_flow
-- ----------------------------
DROP TABLE IF EXISTS `production_status_flow`;
CREATE TABLE `production_status_flow` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `from_status_value` varchar(50) NOT NULL COMMENT '起始状态值',
  `to_status_value` varchar(50) NOT NULL COMMENT '目标状态值',
  `action_name` varchar(50) NOT NULL COMMENT '操作名称（如：入库、出货、完成）',
  `action_code` varchar(50) NOT NULL COMMENT '操作编码（前端按钮标识）',
  `is_enabled` tinyint DEFAULT '1' COMMENT '是否启用：1=启用，0=禁用',
  `sort` int DEFAULT '0' COMMENT '排序',
  `remark` varchar(255) DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_from_to` (`from_status_value`,`to_status_value`),
  KEY `idx_action_code` (`action_code`)
) ENGINE=InnoDB AUTO_INCREMENT=19 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='生产订单状态流转规则表';

-- ----------------------------
-- Table structure for production_stock_in
-- ----------------------------
DROP TABLE IF EXISTS `production_stock_in`;
CREATE TABLE `production_stock_in` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '入库ID',
  `stock_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '入库单号（唯一）',
  `order_id` bigint DEFAULT NULL COMMENT '订单ID（直接入库时为空）',
  `warehouse_id` bigint NOT NULL COMMENT '仓库ID',
  `mold_id` bigint DEFAULT NULL COMMENT '模具ID（自有模具）',
  `external_mold_code` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '外来模具编号（外来模具）',
  `product_id` bigint DEFAULT NULL COMMENT '产品ID（product.id），用于区分同一模具不同产品的入库',
  `product_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '产品名称（入库时的产品名称）',
  `quantity` int NOT NULL COMMENT '入库数量',
  `batch_no` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '批次号',
  `stock_in_date` datetime DEFAULT NULL COMMENT '入库日期',
  `operator_id` bigint DEFAULT NULL COMMENT '操作人ID',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_stock_code` (`stock_code`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_warehouse_id` (`warehouse_id`),
  KEY `idx_stock_in_date` (`stock_in_date`),
  KEY `idx_mold_id` (`mold_id`),
  KEY `idx_external_mold_code` (`external_mold_code`),
  KEY `idx_product_name` (`product_name`),
  KEY `idx_stock_in_product_id` (`product_id`),
  KEY `idx_stock_code` (`stock_code`),
  KEY `idx_mold_product` (`mold_id`,`product_id`),
  KEY `idx_mold_date` (`mold_id`,`stock_in_date` DESC),
  KEY `idx_external_product` (`external_mold_code`,`product_id`),
  CONSTRAINT `fk_production_stock_in_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_psi_order` FOREIGN KEY (`order_id`) REFERENCES `production_order` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_psi_warehouse` FOREIGN KEY (`warehouse_id`) REFERENCES `warehouse` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='生产入库表';

-- ----------------------------
-- Table structure for production_stock_out
-- ----------------------------
DROP TABLE IF EXISTS `production_stock_out`;
CREATE TABLE `production_stock_out` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '出库ID',
  `stock_code` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '出库单号（正常发货时有值，其他出库类型为NULL）',
  `order_id` bigint DEFAULT NULL COMMENT '订单ID（直接出库时为空）',
  `warehouse_id` bigint NOT NULL COMMENT '仓库ID',
  `mold_id` bigint DEFAULT NULL COMMENT '模具ID（自有模具）',
  `external_mold_code` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '外来模具编号（外来模具）',
  `product_id` bigint DEFAULT NULL COMMENT '产品ID（product.id），用于区分同一模具不同产品的出库',
  `product_name` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '产品名称（出库时的产品名称）',
  `customer_name` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '接收客户名称',
  `quantity` int NOT NULL DEFAULT '0' COMMENT '出库数量',
  `batch_no` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '批次号',
  `stock_out_date` datetime DEFAULT NULL COMMENT '出库日期',
  `out_type` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT 'normal' COMMENT '出库类型(normal=正常发货, other=其他原因出库)',
  `out_reason` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '出库原因(expired=过期丢弃, damaged=损坏, unqualified=不达标, other=其他原因)',
  `out_reason_remark` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '出库原因备注',
  `delivery_date` date DEFAULT NULL COMMENT '发货日期',
  `delivery_method` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发货方式(字典值: express/lalamove/pickup/logistics)',
  `express_company` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '快递公司(字典值)',
  `express_no` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '快递单号',
  `photos` json DEFAULT NULL COMMENT '出库照片(JSON数组)',
  `total_weight` decimal(10,2) DEFAULT NULL COMMENT '总重量(kg)',
  `driver_phone` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '司机电话(货拉拉用)',
  `driver_plate` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '车牌号(货拉拉用)',
  `operator_id` bigint DEFAULT NULL COMMENT '操作人ID',
  `remark` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_order_id` (`order_id`),
  KEY `idx_warehouse_id` (`warehouse_id`),
  KEY `idx_mold_id` (`mold_id`),
  KEY `idx_external_mold_code` (`external_mold_code`),
  KEY `idx_stock_out_date` (`stock_out_date`),
  KEY `idx_product_name` (`product_name`),
  KEY `idx_stock_code` (`stock_code`),
  KEY `idx_out_type` (`out_type`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_mold_product` (`mold_id`,`product_id`),
  KEY `idx_mold_date` (`mold_id`,`stock_out_date` DESC),
  KEY `idx_external_product` (`external_mold_code`,`product_id`),
  CONSTRAINT `fk_stock_out_mold` FOREIGN KEY (`mold_id`) REFERENCES `mold` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_stock_out_order` FOREIGN KEY (`order_id`) REFERENCES `production_order` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_stock_out_warehouse` FOREIGN KEY (`warehouse_id`) REFERENCES `warehouse` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='生产出库表';

-- ----------------------------
-- Table structure for receivable_payments
-- ----------------------------
DROP TABLE IF EXISTS `receivable_payments`;
CREATE TABLE `receivable_payments` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '收款记录ID',
  `receivable_id` bigint NOT NULL COMMENT '应收账款ID（accounts_receivable.id）',
  `payment_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '收款金额',
  `discount_amount` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '本次优惠金额',
  `payment_date` date NOT NULL COMMENT '收款日期',
  `payment_method` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收款方式：bank_transfer=银行转账, cash=现金, check=支票, wechat=微信, alipay=支付宝',
  `transaction_no` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '交易流水号/回单号',
  `receiving_account` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '收款账户',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `receipt_file_id` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '回单文件ID（多个ID用逗号分隔）',
  `invoice_file_id` varchar(500) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '发票文件ID（多个ID用逗号分隔）',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_receivable_id` (`receivable_id`),
  KEY `idx_payment_date` (`payment_date`),
  KEY `idx_payment_method` (`payment_method`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收款记录表';

-- ----------------------------
-- Table structure for receiving_accounts
-- ----------------------------
DROP TABLE IF EXISTS `receiving_accounts`;
CREATE TABLE `receiving_accounts` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '账户ID',
  `account_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '账户名称（如：公司公账、法人银行卡）',
  `account_no` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '账号',
  `bank_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '开户行',
  `account_type` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '账户类型：company=对公, personal=个人, wechat=微信, alipay=支付宝',
  `sort` int DEFAULT '0' COMMENT '排序',
  `status` tinyint DEFAULT '1' COMMENT '状态：1=启用, 0=禁用',
  `remark` text COLLATE utf8mb4_unicode_ci COMMENT '备注',
  `created_by` bigint DEFAULT NULL COMMENT '创建人ID',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_account_name_no` (`account_name`,`account_no`),
  KEY `idx_account_type` (`account_type`),
  KEY `idx_status` (`status`),
  KEY `idx_sort` (`sort`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收款账户表';

-- ----------------------------
-- Table structure for status_flow_config
-- ----------------------------
DROP TABLE IF EXISTS `status_flow_config`;
CREATE TABLE `status_flow_config` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '配置ID',
  `work_order_status` varchar(50) DEFAULT NULL COMMENT '工单状态值（语义化字符串）',
  `mold_order_status` varchar(50) DEFAULT NULL COMMENT '订单状态值（语义化字符串）',
  `mold_status` varchar(50) DEFAULT NULL COMMENT '模具档案状态值（语义化字符串）',
  `description` varchar(500) DEFAULT NULL COMMENT '配置说明',
  `priority` int DEFAULT '0' COMMENT '优先级（数字越小优先级越高）',
  `is_active` tinyint DEFAULT '1' COMMENT '是否启用：1=启用，0=禁用',
  `status_order` int DEFAULT '0' COMMENT '状态顺序号（用于状态比较）',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_work_order` (`work_order_status`),
  KEY `idx_mold_order` (`mold_order_status`),
  KEY `idx_mold` (`mold_status`),
  KEY `idx_active` (`is_active`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='状态同步配置表';

-- ----------------------------
-- Table structure for stock
-- ----------------------------
DROP TABLE IF EXISTS `stock`;
CREATE TABLE `stock` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '库存 ID',
  `material_id` bigint NOT NULL COMMENT '关联物料 ID（material.id）',
  `product_id` bigint DEFAULT NULL COMMENT '产品ID（product.id），用于区分同一物料不同产品的库存',
  `wh_id` bigint NOT NULL COMMENT '关联仓库 ID（warehouse.id）',
  `stock_qty` decimal(12,2) DEFAULT '0.00' COMMENT '当前库存数量',
  `safe_stock` decimal(12,2) DEFAULT '0.00' COMMENT '安全库存数量',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_mat_wh` (`material_id`,`wh_id`) COMMENT '唯一索引，避免同一物料同一仓库重复记录',
  KEY `idx_wh_id` (`wh_id`) COMMENT '仓库 ID 索引',
  KEY `idx_stock_product_id` (`product_id`),
  CONSTRAINT `fk_stock_material` FOREIGN KEY (`material_id`) REFERENCES `material` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_stock_warehouse` FOREIGN KEY (`wh_id`) REFERENCES `warehouse` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=23 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='库存表';

-- ----------------------------
-- Table structure for supplier
-- ----------------------------
DROP TABLE IF EXISTS `supplier`;
CREATE TABLE `supplier` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '供应商 ID',
  `sup_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商编号（唯一）',
  `sup_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '供应商名称',
  `contact` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系人',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '电子邮箱',
  `address` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '供应商地址',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `status` tinyint DEFAULT '1' COMMENT '状态：0=停用，1=正常',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sup_code` (`sup_code`) COMMENT '供应商编号唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='供应商表';

-- ----------------------------
-- Table structure for sys_company
-- ----------------------------
DROP TABLE IF EXISTS `sys_company`;
CREATE TABLE `sys_company` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `company_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '公司全称',
  `short_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '公司简称',
  `english_name` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '英文全称',
  `english_short_name` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '英文简称',
  `logo` varchar(500) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '公司 logo URL（建议尺寸 336*44px）',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公司信息表';

-- ----------------------------
-- Table structure for sys_dept
-- ----------------------------
DROP TABLE IF EXISTS `sys_dept`;
CREATE TABLE `sys_dept` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '部门 ID',
  `dept_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '部门名称',
  `parent_id` bigint DEFAULT '0' COMMENT '上级部门 ID，0 表示顶级部门',
  `sort` int DEFAULT '0' COMMENT '部门排序',
  `leader` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '部门负责人',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `status` tinyint DEFAULT '1' COMMENT '部门状态：0=停用，1=正常',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='部门表';

-- ----------------------------
-- Table structure for sys_dict_data
-- ----------------------------
DROP TABLE IF EXISTS `sys_dict_data`;
CREATE TABLE `sys_dict_data` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '字典数据 ID',
  `dict_type_id` bigint NOT NULL COMMENT '字典类型 ID（外键）',
  `dict_label` varchar(100) NOT NULL COMMENT '字典显示标签',
  `dict_value` varchar(100) NOT NULL COMMENT '字典存储值',
  `color` varchar(20) DEFAULT 'default' COMMENT '颜色标签',
  `sort` int DEFAULT '0' COMMENT '排序',
  `status` tinyint DEFAULT '1' COMMENT '状态：1=启用，0=禁用',
  `is_locked` tinyint DEFAULT '0' COMMENT '锁定级别：0=未锁定，1=可解锁锁定，2=永久锁定（不可删除）',
  `allow_production` tinyint DEFAULT '0' COMMENT '是否可用于生产：1=是，0=否（仅模具状态字典有效）',
  `allow_work_report` bigint NOT NULL DEFAULT '1' COMMENT '是否在生产报工页面展示：1=是，0=否',
  `needs_delivery_unlock` tinyint DEFAULT '0' COMMENT '是否需要交付后解锁：1=是，0=否（仅款项类型字典有效）',
  `remark` varchar(255) DEFAULT '' COMMENT '备注',
  `ext` text COMMENT '扩展配置（JSON格式，如付款方式对应的可用账户类型）',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dict_type_label` (`dict_type_id`,`dict_label`),
  UNIQUE KEY `uk_dict_type_value` (`dict_type_id`,`dict_value`),
  KEY `idx_dict_type_id` (`dict_type_id`) COMMENT '字典类型 ID 索引',
  KEY `idx_sort` (`sort`) COMMENT '排序索引',
  KEY `idx_sys_dict_data_type_sort_status` (`dict_type_id`,`sort`,`status`),
  KEY `idx_status` (`status`),
  CONSTRAINT `fk_dict_data_type` FOREIGN KEY (`dict_type_id`) REFERENCES `sys_dict_type` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=231 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='字典数据表';

-- ----------------------------
-- Table structure for sys_dict_type
-- ----------------------------
DROP TABLE IF EXISTS `sys_dict_type`;
CREATE TABLE `sys_dict_type` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '字典类型 ID',
  `dict_name` varchar(100) NOT NULL COMMENT '字典名称',
  `dict_code` varchar(50) NOT NULL COMMENT '字典编码（如：mold-status）',
  `remark` varchar(255) DEFAULT '' COMMENT '备注',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_dict_code` (`dict_code`) COMMENT '字典编码唯一索引',
  KEY `idx_sys_dict_type_id` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=57 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='字典类型表';

-- ----------------------------
-- Table structure for sys_menu
-- ----------------------------
DROP TABLE IF EXISTS `sys_menu`;
CREATE TABLE `sys_menu` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
  `parent_id` bigint DEFAULT '0' COMMENT '父菜单ID，0=顶级',
  `menu_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '菜单名称',
  `menu_type` tinyint DEFAULT '1' COMMENT '1=目录 2=菜单 3=按钮 4=页面元素',
  `permission` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '权限标识，如 system:user:list',
  `path` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '路由路径',
  `component` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '组件路径',
  `icon` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '图标',
  `sort` int DEFAULT '0' COMMENT '排序',
  `status` tinyint DEFAULT '1' COMMENT '0=禁用 1=启用',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permission` (`permission`) COMMENT '权限标识唯一',
  KEY `idx_parent_id` (`parent_id`) COMMENT '父菜单索引'
) ENGINE=InnoDB AUTO_INCREMENT=88 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='菜单权限表';

-- ----------------------------
-- Table structure for sys_notification
-- ----------------------------
DROP TABLE IF EXISTS `sys_notification`;
CREATE TABLE `sys_notification` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `user_id` bigint NOT NULL COMMENT '接收用户 ID',
  `title` varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '通知标题',
  `content` varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '通知内容',
  `type` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '通知类型（如 mold:work_order_status、production:order_status 等）',
  `type_label` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '类型显示名称（如 模具工单、生产订单）',
  `quality` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'low' COMMENT '优先级：low=低 middle=中 high=高',
  `ref_id` bigint DEFAULT '0' COMMENT '关联业务记录 ID',
  `ref_url` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT '' COMMENT '关联跳转链接',
  `is_read` tinyint NOT NULL DEFAULT '0' COMMENT '是否已读：0=未读 1=已读',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`) COMMENT '用户索引',
  KEY `idx_user_read` (`user_id`,`is_read`) COMMENT '用户未读查询索引',
  KEY `idx_create_time` (`create_time`) COMMENT '时间排序索引'
) ENGINE=InnoDB AUTO_INCREMENT=102 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统通知表';

-- ----------------------------
-- Table structure for sys_notification_config
-- ----------------------------
DROP TABLE IF EXISTS `sys_notification_config`;
CREATE TABLE `sys_notification_config` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `type_pattern` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '通知类型匹配模式，支持 LIKE 通配符，如 mold:% 匹配所有模具通知',
  `target_type` varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '目标类型：role=角色 user=用户',
  `target_id` bigint NOT NULL COMMENT '目标 ID（角色 ID 或用户 ID）',
  `enabled` tinyint NOT NULL DEFAULT '1' COMMENT '是否启用：1=启用 0=禁用',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_type_target` (`type_pattern`,`target_type`,`target_id`),
  KEY `idx_type_pattern` (`type_pattern`),
  KEY `idx_target` (`target_type`,`target_id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知配置表';

-- ----------------------------
-- Table structure for sys_oper_log
-- ----------------------------
DROP TABLE IF EXISTS `sys_oper_log`;
CREATE TABLE `sys_oper_log` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `title` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作模块标题',
  `oper_type` tinyint NOT NULL DEFAULT '0' COMMENT '操作类型：1=增 2=删 3=改 4=查 5=导入 6=导出 7=登录 8=登出',
  `business_type` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '业务类型标识',
  `business_id` bigint NOT NULL DEFAULT '0' COMMENT '业务记录ID',
  `method` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求方法（路由路径）',
  `request_url` varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求URL',
  `request_method` varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '请求方式 GET/POST/PUT/DELETE',
  `oper_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作人员',
  `dept_name` varchar(64) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '部门名称',
  `oper_ip` varchar(128) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '操作IP',
  `request_params` text COLLATE utf8mb4_unicode_ci COMMENT '请求参数',
  `json_result` text COLLATE utf8mb4_unicode_ci COMMENT '返回结果',
  `status` tinyint NOT NULL DEFAULT '0' COMMENT '操作状态：0=正常 1=异常',
  `error_msg` varchar(2048) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '' COMMENT '错误消息',
  `cost_time` bigint NOT NULL DEFAULT '0' COMMENT '耗时(ms)',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_oper_name` (`oper_name`) COMMENT '按操作人查询',
  KEY `idx_oper_type` (`oper_type`) COMMENT '按操作类型查询',
  KEY `idx_title` (`title`) COMMENT '按模块查询',
  KEY `idx_business_type` (`business_type`) COMMENT '按业务类型查询',
  KEY `idx_status` (`status`) COMMENT '按状态查询',
  KEY `idx_create_time` (`create_time`) COMMENT '按时间排序查询',
  KEY `idx_oper_log_business` (`business_type`,`business_id`)
) ENGINE=InnoDB AUTO_INCREMENT=69 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志记录';

-- ----------------------------
-- Table structure for sys_role
-- ----------------------------
DROP TABLE IF EXISTS `sys_role`;
CREATE TABLE `sys_role` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '角色 ID',
  `role_name` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '角色名称',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '角色备注',
  `role_type` tinyint DEFAULT '0' COMMENT '0=自定义角色 1=系统内置',
  `data_scope` tinyint DEFAULT '1' COMMENT '1=全部数据 2=本部门及以下 3=本部门 4=仅本人',
  `status` tinyint DEFAULT '1' COMMENT '0=禁用 1=启用',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_name` (`role_name`) COMMENT '角色名唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- ----------------------------
-- Table structure for sys_role_menu
-- ----------------------------
DROP TABLE IF EXISTS `sys_role_menu`;
CREATE TABLE `sys_role_menu` (
  `role_id` bigint NOT NULL COMMENT '角色ID',
  `menu_id` bigint NOT NULL COMMENT '菜单ID',
  PRIMARY KEY (`role_id`,`menu_id`),
  KEY `idx_role_menu_menu` (`menu_id`),
  CONSTRAINT `fk_role_menu_menu` FOREIGN KEY (`menu_id`) REFERENCES `sys_menu` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_role_menu_role` FOREIGN KEY (`role_id`) REFERENCES `sys_role` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色菜单关联表';

-- ----------------------------
-- Table structure for sys_security_config
-- ----------------------------
DROP TABLE IF EXISTS `sys_security_config`;
CREATE TABLE `sys_security_config` (
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
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='安全配置表';

-- ----------------------------
-- Table structure for sys_system_info
-- ----------------------------
DROP TABLE IF EXISTS `sys_system_info`;
CREATE TABLE `sys_system_info` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `system_name` varchar(100) NOT NULL DEFAULT '' COMMENT '系统名称',
  `system_version` varchar(50) NOT NULL DEFAULT '' COMMENT '系统版本',
  `system_description` varchar(500) NOT NULL DEFAULT '' COMMENT '系统简介',
  `tech_stack` varchar(500) NOT NULL DEFAULT '' COMMENT '技术栈',
  `favicon_url` varchar(500) NOT NULL DEFAULT '' COMMENT 'Favicon图标URL',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP,
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='系统信息配置表';

-- ----------------------------
-- Table structure for sys_user
-- ----------------------------
DROP TABLE IF EXISTS `sys_user`;
CREATE TABLE `sys_user` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
  `username` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '登录账号',
  `password` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '密码（加密存储）',
  `real_name` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '真实姓名',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '手机号',
  `email` varchar(100) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '邮箱',
  `position` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '职位',
  `hire_date` date DEFAULT NULL COMMENT '入职日期',
  `dept_id` bigint DEFAULT NULL COMMENT '关联部门 ID',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '用户状态：0=禁用, 1=正常',
  `login_attempts` int NOT NULL DEFAULT '0' COMMENT '连续登录失败次数',
  `locked_until` datetime DEFAULT NULL COMMENT '账号锁定截止时间，NULL表示未锁定',
  `password_version` int NOT NULL DEFAULT '1' COMMENT '密码版本号，修改密码时自增，旧 Token 自动失效',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`) COMMENT '用户名唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统用户表';

-- ----------------------------
-- Table structure for sys_user_role
-- ----------------------------
DROP TABLE IF EXISTS `sys_user_role`;
CREATE TABLE `sys_user_role` (
  `user_id` bigint NOT NULL COMMENT '用户 ID',
  `role_id` bigint NOT NULL COMMENT '角色 ID',
  PRIMARY KEY (`user_id`,`role_id`) COMMENT '联合主键',
  KEY `fk_user_role_role` (`role_id`),
  CONSTRAINT `fk_user_role_role` FOREIGN KEY (`role_id`) REFERENCES `sys_role` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_user_role_user` FOREIGN KEY (`user_id`) REFERENCES `sys_user` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- ----------------------------
-- Table structure for warehouse
-- ----------------------------
DROP TABLE IF EXISTS `warehouse`;
CREATE TABLE `warehouse` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '仓库 ID',
  `wh_code` varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '仓库编码（唯一）',
  `wh_name` varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '仓库名称',
  `address` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '仓库地址',
  `manager` varchar(50) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '仓库管理员',
  `phone` varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '联系电话',
  `remark` varchar(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT '备注',
  `status` tinyint DEFAULT '1' COMMENT '仓库状态：0=停用，1=正常',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_wh_code` (`wh_code`) COMMENT '仓库编码唯一索引'
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='仓库表';

-- ----------------------------
-- Table structure for work_order_process
-- ----------------------------
DROP TABLE IF EXISTS `work_order_process`;
CREATE TABLE `work_order_process` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `work_order_id` bigint NOT NULL COMMENT '工单ID',
  `process_template_id` bigint NOT NULL COMMENT '工序模板ID',
  `is_completed` tinyint NOT NULL DEFAULT '0' COMMENT '工序状态：0=未完成，1=已完成，2=已跳过',
  `completed_at` datetime DEFAULT NULL COMMENT '完成时间',
  `remark` varchar(500) DEFAULT NULL COMMENT '备注',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_work_process` (`work_order_id`,`process_template_id`),
  KEY `idx_work_order` (`work_order_id`),
  KEY `idx_template` (`process_template_id`),
  KEY `idx_completed` (`is_completed`),
  CONSTRAINT `work_order_process_ibfk_1` FOREIGN KEY (`work_order_id`) REFERENCES `mold_work_order` (`id`) ON DELETE CASCADE,
  CONSTRAINT `work_order_process_ibfk_2` FOREIGN KEY (`process_template_id`) REFERENCES `process_template` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=672 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='工单工序关联表';

-- ----------------------------
-- Table structure for sys_migration
-- ----------------------------
CREATE TABLE `sys_migration` (
  `id`          bigint       NOT NULL AUTO_INCREMENT,
  `version`     varchar(32)  NOT NULL COMMENT '迁移版本号，如 v0.2.0',
  `filename`    varchar(255) NOT NULL COMMENT '迁移文件名',
  `executed_at` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '执行时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_version` (`version`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据库迁移版本追踪表';
