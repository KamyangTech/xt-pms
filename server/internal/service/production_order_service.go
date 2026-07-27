package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
)

type ProductionOrderService struct {
	orderDAO                   *dao.ProductionOrderDAO
	orderItemDAO               *dao.ProductionOrderItemDAO
	scheduleDAO                *dao.ProductionScheduleDAO
	reportDAO                  *dao.ProductionReportDAO
	outsourcingDAO             *dao.OutsourcingRecordDAO
	outsourcingReportDAO       *dao.OutsourcingReportDAO
	stockInDAO                 *dao.ProductionStockInDAO
	stockOutDAO                *dao.ProductionStockOutDAO
	stockService               *StockService
	materialDAO                *dao.MaterialDAO
	moldDAO                    *dao.MoldDAO
	fileDAO                    *dao.FileDAO // 文件DAO，用于维护引用计数和查询模具主图
	usageService               *MoldUsageService
	flowService                *ProductionStatusFlowService // 状态流转服务
	externalMoldArchiveService *ExternalMoldArchiveService  // 外来模具档案服务
	moldProductRelationService *MoldProductRelationService  // 模具产品关联服务
}

func NewProductionOrderService() *ProductionOrderService {
	return &ProductionOrderService{
		orderDAO:                   dao.NewProductionOrderDAO(),
		orderItemDAO:               dao.NewProductionOrderItemDAO(),
		scheduleDAO:                dao.NewProductionScheduleDAO(),
		reportDAO:                  dao.NewProductionReportDAO(),
		outsourcingDAO:             dao.NewOutsourcingRecordDAO(),
		outsourcingReportDAO:       dao.NewOutsourcingReportDAO(),
		stockInDAO:                 dao.NewProductionStockInDAO(),
		stockOutDAO:                dao.NewProductionStockOutDAO(),
		stockService:               NewStockService(),
		materialDAO:                dao.NewMaterialDAO(),
		moldDAO:                    dao.NewMoldDAO(),
		fileDAO:                    dao.NewFileDAO(),
		usageService:               NewMoldUsageService(),
		flowService:                NewProductionStatusFlowService(),
		externalMoldArchiveService: NewExternalMoldArchiveService(),
		moldProductRelationService: NewMoldProductRelationService(),
	}
}

// GenerateStockInCode 生成入库单号 SR-YYYYMMDD-001
func (s *ProductionOrderService) GenerateStockInCode() (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")
	prefix := fmt.Sprintf("SR-%s-", dateStr)

	// 查询当天最大序号
	var maxCode string
	dao.GetDB().Model(&entity.ProductionStockIn{}).
		Where("stock_code LIKE ?", prefix+"%").
		Order("stock_code DESC").
		Limit(1).
		Pluck("stock_code", &maxCode)

	seq := 1
	if maxCode != "" {
		seqStr := strings.TrimPrefix(maxCode, prefix)
		if num, err := strconv.Atoi(seqStr); err == nil {
			seq = num + 1
		}
	}

	return fmt.Sprintf("%s%03d", prefix, seq), nil
}

// GenerateBatchNo 生成入库批次号 SR260427-001-01
// 格式：SR + YYMMDD(6位) + 模具ID(或外来模具编号) + 批次流水号(2位)
// 自有模具示例：SR260427-001-01 （模具ID=13）
// 外来模具示例：SR260427-ABC-01 （外来模具编号=ABC）
func (s *ProductionOrderService) GenerateBatchNo(moldID *int64, externalMoldCode *string, productID *int64) string {
	now := time.Now()
	yyMMdd := now.Format("060102") // YYMMDD格式

	var baseIdentifier string
	if moldID != nil && *moldID > 0 {
		baseIdentifier = fmt.Sprintf("%03d", *moldID)
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		code := *externalMoldCode
		if len(code) > 3 {
			code = code[:3]
		}
		baseIdentifier = code
	} else {
		baseIdentifier = "000"
	}

	// 查询当天该模具的最大批次流水号
	var maxBatchNo string
	query := dao.GetDB().Model(&entity.ProductionStockIn{}).
		Where("batch_no LIKE ?", fmt.Sprintf("SR%s-%s-%%", yyMMdd, baseIdentifier)).
		Order("batch_no DESC").
		Limit(1)

	// 如果传入了productID，增加product_id过滤条件
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	} else {
		query = query.Where("product_id IS NULL")
	}

	query.Pluck("batch_no", &maxBatchNo)

	seq := 1
	if maxBatchNo != "" {
		parts := strings.Split(maxBatchNo, "-")
		if len(parts) == 3 {
			if num, err := strconv.Atoi(parts[2]); err == nil {
				seq = num + 1
			}
		}
	}

	return fmt.Sprintf("SR%s-%s-%02d", yyMMdd, baseIdentifier, seq)
}

// GenerateStockOutBatchNo 生成出库批次号 SC260427-001-01
// 格式：SC + YYMMDD(6位) + 模具ID(或外来模具编号) + 批次流水号(2位)
// 自有模具示例：SC260427-001-01 （模具ID=13）
// 外来模具示例：SC260427-ABC-01 （外来模具编号=ABC）
func (s *ProductionOrderService) GenerateStockOutBatchNo(moldID *int64, externalMoldCode *string, productID *int64) string {
	now := time.Now()
	yyMMdd := now.Format("060102") // YYMMDD格式

	var baseIdentifier string
	if moldID != nil && *moldID > 0 {
		baseIdentifier = fmt.Sprintf("%03d", *moldID)
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		code := *externalMoldCode
		if len(code) > 3 {
			code = code[:3]
		}
		baseIdentifier = code
	} else {
		baseIdentifier = "000"
	}

	// 查询当天该模具的最大批次流水号
	var maxBatchNo string
	query := dao.GetDB().Model(&entity.ProductionStockOut{}).
		Where("batch_no LIKE ?", fmt.Sprintf("SC%s-%s-%%", yyMMdd, baseIdentifier)).
		Order("batch_no DESC").
		Limit(1)

	// 如果传入了productID，增加product_id过滤条件
	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	} else {
		query = query.Where("product_id IS NULL")
	}

	query.Pluck("batch_no", &maxBatchNo)

	seq := 1
	if maxBatchNo != "" {
		parts := strings.Split(maxBatchNo, "-")
		if len(parts) == 3 {
			if num, err := strconv.Atoi(parts[2]); err == nil {
				seq = num + 1
			}
		}
	}

	return fmt.Sprintf("SC%s-%s-%02d", yyMMdd, baseIdentifier, seq)
}

func getDependencyStatus(moldSourceType string, moldIDs []int64, externalMoldCode *string) string {
	// 外来模具：不需要验收，直接可以生产
	if moldSourceType == "CUSTOMER_PROVIDED" {
		return "READY"
	}

	// 自有模具：需要判断模具状态
	if moldSourceType == "SELF_MADE" && len(moldIDs) > 0 {
		// 检查每个模具的状态,如果有任何一个模具未达到可生产状态,则返回 WAITING_MOLD
		for _, moldID := range moldIDs {
			var mold entity.Mold
			if err := dao.DB.First(&mold, moldID).Error; err == nil {
				// 如果模具状态不是可生产状态（如:已试模、库存中、在用等），则需要等待
				// 可生产的模具状态包括:TESTED(已试模), IN_STOCK(库存中), DELIVERED(已交付)
				allowedStatuses := []string{
					constant.MoldStatusTested,    // 已试模
					constant.MoldStatusInStock,        // 库存中
					constant.MoldStatusDelivered,      // 已交付(客户验收)
				}

				isReady := false
				for _, allowedStatus := range allowedStatuses {
					if mold.LifeCycleStatus == allowedStatus {
						isReady = true
						break
					}
				}

				// 如果有任何一个模具未达到可生产状态,则返回 WAITING_MOLD
				if !isReady {
					return "WAITING_MOLD"
				}
			} else {
				// 模具不存在,需要等待
				return "WAITING_MOLD"
			}
		}

		// 所有模具都达到可生产状态
		return "READY"
	}

	// 其他情况：无依赖
	return "NONE"
}

// computeDependencyStatus 根据模具来源类型和预加载的模具 map 计算依赖状态（避免 DB 查询）
func (s *ProductionOrderService) computeDependencyStatus(moldSourceType string, moldIDs []int64, externalMoldCode *string, moldMap map[int64]*entity.Mold) string {
	if moldSourceType == "CUSTOMER_PROVIDED" {
		return "READY"
	}
	if moldSourceType == "SELF_MADE" && len(moldIDs) > 0 {
		allowedStatuses := []string{
			constant.MoldStatusTested,
			constant.MoldStatusInStock,
			constant.MoldStatusDelivered,
		}
		for _, moldID := range moldIDs {
			mold, ok := moldMap[moldID]
			if !ok {
				return "WAITING_MOLD"
			}
			isReady := false
			for _, allowedStatus := range allowedStatuses {
				if mold.LifeCycleStatus == allowedStatus {
					isReady = true
					break
				}
			}
			if !isReady {
				return "WAITING_MOLD"
			}
		}
		return "READY"
	}
	return "NONE"
}

// GenerateOrderCode 生成订单编号 PO-YYMMDD-01
func (s *ProductionOrderService) GenerateOrderCode() (string, error) {
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD 6位日期
	prefix := fmt.Sprintf("PO-%s-", dateStr)

	// 查询当天最大序号
	var maxCode string
	dao.GetDB().Model(&entity.ProductionOrder{}).
		Where("order_code LIKE ?", prefix+"%").
		Order("order_code DESC").
		Limit(1).
		Pluck("order_code", &maxCode)

	seq := 1
	if maxCode != "" {
		seqStr := strings.TrimPrefix(maxCode, prefix)
		if num, err := strconv.Atoi(seqStr); err == nil {
			seq = num + 1
		}
	}

	return fmt.Sprintf("%s%02d", prefix, seq), nil
}

// CreateProductionOrder 创建生产订单
func (s *ProductionOrderService) CreateProductionOrder(orderDTO dto.ProductionOrderDTO, userID int64) (*entity.ProductionOrder, error) {
	if orderDTO.OrderCode != "" {
		existing, _ := s.orderDAO.FindByCode(orderDTO.OrderCode)
		if existing != nil {
			return nil, errors.New("订单编号已存在")
		}
	}

	orderCode := orderDTO.OrderCode
	if orderCode == "" {
		var err error
		orderCode, err = s.GenerateOrderCode()
		if err != nil {
			return nil, errors.New("生成订单编号失败")
		}
	}

	prodType := ""
	moldSourceType := ""
	if orderDTO.ProductionType != nil {
		prodType = *orderDTO.ProductionType
	}
	if orderDTO.MoldSourceType != nil {
		moldSourceType = *orderDTO.MoldSourceType
	}
	if prodType == "" || moldSourceType == "" {
		return nil, errors.New("请选择生产类型和模具来源")
	}
	// 自有模具验证：检查明细项中是否有模具选择
	if prodType == "IN_HOUSE" && moldSourceType == "SELF_MADE" {
		if len(orderDTO.Items) == 0 {
			return nil, errors.New("请添加订单明细")
		}
		for i, item := range orderDTO.Items {
			if item.MoldID == nil || *item.MoldID == 0 {
				return nil, fmt.Errorf("第%d项明细必须选择模具", i+1)
			}
		}
	}
	if prodType == "OUTSOURCED" && orderDTO.ProcessorID == nil {
		return nil, errors.New("外发生产必须选择加工商")
	}

	// 【优化】先收集所有模具ID，批量查询一次，避免 N+1 循环查询
	var moldMap map[int64]*entity.Mold
	if moldSourceType == "SELF_MADE" {
		moldIDSet := make(map[int64]struct{})
		for _, item := range orderDTO.Items {
			if item.MoldID != nil && *item.MoldID > 0 {
				moldIDSet[*item.MoldID] = struct{}{}
			}
			for _, mid := range item.MoldIDs {
				if mid > 0 {
					moldIDSet[mid] = struct{}{}
				}
			}
		}
		if len(moldIDSet) > 0 {
			allMoldIDs := make([]int64, 0, len(moldIDSet))
			for mid := range moldIDSet {
				allMoldIDs = append(allMoldIDs, mid)
			}
			var molds []entity.Mold
			if err := dao.GetDB().Where("id IN ?", allMoldIDs).Find(&molds).Error; err != nil {
				return nil, errors.New("查询模具信息失败")
			}
			moldMap = make(map[int64]*entity.Mold, len(molds))
			for i := range molds {
				moldMap[molds[i].ID] = &molds[i]
			}
		}

		for _, item := range orderDTO.Items {
			if item.MoldID != nil && *item.MoldID > 0 {
				mold, ok := moldMap[*item.MoldID]
				if !ok {
					return nil, fmt.Errorf("模具ID %d 不存在", *item.MoldID)
				}
				if mold.CurrentLocation == "OUTSOURCING" {
					return nil, errors.New("模具正在外发中，不能使用")
				}
				if mold.LifeCycleStatus == constant.MoldStatusRepairing {
					return nil, errors.New("模具正在维修中，不能使用")
				}
			}
		}
	}

	var planStartDate, planEndDate *time.Time
	if orderDTO.PlanStartDate != nil && *orderDTO.PlanStartDate != "" {
		t, err := time.ParseInLocation("2006-01-02", *orderDTO.PlanStartDate, time.Local)
		if err == nil {
			planStartDate = &t
		}
	}
	if orderDTO.PlanEndDate != nil && *orderDTO.PlanEndDate != "" {
		t, err := time.ParseInLocation("2006-01-02", *orderDTO.PlanEndDate, time.Local)
		if err == nil {
			planEndDate = &t
		}
	}

	totalPlanQty := 0
	for _, itemDTO := range orderDTO.Items {
		totalPlanQty += itemDTO.PlanQty
	}

	order := &entity.ProductionOrder{
		OrderCode:            orderCode,
		CustomerID:           orderDTO.CustomerID,
		ProductName:          orderDTO.ProductName,
		ProductionType:       prodType,
		MoldSourceType:       moldSourceType,
		MoldID:               orderDTO.MoldID,
		ExternalMoldCode:     orderDTO.ExternalMoldCode,
		ExternalMoldInfo:     orderDTO.ExternalMoldInfo,
		PlanQty:              totalPlanQty, // 从明细项汇总
		PlanStartDate:        planStartDate,
		PlanEndDate:          planEndDate,
		DeviceID:             orderDTO.DeviceID,
		WorkerID:             orderDTO.WorkerID,
		ProcessorID:          orderDTO.ProcessorID,
		OutsourcePrice:       orderDTO.OutsourcePrice,
		MaterialCost:         orderDTO.MaterialCost,
		ProcessingCost:       orderDTO.ProcessingCost,
		UnitPrice:            orderDTO.UnitPrice,
		Priority:             orderDTO.Priority,
		Remark:               orderDTO.Remark,
		CreateUserID:         &userID,
		Status:               constant.ProductionOrderStatusPendingSchedule, // 待排产
		CompletedQty:         0,
		OkQty:                0,
		BadQty:               0,
		MoldShots:            0,
		TotalAmount:          float64(totalPlanQty) * orderDTO.UnitPrice,
		OutsourceTotalAmount: float64(totalPlanQty) * orderDTO.OutsourcePrice,
	}

	// 如果是外发生产，自动计算外发总金额
	if order.ProductionType == "OUTSOURCED" {
		order.OutsourceTotalAmount = float64(order.PlanQty) * order.OutsourcePrice
	}

	// 使用事务确保数据一致性
	tx := dao.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单失败: " + err.Error())
	}

	// 【优化】批量查询模具主图（明细循环前一次性加载，避免 N+1）
	allMoldIDsForImage := make([]int64, 0)
	moldIDSeen := make(map[int64]struct{})
	for _, itemDTO := range orderDTO.Items {
		var mid int64
		if itemDTO.MoldID != nil {
			mid = *itemDTO.MoldID
		} else if len(itemDTO.MoldIDs) > 0 {
			mid = itemDTO.MoldIDs[0]
		}
		if mid > 0 {
			if _, ok := moldIDSeen[mid]; !ok {
				moldIDSeen[mid] = struct{}{}
				allMoldIDsForImage = append(allMoldIDsForImage, mid)
			}
		}
	}

	// 批量查询模具主图
	primaryImageMap := make(map[int64]*entity.File)
	if len(allMoldIDsForImage) > 0 {
		var primaryImages []entity.File
		if err := dao.GetDB().Where("business_type = 'mold' AND business_id IN ? AND is_primary = 1", allMoldIDsForImage).Find(&primaryImages).Error; err == nil {
			for i := range primaryImages {
				if primaryImages[i].BusinessID != nil {
					primaryImageMap[*primaryImages[i].BusinessID] = &primaryImages[i]
				}
			}
		}
	}

	for _, itemDTO := range orderDTO.Items {
		// 处理产品图片：如果没有指定图片，尝试从模具主图继承
		var productImageURL string
		if itemDTO.ProductImageURL != nil && *itemDTO.ProductImageURL != "" {
			productImageURL = *itemDTO.ProductImageURL
		} else if moldSourceType == "SELF_MADE" && itemDTO.MoldID != nil {
			// 从预加载的模具主图 map 中查找
			if primaryImage, ok := primaryImageMap[*itemDTO.MoldID]; ok {
				productImageURL = primaryImage.URL
			}
		}

		// 确定模具IDs
		var moldIDs []int64
		if len(itemDTO.MoldIDs) > 0 {
			// 多模具模式
			moldIDs = itemDTO.MoldIDs
		} else if itemDTO.MoldID != nil {
			// 单模具模式（兼容旧数据）
			moldIDs = []int64{*itemDTO.MoldID}
		}

		item := &entity.ProductionOrderItem{
			OrderID:           order.ID,
			ProductID:         itemDTO.ProductID,    // 优先使用前端传来的 product_id
			ProductName:       *itemDTO.ProductName, // 必填
			ProductImageURL:   productImageURL,
			MoldID:            itemDTO.MoldID, // 保留第一个模具ID用于兼容
			MoldIDs:           moldIDs,
			ExternalMoldCode:  itemDTO.ExternalMoldCode,
			DeviceID:          itemDTO.DeviceID,
			PlasticMaterialID: itemDTO.PlasticMaterialID,
			PlanQty:           itemDTO.PlanQty,
			Sort:              itemDTO.Sort,
			Remark:            itemDTO.Remark,
			// 设置依赖状态（使用预加载的 moldMap，避免 DB 查询）
			DependencyStatus: s.computeDependencyStatus(moldSourceType, moldIDs, itemDTO.ExternalMoldCode, moldMap),
			// ✨ 款项配置字段（金额计算）
			UnitPrice:      itemDTO.UnitPrice,
			PackagingFee:   itemDTO.PackagingFee,
			FreightFee:     itemDTO.FreightFee,
			OutsourcePrice: itemDTO.OutsourcePrice,
		}

		// ✨ 始终同步产品档案信息（创建或补全），确保图片等字段不丢失
		if item.ProductName != "" {
			syncedProductID := s.ensureProductRecord(tx, item.ProductName, item.MoldID, productImageURL)
			if item.ProductID == nil && syncedProductID != nil {
				item.ProductID = syncedProductID
			}
		}
		if err := tx.Create(item).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("创建订单明细失败: " + err.Error())
		}

		// 维护文件引用计数：订单明细引用了图片，对应文件的 ref_count + 1
		if productImageURL != "" {
			if err := s.fileDAO.IncrementRefCount(productImageURL, tx); err != nil {
				logger.Warn("图片引用计数增加失败（不影响主流程）",
					zap.String("url", productImageURL),
					zap.Error(err))
			}
		}

		// 如果是自有模具，创建模具产品关联记录
		if moldSourceType == "SELF_MADE" && len(moldIDs) > 0 {
			relations := make([]*entity.MoldProductRelation, 0)
			for _, moldID := range moldIDs {
				relation := &entity.MoldProductRelation{
					MoldID:                moldID,
					ProductID:             itemDTO.ProductID, // 可选
					ProductionOrderItemID: &item.ID,
					ProductName:           *itemDTO.ProductName, // 必填：产品名称快照
					ProductImageURL:       productImageURL,      // 必填：产品图片URL
					PlanQty:               itemDTO.PlanQty,      // 必填：计划数量
					RelationType:          "PRODUCTION",
				}
				relations = append(relations, relation)
			}

			if err := s.moldProductRelationService.CreateRelationsBatch(tx, relations); err != nil {
				tx.Rollback()
				return nil, errors.New("创建模具产品关联失败: " + err.Error())
			}
		}

		// 如果是外来模具，自动创建或更新档案
		if moldSourceType == "CUSTOMER_PROVIDED" && itemDTO.ExternalMoldCode != nil && *itemDTO.ExternalMoldCode != "" && itemDTO.ProductName != nil && *itemDTO.ProductName != "" {
			_, err := s.externalMoldArchiveService.CreateOrUpdateArchive(
				*itemDTO.ExternalMoldCode,
				*itemDTO.ProductName,
				&order.CustomerID,
				nil,
				order.ID,
			)
			if err != nil {
				// 档案验证失败，回滚事务
				tx.Rollback()
				return nil, fmt.Errorf("外来模具档案验证失败: %v", err)
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交事务失败: " + err.Error())
	}

	// ✨ 自动生成应收账款（在主事务提交后执行）
	if order.TotalAmount > 0 {
		receivableService := NewAccountsReceivableService()
		existing, _ := receivableService.GetReceivableByOrder(order.ID, "production_order")
		if existing == nil {
			dueDate := &dto.Date{Time: time.Now().AddDate(0, 0, 30)}
			ar := &entity.AccountsReceivable{
				CustomerID:  order.CustomerID,
				OrderID:     &order.ID,
				OrderType:   "production_order",
				PaymentType: "total",
				Amount:      order.TotalAmount,
				DueDate:     dueDate,
				Status:      "PENDING",
				Remark:      fmt.Sprintf("生产订单 %s 自动生成应收账款", order.OrderCode),
			}
			if err := receivableService.CreateReceivable(ar); err != nil {
				logger.Error("创建应收账款失败", zap.Error(err))
			} else {
				logger.Info("自动创建应收账款成功",
					zap.Int64("order_id", order.ID),
					zap.String("order_code", order.OrderCode),
					zap.Float64("amount", order.TotalAmount))
			}
		}
	}

	// 如果是外发生产，自动创建外发记录
	if order.ProductionType == "OUTSOURCED" && order.ProcessorID != nil {
		outsourcingService := NewOutsourcingService()
		recordCode, err := outsourcingService.GenerateRecordCode()
		if err == nil {
			outsourcingRecord := &entity.OutsourcingRecord{
				RecordCode:         recordCode,
				OrderID:            &order.ID, // ✨ 生产外发需要关联生产订单
				OutsourceType:      "PRODUCTION",
				ProcessorID:        *order.ProcessorID,
				Quantity:           order.PlanQty,
				UnitPrice:          order.OutsourcePrice,
				TotalAmount:        order.OutsourceTotalAmount,
				OutsourceDate:      planStartDate,
				ExpectedReturnDate: planEndDate,
				Status:             "PENDING",
				CreateUserID:       &userID,
			}
			if err := outsourcingService.outsourcingDAO.Create(outsourcingRecord); err == nil {
				logger.Info("外发生产订单创建成功，已自动生成外发记录",
					zap.Int64("order_id", order.ID),
					zap.String("record_code", recordCode))
			} else {
				logger.Error("创建外发记录失败",
					zap.Int64("order_id", order.ID),
					zap.Error(err))
			}
		}
	}

	return order, nil
}

// GetOrderByID 获取订单详情
func (s *ProductionOrderService) GetOrderByID(id int64) (*entity.ProductionOrder, error) {
	return s.orderDAO.FindByID(id)
}

// GetProductionItemsByMoldID 根据模具ID获取使用该模具的所有生产订单明细
func (s *ProductionOrderService) GetProductionItemsByMoldID(moldID int64) ([]entity.ProductionOrderItem, error) {
	return s.orderItemDAO.FindByMoldID(moldID)
}

// UpdateProductionOrder 更新生产订单
func (s *ProductionOrderService) UpdateProductionOrder(orderDTO dto.ProductionOrderDTO, userID int64) error {
	order, err := s.orderDAO.FindByID(orderDTO.ID)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	// 订单已开始生产后，限制只能修改部分字段
	// 可编辑字段：优先级、产品名称、计划结束日期、成本信息、备注
	// 不可编辑字段：客户、生产类型、模具来源、模具信息、计划开始日期、负责人、外发信息
	isProductionStarted := order.Status != constant.ProductionOrderStatusPendingSchedule

	if isProductionStarted {
		// 订单已开始，只允许更新部分字段
		// 使用 map 来明确指定要更新的字段，避免 GORM 忽略零值
		updateData := make(map[string]interface{})

		if orderDTO.Priority != "" {
			updateData["priority"] = orderDTO.Priority
		}

		if orderDTO.ProductName != "" {
			updateData["product_name"] = orderDTO.ProductName
		}

		if orderDTO.PlanEndDate != nil && *orderDTO.PlanEndDate != "" {
			if t, err := time.ParseInLocation("2006-01-02", *orderDTO.PlanEndDate, time.Local); err == nil {
				updateData["plan_end_date"] = t
			}
		}

		updateData["material_cost"] = orderDTO.MaterialCost
		updateData["processing_cost"] = orderDTO.ProcessingCost
		updateData["unit_price"] = orderDTO.UnitPrice
		updateData["total_amount"] = float64(order.PlanQty) * orderDTO.UnitPrice

		updateData["remark"] = orderDTO.Remark

		// 使用 Updates + map 确保所有字段都被更新
		if err := dao.GetDB().Model(&entity.ProductionOrder{}).Where("id = ?", order.ID).Updates(updateData).Error; err != nil {
			return err
		}

		// 订单已开始生产后，也允许更新订单明细
		if len(orderDTO.Items) > 0 {
			itemDAO := dao.NewProductionOrderItemDAO()

			// 收集旧明细的URL用于引用计数维护
			oldItems, err := itemDAO.FindByOrderID(order.ID)
			if err != nil {
				return err
			}

			newURLs := make(map[string]bool)
			for _, itemDTO := range orderDTO.Items {
				if itemDTO.ProductImageURL != nil && *itemDTO.ProductImageURL != "" {
					newURLs[*itemDTO.ProductImageURL] = true
				}
			}

			oldURLs := make(map[string]bool)
			for _, oldItem := range oldItems {
				if oldItem.ProductImageURL != "" {
					oldURLs[oldItem.ProductImageURL] = true
				}
			}

			// 2. 删除旧的明细数据库记录
			if err := itemDAO.DeleteByOrderID(order.ID); err != nil {
				return errors.New("删除旧明细失败: " + err.Error())
			}

			for _, itemDTO := range orderDTO.Items {
				var productImageURL string
				if itemDTO.ProductImageURL != nil && *itemDTO.ProductImageURL != "" {
					productImageURL = *itemDTO.ProductImageURL
				}

			item := &entity.ProductionOrderItem{
				OrderID:           order.ID,
				ProductID:         itemDTO.ProductID,
				ProductName:       *itemDTO.ProductName,
				ProductImageURL:   productImageURL,
				MoldID:            itemDTO.MoldID,
				ExternalMoldCode:  itemDTO.ExternalMoldCode,
				DeviceID:          itemDTO.DeviceID,
				PlasticMaterialID: itemDTO.PlasticMaterialID,
				PlanQty:           itemDTO.PlanQty,
				Sort:              itemDTO.Sort,
				Remark:            itemDTO.Remark,
				DependencyStatus:  getDependencyStatus(order.MoldSourceType, nil, itemDTO.ExternalMoldCode),
				// ✨ 款项配置字段（金额计算）
				UnitPrice:      itemDTO.UnitPrice,
				PackagingFee:   itemDTO.PackagingFee,
				FreightFee:     itemDTO.FreightFee,
				OutsourcePrice: itemDTO.OutsourcePrice,
			}
			// ✨ 始终同步产品档案信息（创建或补全），确保图片等字段不丢失
			if item.ProductName != "" {
				syncedProductID := s.ensureProductRecord(nil, item.ProductName, item.MoldID, productImageURL)
				if item.ProductID == nil && syncedProductID != nil {
					item.ProductID = syncedProductID
				}
			}
			if err := itemDAO.Create(item); err != nil {
				return errors.New("创建新明细失败: " + err.Error())
			}
			}

			// 维护文件引用计数：旧有但新无 → 减；新有但旧无 → 加
			for url := range oldURLs {
				if !newURLs[url] {
					if err := s.fileDAO.DecrementRefCount(url, dao.GetDB()); err != nil {
						logger.Warn("图片引用计数减少失败（不影响主流程）",
							zap.String("url", url), zap.Error(err))
					}
				}
			}
			for url := range newURLs {
				if !oldURLs[url] {
					if err := s.fileDAO.IncrementRefCount(url, dao.GetDB()); err != nil {
						logger.Warn("图片引用计数增加失败（不影响主流程）",
							zap.String("url", url), zap.Error(err))
					}
				}
			}
		}

		return nil
	}

	// 订单未开始，允许更新所有字段
	order.ProductName = orderDTO.ProductName
	if orderDTO.ProductionType != nil {
		order.ProductionType = *orderDTO.ProductionType
	}
	if orderDTO.MoldSourceType != nil {
		order.MoldSourceType = *orderDTO.MoldSourceType
	}
	order.MoldID = orderDTO.MoldID
	order.ExternalMoldInfo = orderDTO.ExternalMoldInfo

	totalPlanQty := 0
	for _, item := range orderDTO.Items {
		totalPlanQty += item.PlanQty
	}
	order.PlanQty = totalPlanQty

	if orderDTO.PlanStartDate != nil && *orderDTO.PlanStartDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", *orderDTO.PlanStartDate, time.Local); err == nil {
			order.PlanStartDate = &t
		}
	} else {
		order.PlanStartDate = nil
	}
	if orderDTO.PlanEndDate != nil && *orderDTO.PlanEndDate != "" {
		if t, err := time.ParseInLocation("2006-01-02", *orderDTO.PlanEndDate, time.Local); err == nil {
			order.PlanEndDate = &t
		}
	} else {
		order.PlanEndDate = nil
	}

	order.DeviceID = orderDTO.DeviceID
	order.WorkerID = orderDTO.WorkerID
	order.ProcessorID = orderDTO.ProcessorID
	order.OutsourcePrice = orderDTO.OutsourcePrice
	order.MaterialCost = orderDTO.MaterialCost
	order.ProcessingCost = orderDTO.ProcessingCost
	order.UnitPrice = orderDTO.UnitPrice
	order.Priority = orderDTO.Priority
	order.Remark = orderDTO.Remark
	order.TotalAmount = float64(order.PlanQty) * orderDTO.UnitPrice
	order.OutsourceTotalAmount = float64(order.PlanQty) * orderDTO.OutsourcePrice

	itemDAO := dao.NewProductionOrderItemDAO()

	// 查询旧明细，收集新旧URL集合用于引用计数维护
	oldItems, err := itemDAO.FindByOrderID(order.ID)
	if err != nil {
		return err
	}

	newURLs := make(map[string]bool)
	for _, itemDTO := range orderDTO.Items {
		if itemDTO.ProductImageURL != nil && *itemDTO.ProductImageURL != "" {
			newURLs[*itemDTO.ProductImageURL] = true
		}
	}

	oldURLs := make(map[string]bool)
	for _, oldItem := range oldItems {
		if oldItem.ProductImageURL != "" {
			oldURLs[oldItem.ProductImageURL] = true
		}
	}

	// 2. 删除旧明细数据库记录
	if err := itemDAO.DeleteByOrderID(order.ID); err != nil {
		return errors.New("删除旧明细失败: " + err.Error())
	}

	for _, itemDTO := range orderDTO.Items {
		// 处理产品图片：如果没有指定图片，尝试从模具主图继承
		var productImageURL string
		if itemDTO.ProductImageURL != nil && *itemDTO.ProductImageURL != "" {
			productImageURL = *itemDTO.ProductImageURL
		} else if order.MoldSourceType == "SELF_MADE" && itemDTO.MoldID != nil {
			// 尝试获取模具的主图
			if primaryImage, err := s.fileDAO.FindPrimary("mold", *itemDTO.MoldID); err == nil && primaryImage != nil {
				productImageURL = primaryImage.URL
			}
		}

		item := &entity.ProductionOrderItem{
			OrderID:           order.ID,
			ProductID:         itemDTO.ProductID,
			ProductName:       *itemDTO.ProductName, // 必填
			ProductImageURL:   productImageURL,
			MoldID:            itemDTO.MoldID,
			ExternalMoldCode:  itemDTO.ExternalMoldCode,
			DeviceID:          itemDTO.DeviceID,
			PlasticMaterialID: itemDTO.PlasticMaterialID,
			PlanQty:           itemDTO.PlanQty,
			Sort:              itemDTO.Sort,
			Remark:            itemDTO.Remark,
			// ✨ 设置依赖状态：根据模具来源类型决定
			DependencyStatus: getDependencyStatus(order.MoldSourceType, nil, itemDTO.ExternalMoldCode),
			// ✨ 款项配置字段（金额计算）
			UnitPrice:      itemDTO.UnitPrice,
			PackagingFee:   itemDTO.PackagingFee,
			FreightFee:     itemDTO.FreightFee,
			OutsourcePrice: itemDTO.OutsourcePrice,
		}
		// ✨ 始终同步产品档案信息（创建或补全），确保图片等字段不丢失
		if item.ProductName != "" {
			syncedProductID := s.ensureProductRecord(nil, item.ProductName, item.MoldID, productImageURL)
			if item.ProductID == nil && syncedProductID != nil {
				item.ProductID = syncedProductID
			}
		}
		if err := itemDAO.Create(item); err != nil {
			return errors.New("创建订单明细失败: " + err.Error())
		}
	}

	// 维护文件引用计数：旧有但新无 → 减；新有但旧无 → 加
	for url := range oldURLs {
		if !newURLs[url] {
			if err := s.fileDAO.DecrementRefCount(url, dao.GetDB()); err != nil {
				logger.Warn("图片引用计数减少失败（不影响主流程）",
					zap.String("url", url), zap.Error(err))
			}
		}
	}
	for url := range newURLs {
		if !oldURLs[url] {
			if err := s.fileDAO.IncrementRefCount(url, dao.GetDB()); err != nil {
				logger.Warn("图片引用计数增加失败（不影响主流程）",
					zap.String("url", url), zap.Error(err))
			}
		}
	}

	if err := s.orderDAO.Update(order); err != nil {
		return err
	}

	// ✨ 处理外发记录（允许生产过程中切换）
	outsourcingService := NewOutsourcingService()

	// 使用传入的当前登录用户ID

	if order.ProductionType == "OUTSOURCED" && order.ProcessorID != nil {
		// 切换到外发模式，创建或更新外发记录
		if err := outsourcingService.CreateOutsourcingRecordForProductionOrder(order, userID); err != nil {
			logger.Error("创建生产外发记录失败", zap.Error(err), zap.Int64("order_id", order.ID))
			// 不阻断流程，仅记录日志
		}
	} else {
		// 切换回本厂模式，撤回外发记录（如果存在）
		reason := "生产方式变更为：本厂生产"
		if err := outsourcingService.CancelOutsourcingRecordByProductionOrderID(order.ID, reason); err != nil {
			logger.Error("撤回生产外发记录失败", zap.Error(err), zap.Int64("order_id", order.ID))
			// 不阻断流程，仅记录日志
		}
	}

	return nil
}

// CancelOrder 取消订单
func (s *ProductionOrderService) CancelOrder(id int64, reason string) error {
	order, err := s.orderDAO.FindByID(id)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	// 已入库的订单不能取消
	if order.Status == constant.ProductionOrderStatusStockedIn ||
		order.Status == constant.ProductionOrderStatusShipped ||
		order.Status == constant.ProductionOrderStatusCompleted {
		return errors.New("订单已入库，不能取消")
	}

	order.Status = constant.ProductionOrderStatusCancelled
	order.Remark = reason
	if err := s.orderDAO.Update(order); err != nil {
		return err
	}

	// 如果该订单是外发生产，同步撤回对应的外发记录
	if order.ProductionType == "OUTSOURCED" {
		cancelReason := reason
		if cancelReason == "" {
			cancelReason = "生产订单已取消"
		}
		outsourcingService := NewOutsourcingService()
		if err := outsourcingService.CancelOutsourcingRecordByProductionOrderID(order.ID, cancelReason); err != nil {
			logger.Warn("撤回外发记录失败", zap.Error(err), zap.Int64("order_id", order.ID))
		}
	}

	s.syncMasterOrderStatusAfterUpdate(order)
	return nil
}

// StartProduction 开始生产
func (s *ProductionOrderService) StartProduction(id int64) error {
	order, err := s.orderDAO.FindByID(id)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	// 只有待排产的订单可以开始生产
	if order.Status != constant.ProductionOrderStatusPendingSchedule {
		return errors.New("订单状态不允许开始生产")
	}

	if order.MasterCancelled == 1 {
		return errors.New("父综合订单已取消，无法开始生产")
	}

	if order.DependencyStatus == "WAITING_MOLD" {
		return errors.New("模具尚未就绪，无法开始生产")
	}

	now := time.Now()
	order.Status = constant.ProductionOrderStatusProducing // 生产中
	order.ActualStartDate = &now

	if err := s.orderDAO.Update(order); err != nil {
		return err
	}
	s.syncMasterOrderStatusAfterUpdate(order)
	return nil
}

// CompleteProduction 完成生产
func (s *ProductionOrderService) CompleteProduction(id int64) error {
	order, err := s.orderDAO.FindByID(id)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	// 只有生产中的订单可以完成
	if order.Status != constant.ProductionOrderStatusProducing {
		return errors.New("订单不在生产中")
	}

	if order.MasterCancelled == 1 {
		return errors.New("父综合订单已取消，无法完成生产")
	}

	stockIns, err := s.stockInDAO.FindByOrderID(id)
	if err != nil || len(stockIns) == 0 {
		return errors.New("没有找到入库记录，请先完成报工并确保有良品入库")
	}

	now := time.Now()
	order.Status = constant.ProductionOrderStatusStockedIn // 已入库
	order.ActualEndDate = &now

	// 如果是自有模具，更新模具使用次数
	if order.MoldSourceType == "SELF_MADE" && order.MoldID != nil {
		shots := order.CompletedQty
		order.MoldShots = shots

		if err := s.usageService.UpdateMoldShots(*order.MoldID, shots); err != nil {
			// 记录错误但不影响主流程
			logger.Error("更新模具模次失败", zap.Error(err))
		}
	}

	if err := s.orderDAO.Update(order); err != nil {
		return err
	}
	s.syncMasterOrderStatusAfterUpdate(order)
	return nil
}

// StockInOrder 生产订单入库
func (s *ProductionOrderService) StockInOrder(dto dto.ProductionStockInDTO, operatorID int64) error {
	order, err := s.orderDAO.FindByID(dto.OrderID)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	// 【修改】现在支持明细项级别的独立状态管理
	var targetItemID int64
	var targetItem entity.ProductionOrderItem

	if dto.ItemID != nil && *dto.ItemID > 0 {
		// 新逻辑：检查明细项状态
		targetItemID = *dto.ItemID
		if err := dao.GetDB().First(&targetItem, targetItemID).Error; err != nil {
			return errors.New("订单明细不存在")
		}
		// 已入库或生产中的明细项都可以入库（支持继续入库/手动补录场景）
		if targetItem.Status != constant.ProductionOrderStatusStockedIn &&
			targetItem.Status != constant.ProductionOrderStatusProducing {
			return errors.New("该明细项状态不允许入库")
		}
		// 自动补充产品信息
		if dto.MoldID == nil && targetItem.MoldID != nil {
			dto.MoldID = targetItem.MoldID
		}
		if dto.ExternalMoldCode == nil && targetItem.ExternalMoldCode != nil {
			dto.ExternalMoldCode = targetItem.ExternalMoldCode
		}
		if dto.ProductID == nil && targetItem.ProductID != nil {
			dto.ProductID = targetItem.ProductID
		}
		if dto.ProductName == "" {
			dto.ProductName = targetItem.ProductName
		}
	} else {
		// 旧逻辑：检查订单状态（兼容旧数据）
		// 获取当前状态对应的字典值（语义化字符串）
		currentStatusValue, err := s.flowService.GetDictValue("production-order-status", order.Status)
		if err != nil {
			return errors.New("订单状态配置异常")
		}

		_, err = s.flowService.ValidateAction(currentStatusValue, "stock_in")
		if err != nil {
			return errors.New("订单状态不允许入库，" + err.Error())
		}
	}

	stockCode, err := s.GenerateStockInCode()
	if err != nil {
		return errors.New("生成入库单号失败: " + err.Error())
	}

	// 自动生成批次号（如果前端未提供）
	batchNo := dto.BatchNo
	if batchNo == "" {
		batchNo = s.GenerateBatchNo(dto.MoldID, dto.ExternalMoldCode, dto.ProductID)
	}

	now := time.Now()

	// 直接使用前端传入的 product_id 和 product_name
	productID := dto.ProductID
	productName := dto.ProductName

	// 如果 product_id 为空但 product_name 有值，尝试按名称匹配产品档案
	if productID == nil && productName != "" {
		if p, err := dao.NewProductDAO().FindByName(productName); err == nil {
			productID = &p.ID
		}
	}

	// 如果前端没有传产品名称，从订单明细中补充
	if productName == "" && productID != nil {
		for _, item := range order.Items {
			if item.ProductID != nil && *item.ProductID == *productID {
				productName = item.ProductName // 现在是 string 类型，直接赋值
				break
			}
		}
	}

	stockIn := &entity.ProductionStockIn{
		StockCode:        stockCode,
		OrderID:          &dto.OrderID,
		WarehouseID:      dto.WarehouseID,
		MoldID:           dto.MoldID,
		ExternalMoldCode: dto.ExternalMoldCode,
		ProductID:        productID,
		ProductName:      productName,
		Quantity:         dto.Quantity,
		BatchNo:          batchNo, // 使用自动生成的批次号
		StockInDate:      &now,
		OperatorID:       &operatorID,
		Remark:           dto.Remark,
	}

	if err := s.stockInDAO.Create(stockIn); err != nil {
		return errors.New("创建入库记录失败: " + err.Error())
	}

	// 【关键】更新实际库存表（stock 表）
	// 对于自有模具，入库记录保存的是 mold_id，但 stock 表关联的是 material_id
	// 对于外来模具，使用 external_mold_code 作为标识
	// 【新增】支持产品维度：通过 product_id 区分同一模具不同产品的库存
	if dto.MoldID != nil && *dto.MoldID > 0 {
		// 自有模具：找到或创建对应的物料记录（mat_type=3 表示产品）
		materialID, err := s.ensureMoldMaterial(*dto.MoldID, dto.ProductID)
		if err != nil {
			return errors.New("更新库存失败: " + err.Error())
		}

		// 传入 product_id 以支持按产品维度管理库存
		if err := s.stockService.InStockWithProduct(materialID, productID, dto.WarehouseID, float64(dto.Quantity)); err != nil {
			return errors.New("更新库存失败: " + err.Error())
		}
	} else if dto.ExternalMoldCode != nil && *dto.ExternalMoldCode != "" {
		// 外来模具：使用外来模具编号作为物料标识
		// 查找或创建外来模具对应的物料
		materialID, err := s.ensureExternalMoldMaterial(*dto.ExternalMoldCode, order.ProductName)
		if err != nil {
			return errors.New("更新库存失败: " + err.Error())
		}

		if err := s.stockService.InStockWithProduct(materialID, productID, dto.WarehouseID, float64(dto.Quantity)); err != nil {
			return errors.New("更新库存失败: " + err.Error())
		}
	}

	// 【修改】不再自动更新订单状态为"已入库"
	// 入库操作只创建入库记录和更新库存，不改变订单状态
	// 只有当用户手动完成订单时，才更新状态为"已完成"

	// ✨ 新增：同步更新明细项和订单的完成数量
	if dto.ItemID != nil && *dto.ItemID > 0 {
		targetItem.OkQty += dto.Quantity
		targetItem.CompletedQty += dto.Quantity
		if err := dao.GetDB().Save(&targetItem).Error; err != nil {
			return errors.New("更新明细项数量失败: " + err.Error())
		}

		// 同步汇总到订单级别（使用 SQL SUM 替代 Go 侧聚合）
		var orderSummary struct{ OkQty, CompletedQty int64 }
		if err := dao.GetDB().Model(&entity.ProductionOrderItem{}).
			Select("COALESCE(SUM(ok_qty),0) as ok_qty, COALESCE(SUM(completed_qty),0) as completed_qty").
			Where("order_id = ?", dto.OrderID).Scan(&orderSummary).Error; err == nil {
			order.OkQty = int(orderSummary.OkQty)
			order.CompletedQty = int(orderSummary.CompletedQty)
			dao.GetDB().Save(order)
		}
	} else {
		// 旧逻辑：按订单总体入库
		order.OkQty += dto.Quantity
		order.CompletedQty += dto.Quantity
		dao.GetDB().Save(order)
	}

	// ✨ 新增：如果是按明细项入库且当前是生产中状态，更新为已入库
	if dto.ItemID != nil && *dto.ItemID > 0 && targetItem.Status == constant.ProductionOrderStatusProducing {
		now := time.Now()
		targetItem.Status = constant.ProductionOrderStatusStockedIn // 已入库
		targetItem.ActualCompletionTime = &now

		if err := dao.GetDB().Save(&targetItem).Error; err != nil {
			return errors.New("更新明细项状态失败: " + err.Error())
		}
	}

	return nil
}

// checkAllMoldsStockedIn 检查订单的所有模具是否都已入库
func (s *ProductionOrderService) checkAllProductsStockedIn(orderID int64) bool {
	// 1. 获取订单的所有明细项
	itemDAO := dao.NewProductionOrderItemDAO()
	items, err := itemDAO.FindByOrderID(orderID)
	if err != nil || len(items) == 0 {
		// 如果没有明细项，认为已入库（兼容旧数据）
		return true
	}

	// 2. 获取该订单的所有入库记录
	stockIns, err := s.stockInDAO.FindByOrderID(orderID)
	if err != nil {
		return false
	}

	// 3. 检查每个模具是否都有入库记录
	for _, item := range items {
		// 自有模具：检查 mold_id
		if item.MoldID != nil && *item.MoldID > 0 {
			hasStockIn := false
			for _, si := range stockIns {
				if si.MoldID != nil && *si.MoldID == *item.MoldID {
					hasStockIn = true
					break
				}
			}
			if !hasStockIn {
				return false // 有模具还未入库
			}
		} else if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
			// 外来模具：检查 external_mold_code
			hasStockIn := false
			for _, si := range stockIns {
				if si.ExternalMoldCode != nil && *si.ExternalMoldCode == *item.ExternalMoldCode {
					hasStockIn = true
					break
				}
			}
			if !hasStockIn {
				return false // 有外来模具还未入库
			}
		}
		// 如果两者都为空，跳过该项
	}

	return true // 所有模具都已入库
}

// ensureMoldMaterial 确保模具有对应的物料记录（mat_type=3）
// 支持一个模具多个产品：物料编码 = 模具编号-产品ID
func (s *ProductionOrderService) ensureMoldMaterial(moldID int64, productID *int64) (int64, error) {
	// 1. 查找模具（使用结构体缓存的 DAO）
	mold, err := s.moldDAO.FindByID(moldID)
	if err != nil {
		return 0, errors.New("模具不存在")
	}

	// 2. 生成物料编码：模具编号-产品ID（支持一个模具多个产品）
	var matCode string
	if productID != nil && *productID > 0 {
		matCode = fmt.Sprintf("%s-P%d", mold.MoldCode, *productID)
	} else {
		matCode = mold.MoldCode // 没有产品ID时，只用模具编号
	}

	// 3. 查找是否已有对应的物料记录
	existingMaterial, err := s.materialDAO.FindByCode(matCode)
	if err == nil && existingMaterial != nil {
		// 找到了，直接返回
		return existingMaterial.ID, nil
	}

	// 4. 如果不存在，创建新的物料记录
	// 获取产品名称（如果有product_id）
	var productName string
	if productID != nil && *productID > 0 {
		var product entity.Product
		if err := dao.GetDB().First(&product, *productID).Error; err == nil {
			productName = product.ProName
		}
	}

	// 如果没有产品名称，使用模具名称
	if productName == "" {
		productName = mold.MoldName
	}

	newMaterial := &entity.Material{
		MatCode:   matCode,
		MatName:   productName,
		MatType:   3, // 3=产品
		Unit:      "个",
		Price:     0,
		StockWarn: 0,
	}
	if err := s.materialDAO.Create(newMaterial); err != nil {
		return 0, errors.New("创建物料记录失败: " + err.Error())
	}
	return newMaterial.ID, nil
}

// ensureExternalMoldMaterial 确保外来模具有对应的物料记录（mat_type=3）
func (s *ProductionOrderService) ensureExternalMoldMaterial(externalMoldCode string, productName string) (int64, error) {
	// 1. 查找是否已有对应的物料记录（mat_code 有唯一索引）
	matCode := fmt.Sprintf("EM-%s", externalMoldCode)
	existingMaterial, err := s.materialDAO.FindByCode(matCode)
	if err == nil && existingMaterial != nil {
		return existingMaterial.ID, nil
	}

	// 2. 如果不存在，创建新的物料记录
	newMaterial := &entity.Material{
		MatCode:   matCode,
		MatName:   productName,
		MatType:   3, // 3=产品
		Unit:      "个",
		Price:     0,
		StockWarn: 0,
	}
	if err := s.materialDAO.Create(newMaterial); err != nil {
		return 0, errors.New("创建物料记录失败: " + err.Error())
	}
	return newMaterial.ID, nil
}

// EnsureMoldMaterial 确保自有模具有对应的物料记录（mat_type=3）
func (s *ProductionOrderService) EnsureMoldMaterial(moldID int64, productID *int64) (int64, error) {
	return s.ensureMoldMaterial(moldID, productID)
}

// EnsureExternalMoldMaterial 确保外来模具有对应的物料记录（mat_type=3）
func (s *ProductionOrderService) EnsureExternalMoldMaterial(externalMoldCode string, productName string) (int64, error) {
	return s.ensureExternalMoldMaterial(externalMoldCode, productName)
}

// InStockDirectly 直接增加库存（公开方法，供外部调用）
func (s *ProductionOrderService) InStockDirectly(materialID, warehouseID int64, quantity float64) error {
	return s.stockService.InStock(materialID, warehouseID, quantity)
}

// GetStockByMaterialAndWarehouse 获取指定物料在指定仓库的库存（公开方法，供外部调用）
func (s *ProductionOrderService) GetStockByMaterialAndWarehouse(materialID, warehouseID int64) (*entity.Stock, error) {
	return s.stockService.GetStockByMaterialAndWarehouse(materialID, warehouseID)
}

// UpdateStockRecord 更新库存记录（公开方法，供外部调用）
func (s *ProductionOrderService) UpdateStockRecord(stock *entity.Stock) error {
	return s.stockService.UpdateStock(stock)
}

// GetStockByMaterialProductAndWarehouse 获取指定物料、产品和仓库的库存（公开方法，供外部调用）
func (s *ProductionOrderService) GetStockByMaterialProductAndWarehouse(materialID int64, productID *int64, warehouseID int64) (*entity.Stock, error) {
	return s.stockService.GetStockByMaterialProductAndWarehouse(materialID, productID, warehouseID)
}

// CheckStock 盘点库存：直接设置库存为绝对值
func (s *ProductionOrderService) CheckStock(materialID int64, productID *int64, warehouseID int64, actualQty float64) (diff float64, err error) {
	return s.stockService.SetStockQuantity(materialID, productID, warehouseID, actualQty)
}

// calculateStockByProductName 按 product_name + 仓库 + 模具 动态计算库存
// 与 GetCurrentStock 接口保持一致：统计 production_stock_in 入库数量减去 production_stock_out 出库数量
// 用于兼容历史数据 product_id 为 NULL、仅按产品名+模具+仓库维度管理库存的场景
func (s *ProductionOrderService) calculateStockByProductName(productName string, warehouseID int64, moldID *int64, externalMoldCode *string) (float64, error) {
	if productName == "" {
		return 0, errors.New("产品名称不能为空")
	}

	// 入库数量
	stockInQuery := dao.GetDB().Table("production_stock_in").
		Select("COALESCE(SUM(quantity), 0)").
		Where("product_name = ?", productName).
		Where("warehouse_id = ?", warehouseID)
	if moldID != nil && *moldID > 0 {
		stockInQuery = stockInQuery.Where("mold_id = ?", *moldID)
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		stockInQuery = stockInQuery.Where("external_mold_code = ?", *externalMoldCode)
	}

	// 出库数量
	stockOutQuery := dao.GetDB().Model(&entity.ProductionStockOut{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("product_name = ?", productName).
		Where("warehouse_id = ?", warehouseID)
	if moldID != nil && *moldID > 0 {
		stockOutQuery = stockOutQuery.Where("mold_id = ?", *moldID)
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		stockOutQuery = stockOutQuery.Where("external_mold_code = ?", *externalMoldCode)
	}

	var totalIn, totalOut float64
	if err := stockInQuery.Scan(&totalIn).Error; err != nil {
		return 0, err
	}
	if err := stockOutQuery.Scan(&totalOut).Error; err != nil {
		return 0, err
	}

	current := totalIn - totalOut
	if current < 0 {
		current = 0
	}
	return current, nil
}

// decrementStock 扣减库存：
// - 有 productID：使用 stockService 更新 stock 表汇总
// - productID 为空（product_name 路径）：跳过 stock 表更新，库存由 production_stock_in/out 流水表动态计算
func (s *ProductionOrderService) decrementStock(materialID int64, productID *int64, warehouseID int64, qty float64) error {
	// productID 有值：按原有逻辑更新 stock 表
	if productID != nil && *productID > 0 {
		return s.stockService.OutStockWithProduct(materialID, productID, warehouseID, qty)
	}

	// productID 为空：基于流水的库存自动随出库记录扣减，无需手动更新 stock 表
	// 尝试同步更新 product_id IS NULL 的 stock 汇总（兼容老数据），但失败不阻塞出库流程
	stock, err := s.stockService.GetStockByMaterialAndWarehouse(materialID, warehouseID)
	if err == nil && stock != nil {
		if stock.StockQty >= qty {
			stock.StockQty -= qty
			if updateErr := s.stockService.UpdateStock(stock); updateErr != nil {
				logger.Warn("同步更新 product_id 为空的库存汇总失败（不影响出库）", zap.Error(updateErr))
			}
		}
	}
	return nil
}

// ShipmentOrder 生产订单出货
func (s *ProductionOrderService) ShipmentOrder(id int64, remark string) error {
	order, err := s.orderDAO.FindByID(id)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	// 获取当前状态对应的字典值（语义化字符串）
	currentStatusValue, err := s.flowService.GetDictValue("production-order-status", order.Status)
	if err != nil {
		return errors.New("订单状态配置异常")
	}

	flow, err := s.flowService.ValidateAction(currentStatusValue, "shipment")
	if err != nil {
		return errors.New("订单状态不允许出货，" + err.Error())
	}

	// 根据流转规则更新订单状态为目标状态
	order.Status = flow.ToStatusValue // 直接使用语义化字符串
	if remark != "" {
		order.Remark = remark
	}
	if err := s.orderDAO.Update(order); err != nil {
		return err
	}
	s.syncMasterOrderStatusAfterUpdate(order)
	return nil
}

// CompleteOrder 完成订单
func (s *ProductionOrderService) CompleteOrder(id int64) error {
	order, err := s.orderDAO.FindByID(id)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	shippedQty, err := s.GetShippedQty(id)
	if err != nil {
		return errors.New("查询出货记录失败")
	}
	if shippedQty < order.PlanQty {
		return fmt.Errorf("出货数量(%d)不足计划数量(%d)，无法完成订单", shippedQty, order.PlanQty)
	}

	items, err := s.orderItemDAO.FindByOrderID(id)
	if err != nil {
		return errors.New("查询订单明细失败")
	}

	var currentStatusValue string

	// 如果有明细项，检查所有明细项的状态
	if len(items) > 0 {
		firstItemStatus := items[0].Status
		allSameStatus := true
		for _, item := range items {
			if item.Status != firstItemStatus {
				allSameStatus = false
				break
			}
		}

		if !allSameStatus {
			return errors.New("订单明细状态不一致，无法完成订单")
		}

		// 使用明细项的状态
		currentStatusValue, err = s.flowService.GetDictValue("production-order-status", firstItemStatus)
		if err != nil {
			return errors.New("订单明细状态配置异常")
		}
	} else {
		// 没有明细项，使用订单主表状态（兼容旧数据）
		currentStatusValue, err = s.flowService.GetDictValue("production-order-status", order.Status)
		if err != nil {
			return errors.New("订单状态配置异常")
		}
	}

	flow, err := s.flowService.ValidateAction(currentStatusValue, "finish_order")
	if err != nil {
		return errors.New("订单状态不允许完成，" + err.Error())
	}

	// 根据流转规则更新订单状态为目标状态
	order.Status = flow.ToStatusValue // 直接使用语义化字符串
	now := time.Now()
	order.ActualEndDate = &now

	// 如果有明细项，同步更新所有明细项的状态
	if len(items) > 0 {
		for i := range items {
			items[i].Status = flow.ToStatusValue
			if err := s.orderItemDAO.Update(&items[i]); err != nil {
				return errors.New("更新订单明细状态失败")
			}
		}
	}

	if err := s.orderDAO.Update(order); err != nil {
		return err
	}
	s.syncMasterOrderStatusAfterUpdate(order)
	return nil
}

// CompleteProductionItem 完成单个生产明细项
func (s *ProductionOrderService) CompleteProductionItem(itemId int64) error {
	item, err := s.orderItemDAO.FindByID(itemId)
	if err != nil {
		return errors.New("订单明细不存在")
	}

	// 获取当前状态对应的字典值（语义化字符串）
	currentStatusValue, err := s.flowService.GetDictValue("production-order-status", item.Status)
	if err != nil {
		return errors.New("订单明细状态配置异常")
	}

	flow, err := s.flowService.ValidateAction(currentStatusValue, constant.ProductionActionCompleteProduction)
	if err != nil {
		return errors.New("订单明细状态不允许结束生产，" + err.Error())
	}

	item.Status = flow.ToStatusValue
	now := time.Now()
	item.ActualCompletionTime = &now
	if err := s.orderItemDAO.Update(item); err != nil {
		return errors.New("更新订单明细状态失败")
	}

	items, err := s.orderItemDAO.FindByOrderID(item.OrderID)
	if err != nil {
		return errors.New("查询订单明细失败")
	}

	allCompleted := true
	for _, i := range items {
		if i.Status != flow.ToStatusValue {
			allCompleted = false
			break
		}
	}

	// 如果所有明细项都已完成，更新订单主表状态
	if allCompleted {
		order, err := s.orderDAO.FindByID(item.OrderID)
		if err == nil {
			order.Status = flow.ToStatusValue
			order.ActualEndDate = &now
			s.orderDAO.Update(order)
		}
	}

	return nil
}

// GetOrdersByPage 分页查询订单
func (s *ProductionOrderService) GetOrdersByPage(query dto.ProductionOrderQueryDTO) ([]entity.ProductionOrder, int64, error) {
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 || query.PageSize > 100 {
		query.PageSize = 10
	}

	result, err := s.orderDAO.FindByPage(query)
	if err != nil {
		return nil, 0, err
	}
	// result.List 是 *[]entity.ProductionOrder，需要解引用
	if result.List == nil {
		return []entity.ProductionOrder{}, result.Total, nil
	}
	return *(result.List.(*[]entity.ProductionOrder)), result.Total, nil
}

// BatchGetActionCodes 批量获取订单的可用操作编码，直接更新 orders 的 AvailableActionCodes 字段
func (s *ProductionOrderService) BatchGetActionCodes(orders []entity.ProductionOrder) {
	if len(orders) == 0 {
		return
	}

	// 第一步：确定每个订单的当前状态值
	type orderStatus struct {
		statusValue string
		hasItems    bool
	}
	orderStatuses := make(map[int64]orderStatus)

	for _, order := range orders {
		// DependencyStatus == WAITING_MOLD 不返回任何操作
		if order.DependencyStatus == "WAITING_MOLD" {
			order.Status = "" // 标记为无操作，但保留其他数据
			orderStatuses[order.ID] = orderStatus{statusValue: "", hasItems: false}
			continue
		}

		// 优先使用明细项状态
		if len(order.Items) > 0 {
			firstItemStatus := order.Items[0].Status
			allSameStatus := true
			for _, item := range order.Items {
				if item.Status != firstItemStatus {
					allSameStatus = false
					break
				}
			}
			if allSameStatus {
				// 直接使用明细项的状态值（Status 就是 dict_value）
				orderStatuses[order.ID] = orderStatus{statusValue: firstItemStatus, hasItems: true}
				continue
			}
			// 明细项状态不一致，不返回操作
			orderStatuses[order.ID] = orderStatus{statusValue: "", hasItems: true}
			continue
		}

		// 没有明细项，使用订单主表状态
		if order.Status != "" {
			orderStatuses[order.ID] = orderStatus{statusValue: order.Status, hasItems: false}
		} else {
			orderStatuses[order.ID] = orderStatus{statusValue: "", hasItems: false}
		}
	}

	// 第二步：收集所有非空的状态值，批量查询流转规则
	statusValueSet := make(map[string]bool)
	for _, os := range orderStatuses {
		if os.statusValue != "" {
			statusValueSet[os.statusValue] = true
		}
	}

	if len(statusValueSet) > 0 {
		statusValues := make([]string, 0, len(statusValueSet))
		for v := range statusValueSet {
			statusValues = append(statusValues, v)
		}

		flows, err := s.flowService.flowDAO.FindByFromStatusValues(statusValues)
		if err == nil && len(flows) > 0 {
			actionMap := make(map[string][]string)
			for _, flow := range flows {
				actionMap[flow.FromStatusValue] = append(actionMap[flow.FromStatusValue], flow.ActionCode)
			}

			// 第三步：为每个订单设置动作编码
			for i := range orders {
				os, ok := orderStatuses[orders[i].ID]
				if !ok || os.statusValue == "" {
					orders[i].AvailableActionCodes = []string{}
					continue
				}
				if codes, ok := actionMap[os.statusValue]; ok {
					orders[i].AvailableActionCodes = codes
				} else {
					orders[i].AvailableActionCodes = []string{}
				}
			}
		}
	}
}

// GetStatistics 获取订单统计
func (s *ProductionOrderService) GetStatistics() (map[string]int64, error) {
	return s.orderDAO.GetStatistics()
}

// CreateMoldUsageRecord 创建模具使用记录(在生产订单入库时调用)
func (s *ProductionOrderService) CreateMoldUsageRecord(orderID int64, shots, okQty, badQty int) error {
	order, err := s.orderDAO.FindByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}

	// 只有自有模具才创建使用记录
	if order.MoldSourceType != "SELF_MADE" || order.MoldID == nil {
		return nil
	}

	// 使用事务创建使用记录并更新模具信息
	tx := dao.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.usageService.CreateUsageRecord(tx, *order.MoldID, orderID, shots, okQty, badQty); err != nil {
		tx.Rollback()
		return err
	}

	order.MoldShots = shots
	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return errors.New("更新订单失败: " + err.Error())
	}

	return tx.Commit().Error
}

// GetMoldUsageHistory 获取模具使用历史
func (s *ProductionOrderService) GetMoldUsageHistory(moldID int64, page, pageSize int, moldCode string) ([]entity.MoldUsageRecord, int64, error) {
	return s.usageService.GetUsageHistory(moldID, page, pageSize, moldCode)
}

// GetMoldUsageStats 获取模具使用统计
func (s *ProductionOrderService) GetMoldUsageStats(moldID int64) (map[string]interface{}, error) {
	return s.usageService.GetUsageStats(moldID)
}

// GetStockInList 分页查询入库记录
func (s *ProductionOrderService) GetStockInList(page, pageSize int, orderCode string, productName string, stockCode string) ([]entity.ProductionStockIn, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.stockInDAO.FindByPage(page, pageSize, orderCode, productName, stockCode)
}

// GetStockInByID 获取入库记录详情
func (s *ProductionOrderService) GetStockInByID(id int64) (*entity.ProductionStockIn, error) {
	return s.stockInDAO.FindByID(id)
}

// UpdateStockIn 更新入库记录
func (s *ProductionOrderService) UpdateStockIn(stockIn *entity.ProductionStockIn) error {
	existing, err := s.stockInDAO.FindByID(stockIn.ID)
	if err := utils.EnsureExists(existing, err, "入库记录"); err != nil {
		return err
	}

	updateData := make(map[string]interface{})
	// 注意：批次号是系统自动生成的，不允许编辑
	updateData["remark"] = stockIn.Remark
	updateData["quantity"] = stockIn.Quantity
	if !stockIn.StockInDate.IsZero() {
		updateData["stock_in_date"] = stockIn.StockInDate
	}

	// 执行更新
	return dao.GetDB().Model(&entity.ProductionStockIn{}).Where("id = ?", stockIn.ID).Updates(updateData).Error
}

// GetStockStatistics 获取库存统计数据
func (s *ProductionOrderService) GetStockStatistics() (map[string]interface{}, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 使用单次查询获取所有统计数据（优化）
	type StockStats struct {
		TotalStockIn int64 `gorm:"column:total_stock_in"`
		TotalRecords int64 `gorm:"column:total_records"`
		TodayStockIn int64 `gorm:"column:today_stock_in"`
	}

	var stats StockStats
	err := dao.GetDB().Raw(`
		SELECT 
			COALESCE(SUM(quantity), 0) as total_stock_in,
			COUNT(*) as total_records,
			COALESCE(SUM(CASE WHEN stock_in_date >= ? THEN quantity ELSE 0 END), 0) as today_stock_in
		FROM production_stock_in
	`, todayStart).Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_stock_in": stats.TotalStockIn,
		"total_records":  stats.TotalRecords,
		"today_stock_in": stats.TodayStockIn,
	}, nil
}

// GetStockTrend 获取库存趋势数据（按日期统计入库和出库）
func (s *ProductionOrderService) GetStockTrend(days int) ([]map[string]interface{}, error) {
	if days <= 0 {
		days = 30 // 默认30天
	}

	now := time.Now()
	startDate := now.AddDate(0, 0, -days+1)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())

	logger.Debug("库存趋势查询参数",
		zap.Int("days", days),
		zap.Time("start_date", startDate),
		zap.Time("now", now))

	type DailyStock struct {
		Date  string `json:"date"`
		Total int64  `json:"total"`
	}

	var dailyStockIns []DailyStock
	dbResultIn := dao.GetDB().Model(&entity.ProductionStockIn{}).
		Select("DATE_FORMAT(stock_in_date, '%Y-%m-%d') as date, SUM(quantity) as total").
		Where("stock_in_date >= ?", startDate).
		Group("DATE_FORMAT(stock_in_date, '%Y-%m-%d')").
		Order("date ASC").
		Scan(&dailyStockIns)

	logger.Debug("库存趋势-入库数据查询结果",
		zap.Int64("rows_affected", dbResultIn.RowsAffected),
		zap.Any("data", dailyStockIns))

	var dailyStockOuts []DailyStock
	dbResultOut := dao.GetDB().Model(&entity.ProductionStockOut{}).
		Select("DATE_FORMAT(stock_out_date, '%Y-%m-%d') as date, SUM(quantity) as total").
		Where("stock_out_date >= ?", startDate).
		Group("DATE_FORMAT(stock_out_date, '%Y-%m-%d')").
		Order("date ASC").
		Scan(&dailyStockOuts)

	logger.Debug("库存趋势-出库数据查询结果",
		zap.Int64("rows_affected", dbResultOut.RowsAffected),
		zap.Any("data", dailyStockOuts))

	result := make([]map[string]interface{}, 0)
	currentDate := startDate
	for i := 0; i < days; i++ {
		dateStr := currentDate.Format("2006-01-02")
		stockInTotal := int64(0)
		stockOutTotal := int64(0)

		// 查找该日期的入库数据
		for _, ds := range dailyStockIns {
			if ds.Date == dateStr {
				stockInTotal = ds.Total
				break
			}
		}

		// 查找该日期的出库数据
		for _, ds := range dailyStockOuts {
			if ds.Date == dateStr {
				stockOutTotal = ds.Total
				break
			}
		}

		result = append(result, map[string]interface{}{
			"date":      dateStr,
			"stock_in":  stockInTotal,
			"stock_out": stockOutTotal,
		})

		currentDate = currentDate.AddDate(0, 0, 1)
	}

	if len(result) > 0 {
		logger.Debug("库存趋势最终结果",
			zap.Int("count", len(result)),
			zap.Any("first_item", result[0]),
			zap.Any("last_item", result[len(result)-1]))
	}

	return result, nil
}

// GetProductStockList 获取产品库存汇总列表（按模具分组）
func (s *ProductionOrderService) GetProductStockList(page, pageSize int, moldCode string, productName string) ([]dao.ProductStockSummary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.stockInDAO.FindProductStockSummary(page, pageSize, moldCode, productName)
}

// GetStockRecordsByProductID 获取指定产品的出入库记录（合并入库和出库）
// GetStockRecordsByProductID 获取指定模具的出入库记录（保留用于向后兼容）
func (s *ProductionOrderService) GetStockRecordsByProductID(moldID int64, page, pageSize int, sortOrder, statusFilters string) ([]map[string]interface{}, int64, dao.ProductStockSummary, error) {
	return s.GetStockRecordsByMoldIDAndProduct(moldID, nil, page, pageSize, sortOrder, statusFilters)
}

// GetStockRecordsByMoldIDAndProduct 获取指定模具和产品的出入库记录
func (s *ProductionOrderService) GetStockRecordsByMoldIDAndProduct(moldID int64, productID *int64, page, pageSize int, sortOrder, statusFilters string) ([]map[string]interface{}, int64, dao.ProductStockSummary, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc" // 默认降序
	}

	// 【优化】使用 UNION ALL 在数据库层合并、排序、分页，避免全量查询+内存处理
	records, total, err := s.stockInDAO.FindMergedStockRecords(moldID, productID, page, pageSize, sortOrder, statusFilters)
	if err != nil {
		return nil, 0, dao.ProductStockSummary{}, err
	}

	var summary dao.ProductStockSummary
	if productID != nil {
		summary, err = s.stockInDAO.FindProductStockSummaryByMoldIDAndProduct(moldID, productID)
	} else {
		summary, err = s.stockInDAO.FindProductStockSummaryByMoldID(moldID)
	}
	if err != nil {
		return nil, 0, dao.ProductStockSummary{}, err
	}

	return records, total, summary, nil
}

// GetStockRecordsByExternalMoldCode 获取指定外来模具的出入库记录（合并入库和出库）
func (s *ProductionOrderService) GetStockRecordsByExternalMoldCode(externalMoldCode string, page, pageSize int, sortOrder, statusFilters string) ([]interface{}, int64, dao.ProductStockSummary, error) {
	return s.GetStockRecordsByExternalMoldCodeAndProduct(externalMoldCode, nil, page, pageSize, sortOrder, statusFilters)
}

// GetStockRecordsByExternalMoldCodeAndProduct 获取指定外来模具和产品的出入库记录
// 【优化】使用 UNION ALL 在数据库层合并、排序、分页，避免全量查询+内存处理
func (s *ProductionOrderService) GetStockRecordsByExternalMoldCodeAndProduct(externalMoldCode string, productID *int64, page, pageSize int, sortOrder, statusFilters string) ([]interface{}, int64, dao.ProductStockSummary, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// 使用 UNION ALL 在数据库层合并、排序、分页
	records, total, err := s.stockInDAO.FindMergedStockRecordsByExternalMoldCode(externalMoldCode, productID, page, pageSize, sortOrder, statusFilters)
	if err != nil {
		return nil, 0, dao.ProductStockSummary{}, err
	}

	// 转换为 interface{} 以兼容原有返回类型
	resultRecords := make([]interface{}, len(records))
	for i, record := range records {
		resultRecords[i] = record
	}

	var summary dao.ProductStockSummary
	if productID != nil {
		summary, err = s.stockInDAO.FindProductStockSummaryByExternalMoldCodeAndProduct(externalMoldCode, productID)
	} else {
		summary, err = s.stockInDAO.FindProductStockSummaryByExternalMoldCode(externalMoldCode)
	}
	if err != nil {
		return nil, 0, dao.ProductStockSummary{}, err
	}

	return resultRecords, total, summary, nil
}

// GenerateStockOutCode 生成出库单号 SO-YYYYMMDD-001
func (s *ProductionOrderService) GenerateStockOutCode() (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")
	prefix := fmt.Sprintf("SO-%s-", dateStr)

	// 查询当天最大序号
	var maxCode string
	dao.GetDB().Model(&entity.ProductionStockOut{}).
		Where("stock_code LIKE ?", prefix+"%").
		Order("stock_code DESC").
		Limit(1).
		Pluck("stock_code", &maxCode)

	seq := 1
	if maxCode != "" {
		seqStr := strings.TrimPrefix(maxCode, prefix)
		if num, err := strconv.Atoi(seqStr); err == nil {
			seq = num + 1
		}
	}

	return fmt.Sprintf("%s%03d", prefix, seq), nil
}

// StockOutOrder 生产订单出库
func (s *ProductionOrderService) StockOutOrder(orderID int64, warehouseID int64, moldID *int64, externalMoldCode *string, productID *int64, productName string, quantity int, batchNo string, remark string, customerName *string, deliveryDate *time.Time, deliveryMethod *string, expressCompany *string, expressNo *string, photos *string, totalWeight *float64, driverPhone *string, driverPlate *string, outType string, outReason *string, outReasonRemark *string, operatorID int64) error {
	if quantity <= 0 {
		return errors.New("出库数量必须大于0")
	}

	// 如果 product_id 为空但 product_name 有值，尝试按名称匹配产品档案
	if productID == nil && productName != "" {
		if p, err := dao.NewProductDAO().FindByName(productName); err == nil {
			productID = &p.ID
		}
	}

	var materialID int64
	var err error

	if moldID != nil && *moldID > 0 {
		// 自有模具：找到对应的物料记录
		materialID, err = s.ensureMoldMaterial(*moldID, productID)
		if err != nil {
			return errors.New("获取物料信息失败: " + err.Error())
		}
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		// 外来模具：使用外来模具编号作为物料标识
		materialID, err = s.ensureExternalMoldMaterial(*externalMoldCode, productName)
		if err != nil {
			return errors.New("获取外来模具物料信息失败: " + err.Error())
		}
	} else {
		return errors.New("请提供模具ID或外来模具编号")
	}

	// 优先按 product_id 查；查不到时按 product_name + 仓库 + 模具 动态计算（兼容历史数据 product_id 为 NULL 的场景）
	var currentStockQty float64
	if productID != nil && *productID > 0 {
		stock, err := s.stockService.GetStockByMaterialProductAndWarehouse(materialID, productID, warehouseID)
		if err != nil {
			// 找不到包含 product_id 的库存，尝试不区分产品的库存（向后兼容）
			stock, err = s.stockService.GetStockByMaterialAndWarehouse(materialID, warehouseID)
			if err != nil {
				return errors.New("库存记录不存在，请先进行入库操作")
			}
			currentStockQty = stock.StockQty
		} else {
			currentStockQty = stock.StockQty
		}
	} else {
		// productID 为空：按 product_name + 仓库 + 模具 在 production_stock_in/out 中动态计算库存
		// 与 GetCurrentStock 接口保持一致，确保前端库存数量与出库校验口径统一
		currentStockQty, err = s.calculateStockByProductName(productName, warehouseID, moldID, externalMoldCode)
		if err != nil {
			return errors.New("查询库存失败: " + err.Error())
		}
	}

	if currentStockQty < float64(quantity) {
		return errors.New("库存不足，当前库存: " + fmt.Sprintf("%.2f", currentStockQty))
	}

	var stockCode string
	if outType == "normal" {
		code, err := s.GenerateStockOutCode()
		if err != nil {
			return errors.New("生成出库单号失败: " + err.Error())
		}
		stockCode = code
	}

	// 自动生成批次号（如果前端未提供）
	if batchNo == "" {
		batchNo = s.GenerateStockOutBatchNo(moldID, externalMoldCode, productID)
	}

	now := time.Now()

	// 处理 order_id: 如果为 0 或负数，则设为 nil
	var orderIDPtr *int64
	if orderID > 0 {
		orderIDPtr = &orderID
	}

	// 如果 customer_name 为空，尝试通过模具查询关联的生产订单获取客户名称
	if (customerName == nil || *customerName == "") && moldID != nil && *moldID > 0 {
		// 通过模具ID查询最新的生产订单
		var order entity.ProductionOrder
		if err := dao.GetDB().
			Preload("Customer").
			Joins("JOIN production_order_item ON production_order_item.order_id = production_order.id").
			Where("production_order_item.mold_id = ?", *moldID).
			Order("production_order.create_time DESC").
			First(&order).Error; err == nil && order.Customer != nil {
			// 找到关联订单且有客户信息，自动填充客户名称
			customerName = &order.Customer.CusName
		}
	}

	// 如果 customer_name 仍然为空，且是外来模具，尝试从外来模具档案中获取客户名称
	if (customerName == nil || *customerName == "") && externalMoldCode != nil && *externalMoldCode != "" {
		var archive entity.ExternalMoldArchive
		if err := dao.GetDB().
			Where("mold_code = ?", *externalMoldCode).
			First(&archive).Error; err == nil && archive.CustomerName != nil && *archive.CustomerName != "" {
			// 从外来模具档案中找到客户名称
			customerName = archive.CustomerName
		}
	}

	stockOut := &entity.ProductionStockOut{
		StockCode:        stockCode,
		OrderID:          orderIDPtr,
		WarehouseID:      warehouseID,
		MoldID:           moldID,
		ExternalMoldCode: externalMoldCode,
		ProductID:        productID, // 保存产品ID，用于区分同一模具不同产品的出库
		ProductName:      productName,
		CustomerName:     customerName,
		Quantity:         quantity,
		BatchNo:          batchNo,
		StockOutDate:     &now,
		OutType:          outType,
		OutReason:        outReason,
		OutReasonRemark:  outReasonRemark,
		DeliveryDate:     deliveryDate,
		DeliveryMethod:   deliveryMethod,
		ExpressCompany:   expressCompany,
		ExpressNo:        expressNo,
		Photos:           photos,
		TotalWeight:      totalWeight,
		DriverPhone:      driverPhone,
		DriverPlate:      driverPlate,
		OperatorID:       &operatorID,
		Remark:           remark,
	}

	if err := s.stockOutDAO.Create(stockOut); err != nil {
		return errors.New("创建出库记录失败: " + err.Error())
	}

	// 如果stock的product_id为NULL但当前请求有product_id，需要更新stock记录的product_id
	// 注意：仅在按 productID 查询路径下有 stock 记录；按 product_name 动态计算路径下没有 stock 记录，跳过此步
	if productID != nil && *productID > 0 {
		stock, stockErr := s.stockService.GetStockByMaterialProductAndWarehouse(materialID, productID, warehouseID)
		if stockErr == nil && stock.ProductID == nil {
			stock.ProductID = productID
			if updateErr := s.stockService.UpdateStock(stock); updateErr != nil {
				return errors.New("更新库存记录失败: " + updateErr.Error())
			}
		}
	}

	if err := s.decrementStock(materialID, productID, warehouseID, float64(quantity)); err != nil {
		return errors.New("更新库存失败: " + err.Error())
	}

	// 【关键】如果是正常发货，检查是否需要更新订单状态
	if outType == "normal" && orderIDPtr != nil && *orderIDPtr > 0 {
		// 有明确的订单ID，直接更新该订单状态
		if err := s.updateOrderStatusAfterShipment(*orderIDPtr); err != nil {
			// 记录错误但不影响出库流程
			logger.Error("更新订单状态失败", zap.Error(err))
		}
	} else if outType == "normal" && (moldID != nil || externalMoldCode != nil) {
		// 没有订单ID，但有模具信息，尝试自动关联订单并更新状态
		if err := s.autoUpdateOrderStatusAfterShipment(moldID, externalMoldCode, productID); err != nil {
			// 记录错误但不影响出库流程
			logger.Error("自动更新订单状态失败", zap.Error(err))
		}
	}

	return nil
}

// GetStockOutList 分页查询出库记录
func (s *ProductionOrderService) GetStockOutList(page, pageSize int, orderCode string, productName string, stockCode string) ([]entity.ProductionStockOut, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.stockOutDAO.FindByPage(page, pageSize, orderCode, productName, stockCode)
}

// GetStockOutByID 获取出库记录详情
func (s *ProductionOrderService) GetStockOutByID(id int64) (*entity.ProductionStockOut, error) {
	return s.stockOutDAO.FindByID(id)
}

// GetStockOutRecordsByMoldID 获取指定模具的出库记录
func (s *ProductionOrderService) GetStockOutRecordsByMoldID(moldID int64, page, pageSize int) ([]entity.ProductionStockOut, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.stockOutDAO.FindStockRecordsByMoldID(moldID, page, pageSize, "desc")
}

// DeleteStockByID 删除库存记录
func (s *ProductionOrderService) DeleteStockByID(id int64) error {
	return s.stockService.DeleteStockByID(id)
}

// DeleteStockByMoldAndProduct 根据模具和产品删除库存记录（删除对应的出入库记录）
func (s *ProductionOrderService) DeleteStockByMoldAndProduct(moldID *int64, externalMoldCode *string, productID *int64) error {
	// 使用事务确保入库和出库记录同时删除
	return dao.GetDB().Transaction(func(tx *gorm.DB) error {
		inQuery := tx.Model(&entity.ProductionStockIn{})
		outQuery := tx.Model(&entity.ProductionStockOut{})

		if moldID != nil && *moldID > 0 {
			// 自有模具
			inQuery = inQuery.Where("mold_id = ?", *moldID)
			outQuery = outQuery.Where("mold_id = ?", *moldID)
		} else if externalMoldCode != nil && *externalMoldCode != "" {
			// 外来模具
			inQuery = inQuery.Where("external_mold_code = ?", *externalMoldCode)
			outQuery = outQuery.Where("external_mold_code = ?", *externalMoldCode)
		} else {
			return errors.New("请提供模具ID或外来模具编号")
		}

		// 入库记录支持 product_id 条件（入库表有 product_id 字段）
		if productID != nil && *productID > 0 {
			inQuery = inQuery.Where("product_id = ?", *productID)
		} else {
			inQuery = inQuery.Where("product_id IS NULL")
		}

		// 注意：出库记录表没有 product_id 字段，只按模具删除

		if err := inQuery.Delete(&entity.ProductionStockIn{}).Error; err != nil {
			return errors.New("删除入库记录失败: " + err.Error())
		}

		if err := outQuery.Delete(&entity.ProductionStockOut{}).Error; err != nil {
			return errors.New("删除出库记录失败: " + err.Error())
		}

		return nil
	})
}

// GetStockOutRecordsByExternalMoldCode 获取指定外来模具的出库记录
func (s *ProductionOrderService) GetStockOutRecordsByExternalMoldCode(externalMoldCode string, page, pageSize int) ([]entity.ProductionStockOut, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.stockOutDAO.FindStockRecordsByExternalMoldCode(externalMoldCode, page, pageSize, "desc")
}

// UpdateStockOut 更新出库记录(仅允许编辑发货信息,不允许修改数量、模具等核心字段)
func (s *ProductionOrderService) UpdateStockOut(id int64, stockOutDate *time.Time, customerName *string, deliveryDate *time.Time, deliveryMethod *string, expressCompany *string, expressNo *string, photos *string, totalWeight *float64, driverPhone *string, driverPlate *string, remark string) error {
	stockOut, err := s.stockOutDAO.FindByID(id)
	if err := utils.EnsureExists(stockOut, err, "出库记录"); err != nil {
		return err
	}

	// 如果更新了照片，清理旧文件
	if photos != nil {
		logger.Debug("更新出库记录-开始处理照片更新", zap.Int64("id", id))

		oldPhotos := []string{}
		if stockOut.Photos != nil && *stockOut.Photos != "" && *stockOut.Photos != "[]" {
			if err := json.Unmarshal([]byte(*stockOut.Photos), &oldPhotos); err != nil {
				logger.Warn("更新出库记录-解析旧照片列表失败", zap.Error(err))
			}
		}

		newPhotos := []string{}
		if *photos != "" && *photos != "[]" {
			if err := json.Unmarshal([]byte(*photos), &newPhotos); err != nil {
				logger.Warn("更新出库记录-解析新照片列表失败", zap.Error(err))
			}
		}

		logger.Debug("更新出库记录-照片数量对比",
			zap.Int("old_count", len(oldPhotos)),
			zap.Int("new_count", len(newPhotos)))

		// 找出被删除的照片（旧列表中有但新列表中没有的）
		newPhotoSet := make(map[string]bool)
		for _, p := range newPhotos {
			newPhotoSet[p] = true
		}

		deletedCount := 0
		for _, oldPhoto := range oldPhotos {
			if !newPhotoSet[oldPhoto] {
				// 这个旧照片被删除了，删除物理文件
				// 数据库存储的路径是 /uploads/xxx，直接去掉开头的 / 就是相对于 server 目录的路径
				fullPath := strings.TrimPrefix(oldPhoto, "/")
				logger.Debug("更新出库记录-准备删除旧照片文件", zap.String("path", fullPath))
				if err := os.Remove(fullPath); err == nil {
					logger.Debug("更新出库记录-已删除旧照片文件", zap.String("path", fullPath))
					deletedCount++
				} else {
					logger.Warn("更新出库记录-删除旧照片文件失败",
						zap.String("path", fullPath),
						zap.Error(err))
				}
			}
		}

		logger.Info("更新出库记录-照片清理完成", zap.Int("deleted_count", deletedCount))

		stockOut.Photos = photos
	}

	if stockOutDate != nil {
		stockOut.StockOutDate = stockOutDate
	}
	if customerName != nil {
		stockOut.CustomerName = customerName
	}
	if deliveryDate != nil {
		stockOut.DeliveryDate = deliveryDate
	}
	if deliveryMethod != nil {
		stockOut.DeliveryMethod = deliveryMethod
	}
	if expressCompany != nil {
		stockOut.ExpressCompany = expressCompany
	}
	if expressNo != nil {
		stockOut.ExpressNo = expressNo
	}
	if totalWeight != nil {
		stockOut.TotalWeight = totalWeight
	}
	if driverPhone != nil {
		stockOut.DriverPhone = driverPhone
	}
	if driverPlate != nil {
		stockOut.DriverPlate = driverPlate
	}
	if remark != "" {
		stockOut.Remark = remark
	}

	if err := s.stockOutDAO.Update(stockOut); err != nil {
		return errors.New("更新出库记录失败: " + err.Error())
	}

	return nil
}

// GetStockOutListForPrint 获取用于打印的出库记录列表
func (s *ProductionOrderService) GetStockOutListForPrint(ids []int64) ([]entity.ProductionStockOut, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("出库记录ID不能为空")
	}

	// 使用批量查询优化 N+1 问题（从 N 次查询优化为 1 次）
	records, err := s.stockOutDAO.FindByIDs(ids)
	if err != nil {
		return nil, fmt.Errorf("查询出库记录失败: %w", err)
	}

	// 按传入的 ID 顺序排序，确保打印顺序一致
	recordMap := make(map[int64]entity.ProductionStockOut)
	for _, record := range records {
		recordMap[record.ID] = record
	}

	var sortedRecords []entity.ProductionStockOut
	for _, id := range ids {
		if record, ok := recordMap[id]; ok {
			sortedRecords = append(sortedRecords, record)
		}
	}

	return sortedRecords, nil
}

// updateOrderStatusAfterShipment 出库后更新订单状态
func (s *ProductionOrderService) updateOrderStatusAfterShipment(orderID int64) error {
	// 1. 获取订单信息
	order, err := s.orderDAO.FindByID(orderID)
	if err != nil {
		return errors.New("订单不存在")
	}

	// 2. 检查订单当前状态是否允许更新为已出货
	// 只有已入库状态的订单才能转为已出货
	if order.Status != constant.ProductionOrderStatusStockedIn {
		// 订单不是已入库状态，不需要更新
		return nil
	}

	// 3. 检查是否存在该订单的正常发货出库记录
	orderStockOuts, err := s.stockOutDAO.FindByOrderID(orderID)
	if err != nil {
		return errors.New("查询出库记录失败")
	}

	// 筛选正常发货的出库记录
	hasShipment := false
	var totalShippedQty int
	for _, stockOut := range orderStockOuts {
		if stockOut.OutType == "normal" {
			hasShipment = true
			totalShippedQty += stockOut.Quantity
		}
	}

	// 4. 只要有发货记录，就更新订单状态为已出货
	if hasShipment {
		currentStatusValue, err := s.flowService.GetDictValue("production-order-status", order.Status)
		if err != nil {
			return errors.New("订单状态配置异常")
		}

		flow, err := s.flowService.ValidateAction(currentStatusValue, "shipment")
		if err != nil {
			return errors.New("订单状态不允许出货")
		}

		order.Status = flow.ToStatusValue
		if err := s.orderDAO.Update(order); err != nil {
			return errors.New("更新订单状态失败")
		}

		logger.Info("订单已自动更新为已出货状态",
			zap.String("order_code", order.OrderCode),
			zap.Int("total_shipped_qty", totalShippedQty),
			zap.Int("plan_qty", order.PlanQty),
		)
	}

	return nil
}

// autoUpdateOrderStatusAfterShipment 自动关联订单并更新状态
func (s *ProductionOrderService) autoUpdateOrderStatusAfterShipment(moldID *int64, externalMoldCode *string, productID *int64) error {
	// 1. 根据模具和产品查找关联的生产订单
	var order entity.ProductionOrder
	query := dao.GetDB().Model(&entity.ProductionOrder{}).Preload("Items")

	if moldID != nil && *moldID > 0 {
		// 自有模具：通过订单明细关联
		if productID != nil && *productID > 0 {
			// 同时匹配模具和产品
			err := query.Joins("JOIN production_order_item ON production_order_item.order_id = production_order.id").
				Where("production_order_item.mold_id = ? AND production_order_item.product_id = ?", *moldID, *productID).
				Order("production_order.create_time DESC").
				First(&order).Error
			if err != nil {
				// 没有找到匹配的订单
				return nil
			}
		} else {
			// 只匹配模具
			err := query.Joins("JOIN production_order_item ON production_order_item.order_id = production_order.id").
				Where("production_order_item.mold_id = ?", *moldID).
				Order("production_order.create_time DESC").
				First(&order).Error
			if err != nil {
				return nil
			}
		}
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		// 外来模具：通过外部模具编号匹配
		if productID != nil && *productID > 0 {
			err := query.Joins("JOIN production_order_item ON production_order_item.order_id = production_order.id").
				Where("production_order_item.external_mold_code = ? AND production_order_item.product_id = ?", *externalMoldCode, *productID).
				Order("production_order.create_time DESC").
				First(&order).Error
			if err != nil {
				return nil
			}
		} else {
			err := query.Joins("JOIN production_order_item ON production_order_item.order_id = production_order.id").
				Where("production_order_item.external_mold_code = ?", *externalMoldCode).
				Order("production_order.create_time DESC").
				First(&order).Error
			if err != nil {
				return nil
			}
		}
	} else {
		return nil
	}

	// 2. 找到订单后，更新其状态
	return s.updateOrderStatusAfterShipment(order.ID)
}

// GetShippedQty 获取订单已出货数量
func (s *ProductionOrderService) GetShippedQty(orderID int64) (int, error) {
	return s.stockOutDAO.GetOrderTotalStockOutQty(orderID)
}

// BatchGetShippedQty 批量获取多个订单的已出货数量（key: orderID, value: shippedQty）
func (s *ProductionOrderService) BatchGetShippedQty(orderIDs []int64) (map[int64]int, error) {
	return s.stockOutDAO.GetOrdersTotalStockOutQty(orderIDs)
}

// ReplenishShipment 补发（已完成订单的再次出货）
func (s *ProductionOrderService) ReplenishShipment(orderID int64, remark string) error {
	order, err := s.orderDAO.FindByID(orderID)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	if order.Status != constant.ProductionOrderStatusCompleted {
		return errors.New("仅已完成订单可补发")
	}

	shippedQty, _ := s.GetShippedQty(orderID)
	remaining := order.OkQty - shippedQty
	if remaining <= 0 {
		return errors.New("库存不足，无可补发数量")
	}

	var warehouse entity.Warehouse
	if err := dao.GetDB().Where("status = ?", 1).First(&warehouse).Error; err != nil {
		return errors.New("未找到可用仓库")
	}

	now := time.Now()
	stockCode := fmt.Sprintf("RP-%s-%04d", now.Format("20060102"), orderID)
	remarkStr := "补发"
	if remark != "" {
		remarkStr += ": " + remark
	}

	stockOut := &entity.ProductionStockOut{
		StockCode:    stockCode,
		OrderID:      &orderID,
		WarehouseID:  warehouse.ID,
		ProductName:  order.ProductName,
		Quantity:     remaining,
		StockOutDate: &now,
		OutType:      "replenish",
		Remark:       remarkStr,
	}
	if err := s.stockOutDAO.Create(stockOut); err != nil {
		return errors.New("创建补发出库记录失败")
	}

	return nil
}

// syncMasterOrderStatusAfterUpdate 在子订单状态变更后同步综合订单状态
func (s *ProductionOrderService) syncMasterOrderStatusAfterUpdate(order *entity.ProductionOrder) {
	if order.MasterOrderID == nil || *order.MasterOrderID == 0 {
		return
	}
	// 异步同步，不影响主流程
	go func(masterID int64) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("异步同步综合订单状态 panic", zap.Int64("master_order_id", masterID), zap.Any("panic", r))
			}
		}()
		masterService := NewMasterOrderService(dao.GetDB())
		if err := masterService.SyncMasterOrderStatus(masterID); err != nil {
			logger.Error("同步综合订单状态失败",
				zap.Int64("master_order_id", masterID),
				zap.Error(err))
		}
	}(*order.MasterOrderID)
}

// generateProductCode 生成产品编码 P-YY-NNNNNN
func (s *ProductionOrderService) generateProductCode() (string, error) {
	now := time.Now()
	yearStr := now.Format("06")
	prefix := fmt.Sprintf("P-%s-", yearStr)

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

// ensureProductRecord 确保产品档案存在：已存在则补全缺失字段，不存在则自动创建
// db 用于指定数据库连接（事务中传入 tx，否则传 nil 使用全局 DB）
func (s *ProductionOrderService) ensureProductRecord(db *gorm.DB, productName string, moldID *int64, productImageURL string) *int64 {
	if productName == "" {
		return nil
	}

	var productDAO *dao.ProductDAO
	if db != nil {
		productDAO = dao.NewProductDAOWithDB(db)
	} else {
		productDAO = dao.NewProductDAO()
	}
	product, err := productDAO.FindByName(productName)
	if err != nil {
		// 产品不存在，创建新档案
		code, err := s.generateProductCode()
		if err != nil {
			logger.Error("生成产品编码失败", zap.String("product_name", productName), zap.Error(err))
			return nil
		}

		newProduct := &entity.Product{
			ProCode:  code,
			ProName:  productName,
			MoldID:   moldID,
			ImageURL: productImageURL,
			Unit:     "个",
			Price:    0.00,
			Weight:   0.00,
			Status:   1,
			Remark:   "自动创建自生产订单",
		}

		if err := productDAO.Create(newProduct); err != nil {
			// 并发场景下唯一约束冲突，回退查询已有记录
			if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "duplicate key") {
				existing, findErr := productDAO.FindByName(productName)
				if findErr == nil && existing != nil {
					logger.Info("并发创建产品时发现已存在，复用已有记录",
						zap.String("pro_name", productName))
					return &existing.ID
				}
			}
			logger.Error("创建产品档案失败", zap.String("product_name", productName), zap.Error(err))
			return nil
		}

		logger.Info("自动创建产品档案",
			zap.String("pro_code", code),
			zap.String("pro_name", productName))
		return &newProduct.ID
	}

	// 产品已存在，补全缺失字段
	updated := false
	if product.MoldID == nil && moldID != nil {
		product.MoldID = moldID
		updated = true
	}
	if product.ImageURL == "" && productImageURL != "" {
		product.ImageURL = productImageURL
		updated = true
	}
	if updated {
		if err := productDAO.Update(product); err != nil {
			logger.Error("补全产品档案失败", zap.String("product_name", productName), zap.Error(err))
			return &product.ID
		}
		logger.Info("补全产品档案",
			zap.String("pro_name", productName),
			zap.Bool("mold_updated", product.MoldID != nil && moldID != nil),
			zap.Bool("image_updated", product.ImageURL != "" && productImageURL != ""))
	}

	return &product.ID
}
