package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProductService struct {
	productDAO *dao.ProductDAO
}

func NewProductService() *ProductService {
	return &ProductService{productDAO: dao.NewProductDAO()}
}

// GenerateProductCode 生成产品编码 P-YYNNNNNN (YY是年份后两位,NNNNNN是6位序号)
func (s *ProductService) GenerateProductCode() (string, error) {
	now := time.Now()
	yearStr := now.Format("06") // 年份后两位
	prefix := fmt.Sprintf("P-%s-", yearStr)

	// 查询当年最大序号
	var maxCode string
	dao.GetDB().Model(&entity.Product{}).
		Where("pro_code LIKE ?", prefix+"%").
		Order("pro_code DESC").
		Limit(1).
		Pluck("pro_code", &maxCode)

	seq := 1
	if maxCode != "" {
		parts := strings.Split(maxCode, "-")
		if len(parts) == 3 {
			if num, err := strconv.Atoi(parts[2]); err == nil {
				seq = num + 1
			}
		}
	}

	return fmt.Sprintf("%s%06d", prefix, seq), nil
}

func (s *ProductService) CreateProduct(productDTO dto.ProductDTO) (*entity.Product, error) {
	// 如果用户未提供编码，则自动生成
	productCode := productDTO.ProCode
	if productCode == "" {
		var err error
		productCode, err = s.GenerateProductCode()
		if err != nil {
			return nil, errors.New("生成产品编码失败：" + err.Error())
		}
	}

	existing, _ := s.productDAO.FindByCode(productCode)
	if existing != nil {
		return nil, errors.New("产品编码已存在")
	}

	// 检查产品名称是否已存在（避免重复创建）
	existingByName, _ := s.productDAO.FindByName(productDTO.ProName)
	if existingByName != nil {
		return existingByName, nil // 返回已存在的产品
	}

	product := &entity.Product{
		ProCode:    productCode,
		ProName:    productDTO.ProName,
		Spec:       productDTO.Spec,
		SpecDetail: productDTO.SpecDetail,
		MoldID:     productDTO.MoldID,
		Unit:       productDTO.Unit,
		Price:      productDTO.Price,
		Weight:     productDTO.Weight,
		Status:     productDTO.Status,
		Remark:     productDTO.Remark,
		ImageURL:   productDTO.ImageURL,
	}

	if product.Status == 0 {
		product.Status = 1
	}

	err := s.productDAO.Create(product)
	if err != nil {
		// 并发场景下可能唯一约束冲突，回退查询已有记录
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "duplicate key") {
			existingByName, findErr := s.productDAO.FindByName(productDTO.ProName)
			if findErr == nil && existingByName != nil {
				return existingByName, nil
			}
		}
		return nil, err
	}
	return product, nil
}

func (s *ProductService) GetProductByID(id int64) (*entity.Product, error) {
	return s.productDAO.FindByID(id)
}

func (s *ProductService) GetAllProducts() ([]*entity.Product, error) {
	return s.productDAO.FindAll()
}

func (s *ProductService) GetProductsByPage(page, pageSize int) ([]*entity.Product, int64, error) {
	result, err := s.productDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	// result.List 是 *[]*entity.Product，需要解引用
	if result.List == nil {
		return []*entity.Product{}, result.Total, nil
	}
	return *(result.List.(*[]*entity.Product)), result.Total, nil
}

// GetProductsWithStockByPage 获取带库存信息的产品列表（分页）
// 优化：批量查询替代 N+1 循环查询，将 ~41 次数据库查询压缩为 ~6 次
func (s *ProductService) GetProductsWithStockByPage(page, pageSize int) ([]dto.ProductWithStockVO, int64, error) {
	// 先获取产品列表
	products, total, err := s.GetProductsByPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	if len(products) == 0 {
		return []dto.ProductWithStockVO{}, total, nil
	}

	// 收集 IDs 和产品名，用于批量查询
	productNames := make([]string, len(products))
	moldIDs := make([]int64, 0)
	productIDs := make([]int64, 0, len(products))
	for i, p := range products {
		productNames[i] = p.ProName
		if p.MoldID != nil {
			moldIDs = append(moldIDs, *p.MoldID)
		}
		productIDs = append(productIDs, p.ID)
	}

	// ========== 批量查询 1：模具编号 ==========
	moldCodeMap := make(map[int64]string)
	if len(moldIDs) > 0 {
		var molds []entity.Mold
		dao.GetDB().Where("id IN ?", moldIDs).Find(&molds)
		for _, m := range molds {
			moldCodeMap[m.ID] = m.MoldCode
		}
	}

	// ========== 批量查询 2+3：入库和出库汇总（按 product_name 分组） ==========
	type stockSummary struct {
		ProductName string
		TotalQty    int
	}

	stockInMap := make(map[string]int)
	if len(productNames) > 0 {
		var inResults []stockSummary
		dao.GetDB().Table("production_stock_in").
			Select("product_name, COALESCE(SUM(quantity), 0) AS total_qty").
			Where("product_name IN ?", productNames).
			Group("product_name").
			Find(&inResults)
		for _, r := range inResults {
			stockInMap[r.ProductName] = r.TotalQty
		}
	}

	stockOutMap := make(map[string]int)
	if len(productNames) > 0 {
		var outResults []stockSummary
		dao.GetDB().Table("production_stock_out").
			Select("product_name, COALESCE(SUM(quantity), 0) AS total_qty").
			Where("product_name IN ?", productNames).
			Group("product_name").
			Find(&outResults)
		for _, r := range outResults {
			stockOutMap[r.ProductName] = r.TotalQty
		}
	}

	// ========== 批量查询 4：产品图片（回退用） ==========
	prodImageMap := make(map[int64]string)
	if len(productIDs) > 0 {
		type imageRow struct {
			ProductID int64
			ImageURL  string
		}
		// 对每个 product_id 取最新的一条图片
		var images []imageRow
		dao.GetDB().Table("production_order_item").
			Select("product_id, product_image_url AS image_url").
			Where("product_id IN ? AND product_image_url IS NOT NULL AND product_image_url != ''", productIDs).
			Order("id DESC").
			Find(&images)
		// 取每个 product_id 的第一条（因为按 id DESC 排序，第一条即最新）
		for _, img := range images {
			if _, exists := prodImageMap[img.ProductID]; !exists {
				prodImageMap[img.ProductID] = img.ImageURL
			}
		}
	}

	// ========== 组装结果（纯内存操作，无 DB 查询） ==========
	result := make([]dto.ProductWithStockVO, 0, len(products))
	for _, product := range products {
		totalIn := stockInMap[product.ProName]
		totalOut := stockOutMap[product.ProName]

		vo := dto.ProductWithStockVO{
			ID:           product.ID,
			ProCode:      product.ProCode,
			ProName:      product.ProName,
			Spec:         product.Spec,
			SpecDetail:   product.SpecDetail,
			MoldID:       product.MoldID,
			Unit:         product.Unit,
			Price:        product.Price,
			Weight:       product.Weight,
			Status:       product.Status,
			Remark:       product.Remark,
			ImageURL:     product.ImageURL,
			TotalIn:      totalIn,
			TotalOut:     totalOut,
			CurrentStock: totalIn - totalOut,
			CreateTime:   product.CreateTime,
			UpdateTime:   product.UpdateTime,
		}

		// 模具编号
		if product.MoldID != nil {
			vo.MoldCode = moldCodeMap[*product.MoldID]
		}

		// 产品图片：优先手动上传，回退到生产记录
		if product.ImageURL == "" {
			vo.ProductImageURL = prodImageMap[product.ID]
		} else {
			vo.ProductImageURL = product.ImageURL
		}

		result = append(result, vo)
	}

	return result, total, nil
}

func (s *ProductService) UpdateProduct(productDTO dto.ProductUpdateDTO) error {
	product, err := s.productDAO.FindByID(productDTO.ID)
	if err != nil {
		return errors.New("产品不存在")
	}

	if productDTO.ProCode != "" {
		product.ProCode = productDTO.ProCode
	}
	if productDTO.ProName != "" {
		product.ProName = productDTO.ProName
	}
	product.Spec = productDTO.Spec
	product.SpecDetail = productDTO.SpecDetail
	product.MoldID = productDTO.MoldID
	product.Unit = productDTO.Unit
	product.Price = productDTO.Price
	product.Weight = productDTO.Weight
	product.Status = productDTO.Status
	product.Remark = productDTO.Remark
	product.ImageURL = productDTO.ImageURL

	return s.productDAO.Update(product)
}

func (s *ProductService) DeleteProduct(id int64) error {
	// 1. 检查是否被活跃的生产订单引用
	var activeOrderCount int64
	dao.DB.Table("production_order_item poi").
		Joins("JOIN production_order po ON poi.order_id = po.id").
		Where("poi.product_id = ? AND po.status NOT IN (?, ?)",
			id, constant.ProductionOrderStatusCancelled, constant.ProductionOrderStatusCompleted).
		Count(&activeOrderCount)
	if activeOrderCount > 0 {
		return fmt.Errorf("该产品被 %d 个未完成的生产订单引用，无法删除，请先完成或取消相关订单", activeOrderCount)
	}

	// 2. 检查是否有 BOM 记录
	var bomCount int64
	dao.DB.Model(&entity.ProductBOM{}).
		Where("product_id = ?", id).
		Count(&bomCount)
	if bomCount > 0 {
		return fmt.Errorf("该产品关联了 %d 条 BOM 记录，无法删除，请先清理 BOM 数据", bomCount)
	}

	// 3. 检查是否有模具关联
	var relationCount int64
	dao.DB.Model(&entity.MoldProductRelation{}).
		Where("product_id = ?", id).
		Count(&relationCount)
	if relationCount > 0 {
		return fmt.Errorf("该产品关联了 %d 个模具，无法删除，请先解除模具关联", relationCount)
	}

	return s.productDAO.Delete(id)
}

func (s *ProductService) UpdateProductStatus(id int64, status int8) error {
	return s.productDAO.UpdateStatus(id, status)
}

// SyncProductsFromProductionRecords 从生产记录中同步产品
func (s *ProductService) SyncProductsFromProductionRecords() error {
	// 从 production_stock_in 表中提取产品
	s.syncProductsFromTable("production_stock_in")
	// 从 production_stock_out 表中提取产品
	s.syncProductsFromTable("production_stock_out")
	// 从 production_order_item 表中提取产品
	s.syncProductsFromTable("production_order_item")
	return nil
}

// syncProductsFromTable 从指定表中同步产品（批量查询避免 N+1）
func (s *ProductService) syncProductsFromTable(tableName string) {
	var productNames []string

	dao.GetDB().Table(tableName).
		Select("DISTINCT product_name").
		Where("product_name IS NOT NULL AND product_name != ''").
		Pluck("product_name", &productNames)

	if len(productNames) == 0 {
		return
	}

	// 批量查询已存在的产品（避免 N+1）
	existingProducts, _ := s.productDAO.FindByNames(productNames)
	existingMap := make(map[string]bool, len(existingProducts))
	for _, p := range existingProducts {
		existingMap[p.ProName] = true
	}

	// 为每个不存在的新产品名称创建产品档案
	for _, name := range productNames {
		if existingMap[name] {
			continue
		}

		code, err := s.GenerateProductCode()
		if err != nil {
			continue
		}

		product := &entity.Product{
			ProCode: code,
			ProName: name,
			Spec:    "",
			Unit:    "个",
			Price:   0.00,
			Weight:  0.00,
			Status:  1,
			Remark:  "自动创建自历史数据",
		}

		s.productDAO.Create(product)
	}
}
