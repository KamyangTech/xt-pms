package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

type PermissionService struct {
	menuDAO     *dao.MenuDAO
	roleMenuDAO *dao.RoleMenuDAO
	cache       *utils.CacheService
}

var (
	permissionServiceInstance *PermissionService
	permissionServiceOnce     sync.Once
)

// NewPermissionService 创建权限服务实例（单例）
func NewPermissionService() *PermissionService {
	permissionServiceOnce.Do(func() {
		permissionServiceInstance = &PermissionService{
			menuDAO:     dao.NewMenuDAO(),
			roleMenuDAO: dao.NewRoleMenuDAO(),
			cache:       utils.GetCache(),
		}
	})
	return permissionServiceInstance
}

// GetUserPermissions 获取用户权限标识列表（去重）
func (s *PermissionService) GetUserPermissions(userID int64) ([]string, error) {
	cacheKey := fmt.Sprintf("user_permissions:%d", userID)

	// 先查缓存
	if cached, ok := s.cache.Get(cacheKey); ok {
		if perms, ok := cached.([]string); ok {
			return perms, nil
		}
	}

	superAdminID := config.GlobalConfig.Auth.SuperAdminUserID
	superAdminRoleType := config.GlobalConfig.Auth.SuperAdminRoleType

	// 用户 ID 为超级管理员，拥有全部权限
	if userID == superAdminID {
		return s.getAllPermissions(cacheKey)
	}

	// 查用户角色，并检查是否拥有超级管理员角色
	userDAO := dao.NewSysUserDAO()
	user, err := userDAO.FindByIDWithRoles(userID)
	if err != nil {
		return nil, err
	}

	roleIDs := make([]int64, 0, len(user.Roles))
	for _, role := range user.Roles {
		if role.RoleType == superAdminRoleType {
			return s.getAllPermissions(cacheKey)
		}
		roleIDs = append(roleIDs, role.ID)
	}

	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	// 批量查询角色关联菜单，替代循环 N+1
	menus, err := s.menuDAO.FindByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]struct{})
	for _, menu := range menus {
		if menu.Permission != "" {
			permSet[menu.Permission] = struct{}{}
		}
	}

	permissions := make([]string, 0, len(permSet))
	for perm := range permSet {
		permissions = append(permissions, perm)
	}

	// 写入缓存（10分钟过期）
	s.cache.Set(cacheKey, permissions, 10*time.Minute)

	return permissions, nil
}

// HasPermission 判断用户是否有某权限
func (s *PermissionService) HasPermission(userID int64, permission string) bool {
	// 超级管理员拥有所有权限（ID 匹配 或 拥有超级管理员角色）
	superAdminID := config.GlobalConfig.Auth.SuperAdminUserID
	superAdminRoleType := config.GlobalConfig.Auth.SuperAdminRoleType

	if userID == superAdminID {
		return true
	}

	if s.isSuperAdminByRole(userID, superAdminRoleType) {
		return true
	}

	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false
	}
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// isSuperAdminByRole 检查用户是否拥有超级管理员角色
func (s *PermissionService) isSuperAdminByRole(userID int64, roleType int8) bool {
	userDAO := dao.NewSysUserDAO()
	user, err := userDAO.FindByIDWithRoles(userID)
	if err != nil {
		return false
	}
	for _, role := range user.Roles {
		if role.RoleType == roleType {
			return true
		}
	}
	return false
}

// InvalidateUserCache 清除用户权限缓存
func (s *PermissionService) InvalidateUserCache(userID int64) {
	cacheKey := fmt.Sprintf("user_permissions:%d", userID)
	s.cache.Delete(cacheKey)
}

// GetUserMenuTree 获取用户可见的菜单树
func (s *PermissionService) GetUserMenuTree(userID int64) ([]*entity.SysMenu, error) {
	allMenus, err := s.menuDAO.FindAllWithStatus(1)
	if err != nil {
		return nil, err
	}

	superAdminID := config.GlobalConfig.Auth.SuperAdminUserID
	superAdminRoleType := config.GlobalConfig.Auth.SuperAdminRoleType

	// 超级管理员（ID匹配 或 拥有超级管理员角色）返回全部菜单
	if userID == superAdminID {
		return BuildMenuTree(allMenus), nil
	}

	userDAO := dao.NewSysUserDAO()
	user, err := userDAO.FindByIDWithRoles(userID)
	if err != nil {
		return nil, err
	}

	for _, role := range user.Roles {
		if role.RoleType == superAdminRoleType {
			return BuildMenuTree(allMenus), nil
		}
	}

	roleIDs := make([]int64, 0, len(user.Roles))
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
	}

	if len(roleIDs) == 0 {
		return BuildMenuTree(nil), nil
	}

	// 批量查询所有角色关联的菜单 ID，替代循环 N+1
	menuIDs, err := s.roleMenuDAO.FindMenuIDsByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	allMenuMap := make(map[int64]*entity.SysMenu, len(allMenus))
	for _, m := range allMenus {
		allMenuMap[m.ID] = m
	}

	menuIDSet := make(map[int64]bool, len(menuIDs))
	for _, mid := range menuIDs {
		menuIDSet[mid] = true
	}

	// 将没有直接分配的根也包含进来（用于路由导航）
	for menuID := range menuIDSet {
		ptr := menuID
		for ptr > 0 {
			menuIDSet[ptr] = true
			if parent, ok := allMenuMap[ptr]; ok {
				ptr = parent.ParentID
			} else {
				break
			}
		}
	}

	var userMenus []*entity.SysMenu
	for _, m := range allMenus {
		if menuIDSet[m.ID] {
			userMenus = append(userMenus, m)
		}
	}

	return BuildMenuTree(userMenus), nil
}

// GetUserPermissionsWithMenus 合并获取用户权限和菜单树（只需一次用户查询）
func (s *PermissionService) GetUserPermissionsWithMenus(userID int64) ([]string, []*entity.SysMenu, error) {
	// 先查权限缓存
	cacheKey := fmt.Sprintf("user_permissions:%d", userID)
	var cachedPerms []string
	if cached, ok := s.cache.Get(cacheKey); ok {
		if perms, ok := cached.([]string); ok {
			cachedPerms = perms
		}
	}

	allMenus, err := s.menuDAO.FindAllWithStatus(1)
	if err != nil {
		return nil, nil, err
	}

	superAdminID := config.GlobalConfig.Auth.SuperAdminUserID
	superAdminRoleType := config.GlobalConfig.Auth.SuperAdminRoleType

	// 超级管理员（ID匹配）返回全部权限和菜单
	if userID == superAdminID {
		permissions := cachedPerms
		if permissions == nil {
			permissions = make([]string, 0)
			for _, menu := range allMenus {
				if menu.Permission != "" {
					permissions = append(permissions, menu.Permission)
				}
			}
			s.cache.Set(cacheKey, permissions, 10*time.Minute)
		}
		return permissions, BuildMenuTree(allMenus), nil
	}

	// 一次性加载用户及角色
	userDAO := dao.NewSysUserDAO()
	user, err := userDAO.FindByIDWithRoles(userID)
	if err != nil {
		return nil, nil, err
	}

	var roleIDs []int64
	for _, role := range user.Roles {
		if role.RoleType == superAdminRoleType {
			permissions := cachedPerms
			if permissions == nil {
				permissions = make([]string, 0)
				for _, menu := range allMenus {
					if menu.Permission != "" {
						permissions = append(permissions, menu.Permission)
					}
				}
				s.cache.Set(cacheKey, permissions, 10*time.Minute)
			}
			return permissions, BuildMenuTree(allMenus), nil
		}
		roleIDs = append(roleIDs, role.ID)
	}

	// 计算权限（缓存未命中时）
	var permissions []string
	if cachedPerms != nil {
		permissions = cachedPerms
	} else if len(roleIDs) > 0 {
		menus, err := s.menuDAO.FindByRoleIDs(roleIDs)
		if err != nil {
			return nil, nil, err
		}
		permSet := make(map[string]struct{})
		for _, menu := range menus {
			if menu.Permission != "" {
				permSet[menu.Permission] = struct{}{}
			}
		}
		permissions = make([]string, 0, len(permSet))
		for perm := range permSet {
			permissions = append(permissions, perm)
		}
		s.cache.Set(cacheKey, permissions, 10*time.Minute)
	} else {
		permissions = []string{}
	}

	if len(roleIDs) == 0 {
		return permissions, BuildMenuTree(nil), nil
	}

	menuIDs, err := s.roleMenuDAO.FindMenuIDsByRoleIDs(roleIDs)
	if err != nil {
		return nil, nil, err
	}

	allMenuMap := make(map[int64]*entity.SysMenu, len(allMenus))
	for _, m := range allMenus {
		allMenuMap[m.ID] = m
	}

	menuIDSet := make(map[int64]bool, len(menuIDs))
	for _, mid := range menuIDs {
		menuIDSet[mid] = true
	}

	for menuID := range menuIDSet {
		ptr := menuID
		for ptr > 0 {
			menuIDSet[ptr] = true
			if parent, ok := allMenuMap[ptr]; ok {
				ptr = parent.ParentID
			} else {
				break
			}
		}
	}

	var userMenus []*entity.SysMenu
	for _, m := range allMenus {
		if menuIDSet[m.ID] {
			userMenus = append(userMenus, m)
		}
	}

	return permissions, BuildMenuTree(userMenus), nil
}

// getAllPermissions 获取所有菜单的权限标识（用于超级管理员）
func (s *PermissionService) getAllPermissions(cacheKey string) ([]string, error) {
	allMenus, err := s.menuDAO.FindAll()
	if err != nil {
		return nil, err
	}
	permSet := make(map[string]struct{})
	for _, menu := range allMenus {
		if menu.Permission != "" {
			permSet[menu.Permission] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for perm := range permSet {
		permissions = append(permissions, perm)
	}
	s.cache.Set(cacheKey, permissions, 10*time.Minute)
	return permissions, nil
}
