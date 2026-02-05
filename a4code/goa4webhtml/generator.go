package goa4webhtml

import (
	"github.com/arran4/goa4web/a4code/html"
)

type Generator struct {
	*html.Generator
}

func NewGenerator() *Generator {
	return &Generator{
		Generator: html.NewGenerator(),
	}
}
