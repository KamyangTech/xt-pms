package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type ProcessDAO struct{}

// NewProcessDAO 创建工序 DAO 实例
func NewProcessDAO() *ProcessDAO {
	return &ProcessDAO{}
}

// Create 创建工序
func (d *ProcessDAO) Create(process *entity.Process) error {
	return DB.Create(process).Error
}

// FindByID 根据 ID 查询工序
func (d *ProcessDAO) FindByID(id int64) (*entity.Process, error) {
	var process entity.Process
	err := DB.First(&process, id).Error
	if err != nil {
		return nil, err
	}
	return &process, nil
}

// FindByCode 根据编号查询工序
func (d *ProcessDAO) FindByCode(code string) (*entity.Process, error) {
	var process entity.Process
	err := DB.Where("process_code = ?", code).First(&process).Error
	if err != nil {
		return nil, err
	}
	return &process, nil
}

// FindAll 查询所有工序
func (d *ProcessDAO) FindAll() ([]*entity.Process, error) {
	var processes []*entity.Process
	err := DB.Order("id DESC").Find(&processes).Error
	return processes, err
}

// FindPage 分页查询工序
func (d *ProcessDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var processes []*entity.Process
	return utils.Paginate(DB, page, pageSize, &processes, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

// Update 更新工序
func (d *ProcessDAO) Update(process *entity.Process) error {
	return DB.Save(process).Error
}

// Delete 删除工序
func (d *ProcessDAO) Delete(id int64) error {
	return DB.Delete(&entity.Process{}, id).Error
}
