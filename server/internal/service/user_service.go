package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mojocn/base64Captcha"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// captchaAnswers 自定义验证码答案存储（id → answer）
var captchaAnswers = &sync.Map{}

type mathCaptchaDriver struct {
	*base64Captcha.DriverMath
}

// cryptoRandIntn 使用 crypto/rand 生成 [0, n) 范围的随机整数
func cryptoRandIntn(n int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		// 加密随机数生成失败属于系统级错误，无法继续
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return int(r.Int64())
}

// GenerateIdQuestionAnswer 生成个位数加减算式（×÷符号渲染不清，已移除）
func (d *mathCaptchaDriver) GenerateIdQuestionAnswer() (id, question, answer string) {
	a := cryptoRandIntn(9) + 1
	b := cryptoRandIntn(9) + 1

	if cryptoRandIntn(2) == 0 { // 加法
		answer = strconv.Itoa(a + b)
		question = fmt.Sprintf("%d + %d = ?", a, b)
	} else { // 减法（保证大减小）
		if a < b {
			a, b = b, a
		}
		answer = strconv.Itoa(a - b)
		question = fmt.Sprintf("%d - %d = ?", a, b)
	}

	return uuid.New().String(), question, answer
}

// GenerateCaptcha 生成验证码，返回 captcha_id 和 base64 图片
func GenerateCaptcha() (string, string, error) {
	driver := &mathCaptchaDriver{
		DriverMath: &base64Captcha.DriverMath{
			Height:          100,
			Width:           260,
			NoiseCount:      0,
			ShowLineOptions: 0,
		},
	}
	cp := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)

	var captchaID, b64s, answer string
	var err error
	for i := 0; i < 5; i++ {
		captchaID, b64s, answer, err = cp.Generate()
		if err == nil && answer != "" {
			break
		}
	}
	if err != nil {
		return "", "", err
	}

	// 自己存储 answer，不依赖 DefaultMemStore
	captchaAnswers.Store(captchaID, answer)
	// 5 分钟自动过期
	time.AfterFunc(5*time.Minute, func() {
		captchaAnswers.Delete(captchaID)
	})

	return captchaID, b64s, nil
}

// VerifyCaptcha 校验验证码（验证后自动删除，防止重放）
func VerifyCaptcha(captchaID, captchaVal string) bool {
	if captchaID == "" || captchaVal == "" {
		return false
	}
	val, ok := captchaAnswers.Load(captchaID)
	if !ok {
		return false
	}
	captchaAnswers.Delete(captchaID)
	return val.(string) == captchaVal
}

type UserService struct {
	userDAO           *dao.SysUserDAO
	securityCfgSvc    *SecurityConfigService
}

func NewUserService() *UserService {
	return &UserService{
		userDAO:        dao.NewSysUserDAO(),
		securityCfgSvc: GetSecurityConfigService(),
	}
}

// Login 用户登录（含防暴力破解：IP限流、验证码、账号锁定）
func (s *UserService) Login(loginDTO dto.LoginDTO) (string, map[string]interface{}, error) {
	// 0. 验证码检查（仅在提供了 captcha_id 时校验，首次失败前不强制）
	if loginDTO.CaptchaID != "" {
		if !VerifyCaptcha(loginDTO.CaptchaID, loginDTO.CaptchaVal) {
			return "", nil, errors.New("验证码错误")
		}
	}

	// 1. 查询用户
	user, err := s.userDAO.FindByUsername(loginDTO.Username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	// 2. 检查用户状态
	if user.Status != 1 {
		return "", nil, errors.New("用户已被禁用")
	}

	// 3. 检查账号是否被锁定
	secCfg, _ := s.securityCfgSvc.Get()
	if secCfg != nil && secCfg.MaxLoginAttempts > 0 {
		if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
			remaining := int(time.Until(*user.LockedUntil).Seconds())
			logger.Warn("登录失败：账号已被锁定",
				zap.String("username", loginDTO.Username),
				zap.Int("remaining_seconds", remaining),
			)
			return "", nil, fmt.Errorf("账号已被锁定，请 %d 秒后再试", remaining)
		}
		// 锁定期已过，自动重置
		if user.LockedUntil != nil && !user.LockedUntil.After(time.Now()) {
			_ = s.userDAO.ResetLoginAttempts(user.ID)
		}
	}

	// 4. 验证密码
	if !utils.VerifyPassword(user.Password, loginDTO.Password) {
		// 记录失败次数
		if secCfg != nil && secCfg.MaxLoginAttempts > 0 {
			_ = s.userDAO.IncrementLoginAttempts(user.ID)

			// 重新查询以获取最新的尝试次数
			updatedUser, _ := s.userDAO.FindByID(user.ID)
			if updatedUser != nil && updatedUser.LoginAttempts >= secCfg.MaxLoginAttempts {
				lockDuration := time.Duration(secCfg.LockDuration) * time.Second
				if lockDuration <= 0 {
					lockDuration = 900 * time.Second // 默认15分钟
				}
				lockedUntil := time.Now().Add(lockDuration)
				_ = s.userDAO.LockAccount(user.ID, lockedUntil)
				logger.Warn("账号已被锁定：登录失败次数超限",
					zap.String("username", loginDTO.Username),
					zap.Int("failed_attempts", updatedUser.LoginAttempts),
					zap.Int("max_attempts", secCfg.MaxLoginAttempts),
					zap.Time("locked_until", lockedUntil),
				)
			}
		}
		return "", nil, errors.New("用户名或密码错误")
	}

	// 5. 登录成功：重置失败计数（仅在配置了锁定机制时）
	if secCfg != nil && secCfg.MaxLoginAttempts > 0 {
		_ = s.userDAO.ResetLoginAttempts(user.ID)
	}

	// 6. 查询用户角色
	userWithRoles, err := s.userDAO.FindByIDWithRoles(user.ID)
	if err != nil {
		return "", nil, errors.New("查询用户角色失败")
	}

	// 7. 提取角色名称
	roleNames := make([]string, len(userWithRoles.Roles))
	for i, role := range userWithRoles.Roles {
		roleNames[i] = role.RoleName
	}

	// 8. 生成 token（附带密码版本号，用于密码修改后自动吊销旧 Token）
	token, err := middleware.GenerateToken(user.ID, user.Username, loginDTO.KeepLogin, user.PasswordVersion)
	if err != nil {
		return "", nil, errors.New("生成 token 失败")
	}

	// 9. 返回用户信息（用于前端存储）
	userInfo := map[string]interface{}{
		"user_id":   user.ID,
		"username":  user.Username,
		"real_name": user.RealName,
		"roles":     roleNames,
	}

	return token, userInfo, nil
}

// CreateUser 创建用户
func (s *UserService) CreateUser(userDTO dto.UserDTO) (*entity.SysUser, error) {
	existingUser, _ := s.userDAO.FindByUsername(userDTO.Username)
	if existingUser != nil {
		return nil, errors.New("用户名已存在")
	}

	// 加密密码
	hashedPassword, err := utils.HashPassword(userDTO.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	user := &entity.SysUser{
		Username: userDTO.Username,
		Password: string(hashedPassword),
		RealName: userDTO.RealName,
		Phone:    userDTO.Phone,
		Email:    userDTO.Email,
		Position: userDTO.Position,
		DeptID:   userDTO.DeptID,
		Status:   userDTO.Status,
	}

	if userDTO.HireDate != nil && userDTO.HireDate.Time != nil {
		user.HireDate = userDTO.HireDate.Time
	}

	if user.Status == 0 {
		user.Status = entity.UserStatusActive
	}

	err = s.userDAO.Create(user)
	if err != nil {
		return nil, err
	}

	// 分配角色
	if len(userDTO.RoleIDs) > 0 {
		err = s.userDAO.AssignRoles(user.ID, userDTO.RoleIDs)
		if err != nil {
			return nil, errors.New("分配角色失败")
		}
		// 清除用户权限缓存，使新分配的权限立即生效
		NewPermissionService().InvalidateUserCache(user.ID)
	}

	return user, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(id int64) (*entity.SysUser, error) {
	return s.userDAO.FindByIDWithRoles(id)
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() ([]*entity.SysUser, error) {
	return s.userDAO.FindAll()
}

// GetUsersByPage 分页获取用户（支持数据权限过滤）
func (s *UserService) GetUsersByPage(page, pageSize int, username, realName string, status *int8, deptID *int64, currentUserID int64) (*utils.PageResult, error) {
	// 查询当前用户的角色和数据范围
	currentUser, err := s.userDAO.FindByIDWithRoles(currentUserID)
	if err != nil {
		return s.userDAO.FindPage(page, pageSize, username, realName, status, deptID)
	}

	var roleScopes []int8
	for _, role := range currentUser.Roles {
		roleScopes = append(roleScopes, role.DataScope)
	}
	maxDataScope := utils.GetUserMaxDataScope(roleScopes)

	// 如果是全部数据权限，直接查询
	if maxDataScope == utils.DataScopeAll {
		return s.userDAO.FindPage(page, pageSize, username, realName, status, deptID)
	}

	// 根据数据范围构建过滤条件（在service层处理，避免循环依赖）
	var scopeCondition string
	var scopeParams []interface{}

	switch maxDataScope {
	case utils.DataScopeDeptAndChild:
		// 本部门及以下：查询所有子部门ID
		deptDAO := dao.NewSysDeptDAO()
		childDeptIDs, err := deptDAO.GetAllChildDeptIDs(currentUser.DeptID)
		if err != nil {
			return nil, fmt.Errorf("查询子部门失败: %w", err)
		}
		allDeptIDs := append(childDeptIDs, currentUser.DeptID)
		scopeCondition = "dept_id IN ?"
		scopeParams = []interface{}{allDeptIDs}

	case utils.DataScopeDept:
		// 仅本部门
		scopeCondition = "dept_id = ?"
		scopeParams = []interface{}{currentUser.DeptID}

	case utils.DataScopeSelf:
		// 仅本人
		scopeCondition = "id = ?"
		scopeParams = []interface{}{currentUserID}

	default:
		// 默认全部数据
		return s.userDAO.FindPage(page, pageSize, username, realName, status, deptID)
	}

	// 使用带数据范围的分页查询
	return s.userDAO.FindPageWithDataScope(page, pageSize, username, realName, status, deptID, scopeCondition, scopeParams)
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(userDTO dto.UserUpdateDTO) error {
	user, err := s.userDAO.FindByID(userDTO.ID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if userDTO.Username != "" && userDTO.Username != user.Username {
		existingUser, _ := s.userDAO.FindByUsername(userDTO.Username)
		if existingUser != nil && existingUser.ID != user.ID {
			return errors.New("用户名已存在")
		}
		user.Username = userDTO.Username
	}
	user.RealName = userDTO.RealName
	user.Phone = userDTO.Phone
	user.Email = userDTO.Email
	user.Position = userDTO.Position
	user.DeptID = userDTO.DeptID
	user.Status = userDTO.Status

	if userDTO.HireDate != nil {
		user.HireDate = userDTO.HireDate.Time
	}

	err = s.userDAO.Update(user)
	if err != nil {
		return err
	}

	if userDTO.RoleIDs != nil {
		err = s.userDAO.AssignRoles(userDTO.ID, userDTO.RoleIDs)
		if err != nil {
			return errors.New("分配角色失败")
		}
		// 清除用户权限缓存，使权限变更立即生效
		NewPermissionService().InvalidateUserCache(userDTO.ID)
	}

	return nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int64) error {
	return s.userDAO.Delete(id)
}

// BatchUpdateStatus 批量更新用户状态
func (s *UserService) BatchUpdateStatus(ids []int64, status int8) error {
	if len(ids) == 0 {
		return errors.New("请选择要操作的用户")
	}
	return s.userDAO.BatchUpdateStatus(ids, status)
}

// ResetPassword 管理员重置用户密码
func (s *UserService) ResetPassword(userID int64, pwdDTO dto.ResetPasswordDTO) error {
	_, err := s.userDAO.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	hashedPassword, err := utils.HashPassword(pwdDTO.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	if err := s.userDAO.UpdatePassword(userID, string(hashedPassword)); err != nil {
		return err
	}
	// 清除密码版本号缓存，使旧 JWT 立即失效
	middleware.InvalidatePwdVersionCache(userID)
	return nil
}

// UnlockUser 管理员手动解锁账号（重置登录失败计数并解除锁定）
func (s *UserService) UnlockUser(userID int64) error {
	user, err := s.userDAO.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if user.LockedUntil == nil || !user.LockedUntil.After(time.Now()) {
		return nil // 未锁定，无需解锁
	}

	logger.Info("管理员手动解锁账号",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
	)
	return s.userDAO.ResetLoginAttempts(user.ID)
}

// UpdateProfile 更新当前用户信息
func (s *UserService) UpdateProfile(profileDTO dto.ProfileUpdateDTO) error {
	user, err := s.userDAO.FindByID(profileDTO.ID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.RealName = profileDTO.RealName
	user.Phone = profileDTO.Phone
	user.Email = profileDTO.Email
	user.Position = profileDTO.Position

	if profileDTO.HireDate != nil {
		user.HireDate = profileDTO.HireDate.Time
	}

	return s.userDAO.Update(user)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID int64, pwdDTO dto.ChangePasswordDTO) error {
	user, err := s.userDAO.FindByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if !utils.VerifyPassword(user.Password, pwdDTO.OldPassword) {
		return errors.New("原密码错误")
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(pwdDTO.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	if err := s.userDAO.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
		return err
	}
	// 清除密码版本号缓存，使旧 JWT 立即失效
	middleware.InvalidatePwdVersionCache(user.ID)
	return nil
}

// CountUserMoldOrders 统计指定用户在某时间范围内新增的模具订单数
func (s *UserService) CountUserMoldOrders(userID int64, startTime, endTime time.Time) (int64, error) {
	var count int64
	err := dao.DB.Model(&entity.MoldOrder{}).
		Where("create_user_id = ? AND create_time >= ? AND create_time < ?", userID, startTime, endTime).
		Count(&count).Error
	return count, err
}
