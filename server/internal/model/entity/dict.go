package entity

import "time"

type SysDictType struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:字典类型 ID" json:"id"`
	DictName   string    `gorm:"size:100;not null;comment:字典名称" json:"dict_name"`
	DictCode   string    `gorm:"size:50;not null;uniqueIndex;comment:字典编码" json:"dict_code"`
	Remark     string    `gorm:"size:255;comment:备注" json:"remark"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	DataList []*SysDictData `gorm:"foreignKey:DictTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"data_list,omitempty"`
}

func (SysDictType) TableName() string {
	return "sys_dict_type"
}

type SysDictData struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;comment:字典数据 ID" json:"id"`
	DictTypeID      int64     `gorm:"not null;index;comment:字典类型 ID" json:"dict_type_id"`
	DictLabel       string    `gorm:"size:100;not null;comment:显示标签" json:"dict_label"`
	DictValue       string    `gorm:"size:100;not null;comment:实际值（语义化字符串，如 TESTING、COMPLETED）" json:"dict_value"`
	Color           string    `gorm:"size:20;default:default;comment:颜色标签" json:"color"`
	Sort            int       `gorm:"default:0;comment:排序" json:"sort"`
	Status          int8      `gorm:"default:1;comment:状态：1=启用，0=禁用" json:"status"`
	IsLocked        int8      `gorm:"default:0;comment:锁定级别：0=未锁定，1=可解锁锁定，2=永久锁定" json:"is_locked"`
	AllowProduction         int8      `gorm:"default:0;comment:是否可用于生产：1=是，0=否（仅模具状态字典有效）" json:"allow_production"`
	AllowWorkReport         int8      `gorm:"default:1;comment:是否在生产报工页面展示：1=是，0=否（仅生产订单状态字典有效）" json:"allow_work_report"`
	NeedsDeliveryUnlock     int8      `gorm:"default:0;comment:是否需要交付后解锁：1=是，0=否（仅款项类型字典有效）" json:"needs_delivery_unlock"`
	Remark          string    `gorm:"size:255;comment:备注" json:"remark"`
	Ext             string    `gorm:"type:text;comment:扩展配置（JSON格式，如付款方式关联的账户类型）" json:"ext"`
	CreateTime      time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime      time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	DictType *SysDictType `gorm:"foreignKey:DictTypeID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"dict_type,omitempty"`
}

func (SysDictData) TableName() string {
	return "sys_dict_data"
}
