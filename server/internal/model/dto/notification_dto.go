package dto

// NotificationListDTO 通知列表返回 DTO
type NotificationListDTO struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	TypeLabel  string `json:"type_label"`
	Quality    string `json:"quality"`
	RefID      int64  `json:"ref_id"`
	RefURL     string `json:"ref_url"`
	IsRead     int8   `json:"is_read"`
	CreateTime string `json:"create_time"`
}

type NotificationPageQuery struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	IsRead   *int8  `json:"is_read" form:"is_read"`
	Type     string `json:"type" form:"type"` // 按通知类型筛选（如 mold:work_order_status）
}

type BatchReadDTO struct {
	IDs []int64 `json:"ids" binding:"required"`
}
