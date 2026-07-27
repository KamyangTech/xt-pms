package service

import (
	"errors"
	"fmt"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// ProductBOMService 产品 BOM 服务
type ProductBOMService struct {
	productBOMDAO *dao.ProductBOMDAO
}

func NewProductBOMService() *ProductBOMService {
	return &ProductBOMService{productBOMDAO: dao.NewProductBOMDAO()}
}

// CreateProductBOM 创建产品 BOM（支持批量创建）
func (s *ProductBOMService) CreateProductBOM(bom *entity.ProductBOM) (*entity.ProductBOM, error) {
	productDAO := dao.NewProductDAO()
	materialDAO := dao.NewMaterialDAO()

	if _, err := productDAO.FindByID(bom.ProductID); err != nil {
		return nil, fmt.Errorf("产品不存在: %w", err)
	}

	if _, err := materialDAO.FindByID(bom.MaterialID); err != nil {
		return nil, fmt.Errorf("物料不存在: %w", err)
	}

	err := s.productBOMDAO.Create(bom)
	if err != nil {
		return nil, err
	}
	return bom, nil
}

// GetProductBOMByProductID 获取产品的完整 BOM 清单
func (s *ProductBOMService) GetProductBOMByProductID(productID int64) ([]*entity.ProductBOM, error) {
	boms, err := s.productBOMDAO.FindByProductID(productID)
	if err != nil {
		return nil, err
	}

	totalCost := 0.0
	for _, bom := range boms {
		totalCost += bom.TotalPrice
	}

	return boms, nil
}

// UpdateProductBOM 更新产品 BOM
func (s *ProductBOMService) UpdateProductBOM(bom *entity.ProductBOM) error {
	existing, err := s.productBOMDAO.FindByID(bom.ID)
	if err != nil {
		return errors.New("BOM 记录不存在")
	}

	// 保持产品 ID 不变
	bom.ProductID = existing.ProductID

	return s.productBOMDAO.Update(bom)
}

// DeleteProductBOM 删除单个 BOM 记录
func (s *ProductBOMService) DeleteProductBOM(id int64) error {
	return s.productBOMDAO.Delete(id)
}

// DeleteProductBOMByProductID 删除整个产品的所有 BOM（用于重新设计 BOM）
func (s *ProductBOMService) DeleteProductBOMByProductID(productID int64) error {
	return s.productBOMDAO.DeleteByProductID(productID)
}
