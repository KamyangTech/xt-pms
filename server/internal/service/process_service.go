package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

type ProcessService struct {
	processDAO *dao.ProcessDAO
}

func NewProcessService() *ProcessService {
	return &ProcessService{
		processDAO: dao.NewProcessDAO(),
	}
}

// CreateProcess 创建工序
func (s *ProcessService) CreateProcess(process *entity.Process) (*entity.Process, error) {
	existing, _ := s.processDAO.FindByCode(process.ProcessCode)
	if existing != nil {
		return nil, errors.New("工序编码已存在")
	}

	err := s.processDAO.Create(process)
	if err != nil {
		return nil, err
	}

	return process, nil
}

// GetProcessByID 根据 ID 获取工序
func (s *ProcessService) GetProcessByID(id int64) (*entity.Process, error) {
	return s.processDAO.FindByID(id)
}

// GetAllProcesses 获取所有工序
func (s *ProcessService) GetAllProcesses() ([]*entity.Process, error) {
	return s.processDAO.FindAll()
}

// GetProcessesByPage 分页获取工序
func (s *ProcessService) GetProcessesByPage(page, pageSize int) ([]*entity.Process, int64, error) {
	result, err := s.processDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return result.List.([]*entity.Process), result.Total, nil
}

// UpdateProcess 更新工序
func (s *ProcessService) UpdateProcess(process *entity.Process) error {
	existing, err := s.processDAO.FindByID(process.ID)
	if err := utils.EnsureExists(existing, err, "工序"); err != nil {
		return err
	}

	return s.processDAO.Update(process)
}

// DeleteProcess 删除工序
func (s *ProcessService) DeleteProcess(id int64) error {
	return s.processDAO.Delete(id)
}
