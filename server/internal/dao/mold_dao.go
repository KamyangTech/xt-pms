package dao

import (
	"strings"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type MoldDAO struct{}

// NewMoldDAO 创建模具 DAO 实例
func NewMoldDAO() *MoldDAO {
	return &MoldDAO{}
}

// Create 创建模具档案
func (d *MoldDAO) Create(mold *entity.Mold) error {
	return DB.Create(mold).Error
}

// FindByID 根据 ID 查询模具
func (d *MoldDAO) FindByID(id int64) (*entity.Mold, error) {
	var mold entity.Mold
	err := DB.Preload("Customer").First(&mold, id).Error
	if err != nil {
		return nil, err
	}

	return &mold, nil
}

// FindByCode 根据编号查询模具
func (d *MoldDAO) FindByCode(code string) (*entity.Mold, error) {
	var mold entity.Mold
	err := DB.Where("mold_code = ?", code).First(&mold).Error
	if err != nil {
		return nil, err
	}
	return &mold, nil
}

// FindCodesByIDs 根据ID列表批量查询模具编号映射（id -> mold_code）
func (d *MoldDAO) FindCodesByIDs(ids []int64) (map[int64]string, error) {
	if len(ids) == 0 {
		return make(map[int64]string), nil
	}
	type moldCodeRow struct {
		ID       int64
		MoldCode string
	}
	var rows []moldCodeRow
	err := DB.Model(&entity.Mold{}).
		Select("id, mold_code").
		Where("id IN ?", ids).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]string, len(rows))
	for _, r := range rows {
		result[r.ID] = r.MoldCode
	}
	return result, nil
}

// FindAll 查询所有模具（下拉选择框用，限制 1000 条）
func (d *MoldDAO) FindAll() ([]*entity.Mold, error) {
	var molds []*entity.Mold
	err := DB.Select("id", "mold_code", "mold_name", "customer_id").Order("id DESC").Limit(1000).Find(&molds).Error
	return molds, err
}

// FindPage 分页查询模具
func (d *MoldDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var molds []*entity.Mold
	return utils.Paginate(DB, page, pageSize, &molds, func(db *gorm.DB) *gorm.DB {
		return db.Preload("Customer").Order("id DESC")
	})
}

// Update 更新模具
func (d *MoldDAO) Update(mold *entity.Mold) error {
	// 使用 Select 明确指定要更新的字段，避免更新关联对象
	return DB.Model(&entity.Mold{}).
		Where("id = ?", mold.ID).
		Select("mold_code", "mold_name", "customer_id", "ownership_type",
			"production_date", "process_status", "current_location",
			"life_cycle_status", "mold_type", "mold_process", "material_type", "remark", "update_time").
		Updates(mold).Error
}

// FindByIDWithTx 使用指定db连接查询模具（用于事务内）
func (d *MoldDAO) FindByIDWithTx(db *gorm.DB, id int64) (*entity.Mold, error) {
	var mold entity.Mold
	err := db.First(&mold, id).Error
	if err != nil {
		return nil, err
	}
	return &mold, nil
}

// UpdateWithMap 使用map更新模具字段
func (d *MoldDAO) UpdateWithMap(id int64, updates map[string]interface{}) error {
	return DB.Model(&entity.Mold{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateWithMapWithTx 使用指定db连接更新模具字段（用于事务内）
func (d *MoldDAO) UpdateWithMapWithTx(db *gorm.DB, id int64, updates map[string]interface{}) error {
	return db.Model(&entity.Mold{}).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除模具
func (d *MoldDAO) Delete(id int64) error {
	return DB.Delete(&entity.Mold{}, id).Error
}

// UpdateStatus 更新模具状态
func (d *MoldDAO) UpdateStatus(id int64, status int8) error {
	return DB.Model(&entity.Mold{}).Where("id = ?", id).Update("status", status).Error
}

// FindPageWithFilter 带筛选条件的分页查询模具
func (d *MoldDAO) FindPageWithFilter(offset, limit int, moldCode, moldName string, customerID *int64, lifeCycleStatus, moldType, moldProcess, materialType, startDate, endDate string) ([]*entity.Mold, int64, error) {
	var molds []*entity.Mold
	var total int64

	query := DB.Model(&entity.Mold{})

	if moldCode != "" {
		query = query.Where("mold_code LIKE ?", moldCode+"%")
	}
	if moldName != "" {
		query = query.Where("mold_name LIKE ?", "%"+moldName+"%")
	}
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}
	if lifeCycleStatus != "" {
		// 支持逗号分隔的多状态筛选（如 "PRODUCING,IN_STOCK,DELIVERED"）
		statusList := strings.Split(lifeCycleStatus, ",")
		if len(statusList) == 1 {
			query = query.Where("life_cycle_status = ?", strings.TrimSpace(statusList[0]))
		} else {
			trimmed := make([]string, 0, len(statusList))
			for _, s := range statusList {
				s = strings.TrimSpace(s)
				if s != "" {
					trimmed = append(trimmed, s)
				}
			}
			query = query.Where("life_cycle_status IN ?", trimmed)
		}
	}
	if moldType != "" {
		query = query.Where("mold_type = ?", moldType)
	}
	if moldProcess != "" {
		query = query.Where("mold_process = ?", moldProcess)
	}
	if materialType != "" {
		query = query.Where("material_type = ?", materialType)
	}
	if startDate != "" && endDate != "" {
		query = query.Where("production_date >= ? AND production_date < ?", startDate, endDate+" 23:59:59")
	} else if startDate != "" {
		query = query.Where("production_date >= ?", startDate)
	} else if endDate != "" {
		query = query.Where("production_date <= ?", endDate+" 23:59:59")
	}

	query.Count(&total)
	err := query.Preload("Customer").Offset(offset).Limit(limit).Order("id DESC").Find(&molds).Error

	return molds, total, err
}

// GetPrimaryImagesByMoldIDs 根据模具ID列表批量查询主图（如果没有主图则返回第一张图片）
func (d *MoldDAO) GetPrimaryImagesByMoldIDs(moldIDs []int64) (map[int64]string, error) {
	if len(moldIDs) == 0 {
		return make(map[int64]string), nil
	}

	var images []struct {
		BusinessID int64  `gorm:"column:business_id"`
		ImageURL   string `gorm:"column:url"`
		IsPrimary  int8   `gorm:"column:is_primary"`
		Sort       int    `gorm:"column:sort"`
	}

	// 从统一的 files 表查询
	err := DB.Table("files").
		Select("business_id, url, is_primary, sort").
		Where("business_type = 'mold' AND business_id IN ? AND deleted_at IS NULL", moldIDs).
		Order("is_primary DESC, sort ASC").
		Find(&images).Error

	if err != nil {
		return nil, err
	}

	result := make(map[int64]string)
	for _, img := range images {
		if _, exists := result[img.BusinessID]; !exists {
			result[img.BusinessID] = img.ImageURL
		}
	}

	return result, nil
}

// CountAll 统计模具总数
func (d *MoldDAO) CountAll() (int64, error) {
	var count int64
	err := DB.Model(&entity.Mold{}).Count(&count).Error
	return count, err
}

// CountByStatus 根据状态统计模具数量
func (d *MoldDAO) CountByStatus(status int8) (int64, error) {
	var count int64
	err := DB.Model(&entity.Mold{}).Where("life_cycle_status = ?", status).Count(&count).Error
	return count, err
}

// CountByStatusBatch 批量统计各状态模具数量（一次查询）
func (d *MoldDAO) CountByStatusBatch() (map[string]int64, error) {
	type StatusCount struct {
		LifeCycleStatus string
		Count           int64
	}

	var results []StatusCount
	err := DB.Model(&entity.Mold{}).
		Select("life_cycle_status, COUNT(*) as count").
		Group("life_cycle_status").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	statusMap := make(map[string]int64)
	for _, r := range results {
		statusMap[r.LifeCycleStatus] = r.Count
	}

	return statusMap, nil
}
