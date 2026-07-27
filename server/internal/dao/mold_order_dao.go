package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type MoldOrderDAO struct{}

// NewMoldOrderDAO 创建模具订单 DAO 实例
func NewMoldOrderDAO() *MoldOrderDAO {
	return &MoldOrderDAO{}
}

// Create 创建模具订单
func (d *MoldOrderDAO) Create(order *entity.MoldOrder) error {
	return DB.Create(order).Error
}

// FindByID 根据 ID 查询模具订单（预加载客户、模具、综合订单）
func (d *MoldOrderDAO) FindByID(id int64) (*entity.MoldOrder, error) {
	var order entity.MoldOrder
	err := DB.Preload("Customer").Preload("Mold").Preload("MasterOrder").First(&order, id).Error
	if err != nil {
		return nil, err
	}

	// 填充 MasterOrderCode
	if order.MasterOrder != nil {
		order.MasterOrderCode = order.MasterOrder.OrderCode
	}

	return &order, nil
}

// FindByCode 根据编号查询模具订单
func (d *MoldOrderDAO) FindByCode(code string) (*entity.MoldOrder, error) {
	var order entity.MoldOrder
	err := DB.Where("order_code = ?", code).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// FindAll 查询所有模具订单（下拉选择框用，限制 1000 条）
func (d *MoldOrderDAO) FindAll() ([]*entity.MoldOrder, error) {
	var orders []*entity.MoldOrder
	err := DB.Select("id", "order_code", "customer_id", "status").Order("id DESC").Limit(1000).Find(&orders).Error
	return orders, err
}

// FindPage 分页查询模具订单
func (d *MoldOrderDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var orders []*entity.MoldOrder
	return utils.Paginate(DB, page, pageSize, &orders, func(db *gorm.DB) *gorm.DB {
		return db.Preload("Customer").Preload("Mold").Preload("MasterOrder").Order("id DESC")
	})
}

// FindPageWithFilter 带筛选条件的分页查询模具订单
func (d *MoldOrderDAO) FindPageWithFilter(offset, limit int, orderCode string, moldCode string, customerID *int64, status *string, createTimeStart, createTimeEnd, planEndDateStart, planEndDateEnd string) ([]*entity.MoldOrder, int64, error) {
	var orders []*entity.MoldOrder
	var total int64

	query := DB.Model(&entity.MoldOrder{})

	if orderCode != "" {
		query = query.Where("order_code LIKE ?", orderCode+"%")
	}
	// 订单通过工单表（mold_work_order）关联模具，而非 mold_order.mold_id（该字段可能为空）
	if moldCode != "" {
		subQuery := DB.Table("mold_work_order").
			Select("DISTINCT mold_work_order.mold_order_id").
			Joins("LEFT JOIN mold ON mold.id = mold_work_order.mold_id").
			Where("mold.mold_code LIKE ?", moldCode+"%")
		query = query.Where("mold_order.id IN (?)", subQuery)
	}
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if createTimeStart != "" {
		query = query.Where("create_time >= ?", createTimeStart)
	}
	if createTimeEnd != "" {
		query = query.Where("create_time <= ?", createTimeEnd+" 23:59:59")
	}
	if planEndDateStart != "" {
		query = query.Where("plan_end_date >= ?", planEndDateStart)
	}
	if planEndDateEnd != "" {
		query = query.Where("plan_end_date <= ?", planEndDateEnd)
	}

	query.Count(&total)
	err := query.Preload("Customer").Preload("Mold").Preload("MasterOrder").Offset(offset).Limit(limit).Order("id DESC").Find(&orders).Error

	// 填充 MasterOrderCode
	if err == nil {
		for _, order := range orders {
			if order.MasterOrder != nil {
				order.MasterOrderCode = order.MasterOrder.OrderCode
			}
		}
	}

	return orders, total, err
}

// Update 更新模具订单
func (d *MoldOrderDAO) Update(order *entity.MoldOrder) error {
	// 使用 Updates 只更新非零值字段，避免覆盖 create_time
	return DB.Model(&entity.MoldOrder{}).Where("id = ?", order.ID).Updates(order).Error
}

// Delete 删除模具订单
func (d *MoldOrderDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldOrder{}, id).Error
}

// UpdateStatus 更新模具订单状态
func (d *MoldOrderDAO) UpdateStatus(id int64, status int8) error {
	return DB.Model(&entity.MoldOrder{}).Where("id = ?", id).Update("status", status).Error
}

// FindMaxOrderSequenceByPrefix 查找指定前缀的最大序列号
func (d *MoldOrderDAO) FindMaxOrderSequenceByPrefix(prefix string) (int, error) {
	var maxSeq int
	err := DB.Model(&entity.MoldOrder{}).
		Where("order_code LIKE ?", prefix+"%").
		Select("COALESCE(MAX(CAST(SUBSTRING_INDEX(order_code, '-', -1) AS UNSIGNED)), 0)").
		Scan(&maxSeq).Error
	return maxSeq, err
}
