package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysDeptDAO struct{}

// NewSysDeptDAO 创建部门 DAO 实例
func NewSysDeptDAO() *SysDeptDAO {
	return &SysDeptDAO{}
}

// Create 创建部门
func (d *SysDeptDAO) Create(dept *entity.SysDept) error {
	return DB.Create(dept).Error
}

// FindByID 根据 ID 查询部门
func (d *SysDeptDAO) FindByID(id int64) (*entity.SysDept, error) {
	var dept entity.SysDept
	err := DB.First(&dept, id).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// FindAll 查询所有部门（parent_id=0 的排在前面，然后按 sort 排序）
func (d *SysDeptDAO) FindAll() ([]*entity.SysDept, error) {
	var depts []*entity.SysDept
	err := DB.Order("parent_id = 0 DESC, sort ASC, id DESC").Find(&depts).Error
	return depts, err
}

// FindByParentID 根据父 ID 查询部门
func (d *SysDeptDAO) FindByParentID(parentID int64) ([]*entity.SysDept, error) {
	var depts []*entity.SysDept
	err := DB.Where("parent_id = ?", parentID).Order("sort ASC").Find(&depts).Error
	return depts, err
}

// Update 更新部门（只更新指定字段，不更新 create_time）
func (d *SysDeptDAO) Update(dept *entity.SysDept) error {
	return DB.Model(&entity.SysDept{}).Where("id = ?", dept.ID).Updates(map[string]interface{}{
		"dept_name": dept.DeptName,
		"parent_id": dept.ParentID,
		"sort":      dept.Sort,
		"leader":    dept.Leader,
		"phone":     dept.Phone,
		"status":    dept.Status,
	}).Error
}

// CountByParentID 统计指定父 ID 下的子部门数量
func (d *SysDeptDAO) CountByParentID(parentID int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.SysDept{}).Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// CountByNameAndParent 统计同级下同名部门数量（创建时 excludeID 传 0，编辑时排除自身）
func (d *SysDeptDAO) CountByNameAndParent(deptName string, parentID int64, excludeID int64) (int64, error) {
	var count int64
	query := DB.Model(&entity.SysDept{}).Where("dept_name = ? AND parent_id = ?", deptName, parentID)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count, err
}

// CountUsersByDeptID 统计指定部门下的用户数量
func (d *SysDeptDAO) CountUsersByDeptID(deptID int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.SysUser{}).Where("dept_id = ?", deptID).Count(&count).Error
	return count, err
}

// UpdateStatus 更新部门状态
func (d *SysDeptDAO) UpdateStatus(id int64, status int8) error {
	return DB.Model(&entity.SysDept{}).Where("id = ?", id).Update("status", status).Error
}

// Delete 删除部门
func (d *SysDeptDAO) Delete(id int64) error {
	return DB.Delete(&entity.SysDept{}, id).Error
}

// GetAllChildDeptIDs 递归获取指定部门及其所有子部门的ID列表（用于数据权限过滤）
func (d *SysDeptDAO) GetAllChildDeptIDs(deptID int64) ([]int64, error) {
	var childIDs []int64

	// 使用递归CTE查询所有子部门ID（MySQL 8.0+）
	// 如果数据库版本不支持CTE，回退到递归查询
	query := `
		WITH RECURSIVE dept_tree AS (
			SELECT id, parent_id FROM sys_dept WHERE parent_id = ?
			UNION ALL
			SELECT d.id, d.parent_id FROM sys_dept d INNER JOIN dept_tree t ON d.parent_id = t.id
		)
		SELECT id FROM dept_tree
	`
	err := DB.Raw(query, deptID).Scan(&childIDs).Error
	if err != nil {
		// 回退方案：简单查询一级子部门（兼容旧版MySQL）
		var directChildren []struct {
			ID int64 `gorm:"column:id"`
		}
		if err := DB.Where("parent_id = ?", deptID).Find(&directChildren).Error; err != nil {
			return nil, err
		}
		for _, child := range directChildren {
			childIDs = append(childIDs, child.ID)
		}
	}

	return childIDs, nil
}
