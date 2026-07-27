package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// 最大部门层级（1=顶级，2=子级，3=孙级）
const MaxDeptLevel = 3

type SysDeptService struct {
	deptDAO *dao.SysDeptDAO
}

func NewSysDeptService() *SysDeptService {
	return &SysDeptService{
		deptDAO: dao.NewSysDeptDAO(),
	}
}

// getDeptLevel 计算部门的层级（0=根级别，1=一级子部门，2=二级孙部门）
func getDeptLevel(allDepts []*entity.SysDept, id int64) int {
	if id == 0 {
		return 0
	}
	deptMap := make(map[int64]*entity.SysDept, len(allDepts))
	for _, dept := range allDepts {
		deptMap[dept.ID] = dept
	}

	level := 0
	currentID := id
	for currentID > 0 {
		level++
		if parent, ok := deptMap[currentID]; ok {
			currentID = parent.ParentID
		} else {
			break
		}
		if level > 100 { // 防止无限循环
			return level
		}
	}
	return level
}

// checkDeptLevel 检查部门层级是否超过最大限制
func (s *SysDeptService) checkDeptLevel(parentID int64) error {
	if parentID == 0 {
		return nil // 顶级部门不限
	}
	allDepts, err := s.deptDAO.FindAll()
	if err != nil {
		return err
	}
	parentLevel := getDeptLevel(allDepts, parentID)
	if parentLevel >= MaxDeptLevel {
		return errors.New("部门层级已达最大限制（最多3级）")
	}
	return nil
}

// CreateDept 创建部门
func (s *SysDeptService) CreateDept(dept *entity.SysDept) (*entity.SysDept, error) {
	count, err := s.deptDAO.CountByNameAndParent(dept.DeptName, dept.ParentID, 0)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("同级部门下已存在相同名称的部门")
	}

	if err := s.checkDeptLevel(dept.ParentID); err != nil {
		return nil, err
	}

	err = s.deptDAO.Create(dept)
	if err != nil {
		return nil, err
	}

	return dept, nil
}

// GetDeptByID 根据 ID 获取部门
func (s *SysDeptService) GetDeptByID(id int64) (*entity.SysDept, error) {
	return s.deptDAO.FindByID(id)
}

// GetAllDepts 获取所有部门
func (s *SysDeptService) GetAllDepts() ([]*entity.SysDept, error) {
	return s.deptDAO.FindAll()
}

// GetDeptsByParentID 根据父 ID 获取部门
func (s *SysDeptService) GetDeptsByParentID(parentID int64) ([]*entity.SysDept, error) {
	return s.deptDAO.FindByParentID(parentID)
}

// UpdateDept 更新部门
func (s *SysDeptService) UpdateDept(dept *entity.SysDept) error {
	_, err := s.deptDAO.FindByID(dept.ID)
	if err != nil {
		return errors.New("部门不存在")
	}

	count, err := s.deptDAO.CountByNameAndParent(dept.DeptName, dept.ParentID, dept.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("同级部门下已存在相同名称的部门")
	}

	if err := s.checkDeptLevel(dept.ParentID); err != nil {
		return err
	}

	return s.deptDAO.Update(dept)
}

func BuildDeptTree(depts []*entity.SysDept) []*entity.SysDept {
	if len(depts) == 0 {
		return depts
	}

	deptMap := make(map[int64]*entity.SysDept, len(depts))
	for _, dept := range depts {
		dept.Children = nil
		deptMap[dept.ID] = dept
	}

	var roots []*entity.SysDept
	for _, dept := range depts {
		if dept.ParentID == 0 {
			roots = append(roots, dept)
		} else if parent, ok := deptMap[dept.ParentID]; ok {
			parent.Children = append(parent.Children, dept)
		} else {
			// 父部门不存在，作为根节点处理
			roots = append(roots, dept)
		}
	}

	return roots
}

// DeleteDept 删除部门（先检查是否有子部门或关联用户）
func (s *SysDeptService) DeleteDept(id int64) error {
	subCount, err := s.deptDAO.CountByParentID(id)
	if err != nil {
		return err
	}
	if subCount > 0 {
		return errors.New("该部门下存在子部门，无法删除")
	}

	userCount, err := s.deptDAO.CountUsersByDeptID(id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return errors.New("该部门下存在关联用户，无法删除")
	}

	return s.deptDAO.Delete(id)
}

// UpdateDeptStatus 更新部门状态
func (s *SysDeptService) UpdateDeptStatus(id int64, status int8) error {
	return s.deptDAO.UpdateStatus(id, status)
}
