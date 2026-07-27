package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type MoldWorkOrderDAO struct{}

func NewMoldWorkOrderDAO() *MoldWorkOrderDAO {
	return &MoldWorkOrderDAO{}
}

func (d *MoldWorkOrderDAO) Create(workOrder *entity.MoldWorkOrder) error {
	return DB.Create(workOrder).Error
}

func (d *MoldWorkOrderDAO) FindByID(id int64) (*entity.MoldWorkOrder, error) {
	var workOrder entity.MoldWorkOrder
	// 使用 Preload 通过 LEFT JOIN 一次性加载所有关联（替代原先 4 次独立查询）
	err := DB.Preload("Mold.Customer").Preload("MoldOrder").First(&workOrder, id).Error
	if err != nil {
		return nil, err
	}
	return &workOrder, nil
}

func (d *MoldWorkOrderDAO) FindByCode(code string) (*entity.MoldWorkOrder, error) {
	var workOrder entity.MoldWorkOrder
	err := DB.Where("work_code = ?", code).First(&workOrder).Error
	if err != nil {
		return nil, err
	}
	return &workOrder, nil
}

// FindByMoldOrderIDAndMoldID 根据模具订单ID和模具ID查找工单
func (d *MoldWorkOrderDAO) FindByMoldOrderIDAndMoldID(moldOrderID int64, moldID *int64) (*entity.MoldWorkOrder, error) {
	var workOrder entity.MoldWorkOrder
	query := DB.Where("mold_order_id = ?", moldOrderID)
	if moldID != nil {
		query = query.Where("mold_id = ?", *moldID)
	}
	err := query.First(&workOrder).Error
	if err != nil {
		return nil, err
	}
	return &workOrder, nil
}

// FindByMoldOrderID 根据订单ID查询所有工单
func (d *MoldWorkOrderDAO) FindByMoldOrderID(moldOrderID int64) ([]*entity.MoldWorkOrder, error) {
	var workOrders []*entity.MoldWorkOrder
	err := DB.Where("mold_order_id = ?", moldOrderID).Order("id ASC").Find(&workOrders).Error
	return workOrders, err
}

// FindByCodeWithTx 在指定事务中根据工单号查找工单
func (d *MoldWorkOrderDAO) FindByCodeWithTx(tx *gorm.DB, code string) (*entity.MoldWorkOrder, error) {
	var workOrder entity.MoldWorkOrder
	err := tx.Where("work_code = ?", code).First(&workOrder).Error
	if err != nil {
		return nil, err
	}
	return &workOrder, nil
}

func (d *MoldWorkOrderDAO) FindAll() ([]*entity.MoldWorkOrder, error) {
	var workOrders []*entity.MoldWorkOrder
	err := DB.Select("id", "work_code", "mold_id", "mold_order_id", "status").Order("id DESC").Limit(1000).Find(&workOrders).Error
	return workOrders, err
}

func (d *MoldWorkOrderDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var workOrders []*entity.MoldWorkOrder
	result, err := utils.Paginate(DB, page, pageSize, &workOrders, func(db *gorm.DB) *gorm.DB {
		return db.Preload("Mold", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "mold_code", "mold_name", "customer_id")
		}).Preload("MoldOrder", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "order_code", "status")
		}).Order("id DESC")
	})
	if err != nil {
		return nil, err
	}

	// 批量加载 Customer（一次查询）
	workOrders = result.List.([]*entity.MoldWorkOrder)
	if len(workOrders) > 0 {
		customerIDs := make([]int64, 0)
		customerIDSet := make(map[int64]bool)
		for _, wo := range workOrders {
			if wo.Mold != nil && wo.Mold.CustomerID > 0 && !customerIDSet[wo.Mold.CustomerID] {
				customerIDs = append(customerIDs, wo.Mold.CustomerID)
				customerIDSet[wo.Mold.CustomerID] = true
			}
		}

		if len(customerIDs) > 0 {
			var customers []*entity.Customer
			DB.Select("id", "cus_name").Where("id IN ?", customerIDs).Find(&customers)

			// 填充 Customer 数据
			customerMap := make(map[int64]*entity.Customer)
			for _, c := range customers {
				customerMap[c.ID] = c
			}
			for _, wo := range workOrders {
				if wo.Mold != nil && wo.Mold.CustomerID > 0 {
					wo.Mold.Customer = customerMap[wo.Mold.CustomerID]
				}
			}
		}
	}

	result.List = workOrders
	return result, nil
}

// FindPageWithFilter 带筛选条件的分页查询模具工单
func (d *MoldWorkOrderDAO) FindPageWithFilter(offset, limit int, workCode string, moldCode string, customerID *int64, status *int8, orderType string, processStage *string) ([]*entity.MoldWorkOrder, int64, error) {
	var workOrders []*entity.MoldWorkOrder
	var total int64

	query := DB.Model(&entity.MoldWorkOrder{})

	if workCode != "" {
		query = query.Where("work_code LIKE ?", workCode+"%")
	}
	// 添加模具编号筛选条件（使用子查询避免影响 COUNT）
	if moldCode != "" {
		subQuery := DB.Model(&entity.Mold{}).Where("mold_code LIKE ?", moldCode+"%").Select("id")
		query = query.Where("mold_id IN (?)", subQuery)
	}
	if customerID != nil {
		subQuery := DB.Model(&entity.Mold{}).Where("customer_id = ?", *customerID).Select("id")
		query = query.Where("mold_id IN (?)", subQuery)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if orderType != "" {
		query = query.Where("order_type = ?", orderType)
	}
	if processStage != nil && *processStage != "" {
		// 根据工序阶段名称查找对应的工序模板ID
		var templateIDs []int64
		DB.Model(&entity.ProcessTemplate{}).
			Where("process_name LIKE ? OR group_name LIKE ?", "%"+*processStage+"%", "%"+*processStage+"%").
			Pluck("id", &templateIDs)

		if len(templateIDs) > 0 {
			// 查找包含这些工序模板且未完成的工单
			subQuery := DB.Model(&entity.WorkOrderProcess{}).
				Where("process_template_id IN ? AND is_completed = 0", templateIDs).
				Select("work_order_id")
			query = query.Where("id IN (?)", subQuery)
		}
	}

	query.Count(&total)
	// 优化：只预加载必要的关联数据
	err := query.Preload("Mold", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "mold_code", "mold_name", "customer_id") // 只查询需要的字段
	}).Preload("MoldOrder", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "order_code", "status") // 只查询需要的字段
	}).Offset(offset).Limit(limit).Order("id DESC").Find(&workOrders).Error
	if err != nil {
		return nil, 0, err
	}

	// 批量加载 Customer（一次查询）
	if len(workOrders) > 0 {
		customerIDs := make([]int64, 0)
		customerIDSet := make(map[int64]bool)
		for _, wo := range workOrders {
			if wo.Mold != nil && wo.Mold.CustomerID > 0 && !customerIDSet[wo.Mold.CustomerID] {
				customerIDs = append(customerIDs, wo.Mold.CustomerID)
				customerIDSet[wo.Mold.CustomerID] = true
			}
		}

		if len(customerIDs) > 0 {
			var customers []*entity.Customer
			DB.Select("id", "cus_name").Where("id IN ?", customerIDs).Find(&customers) // 只查询需要的字段

			// 填充 Customer 数据
			customerMap := make(map[int64]*entity.Customer)
			for _, c := range customers {
				customerMap[c.ID] = c
			}
			for _, wo := range workOrders {
				if wo.Mold != nil && wo.Mold.CustomerID > 0 {
					wo.Mold.Customer = customerMap[wo.Mold.CustomerID]
				}
			}
		}
	}

	return workOrders, total, nil
}

func (d *MoldWorkOrderDAO) Update(workOrder *entity.MoldWorkOrder) error {
	return DB.Model(&entity.MoldWorkOrder{}).Where("id = ?", workOrder.ID).Updates(workOrder).Error
}

func (d *MoldWorkOrderDAO) Updates(id int64, updates map[string]interface{}) error {
	return DB.Model(&entity.MoldWorkOrder{}).Where("id = ?", id).Updates(updates).Error
}

func (d *MoldWorkOrderDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldWorkOrder{}, id).Error
}

func (d *MoldWorkOrderDAO) UpdateStatus(id int64, status int8) error {
	return DB.Model(&entity.MoldWorkOrder{}).Where("id = ?", id).Update("status", status).Error
}

// FindMaxSequenceByPrefix 查找指定前缀的最大序列号
func (d *MoldWorkOrderDAO) FindMaxSequenceByPrefix(prefix string) (int, error) {
	var maxSeq int
	err := DB.Model(&entity.MoldWorkOrder{}).
		Where("work_code LIKE ?", prefix+"%").
		Select("COALESCE(MAX(CAST(SUBSTRING_INDEX(work_code, '-', -1) AS UNSIGNED)), 0)").
		Scan(&maxSeq).Error
	return maxSeq, err
}

// FindByIDs 根据ID列表批量查询工单
func (d *MoldWorkOrderDAO) FindByIDs(ids []int64) ([]*entity.MoldWorkOrder, error) {
	var workOrders []*entity.MoldWorkOrder
	err := DB.Where("id IN ?", ids).Find(&workOrders).Error
	return workOrders, err
}
