package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

type ProcessorService struct {
	processorDAO *dao.ProcessorDAO
}

func NewProcessorService() *ProcessorService {
	return &ProcessorService{
		processorDAO: dao.NewProcessorDAO(),
	}
}

// CreateProcessor 创建加工商
func (s *ProcessorService) CreateProcessor(processor *entity.Processor) (*entity.Processor, error) {
	existing, _ := s.processorDAO.FindByCode(processor.ProcessorCode)
	if existing != nil {
		return nil, errors.New("加工商编号已存在")
	}

	err := s.processorDAO.Create(processor)
	if err != nil {
		return nil, err
	}

	return processor, nil
}

// GetProcessorByID 根据 ID 获取加工商
func (s *ProcessorService) GetProcessorByID(id int64) (*entity.Processor, error) {
	return s.processorDAO.FindByID(id)
}

// GetAllProcessors 获取所有加工商
func (s *ProcessorService) GetAllProcessors() ([]*entity.Processor, error) {
	return s.processorDAO.FindAll()
}

// GetProcessorsByPageWithFilter 分页获取加工商（带搜索条件）
func (s *ProcessorService) GetProcessorsByPageWithFilter(page, pageSize int, processorCode, processorName, contact, phone string, status int) ([]*entity.Processor, int64, error) {
	offset := (page - 1) * pageSize
	return s.processorDAO.FindPageWithFilter(offset, pageSize, processorCode, processorName, contact, phone, status)
}

// UpdateProcessor 更新加工商
func (s *ProcessorService) UpdateProcessor(processor *entity.Processor) error {
	existing, err := s.processorDAO.FindByID(processor.ID)
	if err := utils.EnsureExists(existing, err, "加工商"); err != nil {
		return err
	}

	// 如果编号被修改，检查新编号是否已被其他加工商占用
	if processor.ProcessorCode != existing.ProcessorCode {
		other, _ := s.processorDAO.FindByCode(processor.ProcessorCode)
		if other != nil && other.ID != processor.ID {
			return errors.New("加工商编号已被其他加工商使用")
		}
	}

	return s.processorDAO.Update(processor)
}

// DeleteProcessor 删除加工商
func (s *ProcessorService) DeleteProcessor(id int64) error {
	return s.processorDAO.Delete(id)
}

// BatchDeleteProcessors 批量删除加工商
func (s *ProcessorService) BatchDeleteProcessors(ids []int64) error {
	return s.processorDAO.BatchDelete(ids)
}
