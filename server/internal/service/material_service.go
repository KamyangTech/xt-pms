package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

type MaterialService struct {
	materialDAO *dao.MaterialDAO
}

func NewMaterialService() *MaterialService {
	return &MaterialService{
		materialDAO: dao.NewMaterialDAO(),
	}
}

// CreateMaterial 创建物料
func (s *MaterialService) CreateMaterial(material *entity.Material) (*entity.Material, error) {
	existing, _ := s.materialDAO.FindByCode(material.MatCode)
	if existing != nil {
		return nil, errors.New("物料编号已存在")
	}

	err := s.materialDAO.Create(material)
	if err != nil {
		return nil, err
	}

	return material, nil
}

// GetMaterialByID 根据 ID 获取物料
func (s *MaterialService) GetMaterialByID(id int64) (*entity.Material, error) {
	return s.materialDAO.FindByID(id)
}

// GetAllMaterials 获取所有物料
func (s *MaterialService) GetAllMaterials() ([]*entity.Material, error) {
	return s.materialDAO.FindAll()
}

// GetMaterialsByPage 分页获取物料
func (s *MaterialService) GetMaterialsByPage(page, pageSize int) ([]*entity.Material, int64, error) {
	result, err := s.materialDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return result.List.([]*entity.Material), result.Total, nil
}

// UpdateMaterial 更新物料
func (s *MaterialService) UpdateMaterial(material *entity.Material) error {
	existing, err := s.materialDAO.FindByID(material.ID)
	if err := utils.EnsureExists(existing, err, "物料"); err != nil {
		return err
	}

	return s.materialDAO.Update(material)
}

// DeleteMaterial 删除物料
func (s *MaterialService) DeleteMaterial(id int64) error {
	return s.materialDAO.Delete(id)
}
