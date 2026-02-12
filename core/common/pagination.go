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
}

// OffsetPagination implements Pagination for offset-based paging.
type OffsetPagination struct {
	TotalItems int
	PageSize   int
	Offset     int
	BaseURL    string
	ParamName  string // Defaults to "offset"
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
	param := op.ParamName
	if param == "" {
		param = "offset"
	}

	var links []PageLink
	sep := "?"
	if strings.Contains(op.BaseURL, "?") {
		sep = "&"
	}

	for i := 1; i <= pages; i++ {
		link := fmt.Sprintf("%s%s%s=%d", op.BaseURL, sep, param, (i-1)*op.PageSize)
		links = append(links, PageLink{
			Num:    i,
			Link:   link,
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

func (pp *PageNumberPagination) GetLinks() []PageLink {
	if pp.PageSize <= 0 {
		return nil
	}
	pages := (pp.TotalItems + pp.PageSize - 1) / pp.PageSize
	if pages <= 1 {
		return nil
	}

	param := pp.ParamName
	if param == "" {
		param = "page"
	}

	var links []PageLink
	sep := "?"
	if strings.Contains(pp.BaseURL, "?") {
		sep = "&"
	}

	for i := 1; i <= pages; i++ {
		link := fmt.Sprintf("%s%s%s=%d", pp.BaseURL, sep, param, i)
		links = append(links, PageLink{
			Num:    i,
			Link:   link,
			Active: i == pp.CurrentPage,
		})
	}
	return links
}
