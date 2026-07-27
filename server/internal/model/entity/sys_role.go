package entity

import "time"

// 角色类型常量
const (
	RoleTypeCustom     int8 = 0 // 自定义角色 — 完全可编辑/删除
	RoleTypeSuperAdmin int8 = 1 // 超级管理员 — 不可编辑/删除/修改权限
	RoleTypeBuiltin    int8 = 2 // 业务内置角色 — 不可删除，名称/备注不可改
)

type SysRole struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:角色 ID" json:"id"`
	RoleName   string    `gorm:"size:50;not null;comment:角色名称" json:"role_name"`
	Remark     string    `gorm:"size:255;comment:角色备注" json:"remark"`
	RoleType   int8      `gorm:"default:0;comment:0=自定义 1=超级管理员 2=业务内置" json:"role_type"`
	DataScope  int8      `gorm:"default:1;comment:1=全部数据 2=本部门及以下 3=本部门 4=仅本人" json:"data_scope"`
	Status     int8      `gorm:"default:1;comment:0=禁用 1=启用" json:"status"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
	// 关联查询字段（不存储到数据库）
	Users []*SysUser `gorm:"many2many:sys_user_role;joinForeignKey:role_id;joinReferences:user_id" json:"users,omitempty"`
	Menus []*SysMenu `gorm:"many2many:sys_role_menu;joinForeignKey:role_id;joinReferences:menu_id" json:"menus,omitempty"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

type SysDept struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:部门 ID" json:"id"`
	DeptName   string     `gorm:"size:50;not null;comment:部门名称" json:"dept_name"`
	ParentID   int64      `gorm:"default:0;comment:上级部门 ID，0 表示顶级部门" json:"parent_id"`
	Sort       int        `gorm:"default:0;comment:部门排序" json:"sort"`
	Leader     string     `gorm:"size:50;comment:部门负责人" json:"leader"`
	Phone      string     `gorm:"size:20;comment:联系电话" json:"phone"`
	Status     int8       `gorm:"default:1;comment:部门状态：0=停用，1=正常" json:"status"`
	CreateTime time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
	Children   []*SysDept `gorm:"-" json:"children,omitempty"`
}

func (SysDept) TableName() string {
	return "sys_dept"
}
