package common

import (
	"fmt"
	"html/template"
)

func (og *OpenGraph) URLMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:url" content="%s" />`, og.URL))
}

func (og *OpenGraph) ImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image" content="%s" />`, og.Image))
}

func (og *OpenGraph) SecureImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image:secure_url" content="%s" />`, og.Image))
}
