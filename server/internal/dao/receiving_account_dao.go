package dao

import "github.com/KamyangTech/xt-pms/server/internal/model/entity"

type ReceivingAccountDAO struct{}

// NewReceivingAccountDAO 创建收款账户DAO实例
func NewReceivingAccountDAO() *ReceivingAccountDAO {
	return &ReceivingAccountDAO{}
}

// FindAll 获取所有启用状态的收款账户列表
func (d *ReceivingAccountDAO) FindAll() ([]*entity.ReceivingAccount, error) {
	var list []*entity.ReceivingAccount
	err := DB.Where("status = 1").Order("sort ASC, id DESC").Find(&list).Error
	return list, err
}

// FindAllWithDisabled 获取所有收款账户（含禁用）
func (d *ReceivingAccountDAO) FindAllWithDisabled() ([]*entity.ReceivingAccount, error) {
	var list []*entity.ReceivingAccount
	err := DB.Order("sort ASC, id DESC").Find(&list).Error
	return list, err
}

// FindPageWithFilter 分页查询收款账户（带搜索条件）
func (d *ReceivingAccountDAO) FindPageWithFilter(offset, limit int, accountName, accountNo, bankName, accountType string, status int) ([]*entity.ReceivingAccount, int64, error) {
	var list []*entity.ReceivingAccount
	var total int64

	query := DB.Model(&entity.ReceivingAccount{})

	if accountName != "" {
		query = query.Where("account_name LIKE ?", "%"+accountName+"%")
	}
	if accountNo != "" {
		query = query.Where("account_no LIKE ?", "%"+accountNo+"%")
	}
	if bankName != "" {
		query = query.Where("bank_name LIKE ?", "%"+bankName+"%")
	}
	if accountType != "" {
		query = query.Where("account_type = ?", accountType)
	}
	if status != 0 {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	err := query.Offset(offset).Limit(limit).Order("sort ASC, id DESC").Find(&list).Error

	return list, total, err
}

// FindByID 根据ID获取收款账户
func (d *ReceivingAccountDAO) FindByID(id int64) (*entity.ReceivingAccount, error) {
	var account entity.ReceivingAccount
	err := DB.First(&account, id).Error
	return &account, err
}

// Create 创建收款账户
func (d *ReceivingAccountDAO) Create(account *entity.ReceivingAccount) error {
	return DB.Create(account).Error
}

// Update 更新收款账户
func (d *ReceivingAccountDAO) Update(account *entity.ReceivingAccount) error {
	return DB.Save(account).Error
}

// Delete 删除收款账户
func (d *ReceivingAccountDAO) Delete(id int64) error {
	return DB.Delete(&entity.ReceivingAccount{}, id).Error
}

// BatchDelete 批量删除收款账户
func (d *ReceivingAccountDAO) BatchDelete(ids []int64) error {
	return DB.Delete(&entity.ReceivingAccount{}, ids).Error
}

// UpdateStatus 更新收款账户状态
func (d *ReceivingAccountDAO) UpdateStatus(id int64, status int8) error {
	return DB.Model(&entity.ReceivingAccount{}).Where("id = ?", id).Update("status", status).Error
}
