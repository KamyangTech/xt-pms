package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type WarehouseDAO struct{}

// NewWarehouseDAO 创建仓库 DAO 实例
func NewWarehouseDAO() *WarehouseDAO {
	return &WarehouseDAO{}
}

// Create 创建仓库
func (d *WarehouseDAO) Create(warehouse *entity.Warehouse) error {
	return DB.Create(warehouse).Error
}

// FindByID 根据 ID 查询仓库
func (d *WarehouseDAO) FindByID(id int64) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	err := DB.First(&warehouse, id).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// FindByCode 根据编号查询仓库
func (d *WarehouseDAO) FindByCode(code string) (*entity.Warehouse, error) {
	var warehouse entity.Warehouse
	err := DB.Where("wh_code = ?", code).First(&warehouse).Error
	if err != nil {
		return nil, err
	}
	return &warehouse, nil
}

// FindAll 查询所有仓库
func (d *WarehouseDAO) FindAll() ([]*entity.Warehouse, error) {
	var warehouses []*entity.Warehouse
	err := DB.Order("id DESC").Find(&warehouses).Error
	return warehouses, err
}

// FindPage 分页查询仓库
func (d *WarehouseDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var warehouses []*entity.Warehouse
	return utils.Paginate(DB, page, pageSize, &warehouses, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

// FindPageWithFilter 分页查询仓库（带搜索条件）
func (d *WarehouseDAO) FindPageWithFilter(offset, limit int, whCode, whName, manager string, status int) ([]*entity.Warehouse, int64, error) {
	var warehouses []*entity.Warehouse
	var total int64

	query := DB.Model(&entity.Warehouse{})

	if whCode != "" {
		query = query.Where("wh_code LIKE ?", whCode+"%")
	}
	if whName != "" {
		query = query.Where("wh_name LIKE ?", "%"+whName+"%")
	}
	if manager != "" {
		query = query.Where("manager LIKE ?", "%"+manager+"%")
	}
	if status != 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&warehouses).Error

	return warehouses, total, err
}

// BatchDelete 批量删除仓库
func (d *WarehouseDAO) BatchDelete(ids []int64) error {
	return DB.Delete(&entity.Warehouse{}, ids).Error
}

// Update 更新仓库
func (d *WarehouseDAO) Update(warehouse *entity.Warehouse) error {
	return DB.Save(warehouse).Error
}

// Delete 删除仓库
func (d *WarehouseDAO) Delete(id int64) error {
	return DB.Delete(&entity.Warehouse{}, id).Error
}
