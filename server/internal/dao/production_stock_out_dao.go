package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProductionStockOutDAO struct{}

// NewProductionStockOutDAO 创建生产出库DAO
func NewProductionStockOutDAO() *ProductionStockOutDAO {
	return &ProductionStockOutDAO{}
}

// Create 创建出库记录
func (d *ProductionStockOutDAO) Create(stockOut *entity.ProductionStockOut) error {
	return GetDB().Create(stockOut).Error
}

// FindByID 根据ID查询出库记录
func (d *ProductionStockOutDAO) FindByID(id int64) (*entity.ProductionStockOut, error) {
	var stockOut entity.ProductionStockOut
	// Preload Order 并加载 Order 的 Items 关联
	err := GetDB().Preload("Order.Customer").Preload("Order.Items").Preload("Warehouse").Preload("Mold").First(&stockOut, id).Error
	if err != nil {
		return nil, err
	}

	// 如果有 Order 和 Items,从 Order.Items 中通过 product_id 匹配获取 product_image_url
	if stockOut.Order != nil && len(stockOut.Order.Items) > 0 && stockOut.ProductID != nil {
		for _, item := range stockOut.Order.Items {
			if item.ProductID != nil && *item.ProductID == *stockOut.ProductID {
				if item.ProductImageURL != "" {
					stockOut.ProductImageURL = &item.ProductImageURL
				}
				break
			}
		}
	}

	// 如果通过 product_id 没找到，尝试通过模具编号和产品名称匹配
	if stockOut.ProductImageURL == nil && stockOut.Order != nil && len(stockOut.Order.Items) > 0 {
		for _, item := range stockOut.Order.Items {
			// 匹配模具编号
			moldMatch := false
			if stockOut.MoldID != nil && item.MoldID != nil && *stockOut.MoldID == *item.MoldID {
				moldMatch = true
			}
			if stockOut.ExternalMoldCode != nil && item.ExternalMoldCode != nil && *stockOut.ExternalMoldCode == *item.ExternalMoldCode {
				moldMatch = true
			}

			// 匹配产品名称
			nameMatch := false
			if stockOut.ProductName != "" && item.ProductName != "" && stockOut.ProductName == item.ProductName {
				nameMatch = true
			}

			// 如果模具和名称都匹配，获取产品图片
			if moldMatch && nameMatch && item.ProductImageURL != "" {
				stockOut.ProductImageURL = &item.ProductImageURL
				break
			}
		}
	}

	// 兜底：订单明细没有匹配到图片时，从 production_order_item 表查询
	if stockOut.ProductImageURL == nil && stockOut.ProductID != nil && *stockOut.ProductID > 0 {
		var imageUrl string
		GetDB().Table("production_order_item").
			Select("product_image_url").
			Where("product_id = ? AND product_image_url IS NOT NULL AND product_image_url != ''", *stockOut.ProductID).
			Order("id DESC").
			Limit(1).
			Pluck("product_image_url", &imageUrl)
		if imageUrl != "" {
			stockOut.ProductImageURL = &imageUrl
		}
	}

	if stockOut.ProductImageURL == nil && stockOut.MoldID != nil && *stockOut.MoldID > 0 {
		var imageUrl string
		GetDB().Table("production_order_item").
			Select("product_image_url").
			Where("mold_id = ? AND product_image_url IS NOT NULL AND product_image_url != ''", *stockOut.MoldID).
			Order("id DESC").
			Limit(1).
			Pluck("product_image_url", &imageUrl)
		if imageUrl != "" {
			stockOut.ProductImageURL = &imageUrl
		}
	}

	if stockOut.ProductImageURL == nil && stockOut.ExternalMoldCode != nil && *stockOut.ExternalMoldCode != "" {
		var imageUrl string
		GetDB().Table("production_order_item").
			Select("product_image_url").
			Where("external_mold_code = ? AND product_image_url IS NOT NULL AND product_image_url != ''", *stockOut.ExternalMoldCode).
			Order("id DESC").
			Limit(1).
			Pluck("product_image_url", &imageUrl)
		if imageUrl != "" {
			stockOut.ProductImageURL = &imageUrl
		}
	}

	return &stockOut, nil
}

// FindByIDs 批量查询出库记录（优化 N+1 问题）
func (d *ProductionStockOutDAO) FindByIDs(ids []int64) ([]entity.ProductionStockOut, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	var stockOuts []entity.ProductionStockOut
	err := GetDB().Where("id IN ?", ids).
		Preload("Order.Customer").
		Preload("Order.Items").
		Preload("Warehouse").
		Preload("Mold").
		Find(&stockOuts).Error

	if err != nil {
		return nil, err
	}

	// 为每个出库记录填充 product_image_url
	for i := range stockOuts {
		// 优先通过 product_id 匹配
		if stockOuts[i].Order != nil && len(stockOuts[i].Order.Items) > 0 && stockOuts[i].ProductID != nil {
			for _, item := range stockOuts[i].Order.Items {
				if item.ProductID != nil && *item.ProductID == *stockOuts[i].ProductID {
					if item.ProductImageURL != "" {
						stockOuts[i].ProductImageURL = &item.ProductImageURL
					}
					break
				}
			}
		}

		// 如果通过 product_id 没找到，尝试通过模具编号和产品名称匹配
		if stockOuts[i].ProductImageURL == nil && stockOuts[i].Order != nil && len(stockOuts[i].Order.Items) > 0 {
			for _, item := range stockOuts[i].Order.Items {
				// 匹配模具编号
				moldMatch := false
				if stockOuts[i].MoldID != nil && item.MoldID != nil && *stockOuts[i].MoldID == *item.MoldID {
					moldMatch = true
				}
				if stockOuts[i].ExternalMoldCode != nil && item.ExternalMoldCode != nil && *stockOuts[i].ExternalMoldCode == *item.ExternalMoldCode {
					moldMatch = true
				}

				// 匹配产品名称
				nameMatch := false
				if stockOuts[i].ProductName != "" && item.ProductName != "" && stockOuts[i].ProductName == item.ProductName {
					nameMatch = true
				}

			// 如果模具和名称都匹配，获取产品图片
			if moldMatch && nameMatch && item.ProductImageURL != "" {
				stockOuts[i].ProductImageURL = &item.ProductImageURL
				break
			}
		}
	}
}

	// 兜底：从 production_order_item 表批量查询缺失的图片（不依赖 order_id）
	directOutProductIDs := make([]int64, 0)
	directOutMoldIDs := make([]int64, 0)
	directOutExtMoldCodes := make([]string, 0)
	for i := range stockOuts {
		if stockOuts[i].ProductImageURL == nil && stockOuts[i].ProductID != nil && *stockOuts[i].ProductID > 0 {
			directOutProductIDs = append(directOutProductIDs, *stockOuts[i].ProductID)
		}
		if stockOuts[i].ProductImageURL == nil && stockOuts[i].MoldID != nil && *stockOuts[i].MoldID > 0 {
			directOutMoldIDs = append(directOutMoldIDs, *stockOuts[i].MoldID)
		}
		if stockOuts[i].ProductImageURL == nil && stockOuts[i].ExternalMoldCode != nil && *stockOuts[i].ExternalMoldCode != "" {
			directOutExtMoldCodes = append(directOutExtMoldCodes, *stockOuts[i].ExternalMoldCode)
		}
	}

	// 批量查询这些product_id对应的产品图片
	if len(directOutProductIDs) > 0 {
		var items []entity.ProductionOrderItem
		GetDB().Where("product_id IN ? AND product_image_url IS NOT NULL AND product_image_url != ''", directOutProductIDs).
			Order("id DESC").
			Find(&items)

		productImageMap := make(map[int64]string)
		for _, item := range items {
			if item.ProductID != nil && item.ProductImageURL != "" {
				if _, exists := productImageMap[*item.ProductID]; !exists {
					productImageMap[*item.ProductID] = item.ProductImageURL
				}
			}
		}

		for i := range stockOuts {
			if stockOuts[i].ProductImageURL == nil && stockOuts[i].ProductID != nil {
				if imageURL, ok := productImageMap[*stockOuts[i].ProductID]; ok {
					stockOuts[i].ProductImageURL = &imageURL
				}
			}
		}
	}

	// 批量查询 mold_id 对应的产品图片
	if len(directOutMoldIDs) > 0 {
		var moldItems []entity.ProductionOrderItem
		GetDB().Where("mold_id IN ? AND product_image_url IS NOT NULL AND product_image_url != ''", directOutMoldIDs).
			Order("id DESC").
			Find(&moldItems)

		moldImageMap := make(map[int64]string)
		for _, item := range moldItems {
			if item.MoldID != nil && item.ProductImageURL != "" {
				if _, exists := moldImageMap[*item.MoldID]; !exists {
					moldImageMap[*item.MoldID] = item.ProductImageURL
				}
			}
		}

		for i := range stockOuts {
			if stockOuts[i].ProductImageURL == nil && stockOuts[i].MoldID != nil {
				if imageURL, ok := moldImageMap[*stockOuts[i].MoldID]; ok {
					stockOuts[i].ProductImageURL = &imageURL
				}
			}
		}
	}

	// 批量查询 external_mold_code 对应的产品图片
	if len(directOutExtMoldCodes) > 0 {
		var extMoldItems []entity.ProductionOrderItem
		GetDB().Where("external_mold_code IN ? AND product_image_url IS NOT NULL AND product_image_url != ''", directOutExtMoldCodes).
			Order("id DESC").
			Find(&extMoldItems)

		extMoldImageMap := make(map[string]string)
		for _, item := range extMoldItems {
			if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" && item.ProductImageURL != "" {
				if _, exists := extMoldImageMap[*item.ExternalMoldCode]; !exists {
					extMoldImageMap[*item.ExternalMoldCode] = item.ProductImageURL
				}
			}
		}

		for i := range stockOuts {
			if stockOuts[i].ProductImageURL == nil && stockOuts[i].ExternalMoldCode != nil {
				if imageURL, ok := extMoldImageMap[*stockOuts[i].ExternalMoldCode]; ok {
					stockOuts[i].ProductImageURL = &imageURL
				}
			}
		}
	}

	return stockOuts, nil
}

// FindByCode 根据出库单号查询
func (d *ProductionStockOutDAO) FindByCode(code string) (*entity.ProductionStockOut, error) {
	var stockOut entity.ProductionStockOut
	err := GetDB().Where("stock_code = ?", code).First(&stockOut).Error
	if err != nil {
		return nil, err
	}
	return &stockOut, nil
}

// Update 更新出库记录
func (d *ProductionStockOutDAO) Update(stockOut *entity.ProductionStockOut) error {
	return GetDB().Save(stockOut).Error
}

// Delete 删除出库记录
func (d *ProductionStockOutDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.ProductionStockOut{}, id).Error
}

// FindByOrderID 根据订单ID查询出库记录列表
func (d *ProductionStockOutDAO) FindByOrderID(orderID int64) ([]entity.ProductionStockOut, error) {
	var stockOuts []entity.ProductionStockOut
	err := GetDB().Where("order_id = ?", orderID).
		Preload("Warehouse").
		Order("stock_out_date DESC").
		Find(&stockOuts).Error
	return stockOuts, err
}

// GetOrderTotalStockOutQty 获取订单的累计出库数量
func (d *ProductionStockOutDAO) GetOrderTotalStockOutQty(orderID int64) (int, error) {
	var totalQty int
	err := GetDB().Model(&entity.ProductionStockOut{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("order_id = ?", orderID).
		Scan(&totalQty).Error
	return totalQty, err
}

// GetOrdersTotalStockOutQty 批量查询多个订单的累计出库数量，返回 map[orderID]totalQty
func (d *ProductionStockOutDAO) GetOrdersTotalStockOutQty(orderIDs []int64) (map[int64]int, error) {
	if len(orderIDs) == 0 {
		return make(map[int64]int), nil
	}

	type Result struct {
		OrderID  int64
		TotalQty int
	}
	var results []Result
	err := GetDB().Model(&entity.ProductionStockOut{}).
		Select("order_id, COALESCE(SUM(quantity), 0) as total_qty").
		Where("order_id IN ? AND out_type = ?", orderIDs, "normal").
		Group("order_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	qtyMap := make(map[int64]int, len(results))
	for _, r := range results {
		qtyMap[r.OrderID] = r.TotalQty
	}
	return qtyMap, nil
}

// FindByPage 分页查询出库记录（仅查询正常发货类型）
func (d *ProductionStockOutDAO) FindByPage(page, pageSize int, orderCode string, productName string, stockCode string) ([]entity.ProductionStockOut, int64, error) {
	var stockOuts []entity.ProductionStockOut
	var total int64

	query := GetDB().Model(&entity.ProductionStockOut{})

	// 只查询正常发货类型（out_type = 'normal'）
	query = query.Where("out_type = ?", "normal")

	if orderCode != "" {
		query = query.Joins("JOIN production_order ON production_order.id = production_stock_out.order_id").
			Where("production_order.order_code LIKE ?", orderCode+"%")
	}

	if productName != "" {
		query = query.Where("product_name LIKE ?", "%"+productName+"%")
	}

	if stockCode != "" {
		query = query.Where("stock_code LIKE ?", stockCode+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	// 优化：先查询主表数据，再按需加载关联数据
	err = query.Order("stock_out_date DESC").Offset(offset).Limit(pageSize).Find(&stockOuts).Error
	if err != nil {
		return nil, 0, err
	}

	// 批量加载关联数据（避免N+1查询）
	if len(stockOuts) > 0 {
		orderIDs := make([]int64, 0)
		warehouseIDs := make([]int64, 0)
		moldIDs := make([]int64, 0)
		orderIDSet := make(map[int64]bool)
		warehouseIDSet := make(map[int64]bool)
		moldIDSet := make(map[int64]bool)

		for _, so := range stockOuts {
			if so.OrderID != nil && *so.OrderID > 0 && !orderIDSet[*so.OrderID] {
				orderIDs = append(orderIDs, *so.OrderID)
				orderIDSet[*so.OrderID] = true
			}
			if so.WarehouseID > 0 && !warehouseIDSet[so.WarehouseID] {
				warehouseIDs = append(warehouseIDs, so.WarehouseID)
				warehouseIDSet[so.WarehouseID] = true
			}
			if so.MoldID != nil && *so.MoldID > 0 && !moldIDSet[*so.MoldID] {
				moldIDs = append(moldIDs, *so.MoldID)
				moldIDSet[*so.MoldID] = true
			}
		}

		// 批量查询订单
		if len(orderIDs) > 0 {
			var orders []entity.ProductionOrder
			// 关键：必须 Preload Items 才能获取 product_image_url
			GetDB().Preload("Items").Where("id IN ?", orderIDs).Find(&orders)
			orderMap := make(map[int64]*entity.ProductionOrder)
			for i := range orders {
				orderMap[orders[i].ID] = &orders[i]
			}
			for i := range stockOuts {
				if stockOuts[i].OrderID != nil {
					if order, ok := orderMap[*stockOuts[i].OrderID]; ok {
						stockOuts[i].Order = order

						// 从 Order.Items 中通过 product_id 匹配获取 product_image_url
						if len(order.Items) > 0 && stockOuts[i].ProductID != nil {
							for _, item := range order.Items {
								if item.ProductID != nil && *item.ProductID == *stockOuts[i].ProductID {
									if item.ProductImageURL != "" {
										stockOuts[i].ProductImageURL = &item.ProductImageURL
									}
									break
								}
							}
						}

						// 如果通过 product_id 没找到，尝试通过模具编号和产品名称匹配
						if stockOuts[i].ProductImageURL == nil && len(order.Items) > 0 {
							for _, item := range order.Items {
								// 匹配模具编号
								moldMatch := false
								if stockOuts[i].MoldID != nil && item.MoldID != nil && *stockOuts[i].MoldID == *item.MoldID {
									moldMatch = true
								}
								if stockOuts[i].ExternalMoldCode != nil && item.ExternalMoldCode != nil && *stockOuts[i].ExternalMoldCode == *item.ExternalMoldCode {
									moldMatch = true
								}

								// 匹配产品名称
								nameMatch := false
								if stockOuts[i].ProductName != "" && item.ProductName != "" && stockOuts[i].ProductName == item.ProductName {
									nameMatch = true
								}

								// 如果模具和名称都匹配，获取产品图片
								if moldMatch && nameMatch && item.ProductImageURL != "" {
									stockOuts[i].ProductImageURL = &item.ProductImageURL
									break
								}
							}
						}
					}
				}
			}
		}

		// 批量查询仓库
		if len(warehouseIDs) > 0 {
			var warehouses []entity.Warehouse
			GetDB().Where("id IN ?", warehouseIDs).Find(&warehouses)
			warehouseMap := make(map[int64]*entity.Warehouse)
			for i := range warehouses {
				warehouseMap[warehouses[i].ID] = &warehouses[i]
			}
			for i := range stockOuts {
				if warehouse, ok := warehouseMap[stockOuts[i].WarehouseID]; ok {
					stockOuts[i].Warehouse = warehouse
				}
			}
		}

		// 批量查询模具
		if len(moldIDs) > 0 {
			var molds []entity.Mold
			GetDB().Where("id IN ?", moldIDs).Find(&molds)
			moldMap := make(map[int64]*entity.Mold)
			for i := range molds {
				moldMap[molds[i].ID] = &molds[i]
			}
			for i := range stockOuts {
				if stockOuts[i].MoldID != nil {
					if mold, ok := moldMap[*stockOuts[i].MoldID]; ok {
						stockOuts[i].Mold = mold
						// 【修复】如果 product_name 为空，使用模具名称填充
						if stockOuts[i].ProductName == "" && mold.MoldName != "" {
							stockOuts[i].ProductName = mold.MoldName
						}
					}
				}
			}
		}

		directOutProductIDs := make([]int64, 0)
		directOutProductIDSet := make(map[int64]bool)

		for i := range stockOuts {
			if stockOuts[i].ProductImageURL == nil && stockOuts[i].ProductID != nil && *stockOuts[i].ProductID > 0 {
				if !directOutProductIDSet[*stockOuts[i].ProductID] {
					directOutProductIDs = append(directOutProductIDs, *stockOuts[i].ProductID)
					directOutProductIDSet[*stockOuts[i].ProductID] = true
				}
			}
		}

		// 批量查询这些product_id对应的产品图片（不限制 order_id，为所有缺失图片的记录兜底）
		if len(directOutProductIDs) > 0 {
			var items []entity.ProductionOrderItem
			GetDB().Where("product_id IN ? AND product_image_url IS NOT NULL AND product_image_url != ''", directOutProductIDs).
				Order("id DESC").
				Find(&items)

			productImageMap := make(map[int64]string)
			for _, item := range items {
				if item.ProductID != nil && item.ProductImageURL != "" {
					if _, exists := productImageMap[*item.ProductID]; !exists {
						productImageMap[*item.ProductID] = item.ProductImageURL
					}
				}
			}

			// 填充所有缺失图片的记录
			for i := range stockOuts {
				if stockOuts[i].ProductImageURL == nil && stockOuts[i].ProductID != nil {
					if imageURL, ok := productImageMap[*stockOuts[i].ProductID]; ok {
						stockOuts[i].ProductImageURL = &imageURL
					}
				}
			}
		}

		directOutMoldIDs := make([]int64, 0)
		directOutMoldIDSet := make(map[int64]bool)
		directOutExtMoldCodes := make([]string, 0)
		directOutExtMoldCodeSet := make(map[string]bool)

		for i := range stockOuts {
			if stockOuts[i].ProductImageURL == nil && stockOuts[i].ProductID == nil {
				if stockOuts[i].MoldID != nil && *stockOuts[i].MoldID > 0 && !directOutMoldIDSet[*stockOuts[i].MoldID] {
					directOutMoldIDs = append(directOutMoldIDs, *stockOuts[i].MoldID)
					directOutMoldIDSet[*stockOuts[i].MoldID] = true
				}
				if stockOuts[i].ExternalMoldCode != nil && *stockOuts[i].ExternalMoldCode != "" && !directOutExtMoldCodeSet[*stockOuts[i].ExternalMoldCode] {
					directOutExtMoldCodes = append(directOutExtMoldCodes, *stockOuts[i].ExternalMoldCode)
					directOutExtMoldCodeSet[*stockOuts[i].ExternalMoldCode] = true
				}
			}
		}

		// 批量查询 mold_id 对应的产品图片
		if len(directOutMoldIDs) > 0 {
			var moldItems []entity.ProductionOrderItem
			GetDB().Where("mold_id IN ? AND product_image_url IS NOT NULL AND product_image_url != ''", directOutMoldIDs).
				Order("id DESC").
				Find(&moldItems)

			moldImageMap := make(map[int64]string)
			for _, item := range moldItems {
				if item.MoldID != nil && item.ProductImageURL != "" {
					if _, exists := moldImageMap[*item.MoldID]; !exists {
						moldImageMap[*item.MoldID] = item.ProductImageURL
					}
				}
			}

			for i := range stockOuts {
				if stockOuts[i].ProductImageURL == nil && stockOuts[i].ProductID == nil {
					if stockOuts[i].MoldID != nil {
						if imageURL, ok := moldImageMap[*stockOuts[i].MoldID]; ok {
							stockOuts[i].ProductImageURL = &imageURL
						}
					}
				}
			}
		}

		// 批量查询 external_mold_code 对应的产品图片
		if len(directOutExtMoldCodes) > 0 {
			var extMoldItems []entity.ProductionOrderItem
			GetDB().Where("external_mold_code IN ? AND product_image_url IS NOT NULL AND product_image_url != ''", directOutExtMoldCodes).
				Order("id DESC").
				Find(&extMoldItems)

			extMoldImageMap := make(map[string]string)
			for _, item := range extMoldItems {
				if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" && item.ProductImageURL != "" {
					if _, exists := extMoldImageMap[*item.ExternalMoldCode]; !exists {
						extMoldImageMap[*item.ExternalMoldCode] = item.ProductImageURL
					}
				}
			}

			for i := range stockOuts {
				if stockOuts[i].ProductImageURL == nil && stockOuts[i].ProductID == nil {
					if stockOuts[i].ExternalMoldCode != nil {
						if imageURL, ok := extMoldImageMap[*stockOuts[i].ExternalMoldCode]; ok {
							stockOuts[i].ProductImageURL = &imageURL
						}
					}
				}
			}
		}
	}

	return stockOuts, total, nil
}

// FindStockRecordsByMoldID 查询指定模具的所有出库记录
func (d *ProductionStockOutDAO) FindStockRecordsByMoldID(moldID int64, page, pageSize int, sortOrder string) ([]entity.ProductionStockOut, int64, error) {
	return d.FindStockRecordsByMoldIDAndProduct(moldID, nil, page, pageSize, sortOrder)
}

// FindStockRecordsByMoldIDAndProduct 查询指定模具和产品的所有出库记录
func (d *ProductionStockOutDAO) FindStockRecordsByMoldIDAndProduct(moldID int64, productID *int64, page, pageSize int, sortOrder string) ([]entity.ProductionStockOut, int64, error) {
	var records []entity.ProductionStockOut
	var total int64

	query := GetDB().Model(&entity.ProductionStockOut{}).
		Preload("Order").
		Preload("Warehouse").
		Preload("Mold").
		Where("mold_id = ?", moldID)

	// 如果指定了productID，增加product_id过滤条件
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}
	// 不指定productID时，查询所有产品（不添加product_id过滤条件）

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	// 根据排序参数动态排序
	orderClause := "stock_out_date DESC"
	if sortOrder == "asc" {
		orderClause = "stock_out_date ASC"
	}

	err = query.Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// FindStockRecordsByExternalMoldCode 查询指定外来模具的所有出库记录
func (d *ProductionStockOutDAO) FindStockRecordsByExternalMoldCode(externalMoldCode string, page, pageSize int, sortOrder string) ([]entity.ProductionStockOut, int64, error) {
	return d.FindStockRecordsByExternalMoldCodeAndProduct(externalMoldCode, nil, page, pageSize, sortOrder)
}

// FindStockRecordsByExternalMoldCodeAndProduct 查询指定外来模具和产品的所有出库记录
func (d *ProductionStockOutDAO) FindStockRecordsByExternalMoldCodeAndProduct(externalMoldCode string, productID *int64, page, pageSize int, sortOrder string) ([]entity.ProductionStockOut, int64, error) {
	var records []entity.ProductionStockOut
	var total int64

	query := GetDB().Model(&entity.ProductionStockOut{}).
		Preload("Order").
		Preload("Warehouse").
		Where("external_mold_code = ?", externalMoldCode)

	// 如果指定了productID，增加product_id过滤条件
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}
	// 不指定productID时，查询所有产品（不添加product_id过滤条件）

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	// 根据排序参数动态排序
	orderClause := "stock_out_date DESC"
	if sortOrder == "asc" {
		orderClause = "stock_out_date ASC"
	}

	err = query.Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetTotalStockOutByMoldID 获取指定模具的总出库数量
func (d *ProductionStockOutDAO) GetTotalStockOutByMoldID(moldID int64) (int, error) {
	return d.GetTotalStockOutByMoldIDAndProduct(moldID, nil)
}

// GetTotalStockOutByMoldIDAndProduct 获取指定模具和产品的总出库数量
func (d *ProductionStockOutDAO) GetTotalStockOutByMoldIDAndProduct(moldID int64, productID *int64) (int, error) {
	var totalQty int
	query := GetDB().Model(&entity.ProductionStockOut{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("mold_id = ?", moldID)

	// 如果指定了productID，增加product_id过滤条件
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	} else {
		query = query.Where("product_id IS NULL")
	}

	err := query.Scan(&totalQty).Error
	return totalQty, err
}

// GetTotalStockOutByExternalMoldCode 获取指定外来模具的总出库数量
func (d *ProductionStockOutDAO) GetTotalStockOutByExternalMoldCode(externalMoldCode string) (int, error) {
	return d.GetTotalStockOutByExternalMoldCodeAndProduct(externalMoldCode, nil)
}

// GetTotalStockOutByExternalMoldCodeAndProduct 获取指定外来模具和产品的总出库数量
func (d *ProductionStockOutDAO) GetTotalStockOutByExternalMoldCodeAndProduct(externalMoldCode string, productID *int64) (int, error) {
	var totalQty int
	query := GetDB().Model(&entity.ProductionStockOut{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("external_mold_code = ?", externalMoldCode)

	// 如果指定了productID，增加product_id过滤条件
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	} else {
		query = query.Where("product_id IS NULL")
	}

	err := query.Scan(&totalQty).Error
	return totalQty, err
}
