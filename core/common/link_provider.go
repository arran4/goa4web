package common

import (
	"context"
	"fmt"
	"html"
	"strings"

	"github.com/arran4/goa4web/a4code/a4code2html"
)

type Goa4WebLinkProvider struct {
	cd  *CoreData
	ctx context.Context
}

func NewGoa4WebLinkProvider(cd *CoreData, ctx context.Context) *Goa4WebLinkProvider {
	return &Goa4WebLinkProvider{
		cd:  cd,
		ctx: ctx,
	}
}

func (p *Goa4WebLinkProvider) MapImageURL(tag, val string) string {
	if tag == "img" {
		return p.cd.MapImageURL(tag, val)
	}
	return val
}

func buildTooltip(title, description string) string {
	if title != "" && description != "" {
		return title + " - " + description
	}
	if title != "" {
		return title
	}
	return description
}

func (p *Goa4WebLinkProvider) RenderLink(rawURL string, isBlock bool, isImmediateClose bool) (string, string, bool) {
	safe, ok := a4code2html.SanitizeURL(rawURL)
	if !ok {
		return html.EscapeString(rawURL), "", false
	}

	targetURL := safe
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		targetURL = p.cd.SignLinkURL(rawURL)
	}

	var title, description, imageURL, faviconURL string
	var hasData bool

	if p.cd.Queries() != nil {
		link, err := p.cd.Queries().GetExternalLink(p.ctx, rawURL)
		if err == nil {
			hasData = true
			title = link.CardTitle.String
			description = link.CardDescription.String
			imageURL = link.CardImage.String
			if link.CardImageCache.Valid && link.CardImageCache.String != "" {
				imageURL = p.cd.MapImageURL("img", link.CardImageCache.String)
			}
			if link.FaviconCache.Valid && link.FaviconCache.String != "" {
				faviconURL = p.cd.MapImageURL("img", link.FaviconCache.String)
			}
		}
	}

	var tooltip string
	if hasData {
		meta := buildTooltip(title, description)
		if meta != "" {
			tooltip = rawURL + " - " + meta
		} else {
			tooltip = rawURL
		}
	} else {
		tooltip = rawURL
	}

	hasUserTitle := !isImmediateClose

	if !isBlock {
		faviconHTML := ""
		if faviconURL != "" {
			safeFav, favOk := a4code2html.SanitizeURL(faviconURL)
			if favOk {
				faviconHTML = fmt.Sprintf(`<img src="%s" class="a4code-inline-favicon" />`, safeFav)
			}
		}

		linkOpen := fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="%s">`, targetURL, html.EscapeString(tooltip))

		if hasUserTitle {
			// [link=url]Title[/link]
			return linkOpen + faviconHTML, "</a>", false
		} else {
			// [link url] -> [link]url[/link] logic in immediate close
			linkText := html.EscapeString(rawURL)
			if hasData {
				displayText := title
				if displayText == "" {
					if len(description) > 50 {
						displayText = description[:47] + "..."
					} else {
						displayText = description
					}
				}
				if displayText != "" {
					linkText = html.EscapeString(displayText)
				}
			}
			return fmt.Sprintf(`%s%s%s</a>`, linkOpen, faviconHTML, linkText), "", true
		}
	}

	// Block rendering logic

	// Online + Title
	if hasUserTitle {
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="%s">`, targetURL, html.EscapeString(tooltip)), "</a>", false
	}

	// Online + No Title (Card or Text)
	if !hasData {
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="%s">%s</a>`, targetURL, html.EscapeString(tooltip), html.EscapeString(rawURL)), "", true
	}

	// Card
	imageHTML := ""
	if imageURL != "" {
		safeImg, imgOk := a4code2html.SanitizeURL(imageURL)
		if imgOk {
			imageHTML = fmt.Sprintf(`<img src="%s" class="external-link-image" />`, safeImg)
		}
	}

	displayTitle := title
	if displayTitle == "" {
		displayTitle = rawURL
	}

	htmlStr := fmt.Sprintf(
		`<div class="external-link-card"><a href="%s" target="_blank" rel="noopener noreferrer" class="external-link-card-inner" title="%s">%s<div class="external-link-content"><div class="external-link-title">%s</div><div class="external-link-description">%s</div></div></a></div>`,
		targetURL, html.EscapeString(tooltip), imageHTML, html.EscapeString(displayTitle), html.EscapeString(description))

	return htmlStr, "", true
}
