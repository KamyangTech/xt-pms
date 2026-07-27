package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
)

type OutsourcingService struct {
	outsourcingDAO       *dao.OutsourcingRecordDAO
	outsourcingReportDAO *dao.OutsourcingReportDAO
	orderDAO             *dao.ProductionOrderDAO
	moldDAO              *dao.MoldDAO
	workOrderDAO         *dao.MoldWorkOrderDAO
}

func NewOutsourcingService() *OutsourcingService {
	return &OutsourcingService{
		outsourcingDAO:       dao.NewOutsourcingRecordDAO(),
		outsourcingReportDAO: dao.NewOutsourcingReportDAO(),
		orderDAO:             dao.NewProductionOrderDAO(),
		moldDAO:              dao.NewMoldDAO(),
		workOrderDAO:         dao.NewMoldWorkOrderDAO(),
	}
}

// GenerateRecordCode 生成外发单号 MO-YYMMDD-01（使用默认 DB 连接）
func (s *OutsourcingService) GenerateRecordCode() (string, error) {
	return s.generateRecordCode(dao.GetDB())
}

// GenerateRecordCodeWithDB 生成外发单号 MO-YYMMDD-01（使用指定 DB 连接，支持事务内查询未提交记录）
func (s *OutsourcingService) GenerateRecordCodeWithDB(db *gorm.DB) (string, error) {
	return s.generateRecordCode(db)
}

// generateRecordCode 内部实现：根据传入的 DB 连接生成外发单号
func (s *OutsourcingService) generateRecordCode(db *gorm.DB) (string, error) {
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD 格式
	prefix := fmt.Sprintf("MO-%s-", dateStr)

	var maxCode string
	db.Model(&entity.OutsourcingRecord{}).
		Where("record_code LIKE ?", prefix+"%").
		Order("record_code DESC").
		Limit(1).
		Pluck("record_code", &maxCode)

	seq := 1
	if maxCode != "" {
		seqStr := maxCode[len(prefix):]
		fmt.Sscanf(seqStr, "%d", &seq)
		seq++
	}

	return fmt.Sprintf("%s%02d", prefix, seq), nil
}

// CreateOutsourcingRecord 创建外发记录(支持生产外发和模具外发)
func (s *OutsourcingService) CreateOutsourcingRecord(recordDTO dto.OutsourcingRecordDTO, userID int64) error {
	// 确定外发类型
	outsourceType := recordDTO.OutsourceType
	if outsourceType == "" {
		outsourceType = "PRODUCTION" // 默认为生产外发
	}

	// 根据外发类型进行不同的验证
	if outsourceType == "PRODUCTION" {
		// 生产外发验证
		return s.createProductionOutsourcing(recordDTO, userID)
	} else if outsourceType == "MOLD" {
		// 模具外发验证
		return s.createMoldOutsourcing(recordDTO, userID)
	}

	return errors.New("不支持的外发类型")
}

// createProductionOutsourcing 创建生产外发记录
func (s *OutsourcingService) createProductionOutsourcing(recordDTO dto.OutsourcingRecordDTO, userID int64) error {
	order, err := s.orderDAO.FindByID(recordDTO.OrderID)
	if err != nil {
		return errors.New("订单不存在")
	}
	if order.ProductionType != "OUTSOURCED" {
		return errors.New("该订单不是外发生产订单")
	}

	existing, _ := s.outsourcingDAO.FindByOrderID(recordDTO.OrderID)
	if existing != nil {
		return errors.New("该订单已有外发记录")
	}

	recordCode, err := s.GenerateRecordCode()
	if err != nil {
		return errors.New("生成外发单号失败")
	}

	record := &entity.OutsourcingRecord{
		RecordCode:         recordCode,
		OrderID:            &recordDTO.OrderID, // ✨ 生产外发需要关联生产订单
		WorkOrderID:        nil,
		OutsourceType:      "PRODUCTION", // 生产外发
		ProcessorID:        recordDTO.ProcessorID,
		MoldID:             recordDTO.MoldID,
		ExternalMoldInfo:   recordDTO.ExternalMoldInfo,
		OutsourceDate:      flexibleTimeToPtr(recordDTO.OutsourceDate),
		ExpectedReturnDate: flexibleTimeToPtr(recordDTO.ExpectedReturnDate),
		Quantity:           recordDTO.Quantity,
		UnitPrice:          recordDTO.UnitPrice,
		TotalAmount:        float64(recordDTO.Quantity) * recordDTO.UnitPrice,
		LogisticsCompany:   recordDTO.LogisticsCompany,
		TrackingNo:         recordDTO.TrackingNo,
		DeliveryAddress:    recordDTO.DeliveryAddress,
		ReceiverName:       recordDTO.ReceiverName,
		ReceiverPhone:      recordDTO.ReceiverPhone,
		Status:             constant.OutsourcingStatusPending, // 待外发
		CreateUserID:       &userID,
	}

	if err := s.outsourcingDAO.Create(record); err != nil {
		return errors.New("创建外发记录失败")
	}

	// ✨ 自动创建应付账款（使用统一财务服务，防止重复创建）
	if record.TotalAmount > 0 {
		now := time.Now()
		dueDate := &dto.Date{Time: now.AddDate(0, 0, 30)}

		financeService := NewFinanceService()
		paymentItems := []PaymentItem{
			{
				PaymentType: "total",
				Amount:      record.TotalAmount,
				DueDate:     dueDate,
				Remark:      fmt.Sprintf("外发记录 %s 创建时自动生成", record.RecordCode),
				SourceType:  "production",
			},
		}

		if err := financeService.CreatePayablesForOutsourcing(record.ProcessorID, record.ID, paymentItems, "production"); err != nil {
			logger.Warn("创建应付账款失败", zap.Error(err))
		}
	}

	order.Status = constant.OutsourcingStatusProducing
	s.orderDAO.Update(order)

	return nil
}

// createMoldOutsourcing 创建模具外发记录
func (s *OutsourcingService) createMoldOutsourcing(recordDTO dto.OutsourcingRecordDTO, userID int64) error {
	if recordDTO.WorkOrderID == nil {
		return errors.New("模具外发必须指定模具工单")
	}

	workOrder, err := s.workOrderDAO.FindByID(*recordDTO.WorkOrderID)
	if err != nil {
		return errors.New("模具工单不存在")
	}

	existing, _ := s.outsourcingDAO.FindByWorkOrderID(*recordDTO.WorkOrderID)
	if existing != nil {
		return errors.New("该工单已有外发记录")
	}

	recordCode, err := s.GenerateRecordCode()
	if err != nil {
		return errors.New("生成外发单号失败")
	}

	record := &entity.OutsourcingRecord{
		RecordCode:         recordCode,
		OrderID:            nil, // ✨ 模具外发不需要关联生产订单
		WorkOrderID:        recordDTO.WorkOrderID,
		OutsourceType:      "MOLD", // 模具外发
		ProcessorID:        recordDTO.ProcessorID,
		MoldID:             workOrder.MoldID,
		ExternalMoldInfo:   nil,
		OutsourceDate:      flexibleTimeToPtr(recordDTO.OutsourceDate),
		ExpectedReturnDate: flexibleTimeToPtr(recordDTO.ExpectedReturnDate),
		Quantity:           recordDTO.Quantity,
		UnitPrice:          recordDTO.UnitPrice,
		TotalAmount:        float64(recordDTO.Quantity) * recordDTO.UnitPrice,
		LogisticsCompany:   recordDTO.LogisticsCompany,
		TrackingNo:         recordDTO.TrackingNo,
		DeliveryAddress:    recordDTO.DeliveryAddress,
		ReceiverName:       recordDTO.ReceiverName,
		ReceiverPhone:      recordDTO.ReceiverPhone,
		Status:             constant.OutsourcingStatusPending, // 待外发
		CreateUserID:       &userID,
	}

	if err := s.outsourcingDAO.Create(record); err != nil {
		return errors.New("创建外发记录失败")
	}

	// ✨ 自动创建应付账款（创建外发记录时即产生应付）
	if record.TotalAmount > 0 {
		apService := NewAccountsPayableService()
		apDAO := dao.NewAccountsPayableDAO()

		existingAP, _ := apDAO.FindByOrderID(record.ID, "outsourcing")
		if existingAP == nil {
			// 计算到期日期（默认30天）
			now := time.Now()
			dueDate := &dto.Date{Time: now.AddDate(0, 0, 30)}

			// 模具外发 → 来源为"模具"
			ap := &entity.AccountsPayable{
				SupplierID:     record.ProcessorID,
				OrderID:        &record.ID,
				OrderType:      "outsourcing",
				PaymentType:    "total",
				SourceType:     "mold",
				OrderCode:      record.RecordCode,
				Amount:         record.TotalAmount,
				DueDate:        dueDate,
				Status:  "PENDING",
				Remark:  fmt.Sprintf("外发记录 %s 创建时自动生成", record.RecordCode),
			}

			if err := apService.CreatePayable(ap); err != nil {
				logger.Warn("创建应付账款失败", zap.Error(err))
			} else {
				logger.Info("应付账款创建成功",
					zap.Int64("ap_id", ap.ID),
					zap.Int64("record_id", record.ID),
					zap.Float64("amount", record.TotalAmount),
				)
			}
		}
	}

	if workOrder.MoldID != nil {
		moldUpdateData := map[string]interface{}{
			"current_location": 2, // 外发中
		}
		s.moldDAO.UpdateWithMap(*workOrder.MoldID, moldUpdateData)
	}

	return nil
}

// SendOutsourcing 确认外发
func (s *OutsourcingService) SendOutsourcing(id int64) error {
	record, err := s.outsourcingDAO.FindByID(id)
	if err != nil {
		return errors.New("外发记录不存在")
	}

	if record.Status != constant.OutsourcingStatusPending {
		return errors.New("外发记录状态不正确")
	}

	now := time.Now()
	record.Status = constant.OutsourcingStatusSent // 已外发
	record.OutsourceDate = &now

	return s.outsourcingDAO.Update(record)
}

// ReceiveOutsourcing 回收确认
func (s *OutsourcingService) ReceiveOutsourcing(id int64) error {
	record, err := s.outsourcingDAO.FindByID(id)
	if err != nil {
		return errors.New("外发记录不存在")
	}

	if record.Status != constant.OutsourcingStatusSent && record.Status != constant.OutsourcingStatusProducing {
		return errors.New("外发记录状态不正确")
	}

	now := time.Now()
	record.Status = constant.OutsourcingStatusReturned // 已返回
	record.ActualReturnDate = &now

	if err := s.outsourcingDAO.Update(record); err != nil {
		return err
	}

	// 根据外发类型进行不同的处理
	if record.OutsourceType == "PRODUCTION" {
		// 生产外发：更新订单状态为已入库
		if record.OrderID != nil {
			order, _ := s.orderDAO.FindByID(*record.OrderID)
			if order != nil {
				order.Status = "STOCKED_IN" // 已入库
				s.orderDAO.Update(order)
			}
		}
	} else if record.OutsourceType == "MOLD" {
		// 模具外发：更新模具位置为厂内
		if record.WorkOrderID != nil {
			workOrder, err := s.workOrderDAO.FindByID(*record.WorkOrderID)
			if err == nil && workOrder.MoldID != nil {
				moldUpdateData := map[string]interface{}{
					"current_location": 1, // 厂内
				}
				s.moldDAO.UpdateWithMap(*workOrder.MoldID, moldUpdateData)
			}
		}
	}

	// ✨ 新增：自动创建应付账款（使用统一财务服务，防止重复创建）
	if record.TotalAmount > 0 {
		dueDate := &dto.Date{Time: now.AddDate(0, 0, 30)}

		// 根据外发类型确定款项来源
		sourceType := "production"
		if record.OutsourceType == "MOLD" {
			sourceType = "mold"
		}

		financeService := NewFinanceService()
		paymentItems := []PaymentItem{
			{
				PaymentType: "total",
				Amount:      record.TotalAmount,
				DueDate:     dueDate,
				Remark:      fmt.Sprintf("外发记录 %s 返回自动生成", record.RecordCode),
				SourceType:  sourceType,
			},
		}

		if err := financeService.CreatePayablesForOutsourcing(record.ProcessorID, record.ID, paymentItems, sourceType); err != nil {
			// 记录日志但不阻断流程
			logger.Warn("创建应付账款失败", zap.Error(err))
		}
	}

	return nil
}

// AddOutsourcingReport 添加外发报工
func (s *OutsourcingService) AddOutsourcingReport(reportDTO dto.OutsourcingReportDTO) error {
	record, err := s.outsourcingDAO.FindByID(reportDTO.OutsourcingID)
	if err != nil {
		return errors.New("外发记录不存在")
	}

	report := &entity.OutsourcingReport{
		OutsourcingID: reportDTO.OutsourcingID,
		ReportDate:    flexibleTimeToPtr(reportDTO.ReportDate),
		OkQty:         reportDTO.OkQty,
		BadQty:        reportDTO.BadQty,
		DefectReason:  reportDTO.DefectReason,
		Photos:        reportDTO.Photos,
		Remark:        reportDTO.Remark,
	}

	if err := s.outsourcingReportDAO.Create(report); err != nil {
		return errors.New("创建外发报工记录失败")
	}

	if record.Status == constant.OutsourcingStatusSent {
		record.Status = constant.OutsourcingStatusProducing // 生产中
		s.outsourcingDAO.Update(record)
	}

	// 汇总外发报工数量并更新订单
	totalOk, totalBad, _ := s.outsourcingReportDAO.GetOutsourcingTotalQty(record.ID)
	if record.OrderID != nil {
		order, _ := s.orderDAO.FindByID(*record.OrderID)
		if order != nil {
			order.OkQty = totalOk
			order.BadQty = totalBad
			order.CompletedQty = totalOk + totalBad

			// 同步更新订单状态：如果订单处于待排产状态，有报工后应转为生产中
			if order.Status == constant.ProductionOrderStatusPendingSchedule {
				order.Status = constant.ProductionOrderStatusProducing
				now := time.Now()
				order.ActualStartDate = &now
			}

			s.orderDAO.Update(order)
		}
	}

	return nil
}

// GetOutsourcingRecords 获取外发记录列表
func (s *OutsourcingService) GetOutsourcingRecords(status *int8) ([]entity.OutsourcingRecord, error) {
	if status != nil {
		return s.outsourcingDAO.FindByStatus(*status)
	}
	return []entity.OutsourcingRecord{}, nil
}

// CreateOutsourcingRecordForWorkOrder 为模具工单创建外发记录（自动调用）
// 如果传入 tx 为 nil，则使用默认数据库连接；否则使用事务
func (s *OutsourcingService) CreateOutsourcingRecordForWorkOrder(tx *gorm.DB, workOrder *entity.MoldWorkOrder, processorID *int64, outsourcePrice *float64, userID *int64) error {
	if workOrder.ID == 0 {
		return errors.New("工单ID不能为0")
	}

	// 使用事务或默认连接
	db := tx
	if db == nil {
		db = dao.GetDB()
	}

	var existing entity.OutsourcingRecord
	err := db.Where("work_order_id = ?", workOrder.ID).First(&existing).Error
	if err == nil && existing.ID > 0 {
		if processorID != nil {
			existing.ProcessorID = *processorID
		}
		if outsourcePrice != nil {
			existing.UnitPrice = *outsourcePrice
			existing.TotalAmount = *outsourcePrice // 假设数量为1
		}
		return db.Save(&existing).Error
	}

	recordCode, err := s.GenerateRecordCode()
	if err != nil {
		return errors.New("生成外发单号失败")
	}

	var pid int64
	if processorID != nil {
		pid = *processorID
	}
	record := &entity.OutsourcingRecord{
		RecordCode:         recordCode,
		OrderID:            nil, // ✨ 模具外发不需要关联生产订单
		WorkOrderID:        &workOrder.ID,
		OutsourceType:      "MOLD",
		ProcessorID:        pid,
		MoldID:             workOrder.MoldID,
		ExternalMoldInfo:   nil,
		OutsourceDate:      workOrder.PlanStart,
		ExpectedReturnDate: workOrder.PlanEnd,
		Quantity:           1, // 模具外发数量为1
		UnitPrice:          0,
		TotalAmount:        0,
		Status:             constant.OutsourcingStatusPending,
		CreateUserID:       userID,
	}

	if outsourcePrice != nil {
		record.UnitPrice = *outsourcePrice
		record.TotalAmount = *outsourcePrice
	}

	if err := db.Create(record).Error; err != nil {
		return errors.New("创建外发记录失败")
	}

	if workOrder.MoldID != nil {
		moldUpdateData := map[string]interface{}{
			"current_location": "OUTSOURCED", // 外发中
		}
		// 使用事务或默认连接更新模具位置
		if tx != nil {
			tx.Model(&entity.Mold{}).Where("id = ?", *workOrder.MoldID).Updates(moldUpdateData)
		} else {
			s.moldDAO.UpdateWithMap(*workOrder.MoldID, moldUpdateData)
		}
	}

	return nil
}

// CancelOutsourcingRecordByWorkOrderID 根据工单ID撤回外发记录（冻结而非删除）
func (s *OutsourcingService) CancelOutsourcingRecordByWorkOrderID(workOrderID int64, reason string) error {
	record, err := s.outsourcingDAO.FindByWorkOrderID(workOrderID)
	if err != nil || record == nil {
		return nil // 没有外发记录，无需处理
	}

	// 只处理待外发或已发出的记录
	if record.Status != constant.OutsourcingStatusPending && record.Status != constant.OutsourcingStatusSent && record.Status != constant.OutsourcingStatusProducing {
		return nil // 已完成或已撤回的记录不需要处理
	}

	// 冻结外发记录（标记为已撤回）
	now := time.Now()
	record.Status = constant.OutsourcingStatusCancelled // 已撤回
	record.ActualReturnDate = &now
	if record.Remark != "" {
		record.Remark = record.Remark + " [撤回原因：" + reason + "]"
	} else {
		record.Remark = "撤回原因：" + reason
	}

	if err := s.outsourcingDAO.Update(record); err != nil {
		return errors.New("撤回外发记录失败")
	}

	// 如果模具还在外发中，恢复为在厂
	workOrder, err := s.workOrderDAO.FindByID(workOrderID)
	if err == nil && workOrder != nil && workOrder.MoldID != nil {
		mold, err := s.moldDAO.FindByID(*workOrder.MoldID)
		if err == nil && mold != nil && mold.CurrentLocation == "OUTSOURCED" {
			moldUpdateData := map[string]interface{}{
				"current_location": "IN_FACTORY", // 在厂
			}
			s.moldDAO.UpdateWithMap(*workOrder.MoldID, moldUpdateData)
		}
	}

	return nil
}

// CreateOutsourcingRecordForProductionOrder 为生产订单创建外发记录（自动调用）
func (s *OutsourcingService) CreateOutsourcingRecordForProductionOrder(order *entity.ProductionOrder, userID int64) error {
	existing, _ := s.outsourcingDAO.FindByOrderID(order.ID)
	if existing != nil {
		if order.ProcessorID != nil {
			existing.ProcessorID = *order.ProcessorID
		}
		if order.OutsourcePrice > 0 {
			existing.UnitPrice = order.OutsourcePrice
			existing.TotalAmount = order.OutsourceTotalAmount
			existing.Quantity = order.PlanQty
		}
		return s.outsourcingDAO.Update(existing)
	}

	recordCode, err := s.GenerateRecordCode()
	if err != nil {
		return errors.New("生成外发单号失败")
	}

	record := &entity.OutsourcingRecord{
		RecordCode:         recordCode,
		OrderID:            &order.ID, // ✨ 生产外发需要关联生产订单
		WorkOrderID:        nil,
		OutsourceType:      "PRODUCTION", // 生产外发
		ProcessorID:        *order.ProcessorID,
		MoldID:             nil,
		ExternalMoldInfo:   nil,
		OutsourceDate:      order.PlanStartDate,
		ExpectedReturnDate: order.PlanEndDate,
		Quantity:           order.PlanQty,
		UnitPrice:          order.OutsourcePrice,
		TotalAmount:        order.OutsourceTotalAmount,
		Status:             constant.OutsourcingStatusPending, // 待外发
		CreateUserID:       &userID,
	}

	if err := s.outsourcingDAO.Create(record); err != nil {
		return errors.New("创建外发记录失败")
	}

	return nil
}

// CancelOutsourcingRecordByProductionOrderID 根据生产订单ID撤回外发记录（冻结而非删除）
func (s *OutsourcingService) CancelOutsourcingRecordByProductionOrderID(orderID int64, reason string) error {
	record, err := s.outsourcingDAO.FindByOrderID(orderID)
	if err != nil || record == nil {
		return nil // 没有外发记录，无需处理
	}

	// 只处理待外发或已发出的记录
	if record.Status != constant.OutsourcingStatusPending && record.Status != constant.OutsourcingStatusSent && record.Status != constant.OutsourcingStatusProducing {
		return nil // 已完成或已撤回的记录不需要处理
	}

	// 冻结外发记录（标记为已撤回）
	now := time.Now()
	record.Status = constant.OutsourcingStatusCancelled // 已撤回
	record.ActualReturnDate = &now
	if record.Remark != "" {
		record.Remark = record.Remark + " [撤回原因：" + reason + "]"
	} else {
		record.Remark = "撤回原因：" + reason
	}

	if err := s.outsourcingDAO.Update(record); err != nil {
		return errors.New("撤回外发记录失败")
	}

	return nil
}

const shareTokenCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
const shareTokenLength = 8

// generateShareToken 生成随机8位分享token（crypto/rand，安全随机）
func generateShareToken() (string, error) {
	token := make([]byte, shareTokenLength)
	for i := range token {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(shareTokenCharset))))
		if err != nil {
			return "", err
		}
		token[i] = shareTokenCharset[n.Int64()]
	}
	return string(token), nil
}

// GenerateShareToken 为外发记录生成分享token（内部API）
func (s *OutsourcingService) GenerateShareToken(id int64) (string, error) {
	record, err := s.outsourcingDAO.FindByID(id)
	if err != nil {
		return "", errors.New("外发记录不存在")
	}

	// 允许已发出的记录生成分享链接
	if record.Status != constant.OutsourcingStatusSent && record.Status != constant.OutsourcingStatusProducing {
		return "", errors.New("只有已发出的外发记录才能生成分享链接")
	}

	// 如果已有 token 且需要重新生成
	token, err := generateShareToken()
	if err != nil {
		return "", errors.New("生成分享token失败")
	}

	record.ShareToken = &token
	if err := s.outsourcingDAO.Update(record); err != nil {
		return "", errors.New("保存分享token失败")
	}

	return token, nil
}

// RegenerateShareToken 重新生成分享token（旧链接失效）
func (s *OutsourcingService) RegenerateShareToken(id int64) (string, error) {
	return s.GenerateShareToken(id)
}

// GetOutsourcingByShareToken 通过分享token获取外发记录（公开API）
func (s *OutsourcingService) GetOutsourcingByShareToken(token string) (*entity.OutsourcingRecord, error) {
	if token == "" {
		return nil, errors.New("无效的分享链接")
	}
	record, err := s.outsourcingDAO.FindByShareToken(token)
	if err != nil {
		return nil, errors.New("分享链接无效或已过期")
	}
	return record, nil
}

// AddPublicReport 通过分享token添加外发报工（公开API，加工商使用）
func (s *OutsourcingService) AddPublicReport(token string, reportDTO dto.OutsourcingPublicReportDTO) error {
	// 通过 token 查找外发记录
	record, err := s.outsourcingDAO.FindByShareToken(token)
	if err != nil {
		return errors.New("分享链接无效或已过期")
	}

	// 只允许已发出或生产中的记录报工
	if record.Status != constant.OutsourcingStatusSent && record.Status != constant.OutsourcingStatusProducing {
		return errors.New("当前状态不允许报工")
	}

	// 数量默认为0
	okQty := 0
	badQty := 0
	if reportDTO.OkQty != nil {
		okQty = *reportDTO.OkQty
	}
	if reportDTO.BadQty != nil {
		badQty = *reportDTO.BadQty
	}

	// 报工日期默认使用当前时间
	reportDate := reportDTO.ReportDate
	if reportDate == nil {
		now := time.Now()
		reportDate = &dto.FlexibleTime{Time: now}
	}

	internalDTO := dto.OutsourcingReportDTO{
		OutsourcingID: record.ID,
		ReportDate:    reportDate,
		OkQty:         okQty,
		BadQty:        badQty,
		DefectReason:  reportDTO.DefectReason,
		Photos:        reportDTO.Photos,
		Remark:        reportDTO.Remark,
	}

	return s.AddOutsourcingReport(internalDTO)
}

// flexibleTimeToPtr 将 *dto.FlexibleTime 转换为 *time.Time
func flexibleTimeToPtr(ft *dto.FlexibleTime) *time.Time {
	if ft == nil {
		return nil
	}
	return &ft.Time
}
