package pagination

import "testing"

func TestNewParams(t *testing.T) {
	tests := []struct {
		name         string
		page         int
		pageSize     int
		expectedPage int
		expectedSize int
	}{
		{"valid params", 1, 20, 1, 20},
		{"valid params page 2", 2, 50, 2, 50},
		{"page less than 1 defaults to 1", 0, 20, 1, 20},
		{"negative page defaults to 1", -5, 20, 1, 20},
		{"pageSize less than 1 defaults to 50", 1, 0, 1, 50},
		{"negative pageSize defaults to 50", 1, -10, 1, 50},
		{"pageSize greater than 100 caps at 100", 1, 200, 1, 100},
		{"page 100", 1, 101, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := NewParams(tt.page, tt.pageSize)
			if params.Page != tt.expectedPage {
				t.Errorf("Page = %d, want %d", params.Page, tt.expectedPage)
			}
			if params.PageSize != tt.expectedSize {
				t.Errorf("PageSize = %d, want %d", params.PageSize, tt.expectedSize)
			}
		})
	}
}

func TestParamsOffset(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedOffset int
	}{
		{"first page", 1, 20, 0},
		{"second page", 2, 20, 20},
		{"third page", 3, 20, 40},
		{"page 10 size 50", 10, 50, 450},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := Params{Page: tt.page, PageSize: tt.pageSize}
			offset := params.Offset()
			if offset != tt.expectedOffset {
				t.Errorf("Offset() = %d, want %d", offset, tt.expectedOffset)
			}
		})
	}
}

func TestParamsLimit(t *testing.T) {
	params := Params{Page: 1, PageSize: 25}
	if params.Limit() != 25 {
		t.Errorf("Limit() = %d, want 25", params.Limit())
	}
}

func TestNewResponse(t *testing.T) {
	tests := []struct {
		name          string
		params        Params
		totalItems    int
		expectedPages int
		expectedTotal int
	}{
		{"exact division", Params{1, 20}, 100, 5, 100},
		{"with remainder", Params{1, 20}, 105, 6, 105},
		{"less than page size", Params{1, 20}, 15, 1, 15},
		{"zero items", Params{1, 20}, 0, 1, 0},
		{"one item", Params{1, 20}, 1, 1, 1},
		{"large dataset", Params{1, 50}, 1000, 20, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := NewResponse(tt.params, tt.totalItems)
			if response.Page != tt.params.Page {
				t.Errorf("Page = %d, want %d", response.Page, tt.params.Page)
			}
			if response.PageSize != tt.params.PageSize {
				t.Errorf("PageSize = %d, want %d", response.PageSize, tt.params.PageSize)
			}
			if response.TotalPages != tt.expectedPages {
				t.Errorf("TotalPages = %d, want %d", response.TotalPages, tt.expectedPages)
			}
			if response.TotalItems != tt.expectedTotal {
				t.Errorf("TotalItems = %d, want %d", response.TotalItems, tt.expectedTotal)
			}
		})
	}
}
