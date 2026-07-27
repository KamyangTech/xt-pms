package dao

import (
	"time"

	"go.uber.org/zap"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"gorm.io/gorm"
)

type AccountsPayableDAO struct{}

// NewAccountsPayableDAO 创建应付账款DAO实例
func NewAccountsPayableDAO() *AccountsPayableDAO {
	return &AccountsPayableDAO{}
}

// Create 创建应付账款记录
func (d *AccountsPayableDAO) Create(ap *entity.AccountsPayable) error {
	return DB.Create(ap).Error
}

// FindByID 根据ID查询应付账款
func (d *AccountsPayableDAO) FindByID(id int64) (*entity.AccountsPayable, error) {
	var ap entity.AccountsPayable
	err := DB.Preload("Supplier").Preload("Processor").First(&ap, id).Error
	if err != nil {
		return nil, err
	}
	return &ap, nil
}

// FindList 分页查询应付账款列表
func (d *AccountsPayableDAO) FindList(page, pageSize int, supplierID *int64, status string) ([]*entity.AccountsPayable, int64, error) {
	var aps []*entity.AccountsPayable
	var total int64

	query := DB.Model(&entity.AccountsPayable{}).Preload("Supplier").Preload("Processor")

	// 应用过滤条件
	if supplierID != nil && *supplierID > 0 {
		query = query.Where("supplier_id = ?", *supplierID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&aps).Error

	return aps, total, err
}

// FindByOrderID 根据订单ID查询应付账款
func (d *AccountsPayableDAO) FindByOrderID(orderID int64, orderType string) (*entity.AccountsPayable, error) {
	var ap entity.AccountsPayable
	err := DB.Where("order_id = ? AND order_type = ?", orderID, orderType).First(&ap).Error
	if err != nil {
		return nil, err
	}
	return &ap, nil
}

// Update 更新应付账款
func (d *AccountsPayableDAO) Update(ap *entity.AccountsPayable) error {
	return DB.Save(ap).Error
}

// Delete 删除应付账款
func (d *AccountsPayableDAO) Delete(id int64) error {
	return DB.Delete(&entity.AccountsPayable{}, id).Error
}

// GetOverdueList 获取逾期应付账款列表
func (d *AccountsPayableDAO) GetOverdueList() ([]*entity.AccountsPayable, error) {
	var aps []*entity.AccountsPayable
	err := DB.Preload("Supplier").Preload("Processor").
		Where("status != ? AND due_date < CURDATE()", "PAID").
		Order("due_date ASC").
		Find(&aps).Error
	return aps, err
}

// Pay 付款（增加已付金额）
func (d *AccountsPayableDAO) Pay(id int64, amount float64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var ap entity.AccountsPayable
		if err := tx.First(&ap, id).Error; err != nil {
			return err
		}

		ap.PaidAmount += amount

		if ap.PaidAmount >= ap.Amount {
			ap.Status = "PAID"
		} else if ap.PaidAmount > 0 {
			ap.Status = "PARTIAL"
		}

		return tx.Save(&ap).Error
	})
}

// PayWithSync 付款并同步到外发记录
func (d *AccountsPayableDAO) PayWithSync(id int64, amount float64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var ap entity.AccountsPayable
		if err := tx.First(&ap, id).Error; err != nil {
			return err
		}

		ap.PaidAmount += amount
		if ap.PaidAmount >= ap.Amount {
			ap.Status = "PAID"
		} else if ap.PaidAmount > 0 {
			ap.Status = "PARTIAL"
		}
		if err := tx.Save(&ap).Error; err != nil {
			return err
		}

		// ✨ 同步更新外发记录
		if ap.OrderType == "outsourcing" && ap.OrderID != nil {
			var outsourcingRecord entity.OutsourcingRecord
			if err := tx.First(&outsourcingRecord, *ap.OrderID).Error; err == nil {
				outsourcingRecord.PaidAmount += amount
				if err := tx.Save(&outsourcingRecord).Error; err != nil {
					// 记录日志但不阻断流程
					logger.Warn("同步外发记录付款金额失败", zap.Error(err))
				}
			}
		}

		return nil
	})
}

// PayWithDetails 付款并创建详细记录
func (d *AccountsPayableDAO) PayWithDetails(id int64, amount float64, discountAmount float64, paymentDate time.Time, paymentMethod, transactionNo, paymentAccount, receiverAccount, remark, receiptFileID, invoiceFileID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var ap entity.AccountsPayable
		if err := tx.First(&ap, id).Error; err != nil {
			return err
		}

		// 1. 创建付款记录
		payment := &entity.PayablePayment{
			PayableID:       id,
			PaymentAmount:   amount,
			DiscountAmount:  discountAmount,
			PaymentDate:     paymentDate,
			PaymentMethod:   paymentMethod,
			TransactionNo:   transactionNo,
			PaymentAccount:  paymentAccount,
			ReceiverAccount: receiverAccount,
			Remark:          remark,
			ReceiptFileID:   receiptFileID,
			InvoiceFileID:   invoiceFileID,
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// 2. 更新应付账款（balance 为数据库生成列，自动计算 amount - paid_amount - discount_amount）
		ap.PaidAmount += amount
		ap.DiscountAmount += discountAmount
		if ap.PaidAmount+ap.DiscountAmount >= ap.Amount {
			ap.Status = "PAID"
		} else if ap.PaidAmount+ap.DiscountAmount > 0 {
			ap.Status = "PARTIAL"
		}
		if err := tx.Save(&ap).Error; err != nil {
			return err
		}

		// 3. 同步更新外发记录
		if ap.OrderType == "outsourcing" && ap.OrderID != nil {
			var outsourcingRecord entity.OutsourcingRecord
			if err := tx.First(&outsourcingRecord, *ap.OrderID).Error; err == nil {
				outsourcingRecord.PaidAmount += amount
				if err := tx.Save(&outsourcingRecord).Error; err != nil {
					logger.Warn("同步外发记录付款金额失败", zap.Error(err))
				}
			}
		}

		return nil
	})
}
