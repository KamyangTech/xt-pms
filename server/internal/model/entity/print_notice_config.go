package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type PrintNoticeConfig struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:配置ID" json:"id"`
	UserID     int64      `gorm:"not null;index;comment:用户ID，0表示全局配置" json:"user_id"`
	ConfigType string     `gorm:"size:50;not null;index;default:'production_order';comment:配置类型" json:"config_type"`
	ConfigData JSONConfig `gorm:"type:json;not null;comment:配置数据（JSON格式）" json:"config_data"`
	IsDefault  int8       `gorm:"default:0;comment:是否为默认配置：1=是，0=否" json:"is_default"`
	CreatedAt  time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (PrintNoticeConfig) TableName() string {
	return "print_notice_config"
}

// JSONConfig 自定义JSON类型，用于GORM的JSON字段
type JSONConfig map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONConfig) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONConfig) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

type PrintNoticeConfigData struct {
	BasicInfo      BasicInfoConfig      `json:"basic_info"`
	ProductionInfo ProductionInfoConfig `json:"production_info"`
	DetailColumns  DetailColumnsConfig  `json:"detail_columns"`
}

type BasicInfoConfig struct {
	OrderCode      bool `json:"order_code"`
	CustomerName   bool `json:"customer_name"`
	ProductionType bool `json:"production_type"`
	MoldSource     bool `json:"mold_source"`
}

type ProductionInfoConfig struct {
	PlanQty       bool `json:"plan_qty"`
	Worker        bool `json:"worker"`
	PlanStartDate bool `json:"plan_start_date"`
	PlanEndDate   bool `json:"plan_end_date"`
	CompletedQty  bool `json:"completed_qty"`
}

type DetailColumnsConfig struct {
	ProductName  bool `json:"product_name"`
	MoldCode     bool `json:"mold_code"`
	ProductImage bool `json:"product_image"`
	PlanQty      bool `json:"plan_qty"`
	CompletedQty bool `json:"completed_qty"`
}

// GetDefaultConfig 获取生产通知单默认配置
func GetDefaultConfig() PrintNoticeConfigData {
	return PrintNoticeConfigData{
		BasicInfo: BasicInfoConfig{
			OrderCode:      true,
			CustomerName:   true,
			ProductionType: true,
			MoldSource:     true,
		},
		ProductionInfo: ProductionInfoConfig{
			PlanQty:       true,
			Worker:        true,
			PlanStartDate: true,
			PlanEndDate:   true,
			CompletedQty:  true,
		},
		DetailColumns: DetailColumnsConfig{
			ProductName:  true,
			MoldCode:     true,
			ProductImage: true,
			PlanQty:      true,
			CompletedQty: true,
		},
	}
}

// ======================== 送货单打印配置 ========================

type DeliveryNoteConfigData struct {
	BasicInfo    DeliveryNoteBasicInfo    `json:"basic_info"`
	DeliveryInfo DeliveryNoteDeliveryInfo `json:"delivery_info"`
	DetailColumn DeliveryNoteDetailColumn `json:"detail_columns"`
}

type DeliveryNoteBasicInfo struct {
	StockCode    bool `json:"stock_code"`
	CustomerName bool `json:"customer_name"`
	MoldCode     bool `json:"mold_code"`
	ProductName  bool `json:"product_name"`
}

type DeliveryNoteDeliveryInfo struct {
	DeliveryDate   bool `json:"delivery_date"`
	OutDate        bool `json:"out_date"`
	DeliveryMethod bool `json:"delivery_method"`
	Quantity       bool `json:"quantity"`
	ExpressCompany bool `json:"express_company"`
	ExpressNo      bool `json:"express_no"`
	DriverPhone    bool `json:"driver_phone"`
	DriverPlate    bool `json:"driver_plate"`
}

type DeliveryNoteDetailColumn struct {
	ProductImage bool `json:"product_image"`
	MoldCode     bool `json:"mold_code"`
	ProductName  bool `json:"product_name"`
	Quantity     bool `json:"quantity"`
	BatchNo      bool `json:"batch_no"`
	Remark       bool `json:"remark"`
}

// GetDefaultDeliveryNoteConfig 获取送货单默认配置
func GetDefaultDeliveryNoteConfig() DeliveryNoteConfigData {
	return DeliveryNoteConfigData{
		BasicInfo: DeliveryNoteBasicInfo{
			StockCode:    true,
			CustomerName: true,
			MoldCode:     true,
			ProductName:  true,
		},
		DeliveryInfo: DeliveryNoteDeliveryInfo{
			DeliveryDate:   true,
			OutDate:        true,
			DeliveryMethod: true,
			Quantity:       true,
			ExpressCompany: true,
			ExpressNo:      true,
			DriverPhone:    true,
			DriverPlate:    true,
		},
		DetailColumn: DeliveryNoteDetailColumn{
			ProductImage: true,
			MoldCode:     true,
			ProductName:  true,
			Quantity:     true,
			BatchNo:      true,
			Remark:       true,
		},
	}
}
