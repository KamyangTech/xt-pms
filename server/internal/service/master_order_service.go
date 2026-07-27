package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MasterOrderService struct {
	db *gorm.DB
}

func NewMasterOrderService(db *gorm.DB) *MasterOrderService {
	return &MasterOrderService{db: db}
}

// GenerateOrderCode 生成综合订单编号 CO-YYMMDD-01
func (s *MasterOrderService) GenerateOrderCode() (string, error) {
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD 6位日期
	prefix := fmt.Sprintf("CO-%s-", dateStr)

	// 查询当天最大序号
	var maxCode string
	s.db.Model(&entity.MasterOrder{}).
		Where("order_code LIKE ?", prefix+"%").
		Order("order_code DESC").
		Limit(1).
		Pluck("order_code", &maxCode)

	seq := 1
	if maxCode != "" {
		for i := len(maxCode) - 1; i >= 0; i-- {
			if maxCode[i] == '-' {
				seqStr := maxCode[i+1:]
				if num, err := strconv.Atoi(seqStr); err == nil {
					seq = num + 1
				}
				break
			}
		}
	}

	return fmt.Sprintf("%s%02d", prefix, seq), nil
}

// CreateCombinedOrder 创建订单（支持综合订单、仅模具、仅生产）
func (s *MasterOrderService) CreateCombinedOrder(master *entity.MasterOrder, moldOrder *entity.MoldOrder, prodOrder *entity.ProductionOrder, paymentPlan *dto.PaymentPlanDTO) (err error) {
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			// 注意：这里不能使用 logger，因为可能还未初始化
			fmt.Printf("CreateCombinedOrder panic: %v\n", r)
			tx.Rollback()
			// 将 panic 转换为错误返回
			err = fmt.Errorf("创建订单时发生错误: %v", r)
		}
	}()

	logger.Debug("开始创建综合订单",
		zap.String("order_type", master.OrderType),
		zap.Int64("customer_id", master.CustomerID))

	// 1. 生成订单号（如果未提供）
	if master.OrderCode == "" {
		orderCode, err := s.GenerateOrderCode()
		if err != nil {
			tx.Rollback()
			logger.Error("生成订单号失败", zap.Error(err))
			return fmt.Errorf("生成订单号失败: %w", err)
		}
		master.OrderCode = orderCode
	}

	// 2. 设置主订单状态
	if master.Status == "" {
		master.Status = constant.MasterOrderStatusPENDING
	}

	// 3. 创建主订单
	master.CreateTime = time.Now()
	master.UpdateTime = time.Now()
	if err := tx.Create(master).Error; err != nil {
		tx.Rollback()
		logger.Error("创建主订单失败", zap.Error(err))
		return fmt.Errorf("创建主订单失败: %w", err)
	}

	// 4. 如果提供了模具订单数据，创建模具订单及其明细
	if moldOrder != nil {
		moldOrder.MasterOrderID = &master.ID
		// 设置默认状态
		if moldOrder.Status == "" {
			moldOrder.Status = constant.MoldOrderStatusNotStarted
		}
		if moldOrder.OrderType == "" {
			moldOrder.OrderType = constant.MoldOrderTypeNew
		}
		// ✨ 确保客户 ID 与主订单一致
		moldOrder.CustomerID = master.CustomerID
		// ✨ 如果前端没有传递备注，使用主订单的备注
		if moldOrder.Remark == "" && master.Remark != "" {
			moldOrder.Remark = master.Remark
		}

		// ✨ 生成模具订单号（如果未提供）
		if moldOrder.OrderCode == "" {
			moldOrderService := NewMoldOrderService()
			orderCode, err := moldOrderService.GenerateOrderCode()
			if err != nil {
				logger.Error("生成模具订单号失败", zap.Error(err))
				tx.Rollback()
				return fmt.Errorf("生成模具订单号失败: %w", err)
			}
			moldOrder.OrderCode = orderCode
			logger.Debug("生成模具订单号", zap.String("order_code", orderCode))
		}

		moldOrder.CreateTime = time.Now()
		moldOrder.UpdateTime = time.Now()

		// ✨ 保存 Items 引用，然后清空 moldOrder.Items，避免 GORM 自动级联创建
		items := moldOrder.Items
		moldOrder.Items = nil

		if err := tx.Create(moldOrder).Error; err != nil {
			logger.Error("创建模具子订单失败", zap.Error(err))
			tx.Rollback()
			return fmt.Errorf("创建模具子订单失败: %w", err)
		}
		logger.Info("模具订单主表创建成功", zap.Int64("mold_order_id", moldOrder.ID))

		// ✨ 恢复 Items 引用
		logger.Debug("恢复 Items 引用", zap.Int("items_count", len(items)))
		moldOrder.Items = items

		// ✨ 创建模具订单明细（items）
		if len(moldOrder.Items) > 0 {
			logger.Debug("开始创建模具订单明细", zap.Int("items_count", len(moldOrder.Items)))
			for i, item := range moldOrder.Items {
				logger.Debug("处理模具订单明细",
					zap.Int("index", i),
					zap.Int64("mold_id", item.MoldID),
					zap.String("mold_code", item.MoldCode),
					zap.String("mold_name", item.MoldName),
					zap.Float64("amount", item.Amount))

				// ✨ 如果 item.MoldID 为 0，说明是新模，需要先创建模具档案
				if item.MoldID == 0 {
					logger.Info("检测到新模具，开始创建模具档案")

					// 生成模具编号（如果未提供）
					moldCode := item.MoldCode
					if moldCode == "" {
						moldCode = fmt.Sprintf("MOLD-%s-%03d", time.Now().Format("060102"), i+1)
					}

					newMold := &entity.Mold{
						MoldCode:      moldCode,
						MoldName:      item.MoldName,
						CustomerID:    master.CustomerID,
						OwnershipType: 1, // 1=自有
						MoldType:      item.MoldType,
						MoldProcess:   item.MoldProcess,
						MaterialType:  item.MaterialType,
						Remark:        item.Remark,
					}

					if err := tx.Create(newMold).Error; err != nil {
						logger.Error("创建模具档案失败", zap.Error(err))
						tx.Rollback()
						return fmt.Errorf("创建模具档案失败: %w", err)
					}

					logger.Info("模具档案创建成功",
						zap.Int64("mold_id", newMold.ID),
						zap.String("mold_code", newMold.MoldCode))

					item.MoldID = newMold.ID

					// 自动创建模具工单（新模）
					logger.Info("开始创建模具工单", zap.Int64("mold_id", newMold.ID))

					// 生成工单号（传入事务，确保序列号在事务内唯一）
					workOrderService := NewMoldWorkOrderService()
					workCode, err := workOrderService.GenerateWorkCodeWithTx(tx)
					if err != nil {
						logger.Error("生成工单号失败", zap.Error(err))
						tx.Rollback()
						return fmt.Errorf("生成工单号失败: %w", err)
					}

					var planStart, planEnd *time.Time
					if moldOrder.PlanStartDate != nil {
						t := moldOrder.PlanStartDate.Time
						planStart = &t
					}
					if moldOrder.PlanEndDate != nil {
						t := moldOrder.PlanEndDate.Time
						planEnd = &t
					}

					workType := "IN_HOUSE" // 默认内部生产
					var processorID *int64
					var outsourcePrice float64
					var outsourceTotalAmount float64

					if item.WorkType == "OUTSOURCED" {
						workType = "OUTSOURCED"
						processorID = item.ProcessorID
						outsourcePrice = item.OutsourcePrice
						outsourceTotalAmount = item.OutsourcePrice * float64(item.Amount) // 外发总金额 = 外发单价 × 金额
					}

					workOrder := &entity.MoldWorkOrder{
						WorkCode:             workCode,
						MoldOrderID:          &moldOrder.ID,
						OrderType:            constant.MoldOrderTypeNew, // 新模工单
						WorkType:             workType,                  // 工单类型：IN_HOUSE/OUTSOURCED
						ProcessorID:          processorID,               // 加工商ID（外发时填写）
						OutsourcePrice:       outsourcePrice,            // 外发单价
						OutsourceTotalAmount: outsourceTotalAmount,      // 外发总金额
						MoldID:               &newMold.ID,
						PlanStart:            planStart,
						PlanEnd:              planEnd,
						Progress:             0,
						Status:               constant.WorkOrderStatusNotStarted,
						Remark:               item.Remark,
					}

					if err := tx.Create(workOrder).Error; err != nil {
						logger.Error("创建模具工单失败", zap.Error(err))
						tx.Rollback()
						return fmt.Errorf("创建模具工单失败: %w", err)
					}

					logger.Info("模具工单创建成功",
						zap.String("work_code", workOrder.WorkCode),
						zap.Int64("work_order_id", workOrder.ID))

					// 如果是外发工单，自动创建外发记录
					if workOrder.WorkType == "OUTSOURCED" && workOrder.ProcessorID != nil {
						logger.Info("检测到外发工单，开始创建外发记录")

						outsourcingService := NewOutsourcingService()
						recordCode, err := outsourcingService.GenerateRecordCodeWithDB(tx)
						if err == nil {
							outsourcingRecord := &entity.OutsourcingRecord{
								RecordCode:         recordCode,
								OrderID:            nil, // ✨ 模具外发不需要关联生产订单
								WorkOrderID:        &workOrder.ID,
								OutsourceType:      "MOLD",
								ProcessorID:        *workOrder.ProcessorID,
								MoldID:             workOrder.MoldID,
								Quantity:           1, // 模具外发数量为1
								UnitPrice:          workOrder.OutsourcePrice,
								TotalAmount:        workOrder.OutsourceTotalAmount,
								OutsourceDate:      workOrder.PlanStart,
								ExpectedReturnDate: workOrder.PlanEnd,
								Status:             "PENDING",
								CreateUserID:       nil,
							}
							if err := tx.Create(outsourcingRecord).Error; err == nil {
								logger.Info("模具外发记录创建成功",
									zap.Int64("work_order_id", workOrder.ID),
									zap.String("record_code", recordCode))

								// ✨ 自动创建应付账款（在事务提交后执行）
								if outsourcingRecord.TotalAmount > 0 {
									apService := NewAccountsPayableService()
									apDAO := dao.NewAccountsPayableDAO()

									existingAP, _ := apDAO.FindByOrderID(outsourcingRecord.ID, "outsourcing")
								if existingAP == nil {
									now := time.Now()
									dueDate := &dto.Date{Time: now.AddDate(0, 0, 30)}

									// 模具外发 → 来源为"模具"
									ap := &entity.AccountsPayable{
										SupplierID:     *workOrder.ProcessorID,
										OrderID:        &outsourcingRecord.ID,
										OrderType:      "outsourcing",
										PaymentType:    "total",
										SourceType:     "mold",
										OrderCode:      recordCode,
										Amount:         outsourcingRecord.TotalAmount,
										DueDate:        dueDate,
											Status:         "PENDING",
											Remark:         fmt.Sprintf("模具外发记录 %s 创建时自动生成", recordCode),
										}

										if err := apService.CreatePayable(ap); err != nil {
											logger.Error("创建应付账款失败", zap.Error(err))
										} else {
											logger.Info("应付账款创建成功",
												zap.Int64("ap_id", ap.ID),
												zap.Int64("record_id", outsourcingRecord.ID),
												zap.Float64("amount", outsourcingRecord.TotalAmount))
										}
									}
								}
							} else {
								logger.Error("创建模具外发记录失败", zap.Error(err))
								// 不阻断流程，仅记录日志
							}
						}
					}

					// 为工单创建工序关联（根据启用的工序模板）
					workOrderProcessService := NewWorkOrderProcessService()
					if err := workOrderProcessService.CreateWorkOrderProcessesInTx(tx, workOrder.ID); err != nil {
						logger.Error("创建工单工序失败", zap.Error(err))
						tx.Rollback()
						return fmt.Errorf("创建工单工序失败: %w", err)
					}

					logger.Info("工单工序创建成功", zap.Int64("work_order_id", workOrder.ID))

					if err := tx.Model(&entity.Mold{}).Where("id = ?", newMold.ID).Updates(map[string]interface{}{
						"latest_order_id":      moldOrder.ID,
						"latest_work_order_id": workOrder.ID,
					}).Error; err != nil {
						logger.Error("更新模具关联字段失败", zap.Error(err))
						// 不阻断流程，仅记录日志
					}
				}

				item.OrderID = moldOrder.ID
				item.CreateTime = time.Now()
				item.UpdateTime = time.Now()
				if err := tx.Create(item).Error; err != nil {
					logger.Error("创建模具订单明细失败", zap.Error(err))
					tx.Rollback()
					return fmt.Errorf("创建模具订单明细失败: %w", err)
				}
				logger.Info("模具订单明细创建成功",
					zap.Int64("mold_id", item.MoldID),
					zap.Float64("amount", item.Amount))
			}
		} else {
			logger.Debug("模具订单没有明细项")
		}
	}

	// 5. 如果提供了生产订单数据，创建生产订单及其明细
	if prodOrder != nil {
		prodOrder.MasterOrderID = &master.ID
		// ✨ 确保客户 ID 与主订单一致
		prodOrder.CustomerID = master.CustomerID
		// 如果是综合订单，生产订单初始状态为等待模具
		if master.OrderType == "COMBINED" {
			prodOrder.DependencyStatus = "WAITING_MOLD"
		} else {
			prodOrder.DependencyStatus = "NONE"
		}
		// 设置默认状态
		if prodOrder.Status == "" {
			prodOrder.Status = constant.ProductionOrderStatusPendingSchedule
		}
		// ✨ 设置默认生产配置（如果前端未传递）
		if prodOrder.ProductionType == "" {
			prodOrder.ProductionType = "IN_HOUSE" // 默认内部生产
		}
		if prodOrder.MoldSourceType == "" {
			prodOrder.MoldSourceType = "SELF_MADE" // 默认自制模具
		}
		if prodOrder.Priority == "" {
			prodOrder.Priority = "NORMAL" // 默认普通优先级
		}
		// ✨ 如果前端没有传递计划数量，从明细中计算
		if prodOrder.PlanQty == 0 && len(prodOrder.Items) > 0 {
			for _, item := range prodOrder.Items {
				prodOrder.PlanQty += item.PlanQty
			}
		}

		// ✨ 综合订单模式下，如果生产订单没有计划日期，从模具订单继承
		if master.OrderType == "COMBINED" && moldOrder != nil {
			if prodOrder.PlanStartDate == nil && moldOrder.PlanStartDate != nil {
				if !moldOrder.PlanStartDate.IsZero() {
					t := moldOrder.PlanStartDate.Time
					prodOrder.PlanStartDate = &t
					logger.Debug("从模具订单继承计划开始日期", zap.Time("plan_start_date", t))
				}
			}
			if prodOrder.PlanEndDate == nil && moldOrder.PlanEndDate != nil {
				if !moldOrder.PlanEndDate.IsZero() {
					t := moldOrder.PlanEndDate.Time
					prodOrder.PlanEndDate = &t
					logger.Debug("从模具订单继承计划结束日期", zap.Time("plan_end_date", t))
				}
			}
		}

		// ✨ 生成生产订单号（如果未提供）
		if prodOrder.OrderCode == "" {
			prodOrderService := NewProductionOrderService()
			orderCode, err := prodOrderService.GenerateOrderCode()
			if err != nil {
				logger.Error("生成生产订单号失败", zap.Error(err))
				tx.Rollback()
				return fmt.Errorf("生成生产订单号失败: %w", err)
			}
			prodOrder.OrderCode = orderCode
			logger.Debug("生成生产订单号", zap.String("order_code", orderCode))
		}

		// ✨ 兜底：产品名称从前端明细中聚合（前端综合订单表单未传 prod_order.product_name）
		if prodOrder.ProductName == "" && len(prodOrder.Items) > 0 {
			names := make([]string, 0, len(prodOrder.Items))
			for _, it := range prodOrder.Items {
				if it.ProductName != "" {
					names = append(names, it.ProductName)
				}
			}
			switch len(names) {
			case 0:
				// 保持空
			case 1:
				prodOrder.ProductName = names[0]
			default:
				prodOrder.ProductName = fmt.Sprintf("%s 等 %d 项产品", names[0], len(names))
			}
		}

		// ✨ 兜底：外发总金额 = 外发单价 × 计划数量（前端未传 outsource_total_amount）
		if prodOrder.OutsourceTotalAmount == 0 && prodOrder.OutsourcePrice > 0 && prodOrder.PlanQty > 0 {
			prodOrder.OutsourceTotalAmount = prodOrder.OutsourcePrice * float64(prodOrder.PlanQty)
		}

		prodOrder.CreateTime = time.Now()
		prodOrder.UpdateTime = time.Now()

		// ✨ 保存 Items 引用，然后清除关联，避免 GORM 自动插入
		prodItems := prodOrder.Items
		prodOrder.Items = nil

		if err := tx.Create(prodOrder).Error; err != nil {
			tx.Rollback()
			logger.Error("创建生产子订单失败", zap.Error(err))
			return fmt.Errorf("创建生产子订单失败: %w", err)
		}

		// ✨ 自动创建外发记录（两个条件满足其一即可）：
		//   1. 生产类型显式设为外发生产 + 有加工商
		//   2. 付款计划中配置了外发应付款 + 有加工商（如用户选了外发应付款但忘了设外发类型）
		shouldCreateOutsource := false
		if prodOrder.ProductionType == "OUTSOURCED" && prodOrder.ProcessorID != nil {
			shouldCreateOutsource = true
		} else if prodOrder.ProcessorID != nil && paymentPlan != nil && paymentPlan.Enabled {
			for _, item := range paymentPlan.Items {
				if item.SourceType == "outsource" {
					shouldCreateOutsource = true
					break
				}
			}
		}

		if shouldCreateOutsource {
			// 如果有外发应付款配置但生产类型未设为外发，自动修正
			if prodOrder.ProductionType != "OUTSOURCED" {
				prodOrder.ProductionType = "OUTSOURCED"
				tx.Model(&entity.ProductionOrder{}).Where("id = ?", prodOrder.ID).Update("production_type", "OUTSOURCED")
				logger.Info("根据外发应付款配置自动将生产类型修正为外发生产",
					zap.Int64("prod_order_id", prodOrder.ID))
			}

			logger.Info("检测到外发生产订单，开始创建外发记录",
				zap.Int64("prod_order_id", prodOrder.ID),
				zap.Int64("processor_id", *prodOrder.ProcessorID))

			outsourcingService := NewOutsourcingService()
			recordCode, err := outsourcingService.GenerateRecordCodeWithDB(tx)
			if err == nil {
				outsourcingRecord := &entity.OutsourcingRecord{
					RecordCode:         recordCode,
					OrderID:            &prodOrder.ID, // ✨ 生产外发需要关联生产订单
					OutsourceType:      "PRODUCTION",
					ProcessorID:        *prodOrder.ProcessorID,
					Quantity:           prodOrder.PlanQty,
					UnitPrice:          prodOrder.OutsourcePrice,
					TotalAmount:        prodOrder.OutsourceTotalAmount,
					OutsourceDate:      prodOrder.PlanStartDate,
					ExpectedReturnDate: prodOrder.PlanEndDate,
					Status:             "PENDING",
					CreateUserID:       nil,
				}
				if err := tx.Create(outsourcingRecord).Error; err == nil {
					logger.Info("生产外发记录创建成功",
						zap.Int64("prod_order_id", prodOrder.ID),
						zap.String("record_code", recordCode))
				} else {
					logger.Error("创建生产外发记录失败", zap.Error(err))
					// 不阻断流程，仅记录日志
				}
			} else {
				logger.Error("生成外发单号失败", zap.Error(err))
			}
		}

		// ✨ 创建生产订单明细（items）
		if len(prodItems) > 0 {
			logger.Debug("开始创建生产订单明细", zap.Int("items_count", len(prodItems)))

			//  构建模具ID列表（用于综合订单模式，按顺序匹配）
			var moldIDs []int64
			if master.OrderType == "COMBINED" && moldOrder != nil && len(moldOrder.Items) > 0 {
				for _, moldItem := range moldOrder.Items {
					if moldItem.MoldID > 0 {
						moldIDs = append(moldIDs, moldItem.MoldID)
						logger.Debug("添加模具ID到列表", zap.Int64("mold_id", moldItem.MoldID))
					}
				}
			}

			// ✨ 批量预加载自有模具状态，避免 N+1（在事务内使用 tx 查询）
		var moldStatusMap map[int64]string
		if prodOrder.MoldSourceType == "SELF_MADE" {
			moldIDsToCheck := make([]int64, 0)
			for _, item := range prodItems {
				if item.MoldID != nil && *item.MoldID > 0 &&
					(item.DependencyStatus == "" || item.DependencyStatus == "NONE") {
					moldIDsToCheck = append(moldIDsToCheck, *item.MoldID)
				}
			}
			if len(moldIDsToCheck) > 0 {
				var molds []entity.Mold
				if err := tx.Where("id IN ?", moldIDsToCheck).Select("id, life_cycle_status").Find(&molds).Error; err == nil {
					moldStatusMap = make(map[int64]string, len(molds))
					for _, m := range molds {
						moldStatusMap[m.ID] = m.LifeCycleStatus
					}
				}
			}
		}

		for idx, item := range prodItems {
			item.OrderID = prodOrder.ID
			item.CreateTime = time.Now()
			item.UpdateTime = time.Now()

			// ✨ 综合订单模式下，自动绑定模具ID
			if master.OrderType == "COMBINED" && len(moldIDs) > 0 {
				// 如果 mold_id 为 0 或 nil，且 external_mold_code 也为空，则需要自动绑定模具
				if (item.MoldID == nil || *item.MoldID == 0) && (item.ExternalMoldCode == nil || *item.ExternalMoldCode == "") {
					if len(moldIDs) == 1 {
						// 只有 1 套模具：所有产品都绑定这 1 套模具
						moldID := moldIDs[0]
						item.MoldID = &moldID
						logger.Info("自动绑定模具ID（1套模具）",
							zap.Int("index", idx),
							zap.Int64("mold_id", moldID))
					} else if idx < len(moldIDs) {
						// 多套模具：按索引匹配
						moldID := moldIDs[idx]
						item.MoldID = &moldID
						logger.Info("自动匹配模具ID（多套模具）",
							zap.Int("index", idx),
							zap.Int64("mold_id", moldID))
					} else {
						logger.Warn("生产明细超出模具数量范围，需要用户手动选择模具",
							zap.Int("index", idx))
					}
				} else if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
					logger.Debug("生产订单明细使用外来模具",
						zap.String("external_mold_code", *item.ExternalMoldCode))
				}
			}

			// ✨ 设置依赖状态：根据模具来源和实际状态决定
			if item.DependencyStatus == "" || item.DependencyStatus == "NONE" {
				// 外来模具：不需要验收，直接可以生产
				if prodOrder.MoldSourceType == "CUSTOMER_PROVIDED" {
					item.DependencyStatus = "READY"
				} else if prodOrder.MoldSourceType == "SELF_MADE" && item.MoldID != nil && *item.MoldID > 0 {
					// 自有模具：从批次查询的 map 中获取状态（避免 N+1）
					if moldStatusMap != nil {
						if lifeCycleStatus, ok := moldStatusMap[*item.MoldID]; ok {
							allowedStatuses := []string{
								constant.MoldStatusTested,
								constant.MoldStatusInStock,
								constant.MoldStatusDelivered,
							}
							isReady := false
							for _, allowedStatus := range allowedStatuses {
								if lifeCycleStatus == allowedStatus {
									isReady = true
									break
								}
							}
							if isReady {
								item.DependencyStatus = "READY"
							} else {
								item.DependencyStatus = "WAITING_MOLD"
							}
						} else {
							item.DependencyStatus = "WAITING_MOLD"
						}
					} else {
						item.DependencyStatus = "WAITING_MOLD"
					}
				} else {
					item.DependencyStatus = "NONE"
				}
				logger.Debug("生产订单明细依赖状态设置",
					zap.String("dependency_status", item.DependencyStatus),
					zap.Int64("item_id", item.ID),
					zap.String("mold_source", prodOrder.MoldSourceType),
					zap.Any("mold_id", item.MoldID))
			}

				// ✨ 同步产品档案（确保 product_id 和图片写入 product 表）
				if item.ProductName != "" {
					prodOrderSvc := NewProductionOrderService()
					syncedProductID := prodOrderSvc.ensureProductRecord(tx, item.ProductName, item.MoldID, item.ProductImageURL)
					if item.ProductID == nil && syncedProductID != nil {
						item.ProductID = syncedProductID
					}
				}

				if err := tx.Create(&item).Error; err != nil {
					tx.Rollback()
					logger.Error("创建生产订单明细失败", zap.Error(err))
					return fmt.Errorf("创建生产订单明细失败: %w", err)
				}
			}
			logger.Info("生产订单明细创建完成")
		} else {
			logger.Debug("生产订单没有明细项")
		}
	}

	// 5. 汇总子订单金额到主订单（在提交事务前）
	calculatedAmount := 0.0
	if moldOrder != nil && moldOrder.TotalAmount > 0 {
		calculatedAmount += moldOrder.TotalAmount
		logger.Info("汇总模具订单金额",
			zap.Int64("mold_order_id", moldOrder.ID),
			zap.Float64("amount", moldOrder.TotalAmount))
	}
	if prodOrder != nil && prodOrder.TotalAmount > 0 {
		calculatedAmount += prodOrder.TotalAmount
		logger.Info("汇总生产订单金额",
			zap.Int64("prod_order_id", prodOrder.ID),
			zap.Float64("amount", prodOrder.TotalAmount))
	}

	// 如果前端未传递主订单金额，使用计算的子订单汇总金额或款项配置总额
	if master.TotalAmount <= 0 && calculatedAmount > 0 {
		master.TotalAmount = calculatedAmount
		logger.Info("更新主订单总金额",
			zap.Float64("total_amount", master.TotalAmount))

		if err := tx.Model(&entity.MasterOrder{}).Where("id = ?", master.ID).Update("total_amount", master.TotalAmount).Error; err != nil {
			logger.Error("更新主订单金额失败", zap.Error(err))
		}
	} else if master.TotalAmount <= 0 && paymentPlan != nil && paymentPlan.Enabled && len(paymentPlan.Items) > 0 {
		// 使用款项配置的汇总金额
		planTotal := 0.0
		for _, item := range paymentPlan.Items {
			planTotal += item.CalculatedAmount
		}
		if planTotal > 0 {
			master.TotalAmount = planTotal
			logger.Info("从款项配置汇总主订单金额",
				zap.Float64("total_amount", master.TotalAmount))
			if err := tx.Model(&entity.MasterOrder{}).Where("id = ?", master.ID).Update("total_amount", master.TotalAmount).Error; err != nil {
				logger.Error("更新主订单金额失败", zap.Error(err))
			}
		}
	}

	logger.Debug("准备提交事务", zap.String("order_code", master.OrderCode))

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		logger.Error("事务提交失败", zap.Error(err))
		return fmt.Errorf("事务提交失败: %w", err)
	}

	// 6. 自动生成应收账款和应付账款（与模具订单逻辑一致）
	var receivableItems []dto.PaymentItemDTO
	var payableItems []dto.PaymentItemDTO

	if paymentPlan != nil && paymentPlan.Enabled && len(paymentPlan.Items) > 0 {
		// 按款项来源分组：应收（mold/production）vs 应付（outsource）
		for _, item := range paymentPlan.Items {
			if item.SourceType == "outsource" {
				payableItems = append(payableItems, item)
			} else {
				receivableItems = append(receivableItems, item)
			}
		}
	} else {
		// ✨ 使用默认的单笔应收（与模具订单逻辑一致）
		if master.TotalAmount > 0 {
			receivableItems = append(receivableItems, dto.PaymentItemDTO{
				PaymentType:      "total",
				CalculatedAmount: master.TotalAmount,
				SourceType:       "production",
			})
			logger.Info("未配置自定义款项，使用默认单笔应收",
				zap.String("order_code", master.OrderCode),
				zap.Float64("total_amount", master.TotalAmount))
		}
	}

	// 6.1 创建应收账款（使用统一方法）
	if len(receivableItems) > 0 {
		if err := SyncReceivablesFromPaymentPlan(dao.GetDB(), master.CustomerID, master.ID, "master_order", master.OrderCode, receivableItems); err != nil {
			return err
		}
	}

	// 6.2 创建应付账款（外发应付款 — 支持分期）
	// 用户在前端"外发应付款"表格中可配置多条分期，每条创建一条独立的应付账款记录
	if len(payableItems) > 0 && prodOrder != nil && prodOrder.ProcessorID != nil {
		outsourcingDAO := dao.NewOutsourcingRecordDAO()
		outsourcingRecord, err := outsourcingDAO.FindByOrderID(prodOrder.ID)
		var outsourceRecordID *int64
		outsourceRecordCode := ""
		if err == nil && outsourcingRecord != nil {
			outsourceRecordID = &outsourcingRecord.ID
			outsourceRecordCode = outsourcingRecord.RecordCode
		} else {
			logger.Warn("未找到生产订单关联的外发记录",
				zap.Int64("prod_order_id", prodOrder.ID))
		}

		apService := NewAccountsPayableService()
		for _, item := range payableItems {
			var dueDate *dto.Date
			if item.DueDate != "" {
				if t, err := time.Parse("2006-01-02", item.DueDate); err == nil {
					dueDate = &dto.Date{Time: t}
				}
			}
			if dueDate == nil {
				now := time.Now()
				defaultDue := now.AddDate(0, 0, 30)
				dueDate = &dto.Date{Time: defaultDue}
			}

			if item.CalculatedAmount <= 0 {
				continue
			}

			ap := &entity.AccountsPayable{
				SupplierID:     *prodOrder.ProcessorID,
				OrderID:        outsourceRecordID,
				OrderType:      "outsourcing",
				PaymentType:    item.PaymentType,
				SourceType:     "production",
				OrderCode:      outsourceRecordCode,
				Amount:         item.CalculatedAmount,
				DueDate:        dueDate,
				Status:         "PENDING",
				Remark:         fmt.Sprintf("订单 %s 外发款项：%s", master.OrderCode, item.PaymentType),
			}

			if err := apService.CreatePayable(ap); err != nil {
				logger.Error("创建外发应付账款失败", zap.Error(err))
				return fmt.Errorf("创建外发应付账款失败: %w", err)
			}
			logger.Info("创建外发应付账款成功",
				zap.Int64("master_order_id", master.ID),
				zap.Int64("payable_id", ap.ID),
				zap.Float64("amount", item.CalculatedAmount),
				zap.String("payment_type", item.PaymentType))
		}
	}

	// 6.3 创建应付账款（模具外发应付款 — 当没有生产订单时，如 MOLD_ONLY 模式）
	if len(payableItems) > 0 && (prodOrder == nil || prodOrder.ProcessorID == nil) && moldOrder != nil {
		// 查询模具订单的外发工单和外发记录（批量查询避免 N+1）
		var workOrders []entity.MoldWorkOrder
		if err := s.db.Where("mold_order_id = ? AND work_type = 'OUTSOURCED'", moldOrder.ID).Find(&workOrders).Error; err == nil && len(workOrders) > 0 {
			var processorID int64
			var outsourceRecordID *int64
			var outsourceRecordCode string

			workOrderIDs := make([]int64, 0)
			woProcessorMap := make(map[int64]int64)
			for _, wo := range workOrders {
				if wo.ProcessorID != nil {
					workOrderIDs = append(workOrderIDs, wo.ID)
					woProcessorMap[wo.ID] = *wo.ProcessorID
				}
			}

			if len(workOrderIDs) > 0 {
				var records []entity.OutsourcingRecord
				if err := s.db.Where("work_order_id IN ?", workOrderIDs).Find(&records).Error; err == nil && len(records) > 0 {
					// 取第一个匹配的外发记录
					for _, wo := range workOrders {
						if wo.ProcessorID != nil {
							for _, record := range records {
								if record.WorkOrderID != nil && *record.WorkOrderID == wo.ID {
									processorID = *wo.ProcessorID
									outsourceRecordID = &record.ID
									outsourceRecordCode = record.RecordCode
									break
								}
							}
							if outsourceRecordID != nil {
								break
							}
						}
					}
				}
			}

			if processorID > 0 && outsourceRecordID != nil {
				apService := NewAccountsPayableService()
				for _, item := range payableItems {
					var dueDate *dto.Date
					if item.DueDate != "" {
						if t, err := time.Parse("2006-01-02", item.DueDate); err == nil {
							dueDate = &dto.Date{Time: t}
						}
					}
					if dueDate == nil {
						now := time.Now()
						defaultDue := now.AddDate(0, 0, 30)
						dueDate = &dto.Date{Time: defaultDue}
					}

					if item.CalculatedAmount <= 0 {
						continue
					}

					ap := &entity.AccountsPayable{
						SupplierID:     processorID,
						OrderID:        outsourceRecordID,
						OrderType:      "outsourcing",
						PaymentType:    item.PaymentType,
						SourceType:     "mold",
						OrderCode:      outsourceRecordCode,
						Amount:         item.CalculatedAmount,
						DueDate:        dueDate,
						Status:         "PENDING",
						Remark:         fmt.Sprintf("订单 %s 模具外发款项：%s", master.OrderCode, item.PaymentType),
					}

					if err := apService.CreatePayable(ap); err != nil {
						logger.Error("创建模具外发应付账款失败", zap.Error(err))
						return fmt.Errorf("创建模具外发应付账款失败: %w", err)
					}
					logger.Info("创建模具外发应付账款成功",
						zap.Int64("master_order_id", master.ID),
						zap.Int64("payable_id", ap.ID),
						zap.Float64("amount", item.CalculatedAmount),
						zap.String("payment_type", item.PaymentType))
				}
			} else {
				logger.Warn("未找到模具外发相关的外发记录，跳过模具外发应付账款创建",
					zap.Int64("mold_order_id", moldOrder.ID))
			}
		} else {
			logger.Warn("未找到模具订单的外发工单，跳过模具外发应付账款创建",
				zap.Int64("mold_order_id", moldOrder.ID))
		}
	}

	logger.Info("事务提交成功", zap.String("order_code", master.OrderCode))
	return nil
}

// GetMasterOrderByID 根据 ID 获取综合订单详情
func (s *MasterOrderService) GetMasterOrderByID(id int64) (*entity.MasterOrder, error) {
	var order entity.MasterOrder
	if err := s.db.Preload("Customer").
		Preload("MoldOrders.Customer").
		Preload("MoldOrders.Items.Mold").
		Preload("ProductionOrders.Customer").
		Preload("ProductionOrders.Mold").
		Preload("ProductionOrders.Items.Product").
		Preload("ProductionOrders.Items.Mold").
		Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}

	if receivables, err := NewAccountsReceivableService().GetReceivableByOrder(id, "master_order"); err == nil {
		for _, ar := range receivables {
			if ar.PaymentType == constant.PaymentTypeFirst {
				order.RemarkSummary = fmt.Sprintf("首付 %.2f / 余额 %.2f", ar.ReceivedAmount, ar.Balance)
			}
		}
		if len(receivables) > 0 {
			order.MoldOrderCount = len(order.MoldOrders)
			order.ProductionOrderCount = len(order.ProductionOrders)
		}
	}

	order.ProgressTimeline = s.calculateProgressTimeline(&order)

	return &order, nil
}

// List 获取综合订单列表
func (s *MasterOrderService) List(page, pageSize int, orderCode, customerName, createTimeStart, createTimeEnd, dueSoon string) ([]entity.MasterOrder, int64, error) {
	var orders []entity.MasterOrder
	var total int64

	offset := (page - 1) * pageSize
	db := s.db.Model(&entity.MasterOrder{})

	if orderCode != "" {
		db = db.Where("order_code LIKE ?", orderCode+"%")
	}
	if customerName != "" {
		db = db.Where("customer_id IN (SELECT id FROM customer WHERE cus_name LIKE ?)", "%"+customerName+"%")
	}
	if createTimeStart != "" {
		db = db.Where("create_time >= ?", createTimeStart)
	}
	if createTimeEnd != "" {
		db = db.Where("create_time <= ?", createTimeEnd+" 23:59:59")
	}
	if dueSoon != "" {
		if days, err := strconv.Atoi(dueSoon); err == nil && days > 0 {
			today := time.Now().Format("2006-01-02")
			endDate := time.Now().AddDate(0, 0, days).Format("2006-01-02")
			db = db.Where(`(
				id IN (SELECT master_order_id FROM mold_order WHERE plan_end_date BETWEEN ? AND ?) OR
				id IN (SELECT master_order_id FROM production_order WHERE plan_end_date BETWEEN ? AND ?)
			)`, today, endDate, today, endDate)
		}
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 主查询：获取综合订单列表（不预加载关联数据）
	if err := db.Offset(offset).Limit(pageSize).Order("id DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	if len(orders) == 0 {
		return orders, total, nil
	}

	orderIDs := make([]int64, len(orders))
	for i, order := range orders {
		orderIDs[i] = order.ID
	}

	// 批量加载 Customer（避免 N+1 查询）
	var customers []entity.Customer
	customerIDs := make(map[int64]bool)
	for _, order := range orders {
		if order.CustomerID > 0 && !customerIDs[order.CustomerID] {
			customerIDs[order.CustomerID] = true
		}
	}

	if len(customerIDs) > 0 {
		customerIDList := make([]int64, 0, len(customerIDs))
		for id := range customerIDs {
			customerIDList = append(customerIDList, id)
		}
		s.db.Where("id IN ?", customerIDList).Find(&customers)

		// 建立 customer map
		customerMap := make(map[int64]*entity.Customer)
		for i := range customers {
			customerMap[customers[i].ID] = &customers[i]
		}

		// 填充 Customer
		for i := range orders {
			if customer, ok := customerMap[orders[i].CustomerID]; ok {
				orders[i].Customer = customer
			}
		}
	}

	// 批量加载 MoldOrders（避免 N+1 查询）
	var moldOrders []entity.MoldOrder
	s.db.Where("master_order_id IN ?", orderIDs).Find(&moldOrders)

	// 按 master_order_id 分组
	moldOrderMap := make(map[int64][]entity.MoldOrder)
	for _, mo := range moldOrders {
		if mo.MasterOrderID != nil {
			moldOrderMap[*mo.MasterOrderID] = append(moldOrderMap[*mo.MasterOrderID], mo)
		}
	}

	// 填充 MoldOrders
	for i := range orders {
		if mos, ok := moldOrderMap[orders[i].ID]; ok {
			orders[i].MoldOrders = mos
		}
	}

	// 批量加载 ProductionOrders（避免 N+1 查询）
	var productionOrders []entity.ProductionOrder
	s.db.Where("master_order_id IN ?", orderIDs).Find(&productionOrders)

	// 按 master_order_id 分组
	prodOrderMap := make(map[int64][]entity.ProductionOrder)
	for _, po := range productionOrders {
		if po.MasterOrderID != nil {
			prodOrderMap[*po.MasterOrderID] = append(prodOrderMap[*po.MasterOrderID], po)
		}
	}

	// 填充 ProductionOrders
	for i := range orders {
		if pos, ok := prodOrderMap[orders[i].ID]; ok {
			orders[i].ProductionOrders = pos
		}
	}

	// 填充统计信息
	for i := range orders {
		order := &orders[i]

		// 1. 统计子订单数量
		order.MoldOrderCount = len(order.MoldOrders)
		order.ProductionOrderCount = len(order.ProductionOrders)

		// 2. 计算最早计划开始日期和最晚计划结束日期
		var earliestStart *string
		var latestEnd *string

		// 从模具订单中获取计划日期（*dto.Date 类型）
		for _, mo := range order.MoldOrders {
			if mo.PlanStartDate != nil {
				dateStr := mo.PlanStartDate.String()
				if earliestStart == nil || dateStr < *earliestStart {
					earliestStart = &dateStr
				}
			}
			if mo.PlanEndDate != nil {
				dateStr := mo.PlanEndDate.String()
				if latestEnd == nil || dateStr > *latestEnd {
					latestEnd = &dateStr
				}
			}
		}

		// 从生产订单中获取计划日期（*time.Time 类型）
		for _, po := range order.ProductionOrders {
			if po.PlanStartDate != nil {
				dateStr := po.PlanStartDate.Format("2006-01-02")
				if earliestStart == nil || dateStr < *earliestStart {
					earliestStart = &dateStr
				}
			}
			if po.PlanEndDate != nil {
				dateStr := po.PlanEndDate.Format("2006-01-02")
				if latestEnd == nil || dateStr > *latestEnd {
					latestEnd = &dateStr
				}
			}
		}

		order.PlanStartDate = earliestStart
		order.PlanEndDate = latestEnd

		// 3. 生成备注摘要（前50个字符）
		if len(order.Remark) > 50 {
			order.RemarkSummary = order.Remark[:50] + "..."
		} else {
			order.RemarkSummary = order.Remark
		}
	}

	return orders, total, nil
}

// CheckMoldReadiness 检查模具是否就绪，用于生产订单开工校验
func (s *MasterOrderService) CheckMoldReadiness(prodOrderID int64) (bool, string, error) {
	var prodOrder entity.ProductionOrder
	if err := s.db.First(&prodOrder, prodOrderID).Error; err != nil {
		return false, "", err
	}

	if prodOrder.MasterOrderID == nil {
		return true, "独立生产订单，无需检查模具依赖", nil
	}

	var moldOrder entity.MoldOrder
	err := s.db.Where("master_order_id = ? AND id != ?", prodOrder.MasterOrderID, prodOrderID).First(&moldOrder).Error
	if err != nil {
		return false, "未找到关联的模具订单", err
	}

	// 假设模具订单状态为 COMPLETED 或 DELIVERED 时视为就绪
	if moldOrder.Status == constant.MoldOrderStatusCompleted || moldOrder.Status == constant.MoldOrderStatusDelivered {
		return true, "模具已就绪", nil
	}

	return false, fmt.Sprintf("模具订单 [%s] 当前状态为 [%s]，尚未完工", moldOrder.OrderCode, moldOrder.Status), nil
}

// calculateProgressTimeline 计算订单进度时间线
func (s *MasterOrderService) calculateProgressTimeline(order *entity.MasterOrder) *entity.ProgressTimeline {
	timeline := &entity.ProgressTimeline{}

	// 1. 模具制造阶段 - 从模具订单的工单获取
	if len(order.MoldOrders) > 0 {
		moldOrder := order.MoldOrders[0]

		// 优先使用模具订单的计划时间作为备选
		if moldOrder.PlanStartDate != nil && !moldOrder.PlanStartDate.IsZero() {
			dateStr := moldOrder.PlanStartDate.Format("2006-01-02")
			timeline.MoldManufacturing = &dateStr
		}
		if moldOrder.PlanEndDate != nil && !moldOrder.PlanEndDate.IsZero() {
			dateStr := moldOrder.PlanEndDate.Format("2006-01-02")
			timeline.MoldManufacturingEnd = &dateStr
		}

		// 如果模具有工单，使用工单的实际时间覆盖（使用 JOIN 替代两次查询）
		if moldOrder.MoldID != nil {
			// 通过 JOIN 一次性获取模具的最新工单信息
			var workOrder entity.MoldWorkOrder
			if err := s.db.Table("mold_work_order").
				Select("mold_work_order.*").
				Joins("JOIN mold ON mold.latest_work_order_id = mold_work_order.id").
				Where("mold.id = ?", *moldOrder.MoldID).
				First(&workOrder).Error; err == nil {
				// 模具制造开始时间（优先使用实际时间）
				if workOrder.ActualStart != nil {
					dateStr := workOrder.ActualStart.Format("2006-01-02")
					timeline.MoldManufacturing = &dateStr
				}

				// 模具制造结束时间（工单实际结束时间）
				if workOrder.ActualEnd != nil {
					dateStr := workOrder.ActualEnd.Format("2006-01-02")
					timeline.MoldManufacturingEnd = &dateStr
				}

			} else {
				logger.Debug("查询工单失败", zap.Error(err))
			}
		} else {
		logger.Debug("模具订单没有关联模具ID", zap.Int64("mold_order_id", moldOrder.ID))
		// 新模订单没有模具ID，直接使用模具订单的更新时间作为制造时间
		if timeline.MoldManufacturing == nil {
			dateStr := moldOrder.UpdateTime.Format("2006-01-02")
			timeline.MoldManufacturing = &dateStr
		}
	}
	}

	// 2. 产品生产阶段 - 从生产订单获取
	if len(order.ProductionOrders) > 0 {
		prodOrder := order.ProductionOrders[0]
		// 产品生产开始时间
		if prodOrder.ActualStartDate != nil {
			dateStr := prodOrder.ActualStartDate.Format("2006-01-02")
			timeline.ProductProduction = &dateStr
		} else if prodOrder.PlanStartDate != nil {
			dateStr := prodOrder.PlanStartDate.Format("2006-01-02")
			timeline.ProductProduction = &dateStr
		}

		// 产品生产结束时间
		if prodOrder.ActualEndDate != nil {
			dateStr := prodOrder.ActualEndDate.Format("2006-01-02")
			timeline.ProductProductionEnd = &dateStr
		}

		// 产品入库时间 - 如果生产订单状态为已入库
		if prodOrder.Status == constant.ProductionOrderStatusStockedIn || prodOrder.Status == constant.ProductionOrderStatusShipped {
			// 可以使用生产订单的更新时间作为入库时间
			if prodOrder.ActualEndDate != nil {
				dateStr := prodOrder.ActualEndDate.Format("2006-01-02")
				timeline.StockIn = &dateStr
			} else {
				dateStr := prodOrder.UpdateTime.Format("2006-01-02")
				timeline.StockIn = &dateStr
			}
		}
	}

	return timeline
}

// EditMasterOrder 编辑综合订单基本信息（仅允许修改交货日期和备注）
func (s *MasterOrderService) EditMasterOrder(id int64, quickPlanEnd *time.Time, remark string) error {
	order, err := s.GetMasterOrderByID(id)
	if err != nil {
		return fmt.Errorf("综合订单不存在: %w", err)
	}
	if order.Status == constant.MasterOrderStatusCancelled {
		return fmt.Errorf("已取消的综合订单无法编辑")
	}

	updates := map[string]interface{}{
		"remark":      remark,
		"update_time": time.Now(),
	}
	if quickPlanEnd != nil {
		updates["quick_plan_end"] = quickPlanEnd
	}

	return s.db.Model(&entity.MasterOrder{}).Where("id = ?", id).Updates(updates).Error
}

// CancelMasterOrder 取消综合订单（级联取消关联的模具/生产订单/工单，标记无法取消的）
func (s *MasterOrderService) CancelMasterOrder(id int64, reason string, userID int64) error {
	order, err := s.GetMasterOrderByID(id)
	if err != nil {
		return fmt.Errorf("综合订单不存在: %w", err)
	}
	if order.Status == constant.MasterOrderStatusCancelled {
		return fmt.Errorf("该综合订单已取消")
	}
	if order.Status == constant.MasterOrderStatusCompleted {
		return fmt.Errorf("已完成的综合订单无法取消")
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 1. 更新综合订单本身
		if err := tx.Model(&entity.MasterOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":       constant.MasterOrderStatusCancelled,
			"cancel_reason": reason,
			"update_time":  now,
		}).Error; err != nil {
			return err
		}

		// 2. 处理关联的模具订单（批量优化：先收集所有 ID，批量查询和更新）
		{
			// 分类模具订单 ID
			var cancelMoldIDs, markMoldIDs []int64
			allMoldOrderIDs := make([]int64, 0, len(order.MoldOrders))
			for _, mo := range order.MoldOrders {
				allMoldOrderIDs = append(allMoldOrderIDs, mo.ID)
				if mo.Status == constant.MoldOrderStatusNotStarted || mo.Status == constant.MoldOrderStatusProcessing {
					cancelMoldIDs = append(cancelMoldIDs, mo.ID)
				} else {
					markMoldIDs = append(markMoldIDs, mo.ID)
				}
			}

			// 批量取消模具订单
			if len(cancelMoldIDs) > 0 {
				if err := tx.Model(&entity.MoldOrder{}).Where("id IN ?", cancelMoldIDs).Updates(map[string]interface{}{
					"status":      constant.MoldOrderStatusCancelled,
					"update_time": now,
				}).Error; err != nil {
					return err
				}
			}
			// 批量标记模具订单
			if len(markMoldIDs) > 0 {
				if err := tx.Model(&entity.MoldOrder{}).Where("id IN ?", markMoldIDs).Update("master_cancelled", 1).Error; err != nil {
					return err
				}
			}

			// 3. 批量处理模具订单下的所有工单和外发记录
			if len(allMoldOrderIDs) > 0 {
				// 一次性查询所有工单
				var allWorkOrders []entity.MoldWorkOrder
				if err := tx.Where("mold_order_id IN ?", allMoldOrderIDs).Find(&allWorkOrders).Error; err != nil {
					return err
				}

				// 分类工单
				var cancelWorkIDs, markWorkIDs []int64
				allWorkOrderIDs := make([]int64, 0, len(allWorkOrders))
				for _, wo := range allWorkOrders {
					allWorkOrderIDs = append(allWorkOrderIDs, wo.ID)
					if wo.Status == constant.WorkOrderStatusNotStarted || wo.Status == constant.WorkOrderStatusProcessing {
						cancelWorkIDs = append(cancelWorkIDs, wo.ID)
					} else {
						markWorkIDs = append(markWorkIDs, wo.ID)
					}
				}

				// 批量取消工单
				if len(cancelWorkIDs) > 0 {
					if err := tx.Model(&entity.MoldWorkOrder{}).Where("id IN ?", cancelWorkIDs).Updates(map[string]interface{}{
						"is_cancelled":     1,
						"master_cancelled": 1,
						"cancel_time":      now,
						"cancel_reason":    reason,
						"cancel_user_id":   userID,
						"update_time":      now,
					}).Error; err != nil {
						return err
					}
				}
				// 批量标记工单
				if len(markWorkIDs) > 0 {
					if err := tx.Model(&entity.MoldWorkOrder{}).Where("id IN ?", markWorkIDs).Update("master_cancelled", 1).Error; err != nil {
						return err
					}
				}

				// 4. 批量取消工单关联的外发记录
				if len(allWorkOrderIDs) > 0 {
					if err := tx.Model(&entity.OutsourcingRecord{}).
						Where("work_order_id IN ? AND status IN ?", allWorkOrderIDs, []string{"PENDING", "SENT", "PRODUCING"}).
						Update("status", "CANCELLED").Error; err != nil {
						return err
					}
				}
			}
		}

		// 5. 处理关联的生产订单（批量优化）
		{
			var cancelProdIDs, markProdIDs []int64
			for _, po := range order.ProductionOrders {
				if po.Status == constant.ProductionOrderStatusPendingSchedule || po.Status == constant.ProductionOrderStatusProducing {
					cancelProdIDs = append(cancelProdIDs, po.ID)
				} else {
					markProdIDs = append(markProdIDs, po.ID)
				}
			}

			// 批量取消生产订单
			if len(cancelProdIDs) > 0 {
				if err := tx.Model(&entity.ProductionOrder{}).Where("id IN ?", cancelProdIDs).Updates(map[string]interface{}{
					"status":      constant.ProductionOrderStatusCancelled,
					"update_time": now,
				}).Error; err != nil {
					return err
				}
				// 批量取消生产订单明细
				if err := tx.Model(&entity.ProductionOrderItem{}).Where("order_id IN ?", cancelProdIDs).Update("status", constant.ProductionOrderStatusCancelled).Error; err != nil {
					return err
				}
			}
			// 批量标记生产订单
			if len(markProdIDs) > 0 {
				if err := tx.Model(&entity.ProductionOrder{}).Where("id IN ?", markProdIDs).Update("master_cancelled", 1).Error; err != nil {
					return err
				}
			}

			// 6. 批量取消生产订单关联的外发记录
			allProdIDs := append(cancelProdIDs, markProdIDs...)
			if len(allProdIDs) > 0 {
				if err := tx.Model(&entity.OutsourcingRecord{}).
					Where("order_id IN ? AND outsource_type = 'PRODUCTION' AND status IN ?", allProdIDs, []string{"PENDING", "SENT", "PRODUCING"}).
					Update("status", "CANCELLED").Error; err != nil {
					return err
				}
			}
		}

		// 7. 处理应收账款
		var receivables []entity.AccountsReceivable
		if err := tx.Where("order_id = ? AND order_type = 'master_order'", id).Find(&receivables).Error; err != nil {
			return err
		}
		for _, ar := range receivables {
			if ar.Status == "PENDING" {
				if err := tx.Model(&entity.AccountsReceivable{}).Where("id = ?", ar.ID).Update("status", "LOCKED").Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// SyncMasterOrderStatus 根据子订单状态同步更新综合订单状态
func (s *MasterOrderService) SyncMasterOrderStatus(masterOrderID int64) error {
	var order entity.MasterOrder
	if err := s.db.Preload("MoldOrders").Preload("ProductionOrders").First(&order, masterOrderID).Error; err != nil {
		return fmt.Errorf("综合订单不存在: %w", err)
	}
	if order.Status == constant.MasterOrderStatusCancelled || order.Status == constant.MasterOrderStatusCompleted {
		return nil // 终态不更新
	}

	newStatus := s.CalculateMasterStatus(&order)
	if newStatus == "" {
		return nil
	}
	if newStatus != order.Status {
		return s.db.Model(&entity.MasterOrder{}).Where("id = ?", masterOrderID).Update("status", newStatus).Error
	}
	return nil
}

// CalculateMasterStatus 计算综合订单的当前阶段状态
func (s *MasterOrderService) CalculateMasterStatus(order *entity.MasterOrder) string {
	hasMold := len(order.MoldOrders) > 0
	hasProd := len(order.ProductionOrders) > 0

	var moldTerminal bool
	if hasMold {
		moldTerminal = true
		for _, mo := range order.MoldOrders {
			if mo.Status != constant.MoldOrderStatusTested &&
				mo.Status != constant.MoldOrderStatusCompleted &&
				mo.Status != constant.MoldOrderStatusDelivered &&
				mo.Status != constant.MoldOrderStatusCancelled {
				moldTerminal = false
				break
			}
		}
	}

	// 根据生产订单状态决定
	if hasProd {
		for _, po := range order.ProductionOrders {
			if po.Status == constant.ProductionOrderStatusCompleted {
				if !hasMold || moldTerminal {
					return constant.MasterOrderStatusCompleted
				}
			}
			if po.Status == constant.ProductionOrderStatusShipped {
				return constant.MasterOrderStatusPendingShip
			}
			if po.Status == constant.ProductionOrderStatusStockedIn {
				return constant.MasterOrderStatusPendingShip
			}
			if po.Status == constant.ProductionOrderStatusProducing {
				return constant.MasterOrderStatusProducing
			}
		}
	}

	// 根据模具订单状态决定
	if hasMold {
		for _, mo := range order.MoldOrders {
			if mo.Status != constant.MoldOrderStatusNotStarted && mo.Status != constant.MoldOrderStatusCancelled {
				return constant.MasterOrderStatusMoldProducing
			}
		}
	}

	return constant.MasterOrderStatusPENDING
}
