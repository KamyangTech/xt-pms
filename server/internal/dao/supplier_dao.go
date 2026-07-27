package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type SupplierDAO struct{}

// NewSupplierDAO 创建供应商 DAO 实例
func NewSupplierDAO() *SupplierDAO {
	return &SupplierDAO{}
}

// Create 创建供应商
func (d *SupplierDAO) Create(supplier *entity.Supplier) error {
	return DB.Create(supplier).Error
}

// FindByID 根据 ID 查询供应商
func (d *SupplierDAO) FindByID(id int64) (*entity.Supplier, error) {
	var supplier entity.Supplier
	err := DB.First(&supplier, id).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// FindByCode 根据编号查询供应商
func (d *SupplierDAO) FindByCode(code string) (*entity.Supplier, error) {
	var supplier entity.Supplier
	err := DB.Where("sup_code = ?", code).First(&supplier).Error
	if err != nil {
		return nil, err
	}
	return &supplier, nil
}

// FindAll 查询所有供应商（下拉选择框用，限制 1000 条）
func (d *SupplierDAO) FindAll() ([]*entity.Supplier, error) {
	var suppliers []*entity.Supplier
	err := DB.Select("id", "sup_code", "sup_name", "status").Order("id DESC").Limit(1000).Find(&suppliers).Error
	return suppliers, err
}

// FindPage 分页查询供应商
func (d *SupplierDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var suppliers []*entity.Supplier
	return utils.Paginate(DB, page, pageSize, &suppliers, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

// FindPageWithFilter 分页查询供应商（带搜索条件）
func (d *SupplierDAO) FindPageWithFilter(offset, limit int, supCode, supName, contact, phone string, status int) ([]*entity.Supplier, int64, error) {
	var suppliers []*entity.Supplier
	var total int64

	query := DB.Model(&entity.Supplier{})

	if supCode != "" {
		query = query.Where("sup_code LIKE ?", supCode+"%")
	}
	if supName != "" {
		query = query.Where("sup_name LIKE ?", "%"+supName+"%")
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
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&suppliers).Error

	return suppliers, total, err
}

// Update 更新供应商
func (d *SupplierDAO) Update(supplier *entity.Supplier) error {
	return DB.Model(supplier).Updates(map[string]interface{}{
		"sup_code": supplier.SupCode,
		"sup_name": supplier.SupName,
		"contact":  supplier.Contact,
		"phone":    supplier.Phone,
		"email":    supplier.Email,
		"address":  supplier.Address,
		"remark":   supplier.Remark,
		"status":   supplier.Status,
	}).Error
}

// Delete 删除供应商
func (d *SupplierDAO) Delete(id int64) error {
	return DB.Delete(&entity.Supplier{}, id).Error
}

// BatchDelete 批量删除供应商
func (d *SupplierDAO) BatchDelete(ids []int64) error {
	return DB.Delete(&entity.Supplier{}, ids).Error
}
