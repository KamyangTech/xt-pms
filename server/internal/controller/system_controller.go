package controller

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/job"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type CleanupConfigRequest struct {
	RetentionDays int `json:"retention_days"`
}

type SystemController struct {
	roleService               *service.SysRoleService
	deptService               *service.SysDeptService
	dictTypeService           *service.SysDictTypeService
	dictDataService           *service.SysDictDataService
	printNoticeConfigService  *service.PrintNoticeConfigService
	receivingAccountService   *service.ReceivingAccountService
	systemInfoService         *service.SysSystemInfoService
	securityConfigService     *service.SecurityConfigService
	permService               *service.PermissionService
	roleMenuDAO               *dao.RoleMenuDAO
}

func NewSystemController() *SystemController {
	return &SystemController{
		roleService:               service.NewSysRoleService(),
		deptService:               service.NewSysDeptService(),
		dictTypeService:           service.NewSysDictTypeService(),
		dictDataService:           service.NewSysDictDataService(),
		printNoticeConfigService:  service.NewPrintNoticeConfigService(),
		receivingAccountService:   service.NewReceivingAccountService(),
		systemInfoService:         service.NewSysSystemInfoService(),
		securityConfigService:     service.GetSecurityConfigService(),
		permService:               service.NewPermissionService(),
		roleMenuDAO:               dao.NewRoleMenuDAO(),
	}
}

// GetRoles 获取角色列表
func (sc *SystemController) GetRoles(c *gin.Context) {
	roles, err := sc.roleService.GetAllRoles()
	if err != nil {
		response.Error(c, response.ERROR, "获取角色列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  roles,
		"total": len(roles),
	})
}

// GetRoleByID 根据 ID 获取角色
func (sc *SystemController) GetRoleByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return sc.roleService.GetRoleByID(id)
	})
}

// CreateRole 创建角色
func (sc *SystemController) CreateRole(c *gin.Context) {
	c.Set("oper_log_title", "创建角色")
	c.Set("oper_log_type", 1) // 1=新增
	c.Set("oper_log_business_type", "role")

	HandleCreate(c, func(role *entity.SysRole) (interface{}, error) {
		return sc.roleService.CreateRole(role)
	})
}

// UpdateRole 更新角色（含数据权限变更）
func (sc *SystemController) UpdateRole(c *gin.Context) {
	c.Set("oper_log_title", "更新角色")
	c.Set("oper_log_type", 3) // 3=修改
	c.Set("oper_log_business_type", "role")

	HandleUpdate(c, func(role *entity.SysRole) error {
		return sc.roleService.UpdateRole(role)
	})
}

// DeleteRole 删除角色
func (sc *SystemController) DeleteRole(c *gin.Context) {
	c.Set("oper_log_title", "删除角色")
	c.Set("oper_log_type", 2) // 2=删除
	c.Set("oper_log_business_type", "role")

	HandleDelete(c, func(id int64) error {
		return sc.roleService.DeleteRole(id)
	})
}

// GetRoleMenus 获取角色已分配的菜单 ID 列表
func (sc *SystemController) GetRoleMenus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}
	menuIDs, err := sc.roleMenuDAO.FindMenuIDsByRoleID(id)
	if err != nil {
		response.Error(c, response.ERROR, "获取角色菜单失败")
		return
	}
	response.Success(c, gin.H{"menu_ids": menuIDs})
}

// AssignRoleMenus 为角色分配菜单
type assignRoleMenusReq struct {
	RoleID  int64   `json:"role_id" binding:"required"`
	MenuIDs []int64 `json:"menu_ids"`
}

func (sc *SystemController) AssignRoleMenus(c *gin.Context) {
	var req assignRoleMenusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	c.Set("oper_log_title", "分配菜单权限")
	c.Set("oper_log_type", 3) // 3=修改
	c.Set("oper_log_business_type", "role_menu")

	// 超级管理员角色不可分配菜单权限（其权限全量固定）
	role, err := sc.roleService.GetRoleByID(req.RoleID)
	if err != nil {
		response.Error(c, response.NOT_FOUND, "角色不存在")
		return
	}
	if role.RoleType == entity.RoleTypeSuperAdmin {
		response.Error(c, response.ERROR, "超级管理员角色无需分配菜单权限")
		return
	}

	err = sc.roleMenuDAO.AssignMenus(req.RoleID, req.MenuIDs)
	if err != nil {
		response.Error(c, response.ERROR, "分配菜单权限失败")
		return
	}

	// 同步清除该角色下所有用户的权限缓存（确保在响应前完成）
	userIDs, cacheErr := dao.DB.Table("sys_user_role").Where("role_id = ?", req.RoleID).Pluck("user_id", &[]int64{}).Rows()
	if cacheErr == nil && userIDs != nil {
		var uids []int64
		for userIDs.Next() {
			var uid int64
			if scanErr := userIDs.Scan(&uid); scanErr == nil {
				sc.permService.InvalidateUserCache(uid)
				uids = append(uids, uid)
			}
		}
		userIDs.Close()
		logger.Info("[权限缓存] 分配菜单后同步清除用户缓存", zap.Int64s("user_ids", uids), zap.Int64("role_id", req.RoleID))
	} else if cacheErr != nil {
		logger.Warn("[权限缓存] 清除用户权限缓存时查询失败", zap.Error(cacheErr), zap.Int64("role_id", req.RoleID))
	}

	response.Success(c, nil)
}

// GetUserPermissions 获取当前用户的权限标识列表和菜单树
func (sc *SystemController) GetUserPermissions(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == nil {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}

	permissions, menuTree, err := sc.permService.GetUserPermissionsWithMenus(*userID)
	if err != nil {
		response.Error(c, response.ERROR, "获取用户权限失败")
		return
	}

	response.Success(c, gin.H{
		"permissions": permissions,
		"menus":       menuTree,
	})
}

// GetDepts 获取部门列表（树形结构）
func (sc *SystemController) GetDepts(c *gin.Context) {
	depts, err := sc.deptService.GetAllDepts()
	if err != nil {
		response.Error(c, response.ERROR, "获取部门列表失败")
		return
	}

	tree := service.BuildDeptTree(depts)

	response.Success(c, gin.H{
		"list":  tree,
		"total": len(tree),
	})
}

// GetDeptByID 根据 ID 获取部门
func (sc *SystemController) GetDeptByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return sc.deptService.GetDeptByID(id)
	})
}

// CreateDept 创建部门
func (sc *SystemController) CreateDept(c *gin.Context) {
	c.Set("oper_log_title", "创建部门")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "dept")
	HandleCreate(c, func(dept *entity.SysDept) (interface{}, error) {
		return sc.deptService.CreateDept(dept)
	})
}

// UpdateDept 更新部门
func (sc *SystemController) UpdateDept(c *gin.Context) {
	c.Set("oper_log_title", "更新部门")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "dept")
	HandleUpdate(c, func(dept *entity.SysDept) error {
		return sc.deptService.UpdateDept(dept)
	})
}

// DeleteDept 删除部门
func (sc *SystemController) DeleteDept(c *gin.Context) {
	c.Set("oper_log_title", "删除部门")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "dept")
	HandleDelete(c, func(id int64) error {
		return sc.deptService.DeleteDept(id)
	})
}

// UpdateDeptStatus 更新部门状态
func (sc *SystemController) UpdateDeptStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	var req struct {
		Status int8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	if req.Status != 0 && req.Status != 1 {
		response.Error(c, response.ERROR, "状态值无效")
		return
	}

	if err := sc.deptService.UpdateDeptStatus(id, req.Status); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// ==================== 字典版本查询 ====================

// GetDictVersion 获取字典版本号（用于前端版本轮询检测变更）
func (sc *SystemController) GetDictVersion(c *gin.Context) {
	response.Success(c, gin.H{
		"version": service.GetDictVersion(),
	})
}

// GetDictTypes 获取字典类型列表（支持分页和筛选，过滤受保护的字典类型）
func (sc *SystemController) GetDictTypes(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	dictName := c.Query("dictName")
	dictCode := c.Query("dictCode")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 如果有筛选条件，走带筛选的分页查询；否则走原有逻辑
	if dictName != "" || dictCode != "" {
		list, total, err := sc.dictTypeService.GetPageWithFilter(page, pageSize, dictName, dictCode)
		if err != nil {
			response.Error(c, response.ERROR, "获取字典类型列表失败")
			return
		}
		response.Success(c, gin.H{
			"list":  list,
			"total": total,
		})
		return
	}

	dictTypes, err := sc.dictTypeService.GetList()
	if err != nil {
		response.Error(c, response.ERROR, "获取字典类型列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  dictTypes,
		"total": len(dictTypes),
	})
}

// GetDictTypeByID 根据 ID 获取字典类型
func (sc *SystemController) GetDictTypeByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return sc.dictTypeService.GetByID(id)
	})
}

// CreateDictType 创建字典类型
func (sc *SystemController) CreateDictType(c *gin.Context) {
	c.Set("oper_log_title", "创建字典类型")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "dict_type")
	HandleCreate(c, func(dictType *entity.SysDictType) (interface{}, error) {
		return nil, sc.dictTypeService.Create(dictType)
	})
}

// UpdateDictType 更新字典类型
func (sc *SystemController) UpdateDictType(c *gin.Context) {
	c.Set("oper_log_title", "更新字典类型")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "dict_type")
	HandleUpdate(c, func(dictType *entity.SysDictType) error {
		return sc.dictTypeService.Update(dictType)
	})
}

// DeleteDictType 删除字典类型
func (sc *SystemController) DeleteDictType(c *gin.Context) {
	c.Set("oper_log_title", "删除字典类型")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "dict_type")
	HandleDelete(c, func(id int64) error {
		return sc.dictTypeService.Delete(id)
	})
}

// ==================== 字典数据管理 ====================

// GetDictDataByType 根据字典类型编码获取字典数据列表
func (sc *SystemController) GetDictDataByType(c *gin.Context) {
	dictCode := c.Query("dict_type")
	if dictCode == "" {
		response.Error(c, response.ERROR, "字典类型编码不能为空")
		return
	}

	dictData, err := sc.dictDataService.GetByTypeCode(dictCode)
	if err != nil {
		response.Error(c, response.ERROR, "获取字典数据列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  dictData,
		"total": len(dictData),
	})
}

// GetDictDataByTypeID 根据字典类型 ID 获取字典数据列表
func (sc *SystemController) GetDictDataByTypeID(c *gin.Context) {
	idStr := c.Query("dict_type_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	dictData, err := sc.dictDataService.GetByTypeID(id)
	if err != nil {
		response.Error(c, response.ERROR, "获取字典数据列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  dictData,
		"total": len(dictData),
	})
}

// GetDictDataPage 分页获取字典数据列表
func (sc *SystemController) GetDictDataPage(c *gin.Context) {
	dictTypeIDStr := c.Query("dict_type_id")

	var dictTypeID int64 = 0
	if dictTypeIDStr != "" {
		var err error
		dictTypeID, err = strconv.ParseInt(dictTypeIDStr, 10, 64)
		if err != nil {
			response.Error(c, response.BAD_REQUEST, "参数错误")
			return
		}
	}

	HandlePageQuery(c, func(page, pageSize int) (interface{}, int64, error) {
		return sc.dictDataService.GetPage(dictTypeID, page, pageSize)
	})
}

// UpdateDictData 更新字典数据（系统内置状态仅允许修改颜色和备注）
func (sc *SystemController) UpdateDictData(c *gin.Context) {
	var req struct {
		ID     int64  `json:"id"`
		Color  string `json:"color"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	if req.ID == 0 {
		response.Error(c, response.ERROR, "缺少ID参数")
		return
	}

	c.Set("oper_log_title", "更新字典数据显示属性")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "dict_data")

	// 先查询已有记录，避免零值覆盖其他字段（如 dict_type_id）
	existing, err := sc.dictDataService.GetByID(req.ID)
	if err != nil {
		response.Error(c, response.ERROR, "未找到该字典数据")
		return
	}

	existing.Color = req.Color
	existing.Remark = req.Remark

	if err := sc.dictDataService.Update(existing); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, existing)
}

// CreateDictData 创建字典数据（业务配置使用，支持全字段）
func (sc *SystemController) CreateDictData(c *gin.Context) {
	var req struct {
		DictType            string `json:"dict_type"`
		DictLabel           string `json:"dict_label"`
		DictValue           string `json:"dict_value"`
		Color               string `json:"color"`
		Sort                int    `json:"sort"`
		Status              int8   `json:"status"`
		NeedsDeliveryUnlock int8   `json:"needs_delivery_unlock"`
		Remark              string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	if req.DictType == "" {
		response.Error(c, response.ERROR, "字典类型编码不能为空")
		return
	}

	c.Set("oper_log_title", "创建字典数据")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "dict_data")

	// 根据字典类型编码查询 ID
	dictType, err := sc.dictTypeService.GetByCode(req.DictType)
	if err != nil {
		response.Error(c, response.NOT_FOUND, "字典类型不存在")
		return
	}

	dictData := &entity.SysDictData{
		DictTypeID:          dictType.ID,
		DictLabel:           req.DictLabel,
		DictValue:           req.DictValue,
		Color:               req.Color,
		Sort:                req.Sort,
		Status:              req.Status,
		NeedsDeliveryUnlock: req.NeedsDeliveryUnlock,
		Remark:              req.Remark,
	}

	if err := sc.dictDataService.Create(dictData); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, dictData)
}

// UpdateBusinessDictData 更新业务字典数据（全字段，用于业务配置页面）
func (sc *SystemController) UpdateBusinessDictData(c *gin.Context) {
	var req struct {
		ID                  int64  `json:"id"`
		DictLabel           string `json:"dict_label"`
		DictValue           string `json:"dict_value"`
		Color               string `json:"color"`
		Status              int8   `json:"status"`
		NeedsDeliveryUnlock int8   `json:"needs_delivery_unlock"`
		Remark              string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	if req.ID == 0 {
		response.Error(c, response.ERROR, "缺少ID参数")
		return
	}

	c.Set("oper_log_title", "更新业务字典数据")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "dict_data")

	existing, err := sc.dictDataService.GetByID(req.ID)
	if err != nil {
		response.Error(c, response.ERROR, "未找到该字典数据")
		return
	}

	dictType, err := sc.dictTypeService.GetByID(existing.DictTypeID)
	if err == nil && dictType != nil {
		if service.IsProtectedDictType(dictType.DictCode) {
			// 受保护类型：禁止修改 dict_value 和 status，但允许修改其他属性
			if req.DictValue != existing.DictValue {
				response.Error(c, response.ERROR, fmt.Sprintf("「%s」的字典值由系统固件控制，不可修改", service.GetProtectedDictTypeName(dictType.DictCode)))
				return
			}
			if req.Status != existing.Status {
				response.Error(c, response.ERROR, fmt.Sprintf("「%s」的启用状态由系统固件控制，不可修改", service.GetProtectedDictTypeName(dictType.DictCode)))
				return
			}
		} else if service.IsProtectedDictDataValue(dictType.DictCode, existing.DictValue) {
			// 受保护字典值：该特定值被后端代码硬编码引用，不可修改
			if req.DictValue != existing.DictValue {
				response.Error(c, response.ERROR, fmt.Sprintf("「%s」字典值由系统默认值控制，不可修改", existing.DictLabel))
				return
			}
		}
	}

	existing.DictLabel = req.DictLabel
	existing.DictValue = req.DictValue
	existing.Color = req.Color
	existing.Status = req.Status
	existing.NeedsDeliveryUnlock = req.NeedsDeliveryUnlock
	existing.Remark = req.Remark

	if err := sc.dictDataService.Update(existing); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, existing)
}

// DeleteDictData 删除字典数据
func (sc *SystemController) DeleteDictData(c *gin.Context) {
	c.Set("oper_log_title", "删除字典数据")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "dict_data")

	HandleDelete(c, func(id int64) error {
		// 先查询字典数据，判断其所属字典类型是否受保护
		existing, err := sc.dictDataService.GetByID(id)
		if err != nil {
			return err
		}
		dictType, err := sc.dictTypeService.GetByID(existing.DictTypeID)
		if err == nil && dictType != nil && service.IsProtectedDictType(dictType.DictCode) {
			return fmt.Errorf("「%s」的字典值由系统固件控制，不可删除", service.GetProtectedDictTypeName(dictType.DictCode))
		}
		return sc.dictDataService.Delete(id)
	})
}

// GetPrintNoticeConfig 获取打印通知单配置
func (sc *SystemController) GetPrintNoticeConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}
	// 类型转换：JWT claims 中的数字会被解析为 float64
	var userIDInt64 int64
	switch v := userID.(type) {
	case float64:
		userIDInt64 = int64(v)
	case int64:
		userIDInt64 = v
	default:
		response.Error(c, response.ERROR, "用户ID类型错误")
		return
	}

	configType := c.DefaultQuery("config_type", "production_order")

	config, err := sc.printNoticeConfigService.GetConfig(userIDInt64, configType)
	if err != nil {
		response.Error(c, response.ERROR, "获取配置失败")
		return
	}

	response.Success(c, config)
}

// SavePrintNoticeConfig 保存打印配置（支持多种配置类型）
func (sc *SystemController) SavePrintNoticeConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}
	// 类型转换：JWT claims 中的数字会被解析为 float64
	var userIDInt64 int64
	switch v := userID.(type) {
	case float64:
		userIDInt64 = int64(v)
	case int64:
		userIDInt64 = v
	default:
		response.Error(c, response.ERROR, "用户ID类型错误")
		return
	}

	configType := c.DefaultQuery("config_type", "production_order")

	var configData map[string]interface{}
	if err := c.ShouldBindJSON(&configData); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := sc.printNoticeConfigService.SaveConfig(userIDInt64, configType, configData)
	if err != nil {
		response.Error(c, response.ERROR, "保存配置失败")
		return
	}

	response.Success(c, nil)
}

// ResetPrintNoticeConfig 重置打印通知单配置为默认值
func (sc *SystemController) ResetPrintNoticeConfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}
	// 类型转换：JWT claims 中的数字会被解析为 float64
	var userIDInt64 int64
	switch v := userID.(type) {
	case float64:
		userIDInt64 = int64(v)
	case int64:
		userIDInt64 = v
	default:
		response.Error(c, response.ERROR, "用户ID类型错误")
		return
	}

	configType := c.DefaultQuery("config_type", "production_order")

	err := sc.printNoticeConfigService.ResetToDefault(userIDInt64, configType)
	if err != nil {
		response.Error(c, response.ERROR, "重置配置失败")
		return
	}

	response.Success(c, nil)
}

// ClearDictCache 清除字典缓存
func (sc *SystemController) ClearDictCache(c *gin.Context) {
	cache := utils.GetCache()

	// 清除字典类型总缓存
	cache.Delete("dict_types_all")

	// 清除所有字典数据缓存
	dictCodes := []string{
		"mold-status",
		"work-order-status",
		"mold-order-status",
		"production-order-status",
		"mold-process",
		"material-type",
		"product-status",
		"inventory-status",
		"production-order-status",
	}

	for _, code := range dictCodes {
		cache.Delete("dict_data_" + code)
	}

	response.Success(c, gin.H{
		"message": "字典缓存已清除，请刷新页面",
	})
}

// RestoreDictSeed 恢复模具状态字典 & 流转配置为默认值
func (sc *SystemController) RestoreDictSeed(c *gin.Context) {
	c.Set("oper_log_title", "恢复状态字典默认配置")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "dict_seed")

	sqlBytes, err := os.ReadFile("migrations/data_seeds/001_dict_status_flow_seed.sql")
	if err != nil {
		logger.Error("[DictSeed] 读取种子文件失败", zap.Error(err))
		response.Error(c, response.ERROR, "读取种子文件失败")
		return
	}

	if !execSeedSQL(sqlBytes) {
		response.Error(c, response.ERROR, "恢复默认配置失败")
		return
	}

	// 清理相关字典缓存
	cache := utils.GetCache()
	cache.Delete("dict_types_all")
	for _, code := range []string{"mold-status", "work-order-status", "mold-order-status"} {
		cache.Delete("dict_data_" + code)
	}

	// 刷新 StatusFlowService 的内存缓存
	service.NewStatusFlowService().RefreshCache()

	response.Success(c, gin.H{
		"message": "已恢复默认配置",
	})
}

// RestoreProductionDictSeed 恢复生产订单状态字典 & 流转规则为默认值
func (sc *SystemController) RestoreProductionDictSeed(c *gin.Context) {
	c.Set("oper_log_title", "恢复生产状态字典默认配置")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "dict_seed")

	sqlBytes, err := os.ReadFile("migrations/data_seeds/002_production_order_status_seed.sql")
	if err != nil {
		logger.Error("[DictSeed] 读取种子文件失败", zap.Error(err))
		response.Error(c, response.ERROR, "读取种子文件失败")
		return
	}

	if !execSeedSQL(sqlBytes) {
		response.Error(c, response.ERROR, "恢复默认配置失败")
		return
	}

	// 清理相关字典缓存
	cache := utils.GetCache()
	cache.Delete("dict_types_all")
	cache.Delete("dict_data_production-order-status")

	response.Success(c, gin.H{
		"message": "已恢复默认配置",
	})
}

// execSeedSQL 执行种子 SQL，成功返回 true，失败返回 false（内部已处理错误响应）
func execSeedSQL(sqlBytes []byte) bool {
	tx := dao.GetDB().Begin()
	if tx.Error != nil {
		return false
	}

	statements := splitSQLStatements(string(sqlBytes))
	for _, stmt := range statements {
		if err := tx.Exec(stmt).Error; err != nil {
			tx.Rollback()
			logger.Error("[DictSeed] 执行种子 SQL 失败", zap.Error(err), zap.String("sql", stmt[:min(len(stmt), 200)]))
			return false
		}
	}

	if err := tx.Commit().Error; err != nil {
		return false
	}

	return true
}

// splitSQLStatements 将 SQL 文本按分号分割，过滤注释和空行
func splitSQLStatements(sql string) []string {
	lines := strings.Split(sql, "\n")
	var statements []string
	var current strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 跳过注释行和空行（但不跳过语句中间的行）
		if trimmed == "" || strings.HasPrefix(trimmed, "--") || strings.HasPrefix(trimmed, "#") {
			if current.Len() > 0 {
				current.WriteString("\n")
			}
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// 如果当前行以分号结束，且不是空语句，则作为一个完整语句
		if strings.HasSuffix(strings.TrimSpace(trimmed), ";") {
			stmt := strings.TrimSpace(current.String())
			if stmt != "" && !strings.HasPrefix(stmt, "--") {
				// 去掉末尾的分号（GORM Exec 不需要分号）
				stmt = strings.TrimSuffix(stmt, ";")
				stmt = strings.TrimSpace(stmt)
				if stmt != "" {
					statements = append(statements, stmt)
				}
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" && !strings.HasPrefix(stmt, "--") {
			stmt = strings.TrimSuffix(stmt, ";")
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				statements = append(statements, stmt)
			}
		}
	}

	return statements
}

// ==================== 系统信息配置 ====================

// GetSystemInfo 获取系统信息
func (sc *SystemController) GetSystemInfo(c *gin.Context) {
	info, err := sc.systemInfoService.Get()
	if err != nil {
		response.Error(c, response.ERROR, "获取系统信息失败")
		return
	}
	response.Success(c, info)
}

// UpdateSystemInfo 更新系统信息
func (sc *SystemController) UpdateSystemInfo(c *gin.Context) {
	var req struct {
		SystemName        string   `json:"system_name"`
		SystemVersion     string   `json:"system_version"`
		SystemDescription string   `json:"system_description"`
		TechStack         []string `json:"tech_stack"`
		FaviconURL        string   `json:"favicon_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	// 将技术栈数组转换为逗号分隔字符串
	techStack := strings.Join(req.TechStack, ",")

	info := &entity.SysSystemInfo{
		SystemName:        req.SystemName,
		SystemVersion:     req.SystemVersion,
		SystemDescription: req.SystemDescription,
		TechStack:         techStack,
		FaviconURL:        req.FaviconURL,
	}

	if err := sc.systemInfoService.Save(info); err != nil {
		response.Error(c, response.ERROR, "保存系统信息失败")
		return
	}

	response.Success(c, nil)
}

// UploadFavicon 上传 Favicon 图标（保存为 favicon.png/ico，覆盖默认）
func (sc *SystemController) UploadFavicon(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, response.ERROR, "获取上传文件失败")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".ico" && ext != ".png" {
		response.Error(c, response.ERROR, "仅支持 .ico 和 .png 格式")
		return
	}

	// 限制文件大小 1MB
	if file.Size > 1024*1024 {
		response.Error(c, response.ERROR, "文件大小不能超过 1MB")
		return
	}

	// 系统配置资源单独存放
	destDir := filepath.Join("system-assets")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		response.Error(c, response.ERROR, "创建目录失败")
		return
	}

	destPath := filepath.Join(destDir, "favicon"+ext)
	src, err := file.Open()
	if err != nil {
		response.Error(c, response.ERROR, "打开文件失败")
		return
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		response.Error(c, response.ERROR, "创建文件失败")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		response.Error(c, response.ERROR, "保存文件失败")
		return
	}

	urlPath := "/system-assets/favicon" + ext
	info, err := sc.systemInfoService.Get()
	if err == nil && info != nil {
		info.FaviconURL = urlPath
		sc.systemInfoService.Save(info)
	}

	response.Success(c, gin.H{
		"url": urlPath,
	})
}

// DeleteFavicon 删除用户自定义 Favicon，回到默认 logo.svg
func (sc *SystemController) DeleteFavicon(c *gin.Context) {
	for _, ext := range []string{".png", ".ico", ".svg"} {
		path := filepath.Join("system-assets", "favicon"+ext)
		os.Remove(path)
	}
	// 清除数据库中的 favicon_url
	info, err := sc.systemInfoService.Get()
	if err == nil && info != nil {
		info.FaviconURL = ""
		sc.systemInfoService.Save(info)
	}
	response.Success(c, nil)
}

// UploadCompanyLogo 上传公司 Logo（保存为 system-assets/company-logo.ext）
func (sc *SystemController) UploadCompanyLogo(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, response.ERROR, "获取上传文件失败")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".gif" && ext != ".svg" {
		response.Error(c, response.ERROR, "仅支持 png/jpg/gif/svg 格式")
		return
	}

	// 限制文件大小 5MB
	if file.Size > 5*1024*1024 {
		response.Error(c, response.ERROR, "文件大小不能超过 5MB")
		return
	}

	// 系统配置资源单独存放
	destDir := filepath.Join("system-assets")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		response.Error(c, response.ERROR, "创建目录失败")
		return
	}

	// 支持多扩展名，删除旧的 logo 文件后保存新的
	for _, oldExt := range []string{".png", ".jpg", ".jpeg", ".gif", ".svg"} {
		oldPath := filepath.Join(destDir, "company-logo"+oldExt)
		os.Remove(oldPath)
	}

	destPath := filepath.Join(destDir, "company-logo"+ext)
	src, err := file.Open()
	if err != nil {
		response.Error(c, response.ERROR, "打开文件失败")
		return
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		response.Error(c, response.ERROR, "创建文件失败")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		response.Error(c, response.ERROR, "保存文件失败")
		return
	}

	urlPath := "/system-assets/company-logo" + ext

	companyController := NewCompanyController()
	company, err := companyController.companyService.GetCompany()
	if err == nil && company != nil {
		company.Logo = urlPath
		companyController.companyService.UpdateCompany(company)
	}

	response.Success(c, gin.H{
		"url": urlPath,
	})
}

// DeleteCompanyLogo 删除公司 Logo，回到无 logo 状态
func (sc *SystemController) DeleteCompanyLogo(c *gin.Context) {
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".gif", ".svg"} {
		path := filepath.Join("system-assets", "company-logo"+ext)
		os.Remove(path)
	}
	// 清除公司信息中的 logo 字段
	companyController := NewCompanyController()
	company, err := companyController.companyService.GetCompany()
	if err == nil && company != nil {
		company.Logo = ""
		companyController.companyService.UpdateCompany(company)
	}
	response.Success(c, nil)
}

// ==================== 系统设置 ====================

// GetOrphanFileCount 获取孤文件数量（含数据库孤文件和磁盘无记录文件）
func (sc *SystemController) GetOrphanFileCount(c *gin.Context) {
	fileDAO := dao.NewFileDAO()
	dbCount, err := fileDAO.CountAllOrphanFiles()
	if err != nil {
		response.Error(c, response.ERROR, "查询孤文件数量失败")
		return
	}

	// 只返回数据库孤文件数量（磁盘孤文件由清理任务异步处理，避免 HTTP 阻塞）
	response.Success(c, gin.H{
		"count":      dbCount,
		"db_count":   dbCount,
		"disk_count": 0,
	})
}

// CleanupOrphanFiles 手动触发孤文件清理
func (sc *SystemController) CleanupOrphanFiles(c *gin.Context) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[文件清理] 清理协程发生 panic 已恢复", zap.Any("recover", r))
			}
		}()
		logger.Info("[文件清理] 收到手动清理请求，开始执行")
		job.CleanupOrphanFiles()
		logger.Info("[文件清理] 手动清理任务执行完毕")
	}()
	response.Success(c, gin.H{
		"message": "孤文件清理任务已启动，请稍后查看结果",
	})
}

// GetCleanupConfig 获取清理配置
func (sc *SystemController) GetCleanupConfig(c *gin.Context) {
	response.Success(c, gin.H{
		"retention_days": job.GetRetentionDays(),
	})
}

// UpdateCleanupConfig 更新清理配置
func (sc *SystemController) UpdateCleanupConfig(c *gin.Context) {
	var req CleanupConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}
	if req.RetentionDays < 1 {
		req.RetentionDays = 1
	}
	if req.RetentionDays > 365 {
		req.RetentionDays = 365
	}
	job.SetRetentionDays(req.RetentionDays)
	response.Success(c, gin.H{
		"retention_days": req.RetentionDays,
		"message":        "清理配置已更新",
	})
}

// ==================== 登录与安全配置 ====================

type SecurityConfigRequest struct {
	TokenExpire            int  `json:"token_expire"`
	MaxLoginAttempts       int  `json:"max_login_attempts"`
	LockDuration           int  `json:"lock_duration"`
	PasswordMinLen         int  `json:"password_min_len"`
	PasswordRequireUpper   bool `json:"password_require_upper"`
	PasswordRequireNumber  bool `json:"password_require_number"`
	PasswordRequireSpecial bool `json:"password_require_special"`
	CookieSecure           bool `json:"cookie_secure"`
	WebSocketSecure        bool `json:"websocket_secure"`
}

// GetSecurityConfig 获取安全配置
func (sc *SystemController) GetSecurityConfig(c *gin.Context) {
	cfg, err := sc.securityConfigService.Get()
	if err != nil {
		response.Error(c, response.ERROR, "获取安全配置失败")
		return
	}
	response.Success(c, cfg)
}

// UpdateSecurityConfig 更新安全配置
func (sc *SystemController) UpdateSecurityConfig(c *gin.Context) {
	var req SecurityConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	cfg := &entity.SysSecurityConfig{
		TokenExpire:            req.TokenExpire,
		MaxLoginAttempts:       req.MaxLoginAttempts,
		LockDuration:           req.LockDuration,
		PasswordMinLen:         req.PasswordMinLen,
		PasswordRequireUpper:   req.PasswordRequireUpper,
		PasswordRequireNumber:  req.PasswordRequireNumber,
		PasswordRequireSpecial: req.PasswordRequireSpecial,
		CookieSecure:           req.CookieSecure,
		WebSocketSecure:        req.WebSocketSecure,
	}

	if err := sc.securityConfigService.Update(cfg); err != nil {
		response.Error(c, response.ERROR, "更新安全配置失败")
		return
	}

	response.Success(c, gin.H{
		"message": "安全配置已更新",
	})
}

// ==================== 收款账户管理 ====================

// GetReceivingAccounts 获取收款账户列表（支持分页与筛选）
func (sc *SystemController) GetReceivingAccounts(c *gin.Context) {
	accountName := c.Query("account_name")
	accountNo := c.Query("account_no")
	bankName := c.Query("bank_name")
	accountType := c.Query("account_type")
	statusStr := c.DefaultQuery("status", "0")
	status, _ := strconv.Atoi(statusStr)

	HandlePageQuery(c, func(page, pageSize int) (interface{}, int64, error) {
		return sc.receivingAccountService.GetListByPage(page, pageSize, accountName, accountNo, bankName, accountType, status)
	})
}

// GetReceivingAccountByID 根据 ID 获取收款账户
func (sc *SystemController) GetReceivingAccountByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return sc.receivingAccountService.GetByID(id)
	})
}

// CreateReceivingAccount 创建收款账户
func (sc *SystemController) CreateReceivingAccount(c *gin.Context) {
	c.Set("oper_log_title", "创建收款账户")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "receiving_account")
	var req struct {
		AccountName string `json:"account_name" binding:"required"`
		AccountNo   string `json:"account_no" binding:"required"`
		BankName    string `json:"bank_name"`
		AccountType string `json:"account_type"`
		Sort        int    `json:"sort"`
		Status      int8   `json:"status"`
		Remark      string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	// 从 JWT 获取当前用户 ID
	var createdBy int64
	if userID, exists := c.Get("user_id"); exists {
		switch v := userID.(type) {
		case float64:
			createdBy = int64(v)
		case int64:
			createdBy = v
		}
	}

	account := &entity.ReceivingAccount{
		AccountName: req.AccountName,
		AccountNo:   req.AccountNo,
		BankName:    req.BankName,
		AccountType: req.AccountType,
		Sort:        req.Sort,
		Status:      req.Status,
		Remark:      req.Remark,
		CreatedBy:   createdBy,
	}

	if err := sc.receivingAccountService.Create(account); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, account)
}

// UpdateReceivingAccount 更新收款账户
func (sc *SystemController) UpdateReceivingAccount(c *gin.Context) {
	c.Set("oper_log_title", "更新收款账户")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "receiving_account")
	var req struct {
		ID          int64  `json:"id" binding:"required"`
		AccountName string `json:"account_name" binding:"required"`
		AccountNo   string `json:"account_no" binding:"required"`
		BankName    string `json:"bank_name"`
		AccountType string `json:"account_type"`
		Sort        int    `json:"sort"`
		Status      int8   `json:"status"`
		Remark      string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	account := &entity.ReceivingAccount{
		ID:          req.ID,
		AccountName: req.AccountName,
		AccountNo:   req.AccountNo,
		BankName:    req.BankName,
		AccountType: req.AccountType,
		Sort:        req.Sort,
		Status:      req.Status,
		Remark:      req.Remark,
	}

	if err := sc.receivingAccountService.Update(account); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, account)
}

// DeleteReceivingAccount 删除收款账户
func (sc *SystemController) DeleteReceivingAccount(c *gin.Context) {
	c.Set("oper_log_title", "删除收款账户")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "receiving_account")
	HandleDelete(c, func(id int64) error {
		return sc.receivingAccountService.Delete(id)
	})
}

// BatchDeleteReceivingAccounts 批量删除收款账户
func (sc *SystemController) BatchDeleteReceivingAccounts(c *gin.Context) {
	c.Set("oper_log_title", "批量删除收款账户")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "receiving_account")
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ERROR, "请求参数格式错误")
		return
	}
	if err := sc.receivingAccountService.BatchDelete(req.IDs); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, nil)
}

// UpdateReceivingAccountStatus 更新收款账户状态
func (sc *SystemController) UpdateReceivingAccountStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	var req struct {
		Status int8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	if req.Status != 0 && req.Status != 1 {
		response.Error(c, response.ERROR, "状态值无效")
		return
	}

	if err := sc.receivingAccountService.UpdateStatus(id, req.Status); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}
