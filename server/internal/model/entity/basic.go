package entity

import "time"

type Customer struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:客户 ID" json:"id"`
	CusCode    string     `gorm:"size:50;not null;uniqueIndex;comment:客户编号（唯一）" json:"cus_code"`
	CusName    string     `gorm:"size:100;not null;comment:客户名称" json:"cus_name"`
	Contact    string     `gorm:"size:50;comment:联系人" json:"contact"`
	Phone      string     `gorm:"size:20;comment:联系电话" json:"phone"`
	Email      string     `gorm:"size:100;comment:电子邮箱" json:"email"`
	Address    string     `gorm:"size:255;comment:客户地址" json:"address"`
	Remark     string     `gorm:"size:255;comment:备注" json:"remark"`
	Status     int8       `gorm:"default:1;comment:客户状态：1=正常，0=停用" json:"status"`
	CreateTime *time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime *time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (Customer) TableName() string {
	return "customer"
}

type Material struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:物料 ID" json:"id"`
	MatCode    string    `gorm:"size:50;not null;uniqueIndex;comment:物料编码（唯一）" json:"mat_code"`
	MatName    string    `gorm:"size:100;not null;comment:物料名称" json:"mat_name"`
	Spec       string    `gorm:"size:255;comment:物料规格型号" json:"spec"`
	Unit       string    `gorm:"size:20;comment:物料单位（如：个、kg）" json:"unit"`
	MatType    int8      `gorm:"default:1;comment:物料类型：1=原料，2=模具配件，3=产品，4=辅料" json:"mat_type"`
	Price      float64   `gorm:"default:0.00;comment:物料单价" json:"price"`
	StockWarn  int       `gorm:"default:0;comment:库存预警数量" json:"stock_warn"`
	Remark     string    `gorm:"size:255;comment:备注" json:"remark"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (Material) TableName() string {
	return "material"
}

type Warehouse struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:仓库 ID" json:"id"`
	WhCode     string     `gorm:"size:50;not null;uniqueIndex;comment:仓库编码（唯一）" json:"wh_code"`
	WhName     string     `gorm:"size:100;not null;comment:仓库名称" json:"wh_name"`
	Address    string     `gorm:"size:255;comment:仓库地址" json:"address"`
	Manager    string     `gorm:"size:50;comment:仓库管理员" json:"manager"`
	Phone      string     `gorm:"size:20;comment:联系电话" json:"phone"`
	Remark     string     `gorm:"size:255;comment:备注" json:"remark"`
	Status     int8       `gorm:"default:1;comment:仓库状态：0=停用，1=正常" json:"status"`
	CreateTime *time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime *time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (Warehouse) TableName() string {
	return "warehouse"
}

type Process struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:工序 ID" json:"id"`
	ProcessCode string    `gorm:"size:50;not null;uniqueIndex;comment:工序编码（唯一）" json:"process_code"`
	ProcessName string    `gorm:"size:50;not null;comment:工序名称（如：CNC、线割、装配）" json:"process_name"`
	ProcessType int8      `gorm:"default:1;comment:工序类型：1=模具工序，2=产品工序" json:"process_type"`
	WorkHour    float64   `gorm:"default:0.00;comment:标准工时" json:"work_hour"`
	Remark      string    `gorm:"size:255;comment:备注" json:"remark"`
	CreateTime  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime  time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (Process) TableName() string {
	return "process"
}

type Processor struct {
	ID            int64      `gorm:"primaryKey;autoIncrement;comment:加工商 ID" json:"id"`
	ProcessorCode string     `gorm:"size:50;not null;uniqueIndex;comment:加工商编号（唯一）" json:"processor_code"`
	ProcessorName string     `gorm:"size:100;not null;comment:加工商名称" json:"processor_name"`
	Contact       string     `gorm:"size:50;comment:联系人" json:"contact"`
	Phone         string     `gorm:"size:20;comment:联系电话" json:"phone"`
	Email         string     `gorm:"size:100;comment:电子邮箱" json:"email"`
	Address       string     `gorm:"size:255;comment:加工商地址" json:"address"`
	ProcessType   string     `gorm:"size:100;comment:加工类型" json:"process_type"`
	Remark        string     `gorm:"size:255;comment:备注" json:"remark"`
	Status        int8       `gorm:"default:1;comment:状态：0=停用，1=正常" json:"status"`
	CreateTime    *time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime    *time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (Processor) TableName() string {
	return "processor"
}

type Supplier struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:供应商 ID" json:"id"`
	SupCode    string     `gorm:"size:50;not null;uniqueIndex;comment:供应商编号（唯一）" json:"sup_code"`
	SupName    string     `gorm:"size:100;not null;comment:供应商名称" json:"sup_name"`
	Contact    string     `gorm:"size:50;comment:联系人" json:"contact"`
	Phone      string     `gorm:"size:20;comment:联系电话" json:"phone"`
	Email      string     `gorm:"size:100;comment:电子邮箱" json:"email"`
	Address    string     `gorm:"size:255;comment:供应商地址" json:"address"`
	Remark     string     `gorm:"size:255;comment:备注" json:"remark"`
	Status     int8       `gorm:"default:1;comment:状态：0=停用，1=正常" json:"status"`
	CreateTime *time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime *time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (Supplier) TableName() string {
	return "supplier"
}
