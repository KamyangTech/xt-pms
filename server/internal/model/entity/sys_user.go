package entity

import "time"

type SysUser struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	Username   string     `gorm:"size:50;not null;comment:登录账号" json:"username"`
	Password   string     `gorm:"size:100;not null;comment:密码（加密存储）" json:"-"`
	RealName   string     `gorm:"size:50;comment:真实姓名" json:"real_name"`
	Phone      string     `gorm:"size:20;comment:手机号" json:"phone"`
	Email      string     `gorm:"size:100;comment:邮箱" json:"email"`
	Position   string     `gorm:"size:50;comment:职位" json:"position"`
	HireDate   *time.Time `gorm:"type:date;comment:入职日期" json:"hire_date"`
	DeptID     int64      `gorm:"comment:关联部门 ID" json:"dept_id"`
	Status         int8       `gorm:"default:1;comment:状态：0=禁用，1=正常" json:"status"`
	LoginAttempts  int        `gorm:"column:login_attempts;default:0;comment:连续登录失败次数" json:"-"`
	LockedUntil    *time.Time `gorm:"column:locked_until;comment:账号锁定截止时间" json:"-"`
	PasswordVersion int       `gorm:"column:password_version;default:1;comment:密码版本号，修改密码时自增" json:"-"`
	CreateTime     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
	// 关联查询字段（不存储到数据库）
	Roles []*SysRole `gorm:"many2many:sys_user_role;joinForeignKey:user_id;joinReferences:role_id" json:"roles,omitempty"`
}

func (SysUser) TableName() string {
	return "sys_user"
}
