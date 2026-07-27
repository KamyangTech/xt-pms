package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type StockDAO struct{}

func NewStockDAO() *StockDAO {
	return &StockDAO{}
}

func (d *StockDAO) Create(stock *entity.Stock) error {
	return DB.Create(stock).Error
}

func (d *StockDAO) CreateDB(db *gorm.DB, stock *entity.Stock) error {
	return db.Create(stock).Error
}

func (d *StockDAO) FindByID(id int64) (*entity.Stock, error) {
	var stock entity.Stock
	err := DB.Preload("Product").First(&stock, id).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindByMaterialAndWarehouse 根据物料和仓库查询库存（向后兼容，不支持产品维度）
func (d *StockDAO) FindByMaterialAndWarehouse(materialID, warehouseID int64) (*entity.Stock, error) {
	var stock entity.Stock
	err := DB.Where("material_id = ? AND wh_id = ? AND product_id IS NULL", materialID, warehouseID).First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindByMaterialProductAndWarehouse 根据物料、产品和仓库查询库存（支持产品维度）
func (d *StockDAO) FindByMaterialProductAndWarehouse(materialID int64, productID *int64, warehouseID int64) (*entity.Stock, error) {
	var stock entity.Stock
	query := DB.Where("material_id = ? AND wh_id = ?", materialID, warehouseID)

	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	} else {
		query = query.Where("product_id IS NULL")
	}

	err := query.Preload("Product").First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindByMaterialProductAndWarehouseDB 根据物料、产品和仓库查询库存（使用指定 DB 连接，支持事务）
func (d *StockDAO) FindByMaterialProductAndWarehouseDB(db *gorm.DB, materialID int64, productID *int64, warehouseID int64) (*entity.Stock, error) {
	var stock entity.Stock
	query := db.Where("material_id = ? AND wh_id = ?", materialID, warehouseID)

	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	} else {
		query = query.Where("product_id IS NULL")
	}

	err := query.Preload("Product").First(&stock).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

// FindByWarehouse 查询仓库下所有库存（支持产品维度）
func (d *StockDAO) FindByWarehouse(warehouseID int64) ([]*entity.Stock, error) {
	var stocks []*entity.Stock
	err := DB.Where("wh_id = ?", warehouseID).Preload("Product").Find(&stocks).Error
	return stocks, err
}

// FindByMaterial 根据物料查询所有仓库的库存
func (d *StockDAO) FindByMaterial(materialID int64) ([]*entity.Stock, error) {
	var stocks []*entity.Stock
	err := DB.Where("material_id = ?", materialID).Preload("Product").Find(&stocks).Error
	return stocks, err
}

// FindAll 查询所有库存记录（限制 1000 条）
func (d *StockDAO) FindAll() ([]*entity.Stock, error) {
	var stocks []*entity.Stock
	err := DB.Preload("Product", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "pro_code", "pro_name")
	}).Limit(1000).Find(&stocks).Error
	return stocks, err
}

func (d *StockDAO) Update(stock *entity.Stock) error {
	return DB.Save(stock).Error
}

func (d *StockDAO) UpdateDB(db *gorm.DB, stock *entity.Stock) error {
	return db.Save(stock).Error
}

func (d *StockDAO) Delete(id int64) error {
	return DB.Delete(&entity.Stock{}, id).Error
}
