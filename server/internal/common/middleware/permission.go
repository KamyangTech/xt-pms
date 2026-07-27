package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
)

// PermissionCheckFunc 权限校验函数类型
// 由调用方传入具体的校验逻辑，避免循环依赖
type PermissionCheckFunc func(userID int64, permission string) bool

// getSuperAdminUserID 从配置获取超级管理员用户ID（内部辅助函数）
func getSuperAdminUserID() int64 {
	return config.GlobalConfig.Auth.SuperAdminUserID
}

// PermissionCheck 权限校验中间件（通用版本，依赖注入模式）
// 路由注册时：
//
//	permService := service.NewPermissionService()
//	auth.GET("/menus", middleware.PermissionCheck(permService.HasPermission), menuController.GetMenuTree)
//
// 同时需要在路由中设置 required_permission 上下文值：
//
//	c.Set("required_permission", "system:menu:list")
func PermissionCheck(checkFn PermissionCheckFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		requiredPerm, exists := c.Get("required_permission")
		if !exists {
			c.Next()
			return
		}

		userID := GetCurrentUserID(c)
		if userID == nil {
			response.Error(c, response.UNAUTHORIZED, "未登录")
			c.Abort()
			return
		}

		// 超级管理员跳过所有权限校验
		if *userID == getSuperAdminUserID() {
			c.Next()
			return
		}

		if !checkFn(*userID, requiredPerm.(string)) {
			response.Error(c, response.FORBIDDEN, "无权限访问")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequirePerm 快捷中间件：等效于 Set("required_permission", perm) + PermissionCheck(checkFn)
// 使用方式：auth.GET("/users", middleware.RequirePerm("system:user:list", permService.HasPermission), handler)
func RequirePerm(perm string, checkFn PermissionCheckFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetCurrentUserID(c)
		if userID == nil {
			response.Error(c, response.UNAUTHORIZED, "未登录")
			c.Abort()
			return
		}

		// 超级管理员跳过
		if *userID == getSuperAdminUserID() {
			c.Next()
			return
		}

		if !checkFn(*userID, perm) {
			response.Error(c, response.FORBIDDEN, "无权限访问")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnyPerm 满足任一权限即可放行（多个权限之间是 OR 关系）
// 适用于写操作需要同时接受对应 list 权限的场景：
//
//	middleware.RequireAnyPerm([]string{"mold:order:update", "mold:order:list"}, permService.HasPermission)
func RequireAnyPerm(perms []string, checkFn PermissionCheckFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetCurrentUserID(c)
		if userID == nil {
			response.Error(c, response.UNAUTHORIZED, "未登录")
			c.Abort()
			return
		}

		// 超级管理员跳过
		if *userID == getSuperAdminUserID() {
			c.Next()
			return
		}

		for _, perm := range perms {
			if checkFn(*userID, perm) {
				c.Next()
				return
			}
		}

		response.Error(c, response.FORBIDDEN, "无权限访问")
		c.Abort()
	}
}
