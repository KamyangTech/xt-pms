package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

// ProcessTemplateDAO 工序模板DAO
type ProcessTemplateDAO struct {
	db *gorm.DB
}

func NewProcessTemplateDAO() *ProcessTemplateDAO {
	return &ProcessTemplateDAO{db: GetDB()}
}

func (d *ProcessTemplateDAO) Create(template *entity.ProcessTemplate) error {
	return d.db.Create(template).Error
}

func (d *ProcessTemplateDAO) FindByID(id int64) (*entity.ProcessTemplate, error) {
	var template entity.ProcessTemplate
	err := d.db.First(&template, id).Error
	return &template, err
}

func (d *ProcessTemplateDAO) FindAll() ([]*entity.ProcessTemplate, error) {
	var templates []*entity.ProcessTemplate
	err := d.db.Order("sort ASC").Find(&templates).Error
	return templates, err
}

func (d *ProcessTemplateDAO) FindActive() ([]*entity.ProcessTemplate, error) {
	var templates []*entity.ProcessTemplate
	err := d.db.Where("status = ?", 1).Order("sort ASC").Find(&templates).Error
	return templates, err
}

func (d *ProcessTemplateDAO) Update(template *entity.ProcessTemplate) error {
	return d.db.Model(&entity.ProcessTemplate{}).Where("id = ?", template.ID).Updates(map[string]interface{}{
		"group_id":     template.GroupID,
		"group_name":   template.GroupName,
		"process_name": template.ProcessName,
		"sort":         template.Sort,
		"status":       template.Status,
		"remark":       template.Remark,
	}).Error
}

func (d *ProcessTemplateDAO) Delete(id int64) error {
	return d.db.Delete(&entity.ProcessTemplate{}, id).Error
}
