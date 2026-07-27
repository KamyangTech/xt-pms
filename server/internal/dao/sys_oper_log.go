package dao

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysOperLogDAO struct{}

// NewSysOperLogDAO 创建操作日志DAO实例
func NewSysOperLogDAO() *SysOperLogDAO {
	return &SysOperLogDAO{}
}

// Create 记录操作日志
func (d *SysOperLogDAO) Create(log *entity.SysOperLog) error {
	return DB.Create(log).Error
}

// BatchCreate 批量写入操作日志
func (d *SysOperLogDAO) BatchCreate(logs []*entity.SysOperLog) error {
	if len(logs) == 0 {
		return nil
	}
	return DB.CreateInBatches(logs, 100).Error
}

// FindPageWithFilter 分页查询操作日志（带筛选条件）
func (d *SysOperLogDAO) FindPageWithFilter(offset, limit int, title, operName string, operType int8, status int8, businessType string, startTime, endTime string) ([]*entity.SysOperLog, int64, error) {
	var list []*entity.SysOperLog
	var total int64

	query := DB.Model(&entity.SysOperLog{})

	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if operName != "" {
		query = query.Where("oper_name LIKE ?", "%"+operName+"%")
	}
	if operType > 0 {
		query = query.Where("oper_type = ?", operType)
	}
	if status == 0 || status == 1 {
		query = query.Where("status = ?", status)
	}
	if businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}
	if startTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local); err == nil {
			query = query.Where("create_time >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local); err == nil {
			query = query.Where("create_time <= ?", t)
		}
	}

	query.Count(&total)
	err := query.Offset(offset).Limit(limit).Order("id DESC").Find(&list).Error
	return list, total, err
}

// FindByID 根据ID获取操作日志详情
func (d *SysOperLogDAO) FindByID(id int64) (*entity.SysOperLog, error) {
	var log entity.SysOperLog
	err := DB.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// FindAllForExport 查询所有匹配的操作日志（不分页，供导出用）
func (d *SysOperLogDAO) FindAllForExport(title, operName string, operType int8, status int8, businessType, startTime, endTime string, limit int) ([]*entity.SysOperLog, error) {
	var list []*entity.SysOperLog

	query := DB.Model(&entity.SysOperLog{}).Select("id, title, oper_type, business_type, method, request_url, request_method, oper_name, oper_ip, status, cost_time, create_time, error_msg")

	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if operName != "" {
		query = query.Where("oper_name LIKE ?", "%"+operName+"%")
	}
	if operType > 0 {
		query = query.Where("oper_type = ?", operType)
	}
	if status == 0 || status == 1 {
		query = query.Where("status = ?", status)
	}
	if businessType != "" {
		query = query.Where("business_type = ?", businessType)
	}
	if startTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local); err == nil {
			query = query.Where("create_time >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local); err == nil {
			query = query.Where("create_time <= ?", t)
		}
	}

	err := query.Order("id DESC").Limit(limit).Find(&list).Error
	return list, err
}

// DeleteOlderThan 删除指定时间之前的操作日志，返回删除条数（用于定时清理）
func (d *SysOperLogDAO) DeleteOlderThan(before time.Time) (int64, error) {
	// 分批删除，避免单条大事务锁表
	var totalAffected int64
	batchSize := 1000
	for {
		result := DB.Where("create_time < ?", before).Limit(batchSize).Delete(&entity.SysOperLog{})
		if result.Error != nil {
			return totalAffected, result.Error
		}
		if result.RowsAffected == 0 {
			break
		}
		totalAffected += result.RowsAffected
	}
	return totalAffected, nil
}
