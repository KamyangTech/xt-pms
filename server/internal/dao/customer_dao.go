package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type CustomerDAO struct{}

// NewCustomerDAO 创建客户 DAO 实例
func NewCustomerDAO() *CustomerDAO {
	return &CustomerDAO{}
}

// Create 创建客户
func (d *CustomerDAO) Create(customer *entity.Customer) error {
	return DB.Create(customer).Error
}

// FindByID 根据 ID 查询客户
func (d *CustomerDAO) FindByID(id int64) (*entity.Customer, error) {
	var customer entity.Customer
	err := DB.First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindByCode 根据编号查询客户
func (d *CustomerDAO) FindByCode(code string) (*entity.Customer, error) {
	var customer entity.Customer
	err := DB.Where("cus_code = ?", code).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindAll 查询所有客户（下拉选择框用，限制 1000 条）
func (d *CustomerDAO) FindAll() ([]*entity.Customer, error) {
	var customers []*entity.Customer
	err := DB.Select("id", "cus_code", "cus_name", "status").Order("id DESC").Limit(1000).Find(&customers).Error
	return customers, err
}

// FindPage 分页查询客户
func (d *CustomerDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var customers []*entity.Customer
	return utils.Paginate(DB, page, pageSize, &customers, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

// FindPageWithFilter 分页查询客户（带搜索条件）
func (d *CustomerDAO) FindPageWithFilter(offset, limit int, cusCode, cusName, contact, phone string, status int) ([]*entity.Customer, int64, error) {
	var customers []*entity.Customer
	var total int64

	query := DB.Model(&entity.Customer{})

	if cusCode != "" {
		query = query.Where("cus_code LIKE ?", cusCode+"%")
	}
	if cusName != "" {
		query = query.Where("cus_name LIKE ?", "%"+cusName+"%")
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
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&customers).Error

	return customers, total, err
}

// Update 更新客户
func (d *CustomerDAO) Update(customer *entity.Customer) error {
	return DB.Save(customer).Error
}

// Delete 删除客户
func (d *CustomerDAO) Delete(id int64) error {
	return DB.Delete(&entity.Customer{}, id).Error
}

// BatchDelete 批量删除客户
func (d *CustomerDAO) BatchDelete(ids []int64) error {
	return DB.Delete(&entity.Customer{}, ids).Error
}
