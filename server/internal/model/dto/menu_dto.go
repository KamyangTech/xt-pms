package dto

// MenuDTO 菜单创建/更新 DTO
type MenuDTO struct {
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

// RoleMenuDTO 角色分配菜单 DTO
type RoleMenuDTO struct {
	RoleID  int64   `json:"role_id" binding:"required"`
	MenuIDs []int64 `json:"menu_ids"`
}
