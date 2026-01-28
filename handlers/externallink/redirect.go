package externallink

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/sign"
	"golang.org/x/net/html"
)

func fetchOpenGraph(urlStr string) (title, desc, image string, err error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", "", "", err
	}

	host := u.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", "", "", err
	}

	for _, ip := range ips {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() {
			return "", "", "", fmt.Errorf("blocked internal ip: %s", ip)
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			// Re-check IP on redirect
			h := req.URL.Hostname()
			ips, err := net.LookupIP(h)
			if err != nil {
				return err
			}
			for _, ip := range ips {
				if ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() {
					return fmt.Errorf("blocked internal ip on redirect: %s", ip)
				}
			}
			return nil
		},
	}
	resp, err := client.Get(urlStr)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return "", "", "", err
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var prop, content, name string
			for _, a := range n.Attr {
				if a.Key == "property" {
					prop = a.Val
				}
				if a.Key == "content" {
					content = a.Val
				}
				if a.Key == "name" {
					name = a.Val
				}
			}
			switch prop {
			case "og:title":
				title = content
			case "og:description":
				desc = content
			case "og:image":
				image = content
			}
			if title == "" && name == "title" {
				title = content
			}
			if desc == "" && name == "description" {
				desc = content
			}
		}
		if n.Type == html.ElementNode && n.Data == "title" && title == "" {
			if n.FirstChild != nil {
				title = n.FirstChild.Data
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return title, desc, image, nil
}

// RedirectHandler shows a confirmation page before leaving the site or
// performs the redirect when the go parameter is present.
func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	if cd.LinkSignKey == "" {
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
		return
	}
	idStr := r.URL.Query().Get("id")
	rawURL := r.URL.Query().Get("u")
	sig := r.URL.Query().Get("sig")
	var linkID int32
	usedURL := false
	var existingLink *db.ExternalLink

	switch {
	case rawURL != "":
		data := "link:" + rawURL
		if err := sign.Verify(data, sig, cd.LinkSignKey, sign.WithOutNonce()); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
			return
		}
		usedURL = true
		if queries := cd.Queries(); queries != nil {
			if link, err := queries.GetExternalLink(r.Context(), rawURL); err == nil && link != nil {
				linkID = link.ID
				existingLink = link
			} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("load external link by url: %v", err)
			}
		}
	case idStr != "":
		id64, err := strconv.ParseInt(idStr, 10, 32)
		data := "link:" + idStr
		if err != nil || sign.Verify(data, sig, cd.LinkSignKey, sign.WithOutNonce()) != nil {
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
			return
		}
		linkID = int32(id64)
		if queries := cd.Queries(); queries != nil {
			if link, err := queries.GetExternalLinkByID(r.Context(), linkID); err == nil && link != nil {
				existingLink = link
				rawURL = link.Url
			}
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
		return
	}

	if r.URL.Query().Get("go") != "" {
		cd.RegisterExternalLinkClick(rawURL)
		http.Redirect(w, r, rawURL, http.StatusTemporaryRedirect)
		return
	}

	// Check if the URL is internal to avoid self-fetching
	isInternal := false
	if u, err := url.Parse(rawURL); err == nil {
		if u.Hostname() == r.Host {
			isInternal = true
		}
		// Also check against configured hostname
		if cd.Config != nil && cd.Config.HTTPHostname != "" {
			if confU, err := url.Parse(cd.Config.HTTPHostname); err == nil {
				if u.Hostname() == confU.Hostname() {
					isInternal = true
				}
			}
		}
	}

	needFetch := false
	if r.URL.Query().Get("reload") == "1" {
		needFetch = true
	} else if linkID == 0 && usedURL {
		needFetch = true
	} else if existingLink != nil && (!existingLink.CardTitle.Valid || !existingLink.CardDescription.Valid) {
		needFetch = true
	}

	if needFetch && rawURL != "" && !isInternal {
		title, desc, img, err := fetchOpenGraph(rawURL)
		if err == nil {
			if linkID == 0 {
				res, err := cd.Queries().CreateExternalLink(r.Context(), rawURL)
				if err == nil {
					id, _ := res.LastInsertId()
					linkID = int32(id)
				} else {
					log.Printf("CreateExternalLink error: %v", err)
				}
			}
			if linkID != 0 {
				err := cd.Queries().UpdateExternalLinkMetadata(r.Context(), db.UpdateExternalLinkMetadataParams{
					CardTitle:       sql.NullString{String: title, Valid: title != ""},
					CardDescription: sql.NullString{String: desc, Valid: desc != ""},
					CardImage:       sql.NullString{String: img, Valid: img != ""},
					ID:              linkID,
				})
				if err != nil {
					log.Printf("UpdateExternalLinkMetadata error: %v", err)
				}
			}
		} else {
			log.Printf("fetchOpenGraph error: %v", err)
		}
	} else if isInternal {
		// Log or handle internal link specifically if needed
		// For now, we just don't fetch metadata
	}

	if linkID != 0 {
		cd.SetCurrentExternalLinkID(linkID)
	}
	link := cd.SelectedExternalLink()
	if rawURL == "" {
		if link == nil {
			w.WriteHeader(http.StatusBadRequest)
			handlers.RenderErrorPage(w, r, fmt.Errorf("invalid link"))
			return
		}
		rawURL = link.Url
	}

	type Data struct {
		URL         string
		RedirectURL string
		ReloadURL   string
	}
	cd.PageTitle = "External Link"
	linkParam := "id"
	linkValue := idStr
	if usedURL {
		linkParam = "u"
		linkValue = url.QueryEscape(rawURL)
	}
	data := Data{
		URL:         rawURL,
		RedirectURL: fmt.Sprintf("/goto?%s=%s&sig=%s&go=1", linkParam, linkValue, sig),
		ReloadURL:   fmt.Sprintf("/goto?%s=%s&sig=%s&reload=1", linkParam, linkValue, sig),
	}
	if err := cd.ExecuteSiteTemplate(w, r, "externalLinkPage.gohtml", data); err != nil {
		log.Printf("Template Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
	}
}
