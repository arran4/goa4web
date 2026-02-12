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

func TestOffsetPagination_NavLinks(t *testing.T) {
	p := &common.OffsetPagination{
		TotalItems: 30,
		PageSize:   10,
		Offset:     10,
		BaseURL:    "/test",
	}

	if link := p.StartLink(); link != "/test?offset=0" {
		t.Errorf("StartLink: got %s, want /test?offset=0", link)
	}
	if link := p.PrevLink(); link != "/test?offset=0" {
		t.Errorf("PrevLink: got %s, want /test?offset=0", link)
	}
	if link := p.NextLink(); link != "/test?offset=20" {
		t.Errorf("NextLink: got %s, want /test?offset=20", link)
	}

	// First page
	p.Offset = 0
	if link := p.StartLink(); link != "" {
		t.Errorf("StartLink (first page): got %s, want empty", link)
	}
	if link := p.PrevLink(); link != "" {
		t.Errorf("PrevLink (first page): got %s, want empty", link)
	}

	// Last page
	p.Offset = 20
	if link := p.NextLink(); link != "" {
		t.Errorf("NextLink (last page): got %s, want empty", link)
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

func TestPageNumberPagination_NavLinks(t *testing.T) {
	p := &common.PageNumberPagination{
		TotalItems:  30,
		PageSize:    10,
		CurrentPage: 2,
		BaseURL:     "/test",
	}

	if link := p.StartLink(); link != "/test?page=1" {
		t.Errorf("StartLink: got %s, want /test?page=1", link)
	}
	if link := p.PrevLink(); link != "/test?page=1" {
		t.Errorf("PrevLink: got %s, want /test?page=1", link)
	}
	if link := p.NextLink(); link != "/test?page=3" {
		t.Errorf("NextLink: got %s, want /test?page=3", link)
	}

	// First page
	p.CurrentPage = 1
	if link := p.StartLink(); link != "" {
		t.Errorf("StartLink (first page): got %s, want empty", link)
	}
	if link := p.PrevLink(); link != "" {
		t.Errorf("PrevLink (first page): got %s, want empty", link)
	}

	// Last page
	p.CurrentPage = 3
	if link := p.NextLink(); link != "" {
		t.Errorf("NextLink (last page): got %s, want empty", link)
	}
}
