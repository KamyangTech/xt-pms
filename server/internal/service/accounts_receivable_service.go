package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
)

type AccountsReceivableService struct {
	arDAO *dao.AccountsReceivableDAO
}

func NewAccountsReceivableService() *AccountsReceivableService {
	return &AccountsReceivableService{
		arDAO: dao.NewAccountsReceivableDAO(),
	}
}

// CreateReceivable 创建应收账款
func (s *AccountsReceivableService) CreateReceivable(ar *entity.AccountsReceivable) error {
	// 参数验证
	if ar.CustomerID == 0 {
		return fmt.Errorf("客户ID不能为空")
	}
	if ar.Amount <= 0 {
		return fmt.Errorf("应收金额必须大于0")
	}
	if ar.Status == "" {
		ar.Status = "PENDING"
	}

	s.calculateAgingCategory(ar)

	return s.arDAO.Create(ar)
}

// GetReceivableByID 根据ID获取应收账款
func (s *AccountsReceivableService) GetReceivableByID(id int64) (*entity.AccountsReceivable, error) {
	ar, err := s.arDAO.FindByID(id)
	if err != nil {
		return nil, err
	}

	// ✨ 如果是模具订单，附加分期收款信息
	if ar.OrderType == "mold_order" && ar.OrderID != nil {
		paymentService := NewMoldOrderPaymentService()
		payments, _ := paymentService.GetPaymentsByOrderID(*ar.OrderID)
		// 将分期信息存储到Remark中（或者可以扩展实体字段）
		if len(payments) > 0 {
			var paymentInfo string
			for _, p := range payments {
				statusText := "未收"
				if p.Status == 2 {
					statusText = "部分收款"
				} else if p.Status == 3 {
					statusText = "已收款"
				}
				paymentInfo += fmt.Sprintf("%s: %.2f/%.2f (%s); ", p.PaymentType, p.ActualAmount, p.ShouldAmount, statusText)
			}
			ar.Remark += "\n分期收款: " + paymentInfo
		}
	}

	if ar.IsLocked && ar.Status == "" {
		ar.Status = "locked"
	}

	return ar, nil
}

// GetReceivableList 获取应收账款列表
// sourceTypes / paymentTypes 为多选过滤条件，空切片表示不过滤
func (s *AccountsReceivableService) GetReceivableList(page, pageSize int, customerID *int64, status string, sourceTypes, paymentTypes []string) ([]*entity.AccountsReceivable, int64, error) {
	list, total, err := s.arDAO.FindListWithOrderInfo(page, pageSize, customerID, status, sourceTypes, paymentTypes)
	if err != nil {
		return nil, 0, err
	}

	for _, ar := range list {
		s.updateAgingInfo(ar)
	}

	return list, total, nil
}

// GetReceivableByOrder 根据订单获取应收账款
func (s *AccountsReceivableService) GetReceivableByOrder(orderID int64, orderType string) ([]*entity.AccountsReceivable, error) {
	return s.arDAO.FindByOrderID(orderID, orderType)
}

// UpdateReceivable 更新应收账款（仅未收款的 pending 状态允许编辑）
func (s *AccountsReceivableService) UpdateReceivable(ar *entity.AccountsReceivable) error {
	if ar.ID == 0 {
		return fmt.Errorf("应收账款ID不能为空")
	}

	existing, err := s.arDAO.FindByID(ar.ID)
	if err != nil {
		return fmt.Errorf("应收账款不存在")
	}

	// 仅允许编辑 PENDING 状态的记录
	if existing.Status != "PENDING" {
		return fmt.Errorf("当前状态为 %s，不允许编辑金额，请通过追加应收方式处理", existing.Status)
	}

	// 保留不可编辑的字段
	ar.Status = existing.Status
	ar.ReceivedAmount = existing.ReceivedAmount
	ar.DiscountAmount = existing.DiscountAmount
	ar.IsLocked = existing.IsLocked
	ar.CreateTime = existing.CreateTime

	// 重新计算账龄分类
	s.calculateAgingCategory(ar)

	return s.arDAO.Update(ar)
}

// DeleteReceivable 删除应收账款
func (s *AccountsReceivableService) DeleteReceivable(id int64) error {
	return s.arDAO.Delete(id)
}

// GetPaymentsByReceivableID 获取应收账款的收款记录列表
func (s *AccountsReceivableService) GetPaymentsByReceivableID(receivableID int64) ([]*entity.ReceivablePayment, error) {
	paymentDAO := dao.NewReceivablePaymentDAO()

	payments, err := paymentDAO.FindByReceivableID(receivableID)
	if err != nil {
		return nil, err
	}

	// 批量查询所有回单和发票文件（优化 N+1）
	allFileIDs := make(map[string]struct{})
	for _, p := range payments {
		if p.ReceiptFileID != "" {
			for _, id := range strings.Split(p.ReceiptFileID, ",") {
				if tid := strings.TrimSpace(id); tid != "" {
					allFileIDs[tid] = struct{}{}
				}
			}
		}
		if p.InvoiceFileID != "" {
			for _, id := range strings.Split(p.InvoiceFileID, ",") {
				if tid := strings.TrimSpace(id); tid != "" {
					allFileIDs[tid] = struct{}{}
				}
			}
		}
	}

	if len(allFileIDs) > 0 {
		ids := make([]string, 0, len(allFileIDs))
		for id := range allFileIDs {
			ids = append(ids, id)
		}

		fileDAO := dao.NewFileDAO()
		fileMap, err := fileDAO.FindByIDsMap(ids)
		if err == nil {
			for _, p := range payments {
				if p.ReceiptFileID != "" {
					p.ReceiptFiles = loadFilesByMap(p.ReceiptFileID, fileMap)
				}
				if p.InvoiceFileID != "" {
					p.InvoiceFiles = loadFilesByMap(p.InvoiceFileID, fileMap)
				}
			}
		}
	}

	return payments, nil
}

// loadFilesByMap 根据逗号分隔的文件ID列表，从已加载的 map 中批量取文件
func loadFilesByMap(fileIDs string, fileMap map[string]*entity.File) []*entity.File {
	ids := strings.Split(fileIDs, ",")
	files := make([]*entity.File, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if f, ok := fileMap[id]; ok {
			files = append(files, f)
		}
	}
	return files
}

// loadFiles 逐条查询文件（保留兼容，新代码请使用 GetPaymentsByReceivableID）
func loadFiles(fileIDs string) []*entity.File {
	fileDAO := dao.NewFileDAO()
	ids := strings.Split(fileIDs, ",")
	var files []*entity.File
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if file, err := fileDAO.FindByID(id); err == nil {
			files = append(files, file)
		}
	}
	return files
}

// UpdateReceivablePayment 更新收款记录（仅允许编辑非金额字段）
func (s *AccountsReceivableService) UpdateReceivablePayment(payment *entity.ReceivablePayment) error {
	if payment.ID == 0 {
		return fmt.Errorf("收款记录ID不能为空")
	}

	paymentDAO := dao.NewReceivablePaymentDAO()
	existing, err := paymentDAO.FindByID(payment.ID)
	if err != nil {
		return fmt.Errorf("收款记录不存在")
	}

	// 仅允许更新以下字段（禁止修改收款金额）
	existing.PaymentDate = payment.PaymentDate
	existing.PaymentMethod = payment.PaymentMethod
	existing.TransactionNo = payment.TransactionNo
	existing.ReceivingAccount = payment.ReceivingAccount
	existing.Remark = payment.Remark
	existing.ReceiptFileID = payment.ReceiptFileID
	existing.InvoiceFileID = payment.InvoiceFileID

	return paymentDAO.Update(existing)
}

// ReceivePayment 收款（支持优惠减免）
func (s *AccountsReceivableService) ReceivePayment(id int64, amount float64, discountAmount float64, paymentDate time.Time, paymentMethod, transactionNo, receivingAccount, remark, receiptFileID, invoiceFileID string) error {
	if amount <= 0 && discountAmount <= 0 {
		return fmt.Errorf("收款金额和优惠金额不能同时为空")
	}
	if amount < 0 || discountAmount < 0 {
		return fmt.Errorf("收款金额和优惠金额不能为负数")
	}

	// 先查询当前记录，检查余额
	ar, err := s.arDAO.FindByID(id)
	if err != nil {
		return err
	}

	if ar.IsLocked {
		return fmt.Errorf("该应收账款已锁定，暂不可收款")
	}
	if PaymentTypeNeedsDeliveryUnlock(ar.PaymentType) && ar.Status == "locked" {
		return fmt.Errorf("该款项未解锁（需交付后才能解锁），暂不可收款")
	}
	if amount+discountAmount > ar.Balance {
		return fmt.Errorf("收款金额+优惠金额(%.2f)不能超过余额 %.2f", amount+discountAmount, ar.Balance)
	}

	return s.arDAO.ReceivePaymentWithDetails(id, amount, discountAmount, paymentDate, paymentMethod, transactionNo, receivingAccount, remark, receiptFileID, invoiceFileID)
}

// UnlockFinalReceivablesByMasterOrder 解锁综合订单的尾款应收
func (s *AccountsReceivableService) UnlockFinalReceivablesByMasterOrder(masterOrderID int64) error {
	ars, err := s.arDAO.FindByOrderID(masterOrderID, "master_order")
	if err != nil {
		return err
	}
	for _, ar := range ars {
		if PaymentTypeNeedsDeliveryUnlock(ar.PaymentType) {
			ar.IsLocked = false
			ar.UnlockRule = ""
			if ar.Status == "locked" {
				ar.Status = "PENDING"
			}
			if err := s.arDAO.Update(ar); err != nil {
				return err
			}
		}
	}
	return nil
}

// ProcessOverdueReceivables 处理逾期应收账款（可定时任务调用）
func (s *AccountsReceivableService) ProcessOverdueReceivables() error {
	overdueList, err := s.arDAO.GetOverdueList()
	if err != nil {
		return err
	}

	updatedCount := 0
	for _, ar := range overdueList {
		s.calculateAgingCategory(ar)

		// 如果状态不是已收款且已逾期，更新为逾期状态
		if ar.AgingDays > 0 && ar.Status != "PAID" && ar.Status != "OVERDUE" {
			oldStatus := ar.Status
			ar.Status = "OVERDUE"

			if err := s.arDAO.Update(ar); err != nil {
				logger.Error("更新应收账款逾期状态失败",
					zap.Int64("id", ar.ID),
					zap.String("old_status", oldStatus),
					zap.Error(err),
				)
				continue
			}

			logger.Info("应收账款逾期状态更新",
				zap.Int64("id", ar.ID),
				zap.String("old_status", oldStatus),
				zap.Int("aging_days", ar.AgingDays),
			)
			updatedCount++
		}
	}

	logger.Info("逾期应收账款处理完成",
		zap.Int("updated_count", updatedCount),
	)
	return nil
}

// GetOverdueList 获取逾期应收账款列表
func (s *AccountsReceivableService) GetOverdueList() ([]*entity.AccountsReceivable, error) {
	list, err := s.arDAO.GetOverdueList()
	if err != nil {
		return nil, err
	}

	for _, ar := range list {
		s.updateAgingInfo(ar)
	}

	return list, nil
}

// GetAgingStatistics 获取账龄统计
func (s *AccountsReceivableService) GetAgingStatistics() (map[string]interface{}, error) {
	return s.arDAO.GetAgingStatistics()
}

// calculateAgingCategory 计算账龄分类
func (s *AccountsReceivableService) calculateAgingCategory(ar *entity.AccountsReceivable) {
	if ar.DueDate == nil {
		ar.AgingCategory = "CURRENT"
		ar.AgingDays = 0
		return
	}

	now := time.Now()
	dueDate := ar.DueDate.Time

	diff := now.Sub(dueDate)
	days := int(math.Ceil(diff.Hours() / 24))

	if days < 0 {
		// 未到期
		ar.AgingDays = 0
		ar.AgingCategory = "CURRENT"
	} else if days <= 30 {
		ar.AgingDays = days
		ar.AgingCategory = "WITHIN_30"
	} else if days <= 60 {
		ar.AgingDays = days
		ar.AgingCategory = "WITHIN_60"
	} else if days <= 90 {
		ar.AgingDays = days
		ar.AgingCategory = "WITHIN_90"
	} else {
		ar.AgingDays = days
		ar.AgingCategory = "OVERDUE"
	}

	// 如果已逾期且状态不是已收款，更新状态为逾期
	if days > 0 && ar.Status != "PAID" {
		ar.Status = "OVERDUE"
	}
}

// updateAgingInfo 更新账龄信息（不保存到数据库）
func (s *AccountsReceivableService) updateAgingInfo(ar *entity.AccountsReceivable) {
	s.calculateAgingCategory(ar)
}

// parseDueDate 解析到期日期，默认30天后
func parseDueDate(dueDateStr string) *dto.Date {
	if dueDateStr != "" {
		if t, err := time.Parse("2006-01-02", dueDateStr); err == nil {
			return &dto.Date{Time: t}
		}
	}
	now := time.Now()
	defaultDue := now.AddDate(0, 0, 30)
	return &dto.Date{Time: defaultDue}
}

// SyncReceivablesFromPaymentPlan 根据付款计划同步创建应收账款（统一方法，供模具订单和综合订单复用）
// db: 数据库连接（支持事务 tx 或普通 DB）
// customerID: 客户ID
// orderID: 关联订单ID
// orderType: 订单类型（mold_order / master_order）
// orderCode: 订单编号
// items: 付款计划中的应收款项（非 outsource）
func SyncReceivablesFromPaymentPlan(db *gorm.DB, customerID int64, orderID int64, orderType string, orderCode string, items []dto.PaymentItemDTO) error {
	if len(items) == 0 {
		return nil
	}

	// 检查是否已有应收账款，避免重复创建
	var existingCount int64
	if err := db.Model(&entity.AccountsReceivable{}).
		Where("order_id = ? AND order_type = ?", orderID, orderType).
		Count(&existingCount).Error; err == nil && existingCount > 0 {
		logger.Info("已存在应收账款，跳过自动创建",
			zap.Int64("order_id", orderID),
			zap.String("order_type", orderType))
		return nil
	}

	for _, item := range items {
		if item.CalculatedAmount <= 0 {
			continue
		}

		sourceType := item.SourceType
		if sourceType == "" {
			sourceType = "mold"
		}

		ar := &entity.AccountsReceivable{
			CustomerID: customerID,
			OrderID:    &orderID,
			OrderType:  orderType,
			PaymentType: item.PaymentType,
			SourceType:  sourceType,
			Amount:      item.CalculatedAmount,
			DueDate:     parseDueDate(item.DueDate),
			Status:      "PENDING",
			Remark:      fmt.Sprintf("订单 %s 自定义款项：%s", orderCode, item.PaymentType),
		}

		if err := db.Create(ar).Error; err != nil {
			return fmt.Errorf("创建应收账款失败(payment_type=%s): %w", item.PaymentType, err)
		}

		logger.Info("创建应收账款成功",
			zap.Int64("order_id", orderID),
			zap.String("order_code", orderCode),
			zap.Float64("amount", item.CalculatedAmount),
			zap.String("payment_type", item.PaymentType),
		)
	}

	return nil
}
