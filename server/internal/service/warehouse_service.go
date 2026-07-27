package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

type WarehouseService struct {
	warehouseDAO *dao.WarehouseDAO
}

func NewWarehouseService() *WarehouseService {
	return &WarehouseService{
		warehouseDAO: dao.NewWarehouseDAO(),
	}
}

// CreateWarehouse 创建仓库
func (s *WarehouseService) CreateWarehouse(warehouse *entity.Warehouse) (*entity.Warehouse, error) {
	existing, _ := s.warehouseDAO.FindByCode(warehouse.WhCode)
	if existing != nil {
		return nil, errors.New("仓库编号已存在")
	}

	err := s.warehouseDAO.Create(warehouse)
	if err != nil {
		return nil, err
	}

	return warehouse, nil
}

// GetWarehouseByID 根据 ID 获取仓库
func (s *WarehouseService) GetWarehouseByID(id int64) (*entity.Warehouse, error) {
	return s.warehouseDAO.FindByID(id)
}

// GetAllWarehouses 获取所有仓库
func (s *WarehouseService) GetAllWarehouses() ([]*entity.Warehouse, error) {
	return s.warehouseDAO.FindAll()
}

// GetWarehousesByPage 分页获取仓库
func (s *WarehouseService) GetWarehousesByPage(page, pageSize int) ([]*entity.Warehouse, int64, error) {
	result, err := s.warehouseDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	// Paginate 返回的 List 是指向切片的指针，需要解引用
	if warehouses, ok := result.List.(*[]*entity.Warehouse); ok && warehouses != nil {
		return *warehouses, result.Total, nil
	}
	return []*entity.Warehouse{}, result.Total, nil
}

// GetWarehousesByPageWithFilter 分页获取仓库（带搜索条件）
func (s *WarehouseService) GetWarehousesByPageWithFilter(page, pageSize int, whCode, whName, manager string, status int) ([]*entity.Warehouse, int64, error) {
	offset := (page - 1) * pageSize
	return s.warehouseDAO.FindPageWithFilter(offset, pageSize, whCode, whName, manager, status)
}

// UpdateWarehouse 更新仓库
func (s *WarehouseService) UpdateWarehouse(warehouse *entity.Warehouse) error {
	existing, err := s.warehouseDAO.FindByID(warehouse.ID)
	if err := utils.EnsureExists(existing, err, "仓库"); err != nil {
		return err
	}

	return s.warehouseDAO.Update(warehouse)
}

// DeleteWarehouse 删除仓库
func (s *WarehouseService) DeleteWarehouse(id int64) error {
	return s.warehouseDAO.Delete(id)
}

// BatchDeleteWarehouses 批量删除仓库
func (s *WarehouseService) BatchDeleteWarehouses(ids []int64) error {
	return s.warehouseDAO.BatchDelete(ids)
}
