package entity

import "time"

type SysOperLog struct {
	ID            int64     `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Title         string    `gorm:"size:128;not null;default:'';comment:操作模块标题" json:"title"`
	OperType      int8      `gorm:"not null;default:0;comment:操作类型 1=增 2=删 3=改 4=查 5=导入 6=导出 7=登录 8=登出" json:"oper_type"`
	BusinessType  string    `gorm:"size:64;not null;default:'';comment:业务类型" json:"business_type"`
	BusinessID    int64     `gorm:"not null;default:0;comment:业务记录ID" json:"business_id"`
	Method        string    `gorm:"size:255;not null;default:'';comment:请求方法（路由路径）" json:"method"`
	RequestURL    string    `gorm:"size:255;not null;default:'';comment:请求URL" json:"request_url"`
	RequestMethod string    `gorm:"size:10;not null;default:'';comment:请求方式" json:"request_method"`
	OperName      string    `gorm:"size:64;not null;default:'';comment:操作人员" json:"oper_name"`
	DeptName      string    `gorm:"size:64;not null;default:'';comment:部门名称" json:"dept_name"`
	OperIP        string    `gorm:"size:128;not null;default:'';comment:操作IP" json:"oper_ip"`
	RequestParams string    `gorm:"type:text;comment:请求参数" json:"request_params,omitempty"`
	JSONResult    string    `gorm:"type:text;comment:返回结果" json:"json_result,omitempty"`
	Status        int8      `gorm:"not null;default:0;comment:操作状态 0=正常 1=异常" json:"status"`
	ErrorMsg      string    `gorm:"size:2048;not null;default:'';comment:错误消息" json:"error_msg,omitempty"`
	CostTime      int64     `gorm:"not null;default:0;comment:耗时(ms)" json:"cost_time"`
	CreateTime    time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
}

// TableName 指定表名
func (SysOperLog) TableName() string {
	return "sys_oper_log"
}
