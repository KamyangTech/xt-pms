package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// MoldBOMDAO 模具 BOM 数据访问对象
type MoldBOMDAO struct{}

// NewMoldBOMDAO 创建模具 BOM DAO 实例
func NewMoldBOMDAO() *MoldBOMDAO {
	return &MoldBOMDAO{}
}

// Create 创建模具 BOM
func (d *MoldBOMDAO) Create(bom *entity.MoldBOM) error {
	return DB.Create(bom).Error
}

// FindByID 根据 ID 查询模具 BOM
func (d *MoldBOMDAO) FindByID(id int64) (*entity.MoldBOM, error) {
	var bom entity.MoldBOM
	err := DB.First(&bom, id).Error
	if err != nil {
		return nil, err
	}
	return &bom, nil
}

// FindByMoldID 根据模具 ID 查询 BOM 列表
func (d *MoldBOMDAO) FindByMoldID(moldID int64) ([]*entity.MoldBOM, error) {
	var boms []*entity.MoldBOM
	err := DB.Where("mold_id = ?", moldID).Find(&boms).Error
	return boms, err
}

// Update 更新模具 BOM
func (d *MoldBOMDAO) Update(bom *entity.MoldBOM) error {
	return DB.Save(bom).Error
}

// Delete 删除模具 BOM
func (d *MoldBOMDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldBOM{}, id).Error
}

// DeleteByMoldID 根据模具 ID 删除所有 BOM
func (d *MoldBOMDAO) DeleteByMoldID(moldID int64) error {
	return DB.Where("mold_id = ?", moldID).Delete(&entity.MoldBOM{}).Error
}
