package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProductionScheduleDAO struct{}

// NewProductionScheduleDAO 创建生产排程DAO
func NewProductionScheduleDAO() *ProductionScheduleDAO {
	return &ProductionScheduleDAO{}
}

// Create 创建排程
func (d *ProductionScheduleDAO) Create(schedule *entity.ProductionSchedule) error {
	return GetDB().Create(schedule).Error
}

// FindByID 根据ID查询排程
func (d *ProductionScheduleDAO) FindByID(id int64) (*entity.ProductionSchedule, error) {
	var schedule entity.ProductionSchedule
	err := GetDB().Preload("Order").Preload("Device").First(&schedule, id).Error
	if err != nil {
		return nil, err
	}
	return &schedule, nil
}

// Update 更新排程
func (d *ProductionScheduleDAO) Update(schedule *entity.ProductionSchedule) error {
	return GetDB().Save(schedule).Error
}

// Delete 删除排程
func (d *ProductionScheduleDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.ProductionSchedule{}, id).Error
}

// FindByOrderID 根据订单ID查询排程列表
func (d *ProductionScheduleDAO) FindByOrderID(orderID int64) ([]entity.ProductionSchedule, error) {
	var schedules []entity.ProductionSchedule
	err := GetDB().Where("order_id = ?", orderID).Preload("Device").Order("schedule_date ASC").Find(&schedules).Error
	return schedules, err
}

// FindByDateRange 根据日期范围查询排程
func (d *ProductionScheduleDAO) FindByDateRange(startDate, endDate string) ([]entity.ProductionSchedule, error) {
	var schedules []entity.ProductionSchedule
	err := GetDB().Where("schedule_date BETWEEN ? AND ?", startDate, endDate).
		Preload("Order").Preload("Device").
		Order("schedule_date ASC, time_slot ASC").
		Find(&schedules).Error
	return schedules, err
}

// CheckDeviceConflict 检查设备时间冲突
func (d *ProductionScheduleDAO) CheckDeviceConflict(deviceID int64, scheduleDate string, excludeID int64) (bool, error) {
	var count int64
	query := GetDB().Model(&entity.ProductionSchedule{}).
		Where("device_id = ? AND schedule_date = ? AND status IN (1, 2)", deviceID, scheduleDate)
	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
