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

func (p *Goa4WebLinkProvider) RenderLink(rawURL string, isBlock bool, isImmediateClose bool) (string, string, bool) {
	safe, ok := a4code2html.SanitizeURL(rawURL)
	if !ok {
		return html.EscapeString(rawURL), "", false
	}

	targetURL := safe
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		targetURL = p.cd.SignLinkURL(rawURL)
	}

	var title, description, imageURL string
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
		}
	}

	hasUserTitle := !isImmediateClose

	// Inline + Title + Without data -> No change
	if !isBlock && hasUserTitle && !hasData {
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">`, targetURL), "</a>", false
	}

	// Inline + Title + With data -> Title & Description added as alt text
	if !isBlock && hasUserTitle && hasData {
		altText := title
		if description != "" {
			altText += " - " + description
		}
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="%s">`, targetURL, html.EscapeString(altText)), "</a>", false
	}

	// Inline + No title + Without data -> URL as link text
	if !isBlock && !hasUserTitle && !hasData {
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, targetURL, html.EscapeString(rawURL)), "", true
	}

	// Inline + No title + With data -> Title/Description as text
	if !isBlock && !hasUserTitle && hasData {
		linkText := title
		if linkText == "" {
			if len(description) > 50 {
				linkText = description[:47] + "..."
			} else {
				linkText = description
			}
		}
		if linkText == "" {
			linkText = rawURL
		}
		altText := description
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="%s">%s</a>`, targetURL, html.EscapeString(altText), html.EscapeString(linkText)), "", true
	}

	// Online + Title + Without data -> No change
	if isBlock && hasUserTitle && !hasData {
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">`, targetURL), "</a>", false
	}

	// Online + Title + With data -> Title & Description added as alt text
	if isBlock && hasUserTitle && hasData {
		altText := title
		if description != "" {
			altText += " - " + description
		}
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer" title="%s">`, targetURL, html.EscapeString(altText)), "</a>", false
	}

	// Online + No title + Without data -> URL as link text
	if isBlock && !hasUserTitle && !hasData {
		return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">%s</a>`, targetURL, html.EscapeString(rawURL)), "", true
	}

	// Online + No title + With data -> Link Card
	if isBlock && !hasUserTitle && hasData {
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
			`<div class="external-link-card"><a href="%s" target="_blank" rel="noopener noreferrer" class="external-link-card-inner">%s<div class="external-link-content"><div class="external-link-title">%s</div><div class="external-link-description">%s</div></div></a></div>`,
			safe, imageHTML, html.EscapeString(displayTitle), html.EscapeString(description))

		return htmlStr, "", true
	}

	return fmt.Sprintf(`<a href="%s" target="_blank" rel="noopener noreferrer">`, safe), "</a>", false
}
