package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProductionOrderItemDAO struct{}

// NewProductionOrderItemDAO 创建生产订单明细DAO
func NewProductionOrderItemDAO() *ProductionOrderItemDAO {
	return &ProductionOrderItemDAO{}
}

// Create 创建生产订单明细
func (d *ProductionOrderItemDAO) Create(item *entity.ProductionOrderItem) error {
	return GetDB().Create(item).Error
}

// FindByOrderID 根据订单ID查询明细列表
func (d *ProductionOrderItemDAO) FindByOrderID(orderID int64) ([]entity.ProductionOrderItem, error) {
	var items []entity.ProductionOrderItem
	err := GetDB().Preload("Product").Preload("Mold").Where("order_id = ?", orderID).Order("sort ASC").Find(&items).Error
	return items, err
}

// FindByID 根据明细项ID查询单个明细项
func (d *ProductionOrderItemDAO) FindByID(id int64) (*entity.ProductionOrderItem, error) {
	var item entity.ProductionOrderItem
	err := GetDB().Preload("Product").Preload("Mold").First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Update 更新生产订单明细
func (d *ProductionOrderItemDAO) Update(item *entity.ProductionOrderItem) error {
	return GetDB().Save(item).Error
}

// Delete 删除生产订单明细
func (d *ProductionOrderItemDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.ProductionOrderItem{}, id).Error
}

// DeleteByOrderID 删除订单下的所有明细
func (d *ProductionOrderItemDAO) DeleteByOrderID(orderID int64) error {
	return GetDB().Where("order_id = ?", orderID).Delete(&entity.ProductionOrderItem{}).Error
}

// FindByMoldID 根据模具ID查询所有使用该模具的订单明细（优化：移除Preload避免N+1查询）
func (d *ProductionOrderItemDAO) FindByMoldID(moldID int64) ([]entity.ProductionOrderItem, error) {
	var items []entity.ProductionOrderItem
	// 支持单模具和多模具查询
	// JSON_CONTAINS需要第二个参数为JSON格式，使用fmt.Sprintf将整数转为JSON数字格式
	err := GetDB().Where("mold_id = ? OR JSON_CONTAINS(mold_ids, CAST(? AS JSON))", moldID, moldID).Order("create_time DESC").Find(&items).Error
	if err != nil {
		return items, err
	}

	// 按需加载关联数据（批量查询优化）
	if len(items) > 0 {
		orderIDs := make([]int64, 0)
		productIDs := make([]int64, 0)
		moldIDs := make([]int64, 0)
		orderIDSet := make(map[int64]bool)
		productIDSet := make(map[int64]bool)
		moldIDSet := make(map[int64]bool)

		for _, item := range items {
			if item.OrderID > 0 && !orderIDSet[item.OrderID] {
				orderIDs = append(orderIDs, item.OrderID)
				orderIDSet[item.OrderID] = true
			}
			if item.ProductID != nil && *item.ProductID > 0 && !productIDSet[*item.ProductID] {
				productIDs = append(productIDs, *item.ProductID)
				productIDSet[*item.ProductID] = true
			}
			// 单模具
			if item.MoldID != nil && *item.MoldID > 0 && !moldIDSet[*item.MoldID] {
				moldIDs = append(moldIDs, *item.MoldID)
				moldIDSet[*item.MoldID] = true
			}
			// 多模具
			if len(item.MoldIDs) > 0 {
				for _, mid := range item.MoldIDs {
					if mid > 0 && !moldIDSet[mid] {
						moldIDs = append(moldIDs, mid)
						moldIDSet[mid] = true
					}
				}
			}
		}

		// 批量查询订单
		if len(orderIDs) > 0 {
			var orders []entity.ProductionOrder
			if err := GetDB().Where("id IN ?", orderIDs).Find(&orders).Error; err == nil {
				orderMap := make(map[int64]*entity.ProductionOrder)
				for i := range orders {
					orderMap[orders[i].ID] = &orders[i]
				}
				for i := range items {
					if order, ok := orderMap[items[i].OrderID]; ok {
						items[i].Order = order
					}
				}
			}
		}

		// 批量查询产品
		if len(productIDs) > 0 {
			var products []entity.Product
			if err := GetDB().Where("id IN ?", productIDs).Find(&products).Error; err == nil {
				productMap := make(map[int64]*entity.Product)
				for i := range products {
					productMap[products[i].ID] = &products[i]
				}
				for i := range items {
					if items[i].ProductID != nil {
						if product, ok := productMap[*items[i].ProductID]; ok {
							items[i].Product = product
						}
					}
				}
			}
		}

		// 批量查询模具
		if len(moldIDs) > 0 {
			var molds []entity.Mold
			if err := GetDB().Where("id IN ?", moldIDs).Find(&molds).Error; err == nil {
				moldMap := make(map[int64]*entity.Mold)
				for i := range molds {
					moldMap[molds[i].ID] = &molds[i]
				}
				for i := range items {
					if items[i].MoldID != nil {
						if mold, ok := moldMap[*items[i].MoldID]; ok {
							items[i].Mold = mold
						}
					}
					if len(items[i].MoldIDs) > 0 {
						items[i].Molds = make([]*entity.Mold, 0)
						for _, mid := range items[i].MoldIDs {
							if mold, ok := moldMap[mid]; ok {
								items[i].Molds = append(items[i].Molds, mold)
							}
						}
					}
				}
			}
		}
	}

	return items, nil
}
