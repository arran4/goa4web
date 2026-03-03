package common

import (
	"fmt"
	"html"
	"html/template"
)

func (og *OpenGraph) URLMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:url" content="%s" />`, html.EscapeString(og.URL)))
}

func (og *OpenGraph) ImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image" content="%s" />`, html.EscapeString(og.Image)))
}

func (og *OpenGraph) SecureImageMeta() template.HTML {
	return template.HTML(fmt.Sprintf(`<meta property="og:image:secure_url" content="%s" />`, html.EscapeString(og.Image)))
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
	return template.HTML(fmt.Sprintf(`<meta name="twitter:image" content="%s" />`, html.EscapeString(og.Image)))
}

func (og *OpenGraph) TypeMeta() template.HTML {
	ogType := "website"
	if og.Type != "" {
		ogType = og.Type
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:type" content="%s" />`, html.EscapeString(ogType)))
}

func (og *OpenGraph) ExpirationTimeMeta() template.HTML {
	if og.ExpirationTime == nil {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="article:expiration_time" content="%s" />`, og.ExpirationTime.Format("2006-01-02T15:04:05Z07:00")))
}

func (og *OpenGraph) PublishedTimeMeta() template.HTML {
	if og.PublishedTime == nil {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="article:published_time" content="%s" />`, og.PublishedTime.Format("2006-01-02T15:04:05Z07:00")))
}

func (og *OpenGraph) ModifiedTimeMeta() template.HTML {
	if og.ModifiedTime == nil {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="article:modified_time" content="%s" />`, og.ModifiedTime.Format("2006-01-02T15:04:05Z07:00")))
}

func (og *OpenGraph) SiteNameMeta() template.HTML {
	if og.SiteName == "" {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:site_name" content="%s" />`, html.EscapeString(og.SiteName)))
}

func (og *OpenGraph) UpdatedTimeMeta() template.HTML {
	if og.UpdatedTime == nil {
		return ""
	}
	return template.HTML(fmt.Sprintf(`<meta property="og:updated_time" content="%s" />`, og.UpdatedTime.Format("2006-01-02T15:04:05Z07:00")))
}
