package entity

import "time"

type ReceivingAccount struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:账户ID" json:"id"`
	AccountName string    `gorm:"size:100;not null;comment:账户名称（如：公司公账、法人银行卡）" json:"account_name"`
	AccountNo   string    `gorm:"size:100;not null;comment:账号" json:"account_no"`
	BankName    string    `gorm:"size:200;comment:开户行" json:"bank_name"`
	AccountType string    `gorm:"size:50;comment:账户类型：company=对公, personal=个人, wechat=微信, alipay=支付宝" json:"account_type"`
	Sort        int       `gorm:"default:0;comment:排序" json:"sort"`
	Status      int8      `gorm:"default:1;comment:状态：1=启用, 0=禁用" json:"status"`
	Remark      string    `gorm:"size:500;comment:备注" json:"remark"`
	CreatedBy   int64     `gorm:"comment:创建人ID" json:"created_by"`
	CreateTime  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime  time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (ReceivingAccount) TableName() string {
	return "receiving_accounts"
}
