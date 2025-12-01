package pagination

type Params struct {
	Page     int
	PageSize int
}

func NewParams(page, pageSize int) Params {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return Params{
		Page:     page,
		PageSize: pageSize,
	}
}

func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p Params) Limit() int {
	return p.PageSize
}

type Response struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

func NewResponse(params Params, totalItems int) Response {
	totalPages := totalItems / params.PageSize
	if totalItems%params.PageSize > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}

	return Response{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}
