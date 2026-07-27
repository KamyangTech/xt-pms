package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldStockDAO struct{}

func NewMoldStockDAO() *MoldStockDAO {
	return &MoldStockDAO{}
}

func (d *MoldStockDAO) Create(stock *entity.MoldStock) error {
	return DB.Create(stock).Error
}

func (d *MoldStockDAO) FindByID(id int64) (*entity.MoldStock, error) {
	var stock entity.MoldStock
	err := DB.First(&stock, id).Error
	if err != nil {
		return nil, err
	}
	return &stock, nil
}

func (d *MoldStockDAO) FindByMoldID(moldID int64) ([]*entity.MoldStock, error) {
	var stocks []*entity.MoldStock
	err := DB.Where("mold_id = ?", moldID).Order("id DESC").Find(&stocks).Error
	return stocks, err
}

func (d *MoldStockDAO) Update(stock *entity.MoldStock) error {
	return DB.Save(stock).Error
}

func (d *MoldStockDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldStock{}, id).Error
}
