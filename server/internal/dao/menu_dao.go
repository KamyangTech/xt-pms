package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MenuDAO struct{}

// NewMenuDAO 创建菜单 DAO 实例
func NewMenuDAO() *MenuDAO {
	return &MenuDAO{}
}

// Create 创建菜单
func (d *MenuDAO) Create(menu *entity.SysMenu) error {
	return DB.Create(menu).Error
}

// FindByID 根据 ID 查询菜单
func (d *MenuDAO) FindByID(id int64) (*entity.SysMenu, error) {
	var menu entity.SysMenu
	err := DB.First(&menu, id).Error
	if err != nil {
		return nil, err
	}
	return &menu, nil
}

// FindAll 查询所有菜单（按排序字段排列）
func (d *MenuDAO) FindAll() ([]*entity.SysMenu, error) {
	var menus []*entity.SysMenu
	err := DB.Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

// FindAllWithStatus 按状态查询所有菜单
func (d *MenuDAO) FindAllWithStatus(status int8) ([]*entity.SysMenu, error) {
	var menus []*entity.SysMenu
	err := DB.Where("status = ?", status).Order("sort ASC, id ASC").Find(&menus).Error
	return menus, err
}

// FindByParentID 根据父 ID 查询菜单
func (d *MenuDAO) FindByParentID(parentID int64) ([]*entity.SysMenu, error) {
	var menus []*entity.SysMenu
	err := DB.Where("parent_id = ?", parentID).Order("sort ASC").Find(&menus).Error
	return menus, err
}

// FindByRoleID 查询某角色拥有的所有菜单
func (d *MenuDAO) FindByRoleID(roleID int64) ([]*entity.SysMenu, error) {
	var menus []*entity.SysMenu
	err := DB.Joins("JOIN sys_role_menu ON sys_role_menu.menu_id = sys_menu.id").
		Where("sys_role_menu.role_id = ?", roleID).
		Order("sys_menu.sort ASC, sys_menu.id ASC").
		Find(&menus).Error
	return menus, err
}

// FindByRoleIDs 批量查询多个角色拥有的所有菜单（已去重）
func (d *MenuDAO) FindByRoleIDs(roleIDs []int64) ([]*entity.SysMenu, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var menus []*entity.SysMenu
	err := DB.Joins("JOIN sys_role_menu ON sys_role_menu.menu_id = sys_menu.id").
		Where("sys_role_menu.role_id IN ?", roleIDs).
		Order("sys_menu.sort ASC, sys_menu.id ASC").
		Find(&menus).Error
	return menus, err
}

// Update 更新菜单
func (d *MenuDAO) Update(menu *entity.SysMenu) error {
	return DB.Model(&entity.SysMenu{}).Where("id = ?", menu.ID).Updates(map[string]interface{}{
		"parent_id":  menu.ParentID,
		"menu_name":  menu.MenuName,
		"menu_type":  menu.MenuType,
		"permission": menu.Permission,
		"path":       menu.Path,
		"component":  menu.Component,
		"icon":       menu.Icon,
		"sort":       menu.Sort,
		"status":     menu.Status,
	}).Error
}

// Delete 删除菜单
func (d *MenuDAO) Delete(id int64) error {
	return DB.Delete(&entity.SysMenu{}, id).Error
}

// CountByParentID 统计指定父 ID 下的子菜单数量
func (d *MenuDAO) CountByParentID(parentID int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.SysMenu{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}
