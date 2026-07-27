package entity

import "time"

type SysCompany struct {
	ID               int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CompanyName      string    `gorm:"column:company_name;type:varchar(100);not null" json:"company_name"`
	ShortName        string    `gorm:"column:short_name;type:varchar(50)" json:"short_name"`
	EnglishName      string    `gorm:"column:english_name;type:varchar(200)" json:"english_name"`
	EnglishShortName string    `gorm:"column:english_short_name;type:varchar(100)" json:"english_short_name"`
	Logo             string    `gorm:"column:logo;type:varchar(500)" json:"logo"`
	CreateTime       time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime       time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

// TableName 指定表名
func (SysCompany) TableName() string {
	return "sys_company"
}
