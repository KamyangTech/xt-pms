package dto

// CompanyDTO 公司信息DTO
type CompanyDTO struct {
	ID               int64  `json:"id"`
	CompanyName      string `json:"company_name" binding:"required"`
	ShortName        string `json:"short_name"`
	EnglishName      string `json:"english_name"`
	EnglishShortName string `json:"english_short_name"`
	Logo             string `json:"logo"`
}
