package common

import "net/http"

type HasBreadcrumb interface {
	Breadcrumb() (name, link string, parent HasBreadcrumb)
}

type HasPageTitle interface {
	PageTitle() string
}

type Page interface {
	http.Handler
}
