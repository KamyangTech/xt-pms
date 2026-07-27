package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type MenuController struct {
	menuService *service.SysMenuService
}

func NewMenuController() *MenuController {
	return &MenuController{
		menuService: service.NewSysMenuService(),
	}
}

// GetMenuTree 获取树形菜单
func (mc *MenuController) GetMenuTree(c *gin.Context) {
	tree, err := mc.menuService.GetMenuTree()
	if err != nil {
		response.Error(c, response.ERROR, "获取菜单树失败")
		return
	}
	response.Success(c, gin.H{
		"list":  tree,
		"total": len(tree),
	})
}

// GetMenuByID 根据 ID 获取菜单
func (mc *MenuController) GetMenuByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return mc.menuService.GetMenuByID(id)
	})
}

// CreateMenu 创建菜单
func (mc *MenuController) CreateMenu(c *gin.Context) {
	c.Set("oper_log_title", "创建菜单")
	c.Set("oper_log_type", 1) // 1=新增
	c.Set("oper_log_business_type", "menu")

	var menuDTO dto.MenuDTO
	if err := c.ShouldBindJSON(&menuDTO); err != nil {
		response.Error(c, response.ERROR, "请求参数格式错误")
		return
	}

	menu := &entity.SysMenu{
		ParentID:   menuDTO.ParentID,
		MenuName:   menuDTO.MenuName,
		MenuType:   menuDTO.MenuType,
		Permission: menuDTO.Permission,
		Path:       menuDTO.Path,
		Component:  menuDTO.Component,
		Icon:       menuDTO.Icon,
		Sort:       menuDTO.Sort,
		Status:     menuDTO.Status,
	}

	created, err := mc.menuService.CreateMenu(menu)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, created)
}

// UpdateMenu 更新菜单（含权限标识变更）
func (mc *MenuController) UpdateMenu(c *gin.Context) {
	c.Set("oper_log_title", "更新菜单")
	c.Set("oper_log_type", 3) // 3=修改
	c.Set("oper_log_business_type", "menu")

	HandleUpdate(c, func(req *menuUpdateRequest) error {
		menu := &entity.SysMenu{
			ID:         req.ID,
			ParentID:   req.ParentID,
			MenuName:   req.MenuName,
			MenuType:   req.MenuType,
			Permission: req.Permission,
			Path:       req.Path,
			Component:  req.Component,
			Icon:       req.Icon,
			Sort:       req.Sort,
			Status:     req.Status,
		}
		return mc.menuService.UpdateMenu(menu)
	})
}

type menuUpdateRequest struct {
	ID         int64  `json:"id" binding:"required"`
	ParentID   int64  `json:"parent_id"`
	MenuName   string `json:"menu_name" binding:"required"`
	MenuType   int8   `json:"menu_type"`
	Permission string `json:"permission"`
	Path       string `json:"path"`
	Component  string `json:"component"`
	Icon       string `json:"icon"`
	Sort       int    `json:"sort"`
	Status     int8   `json:"status"`
}

// DeleteMenu 删除菜单
func (mc *MenuController) DeleteMenu(c *gin.Context) {
	c.Set("oper_log_title", "删除菜单")
	c.Set("oper_log_type", 2) // 2=删除
	c.Set("oper_log_business_type", "menu")

	HandleDelete(c, func(id int64) error {
		return mc.menuService.DeleteMenu(id)
	})
}
