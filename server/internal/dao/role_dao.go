package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysRoleDAO struct{}

// NewSysRoleDAO 创建角色 DAO 实例
func NewSysRoleDAO() *SysRoleDAO {
	return &SysRoleDAO{}
}

// Create 创建角色
func (d *SysRoleDAO) Create(role *entity.SysRole) error {
	return DB.Create(role).Error
}

// FindByID 根据 ID 查询角色
func (d *SysRoleDAO) FindByID(id int64) (*entity.SysRole, error) {
	var role entity.SysRole
	err := DB.First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// FindAll 查询所有角色
func (d *SysRoleDAO) FindAll() ([]*entity.SysRole, error) {
	var roles []*entity.SysRole
	err := DB.Order("id DESC").Find(&roles).Error
	return roles, err
}

// FindByName 根据名称查询角色
func (d *SysRoleDAO) FindByName(name string) (*entity.SysRole, error) {
	var role entity.SysRole
	err := DB.Where("role_name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Update 更新角色
func (d *SysRoleDAO) Update(role *entity.SysRole) error {
	return DB.Save(role).Error
}

// Delete 删除角色
func (d *SysRoleDAO) Delete(id int64) error {
	return DB.Delete(&entity.SysRole{}, id).Error
}
