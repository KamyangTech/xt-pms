package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type StockService struct {
	stockDAO *dao.StockDAO
}

func NewStockService() *StockService {
	return &StockService{stockDAO: dao.NewStockDAO()}
}

// CreateStock 创建库存记录
func (s *StockService) CreateStock(stock *entity.Stock) (*entity.Stock, error) {
	existing, err := s.stockDAO.FindByMaterialAndWarehouse(stock.MaterialID, stock.WhID)
	if err == nil && existing != nil {
		return nil, errors.New("该物料在此仓库已有库存记录，请直接更新")
	}

	err = s.stockDAO.Create(stock)
	if err != nil {
		return nil, err
	}
	return stock, nil
}

// GetStockByID 获取库存记录详情
func (s *StockService) GetStockByID(id int64) (*entity.Stock, error) {
	return s.stockDAO.FindByID(id)
}

// GetStockByMaterialAndWarehouse 查询指定物料在指定仓库的库存
func (s *StockService) GetStockByMaterialAndWarehouse(materialID, warehouseID int64) (*entity.Stock, error) {
	return s.stockDAO.FindByMaterialAndWarehouse(materialID, warehouseID)
}

// GetStockByMaterialProductAndWarehouse 查询指定物料、产品和仓库的库存（支持产品维度）
func (s *StockService) GetStockByMaterialProductAndWarehouse(materialID int64, productID *int64, warehouseID int64) (*entity.Stock, error) {
	return s.stockDAO.FindByMaterialProductAndWarehouse(materialID, productID, warehouseID)
}

// GetStocksByWarehouse 获取某仓库的所有库存
func (s *StockService) GetStocksByWarehouse(warehouseID int64) ([]*entity.Stock, error) {
	return s.stockDAO.FindByWarehouse(warehouseID)
}

// UpdateStock 更新库存数量
func (s *StockService) UpdateStock(stock *entity.Stock) error {
	existing, err := s.stockDAO.FindByID(stock.ID)
	if err != nil {
		return errors.New("库存记录不存在")
	}

	// 保持物料 ID 和仓库 ID 不变
	stock.MaterialID = existing.MaterialID
	stock.WhID = existing.WhID

	return s.stockDAO.Update(stock)
}

// InStock 入库操作（向后兼容，不支持产品维度）
func (s *StockService) InStock(materialID, warehouseID int64, qty float64) error {
	return s.InStockWithProduct(materialID, nil, warehouseID, qty)
}

// InStockWithProduct 入库操作（支持产品维度）
func (s *StockService) InStockWithProduct(materialID int64, productID *int64, warehouseID int64, qty float64) error {
	return s.inStockWithProductDB(dao.GetDB(), materialID, productID, warehouseID, qty)
}

// InStockWithProductTx 入库操作（使用指定事务）
func (s *StockService) InStockWithProductTx(tx *gorm.DB, materialID int64, productID *int64, warehouseID int64, qty float64) error {
	return s.inStockWithProductDB(tx, materialID, productID, warehouseID, qty)
}

// inStockWithProductDB 内部实现（复用逻辑）
func (s *StockService) inStockWithProductDB(db *gorm.DB, materialID int64, productID *int64, warehouseID int64, qty float64) error {
	stock, err := s.stockDAO.FindByMaterialProductAndWarehouseDB(db, materialID, productID, warehouseID)
	if err != nil {
		// 如果不存在，创建新记录
		newStock := &entity.Stock{
			MaterialID: materialID,
			ProductID:  productID,
			WhID:       warehouseID,
			StockQty:   qty,
			SafeStock:  0,
		}
		return s.stockDAO.CreateDB(db, newStock)
	}

	// 累加库存
	stock.StockQty += qty
	return s.stockDAO.UpdateDB(db, stock)
}

// OutStock 出库操作（向后兼容，不支持产品维度）
func (s *StockService) OutStock(materialID, warehouseID int64, qty float64) error {
	return s.OutStockWithProduct(materialID, nil, warehouseID, qty)
}

// OutStockWithProduct 出库操作（支持产品维度）
func (s *StockService) OutStockWithProduct(materialID int64, productID *int64, warehouseID int64, qty float64) error {
	stock, err := s.stockDAO.FindByMaterialProductAndWarehouse(materialID, productID, warehouseID)
	if err != nil {
		return errors.New("库存记录不存在")
	}

	if stock.StockQty < qty {
		return errors.New("库存不足")
	}

	stock.StockQty -= qty
	return s.stockDAO.Update(stock)
}

// CheckSafeStock 检查安全库存
func (s *StockService) CheckSafeStock() ([]*entity.Stock, error) {
	allStocks, err := s.stockDAO.FindByWarehouse(0) // 获取所有
	if err != nil {
		return nil, err
	}

	var lowStocks []*entity.Stock
	for _, stock := range allStocks {
		if stock.SafeStock > 0 && stock.StockQty < stock.SafeStock {
			lowStocks = append(lowStocks, stock)
		}
	}

	return lowStocks, nil
}

// DeleteStockByID 删除库存记录
func (s *StockService) DeleteStockByID(id int64) error {
	stock, err := s.stockDAO.FindByID(id)
	if err != nil {
		return errors.New("库存记录不存在")
	}

	if stock.StockQty > 0 {
		return errors.New("库存数量不为0，无法删除，请先出库")
	}

	return s.stockDAO.Delete(id)
}

// SetStockQuantity 盘点：直接设置库存为绝对值（用于库存盘点矫正）
func (s *StockService) SetStockQuantity(materialID int64, productID *int64, warehouseID int64, actualQty float64) (diff float64, err error) {
	stock, err := s.stockDAO.FindByMaterialProductAndWarehouse(materialID, productID, warehouseID)
	if err != nil {
		// 库存记录不存在，创建新记录
		newStock := &entity.Stock{
			MaterialID: materialID,
			ProductID:  productID,
			WhID:       warehouseID,
			StockQty:   actualQty,
			SafeStock:  0,
		}
		if createErr := s.stockDAO.Create(newStock); createErr != nil {
			return 0, createErr
		}
		return actualQty, nil
	}

	diff = actualQty - stock.StockQty
	stock.StockQty = actualQty
	if updateErr := s.stockDAO.Update(stock); updateErr != nil {
		return 0, updateErr
	}
	return diff, nil
}

// DeleteStockByMaterialAndProduct 根据物料和产品删除库存记录
func (s *StockService) DeleteStockByMaterialAndProduct(materialID int64, productID *int64, warehouseID int64) error {
	stock, err := s.stockDAO.FindByMaterialProductAndWarehouse(materialID, productID, warehouseID)
	if err != nil {
		return errors.New("库存记录不存在")
	}

	// 删除库存记录（允许库存不为0时删除，用于测试数据清理）
	return s.stockDAO.Delete(stock.ID)
}
