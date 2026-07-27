package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProcessTemplateService struct {
	processTemplateDAO *dao.ProcessTemplateDAO
}

func NewProcessTemplateService() *ProcessTemplateService {
	return &ProcessTemplateService{
		processTemplateDAO: dao.NewProcessTemplateDAO(),
	}
}

func (s *ProcessTemplateService) CreateProcessTemplate(template *entity.ProcessTemplate) (*entity.ProcessTemplate, error) {
	err := s.processTemplateDAO.Create(template)
	if err != nil {
		return nil, err
	}
	return template, nil
}

func (s *ProcessTemplateService) GetProcessTemplateByID(id int64) (*entity.ProcessTemplate, error) {
	return s.processTemplateDAO.FindByID(id)
}

func (s *ProcessTemplateService) GetAllProcessTemplates() ([]*entity.ProcessTemplate, error) {
	return s.processTemplateDAO.FindAll()
}

func (s *ProcessTemplateService) GetActiveProcessTemplates() ([]*entity.ProcessTemplate, error) {
	return s.processTemplateDAO.FindActive()
}

func (s *ProcessTemplateService) UpdateProcessTemplate(template *entity.ProcessTemplate) error {
	_, err := s.processTemplateDAO.FindByID(template.ID)
	if err != nil {
		return errors.New("工序模板不存在")
	}
	return s.processTemplateDAO.Update(template)
}

func (s *ProcessTemplateService) DeleteProcessTemplate(id int64) error {
	return s.processTemplateDAO.Delete(id)
}
