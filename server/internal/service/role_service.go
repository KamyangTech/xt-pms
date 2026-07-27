package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysRoleService struct {
	roleDAO *dao.SysRoleDAO
}

func NewSysRoleService() *SysRoleService {
	return &SysRoleService{
		roleDAO: dao.NewSysRoleDAO(),
	}
}

// CreateRole 创建角色
func (s *SysRoleService) CreateRole(role *entity.SysRole) (*entity.SysRole, error) {
	existing, _ := s.roleDAO.FindByName(role.RoleName)
	if existing != nil {
		return nil, errors.New("角色名称已存在")
	}

	// 新创建的角色默认是自定义角色
	role.RoleType = entity.RoleTypeCustom

	err := s.roleDAO.Create(role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// GetRoleByID 根据 ID 获取角色
func (s *SysRoleService) GetRoleByID(id int64) (*entity.SysRole, error) {
	return s.roleDAO.FindByID(id)
}

// GetAllRoles 获取所有角色
func (s *SysRoleService) GetAllRoles() ([]*entity.SysRole, error) {
	return s.roleDAO.FindAll()
}

// UpdateRole 更新角色
func (s *SysRoleService) UpdateRole(role *entity.SysRole) error {
	existing, err := s.roleDAO.FindByID(role.ID)
	if err != nil {
		return errors.New("角色不存在")
	}

	// 超级管理员角色不可修改任何字段
	if existing.RoleType == entity.RoleTypeSuperAdmin {
		return errors.New("超级管理员角色不可编辑")
	}

	// 不允许通过更新接口修改角色类型
	role.RoleType = existing.RoleType

	return s.roleDAO.Update(role)
}

// DeleteRole 删除角色
func (s *SysRoleService) DeleteRole(id int64) error {
	existing, err := s.roleDAO.FindByID(id)
	if err != nil {
		return errors.New("角色不存在")
	}

	if existing.RoleType == entity.RoleTypeSuperAdmin {
		return errors.New("超级管理员角色不可删除")
	}
	if existing.RoleType == entity.RoleTypeBuiltin {
		return errors.New("业务内置角色不可删除")
	}

	return s.roleDAO.Delete(id)
}
