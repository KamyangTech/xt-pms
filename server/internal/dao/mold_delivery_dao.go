package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldDeliveryDAO struct{}

func NewMoldDeliveryDAO() *MoldDeliveryDAO {
	return &MoldDeliveryDAO{}
}

func (d *MoldDeliveryDAO) Create(delivery *entity.MoldDelivery) error {
	return DB.Create(delivery).Error
}

func (d *MoldDeliveryDAO) FindByID(id int64) (*entity.MoldDelivery, error) {
	var delivery entity.MoldDelivery
	err := DB.First(&delivery, id).Error
	if err != nil {
		return nil, err
	}
	return &delivery, nil
}

func (d *MoldDeliveryDAO) FindByWorkOrderID(workOrderID int64) ([]*entity.MoldDelivery, error) {
	var deliveries []*entity.MoldDelivery
	err := DB.Where("mold_work_id = ?", workOrderID).Order("id DESC").Find(&deliveries).Error
	return deliveries, err
}

func (d *MoldDeliveryDAO) Update(delivery *entity.MoldDelivery) error {
	return DB.Save(delivery).Error
}

func (d *MoldDeliveryDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldDelivery{}, id).Error
}
