package controller

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: service.NewUserService(),
	}
}

// Login 用户登录
func (u *UserController) Login(c *gin.Context) {
	HandleRequest(c, func(loginDTO *dto.LoginDTO) (interface{}, error) {
		token, userInfo, err := u.userService.Login(*loginDTO)
		if err != nil {
			// 登录失败是客户端错误（用户名密码错误、验证码错误等），返回 401 而不是 500
			return nil, &response.HttpError{
				Code:    response.UNAUTHORIZED,
				Message: err.Error(),
			}
		}

		// 设置 HttpOnly Cookie（更安全）
		secCfgSvc := service.GetSecurityConfigService()
		domain := ""
		if config.GlobalConfig.App.Mode == "development" {
			domain = "localhost"
		}
		c.SetCookie(
			"pms_token",                 // Cookie 名称
			token,                       // Cookie 值
			86400,                       // 过期时间（秒）：24 小时
			"/",                         // 路径
			domain,                      // 域：开发环境 localhost，生产环境留空（自动适配当前域名）
			secCfgSvc.GetCookieSecure(), // secure: 由系统设置→登录与安全 控制，HTTPS部署可开启
			true,                        // httpOnly: 防止 JavaScript 访问
		)

		return gin.H{"token": token, "userInfo": userInfo}, nil
	})
}

// GetCaptcha 获取数学运算验证码
func (u *UserController) GetCaptcha(c *gin.Context) {
	captchaID, b64s, err := service.GenerateCaptcha()
	if err != nil {
		response.Error(c, response.ERROR, "生成验证码失败")
		return
	}
	response.Success(c, gin.H{"captcha_id": captchaID, "captcha_img": b64s})
}

// GetUsers 获取用户列表（支持分页、搜索和数据权限过滤）
func (u *UserController) GetUsers(c *gin.Context) {
	// 获取分页参数，默认第1页，每页20条
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	username := c.Query("username")
	realName := c.Query("real_name")

	var status *int8
	if s := c.Query("status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			sv := int8(v)
			status = &sv
		}
	}
	var deptID *int64
	if d := c.Query("dept_id"); d != "" {
		if v, err := strconv.ParseInt(d, 10, 64); err == nil {
			deptID = &v
		}
	}

	currentUserID := middleware.GetUserID(c)

	result, err := u.userService.GetUsersByPage(page, pageSize, username, realName, status, deptID, currentUserID)
	if err != nil {
		response.Error(c, response.ERROR, "获取用户列表失败")
		return
	}

	response.Success(c, result)
}

// BatchUpdateStatus 批量更新用户状态
func (u *UserController) BatchUpdateStatus(c *gin.Context) {
	var batchDTO dto.BatchStatusDTO
	if err := c.ShouldBindJSON(&batchDTO); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := u.userService.BatchUpdateStatus(batchDTO.IDs, batchDTO.Status)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// ResetPassword 管理员重置用户密码
func (u *UserController) ResetPassword(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	var pwdDTO dto.ResetPasswordDTO
	if err := c.ShouldBindJSON(&pwdDTO); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err = u.userService.ResetPassword(id, pwdDTO)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// UnlockUser 管理员手动解锁账号
func (u *UserController) UnlockUser(c *gin.Context) {
	c.Set("oper_log_title", "解锁账号")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "user")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}
	c.Set("oper_log_business_id", id)

	err = u.userService.UnlockUser(id)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetProfile 获取当前用户信息
func (u *UserController) GetProfile(c *gin.Context) {
	// 从 JWT 中获取用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}

	// 类型转换：JWT claims 中的数字会被解析为 float64
	var userIdInt64 int64
	switch v := userID.(type) {
	case float64:
		userIdInt64 = int64(v)
	case int64:
		userIdInt64 = v
	default:
		response.Error(c, response.BAD_REQUEST, "用户 ID 格式错误")
		return
	}

	user, err := u.userService.GetUserByID(userIdInt64)
	if err != nil {
		response.Error(c, response.NOT_FOUND, "用户不存在")
		return
	}

	// 直接返回用户对象，不要嵌套在 "user" 字段中
	response.Success(c, user)
}

// UpdateProfile 更新当前用户信息
func (u *UserController) UpdateProfile(c *gin.Context) {
	// 从 JWT 中获取用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}

	// 类型转换：JWT claims 中的数字会被解析为 float64
	var userIdInt64 int64
	switch v := userID.(type) {
	case float64:
		userIdInt64 = int64(v)
	case int64:
		userIdInt64 = v
	default:
		response.Error(c, response.BAD_REQUEST, "用户 ID 格式错误")
		return
	}

	var profileDTO dto.ProfileUpdateDTO
	if err := c.ShouldBindJSON(&profileDTO); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	profileDTO.ID = userIdInt64

	err := u.userService.UpdateProfile(profileDTO)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// ChangePassword 修改密码
func (u *UserController) ChangePassword(c *gin.Context) {
	// 从 JWT 中获取用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}

	// 类型转换：JWT claims 中的数字会被解析为 float64
	var userIdInt64 int64
	switch v := userID.(type) {
	case float64:
		userIdInt64 = int64(v)
	case int64:
		userIdInt64 = v
	default:
		response.Error(c, response.BAD_REQUEST, "用户 ID 格式错误")
		return
	}

	var pwdDTO dto.ChangePasswordDTO
	if err := c.ShouldBindJSON(&pwdDTO); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := u.userService.ChangePassword(userIdInt64, pwdDTO)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// CreateUser 创建用户
func (u *UserController) CreateUser(c *gin.Context) {
	c.Set("oper_log_title", "创建用户")
	c.Set("oper_log_type", 1) // 1=新增
	c.Set("oper_log_business_type", "user")

	HandleRequest(c, func(userDTO *dto.UserDTO) (interface{}, error) {
		return u.userService.CreateUser(*userDTO)
	})
}

// GetUserByID 根据 ID 获取用户
func (u *UserController) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	user, err := u.userService.GetUserByID(id)
	if err != nil {
		response.Error(c, response.NOT_FOUND, "用户不存在")
		return
	}

	response.Success(c, user)
}

// UpdateUser 更新用户（含角色分配）
func (u *UserController) UpdateUser(c *gin.Context) {
	var userDTO dto.UserUpdateDTO
	if err := c.ShouldBindJSON(&userDTO); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误")
		return
	}

	c.Set("oper_log_title", "更新用户")
	c.Set("oper_log_type", 3) // 3=修改
	c.Set("oper_log_business_type", "user")
	c.Set("oper_log_business_id", userDTO.ID)

	err := u.userService.UpdateUser(userDTO)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteUser 删除用户
func (u *UserController) DeleteUser(c *gin.Context) {
	c.Set("oper_log_title", "删除用户")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "user")

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	err = u.userService.DeleteUser(id)
	if err != nil {
		response.Error(c, response.ERROR, "删除失败")
		return
	}

	response.Success(c, nil)
}

// GetUserStatistics 获取当前用户的统计数据
func (u *UserController) GetUserStatistics(c *gin.Context) {
	userIDPtr := middleware.GetCurrentUserID(c)
	if userIDPtr == nil {
		response.Error(c, response.UNAUTHORIZED, "未登录")
		return
	}
	userID := *userIDPtr

	// 本月起止时间
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	var moldOrderCount int64
	var prodOrderCount int64

	// 查询当前用户本月新增的模具订单
	moldOrderCount, _ = u.userService.CountUserMoldOrders(userID, startOfMonth, endOfMonth)

	// 查询当前用户本月新增的生产订单
	dao.DB.Model(&entity.ProductionOrder{}).
		Where("create_user_id = ? AND create_time >= ? AND create_time < ?", userID, startOfMonth, endOfMonth).
		Count(&prodOrderCount)

	response.Success(c, gin.H{
		"mold_order_count":     moldOrderCount,
		"production_order_count": prodOrderCount,
	})
}
