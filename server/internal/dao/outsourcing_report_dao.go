package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type OutsourcingReportDAO struct{}

// NewOutsourcingReportDAO 创建外发报工DAO
func NewOutsourcingReportDAO() *OutsourcingReportDAO {
	return &OutsourcingReportDAO{}
}

// Create 创建外发报工记录
func (d *OutsourcingReportDAO) Create(report *entity.OutsourcingReport) error {
	return GetDB().Create(report).Error
}

// FindByID 根据ID查询外发报工记录
func (d *OutsourcingReportDAO) FindByID(id int64) (*entity.OutsourcingReport, error) {
	var report entity.OutsourcingReport
	err := GetDB().Preload("Outsourcing").First(&report, id).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// Update 更新外发报工记录
func (d *OutsourcingReportDAO) Update(report *entity.OutsourcingReport) error {
	return GetDB().Save(report).Error
}

// Delete 删除外发报工记录
func (d *OutsourcingReportDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.OutsourcingReport{}, id).Error
}

// FindByOutsourcingID 根据外发记录ID查询报工列表
func (d *OutsourcingReportDAO) FindByOutsourcingID(outsourcingID int64) ([]entity.OutsourcingReport, error) {
	var reports []entity.OutsourcingReport
	err := GetDB().Where("outsourcing_id = ?", outsourcingID).
		Order("report_date DESC").
		Find(&reports).Error
	return reports, err
}

// GetOutsourcingTotalQty 获取外发记录的累计产量
func (d *OutsourcingReportDAO) GetOutsourcingTotalQty(outsourcingID int64) (okQty, badQty int, err error) {
	type Result struct {
		OkQty  int
		BadQty int
	}
	var result Result
	err = GetDB().Model(&entity.OutsourcingReport{}).
		Select("COALESCE(SUM(ok_qty), 0) as ok_qty, COALESCE(SUM(bad_qty), 0) as bad_qty").
		Where("outsourcing_id = ?", outsourcingID).
		Scan(&result).Error
	return result.OkQty, result.BadQty, err
}

// GetOutsourcingTotalQtyBatch 批量获取外发记录的累计产量
func (d *OutsourcingReportDAO) GetOutsourcingTotalQtyBatch(ids []int64) (map[int64]int, error) {
	if len(ids) == 0 {
		return map[int64]int{}, nil
	}
	type Result struct {
		OutsourcingID int64
		Total         int
	}
	var results []Result
	err := GetDB().Model(&entity.OutsourcingReport{}).
		Select("outsourcing_id, COALESCE(SUM(ok_qty + bad_qty), 0) as total").
		Where("outsourcing_id IN ?", ids).
		Group("outsourcing_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}
	m := make(map[int64]int, len(results))
	for _, r := range results {
		m[r.OutsourcingID] = r.Total
	}
	return m, nil
}
