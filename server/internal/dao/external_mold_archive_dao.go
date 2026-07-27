package dao

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type ExternalMoldArchiveDAO struct {
	db *gorm.DB
}

// NewExternalMoldArchiveDAO 创建外来模具档案DAO
func NewExternalMoldArchiveDAO() *ExternalMoldArchiveDAO {
	return &ExternalMoldArchiveDAO{
		db: GetDB(),
	}
}

// GetByMoldCode 根据模具编号查询档案
func (d *ExternalMoldArchiveDAO) GetByMoldCode(moldCode string) (*entity.ExternalMoldArchive, error) {
	var archive entity.ExternalMoldArchive
	err := d.db.Where("mold_code = ? AND status = 1", moldCode).First(&archive).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &archive, nil
}

// Create 创建档案
func (d *ExternalMoldArchiveDAO) Create(archive *entity.ExternalMoldArchive) error {
	return d.db.Create(archive).Error
}

// Update 更新档案
func (d *ExternalMoldArchiveDAO) Update(archive *entity.ExternalMoldArchive) error {
	return d.db.Save(archive).Error
}

// UpdateUsageInfo 更新使用信息（次数、最近使用时间）
func (d *ExternalMoldArchiveDAO) UpdateUsageInfo(id int64, orderID int64) error {
	now := time.Now()
	return d.db.Model(&entity.ExternalMoldArchive{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_order_id": orderID,
			"last_use_date": now,
			"use_count":     gorm.Expr("use_count + 1"),
		}).Error
}

// List 查询列表（支持分页和筛选）
func (d *ExternalMoldArchiveDAO) List(moldCode string, customerID *int64, page, pageSize int) ([]entity.ExternalMoldArchive, int64, error) {
	var archives []entity.ExternalMoldArchive
	var total int64

	query := d.db.Model(&entity.ExternalMoldArchive{})

	// 筛选条件
	if moldCode != "" {
		query = query.Where("mold_code LIKE ?", moldCode+"%")
	}
	if customerID != nil && *customerID > 0 {
		query = query.Where("customer_id = ?", *customerID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("create_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&archives).Error

	return archives, total, err
}

// Delete 删除档案（软删除，改为禁用状态）
func (d *ExternalMoldArchiveDAO) Delete(id int64) error {
	return d.db.Model(&entity.ExternalMoldArchive{}).
		Where("id = ?", id).
		Update("status", 0).Error
}

// GetByID 根据ID查询档案
func (d *ExternalMoldArchiveDAO) GetByID(id int64) (*entity.ExternalMoldArchive, error) {
	var archive entity.ExternalMoldArchive
	err := d.db.Where("id = ?", id).First(&archive).Error
	if err != nil {
		return nil, err
	}
	return &archive, nil
}

// GetPendingOrdersByMoldCode 查询该模具编号的未完成订单（排除指定订单ID）
func (d *ExternalMoldArchiveDAO) GetPendingOrdersByMoldCode(moldCode string, excludeOrderID *int64) ([]map[string]interface{}, error) {
	var orders []map[string]interface{}

	query := d.db.Table("production_order_item as poi").
		Select("po.id as order_id, po.order_code, po.status, poi.product_name, poi.plan_qty, po.plan_start_date, po.plan_end_date").
		Joins("LEFT JOIN production_order po ON poi.order_id = po.id").
		Where("poi.external_mold_code = ? AND po.status IN (1, 2, 3, 4, 5, 6, 7)", moldCode) // 1=待排产, 2=生产中, 3=已完成, 4=待质检, 5=质检合格, 6=质检不合格, 7=已入库

	// 排除指定订单ID（用于编辑时不显示当前编辑的订单）
	if excludeOrderID != nil && *excludeOrderID > 0 {
		query = query.Where("po.id != ?", *excludeOrderID)
	}

	err := query.Order("po.create_time ASC").
		Find(&orders).Error

	return orders, err
}
