package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldTryTestDAO struct{}

func NewMoldTryTestDAO() *MoldTryTestDAO {
	return &MoldTryTestDAO{}
}

func (d *MoldTryTestDAO) Create(test *entity.MoldTryTest) error {
	return DB.Create(test).Error
}

func (d *MoldTryTestDAO) FindByID(id int64) (*entity.MoldTryTest, error) {
	var test entity.MoldTryTest
	err := DB.First(&test, id).Error
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (d *MoldTryTestDAO) FindByWorkOrderID(workOrderID int64) ([]*entity.MoldTryTest, error) {
	var tests []*entity.MoldTryTest
	err := DB.Where("mold_work_id = ?", workOrderID).Order("id DESC").Find(&tests).Error
	return tests, err
}

func (d *MoldTryTestDAO) Update(test *entity.MoldTryTest) error {
	return DB.Save(test).Error
}

func (d *MoldTryTestDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldTryTest{}, id).Error
}
