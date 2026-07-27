package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProductionReportDAO struct{}

// NewProductionReportDAO 创建生产报工DAO
func NewProductionReportDAO() *ProductionReportDAO {
	return &ProductionReportDAO{}
}

// Create 创建报工记录
func (d *ProductionReportDAO) Create(report *entity.ProductionReport) error {
	return GetDB().Create(report).Error
}

// FindByID 根据ID查询报工记录
func (d *ProductionReportDAO) FindByID(id int64) (*entity.ProductionReport, error) {
	var report entity.ProductionReport
	err := GetDB().Preload("Order").Preload("Schedule").First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// Update 更新报工记录
func (d *ProductionReportDAO) Update(report *entity.ProductionReport) error {
	return GetDB().Save(report).Error
}

// Delete 删除报工记录
func (d *ProductionReportDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.ProductionReport{}, id).Error
}

// FindByOrderID 根据订单ID查询报工记录列表
func (d *ProductionReportDAO) FindByOrderID(orderID int64) ([]entity.ProductionReport, error) {
	var reports []entity.ProductionReport
	err := GetDB().Where("order_id = ?", orderID).
		Order("report_date DESC").
		Find(&reports).Error
	return reports, err
}

// GetOrderTotalQty 获取订单的累计产量
func (d *ProductionReportDAO) GetOrderTotalQty(orderID int64) (okQty, badQty, scrapQty int, err error) {
	type Result struct {
		OkQty    int
		BadQty   int
		ScrapQty int
	}
	var result Result
	err = GetDB().Model(&entity.ProductionReport{}).
		Select("COALESCE(SUM(ok_qty), 0) as ok_qty, COALESCE(SUM(bad_qty), 0) as bad_qty, COALESCE(SUM(scrap_qty), 0) as scrap_qty").
		Where("order_id = ?", orderID).
		Scan(&result).Error
	return result.OkQty, result.BadQty, result.ScrapQty, err
}
