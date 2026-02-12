package common_test

import (
	"testing"

	"github.com/arran4/goa4web/core/common"
)

func TestOffsetPagination_GetLinks(t *testing.T) {
	tests := []struct {
		name          string
		total         int
		pageSize      int
		offset        int
		baseURL       string
		paramName     string
		expectedLinks int
		expectedURLs  []string
		expectedActive int
	}{
		{
			name:          "No links (total <= pageSize)",
			total:         10,
			pageSize:      10,
			offset:        0,
			baseURL:       "/test",
			expectedLinks: 0,
		},
		{
			name:          "Two pages, first active",
			total:         15,
			pageSize:      10,
			offset:        0,
			baseURL:       "/test",
			expectedLinks: 2,
			expectedURLs:  []string{"/test?offset=0", "/test?offset=10"},
			expectedActive: 1,
		},
		{
			name:          "Two pages, second active",
			total:         15,
			pageSize:      10,
			offset:        10,
			baseURL:       "/test",
			expectedLinks: 2,
			expectedURLs:  []string{"/test?offset=0", "/test?offset=10"},
			expectedActive: 2,
		},
		{
			name:          "Custom param, existing query",
			total:         25,
			pageSize:      10,
			offset:        20,
			baseURL:       "/test?foo=bar",
			paramName:     "off",
			expectedLinks: 3,
			expectedURLs:  []string{"/test?foo=bar&off=0", "/test?foo=bar&off=10", "/test?foo=bar&off=20"},
			expectedActive: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &common.OffsetPagination{
				TotalItems: tt.total,
				PageSize:   tt.pageSize,
				Offset:     tt.offset,
				BaseURL:    tt.baseURL,
				ParamName:  tt.paramName,
			}
			links := p.GetLinks()
			if len(links) != tt.expectedLinks {
				t.Errorf("Expected %d links, got %d", tt.expectedLinks, len(links))
			}
			if len(links) > 0 {
				for i, l := range links {
					if l.Link != tt.expectedURLs[i] {
						t.Errorf("Link %d: expected %s, got %s", i, tt.expectedURLs[i], l.Link)
					}
					if (l.Num == tt.expectedActive) != l.Active {
						t.Errorf("Link %d: expected active=%v, got %v", i, l.Num == tt.expectedActive, l.Active)
					}
				}
			}
		})
	}
}

func TestPageNumberPagination_GetLinks(t *testing.T) {
	tests := []struct {
		name          string
		total         int
		pageSize      int
		currentPage   int
		baseURL       string
		paramName     string
		expectedLinks int
		expectedURLs  []string
		expectedActive int
	}{
		{
			name:          "No links",
			total:         10,
			pageSize:      10,
			currentPage:   1,
			baseURL:       "/test",
			expectedLinks: 0,
		},
		{
			name:          "Two pages",
			total:         15,
			pageSize:      10,
			currentPage:   1,
			baseURL:       "/test",
			expectedLinks: 2,
			expectedURLs:  []string{"/test?page=1", "/test?page=2"},
			expectedActive: 1,
		},
		{
			name:          "Existing query",
			total:         15,
			pageSize:      10,
			currentPage:   2,
			baseURL:       "/test?q=a",
			expectedLinks: 2,
			expectedURLs:  []string{"/test?q=a&page=1", "/test?q=a&page=2"},
			expectedActive: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &common.PageNumberPagination{
				TotalItems:  tt.total,
				PageSize:    tt.pageSize,
				CurrentPage: tt.currentPage,
				BaseURL:     tt.baseURL,
				ParamName:   tt.paramName,
			}
			links := p.GetLinks()
			if len(links) != tt.expectedLinks {
				t.Errorf("Expected %d links, got %d", tt.expectedLinks, len(links))
			}
			if len(links) > 0 {
				for i, l := range links {
					if l.Link != tt.expectedURLs[i] {
						t.Errorf("Link %d: expected %s, got %s", i, tt.expectedURLs[i], l.Link)
					}
					if (l.Num == tt.expectedActive) != l.Active {
						t.Errorf("Link %d: expected active=%v, got %v", i, l.Num == tt.expectedActive, l.Active)
					}
				}
			}
		})
	}
}
