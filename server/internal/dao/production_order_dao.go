package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type ProductionOrderDAO struct{}

// NewProductionOrderDAO 创建生产订单DAO
func NewProductionOrderDAO() *ProductionOrderDAO {
	return &ProductionOrderDAO{}
}

// Create 创建生产订单
func (d *ProductionOrderDAO) Create(order *entity.ProductionOrder) error {
	return GetDB().Create(order).Error
}

// FindByID 根据ID查询生产订单
func (d *ProductionOrderDAO) FindByID(id int64) (*entity.ProductionOrder, error) {
	var order entity.ProductionOrder
	err := GetDB().Preload("Customer").Preload("Mold").Preload("Processor").Preload("Items.Product").Preload("Items.Mold").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	// 手动加载多模具关联数据（Molds字段）
	if len(order.Items) > 0 {
		d.loadItemsMolds([]entity.ProductionOrder{order})
	}
	return &order, nil
}

// FindByCode 根据订单编号查询
func (d *ProductionOrderDAO) FindByCode(code string) (*entity.ProductionOrder, error) {
	var order entity.ProductionOrder
	err := GetDB().Where("order_code = ?", code).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// Update 更新生产订单
func (d *ProductionOrderDAO) Update(order *entity.ProductionOrder) error {
	// 使用 Updates 只更新指定字段，避免级联保存关联对象
	return GetDB().Model(&entity.ProductionOrder{}).Where("id = ?", order.ID).Updates(order).Error
}

// Delete 删除生产订单（软删除通过状态控制）
func (d *ProductionOrderDAO) Delete(id int64) error {
	return GetDB().Model(&entity.ProductionOrder{}).Where("id = ?", id).Update("status", 9).Error
}

// FindByPage 分页查询生产订单（用于列表展示）
func (d *ProductionOrderDAO) FindByPage(query dto.ProductionOrderQueryDTO) (*utils.PageResult, error) {
	var orders []entity.ProductionOrder
	db := GetDB().Model(&entity.ProductionOrder{})

	if query.Keyword != "" {
		// 关键词搜索：匹配订单号、订单级产品名称、明细级产品名称、客户名称
		likePattern := "%" + query.Keyword + "%"
		db = db.Where(
			"order_code LIKE ? OR product_name LIKE ? OR customer_id IN (SELECT id FROM customer WHERE cus_name LIKE ?) OR id IN (SELECT order_id FROM production_order_item WHERE product_name LIKE ?)",
			query.Keyword+"%",
			likePattern,
			likePattern,
			likePattern,
		)
	}
	if query.OrderCode != "" {
		db = db.Where("order_code LIKE ?", query.OrderCode+"%")
	}
	if query.CustomerID != nil {
		db = db.Where("customer_id = ?", *query.CustomerID)
	}
	if query.ProductName != "" {
		// 同时搜索订单级产品名称和明细级产品名称
		db = db.Where(
			"product_name LIKE ? OR id IN (SELECT order_id FROM production_order_item WHERE product_name LIKE ?)",
			"%"+query.ProductName+"%",
			"%"+query.ProductName+"%",
		)
	}
	if query.ProductionType != nil {
		db = db.Where("production_type = ?", *query.ProductionType)
	}
	if query.MoldSourceType != nil {
		db = db.Where("mold_source_type = ?", *query.MoldSourceType)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.ItemStatus != nil {
		// 按明细项状态过滤（报工页面使用）
		// 报工页面的卡片展示状态逻辑：
		//   - 订单主状态为 PENDING_SCHEDULE 时显示「待排产」（忽略明细项状态）
		//   - 其他情况显示明细项状态
		//   - 匹配订单主状态（如 PENDING_SCHEDULE）
		//   - 匹配非待排产订单下的明细项状态（排除待排产订单，因为其明细项默认状态也是 PRODUCING）
		db = db.Where(
			`status = ? OR id IN (
				SELECT poi.order_id FROM production_order_item poi
				INNER JOIN production_order po ON po.id = poi.order_id
				WHERE po.status != 'PENDING_SCHEDULE' AND poi.status = ?
			)`,
			*query.ItemStatus, *query.ItemStatus,
		)
	}
	if query.CreateTimeStart != nil && *query.CreateTimeStart != "" {
		db = db.Where("create_time >= ?", *query.CreateTimeStart)
	}
	if query.CreateTimeEnd != nil && *query.CreateTimeEnd != "" {
		db = db.Where("create_time <= ?", *query.CreateTimeEnd+" 23:59:59")
	}

	result, err := utils.Paginate(db, query.Page, query.PageSize, &orders, func(db *gorm.DB) *gorm.DB {
		return db.Preload("MasterOrder").Preload("Customer").Preload("Mold").Preload("Items.Product").Preload("Items.Mold").Order("create_time DESC")
	})

	if err == nil && len(orders) > 0 {
		// 手动加载多模具关联数据（Molds字段）
		d.loadItemsMolds(orders)
	}

	return result, err
}

// FindByPageForWorkReport 分页查询生产订单（用于生产报工页面，只查询允许报工的状态）
func (d *ProductionOrderDAO) FindByPageForWorkReport(query dto.ProductionOrderQueryDTO) ([]entity.ProductionOrder, int64, error) {
	var orders []entity.ProductionOrder
	var total int64

	db := GetDB().Model(&entity.ProductionOrder{})

	if query.OrderCode != "" {
		db = db.Where("order_code LIKE ?", query.OrderCode+"%")
	}
	if query.CustomerID != nil {
		db = db.Where("customer_id = ?", *query.CustomerID)
	}
	if query.ProductName != "" {
		// 同时搜索订单级产品名称和明细级产品名称
		db = db.Where(
			"product_name LIKE ? OR id IN (SELECT order_id FROM production_order_item WHERE product_name LIKE ?)",
			"%"+query.ProductName+"%",
			"%"+query.ProductName+"%",
		)
	}
	if query.ProductionType != nil {
		db = db.Where("production_type = ?", *query.ProductionType)
	}
	if query.MoldSourceType != nil {
		db = db.Where("mold_source_type = ?", *query.MoldSourceType)
	}
	if query.Status != nil {
		db = db.Where("status = ?", *query.Status)
	}
	if query.ItemStatus != nil {
		// 按明细项状态过滤（报工页面使用）
		// 注意：报工页面的卡片展示状态逻辑是：
		//   - 订单主状态为 PENDING_SCHEDULE 时显示「待排产」
		//   - 其他情况显示明细项状态
		// 因此过滤条件要同时匹配订单主状态和明细项状态，否则「待排产」等订单级状态会筛不到
		db = db.Where(
			`status = ? OR id IN (
				SELECT poi.order_id FROM production_order_item poi
				INNER JOIN production_order po ON po.id = poi.order_id
				WHERE po.status != 'PENDING_SCHEDULE' AND poi.status = ?
			)`,
			*query.ItemStatus, *query.ItemStatus,
		)
	}
	if query.Priority != nil {
		db = db.Where("priority = ?", *query.Priority)
	}
	if query.PlanEndDateStart != nil && *query.PlanEndDateStart != "" {
		db = db.Where("plan_end_date >= ?", *query.PlanEndDateStart)
	}
	if query.PlanEndDateEnd != nil && *query.PlanEndDateEnd != "" {
		db = db.Where("plan_end_date <= ?", *query.PlanEndDateEnd)
	}

	// 通过子查询关联 sys_dict_data 表，检查状态的 allow_work_report 字段
	db = db.Where(`status IN (
		SELECT CAST(dict_value AS UNSIGNED) 
		FROM sys_dict_data 
		WHERE dict_type_id = (SELECT id FROM sys_dict_type WHERE dict_code = 'production-order-status')
		AND allow_work_report = 1
	)`)

	// ✨ 新增：自有模具必须达到就绪状态才能报工
	// 对于 mold_source_type = 'SELF_MADE' 的订单，只查询 dependency_status = 'READY' 的明细项
	db = db.Where(`mold_source_type = 'CUSTOMER_PROVIDED' 
		OR id IN (
			SELECT DISTINCT poi.order_id 
			FROM production_order_item poi 
			WHERE poi.dependency_status = 'READY' OR poi.dependency_status = 'NONE'
		)
		OR id NOT IN (
			SELECT DISTINCT poi2.order_id 
			FROM production_order_item poi2 
			WHERE poi2.dependency_status = 'WAITING_MOLD'
		)
	)`)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.Preload("Customer").Preload("Mold").Preload("Items.Product").Preload("Items.Mold").
		Offset(offset).Limit(query.PageSize).
		Order("create_time DESC").
		Find(&orders).Error

	// 手动加载多模具关联数据（Molds字段）
	if err == nil && len(orders) > 0 {
		d.loadItemsMolds(orders)
	}

	return orders, total, err
}

// loadItemsMolds 手动加载订单明细的多模具关联数据
func (d *ProductionOrderDAO) loadItemsMolds(orders []entity.ProductionOrder) {
	moldIDSet := make(map[int64]bool)
	for _, order := range orders {
		for _, item := range order.Items {
			if len(item.MoldIDs) > 0 {
				for _, moldID := range item.MoldIDs {
					if moldID > 0 {
						moldIDSet[moldID] = true
					}
				}
			}
		}
	}

	// 如果没有多模具数据，直接返回
	if len(moldIDSet) == 0 {
		return
	}

	// 批量查询模具
	moldIDs := make([]int64, 0, len(moldIDSet))
	for id := range moldIDSet {
		moldIDs = append(moldIDs, id)
	}

	var molds []entity.Mold
	if err := GetDB().Where("id IN ?", moldIDs).Find(&molds).Error; err != nil {
		return
	}

	moldMap := make(map[int64]*entity.Mold)
	for i := range molds {
		moldMap[molds[i].ID] = &molds[i]
	}

	// 为每个订单明细填充 Molds 字段
	for i := range orders {
		for j := range orders[i].Items {
			if len(orders[i].Items[j].MoldIDs) > 0 {
				orders[i].Items[j].Molds = make([]*entity.Mold, 0)
				for _, moldID := range orders[i].Items[j].MoldIDs {
					if mold, ok := moldMap[moldID]; ok {
						orders[i].Items[j].Molds = append(orders[i].Items[j].Molds, mold)
					}
				}
			}
		}
	}
}

// GetStatistics 获取订单统计
func (d *ProductionOrderDAO) GetStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// 总订单数
	var total int64
	if err := GetDB().Model(&entity.ProductionOrder{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 各状态订单数
	type StatusCount struct {
		Status string
		Count  int64
	}
	var statusCounts []StatusCount
	if err := GetDB().Model(&entity.ProductionOrder{}).Select("status, count(*) as count").Group("status").Scan(&statusCounts).Error; err != nil {
		return nil, err
	}
	for _, sc := range statusCounts {
		key := "status_" + sc.Status
		stats[key] = sc.Count
	}

	return stats, nil
}
