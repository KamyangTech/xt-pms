package entity

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
)

type AccountsReceivable struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;comment:应收账款ID" json:"id"`
	CustomerID     int64      `gorm:"not null;comment:客户ID（customer.id）" json:"customer_id"`
	OrderID        *int64     `gorm:"comment:关联订单ID（mold_order.id 或 production_order.id）" json:"order_id"`
	OrderType      string     `gorm:"size:50;comment:订单类型：mold_order=模具订单, production_order=生产订单, master_order=综合订单" json:"order_type"`
	PaymentType    string     `gorm:"size:50;comment:款项类型：deposit=首付款, final=尾款, production=生产款, total=总应收" json:"payment_type"`
	Amount         float64    `gorm:"type:decimal(12,2);not null;default:0.00;comment:应收金额" json:"amount"`
	ReceivedAmount float64    `gorm:"type:decimal(12,2);default:0.00;comment:已收金额（实收）" json:"received_amount"`
	DiscountAmount float64    `gorm:"type:decimal(12,2);default:0.00;comment:优惠减免金额" json:"discount_amount"`
	Balance        float64    `gorm:"type:decimal(12,2);->;comment:余额（数据库生成列）" json:"balance"`
	DueDate        *dto.Date  `gorm:"type:date;comment:到期日期" json:"due_date"`
	AgingDays      int        `gorm:"comment:逾期天数" json:"aging_days"`
	AgingCategory  string     `gorm:"size:20;comment:账龄分类：CURRENT=当期, WITHIN_30=30天内, WITHIN_60=60天内, WITHIN_90=90天内, OVERDUE=逾期" json:"aging_category"`
	Status         string     `gorm:"size:20;default:PENDING;comment:状态：PENDING=待收款, LOCKED=锁定, PARTIAL=部分收款, PAID=已收款, OVERDUE=逾期" json:"status"`
	IsLocked       bool       `gorm:"default:false;comment:是否锁定" json:"is_locked"`
	UnlockRule     string     `gorm:"size:100;comment:解锁规则" json:"unlock_rule"`
	SourceType     string     `gorm:"size:20;default:'';comment:款项来源：mold=模具款, production=生产款" json:"source_type"`
	Remark         string     `gorm:"type:text;comment:备注" json:"remark"`
	CreatedBy      *int64     `gorm:"comment:创建人ID" json:"created_by"`
	CreateTime     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 非数据库字段：列表展示信息
	OrderCode   string `gorm:"-" json:"order_code,omitempty"`
	MoldCode    string `gorm:"-" json:"mold_code,omitempty"`
	ProductName string `gorm:"-" json:"product_name,omitempty"`

	// 关联关系
	Customer *Customer `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"customer,omitempty"`
}

func (AccountsReceivable) TableName() string {
	return "accounts_receivable"
}

type AccountsPayable struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;comment:应付账款ID" json:"id"`
	SupplierID     int64      `gorm:"not null;comment:供应商/加工商ID（supplier.id 或 processor.id）" json:"supplier_id"`
	OrderID        *int64     `gorm:"comment:关联订单ID（outsourcing_record.id 等）" json:"order_id"`
	OrderType      string     `gorm:"size:50;comment:订单类型：outsourcing=外发订单, purchase=采购订单" json:"order_type"`
	PaymentType    string     `gorm:"size:50;default:'';comment:款项类型：total=全款, first=首期款, final=尾款, installment=分期款" json:"payment_type"`
	SourceType     string     `gorm:"size:20;default:'';comment:款项来源：mold=模具, production=生产, purchase=采购, outsource=外发(旧数据)" json:"source_type"`
	Amount         float64    `gorm:"type:decimal(12,2);not null;default:0.00;comment:应付金额" json:"amount"`
	PaidAmount     float64    `gorm:"type:decimal(12,2);default:0.00;comment:已付金额" json:"paid_amount"`
	DiscountAmount float64    `gorm:"type:decimal(12,2);default:0.00;comment:优惠减免金额" json:"discount_amount"`
	Balance        float64    `gorm:"type:decimal(12,2);->;comment:余额（只读）" json:"balance"`
	DueDate        *dto.Date  `gorm:"type:date;comment:到期日期" json:"due_date"`
	Status         string     `gorm:"size:20;default:PENDING;comment:状态：PENDING=待付款, PARTIAL=部分付款, PAID=已付款, OVERDUE=逾期" json:"status"`
	Remark         string     `gorm:"type:text;comment:备注" json:"remark"`
	CreatedBy      *int64     `gorm:"comment:创建人ID" json:"created_by"`
	CreateTime     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联订单编号（外发单号等）
	OrderCode string `gorm:"size:50;default:'';comment:关联订单编号（外发单号等）" json:"order_code,omitempty"`

	// 关联关系
	Supplier  *Supplier  `gorm:"foreignKey:SupplierID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"supplier,omitempty"`
	Processor *Processor `gorm:"foreignKey:SupplierID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"processor,omitempty"`
}

func (AccountsPayable) TableName() string {
	return "accounts_payable"
}

type FinanceExpense struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;comment:费用ID" json:"id"`
	ExpenseType      string     `gorm:"size:50;not null;comment:费用类型：freight=运费, packaging=包装费, insurance=保险费, other=其他" json:"expense_type"`
	RelatedOrderID   *int64     `gorm:"comment:关联订单ID（mold_order.id 或 sale_order.id）" json:"related_order_id"`
	RelatedOrderType string     `gorm:"size:50;comment:订单类型：mold_order=模具订单, sale_order=销售订单" json:"related_order_type"`
	DeliveryID       *int64     `gorm:"comment:关联交付记录ID（mold_delivery.id）" json:"delivery_id"`
	Amount           float64    `gorm:"type:decimal(12,2);not null;default:0.00;comment:费用金额" json:"amount"`
	PaidAmount       float64    `gorm:"type:decimal(12,2);default:0.00;comment:已付金额" json:"paid_amount"`
	PaymentDate      *time.Time `gorm:"type:date;comment:支付日期" json:"payment_date"`
	Payer            string     `gorm:"size:100;comment:支付方" json:"payer"`
	Receiver         string     `gorm:"size:100;comment:收款方" json:"receiver"`
	Remark           string     `gorm:"type:text;comment:备注" json:"remark"`
	CreatedBy        *int64     `gorm:"comment:创建人ID" json:"created_by"`
	CreateTime       time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime       time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 非数据库字段（由 Service 层填充）
	Balance float64 `gorm:"-" json:"balance"`
	Status  string  `gorm:"-" json:"status"`
}

func (FinanceExpense) TableName() string {
	return "finance_expense"
}

type ReceivablePayment struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;comment:收款记录ID" json:"id"`
	ReceivableID     int64     `gorm:"not null;comment:应收账款ID" json:"receivable_id"`
	PaymentAmount    float64   `gorm:"type:decimal(12,2);not null;default:0.00;comment:收款金额（实际到账）" json:"payment_amount"`
	DiscountAmount   float64   `gorm:"type:decimal(12,2);default:0.00;comment:本次优惠金额" json:"discount_amount"`
	PaymentDate      time.Time `gorm:"type:date;not null;comment:收款日期" json:"payment_date"`
	PaymentMethod    string    `gorm:"size:50;comment:收款方式：bank_transfer/cash/check/wechat/alipay" json:"payment_method"`
	TransactionNo    string    `gorm:"size:100;comment:交易流水号/回单号" json:"transaction_no"`
	ReceivingAccount string    `gorm:"size:100;comment:收款账户" json:"receiving_account"`
	Remark           string    `gorm:"type:text;comment:备注" json:"remark"`
	ReceiptFileID    string    `gorm:"size:500;comment:回单文件ID（多个ID用逗号分隔）" json:"receipt_file_id"`
	InvoiceFileID    string    `gorm:"size:500;comment:发票文件ID（多个ID用逗号分隔）" json:"invoice_file_id"`
	CreatedBy        *int64    `gorm:"comment:创建人ID" json:"created_by"`
	CreateTime       time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime       time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Receivable *AccountsReceivable `gorm:"foreignKey:ReceivableID" json:"receivable,omitempty"`

	// 非数据库字段：文件信息（由 Service 层填充，支持多文件）
	ReceiptFiles []*File `gorm:"-" json:"receipt_files,omitempty"`
	InvoiceFiles []*File `gorm:"-" json:"invoice_files,omitempty"`
}

func (ReceivablePayment) TableName() string {
	return "receivable_payments"
}

type PayablePayment struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;comment:付款记录ID" json:"id"`
	PayableID        int64     `gorm:"not null;comment:应付账款ID" json:"payable_id"`
	PaymentAmount    float64   `gorm:"type:decimal(12,2);not null;default:0.00;comment:付款金额" json:"payment_amount"`
	DiscountAmount   float64   `gorm:"type:decimal(12,2);default:0.00;comment:本次优惠金额" json:"discount_amount"`
	PaymentDate      time.Time `gorm:"type:date;not null;comment:付款日期" json:"payment_date"`
	PaymentMethod    string    `gorm:"size:50;comment:付款方式：bank_transfer/cash/check/wechat/alipay" json:"payment_method"`
	TransactionNo    string    `gorm:"size:100;comment:交易流水号/回单号" json:"transaction_no"`
	PaymentAccount   string    `gorm:"size:100;comment:付款账户" json:"payment_account"`
	ReceiverAccount  string    `gorm:"size:100;comment:收款方账户" json:"receiver_account"`
	Remark           string    `gorm:"type:text;comment:备注" json:"remark"`
	ReceiptFileID    string    `gorm:"size:500;comment:回单文件ID（多个ID用逗号分隔）" json:"receipt_file_id"`
	InvoiceFileID    string    `gorm:"size:500;comment:发票文件ID（多个ID用逗号分隔）" json:"invoice_file_id"`
	CreatedBy        *int64    `gorm:"comment:创建人ID" json:"created_by"`
	CreateTime       time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime       time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Payable *AccountsPayable `gorm:"foreignKey:PayableID" json:"payable,omitempty"`

	// 非数据库字段：文件信息（由 Service 层填充，支持多文件）
	ReceiptFiles []*File `gorm:"-" json:"receipt_files,omitempty"`
	InvoiceFiles []*File `gorm:"-" json:"invoice_files,omitempty"`
}

func (PayablePayment) TableName() string {
	return "payable_payments"
}

type ExpensePayment struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;comment:支付记录ID" json:"id"`
	ExpenseID       int64     `gorm:"not null;comment:费用ID" json:"expense_id"`
	PaymentAmount   float64   `gorm:"type:decimal(12,2);not null;default:0.00;comment:支付金额" json:"payment_amount"`
	PaymentDate     time.Time `gorm:"type:date;comment:支付日期" json:"payment_date"`
	PaymentMethod   string    `gorm:"size:50;comment:支付方式：bank_transfer/cash/check/wechat/alipay" json:"payment_method"`
	TransactionNo   string    `gorm:"size:100;comment:交易流水号/回单号" json:"transaction_no"`
	PaymentAccount  string    `gorm:"size:100;comment:付款账户" json:"payment_account"`
	ReceiverAccount string    `gorm:"size:100;comment:收款方账户" json:"receiver_account"`
	Remark          string    `gorm:"type:text;comment:备注" json:"remark"`
	CreatedBy       *int64    `gorm:"comment:创建人ID" json:"created_by"`
	CreateTime      time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime      time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Expense *FinanceExpense `gorm:"foreignKey:ExpenseID" json:"expense,omitempty"`
}

func (ExpensePayment) TableName() string {
	return "expense_payments"
}
