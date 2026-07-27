package dao

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type MoldUsageRecordDAO struct{}

// NewMoldUsageRecordDAO 创建模具使用记录 DAO 实例
func NewMoldUsageRecordDAO() *MoldUsageRecordDAO {
	return &MoldUsageRecordDAO{}
}

// Create 创建模具使用记录
func (d *MoldUsageRecordDAO) Create(record *entity.MoldUsageRecord) error {
	return DB.Create(record).Error
}

// FindByID 根据 ID 查询使用记录
func (d *MoldUsageRecordDAO) FindByID(id int64) (*entity.MoldUsageRecord, error) {
	var record entity.MoldUsageRecord
	err := DB.Preload("Mold").Preload("ProductionOrder").First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByMoldAndOrder 根据模具ID和生产订单ID查询使用记录
func (d *MoldUsageRecordDAO) FindByMoldAndOrder(moldID, orderID int64) (*entity.MoldUsageRecord, error) {
	var record entity.MoldUsageRecord
	err := DB.Where("mold_id = ? AND production_order_id = ?", moldID, orderID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByMoldAndItem 根据模具ID和生产订单明细ID查询使用记录
func (d *MoldUsageRecordDAO) FindByMoldAndItem(moldID, itemID int64) (*entity.MoldUsageRecord, error) {
	var record entity.MoldUsageRecord
	err := DB.Where("mold_id = ? AND production_order_item_id = ?", moldID, itemID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByMoldAndItemWithTx 在事务内根据模具ID和生产订单明细ID查询使用记录
func (d *MoldUsageRecordDAO) FindByMoldAndItemWithTx(tx *gorm.DB, moldID, itemID int64) (*entity.MoldUsageRecord, error) {
	var record entity.MoldUsageRecord
	err := tx.Where("mold_id = ? AND production_order_item_id = ?", moldID, itemID).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetListByMoldID 根据模具ID获取使用记录列表(分页)
// 【重构】从 production_report（生产报工）和 production_order_item（订单明细）关联查询
// 如果 moldID 为 0，则查询所有模具的使用记录；moldCode 支持模具编号模糊搜索
func (d *MoldUsageRecordDAO) GetListByMoldID(moldID int64, page, pageSize int, moldCode string) ([]entity.MoldUsageRecord, int64, error) {
	// 关联查询：production_report -> production_order_item -> mold
	type UsageRecordResult struct {
		ID                int64      `json:"id"`
		OrderCode         string     `json:"order_code"`
		ProductName       string     `json:"product_name"`
		Shots             int        `json:"shots"`
		OkQty             int        `json:"ok_qty"`
		BadQty            int        `json:"bad_qty"`
		ReportTime        *time.Time `json:"report_time"`
		ProductionOrderID *int64     `json:"production_order_id"`
		MoldID            int64      `json:"mold_id"`
		MoldCode          string     `json:"mold_code"`
		MoldName          string     `json:"mold_name"`
	}

	var records []UsageRecordResult
	var total int64

	query := DB.Table("production_report pr").
		Select(`
			pr.id,
			po.order_code,
			poi.product_name,
			(pr.ok_qty + pr.bad_qty) as shots,
			pr.ok_qty,
			pr.bad_qty,
			pr.report_date as report_time,
			pr.order_id as production_order_id,
			poi.mold_id,
			m.mold_code,
			m.mold_name
		`).
		Joins("LEFT JOIN production_order po ON pr.order_id = po.id").
		Joins("LEFT JOIN production_order_item poi ON pr.item_id = poi.id").
		Joins("LEFT JOIN mold m ON poi.mold_id = m.id").
		Where("poi.mold_id IS NOT NULL") // 只查询有模具的明细项

	// 如果 moldID > 0，则过滤特定模具
	if moldID > 0 {
		query = query.Where("poi.mold_id = ?", moldID)
	}

	// 如果 moldCode 不为空，按模具编号模糊搜索
	if moldCode != "" {
		query = query.Where("m.mold_code LIKE ?", moldCode+"%")
	}

	countQuery := DB.Table("production_report pr").
		Joins("LEFT JOIN production_order_item poi ON pr.item_id = poi.id").
		Where("poi.mold_id IS NOT NULL")

	if moldID > 0 {
		countQuery = countQuery.Where("poi.mold_id = ?", moldID)
	}

	if moldCode != "" {
		countQuery = countQuery.Joins("LEFT JOIN mold m ON poi.mold_id = m.id").
			Where("m.mold_code LIKE ?", moldCode+"%")
	}

	countQuery.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("pr.report_date DESC").Offset(offset).Limit(pageSize).Scan(&records).Error

	// 转换为 MoldUsageRecord 格式（兼容前端）
	result := make([]entity.MoldUsageRecord, 0, len(records))
	for _, r := range records {
		var productionOrderID int64
		if r.ProductionOrderID != nil {
			productionOrderID = *r.ProductionOrderID
		}
		record := entity.MoldUsageRecord{
			ID:                r.ID,
			ProductionOrderID: productionOrderID,
			ProductName:       &r.ProductName,
			TotalShots:        r.Shots,
			OkQty:             r.OkQty,
			BadQty:            r.BadQty,
			UsageDate:         r.ReportTime,
			ReportTime:        r.ReportTime, // 填充报工时间（用于前端显示）
			OrderCode:         r.OrderCode,  // 填充订单编号
			MoldID:            r.MoldID,     // 填充模具ID
			MoldCode:          r.MoldCode,   // 填充模具编号
			MoldName:          r.MoldName,   // 填充模具名称
		}
		result = append(result, record)
	}

	return result, total, err
}

// GetListByOrderID 根据生产订单ID获取使用记录
func (d *MoldUsageRecordDAO) GetListByOrderID(orderID int64) ([]entity.MoldUsageRecord, error) {
	var records []entity.MoldUsageRecord
	err := DB.Where("production_order_id = ?", orderID).Preload("Mold").Find(&records).Error
	return records, err
}

// Update 更新使用记录
func (d *MoldUsageRecordDAO) Update(record *entity.MoldUsageRecord) error {
	return DB.Save(record).Error
}

// Delete 删除使用记录
func (d *MoldUsageRecordDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldUsageRecord{}, id).Error
}
