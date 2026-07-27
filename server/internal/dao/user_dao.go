package dao

import (
	"strings"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type SysUserDAO struct{}

// NewSysUserDAO 创建用户 DAO 实例
func NewSysUserDAO() *SysUserDAO {
	return &SysUserDAO{}
}

// Create 创建用户
func (d *SysUserDAO) Create(user *entity.SysUser) error {
	return DB.Create(user).Error
}

// FindByID 根据 ID 查询用户
func (d *SysUserDAO) FindByID(id int64) (*entity.SysUser, error) {
	var user entity.SysUser
	err := DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername 根据用户名查询用户
func (d *SysUserDAO) FindByUsername(username string) (*entity.SysUser, error) {
	var user entity.SysUser
	err := DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDWithRoles 根据 ID 查询用户并加载角色
func (d *SysUserDAO) FindByIDWithRoles(id int64) (*entity.SysUser, error) {
	var user entity.SysUser
	err := DB.Preload("Roles").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserIDsByRoleID 获取指定角色下的所有用户 ID
func (d *SysUserDAO) GetUserIDsByRoleID(roleID int64) ([]int64, error) {
	var ids []int64
	err := DB.Table("sys_user_role").Where("role_id = ?", roleID).Pluck("user_id", &ids).Error
	return ids, err
}

// FindAll 查询所有用户（下拉选择框用，限制 1000 条）
func (d *SysUserDAO) FindAll() ([]*entity.SysUser, error) {
	var users []*entity.SysUser
	err := DB.Select("id", "username", "real_name", "status").Order("id DESC").Limit(1000).Find(&users).Error
	return users, err
}

// FindPage 分页查询用户（支持按用户名/姓名/状态/部门搜索）
func (d *SysUserDAO) FindPage(page, pageSize int, username, realName string, status *int8, deptID *int64) (*utils.PageResult, error) {
	var users []*entity.SysUser
	return utils.Paginate(DB, page, pageSize, &users, func(db *gorm.DB) *gorm.DB {
		query := db.Preload("Roles").Order("id DESC")
		if username != "" {
			query = query.Where("username LIKE ?", "%"+username+"%")
		}
		if realName != "" {
			query = query.Where("real_name LIKE ?", "%"+realName+"%")
		}
		if status != nil {
			query = query.Where("status = ?", *status)
		}
		if deptID != nil {
			query = query.Where("dept_id = ?", *deptID)
		}
		return query
	})
}

// FindPageWithDataScope 分页查询用户（带数据权限范围过滤）
// scopeCondition: 数据范围SQL条件（如 "dept_id IN ?" 或 "id = ?"）
// scopeParams: 条件参数
func (d *SysUserDAO) FindPageWithDataScope(page, pageSize int, username, realName string, status *int8, deptID *int64, scopeCondition string, scopeParams []interface{}) (*utils.PageResult, error) {
	var users []*entity.SysUser
	return utils.Paginate(DB, page, pageSize, &users, func(db *gorm.DB) *gorm.DB {
		query := db.Preload("Roles").Order("id DESC")

		// 应用数据权限过滤条件
		if scopeCondition != "" && len(scopeParams) > 0 {
			query = query.Where(scopeCondition, scopeParams...)
		}

		// 应用其他搜索条件
		if username != "" {
			query = query.Where("username LIKE ?", "%"+username+"%")
		}
		if realName != "" {
			query = query.Where("real_name LIKE ?", "%"+realName+"%")
		}
		if status != nil {
			query = query.Where("status = ?", *status)
		}
		if deptID != nil {
			query = query.Where("dept_id = ?", *deptID)
		}
		return query
	})
}

// BatchUpdateStatus 批量更新用户状态
func (d *SysUserDAO) BatchUpdateStatus(ids []int64, status int8) error {
	return DB.Model(&entity.SysUser{}).Where("id IN ?", ids).Update("status", status).Error
}

// UpdatePassword 更新用户密码（同时自增密码版本号，使旧 Token 失效）
func (d *SysUserDAO) UpdatePassword(id int64, hashedPassword string) error {
	return DB.Model(&entity.SysUser{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"password":         hashedPassword,
			"password_version": gorm.Expr("password_version + 1"),
		}).Error
}

// Update 更新用户
func (d *SysUserDAO) Update(user *entity.SysUser) error {
	return DB.Save(user).Error
}

// Delete 删除用户
func (d *SysUserDAO) Delete(id int64) error {
	return DB.Delete(&entity.SysUser{}, id).Error
}

// IncrementLoginAttempts 增加登录失败次数
func (d *SysUserDAO) IncrementLoginAttempts(userID int64) error {
	return DB.Model(&entity.SysUser{}).Where("id = ?", userID).
		UpdateColumn("login_attempts", gorm.Expr("login_attempts + 1")).Error
}

// LockAccount 锁定账号（禁止登录到指定时间）
func (d *SysUserDAO) LockAccount(userID int64, lockedUntil time.Time) error {
	return DB.Model(&entity.SysUser{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"login_attempts": 0,
			"locked_until":   lockedUntil,
		}).Error
}

// ResetLoginAttempts 重置登录失败次数并解除锁定
func (d *SysUserDAO) ResetLoginAttempts(userID int64) error {
	return DB.Model(&entity.SysUser{}).Where("id = ?", userID).
		Updates(map[string]interface{}{
			"login_attempts": 0,
			"locked_until":   nil,
		}).Error
}

// AssignRoles 为用户分配角色
func (d *SysUserDAO) AssignRoles(userID int64, roleIDs []int64) error {
	// 开启事务
	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 先删除旧的角色关联
	if err := tx.Exec("DELETE FROM sys_user_role WHERE user_id = ?", userID).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 批量插入新角色（使用 VALUES 多行语法，一次 INSERT 代替逐条）
	if len(roleIDs) > 0 {
		values := make([]string, 0, len(roleIDs))
		args := make([]interface{}, 0, len(roleIDs)*2)
		for _, roleID := range roleIDs {
			values = append(values, "(?, ?)")
			args = append(args, userID, roleID)
		}
		sql := "INSERT INTO sys_user_role (user_id, role_id) VALUES " + strings.Join(values, ", ")
		if err := tx.Exec(sql, args...).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}
