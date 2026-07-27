package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProductionReportService struct {
	db           *gorm.DB
	reportDAO    *dao.ProductionReportDAO
	orderDAO     *dao.ProductionOrderDAO
	usageService *MoldUsageService
	stockInDAO   *dao.ProductionStockInDAO // 新增：入库 DAO
	stockService *StockService             // 新增：库存服务（用于更新实际库存）
}

func NewProductionReportService() *ProductionReportService {
	return &ProductionReportService{
		db:           dao.GetDB(),
		reportDAO:    dao.NewProductionReportDAO(),
		orderDAO:     dao.NewProductionOrderDAO(),
		usageService: NewMoldUsageService(),
		stockInDAO:   dao.NewProductionStockInDAO(),
		stockService: NewStockService(),
	}
}

// SubmitReport 提交报工
func (s *ProductionReportService) SubmitReport(reportDTO dto.ProductionReportDTO, userID int64) error {
	order, err := s.orderDAO.FindByID(reportDTO.OrderID)
	if err != nil {
		return &response.HttpError{Code: response.BAD_REQUEST, Message: "订单不存在"}
	}

	// 只有生产中的订单可以报工
	// 【修改】现在支持明细项级别的独立状态管理
	// 如果订单有明细项，需要检查该明细项的状态
	if reportDTO.ItemID != nil && *reportDTO.ItemID > 0 {
		var item entity.ProductionOrderItem
		if err := s.db.First(&item, *reportDTO.ItemID).Error; err != nil {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "订单明细不存在"}
		}
		// 只有生产中的明细项可以报工
		if item.Status != constant.ProductionOrderStatusProducing && item.Status != "paused" {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "该明细项不在生产中，无法报工"}
		}

		// ✨ 新增：检查模具工单状态（如果有关联模具）
		if item.MoldID != nil && *item.MoldID > 0 {
			var workOrder entity.MoldWorkOrder
			err := s.db.Where("mold_id = ? AND is_cancelled = 0", *item.MoldID).
				Order("status ASC"). // 按状态升序，取最早的一个活跃工单
				First(&workOrder).Error

			if err == nil {
				// 自有模具：必须试模合格才能报工
				allowedStatuses := []string{
					constant.WorkOrderStatusTested,    // 4: 已试模
					constant.WorkOrderStatusStockedIn, // 5: 已入库
					constant.WorkOrderStatusDelivered, // 6: 已交付
				}
				isAllowed := false
				for _, status := range allowedStatuses {
					if workOrder.Status == status {
						isAllowed = true
						break
					}
				}
			if !isAllowed {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "模具尚未加工完成或试模合格，无法进行生产报工"}
			}

			// 进一步校验：必须存在"合格"的试模记录（覆盖修模后需重新试模的逻辑）
			var tryTestCount int64
			s.db.Model(&entity.MoldTryTest{}).
				Where("mold_work_id = ? AND result = ?", workOrder.ID, "合格").
				Count(&tryTestCount)

			if tryTestCount == 0 {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "该模具尚未有合格的试模记录，无法进行生产报工"}
			}
			} else if err == gorm.ErrRecordNotFound {
				// 没有找到工单记录，视为外来模具，默认允许报工
			} else {
				// 其他数据库错误
				return errors.New("查询模具工单状态失败")
			}
		}
	} else {
		// 旧逻辑：检查订单状态（兼容旧数据）
		if order.Status != constant.ProductionOrderStatusProducing {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "订单不在生产中"}
		}
	}

	reportDate, err := time.ParseInLocation("2006-01-02", reportDTO.ReportDate, time.Local)
	if err != nil {
		// 如果解析失败，尝试其他格式
		reportDate, err = time.ParseInLocation(time.RFC3339, reportDTO.ReportDate, time.Local)
		if err != nil {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "报工日期格式错误"}
		}
	}

	report := &entity.ProductionReport{
		OrderID:      reportDTO.OrderID,
		ItemID:       reportDTO.ItemID, //  新增：保存明细项ID
		ScheduleID:   reportDTO.ScheduleID,
		ReportDate:   reportDate,
		Shift:        reportDTO.Shift,
		OkQty:        reportDTO.OkQty,
		BadQty:       reportDTO.BadQty,
		ScrapQty:     reportDTO.ScrapQty,
		DeviceID:     reportDTO.DeviceID,
		WorkerID:     reportDTO.WorkerID,
		WorkHours:    reportDTO.WorkHours,
		DefectReason: reportDTO.DefectReason,
		Remark:       reportDTO.Remark,
		CreateUserID: &userID,
	}

	// 使用事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(report).Error; err != nil {
		tx.Rollback()
		return errors.New("创建报工记录失败")
	}

	// ✨ 根据是否有明细项ID，采用不同的更新逻辑
	// 缓存 item 引用，后续 autoStockIn 和模具使用记录可复用，避免重复查询
	var cachedItemPtr *entity.ProductionOrderItem
	if reportDTO.ItemID != nil && *reportDTO.ItemID > 0 {
		// 按明细项报工：只更新指定的明细项（缓存以复用，避免重复查询）
		cachedItemPtr = &entity.ProductionOrderItem{}
		if err := tx.First(cachedItemPtr, *reportDTO.ItemID).Error; err != nil {
			tx.Rollback()
			return errors.New("订单明细不存在")
		}

		cachedItemPtr.OkQty += reportDTO.OkQty
		cachedItemPtr.BadQty += reportDTO.BadQty
		cachedItemPtr.CompletedQty += (reportDTO.OkQty + reportDTO.BadQty)

		// 【修改】不再自动转为已入库
		// 用户可继续报工（超产），直到手动点击"结束生产"
		// 只有手动"结束生产"才转为 STOCKED_IN

		if err := tx.Save(cachedItemPtr).Error; err != nil {
			tx.Rollback()
			return errors.New("更新订单明细失败: " + err.Error())
		}

		// 统计该订单所有明细项的总完成数量（使用 SQL SUM 替代 Go 侧聚合）
		var orderSummary struct{ OkQty, BadQty, CompletedQty int64 }
		if err := tx.Model(&entity.ProductionOrderItem{}).
			Select("COALESCE(SUM(ok_qty),0) as ok_qty, COALESCE(SUM(bad_qty),0) as bad_qty, COALESCE(SUM(completed_qty),0) as completed_qty").
			Where("order_id = ?", order.ID).Scan(&orderSummary).Error; err != nil {
			tx.Rollback()
			return errors.New("统计订单明细失败")
		}

		order.OkQty = int(orderSummary.OkQty)
		order.BadQty = int(orderSummary.BadQty)
		order.CompletedQty = int(orderSummary.CompletedQty)
	} else {
		// 旧逻辑：没有明细项ID，按订单总体报工（兼容旧数据）
		type QtyResult struct {
			OkQty  int
			BadQty int
		}
		var result QtyResult
		if err := tx.Model(&entity.ProductionReport{}).
			Select("COALESCE(SUM(ok_qty), 0) as ok_qty, COALESCE(SUM(bad_qty), 0) as bad_qty").
			Where("order_id = ?", order.ID).
			Scan(&result).Error; err != nil {
			tx.Rollback()
			return errors.New("统计报工数量失败")
		}
		order.CompletedQty = result.OkQty + result.BadQty
		order.OkQty = result.OkQty
		order.BadQty = result.BadQty

		// 同步更新订单明细的完成数量（将总量按比例分配到各明细项）
		var items []entity.ProductionOrderItem
		if err := tx.Where("order_id = ?", order.ID).Order("sort ASC").Find(&items).Error; err == nil && len(items) > 0 {
			totalPlanQty := 0
			for _, item := range items {
				totalPlanQty += item.PlanQty
			}

			// 按比例分配完成数量到各明细项
			if totalPlanQty > 0 {
				allocatedOk := 0
				allocatedBad := 0
				allocatedCompleted := 0

				for i, item := range items {
					if i == len(items)-1 {
						// 最后一项使用剩余数量，避免精度损失
						item.OkQty = result.OkQty - allocatedOk
						item.BadQty = result.BadQty - allocatedBad
						item.CompletedQty = order.CompletedQty - allocatedCompleted
					} else {
						// 按比例分配
						ratio := float64(item.PlanQty) / float64(totalPlanQty)
						item.OkQty = int(float64(result.OkQty) * ratio)
						item.BadQty = int(float64(result.BadQty) * ratio)
						item.CompletedQty = int(float64(order.CompletedQty) * ratio)
						allocatedOk += item.OkQty
						allocatedBad += item.BadQty
						allocatedCompleted += item.CompletedQty
					}

					// 使用事务更新明细
					if err := tx.Save(&item).Error; err != nil {
						tx.Rollback()
						return errors.New("更新订单明细失败: " + err.Error())
					}
				}
			}
		}
	}

	// 【修改】移除自动转状态的逻辑
	// 原逻辑：当 CompletedQty >= PlanQty 时自动转为待质检
	// 新逻辑：不自动转状态，由用户手动点击“完成生产”按钮来转为待质检
	// 这样可以让用户在生产过程中多次报工，直到确认生产完成后再转待质检

	if err := tx.Save(order).Error; err != nil {
		tx.Rollback()
		return errors.New("更新订单进度失败")
	}

	// ✨ 新增：报工时同步自动入库（良品数量直接进入库存）
	// 简化流程：报工即入库，不再需要单独的入库操作
	if reportDTO.OkQty > 0 {
		if err := s.autoStockIn(tx, order, reportDTO, &reportDate, userID, cachedItemPtr); err != nil {
			tx.Rollback()
			return errors.New("自动入库失败: " + err.Error())
		}
	}

	// ✨ 新增：创建模具使用记录（全部使用事务连接，复用已缓存的 cachedItem）
	if cachedItemPtr != nil && cachedItemPtr.MoldID != nil && *cachedItemPtr.MoldID > 0 {
		existingUsage, _ := dao.NewMoldUsageRecordDAO().FindByMoldAndItemWithTx(tx, *cachedItemPtr.MoldID, *reportDTO.ItemID)

		if existingUsage == nil {
			// 不存在则创建新的使用记录
			productName := cachedItemPtr.ProductName
			if err := s.usageService.CreateUsageRecordForItem(
				tx,
				*cachedItemPtr.MoldID,
				order.ID,
				*reportDTO.ItemID,
				cachedItemPtr.ProductID,
				&productName,
				reportDTO.OkQty+reportDTO.BadQty, // shots 使用总数量
				reportDTO.OkQty,
				reportDTO.BadQty,
			); err != nil {
				// 记录日志但不阻断报工流程
				logger.Warn("创建模具使用记录失败",
					zap.Int64("moldID", *cachedItemPtr.MoldID),
					zap.Int64("itemID", *reportDTO.ItemID),
					zap.Error(err))
			}
		} else {
			// 已存在则更新数量（累加）
			existingUsage.TotalShots += (reportDTO.OkQty + reportDTO.BadQty)
			existingUsage.OkQty += reportDTO.OkQty
			existingUsage.BadQty += reportDTO.BadQty
			existingUsage.EndDate = &reportDate

			if err := tx.Save(existingUsage).Error; err != nil {
				logger.Warn("更新模具使用记录失败",
					zap.Int64("usageID", existingUsage.ID),
					zap.Error(err))
			}

			if err := s.usageService.UpdateMoldShotsWithTx(tx, *cachedItemPtr.MoldID, reportDTO.OkQty+reportDTO.BadQty); err != nil {
				logger.Warn("更新模具累计模次失败",
					zap.Int64("moldID", *cachedItemPtr.MoldID),
					zap.Error(err))
			}
		}
	} else if order.MoldID != nil && *order.MoldID > 0 {
		// 旧逻辑：按订单总体报工（兼容旧数据，使用事务连接）
		existingUsage, _ := dao.NewMoldUsageRecordDAO().FindByMoldAndOrder(*order.MoldID, order.ID)

		if existingUsage == nil {
			// 不存在则创建新的使用记录
			if err := s.usageService.CreateUsageRecord(
				tx,
				*order.MoldID,
				order.ID,
				reportDTO.OkQty+reportDTO.BadQty, // shots 使用总数量
				reportDTO.OkQty,
				reportDTO.BadQty,
			); err != nil {
				logger.Warn("创建模具使用记录失败",
					zap.Int64("moldID", *order.MoldID),
					zap.Int64("orderID", order.ID),
					zap.Error(err))
			}
		} else {
			// 已存在则更新数量（累加）
			existingUsage.TotalShots += (reportDTO.OkQty + reportDTO.BadQty)
			existingUsage.OkQty += reportDTO.OkQty
			existingUsage.BadQty += reportDTO.BadQty
			existingUsage.EndDate = &reportDate

			if err := tx.Save(existingUsage).Error; err != nil {
				logger.Warn("更新模具使用记录失败",
					zap.Int64("usageID", existingUsage.ID),
					zap.Error(err))
			}

			if err := s.usageService.UpdateMoldShots(*order.MoldID, reportDTO.OkQty+reportDTO.BadQty); err != nil {
				logger.Warn("更新模具累计模次失败",
					zap.Int64("moldID", *order.MoldID),
					zap.Error(err))
			}
		}
	}

	return tx.Commit().Error
}

// GetReportsByOrder 获取订单的报工记录
func (s *ProductionReportService) GetReportsByOrder(orderID int64) ([]entity.ProductionReport, error) {
	return s.reportDAO.FindByOrderID(orderID)
}

// autoStockIn 报工时自动入库：将良品数量直接进入库存
// 简化流程：报工即入库，CompletedQty 与实际库存始终一致
// item 参数用于传入已缓存的订单明细，避免重复查询（可选，为 nil 时自动查询）
func (s *ProductionReportService) autoStockIn(tx *gorm.DB, order *entity.ProductionOrder, reportDTO dto.ProductionReportDTO, reportDate *time.Time, userID int64, item *entity.ProductionOrderItem) error {
	// 1. 获取默认仓库（取第一个正常状态的仓库）
	warehouseID, err := s.getDefaultWarehouseID(tx)
	if err != nil {
		return fmt.Errorf("获取默认仓库失败: %w", err)
	}

	// 2. 生成入库单号
	stockCode, err := s.generateStockInCode(tx)
	if err != nil {
		return fmt.Errorf("生成入库单号失败: %w", err)
	}

	// 3. 根据是否有明细项，确定产品信息（优先复用已缓存的 item 避免重复查询）
	var moldID *int64
	var externalMoldCode *string
	var productID *int64
	var productName string

	if item != nil {
		moldID = item.MoldID
		externalMoldCode = item.ExternalMoldCode
		productID = item.ProductID
		if productID == nil && item.ProductID != nil {
			cp := *item.ProductID
			productID = &cp
		}
		productName = item.ProductName
	} else if reportDTO.ItemID != nil && *reportDTO.ItemID > 0 {
		var cacheItem entity.ProductionOrderItem
		if err := tx.First(&cacheItem, *reportDTO.ItemID).Error; err == nil {
			moldID = cacheItem.MoldID
			externalMoldCode = cacheItem.ExternalMoldCode
			productID = cacheItem.ProductID
			productName = cacheItem.ProductName
		}
	} else {
		// 旧逻辑：从订单级别获取
		moldID = order.MoldID
		if len(order.Items) > 0 {
			firstItem := order.Items[0]
			if firstItem.ExternalMoldCode != nil && *firstItem.ExternalMoldCode != "" {
				externalMoldCode = firstItem.ExternalMoldCode
			}
			productID = firstItem.ProductID
			if productID == nil && firstItem.ProductID != nil {
				cp := *firstItem.ProductID
				productID = &cp
			}
			productName = firstItem.ProductName
		}
	}

	// 4. 生成批次号
	batchNo := s.generateBatchNo(moldID, externalMoldCode, productID)

	// 5. 创建入库记录
	stockIn := &entity.ProductionStockIn{
		StockCode:        stockCode,
		OrderID:          &order.ID,
		WarehouseID:      warehouseID,
		MoldID:           moldID,
		ExternalMoldCode: externalMoldCode,
		ProductID:        productID,
		ProductName:      productName,
		Quantity:         reportDTO.OkQty, // 只有良品数量入库
		BatchNo:          batchNo,
		StockInDate:      reportDate,
		OperatorID:       &userID,
		Remark:           "报工自动入库",
	}

	if err := tx.Create(stockIn).Error; err != nil {
		return fmt.Errorf("创建入库记录失败: %w", err)
	}

	// 6. 更新实际库存表（stock 表）
	if moldID != nil && *moldID > 0 {
		// 自有模具：确保物料记录存在，然后更新库存
		materialID, err := s.ensureMoldMaterialForReport(tx, *moldID, productID)
		if err != nil {
			logger.Warn("更新库存-确保物料记录失败",
				zap.Int64("moldID", *moldID),
				zap.Error(err))
			// 不阻断报工流程
		} else if err := s.stockService.InStockWithProductTx(tx, materialID, productID, warehouseID, float64(reportDTO.OkQty)); err != nil {
			logger.Warn("更新实际库存失败(自有模具)",
				zap.Int64("moldID", *moldID),
				zap.Int("qty", reportDTO.OkQty),
				zap.Error(err))
			// 不阻断报工流程
		}
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		// 外来模具
		materialID, err := s.ensureExternalMoldMaterialForReport(tx, *externalMoldCode, order)
		if err != nil {
			logger.Warn("更新库存-确保外来物料记录失败",
				zap.String("extMoldCode", *externalMoldCode),
				zap.Error(err))
		} else if err := s.stockService.InStockWithProductTx(tx, materialID, productID, warehouseID, float64(reportDTO.OkQty)); err != nil {
			logger.Warn("更新实际库存失败(外来模具)",
				zap.String("extMoldCode", *externalMoldCode),
				zap.Int("qty", reportDTO.OkQty),
				zap.Error(err))
		}
	}

	return nil
}

// getDefaultWarehouseID 获取默认仓库 ID（第一个正常状态仓库）
func (s *ProductionReportService) getDefaultWarehouseID(db *gorm.DB) (int64, error) {
	var wh entity.Warehouse
	err := db.Where("status = ?", 1).First(&wh).Error
	if err == nil {
		return wh.ID, nil
	}

	// 没有仓库时，自动创建一个默认仓库
	logger.Warn("未找到可用仓库，自动创建默认仓库")
	defaultWh := entity.Warehouse{
		WhCode: "DEFAULT",
		WhName: "默认仓库",
		Status: 1,
	}
	if err := db.Create(&defaultWh).Error; err != nil {
		return 0, fmt.Errorf("创建默认仓库失败: %w", err)
	}
	return defaultWh.ID, nil
}

// generateStockInCode 生成入库单号 SR-YYYYMMDD-001
func (s *ProductionReportService) generateStockInCode(db *gorm.DB) (string, error) {
	now := time.Now()
	dateStr := now.Format("20060102")
	prefix := fmt.Sprintf("SR-%s-", dateStr)

	var maxCode string
	db.Model(&entity.ProductionStockIn{}).
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

// generateBatchNo 生成批次号 SR260427-001-01
// 格式：SR + YYMMDD(6位) + 模具ID(或外来模具编号前3位) + 批次流水号(2位)
func (s *ProductionReportService) generateBatchNo(moldID *int64, externalMoldCode *string, productID *int64) string {
	// 获取当前日期 YYMMDD
	datePart := time.Now().Format("060102")

	var baseIdentifier string
	if moldID != nil && *moldID > 0 {
		// 自有模具：使用模具ID（3位，不足补0）
		baseIdentifier = fmt.Sprintf("%03d", *moldID)
	} else if externalMoldCode != nil && *externalMoldCode != "" {
		// 外来模具：使用编号（取前3位）
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
		Where("batch_no LIKE ?", fmt.Sprintf("SR%s-%s-%%", datePart, baseIdentifier)).
		Order("batch_no DESC").
		Limit(1)

	if productID != nil && *productID > 0 {
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

	return fmt.Sprintf("SR%s-%s-%02d", datePart, baseIdentifier, seq)
}

// ensureMoldMaterialForReport 报工时确保模具有对应的物料记录（mat_type=3）
func (s *ProductionReportService) ensureMoldMaterialForReport(tx *gorm.DB, moldID int64, productID *int64) (int64, error) {
	var mold entity.Mold
	if err := tx.First(&mold, moldID).Error; err != nil {
		return 0, errors.New("模具不存在")
	}

	matCode := fmt.Sprintf("%s-P%d", mold.MoldCode, func() int64 { if productID != nil { return *productID }; return 0 }())

	var material entity.Material
	err := tx.Where("mat_code = ? AND mat_type = 3", matCode).First(&material).Error
	if err == nil {
		return material.ID, nil
	}

	newMaterial := entity.Material{
		MatCode:    matCode,
		MatName:    fmt.Sprintf("%s-产品%v", mold.MoldName, productID),
		MatType:    3,
		CreateTime: time.Now(),
	}
	if err := tx.Create(&newMaterial).Error; err != nil {
		return 0, errors.New("创建物料记录失败: " + err.Error())
	}
	return newMaterial.ID, nil
}

// ensureExternalMoldMaterialForReport 外来模具的物料记录
func (s *ProductionReportService) ensureExternalMoldMaterialForReport(tx *gorm.DB, externalMoldCode string, order *entity.ProductionOrder) (int64, error) {
	matCode := fmt.Sprintf("EXT-%s", externalMoldCode)

	var material entity.Material
	err := tx.Where("mat_code = ? AND mat_type = 3", matCode).First(&material).Error
	if err == nil {
		return material.ID, nil
	}

	customerName := ""
	if order.CustomerID > 0 {
		var c entity.Customer
		if tx.Select("cus_name").First(&c, order.CustomerID).Error == nil {
			customerName = c.CusName
		}
	}

	newMaterial := entity.Material{
		MatCode:    matCode,
		MatName:    fmt.Sprintf("外来-%s(%s)", externalMoldCode, customerName),
		MatType:    3,
		CreateTime: time.Now(),
	}
	if err := tx.Create(&newMaterial).Error; err != nil {
		return 0, errors.New("创建外来物料记录失败: " + err.Error())
	}
	return newMaterial.ID, nil
}
