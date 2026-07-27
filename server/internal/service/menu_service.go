package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysMenuService struct {
	menuDAO     *dao.MenuDAO
	roleMenuDAO *dao.RoleMenuDAO
}

func NewSysMenuService() *SysMenuService {
	return &SysMenuService{
		menuDAO:     dao.NewMenuDAO(),
		roleMenuDAO: dao.NewRoleMenuDAO(),
	}
}

// GetMenuTree 获取树形菜单
func (s *SysMenuService) GetMenuTree() ([]*entity.SysMenu, error) {
	menus, err := s.menuDAO.FindAll()
	if err != nil {
		return nil, err
	}
	return BuildMenuTree(menus), nil
}

// GetMenuByID 根据 ID 获取菜单
func (s *SysMenuService) GetMenuByID(id int64) (*entity.SysMenu, error) {
	return s.menuDAO.FindByID(id)
}

// CreateMenu 创建菜单
func (s *SysMenuService) CreateMenu(menu *entity.SysMenu) (*entity.SysMenu, error) {
	err := s.menuDAO.Create(menu)
	if err != nil {
		return nil, err
	}
	return menu, nil
}

// UpdateMenu 更新菜单
func (s *SysMenuService) UpdateMenu(menu *entity.SysMenu) error {
	_, err := s.menuDAO.FindByID(menu.ID)
	if err != nil {
		return errors.New("菜单不存在")
	}
	return s.menuDAO.Update(menu)
}

// DeleteMenu 删除菜单
func (s *SysMenuService) DeleteMenu(id int64) error {
	count, err := s.menuDAO.CountByParentID(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该菜单下存在子菜单，无法删除")
	}
	return s.menuDAO.Delete(id)
}

func BuildMenuTree(menus []*entity.SysMenu) []*entity.SysMenu {
	if len(menus) == 0 {
		return menus
	}

	menuMap := make(map[int64]*entity.SysMenu, len(menus))
	for _, menu := range menus {
		menu.Children = nil
		menuMap[menu.ID] = menu
	}

	var roots []*entity.SysMenu
	for _, menu := range menus {
		if menu.ParentID == 0 {
			roots = append(roots, menu)
		} else if parent, ok := menuMap[menu.ParentID]; ok {
			parent.Children = append(parent.Children, menu)
		} else {
			// 父菜单不存在，作为根节点处理
			roots = append(roots, menu)
		}
	}

	return roots
}
