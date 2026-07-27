package dao

import (
	"time"

	"go.uber.org/zap"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"gorm.io/gorm"
)

type AccountsReceivableDAO struct{}

// NewAccountsReceivableDAO 创建应收账款DAO实例
func NewAccountsReceivableDAO() *AccountsReceivableDAO {
	return &AccountsReceivableDAO{}
}

// Create 创建应收账款记录
func (d *AccountsReceivableDAO) Create(ar *entity.AccountsReceivable) error {
	return DB.Create(ar).Error
}

// FindByID 根据ID查询应收账款
func (d *AccountsReceivableDAO) FindByID(id int64) (*entity.AccountsReceivable, error) {
	var ar entity.AccountsReceivable
	if err := DB.First(&ar, id).Error; err != nil {
		return nil, err
	}
	if err := d.attachCustomer(&ar); err != nil {
		return nil, err
	}
	// 补充订单编号等关联信息
	if err := d.attachOrderInfo(&ar); err != nil {
		return nil, err
	}
	return &ar, nil
}

// FindList 分页查询应收账款列表
// sourceTypes / paymentTypes 为多选过滤条件，空切片表示不过滤
func (d *AccountsReceivableDAO) FindList(page, pageSize int, customerID *int64, status string, sourceTypes, paymentTypes []string) ([]*entity.AccountsReceivable, int64, error) {
	var ars []*entity.AccountsReceivable
	var total int64

	query := DB.Model(&entity.AccountsReceivable{})

	// 应用过滤条件
	if customerID != nil && *customerID > 0 {
		query = query.Where("customer_id = ?", *customerID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if len(sourceTypes) > 0 {
		// 多选款项来源（mold=模具, production=生产）
		query = query.Where("source_type IN ?", sourceTypes)
	}
	if len(paymentTypes) > 0 {
		// 多选款项类型（total=全款, deposit=首款, interim=中期款, final=尾款）
		query = query.Where("payment_type IN ?", paymentTypes)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&ars).Error
	if err != nil {
		return nil, 0, err
	}

	if err := d.attachCustomers(ars); err != nil {
		return nil, 0, err
	}

	return ars, total, nil
}

// FindByOrderID 根据订单ID查询应收账款
func (d *AccountsReceivableDAO) FindByOrderID(orderID int64, orderType string) ([]*entity.AccountsReceivable, error) {
	var ars []*entity.AccountsReceivable
	err := DB.Where("order_id = ? AND order_type = ?", orderID, orderType).Order("id ASC").Find(&ars).Error
	if err != nil {
		return nil, err
	}
	return ars, nil
}

// FindListWithOrderInfo 分页查询并补充订单关联信息（优化版：批量查询避免N+1问题）
// sourceTypes / paymentTypes 为多选过滤条件，空切片表示不过滤
func (d *AccountsReceivableDAO) FindListWithOrderInfo(page, pageSize int, customerID *int64, status string, sourceTypes, paymentTypes []string) ([]*entity.AccountsReceivable, int64, error) {
	ars, total, err := d.FindList(page, pageSize, customerID, status, sourceTypes, paymentTypes)
	if err != nil {
		return nil, 0, err
	}

	if err := d.attachCustomers(ars); err != nil {
		return nil, 0, err
	}

	// 优化：批量查询所有订单信息，避免N+1查询
	if err := d.attachOrderInfoInBatch(ars); err != nil {
		return nil, 0, err
	}

	return ars, total, nil
}

// attachOrderInfo 补充单条应收账款的订单信息
func (d *AccountsReceivableDAO) attachOrderInfo(ar *entity.AccountsReceivable) error {
	if ar.OrderID == nil {
		return nil
	}

	switch ar.OrderType {
	case "master_order":
		var mo entity.MasterOrder
		if err := DB.Preload("MoldOrders.Items.Mold").
			Preload("ProductionOrders.Mold").
			Preload("ProductionOrders.Items.Mold").
			Preload("ProductionOrders.Items.Product").
			First(&mo, *ar.OrderID).Error; err != nil {
			return err
		}
		d.fillMasterOrderInfo(ar, &mo)
	case "mold_order":
		var mo entity.MoldOrder
		if err := DB.Preload("Items.Mold").
			First(&mo, *ar.OrderID).Error; err != nil {
			return err
		}
		d.fillMoldOrderInfo(ar, &mo)
	case "production_order":
		var po entity.ProductionOrder
		if err := DB.Preload("Mold").
			Preload("Items.Mold").
			First(&po, *ar.OrderID).Error; err != nil {
			return err
		}
		d.fillProductionOrderInfo(ar, &po)
	}

	return nil
}

// attachOrderInfoInBatch 批量补充订单信息（优化N+1查询问题）
func (d *AccountsReceivableDAO) attachOrderInfoInBatch(ars []*entity.AccountsReceivable) error {
	if len(ars) == 0 {
		return nil
	}

	// 1. 收集所有需要查询的订单ID
	masterOrderIDs := make([]int64, 0)
	moldOrderIDs := make([]int64, 0)
	productionOrderIDs := make([]int64, 0)

	for _, ar := range ars {
		if ar.OrderID == nil {
			continue
		}

		switch ar.OrderType {
		case "master_order":
			masterOrderIDs = append(masterOrderIDs, *ar.OrderID)
		case "mold_order":
			moldOrderIDs = append(moldOrderIDs, *ar.OrderID)
		case "production_order":
			productionOrderIDs = append(productionOrderIDs, *ar.OrderID)
		}
	}

	// 2. 批量查询主订单信息
	masterOrderMap := make(map[int64]*entity.MasterOrder)
	if len(masterOrderIDs) > 0 {
		var masterOrders []*entity.MasterOrder
		if err := DB.Preload("MoldOrders.Items.Mold").
			Preload("ProductionOrders.Mold").
			Preload("ProductionOrders.Items.Mold").
			Preload("ProductionOrders.Items.Product").
			Where("id IN ?", masterOrderIDs).Find(&masterOrders).Error; err == nil {
			for _, mo := range masterOrders {
				masterOrderMap[mo.ID] = mo
			}
		}
	}

	// 3. 批量查询模具订单信息
	moldOrderMap := make(map[int64]*entity.MoldOrder)
	if len(moldOrderIDs) > 0 {
		var moldOrders []*entity.MoldOrder
		if err := DB.Preload("Items.Mold").
			Where("id IN ?", moldOrderIDs).Find(&moldOrders).Error; err == nil {
			for _, mo := range moldOrders {
				moldOrderMap[mo.ID] = mo
			}
		}
	}

	// 4. 批量查询生产订单信息
	productionOrderMap := make(map[int64]*entity.ProductionOrder)
	if len(productionOrderIDs) > 0 {
		var productionOrders []*entity.ProductionOrder
		if err := DB.Preload("Mold").
			Preload("Items.Mold").
			Where("id IN ?", productionOrderIDs).Find(&productionOrders).Error; err == nil {
			for _, po := range productionOrders {
				productionOrderMap[po.ID] = po
			}
		}
	}

	// 5. 从缓存中为每条应收账款填充订单信息
	for _, ar := range ars {
		if ar.OrderID == nil {
			continue
		}

		switch ar.OrderType {
		case "master_order":
			if mo, ok := masterOrderMap[*ar.OrderID]; ok {
				d.fillMasterOrderInfo(ar, mo)
			}
		case "mold_order":
			if mo, ok := moldOrderMap[*ar.OrderID]; ok {
				d.fillMoldOrderInfo(ar, mo)
			}
		case "production_order":
			if po, ok := productionOrderMap[*ar.OrderID]; ok {
				d.fillProductionOrderInfo(ar, po)
			}
		}
	}

	return nil
}

// fillMasterOrderInfo 填充主订单信息
func (d *AccountsReceivableDAO) fillMasterOrderInfo(ar *entity.AccountsReceivable, mo *entity.MasterOrder) {
	ar.OrderCode = mo.OrderCode

	// 从模具订单获取模具编号
	if ar.MoldCode == "" && len(mo.MoldOrders) > 0 {
		for _, moldOrder := range mo.MoldOrders {
			if len(moldOrder.Items) > 0 {
				for _, item := range moldOrder.Items {
					if item.Mold != nil && item.Mold.MoldCode != "" {
						ar.MoldCode = item.Mold.MoldCode
						break
					}
				}
			}
			if ar.MoldCode != "" {
				break
			}
		}
	}

	// 从生产订单获取产品名称和模具编号
	if len(mo.ProductionOrders) > 0 {
		for _, prodOrder := range mo.ProductionOrders {
			// 产品名称
			if ar.ProductName == "" && prodOrder.ProductName != "" {
				ar.ProductName = prodOrder.ProductName
			}

			// 模具编号：优先从生产订单本身的 MoldID 或 ExternalMoldCode 取
			if ar.MoldCode == "" {
				if prodOrder.Mold != nil && prodOrder.Mold.MoldCode != "" {
					ar.MoldCode = prodOrder.Mold.MoldCode
				} else if prodOrder.ExternalMoldCode != nil && *prodOrder.ExternalMoldCode != "" {
					ar.MoldCode = *prodOrder.ExternalMoldCode
				}
			}

			// 模具编号：从生产订单明细取
			if len(prodOrder.Items) > 0 {
				for _, item := range prodOrder.Items {
					if ar.ProductName == "" && item.ProductName != "" {
						ar.ProductName = item.ProductName
					}
					if ar.MoldCode == "" {
						if item.Mold != nil && item.Mold.MoldCode != "" {
							ar.MoldCode = item.Mold.MoldCode
						} else if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
							ar.MoldCode = *item.ExternalMoldCode
						}
					}
				}
			}
		}
	}
}

// fillMoldOrderInfo 填充模具订单信息
func (d *AccountsReceivableDAO) fillMoldOrderInfo(ar *entity.AccountsReceivable, mo *entity.MoldOrder) {
	ar.OrderCode = mo.OrderCode
	if ar.MoldCode == "" && len(mo.Items) > 0 {
		ar.MoldCode = getMoldCodeFromMoldOrderItems(mo.Items)
	}
}

// fillProductionOrderInfo 填充生产订单信息
func (d *AccountsReceivableDAO) fillProductionOrderInfo(ar *entity.AccountsReceivable, po *entity.ProductionOrder) {
	ar.OrderCode = po.OrderCode

	// 产品名称
	if ar.ProductName == "" && po.ProductName != "" {
		ar.ProductName = po.ProductName
	}

	// 模具编号：优先从生产订单本身的 MoldID 或 ExternalMoldCode 取
	if ar.MoldCode == "" {
		if po.Mold != nil && po.Mold.MoldCode != "" {
			ar.MoldCode = po.Mold.MoldCode
		} else if po.ExternalMoldCode != nil && *po.ExternalMoldCode != "" {
			ar.MoldCode = *po.ExternalMoldCode
		}
	}

	// 模具编号：从生产订单明细取（支持外来模具编号）
	if ar.MoldCode == "" && len(po.Items) > 0 {
		for _, item := range po.Items {
			if ar.ProductName == "" && item.ProductName != "" {
				ar.ProductName = item.ProductName
			}
			if ar.MoldCode == "" {
				if item.Mold != nil && item.Mold.MoldCode != "" {
					ar.MoldCode = item.Mold.MoldCode
				} else if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
					ar.MoldCode = *item.ExternalMoldCode
				}
			}
		}
	}
}

// attachMasterOrderInfo 补充综合订单的模具编号和产品名称
func (d *AccountsReceivableDAO) attachMasterOrderInfo(ar *entity.AccountsReceivable) {
	var mo entity.MasterOrder
	if err := DB.Preload("MoldOrders.Items.Mold").
		Preload("ProductionOrders.Mold").
		Preload("ProductionOrders.Items.Mold").
		Preload("ProductionOrders.Items.Product").
		First(&mo, *ar.OrderID).Error; err != nil {
		return
	}

	ar.OrderCode = mo.OrderCode

	// 从模具订单获取模具编号
	if ar.MoldCode == "" && len(mo.MoldOrders) > 0 {
		for _, moldOrder := range mo.MoldOrders {
			if len(moldOrder.Items) > 0 {
				for _, item := range moldOrder.Items {
					if item.Mold != nil && item.Mold.MoldCode != "" {
						ar.MoldCode = item.Mold.MoldCode
						break
					}
				}
			}
			if ar.MoldCode != "" {
				break
			}
		}
	}

	// 从生产订单获取产品名称和模具编号
	if len(mo.ProductionOrders) > 0 {
		for _, prodOrder := range mo.ProductionOrders {
			// 产品名称（直接取生产订单的）
			if ar.ProductName == "" && prodOrder.ProductName != "" {
				ar.ProductName = prodOrder.ProductName
			}

			// 模具编号：优先从生产订单本身的 MoldID 或 ExternalMoldCode 取
			if ar.MoldCode == "" {
				if prodOrder.Mold != nil && prodOrder.Mold.MoldCode != "" {
					ar.MoldCode = prodOrder.Mold.MoldCode
				} else if prodOrder.ExternalMoldCode != nil && *prodOrder.ExternalMoldCode != "" {
					ar.MoldCode = *prodOrder.ExternalMoldCode
				}
			}

			// 模具编号：从生产订单明细取
			if len(prodOrder.Items) > 0 {
				for _, item := range prodOrder.Items {
					if ar.ProductName == "" && item.ProductName != "" {
						ar.ProductName = item.ProductName
					}
					if ar.MoldCode == "" {
						if item.Mold != nil && item.Mold.MoldCode != "" {
							ar.MoldCode = item.Mold.MoldCode
						} else if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
							ar.MoldCode = *item.ExternalMoldCode
						}
					}
				}
			}
		}
	}
}

func getMoldCodeFromMoldOrderItems(items []*entity.MoldOrderItem) string {
	for _, item := range items {
		if item.Mold != nil && item.Mold.MoldCode != "" {
			return item.Mold.MoldCode
		}
	}
	return ""
}

// attachMoldOrderInfo 补充模具订单的模具编号
func (d *AccountsReceivableDAO) attachMoldOrderInfo(ar *entity.AccountsReceivable) {
	var mo entity.MoldOrder
	if err := DB.Preload("Items.Mold").First(&mo, *ar.OrderID).Error; err != nil {
		return
	}
	ar.OrderCode = mo.OrderCode
	if ar.MoldCode == "" && len(mo.Items) > 0 {
		ar.MoldCode = getMoldCodeFromMoldOrderItems(mo.Items)
	}
}

// attachProductionOrderInfo 补充生产订单的模具编号和产品名称
func (d *AccountsReceivableDAO) attachProductionOrderInfo(ar *entity.AccountsReceivable) {
	var po entity.ProductionOrder
	if err := DB.Preload("Mold").Preload("Items.Mold").First(&po, *ar.OrderID).Error; err != nil {
		return
	}
	ar.OrderCode = po.OrderCode

	// 产品名称
	if ar.ProductName == "" && po.ProductName != "" {
		ar.ProductName = po.ProductName
	}

	// 模具编号：优先从生产订单本身的 MoldID 或 ExternalMoldCode 取
	if ar.MoldCode == "" {
		if po.Mold != nil && po.Mold.MoldCode != "" {
			ar.MoldCode = po.Mold.MoldCode
		} else if po.ExternalMoldCode != nil && *po.ExternalMoldCode != "" {
			ar.MoldCode = *po.ExternalMoldCode
		}
	}

	// 模具编号：从生产订单明细取（支持外来模具编号）
	if ar.MoldCode == "" && len(po.Items) > 0 {
		for _, item := range po.Items {
			if ar.ProductName == "" && item.ProductName != "" {
				ar.ProductName = item.ProductName
			}
			if ar.MoldCode == "" {
				if item.Mold != nil && item.Mold.MoldCode != "" {
					ar.MoldCode = item.Mold.MoldCode
				} else if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
					ar.MoldCode = *item.ExternalMoldCode
				}
			}
		}
	}
}

// Update 更新应收账款
func (d *AccountsReceivableDAO) Update(ar *entity.AccountsReceivable) error {
	return DB.Save(ar).Error
}

// Delete 删除应收账款
func (d *AccountsReceivableDAO) Delete(id int64) error {
	return DB.Delete(&entity.AccountsReceivable{}, id).Error
}

func (d *AccountsReceivableDAO) attachCustomer(ar *entity.AccountsReceivable) error {
	if ar == nil {
		return nil
	}
	var customer entity.Customer
	if err := DB.First(&customer, ar.CustomerID).Error; err != nil {
		return err
	}
	ar.Customer = &customer
	return nil
}

func (d *AccountsReceivableDAO) attachCustomers(ars []*entity.AccountsReceivable) error {
	if len(ars) == 0 {
		return nil
	}

	customerIDs := make([]int64, 0)
	customerIDMap := make(map[int64]bool)
	for _, ar := range ars {
		if ar.CustomerID > 0 && !customerIDMap[ar.CustomerID] {
			customerIDs = append(customerIDs, ar.CustomerID)
			customerIDMap[ar.CustomerID] = true
		}
	}

	if len(customerIDs) == 0 {
		return nil
	}

	// 批量查询客户信息
	var customers []*entity.Customer
	if err := DB.Where("id IN ?", customerIDs).Find(&customers).Error; err != nil {
		return err
	}

	customerMap := make(map[int64]*entity.Customer)
	for _, customer := range customers {
		customerMap[customer.ID] = customer
	}

	// 为每条应收账款分配客户信息
	for _, ar := range ars {
		if customer, ok := customerMap[ar.CustomerID]; ok {
			ar.Customer = customer
		}
	}

	return nil
}

// GetOverdueList 获取逾期应收账款列表
func (d *AccountsReceivableDAO) GetOverdueList() ([]*entity.AccountsReceivable, error) {
	var ars []*entity.AccountsReceivable
	err := DB.Where("status != ? AND due_date < CURDATE()", "PAID").
		Order("due_date ASC").
		Find(&ars).Error
	if err != nil {
		return nil, err
	}
	if err := d.attachCustomers(ars); err != nil {
		return nil, err
	}
	return ars, nil
}

// GetAgingStatistics 获取账龄统计
func (d *AccountsReceivableDAO) GetAgingStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 按账龄分类统计
	type AgingStat struct {
		Category string  `gorm:"column:aging_category"`
		Count    int64   `gorm:"column:count"`
		Amount   float64 `gorm:"column:amount"`
	}

	var agingStats []AgingStat
	err := DB.Model(&entity.AccountsReceivable{}).
		Select("aging_category, COUNT(*) as count, SUM(balance) as amount").
		Where("status != ?", "PAID").
		Group("aging_category").
		Scan(&agingStats).Error

	if err != nil {
		return nil, err
	}

	stats["aging"] = agingStats

	// 总体统计
	var totalStats struct {
		TotalAmount    float64 `gorm:"column:total_amount"`
		ReceivedAmount float64 `gorm:"column:received_amount"`
		Balance        float64 `gorm:"column:balance"`
		OverdueCount   int64   `gorm:"column:overdue_count"`
		OverdueAmount  float64 `gorm:"column:overdue_amount"`
	}

	err = DB.Model(&entity.AccountsReceivable{}).
		Select(`
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(received_amount), 0) as received_amount,
			COALESCE(SUM(balance), 0) as balance,
			COALESCE(SUM(CASE WHEN status = 'overdue' THEN 1 ELSE 0 END), 0) as overdue_count,
			COALESCE(SUM(CASE WHEN status = 'overdue' THEN balance ELSE 0 END), 0) as overdue_amount
		`).
		Scan(&totalStats).Error

	if err != nil {
		return nil, err
	}

	stats["summary"] = totalStats

	return stats, nil
}

// ReceivePayment 收款（增加已收金额）
func (d *AccountsReceivableDAO) ReceivePayment(id int64, amount float64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var ar entity.AccountsReceivable
		if err := tx.First(&ar, id).Error; err != nil {
			return err
		}

		ar.ReceivedAmount += amount

		if ar.ReceivedAmount >= ar.Amount {
			ar.Status = "PAID"
		} else if ar.ReceivedAmount > 0 {
			ar.Status = "PARTIAL"
		}

		return tx.Save(&ar).Error
	})
}

// ReceivePaymentWithSync 收款并同步到模具订单款项表
func (d *AccountsReceivableDAO) ReceivePaymentWithSync(id int64, amount float64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var ar entity.AccountsReceivable
		if err := tx.First(&ar, id).Error; err != nil {
			return err
		}

		ar.ReceivedAmount += amount
		if ar.ReceivedAmount >= ar.Amount {
			ar.Status = "PAID"
		} else if ar.ReceivedAmount > 0 {
			ar.Status = "PARTIAL"
		}
		if err := tx.Save(&ar).Error; err != nil {
			return err
		}

		// ✨ 同步更新模具订单款项表
		if ar.OrderType == "mold_order" && ar.OrderID != nil {
			// 查找该订单的所有款项记录
			var payments []entity.MoldOrderPayment
			if err := tx.Where("order_id = ?", *ar.OrderID).Order("id ASC").Find(&payments).Error; err == nil {
				remainingAmount := amount

				// 按顺序分配收款金额到各个款项
				for i := range payments {
					if remainingAmount <= 0 {
						break
					}

					payment := &payments[i]
					if payment.Status >= 3 { // 已收款，跳过
						continue
					}

					unreceived := payment.ShouldAmount - payment.ActualAmount
					if remainingAmount >= unreceived {
						// 该款项完全收清
						payment.ActualAmount = payment.ShouldAmount
						payment.Status = 3 // 已收款
						now := time.Now()
						payment.PaymentDate = &now
						remainingAmount -= unreceived
					} else {
						// 部分收款
						payment.ActualAmount += remainingAmount
						payment.Status = 2 // 部分收款
						remainingAmount = 0
					}

					if err := tx.Save(payment).Error; err != nil {
						logger.Warn("同步模具款项失败", zap.Error(err))
					}
				}
			}
		}

		return nil
	})
}

// ReceivePaymentWithDetails 收款并创建详细记录（支持优惠减免）
func (d *AccountsReceivableDAO) ReceivePaymentWithDetails(id int64, amount float64, discountAmount float64, paymentDate time.Time, paymentMethod, transactionNo, receivingAccount, remark, receiptFileID, invoiceFileID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var ar entity.AccountsReceivable
		if err := tx.First(&ar, id).Error; err != nil {
			return err
		}

		// 1. 创建收款记录
		payment := &entity.ReceivablePayment{
			ReceivableID:     id,
			PaymentAmount:    amount,
			DiscountAmount:   discountAmount,
			PaymentDate:      paymentDate,
			PaymentMethod:    paymentMethod,
			TransactionNo:    transactionNo,
			ReceivingAccount: receivingAccount,
			Remark:           remark,
			ReceiptFileID:    receiptFileID,
			InvoiceFileID:    invoiceFileID,
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// 2. 更新应收账款（balance 为数据库生成列，自动计算 amount - received_amount - discount_amount）
		ar.ReceivedAmount += amount
		ar.DiscountAmount += discountAmount
		if ar.ReceivedAmount+ar.DiscountAmount >= ar.Amount {
			ar.Status = "PAID"
		} else if ar.ReceivedAmount+ar.DiscountAmount > 0 {
			ar.Status = "PARTIAL"
		}
		if err := tx.Save(&ar).Error; err != nil {
			return err
		}

		// 3. 同步更新模具订单款项表（使用事务连接查询，避免读到陈旧数据）
		if ar.OrderType == "mold_order" && ar.OrderID != nil {
			var payments []entity.MoldOrderPayment
			tx.Where("order_id = ?", *ar.OrderID).Find(&payments)
			remainingAmount := amount + discountAmount
			for _, payment := range payments {
				if remainingAmount <= 0 {
					break
				}
				if payment.Status < 3 {
					unpaidAmount := payment.ShouldAmount - payment.ActualAmount
					if remainingAmount >= unpaidAmount {
						payment.ActualAmount = payment.ShouldAmount
						payment.Status = 3
						payment.PaymentDate = &paymentDate
						remainingAmount -= unpaidAmount
					} else {
						payment.ActualAmount += remainingAmount
						payment.Status = 2
						payment.PaymentDate = &paymentDate
						remainingAmount = 0
					}
					if err := tx.Save(payment).Error; err != nil {
						logger.Warn("同步模具款项失败", zap.Error(err))
					}
				}
			}
		}

		return nil
	})
}
