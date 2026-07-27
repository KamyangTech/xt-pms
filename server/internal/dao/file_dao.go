package dao

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type FileDAO struct{}

// NewFileDAO 创建文件DAO实例
func NewFileDAO() *FileDAO {
	return &FileDAO{}
}

// Create 创建文件记录
func (d *FileDAO) Create(file *entity.File) error {
	return DB.Create(file).Error
}

// FindByID 根据ID获取文件
func (d *FileDAO) FindByID(id string) (*entity.File, error) {
	var file entity.File
	err := DB.Where("id = ?", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// FindByIDs 根据ID列表批量查询文件（优化N+1）
func (d *FileDAO) FindByIDs(ids []string) ([]entity.File, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var files []entity.File
	err := DB.Where("id IN ?", ids).Find(&files).Error
	return files, err
}

// FindByIDsMap 批量查询并返回 map[id]*File，方便内存映射
func (d *FileDAO) FindByIDsMap(ids []string) (map[string]*entity.File, error) {
	files, err := d.FindByIDs(ids)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*entity.File, len(files))
	for i := range files {
		result[files[i].ID] = &files[i]
	}
	return result, nil
}

// FindByURL 根据URL查找文件记录
func (d *FileDAO) FindByURL(url string) (*entity.File, error) {
	var file entity.File
	err := DB.Where("url = ?", url).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// FindAllFileURLs 获取数据库中所有文件URL列表（用于磁盘孤文件对比）
func (d *FileDAO) FindAllFileURLs() ([]string, error) {
	var urls []string
	err := DB.Model(&entity.File{}).Select("url").Find(&urls).Error
	return urls, err
}

// FindByBusiness 根据业务类型和业务ID获取文件列表（按创建时间正序）
func (d *FileDAO) FindByBusiness(businessType string, businessID int64) ([]entity.File, error) {
	var files []entity.File
	err := DB.Where("business_type = ? AND business_id = ?", businessType, businessID).
		Order("is_primary DESC, sort ASC, create_time ASC").
		Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// FindPrimary 查找主图
func (d *FileDAO) FindPrimary(businessType string, businessID int64) (*entity.File, error) {
	var file entity.File
	err := DB.Where("business_type = ? AND business_id = ? AND is_primary = 1", businessType, businessID).
		First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// SetPrimary 设置主图（将该业务下其他图片取消主图，并将指定文件设为主图）
func (d *FileDAO) SetPrimary(businessType string, businessID int64, fileID string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		// 将该业务下的所有文件取消主图
		if err := tx.Model(&entity.File{}).
			Where("business_type = ? AND business_id = ?", businessType, businessID).
			Update("is_primary", 0).Error; err != nil {
			return err
		}
		// 将指定文件设为主图并排序置为 0
		if err := tx.Model(&entity.File{}).Where("id = ?", fileID).
			Updates(map[string]interface{}{
				"is_primary": 1,
				"sort":       0,
			}).Error; err != nil {
			return err
		}
		// 其他图片排序 +1
		return tx.Model(&entity.File{}).
			Where("business_type = ? AND business_id = ? AND id != ?", businessType, businessID, fileID).
			Update("sort", gorm.Expr("sort + 1")).Error
	})
}

// IncrementRefCount 增加文件的引用计数（使用指定DB实例，支持事务）
func (d *FileDAO) IncrementRefCount(url string, db *gorm.DB) error {
	return db.Model(&entity.File{}).
		Where("url = ?", url).
		UpdateColumn("ref_count", gorm.Expr("ref_count + 1")).Error
}

// DecrementRefCount 减少文件的引用计数，减到0时自动标记软删除（使用指定DB实例，支持事务）
func (d *FileDAO) DecrementRefCount(url string, db *gorm.DB) error {
	return db.Exec(
		"UPDATE files SET ref_count = GREATEST(ref_count - 1, 0), "+
			"deleted_at = CASE WHEN ref_count <= 1 THEN NOW() ELSE deleted_at END "+
			"WHERE url = ?", url).Error
}

// MarkDeleted 标记引用计数为0的文件为软删除状态
func (d *FileDAO) MarkDeleted(db *gorm.DB) error {
	return db.Model(&entity.File{}).
		Where("ref_count = 0 AND deleted_at IS NULL").
		UpdateColumn("deleted_at", time.Now()).Error
}

// CountAllOrphanFiles 统计所有孤文件数量
func (d *FileDAO) CountAllOrphanFiles() (int64, error) {
	var count int64
	err := DB.Model(&entity.File{}).
		Where("ref_count <= 0 AND deleted_at IS NOT NULL").
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// FindAllOrphanFiles 查找所有孤文件（ref_count <= 0 且已标记软删除）
func (d *FileDAO) FindAllOrphanFiles() ([]entity.File, error) {
	var files []entity.File
	err := DB.Where("ref_count <= 0 AND deleted_at IS NOT NULL").
		Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// FindOrphanFiles 查找超过指定时间的孤文件（供定时任务使用，需传入保留期边界）
func (d *FileDAO) FindOrphanFiles(before time.Time) ([]entity.File, error) {
	var files []entity.File
	err := DB.Where("ref_count <= 0 AND deleted_at IS NOT NULL AND deleted_at < ?", before).
		Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

// HardDeleteByID 硬删除文件记录（仅用于清理孤文件时使用）
func (d *FileDAO) HardDeleteByID(id string) error {
	return DB.Unscoped().Where("id = ?", id).Delete(&entity.File{}).Error
}

// DeleteByID 根据ID删除文件记录
func (d *FileDAO) DeleteByID(id string) error {
	return DB.Where("id = ?", id).Delete(&entity.File{}).Error
}

// DeleteByBusiness 根据业务类型和业务ID删除所有关联文件记录
func (d *FileDAO) DeleteByBusiness(businessType string, businessID int64) error {
	return DB.Where("business_type = ? AND business_id = ?", businessType, businessID).
		Delete(&entity.File{}).Error
}

// Update 更新文件记录（使用 struct，注意将写全所有字段）
func (d *FileDAO) Update(file *entity.File) error {
	return DB.Save(file).Error
}

// UpdateByMap 按需更新文件记录的指定字段（避免零值覆盖已有字段）
func (d *FileDAO) UpdateByMap(id string, values map[string]interface{}) error {
	return DB.Model(&entity.File{}).Where("id = ?", id).Updates(values).Error
}

// UpdateBusinessID 批量更新文件的业务关联信息
func (d *FileDAO) UpdateBusinessID(fileIDs []string, businessType string, businessID int64) error {
	return DB.Model(&entity.File{}).
		Where("id IN ?", fileIDs).
		Updates(map[string]interface{}{
			"business_type": businessType,
			"business_id":   businessID,
		}).Error
}
