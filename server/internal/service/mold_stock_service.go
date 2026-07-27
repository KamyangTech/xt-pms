package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldStockService struct {
	moldStockDAO *dao.MoldStockDAO
}

func NewMoldStockService() *MoldStockService {
	return &MoldStockService{moldStockDAO: dao.NewMoldStockDAO()}
}

func (s *MoldStockService) CreateMoldStock(stock *entity.MoldStock) (*entity.MoldStock, error) {
	err := s.moldStockDAO.Create(stock)
	if err != nil {
		return nil, err
	}
	return stock, nil
}

func (s *MoldStockService) GetMoldStockByID(id int64) (*entity.MoldStock, error) {
	return s.moldStockDAO.FindByID(id)
}

func (s *MoldStockService) GetMoldStocksByMoldID(moldID int64) ([]*entity.MoldStock, error) {
	return s.moldStockDAO.FindByMoldID(moldID)
}

func (s *MoldStockService) UpdateMoldStock(stock *entity.MoldStock) error {
	_, err := s.moldStockDAO.FindByID(stock.ID)
	if err != nil {
		return errors.New("库存记录不存在")
	}
	return s.moldStockDAO.Update(stock)
}

func (s *MoldStockService) DeleteMoldStock(id int64) error {
	return s.moldStockDAO.Delete(id)
}
