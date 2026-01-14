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

func (og *OpenGraph) ImageWidthMeta() template.HTML {
	if og.ImageWidth == 0 {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:image:width" content="%d" />`, og.ImageWidth))
}

func (og *OpenGraph) ImageHeightMeta() template.HTML {
	if og.ImageHeight == 0 {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:image:height" content="%d" />`, og.ImageHeight))
}

func (og *OpenGraph) TwitterImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta name="twitter:image" content="%s" />`, og.Image))
}

func (og *OpenGraph) TypeMeta() template.HTML {
	ogType := "website"
	if og.Type != "" {
		ogType = og.Type
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:type" content="%s" />`, ogType))
}
