package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

type SupplierService struct {
	supplierDAO *dao.SupplierDAO
}

func NewSupplierService() *SupplierService {
	return &SupplierService{
		supplierDAO: dao.NewSupplierDAO(),
	}
}

// CreateSupplier 创建供应商
func (s *SupplierService) CreateSupplier(supplier *entity.Supplier) (*entity.Supplier, error) {
	existing, _ := s.supplierDAO.FindByCode(supplier.SupCode)
	if existing != nil {
		return nil, errors.New("供应商编号已存在")
	}

	err := s.supplierDAO.Create(supplier)
	if err != nil {
		return nil, err
	}

	return supplier, nil
}

// GetSupplierByID 根据 ID 获取供应商
func (s *SupplierService) GetSupplierByID(id int64) (*entity.Supplier, error) {
	return s.supplierDAO.FindByID(id)
}

// GetAllSuppliers 获取所有供应商
func (s *SupplierService) GetAllSuppliers() ([]*entity.Supplier, error) {
	return s.supplierDAO.FindAll()
}

// GetSuppliersByPageWithFilter 分页获取供应商（带搜索条件）
func (s *SupplierService) GetSuppliersByPageWithFilter(page, pageSize int, supCode, supName, contact, phone string, status int) ([]*entity.Supplier, int64, error) {
	offset := (page - 1) * pageSize
	return s.supplierDAO.FindPageWithFilter(offset, pageSize, supCode, supName, contact, phone, status)
}

// UpdateSupplier 更新供应商
func (s *SupplierService) UpdateSupplier(supplier *entity.Supplier) error {
	existing, err := s.supplierDAO.FindByID(supplier.ID)
	if err := utils.EnsureExists(existing, err, "供应商"); err != nil {
		return err
	}

	// 如果编号被修改，检查新编号是否已被其他供应商占用
	if supplier.SupCode != existing.SupCode {
		other, _ := s.supplierDAO.FindByCode(supplier.SupCode)
		if other != nil && other.ID != supplier.ID {
			return errors.New("供应商编号已被其他供应商使用")
		}
	}

	return s.supplierDAO.Update(supplier)
}

// DeleteSupplier 删除供应商
func (s *SupplierService) DeleteSupplier(id int64) error {
	return s.supplierDAO.Delete(id)
}

// BatchDeleteSuppliers 批量删除供应商
func (s *SupplierService) BatchDeleteSuppliers(ids []int64) error {
	return s.supplierDAO.BatchDelete(ids)
}
