package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type ProcessorDAO struct{}

// NewProcessorDAO 创建加工商 DAO 实例
func NewProcessorDAO() *ProcessorDAO {
	return &ProcessorDAO{}
}

// Create 创建加工商
func (d *ProcessorDAO) Create(processor *entity.Processor) error {
	return DB.Create(processor).Error
}

// FindByID 根据 ID 查询加工商
func (d *ProcessorDAO) FindByID(id int64) (*entity.Processor, error) {
	var processor entity.Processor
	err := DB.First(&processor, id).Error
	if err != nil {
		return nil, err
	}
	return &processor, nil
}

// FindByCode 根据编号查询加工商
func (d *ProcessorDAO) FindByCode(code string) (*entity.Processor, error) {
	var processor entity.Processor
	err := DB.Where("processor_code = ?", code).First(&processor).Error
	if err != nil {
		return nil, err
	}
	return &processor, nil
}

// FindAll 查询所有加工商（下拉选择框用，限制 1000 条）
func (d *ProcessorDAO) FindAll() ([]*entity.Processor, error) {
	var processors []*entity.Processor
	err := DB.Select("id", "processor_code", "processor_name", "status").Order("id DESC").Limit(1000).Find(&processors).Error
	return processors, err
}

// FindPage 分页查询加工商
func (d *ProcessorDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var processors []*entity.Processor
	return utils.Paginate(DB, page, pageSize, &processors, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

// FindPageWithFilter 分页查询加工商（带搜索条件）
func (d *ProcessorDAO) FindPageWithFilter(offset, limit int, processorCode, processorName, contact, phone string, status int) ([]*entity.Processor, int64, error) {
	var processors []*entity.Processor
	var total int64

	query := DB.Model(&entity.Processor{})

	if processorCode != "" {
		query = query.Where("processor_code LIKE ?", processorCode+"%")
	}
	if processorName != "" {
		query = query.Where("processor_name LIKE ?", "%"+processorName+"%")
	}
	if contact != "" {
		query = query.Where("contact LIKE ?", "%"+contact+"%")
	}
	if phone != "" {
		query = query.Where("phone LIKE ?", "%"+phone+"%")
	}
	if status != 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&processors).Error

	return processors, total, err
}

// Update 更新加工商
func (d *ProcessorDAO) Update(processor *entity.Processor) error {
	return DB.Model(processor).Updates(map[string]interface{}{
		"processor_code": processor.ProcessorCode,
		"processor_name": processor.ProcessorName,
		"contact":        processor.Contact,
		"phone":          processor.Phone,
		"email":          processor.Email,
		"address":        processor.Address,
		"process_type":   processor.ProcessType,
		"remark":         processor.Remark,
		"status":         processor.Status,
	}).Error
}

// Delete 删除加工商
func (d *ProcessorDAO) Delete(id int64) error {
	return DB.Delete(&entity.Processor{}, id).Error
}

// BatchDelete 批量删除加工商
func (d *ProcessorDAO) BatchDelete(ids []int64) error {
	return DB.Delete(&entity.Processor{}, ids).Error
}
