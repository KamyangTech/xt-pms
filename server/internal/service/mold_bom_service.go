package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// MoldBOMService 模具 BOM 服务
type MoldBOMService struct {
	moldBOMDAO *dao.MoldBOMDAO
}

func NewMoldBOMService() *MoldBOMService {
	return &MoldBOMService{moldBOMDAO: dao.NewMoldBOMDAO()}
}

func (s *MoldBOMService) CreateMoldBOM(bom *entity.MoldBOM) (*entity.MoldBOM, error) {
	err := s.moldBOMDAO.Create(bom)
	if err != nil {
		return nil, err
	}
	return bom, nil
}

func (s *MoldBOMService) GetMoldBOMByID(id int64) (*entity.MoldBOM, error) {
	return s.moldBOMDAO.FindByID(id)
}

func (s *MoldBOMService) GetMoldBOMByMoldID(moldID int64) ([]*entity.MoldBOM, error) {
	return s.moldBOMDAO.FindByMoldID(moldID)
}

func (s *MoldBOMService) UpdateMoldBOM(bom *entity.MoldBOM) error {
	_, err := s.moldBOMDAO.FindByID(bom.ID)
	if err != nil {
		return errors.New("BOM 不存在")
	}
	return s.moldBOMDAO.Update(bom)
}

func (s *MoldBOMService) DeleteMoldBOM(id int64) error {
	return s.moldBOMDAO.Delete(id)
}

func (s *MoldBOMService) DeleteMoldBOMByMoldID(moldID int64) error {
	return s.moldBOMDAO.DeleteByMoldID(moldID)
}
