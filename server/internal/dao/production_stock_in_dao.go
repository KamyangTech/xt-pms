package dao

import (
	"strings"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProductionStockInDAO struct{}

// NewProductionStockInDAO 创建生产入库DAO
func NewProductionStockInDAO() *ProductionStockInDAO {
	return &ProductionStockInDAO{}
}

// Create 创建入库记录
func (d *ProductionStockInDAO) Create(stockIn *entity.ProductionStockIn) error {
	return GetDB().Create(stockIn).Error
}

// FindByID 根据ID查询入库记录
func (d *ProductionStockInDAO) FindByID(id int64) (*entity.ProductionStockIn, error) {
	var stockIn entity.ProductionStockIn
	err := GetDB().Preload("Order").Preload("Warehouse").Preload("Mold").First(&stockIn, id).Error
	if err != nil {
		return nil, err
	}
	return &stockIn, nil
}

// FindByCode 根据入库单号查询
func (d *ProductionStockInDAO) FindByCode(code string) (*entity.ProductionStockIn, error) {
	var stockIn entity.ProductionStockIn
	err := GetDB().Where("stock_code = ?", code).First(&stockIn).Error
	if err != nil {
		return nil, err
	}
	return &stockIn, nil
}

// Update 更新入库记录
func (d *ProductionStockInDAO) Update(stockIn *entity.ProductionStockIn) error {
	return GetDB().Save(stockIn).Error
}

// Delete 删除入库记录
func (d *ProductionStockInDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.ProductionStockIn{}, id).Error
}

// FindByOrderID 根据订单ID查询入库记录列表
func (d *ProductionStockInDAO) FindByOrderID(orderID int64) ([]entity.ProductionStockIn, error) {
	var stockIns []entity.ProductionStockIn
	err := GetDB().Where("order_id = ?", orderID).
		Preload("Warehouse").
		Order("stock_in_date DESC").
		Find(&stockIns).Error
	return stockIns, err
}

// GetOrderTotalStockInQty 获取订单的累计入库数量
func (d *ProductionStockInDAO) GetOrderTotalStockInQty(orderID int64) (int, error) {
	var totalQty int
	err := GetDB().Model(&entity.ProductionStockIn{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("order_id = ?", orderID).
		Scan(&totalQty).Error
	return totalQty, err
}

// FindByPage 分页查询入库记录
func (d *ProductionStockInDAO) FindByPage(page, pageSize int, orderCode string, productName string, stockCode string) ([]entity.ProductionStockIn, int64, error) {
	var stockIns []entity.ProductionStockIn
	var total int64

	query := GetDB().Model(&entity.ProductionStockIn{}).
		Preload("Order").
		Preload("Warehouse").
		Preload("Mold")

	if orderCode != "" {
		query = query.Joins("JOIN production_order ON production_order.id = production_stock_in.order_id").
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
	err = query.Order("stock_in_date DESC").Offset(offset).Limit(pageSize).Find(&stockIns).Error
	if err != nil {
		return nil, 0, err
	}

	return stockIns, total, nil
}

type ProductStockSummary struct {
	MoldID           *int64     `json:"mold_id"`            // 模具ID（自有模具）
	ExternalMoldCode *string    `json:"external_mold_code"` // 外来模具编号
	ProductID        *int64     `json:"product_id"`         // 产品ID
	MoldCode         string     `json:"mold_code"`
	ProductName      string     `json:"product_name"`       // 产品名称
	ProductImageURL  *string    `json:"product_image_url"`  // 产品图片URL
	CustomerName     *string    `json:"customer_name"`      // 订单客户名称
	TotalStockIn     int        `json:"total_stock_in"`     // 总入库数量
	TotalStockOut    int        `json:"total_stock_out"`    // 总出库数量
	CurrentStock     int        `json:"current_stock"`      // 当前库存
	LastStockInDate  *time.Time `json:"last_stock_in_date"` // 最后入库日期
}

// FindProductStockSummary 分页查询产品库存汇总（支持自有模具和外来模具）
// 【优化】使用分步查询替代复杂的 UNION ALL + 子查询，提升性能和可维护性
func (d *ProductionStockInDAO) FindProductStockSummary(page, pageSize int, moldCode string, productName string) ([]ProductStockSummary, int64, error) {
	var summaries []ProductStockSummary
	var total int64
	db := GetDB()

	// ========== 第一步：直接使用原生 SQL 构建 UNION ALL 查询 ==========
	// 自有模具汇总 + 外来模具汇总
	unionSQL := `(
		SELECT 
			psi.mold_id,
			NULL as external_mold_code,
			psi.product_id,
			MAX(psi.order_id) as order_id,
			COALESCE(SUM(psi.quantity), 0) as total_stock_in,
			MAX(psi.stock_in_date) as last_stock_in_date
		FROM production_stock_in psi
		WHERE psi.mold_id IS NOT NULL
		GROUP BY psi.mold_id, psi.product_id
		
		UNION ALL
		
		SELECT 
			NULL as mold_id,
			psi.external_mold_code,
			psi.product_id,
			MAX(psi.order_id) as order_id,
			COALESCE(SUM(psi.quantity), 0) as total_stock_in,
			MAX(psi.stock_in_date) as last_stock_in_date
		FROM production_stock_in psi
		WHERE psi.external_mold_code IS NOT NULL
		GROUP BY psi.external_mold_code, psi.product_id
	) as stock_summary`

	// ========== 第二步：关联模具、订单明细、客户信息，并计算出库数量 ==========
	mainQuery := db.Table(unionSQL).
		Select(`
			stock_summary.mold_id,
			stock_summary.external_mold_code,
			stock_summary.product_id,
			stock_summary.order_id,
			stock_summary.total_stock_in,
			stock_summary.last_stock_in_date,
			CASE 
				WHEN stock_summary.mold_id IS NOT NULL THEN m.mold_code
				ELSE stock_summary.external_mold_code
			END as mold_code,
			poi.product_name,
		COALESCE((
			SELECT poi2.product_image_url FROM production_order_item poi2
			WHERE 
				(
					(stock_summary.product_id IS NOT NULL AND poi2.product_id = stock_summary.product_id)
					OR (stock_summary.product_id IS NULL AND stock_summary.mold_id IS NOT NULL AND poi2.mold_id = stock_summary.mold_id)
					OR (stock_summary.product_id IS NULL AND stock_summary.external_mold_code IS NOT NULL AND poi2.external_mold_code COLLATE utf8mb4_unicode_ci = stock_summary.external_mold_code COLLATE utf8mb4_unicode_ci)
				)
				AND poi2.product_image_url IS NOT NULL AND poi2.product_image_url != ''
			ORDER BY poi2.id DESC LIMIT 1
		), '') as product_image_url,
			c.cus_name as customer_name,
			COALESCE(stock_out_stats.total_stock_out, 0) as total_stock_out,
			stock_summary.total_stock_in - COALESCE(stock_out_stats.total_stock_out, 0) as current_stock
		`).
		Joins("LEFT JOIN mold m ON m.id = stock_summary.mold_id").
		Joins("LEFT JOIN production_order po ON po.id = stock_summary.order_id").
		Joins("LEFT JOIN customer c ON c.id = po.customer_id").
		Joins(`LEFT JOIN production_order_item poi ON poi.order_id = stock_summary.order_id 
			AND ((stock_summary.product_id IS NOT NULL AND poi.product_id = stock_summary.product_id)
			     OR (stock_summary.product_id IS NULL AND (
					  (stock_summary.mold_id IS NOT NULL AND poi.mold_id = stock_summary.mold_id)
					  OR (stock_summary.external_mold_code IS NOT NULL AND poi.external_mold_code COLLATE utf8mb4_unicode_ci = stock_summary.external_mold_code COLLATE utf8mb4_unicode_ci)
				 )))`).
		Joins(`LEFT JOIN (
			SELECT 
				pso.mold_id,
				pso.external_mold_code,
				pso.product_id,
				SUM(pso.quantity) as total_stock_out
			FROM production_stock_out pso
			GROUP BY pso.mold_id, pso.external_mold_code, pso.product_id
		) as stock_out_stats ON (
			(stock_summary.mold_id IS NOT NULL AND stock_out_stats.mold_id = stock_summary.mold_id)
			OR (stock_summary.external_mold_code IS NOT NULL AND stock_out_stats.external_mold_code COLLATE utf8mb4_unicode_ci = stock_summary.external_mold_code COLLATE utf8mb4_unicode_ci)
		)
		AND ((stock_summary.product_id IS NOT NULL AND stock_out_stats.product_id = stock_summary.product_id)
		     OR (stock_summary.product_id IS NULL AND stock_out_stats.product_id IS NULL))`)

	// ========== 第三步：添加过滤条件 ==========
	if moldCode != "" {
		mainQuery = mainQuery.Where("mold_code LIKE ?", moldCode+"%")
	}
	if productName != "" {
		mainQuery = mainQuery.Where("poi.product_name LIKE ?", "%"+productName+"%")
	}

	// ========== 第四步：查询总数 ==========
	countQuery := db.Table(unionSQL)
	if moldCode != "" {
		countQuery = countQuery.Where("mold_code LIKE ?", moldCode+"%")
	}
	if productName != "" {
		countQuery = countQuery.Where("poi.product_name LIKE ?", "%"+productName+"%")
	}
	countQuery.Count(&total)

	// ========== 第五步：分页查询 ==========
	offset := (page - 1) * pageSize
	err := mainQuery.
		Order("stock_summary.last_stock_in_date DESC").
		Limit(pageSize).
		Offset(offset).
		Scan(&summaries).Error

	if err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}

// FindStockRecordsByMoldID 查询指定模具的所有入库记录
func (d *ProductionStockInDAO) FindStockRecordsByMoldID(moldID int64, page, pageSize int, sortOrder string) ([]entity.ProductionStockIn, int64, error) {
	return d.FindStockRecordsByMoldIDAndProduct(moldID, nil, page, pageSize, sortOrder)
}

// FindStockRecordsByMoldIDAndProduct 查询指定模具和产品的所有入库记录
func (d *ProductionStockInDAO) FindStockRecordsByMoldIDAndProduct(moldID int64, productID *int64, page, pageSize int, sortOrder string) ([]entity.ProductionStockIn, int64, error) {
	var records []entity.ProductionStockIn
	var total int64

	query := GetDB().Model(&entity.ProductionStockIn{}).
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
	orderClause := "stock_in_date DESC"
	if sortOrder == "asc" {
		orderClause = "stock_in_date ASC"
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

// FindProductStockSummaryByMoldID 查询指定模具的库存汇总信息
func (d *ProductionStockInDAO) FindProductStockSummaryByMoldID(moldID int64) (ProductStockSummary, error) {
	return d.FindProductStockSummaryByMoldIDAndProduct(moldID, nil)
}

// FindProductStockSummaryByMoldIDAndProduct 查询指定模具和产品的库存汇总信息
func (d *ProductionStockInDAO) FindProductStockSummaryByMoldIDAndProduct(moldID int64, productID *int64) (ProductStockSummary, error) {
	var summary ProductStockSummary

	// 单次查询：将出库总数和产品图片嵌入为子查询，替代原先 3 次独立查询
	var summaryData struct {
		MoldID          *int64     `json:"mold_id"`
		MoldCode        string     `json:"mold_code"`
		ProductName     string     `json:"product_name"`
		CustomerName    *string    `json:"customer_name"`
		LastStockInDate *time.Time `json:"last_stock_in_date"`
		ProductImageURL string     `json:"product_image_url"`
		TotalStockIn    int        `json:"total_stock_in"`
		TotalStockOut   int        `json:"total_stock_out"`
	}

	query := GetDB().Table("production_stock_in psi").
		Select(`
			MAX(psi.mold_id) as mold_id,
			MAX(m.mold_code) as mold_code,
			MAX(psi.product_name) as product_name,
			MAX(c.cus_name) as customer_name,
			MAX(psi.stock_in_date) as last_stock_in_date,
			COALESCE(SUM(psi.quantity), 0) as total_stock_in,
			COALESCE((
				SELECT SUM(pso.quantity) FROM production_stock_out pso
				WHERE pso.mold_id = psi.mold_id
				  AND (pso.product_id = psi.product_id OR (pso.product_id IS NULL AND psi.product_id IS NULL))
			), 0) as total_stock_out,
			COALESCE((
				SELECT poi.product_image_url FROM production_order_item poi
				WHERE poi.mold_id = psi.mold_id AND poi.product_name = psi.product_name
				ORDER BY poi.id ASC LIMIT 1
			), '') as product_image_url
		`).
		Joins("LEFT JOIN mold m ON m.id = psi.mold_id").
		Joins("LEFT JOIN production_order po ON po.id = psi.order_id").
		Joins("LEFT JOIN customer c ON c.id = po.customer_id").
		Where("psi.mold_id = ?", moldID)

	if productID != nil {
		query = query.Where("psi.product_id = ?", *productID)
	} else {
		query = query.Where("psi.product_id IS NULL")
	}

	// 添加 GROUP BY 以兼容 sql_mode=only_full_group_by（子查询引用了 psi.product_name）
	query = query.Group("psi.mold_id, psi.product_id, psi.product_name, m.mold_code, c.cus_name")

	err := query.Scan(&summaryData).Error
	if err != nil {
		return summary, err
	}

	summary.MoldID = &moldID
	summary.MoldCode = summaryData.MoldCode
	summary.ProductName = summaryData.ProductName
	summary.CustomerName = summaryData.CustomerName
	summary.ProductImageURL = &summaryData.ProductImageURL
	summary.LastStockInDate = summaryData.LastStockInDate
	summary.TotalStockIn = summaryData.TotalStockIn
	summary.TotalStockOut = summaryData.TotalStockOut
	summary.CurrentStock = summaryData.TotalStockIn - summaryData.TotalStockOut
	summary.ProductID = productID

	return summary, nil
}

// FindStockRecordsByExternalMoldCode 查询指定外来模具的所有入库记录
func (d *ProductionStockInDAO) FindStockRecordsByExternalMoldCode(externalMoldCode string, page, pageSize int, sortOrder string) ([]entity.ProductionStockIn, int64, error) {
	return d.FindStockRecordsByExternalMoldCodeAndProduct(externalMoldCode, nil, page, pageSize, sortOrder)
}

// FindStockRecordsByExternalMoldCodeAndProduct 查询指定外来模具和产品的所有入库记录
func (d *ProductionStockInDAO) FindStockRecordsByExternalMoldCodeAndProduct(externalMoldCode string, productID *int64, page, pageSize int, sortOrder string) ([]entity.ProductionStockIn, int64, error) {
	var records []entity.ProductionStockIn
	var total int64

	query := GetDB().Model(&entity.ProductionStockIn{}).
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
	orderClause := "stock_in_date DESC"
	if sortOrder == "asc" {
		orderClause = "stock_in_date ASC"
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

// FindProductStockSummaryByExternalMoldCode 查询指定外来模具的库存汇总信息
func (d *ProductionStockInDAO) FindProductStockSummaryByExternalMoldCode(externalMoldCode string) (ProductStockSummary, error) {
	return d.FindProductStockSummaryByExternalMoldCodeAndProduct(externalMoldCode, nil)
}

// FindProductStockSummaryByExternalMoldCodeAndProduct 查询指定外来模具和产品的库存汇总信息
func (d *ProductionStockInDAO) FindProductStockSummaryByExternalMoldCodeAndProduct(externalMoldCode string, productID *int64) (ProductStockSummary, error) {
	var summary ProductStockSummary

	// 单次查询：将出库总数和产品图片嵌入为子查询，替代原先 4 次独立查询
	var summaryData struct {
		ExternalMoldCode string     `json:"external_mold_code"`
		ProductName      string     `json:"product_name"`
		CustomerName     *string    `json:"customer_name"`
		LastStockInDate  *time.Time `json:"last_stock_in_date"`
		ProductImageURL  string     `json:"product_image_url"`
		TotalStockIn     int        `json:"total_stock_in"`
		TotalStockOut    int        `json:"total_stock_out"`
	}

	query := GetDB().Table("production_stock_in").
		Select(`
			MAX(production_stock_in.external_mold_code) as external_mold_code,
			MAX(production_stock_in.product_name) as product_name,
			MAX(c.cus_name) as customer_name,
			MAX(production_stock_in.stock_in_date) as last_stock_in_date,
			COALESCE(SUM(production_stock_in.quantity), 0) as total_stock_in,
			COALESCE((
				SELECT SUM(pso.quantity) FROM production_stock_out pso
				WHERE pso.external_mold_code = production_stock_in.external_mold_code
				  AND (pso.product_id = production_stock_in.product_id
				       OR (pso.product_id IS NULL AND production_stock_in.product_id IS NULL))
			), 0) as total_stock_out,
			COALESCE((
				SELECT poi.product_image_url FROM production_order_item poi
				WHERE poi.external_mold_code = production_stock_in.external_mold_code
				  AND poi.product_name = production_stock_in.product_name
				ORDER BY poi.id ASC LIMIT 1
			), '') as product_image_url
		`).
		Joins("LEFT JOIN production_order po ON po.id = production_stock_in.order_id").
		Joins("LEFT JOIN customer c ON c.id = po.customer_id").
		Where("production_stock_in.external_mold_code = ?", externalMoldCode)

	if productID != nil {
		query = query.Where("production_stock_in.product_id = ?", *productID)
	} else {
		query = query.Where("production_stock_in.product_id IS NULL")
	}

	// 添加 GROUP BY 以兼容 sql_mode=only_full_group_by（子查询引用了 product_name 和 external_mold_code）
	query = query.Group("production_stock_in.external_mold_code, production_stock_in.product_id, production_stock_in.product_name, c.cus_name")

	err := query.Scan(&summaryData).Error
	if err != nil {
		return summary, err
	}

	summary.MoldCode = summaryData.ExternalMoldCode
	summary.ExternalMoldCode = &externalMoldCode
	summary.ProductName = summaryData.ProductName
	summary.CustomerName = summaryData.CustomerName
	summary.ProductImageURL = &summaryData.ProductImageURL
	summary.LastStockInDate = summaryData.LastStockInDate
	summary.TotalStockIn = summaryData.TotalStockIn
	summary.TotalStockOut = summaryData.TotalStockOut
	summary.CurrentStock = summaryData.TotalStockIn - summaryData.TotalStockOut
	summary.ProductID = productID

	return summary, nil
}

// FindMergedStockRecords 使用 UNION ALL 合并查询入库和出库记录（优化版）
// 【优化】在数据库层完成合并、排序、分页，避免全量查询+内存处理
func (d *ProductionStockInDAO) FindMergedStockRecords(moldID int64, productID *int64, page, pageSize int, sortOrder, statusFilters string) ([]map[string]interface{}, int64, error) {
	// 确定排序方向
	orderDirection := "DESC"
	if sortOrder == "asc" {
		orderDirection = "ASC"
	}

	unionSQL := `
		(
			SELECT 
				'in' as record_type,
				psi.id,
				psi.stock_code,
				psi.order_id,
				psi.warehouse_id,
				psi.mold_id,
				psi.product_name,
				psi.quantity,
				psi.batch_no,
				psi.stock_in_date as record_date,
				psi.operator_id,
				psi.remark,
				psi.create_time,
				psi.update_time,
				NULL as customer_name,
				NULL as out_type,
				NULL as out_reason,
				NULL as out_reason_remark,
				NULL as delivery_date,
				NULL as delivery_method,
				NULL as express_company,
				NULL as express_no,
				NULL as photos,
				NULL as total_weight,
				NULL as driver_phone,
				NULL as driver_plate
			FROM production_stock_in psi
			WHERE psi.mold_id = ?
			
			UNION ALL
			
			SELECT 
				'out' as record_type,
				pso.id,
				pso.stock_code,
				pso.order_id,
				pso.warehouse_id,
				pso.mold_id,
				pso.product_name,
				pso.quantity * -1 as quantity,
				pso.batch_no,
				pso.stock_out_date as record_date,
				pso.operator_id,
				pso.remark,
				pso.create_time,
				pso.update_time,
				pso.customer_name,
				pso.out_type,
				pso.out_reason,
				pso.out_reason_remark,
				pso.delivery_date,
				pso.delivery_method,
				pso.express_company,
				pso.express_no,
				pso.photos,
				pso.total_weight,
				pso.driver_phone,
				pso.driver_plate
			FROM production_stock_out pso
			WHERE pso.mold_id = ?
		) as merged_records
	`

	var whereConditions []string
	var whereArgs []interface{}

	// 添加 product_id 过滤（必须外层括号包裹 OR，否则 AND 优先级高于 OR 会破坏语义）
	if productID != nil {
		whereConditions = append(whereConditions, "((record_type = 'in' AND EXISTS (SELECT 1 FROM production_stock_in psi WHERE psi.id = merged_records.id AND psi.product_id = ?)) OR (record_type = 'out' AND EXISTS (SELECT 1 FROM production_stock_out pso WHERE pso.id = merged_records.id AND pso.product_id = ?)))")
		whereArgs = append(whereArgs, *productID, *productID)
	} else {
		whereConditions = append(whereConditions, "((record_type = 'in' AND EXISTS (SELECT 1 FROM production_stock_in psi WHERE psi.id = merged_records.id AND psi.product_id IS NULL)) OR (record_type = 'out' AND EXISTS (SELECT 1 FROM production_stock_out pso WHERE pso.id = merged_records.id AND pso.product_id IS NULL)))")
	}

	// 状态筛选
	if statusFilters != "" {
		filterArray := strings.Split(statusFilters, ",")
		var statusConditions []string
		for _, filter := range filterArray {
			filter = strings.TrimSpace(filter)
			if filter == "in" {
				statusConditions = append(statusConditions, "record_type = 'in'")
			} else if filter == "shipment" {
				statusConditions = append(statusConditions, "(record_type = 'out' AND (out_type IS NULL OR out_type = 'normal' OR out_type = 'shipment'))")
			} else {
				statusConditions = append(statusConditions, "(record_type = 'out' AND out_reason = ?)")
				whereArgs = append(whereArgs, filter)
			}
		}
		if len(statusConditions) > 0 {
			whereConditions = append(whereConditions, "("+strings.Join(statusConditions, " OR ")+")")
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM " + unionSQL + " " + whereClause
	countArgs := append([]interface{}{moldID, moldID}, whereArgs...)
	GetDB().Raw(countSQL, countArgs...).Scan(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	querySQL := "SELECT * FROM " + unionSQL + " " + whereClause + " ORDER BY record_date " + orderDirection + " LIMIT ? OFFSET ?"
	queryArgs := append(countArgs, pageSize, offset)

	var records []map[string]interface{}
	err := GetDB().Raw(querySQL, queryArgs...).Scan(&records).Error
	if err != nil {
		return nil, 0, err
	}

	// 关联查询订单和仓库信息（批量查询避免 N+1）
	if len(records) > 0 {
		orderIDSet := make(map[int64]struct{})
		warehouseIDSet := make(map[int64]struct{})
		for _, record := range records {
			if orderID, ok := record["order_id"].(int64); ok && orderID > 0 {
				orderIDSet[orderID] = struct{}{}
			}
			if warehouseID, ok := record["warehouse_id"].(int64); ok && warehouseID > 0 {
				warehouseIDSet[warehouseID] = struct{}{}
			}
		}

		orderIDs := make([]int64, 0, len(orderIDSet))
		for id := range orderIDSet {
			orderIDs = append(orderIDs, id)
		}
		warehouseIDs := make([]int64, 0, len(warehouseIDSet))
		for id := range warehouseIDSet {
			warehouseIDs = append(warehouseIDs, id)
		}

		// 批量查询订单
		orderMap := make(map[int64]*entity.ProductionOrder)
		if len(orderIDs) > 0 {
			var orders []entity.ProductionOrder
			GetDB().Where("id IN ?", orderIDs).Find(&orders)
			for i := range orders {
				orderMap[orders[i].ID] = &orders[i]
			}
		}

		// 批量查询仓库
		warehouseMap := make(map[int64]*entity.Warehouse)
		if len(warehouseIDs) > 0 {
			var warehouses []entity.Warehouse
			GetDB().Where("id IN ?", warehouseIDs).Find(&warehouses)
			for i := range warehouses {
				warehouseMap[warehouses[i].ID] = &warehouses[i]
			}
		}

		// 填充关联数据
		for i := range records {
			// 【修复】将 record_type 重命名为 type，以匹配前端期望
			if recordType, ok := records[i]["record_type"]; ok {
				records[i]["type"] = recordType
				delete(records[i], "record_type") // 删除旧字段，避免混淆
			}

			if orderID, ok := records[i]["order_id"].(int64); ok && orderID > 0 {
				if order, exists := orderMap[orderID]; exists {
					records[i]["order"] = order
				}
			}
			if warehouseID, ok := records[i]["warehouse_id"].(int64); ok && warehouseID > 0 {
				if warehouse, exists := warehouseMap[warehouseID]; exists {
					records[i]["warehouse"] = warehouse
				}
			}
		}
	}

	return records, total, nil
}

// FindMergedStockRecordsByExternalMoldCode 使用 UNION ALL 合并查询外来模具的入库和出库记录（优化版）
// 在数据库层完成合并、过滤、排序和分页，避免全量加载到 Go 内存中处理
func (d *ProductionStockInDAO) FindMergedStockRecordsByExternalMoldCode(externalMoldCode string, productID *int64, page, pageSize int, sortOrder, statusFilters string) ([]map[string]interface{}, int64, error) {
	// 确定排序方向
	orderDirection := "DESC"
	if sortOrder == "asc" {
		orderDirection = "ASC"
	}

	unionSQL := `
		(
			SELECT 
				'in' as record_type,
				psi.id,
				psi.stock_code,
				psi.order_id,
				psi.warehouse_id,
				psi.mold_id,
				psi.external_mold_code,
				psi.product_name,
				psi.quantity,
				psi.batch_no,
				psi.stock_in_date as record_date,
				psi.operator_id,
				psi.remark,
				psi.create_time,
				psi.update_time,
				NULL as customer_name,
				NULL as out_type,
				NULL as out_reason,
				NULL as out_reason_remark,
				NULL as delivery_date,
				NULL as delivery_method,
				NULL as express_company,
				NULL as express_no,
				NULL as photos,
				NULL as total_weight,
				NULL as driver_phone,
				NULL as driver_plate
			FROM production_stock_in psi
			WHERE psi.external_mold_code = ?
			
			UNION ALL
			
			SELECT 
				'out' as record_type,
				pso.id,
				pso.stock_code,
				pso.order_id,
				pso.warehouse_id,
				pso.mold_id,
				pso.external_mold_code,
				pso.product_name,
				pso.quantity * -1 as quantity,
				pso.batch_no,
				pso.stock_out_date as record_date,
				pso.operator_id,
				pso.remark,
				pso.create_time,
				pso.update_time,
				pso.customer_name,
				pso.out_type,
				pso.out_reason,
				pso.out_reason_remark,
				pso.delivery_date,
				pso.delivery_method,
				pso.express_company,
				pso.express_no,
				pso.photos,
				pso.total_weight,
				pso.driver_phone,
				pso.driver_plate
			FROM production_stock_out pso
			WHERE pso.external_mold_code = ?
		) as merged_records
	`

	var whereConditions []string
	var whereArgs []interface{}

	// 添加 product_id 过滤（必须外层括号包裹 OR，否则 AND 优先级高于 OR 会破坏语义）
	if productID != nil {
		whereConditions = append(whereConditions, "((record_type = 'in' AND EXISTS (SELECT 1 FROM production_stock_in psi WHERE psi.id = merged_records.id AND psi.product_id = ?)) OR (record_type = 'out' AND EXISTS (SELECT 1 FROM production_stock_out pso WHERE pso.id = merged_records.id AND pso.product_id = ?)))")
		whereArgs = append(whereArgs, *productID, *productID)
	}

	// 状态筛选
	if statusFilters != "" {
		filterArray := strings.Split(statusFilters, ",")
		var statusConditions []string
		for _, filter := range filterArray {
			filter = strings.TrimSpace(filter)
			if filter == "in" {
				statusConditions = append(statusConditions, "record_type = 'in'")
			} else if filter == "shipment" {
				statusConditions = append(statusConditions, "(record_type = 'out' AND (out_type IS NULL OR out_type = 'normal' OR out_type = 'shipment'))")
			} else {
				statusConditions = append(statusConditions, "(record_type = 'out' AND out_reason = ?)")
				whereArgs = append(whereArgs, filter)
			}
		}
		if len(statusConditions) > 0 {
			whereConditions = append(whereConditions, "("+strings.Join(statusConditions, " OR ")+")")
		}
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM " + unionSQL + " " + whereClause
	countArgs := append([]interface{}{externalMoldCode, externalMoldCode}, whereArgs...)
	GetDB().Raw(countSQL, countArgs...).Scan(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	querySQL := "SELECT * FROM " + unionSQL + " " + whereClause + " ORDER BY record_date " + orderDirection + " LIMIT ? OFFSET ?"
	queryArgs := append(countArgs, pageSize, offset)

	var records []map[string]interface{}
	err := GetDB().Raw(querySQL, queryArgs...).Scan(&records).Error
	if err != nil {
		return nil, 0, err
	}

	// 关联查询订单和仓库信息（批量查询避免 N+1）
	if len(records) > 0 {
		orderIDSet := make(map[int64]struct{})
		warehouseIDSet := make(map[int64]struct{})
		for _, record := range records {
			if orderID, ok := record["order_id"].(int64); ok && orderID > 0 {
				orderIDSet[orderID] = struct{}{}
			}
			if warehouseID, ok := record["warehouse_id"].(int64); ok && warehouseID > 0 {
				warehouseIDSet[warehouseID] = struct{}{}
			}
		}

		orderIDs := make([]int64, 0, len(orderIDSet))
		for id := range orderIDSet {
			orderIDs = append(orderIDs, id)
		}
		warehouseIDs := make([]int64, 0, len(warehouseIDSet))
		for id := range warehouseIDSet {
			warehouseIDs = append(warehouseIDs, id)
		}

		// 批量查询订单
		orderMap := make(map[int64]*entity.ProductionOrder)
		if len(orderIDs) > 0 {
			var orders []entity.ProductionOrder
			GetDB().Where("id IN ?", orderIDs).Find(&orders)
			for i := range orders {
				orderMap[orders[i].ID] = &orders[i]
			}
		}

		// 批量查询仓库
		warehouseMap := make(map[int64]*entity.Warehouse)
		if len(warehouseIDs) > 0 {
			var warehouses []entity.Warehouse
			GetDB().Where("id IN ?", warehouseIDs).Find(&warehouses)
			for i := range warehouses {
				warehouseMap[warehouses[i].ID] = &warehouses[i]
			}
		}

		// 填充关联数据
		for i := range records {
			if recordType, ok := records[i]["record_type"]; ok {
				records[i]["type"] = recordType
				delete(records[i], "record_type")
			}

			if orderID, ok := records[i]["order_id"].(int64); ok && orderID > 0 {
				if order, exists := orderMap[orderID]; exists {
					records[i]["order"] = order
				}
			}
			if warehouseID, ok := records[i]["warehouse_id"].(int64); ok && warehouseID > 0 {
				if warehouse, exists := warehouseMap[warehouseID]; exists {
					records[i]["warehouse"] = warehouse
				}
			}
		}
	}

	return records, total, nil
}
