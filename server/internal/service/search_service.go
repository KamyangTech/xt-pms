package service

import (
	"fmt"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SearchService struct {
	moldDAO            *dao.MoldDAO
	productDAO         *dao.ProductDAO
	customerDAO        *dao.CustomerDAO
	productionOrderDAO *dao.ProductionOrderDAO
	moldOrderDAO       *dao.MoldOrderDAO
	moldWorkOrderDAO   *dao.MoldWorkOrderDAO
}

func NewSearchService() *SearchService {
	return &SearchService{
		moldDAO:            dao.NewMoldDAO(),
		productDAO:         dao.NewProductDAO(),
		customerDAO:        dao.NewCustomerDAO(),
		productionOrderDAO: dao.NewProductionOrderDAO(),
		moldOrderDAO:       dao.NewMoldOrderDAO(),
		moldWorkOrderDAO:   dao.NewMoldWorkOrderDAO(),
	}
}

// GlobalSearch 跨模块全局搜索
func (s *SearchService) GlobalSearch(keyword string, page, pageSize int) (*dto.SearchResponse, error) {
	if keyword == "" {
		return &dto.SearchResponse{Items: []dto.SearchResultItem{}, Total: 0}, nil
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	like := "%" + keyword + "%"
	var allItems []dto.SearchResultItem

	// 1. 搜索模具档案
	molds, err := s.searchMolds(like)
	if err != nil {
		return nil, fmt.Errorf("搜索模具失败: %w", err)
	}
	allItems = append(allItems, molds...)

	// 2. 搜索产品
	products, err := s.searchProducts(like)
	if err != nil {
		return nil, fmt.Errorf("搜索产品失败: %w", err)
	}
	allItems = append(allItems, products...)

	// 3. 搜索客户
	customers, err := s.searchCustomers(like)
	if err != nil {
		return nil, fmt.Errorf("搜索客户失败: %w", err)
	}
	allItems = append(allItems, customers...)

	// 4. 搜索生产订单
	prodOrders, err := s.searchProductionOrders(like)
	if err != nil {
		return nil, fmt.Errorf("搜索生产订单失败: %w", err)
	}
	allItems = append(allItems, prodOrders...)

	// 5. 搜索模具订单
	moldOrders, err := s.searchMoldOrders(like)
	if err != nil {
		return nil, fmt.Errorf("搜索模具订单失败: %w", err)
	}
	allItems = append(allItems, moldOrders...)

	// 6. 搜索模具工单
	workOrders, err := s.searchMoldWorkOrders(like)
	if err != nil {
		return nil, fmt.Errorf("搜索模具工单失败: %w", err)
	}
	allItems = append(allItems, workOrders...)

	// 7. 搜索综合订单
	masterOrders, err := s.searchMasterOrders(like)
	if err != nil {
		return nil, fmt.Errorf("搜索综合订单失败: %w", err)
	}
	allItems = append(allItems, masterOrders...)

	total := int64(len(allItems))

	// 手动分页
	start := (page - 1) * pageSize
	if start >= int(total) {
		return &dto.SearchResponse{Items: []dto.SearchResultItem{}, Total: total}, nil
	}
	end := start + pageSize
	if end > int(total) {
		end = int(total)
	}

	return &dto.SearchResponse{
		Items: allItems[start:end],
		Total: total,
	}, nil
}

func (s *SearchService) searchMolds(like string) ([]dto.SearchResultItem, error) {
	var molds []entity.Mold
	if err := dao.GetDB().Where("mold_code LIKE ? OR mold_name LIKE ?", like, like).
		Select("id, mold_code, mold_name").
		Limit(20).Find(&molds).Error; err != nil {
		return nil, err
	}
	items := make([]dto.SearchResultItem, 0, len(molds))
	for _, m := range molds {
		items = append(items, dto.SearchResultItem{
			ID:          m.ID,
			Title:       fmt.Sprintf("%s（%s）", m.MoldCode, m.MoldName),
			Description: "模具档案",
			Module:      "mold",
			ModuleName:  "模具管理",
			Path:        fmt.Sprintf("/mold/detail/%d", m.ID),
		})
	}
	return items, nil
}

func (s *SearchService) searchProducts(like string) ([]dto.SearchResultItem, error) {
	var products []entity.Product
	if err := dao.GetDB().Where("pro_code LIKE ? OR pro_name LIKE ?", like, like).
		Select("id, pro_code, pro_name").
		Limit(20).Find(&products).Error; err != nil {
		return nil, err
	}
	items := make([]dto.SearchResultItem, 0, len(products))
	for _, p := range products {
		items = append(items, dto.SearchResultItem{
			ID:          p.ID,
			Title:       fmt.Sprintf("%s (%s)", p.ProName, p.ProCode),
			Description: "产品管理",
			Module:      "product",
			ModuleName:  "产品管理",
			Path:        "/product/index",
		})
	}
	return items, nil
}

func (s *SearchService) searchCustomers(like string) ([]dto.SearchResultItem, error) {
	var customers []entity.Customer
	if err := dao.GetDB().Where("cus_code LIKE ? OR cus_name LIKE ? OR contact LIKE ? OR phone LIKE ?",
		like, like, like, like).
		Select("id, cus_code, cus_name, contact").
		Limit(20).Find(&customers).Error; err != nil {
		return nil, err
	}
	items := make([]dto.SearchResultItem, 0, len(customers))
	for _, c := range customers {
		desc := c.CusCode
		if c.Contact != "" {
			desc += " / " + c.Contact
		}
		items = append(items, dto.SearchResultItem{
			ID:          c.ID,
			Title:       c.CusName,
			Description: desc,
			Module:      "customer",
			ModuleName:  "客户信息管理",
			Path:        "/system/customer",
		})
	}
	return items, nil
}

func (s *SearchService) searchProductionOrders(like string) ([]dto.SearchResultItem, error) {
	var orders []struct {
		ID          int64
		OrderCode   string
		ProductName string
	}
	sql := `SELECT DISTINCT po.id, po.order_code, po.product_name
		FROM production_order po
		LEFT JOIN mold m ON po.mold_id = m.id
		LEFT JOIN production_order_item poi ON poi.order_id = po.id
		LEFT JOIN mold mi ON mi.id = poi.mold_id
		WHERE po.order_code LIKE ? OR po.product_name LIKE ?
		   OR m.mold_code LIKE ? OR m.mold_name LIKE ?
		   OR mi.mold_code LIKE ? OR mi.mold_name LIKE ?
		LIMIT 20`
	if err := dao.GetDB().Raw(sql, like, like, like, like, like, like).Scan(&orders).Error; err != nil {
		return nil, err
	}
	items := make([]dto.SearchResultItem, 0, len(orders))
	for _, o := range orders {
		title := o.OrderCode
		if o.ProductName != "" {
			title += " - " + o.ProductName
		}
		items = append(items, dto.SearchResultItem{
			ID:          o.ID,
			Title:       title,
			Description: "生产订单",
			Module:      "production_order",
			ModuleName:  "生产管理",
			Path:        fmt.Sprintf("/production/order/detail/%d", o.ID),
		})
	}
	return items, nil
}

func (s *SearchService) searchMoldOrders(like string) ([]dto.SearchResultItem, error) {
	var orders []struct {
		ID        int64
		OrderCode string
	}
	sql := `SELECT DISTINCT mo.id, mo.order_code
		FROM mold_order mo
		LEFT JOIN mold m ON mo.mold_id = m.id
		LEFT JOIN mold_order_item moi ON moi.order_id = mo.id
		LEFT JOIN mold mi ON moi.mold_id = mi.id
		WHERE mo.order_code LIKE ?
		   OR m.mold_code LIKE ? OR m.mold_name LIKE ?
		   OR mi.mold_code LIKE ? OR mi.mold_name LIKE ?
		LIMIT 20`
	if err := dao.GetDB().Raw(sql, like, like, like, like, like).Scan(&orders).Error; err != nil {
		return nil, err
	}
	items := make([]dto.SearchResultItem, 0, len(orders))
	for _, o := range orders {
		items = append(items, dto.SearchResultItem{
			ID:          o.ID,
			Title:       o.OrderCode,
			Description: "模具订单",
			Module:      "mold_order",
			ModuleName:  "模具管理",
			Path:        fmt.Sprintf("/mold/order/detail/%d", o.ID),
		})
	}
	return items, nil
}

func (s *SearchService) searchMoldWorkOrders(like string) ([]dto.SearchResultItem, error) {
	var workOrders []struct {
		ID       int64
		WorkCode string
	}
	sql := `SELECT DISTINCT mwo.id, mwo.work_code
		FROM mold_work_order mwo
		LEFT JOIN mold m ON mwo.mold_id = m.id
		WHERE mwo.work_code LIKE ? OR m.mold_code LIKE ? OR m.mold_name LIKE ?
		LIMIT 20`
	if err := dao.GetDB().Raw(sql, like, like, like).Scan(&workOrders).Error; err != nil {
		return nil, err
	}
	items := make([]dto.SearchResultItem, 0, len(workOrders))
	for _, w := range workOrders {
		items = append(items, dto.SearchResultItem{
			ID:          w.ID,
			Title:       w.WorkCode,
			Description: "模具工单",
			Module:      "mold_work_order",
			ModuleName:  "模具管理",
			Path:        "/mold/work-order",
		})
	}
	return items, nil
}

func (s *SearchService) searchMasterOrders(like string) ([]dto.SearchResultItem, error) {
	var orders []entity.MasterOrder
	if err := dao.GetDB().Where("order_code LIKE ?", like).
		Select("id, order_code").
		Limit(20).Find(&orders).Error; err != nil {
		return nil, err
	}
	items := make([]dto.SearchResultItem, 0, len(orders))
	for _, o := range orders {
		items = append(items, dto.SearchResultItem{
			ID:          o.ID,
			Title:       o.OrderCode,
			Description: "综合订单",
			Module:      "master_order",
			ModuleName:  "订单管理",
			Path:        fmt.Sprintf("/order-management/detail/%d", o.ID),
		})
	}
	return items, nil
}
