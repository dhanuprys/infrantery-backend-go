package dto

// PaginationParams represents common pagination parameters
type PaginationParams struct {
	Page     int `form:"page" json:"page"`           // Page number (1-indexed)
	PageSize int `form:"page_size" json:"page_size"` // Items per page
}

// DefaultPaginationParams returns default pagination values
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 20,
	}
}

// Validate validates and normalizes pagination parameters
func (p *PaginationParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 || p.PageSize > 100 {
		p.PageSize = 20
	}
}

// GetOffset calculates the MongoDB skip value
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the page size
func (p *PaginationParams) GetLimit() int {
	return p.PageSize
}

// PaginationMeta represents pagination metadata in responses
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	HasNextPage bool  `json:"has_next_page"`
	HasPrevPage bool  `json:"has_prev_page"`
}

// NewPaginationMeta creates pagination metadata
func NewPaginationMeta(params PaginationParams, totalItems int64) PaginationMeta {
	totalPages := int((totalItems + int64(params.PageSize) - 1) / int64(params.PageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationMeta{
		CurrentPage: params.Page,
		PageSize:    params.PageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNextPage: params.Page < totalPages,
		HasPrevPage: params.Page > 1,
	}
}
