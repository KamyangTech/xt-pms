package service

import (
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ReceivingAccountService struct {
	dao *dao.ReceivingAccountDAO
}

// NewReceivingAccountService 创建收款账户Service实例
func NewReceivingAccountService() *ReceivingAccountService {
	return &ReceivingAccountService{
		dao: dao.NewReceivingAccountDAO(),
	}
}

// GetList 获取所有启用状态的收款账户列表
func (s *ReceivingAccountService) GetList() ([]*entity.ReceivingAccount, error) {
	return s.dao.FindAll()
}

// GetListByPage 分页获取收款账户（带搜索条件）
func (s *ReceivingAccountService) GetListByPage(page, pageSize int, accountName, accountNo, bankName, accountType string, status int) ([]*entity.ReceivingAccount, int64, error) {
	offset := (page - 1) * pageSize
	return s.dao.FindPageWithFilter(offset, pageSize, accountName, accountNo, bankName, accountType, status)
}

// GetListAll 获取所有收款账户（含禁用）
func (s *ReceivingAccountService) GetListAll() ([]*entity.ReceivingAccount, error) {
	return s.dao.FindAllWithDisabled()
}

// GetByID 根据ID获取收款账户
func (s *ReceivingAccountService) GetByID(id int64) (*entity.ReceivingAccount, error) {
	return s.dao.FindByID(id)
}

// Create 创建收款账户
func (s *ReceivingAccountService) Create(account *entity.ReceivingAccount) error {
	return s.dao.Create(account)
}

// Update 更新收款账户
func (s *ReceivingAccountService) Update(account *entity.ReceivingAccount) error {
	_, err := s.dao.FindByID(account.ID)
	if err != nil {
		return err
	}
	return s.dao.Update(account)
}

// Delete 删除收款账户
func (s *ReceivingAccountService) Delete(id int64) error {
	return s.dao.Delete(id)
}

// BatchDelete 批量删除收款账户
func (s *ReceivingAccountService) BatchDelete(ids []int64) error {
	return s.dao.BatchDelete(ids)
}

// UpdateStatus 更新收款账户状态
func (s *ReceivingAccountService) UpdateStatus(id int64, status int8) error {
	return s.dao.UpdateStatus(id, status)
}
