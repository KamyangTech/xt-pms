package dao

import (
	"strings"

	"gorm.io/gorm"
)

type RoleMenuDAO struct{}

// NewRoleMenuDAO 创建角色菜单关联 DAO 实例
func NewRoleMenuDAO() *RoleMenuDAO {
	return &RoleMenuDAO{}
}

// AssignMenus 为角色分配菜单（先删除旧关联，再批量插入新关联）
func (d *RoleMenuDAO) AssignMenus(roleID int64, menuIDs []int64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 先删除旧的菜单关联
		if err := tx.Exec("DELETE FROM sys_role_menu WHERE role_id = ?", roleID).Error; err != nil {
			return err
		}

		// 批量插入新菜单关联（单条多值 SQL 替代逐条 INSERT）
		if len(menuIDs) > 0 {
			values := make([]string, 0, len(menuIDs))
			args := make([]interface{}, 0, len(menuIDs)*2)
			for _, menuID := range menuIDs {
				values = append(values, "(?, ?)")
				args = append(args, roleID, menuID)
			}
			sql := "INSERT INTO sys_role_menu (role_id, menu_id) VALUES " + strings.Join(values, ", ")
			if err := tx.Exec(sql, args...).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindMenuIDsByRoleID 查询角色拥有的菜单 ID 列表
func (d *RoleMenuDAO) FindMenuIDsByRoleID(roleID int64) ([]int64, error) {
	var menuIDs []int64
	err := DB.Table("sys_role_menu").Where("role_id = ?", roleID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

// FindMenuIDsByRoleIDs 批量查询多个角色拥有的菜单 ID（已去重）
func (d *RoleMenuDAO) FindMenuIDsByRoleIDs(roleIDs []int64) ([]int64, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var menuIDs []int64
	err := DB.Table("sys_role_menu").
		Where("role_id IN ?", roleIDs).
		Distinct("menu_id").
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

// DeleteByRoleID 删除角色的所有菜单关联
func (d *RoleMenuDAO) DeleteByRoleID(roleID int64) error {
	return DB.Exec("DELETE FROM sys_role_menu WHERE role_id = ?", roleID).Error
}
