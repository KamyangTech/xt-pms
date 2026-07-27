package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type MaterialDAO struct{}

// NewMaterialDAO 创建物料 DAO 实例
func NewMaterialDAO() *MaterialDAO {
	return &MaterialDAO{}
}

// Create 创建物料
func (d *MaterialDAO) Create(material *entity.Material) error {
	return DB.Create(material).Error
}

// FindByID 根据 ID 查询物料
func (d *MaterialDAO) FindByID(id int64) (*entity.Material, error) {
	var material entity.Material
	err := DB.First(&material, id).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

// FindByCode 根据编号查询物料
func (d *MaterialDAO) FindByCode(code string) (*entity.Material, error) {
	var material entity.Material
	err := DB.Where("mat_code = ?", code).First(&material).Error
	if err != nil {
		return nil, err
	}
	return &material, nil
}

// FindAll 查询所有物料
func (d *MaterialDAO) FindAll() ([]*entity.Material, error) {
	var materials []*entity.Material
	err := DB.Order("id DESC").Find(&materials).Error
	return materials, err
}

// FindPage 分页查询物料
func (d *MaterialDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var materials []*entity.Material
	return utils.Paginate(DB, page, pageSize, &materials, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

// Update 更新物料
func (d *MaterialDAO) Update(material *entity.Material) error {
	return DB.Save(material).Error
}

// Delete 删除物料
func (d *MaterialDAO) Delete(id int64) error {
	return DB.Delete(&entity.Material{}, id).Error
}
