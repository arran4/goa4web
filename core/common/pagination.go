package common

import (
	"fmt"
	"strings"
)

// PageLink represents a numbered pagination link.
type PageLink struct {
	Num    int
	Link   string
	Active bool
}

// Pagination defines an interface for generating pagination links.
type Pagination interface {
	GetLinks() []PageLink
	StartLink() string
	PrevLink() string
	NextLink() string
}

// OffsetPagination implements Pagination for offset-based paging.
type OffsetPagination struct {
	TotalItems int
	PageSize   int
	Offset     int
	BaseURL    string
	ParamName  string // Defaults to "offset"
}

func (op *OffsetPagination) fmtLink(offset int) string {
	sep := "?"
	if strings.Contains(op.BaseURL, "?") {
		sep = "&"
	}
	param := op.ParamName
	if param == "" {
		param = "offset"
	}
	return fmt.Sprintf("%s%s%s=%d", op.BaseURL, sep, param, offset)
}

func (op *OffsetPagination) StartLink() string {
	if op.Offset <= 0 {
		return ""
	}
	return op.fmtLink(0)
}

func (op *OffsetPagination) PrevLink() string {
	if op.Offset <= 0 {
		return ""
	}
	prev := op.Offset - op.PageSize
	if prev < 0 {
		prev = 0
	}
	return op.fmtLink(prev)
}

func (op *OffsetPagination) NextLink() string {
	if op.Offset+op.PageSize >= op.TotalItems {
		return ""
	}
	return op.fmtLink(op.Offset + op.PageSize)
}

func (op *OffsetPagination) GetLinks() []PageLink {
	if op.PageSize <= 0 {
		return nil
	}
	pages := (op.TotalItems + op.PageSize - 1) / op.PageSize
	if pages <= 1 {
		return nil
	}
	currentPage := op.Offset/op.PageSize + 1

	var links []PageLink
	for i := 1; i <= pages; i++ {
		links = append(links, PageLink{
			Num:    i,
			Link:   op.fmtLink((i - 1) * op.PageSize),
			Active: i == currentPage,
		})
	}
	return links
}

// PageNumberPagination implements Pagination for page-number-based paging.
type PageNumberPagination struct {
	TotalItems  int
	PageSize    int
	CurrentPage int
	BaseURL     string
	ParamName   string // Defaults to "page"
}

func (pp *PageNumberPagination) fmtLink(page int) string {
	sep := "?"
	if strings.Contains(pp.BaseURL, "?") {
		sep = "&"
	}
	param := pp.ParamName
	if param == "" {
		param = "page"
	}
	return fmt.Sprintf("%s%s%s=%d", pp.BaseURL, sep, param, page)
}

func (pp *PageNumberPagination) StartLink() string {
	if pp.CurrentPage <= 1 {
		return ""
	}
	return pp.fmtLink(1)
}

func (pp *PageNumberPagination) PrevLink() string {
	if pp.CurrentPage <= 1 {
		return ""
	}
	return pp.fmtLink(pp.CurrentPage - 1)
}

func (pp *PageNumberPagination) NextLink() string {
	if pp.PageSize <= 0 {
		return ""
	}
	pages := (pp.TotalItems + pp.PageSize - 1) / pp.PageSize
	if pp.CurrentPage >= pages {
		return ""
	}
	return pp.fmtLink(pp.CurrentPage + 1)
}

func (pp *PageNumberPagination) GetLinks() []PageLink {
	if pp.PageSize <= 0 {
		return nil
	}
	pages := (pp.TotalItems + pp.PageSize - 1) / pp.PageSize
	if pages <= 1 {
		return nil
	}

	var links []PageLink
	for i := 1; i <= pages; i++ {
		links = append(links, PageLink{
			Num:    i,
			Link:   pp.fmtLink(i),
			Active: i == pp.CurrentPage,
		})
	}
	return links
}
