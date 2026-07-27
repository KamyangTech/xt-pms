package service

import (
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysCompanyService struct {
	companyDAO *dao.SysCompanyDAO
}

// NewSysCompanyService 创建公司信息Service实例
func NewSysCompanyService() *SysCompanyService {
	return &SysCompanyService{
		companyDAO: dao.NewSysCompanyDAO(),
	}
}

// GetCompany 获取公司信息
func (s *SysCompanyService) GetCompany() (*entity.SysCompany, error) {
	return s.companyDAO.GetCompany()
}

// UpdateCompany 更新公司信息
func (s *SysCompanyService) UpdateCompany(company *entity.SysCompany) error {
	return s.companyDAO.UpdateCompany(company)
}

// CreateCompany 创建公司信息
func (s *SysCompanyService) CreateCompany(company *entity.SysCompany) error {
	return s.companyDAO.CreateCompany(company)
}

// SaveCompany 保存公司信息（如果不存在则创建，存在则更新）
func (s *SysCompanyService) SaveCompany(company *entity.SysCompany) error {
	existing, err := s.companyDAO.GetCompany()
	if err != nil || existing == nil {
		// 不存在，创建
		return s.companyDAO.CreateCompany(company)
	}
	// 存在，更新 - 只更新业务字段，保留原有的 create_time
	return s.companyDAO.UpdateCompanyFields(existing.ID, company.CompanyName, company.ShortName, company.EnglishName, company.EnglishShortName, company.Logo)
}
