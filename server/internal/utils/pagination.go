package utils

import (
	"gorm.io/gorm"
)

type PageResult struct {
	List  interface{} `json:"list"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"page_size"`
}

// Paginate 通用分页查询函数
// db: GORM 数据库实例
// page: 当前页码
// pageSize: 每页数量
// result: 接收结果的切片指针
// opts: 可选的 GORM 查询选项（如 Preload, Where 等）
func Paginate(db *gorm.DB, page, pageSize int, result interface{}, opts ...func(*gorm.DB) *gorm.DB) (*PageResult, error) {
	var total int64

	// 克隆 db 以避免影响原始查询对象
	query := db.Session(&gorm.Session{})

	// 应用额外的查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.Model(result).Count(&total).Error; err != nil {
		return nil, err
	}

	// 执行分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(result).Error; err != nil {
		return nil, err
	}

	return &PageResult{
		List:  result,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}
