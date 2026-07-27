package dto

type SearchParams struct {
	Keyword  string `form:"keyword" json:"keyword"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"page_size" json:"page_size"`
}

type SearchResultItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Module      string `json:"module"`
	ModuleName  string `json:"module_name"`
	Path        string `json:"path"`
}

type SearchResponse struct {
	Items []SearchResultItem `json:"items"`
	Total int64              `json:"total"`
}
