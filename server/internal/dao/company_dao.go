package dao

import "github.com/KamyangTech/xt-pms/server/internal/model/entity"

type SysCompanyDAO struct{}

// NewSysCompanyDAO 创建公司信息DAO实例
func NewSysCompanyDAO() *SysCompanyDAO {
	return &SysCompanyDAO{}
}

// GetCompany 获取公司信息（通常只有一条记录）
func (d *SysCompanyDAO) GetCompany() (*entity.SysCompany, error) {
	var company entity.SysCompany
	err := DB.First(&company).Error
	if err != nil {
		return nil, err
	}
	return &company, nil
}

// UpdateCompany 更新公司信息
func (d *SysCompanyDAO) UpdateCompany(company *entity.SysCompany) error {
	return DB.Save(company).Error
}

// UpdateCompanyFields 更新公司信息的指定字段（不更新 create_time）
func (d *SysCompanyDAO) UpdateCompanyFields(id int64, companyName, shortName, englishName, englishShortName, logo string) error {
	return DB.Model(&entity.SysCompany{}).Where("id = ?", id).Updates(map[string]interface{}{
		"company_name":       companyName,
		"short_name":         shortName,
		"english_name":       englishName,
		"english_short_name": englishShortName,
		"logo":               logo,
	}).Error
}

// CreateCompany 创建公司信息
func (d *SysCompanyDAO) CreateCompany(company *entity.SysCompany) error {
	return DB.Create(company).Error
}
