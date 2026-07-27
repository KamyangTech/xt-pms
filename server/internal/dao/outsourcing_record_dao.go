package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type OutsourcingRecordDAO struct{}

// NewOutsourcingRecordDAO 创建外发记录DAO
func NewOutsourcingRecordDAO() *OutsourcingRecordDAO {
	return &OutsourcingRecordDAO{}
}

// Create 创建外发记录
func (d *OutsourcingRecordDAO) Create(record *entity.OutsourcingRecord) error {
	return GetDB().Create(record).Error
}

// FindByID 根据ID查询外发记录
func (d *OutsourcingRecordDAO) FindByID(id int64) (*entity.OutsourcingRecord, error) {
	var record entity.OutsourcingRecord
	err := GetDB().Preload("Order").Preload("Processor").Preload("Mold").First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByCode 根据外发单号查询
func (d *OutsourcingRecordDAO) FindByCode(code string) (*entity.OutsourcingRecord, error) {
	var record entity.OutsourcingRecord
	err := GetDB().Where("record_code = ?", code).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Update 更新外发记录
func (d *OutsourcingRecordDAO) Update(record *entity.OutsourcingRecord) error {
	return GetDB().Save(record).Error
}

// Delete 删除外发记录
func (d *OutsourcingRecordDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.OutsourcingRecord{}, id).Error
}

// FindByOrderID 根据订单ID查询外发记录
func (d *OutsourcingRecordDAO) FindByOrderID(orderID int64) (*entity.OutsourcingRecord, error) {
	var record entity.OutsourcingRecord
	err := GetDB().Where("order_id = ?", orderID).
		Preload("Processor").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByStatus 根据状态查询外发记录列表
func (d *OutsourcingRecordDAO) FindByStatus(status int8) ([]entity.OutsourcingRecord, error) {
	var records []entity.OutsourcingRecord
	err := GetDB().Where("status = ?", status).
		Preload("Order").Preload("Processor").
		Order("create_time DESC").
		Find(&records).Error
	return records, err
}

// FindByWorkOrderID 根据模具工单ID查询外发记录
func (d *OutsourcingRecordDAO) FindByWorkOrderID(workOrderID int64) (*entity.OutsourcingRecord, error) {
	var record entity.OutsourcingRecord
	err := GetDB().Where("work_order_id = ?", workOrderID).
		Preload("Processor").
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByShareToken 根据分享token查询外发记录
func (d *OutsourcingRecordDAO) FindByShareToken(token string) (*entity.OutsourcingRecord, error) {
	var record entity.OutsourcingRecord
	err := GetDB().
		Preload("Order").
		Preload("Processor").
		Preload("Mold").
		Preload("WorkOrder").
		Where("share_token = ?", token).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByTypeWithPage 根据外发类型分页查询外发记录
func (d *OutsourcingRecordDAO) FindByTypeWithPage(outsourceType string, status string, recordCode string, processorID *int64, page int, pageSize int) ([]entity.OutsourcingRecord, int64, error) {
	var records []entity.OutsourcingRecord
	var total int64

	query := GetDB().Model(&entity.OutsourcingRecord{}).
		Preload("Order").
		Preload("Order.Items").
		Preload("Processor").
		Preload("Mold").
		Preload("WorkOrder")

	if outsourceType != "" {
		query = query.Where("outsource_type = ?", outsourceType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if recordCode != "" {
		query = query.Where("record_code LIKE ?", recordCode+"%")
	}
	if processorID != nil {
		query = query.Where("processor_id = ?", *processorID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("create_time DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}
