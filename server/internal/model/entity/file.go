package entity

import "time"

// File 文件表 - 统一文件存储记录
type File struct {
	ID           string     `gorm:"primaryKey;size:36;comment:UUID" json:"id"`
	URL          string     `gorm:"size:500;not null;comment:存储访问路径" json:"url"`
	FileName     string     `gorm:"size:255;not null;comment:原始文件名" json:"file_name"`
	FileSize     int64      `gorm:"not null;default:0;comment:文件大小（字节）" json:"file_size"`
	FileType     string     `gorm:"size:20;not null;default:'';comment:文件扩展名" json:"file_type"`
	RefCount     int        `gorm:"not null;default:0;comment:引用计数" json:"ref_count"`
	DeletedAt    *time.Time `gorm:"index;comment:软删除时间" json:"deleted_at"`
	BusinessType string     `gorm:"size:50;comment:业务类型" json:"business_type"`
	BusinessID   *int64     `gorm:"comment:业务记录ID" json:"business_id"`
	IsPrimary    int8       `gorm:"not null;default:0;comment:是否主图(0否 1是)" json:"is_primary"`
	Sort         int        `gorm:"not null;default:0;comment:排序" json:"sort"`
	CreatedBy    *int64     `gorm:"comment:上传用户ID" json:"created_by"`
	CreateTime   time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
}

func (File) TableName() string {
	return "files"
}
