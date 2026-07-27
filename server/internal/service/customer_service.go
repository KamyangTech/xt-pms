package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type CustomerService struct {
	customerDAO *dao.CustomerDAO
}

func NewCustomerService() *CustomerService {
	return &CustomerService{
		customerDAO: dao.NewCustomerDAO(),
	}
}

// CreateCustomer 创建客户
func (s *CustomerService) CreateCustomer(customer *entity.Customer) (*entity.Customer, error) {
	existing, _ := s.customerDAO.FindByCode(customer.CusCode)
	if existing != nil {
		return nil, errors.New("客户编号已存在")
	}

	err := s.customerDAO.Create(customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomerByID 根据 ID 获取客户
func (s *CustomerService) GetCustomerByID(id int64) (*entity.Customer, error) {
	return s.customerDAO.FindByID(id)
}

// GetAllCustomers 获取所有客户
func (s *CustomerService) GetAllCustomers() ([]*entity.Customer, error) {
	return s.customerDAO.FindAll()
}

// UpdateCustomer 更新客户
func (s *CustomerService) UpdateCustomer(customer *entity.Customer) error {
	if customer.ID == 0 {
		return errors.New("客户 ID 不能为空")
	}

	existing, err := s.customerDAO.FindByID(customer.ID)
	if err != nil {
		return errors.New("客户不存在")
	}

	// 如果编号被修改，检查新编号是否已被其他客户占用
	if customer.CusCode != existing.CusCode {
		other, _ := s.customerDAO.FindByCode(customer.CusCode)
		if other != nil && other.ID != customer.ID {
			return errors.New("客户编号已被其他客户使用")
		}
	}

	return s.customerDAO.Update(customer)
}

// GetCustomersByPageWithFilter 分页获取客户（带搜索条件）
func (s *CustomerService) GetCustomersByPageWithFilter(page, pageSize int, cusCode, cusName, contact, phone string, status int) ([]*entity.Customer, int64, error) {
	offset := (page - 1) * pageSize
	return s.customerDAO.FindPageWithFilter(offset, pageSize, cusCode, cusName, contact, phone, status)
}

// DeleteCustomer 删除客户
func (s *CustomerService) DeleteCustomer(id int64) error {
	return s.customerDAO.Delete(id)
}

// BatchDeleteCustomers 批量删除客户
func (s *CustomerService) BatchDeleteCustomers(ids []int64) error {
	return s.customerDAO.BatchDelete(ids)
}
