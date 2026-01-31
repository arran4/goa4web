package opengraph

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Info contains extracted metadata.
type Info struct {
	Title       string
	Description string
	Image       string
	Duration    string
	UploadDate  string
	Author      string
}

// NewSafeClient returns an http.Client configured to block internal IP addresses.
func NewSafeClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Second,
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
}

func Fetch(urlStr string, client *http.Client) (*Info, error) {
	if client == nil {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}

		host := u.Hostname()
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, err
		}

		for _, ip := range ips {
			if ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() {
				return nil, fmt.Errorf("blocked internal ip: %s", ip)
			}
		}

		client = NewSafeClient()
	}
	resp, err := client.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, err
	}

	info := &Info{}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "script" {
				isLD := false
				for _, a := range n.Attr {
					if a.Key == "type" && a.Val == "application/ld+json" {
						isLD = true
						break
					}
				}
				if isLD && n.FirstChild != nil {
					parseJSONLD(n.FirstChild.Data, info)
				}
			} else if n.Data == "meta" {
				var prop, content, name, itemprop string
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
					if a.Key == "itemprop" {
						itemprop = a.Val
					}
				}
				switch prop {
				case "og:title":
					if info.Title == "" {
						info.Title = content
					}
				case "og:description":
					if info.Description == "" {
						info.Description = content
					}
				case "og:image":
					if info.Image == "" {
						info.Image = content
					}
				}

				if info.Title == "" && name == "title" {
					info.Title = content
				}
				if info.Description == "" && name == "description" {
					info.Description = content
				}

				// Fallbacks for new fields
				if info.Duration == "" && itemprop == "duration" {
					info.Duration = content
				}
				if info.UploadDate == "" && itemprop == "uploadDate" {
					info.UploadDate = content
				}
				// "uploadDate" can also be name? rare but possible.
				if info.UploadDate == "" && name == "uploadDate" {
					info.UploadDate = content
				}
				if info.Author == "" && itemprop == "author" {
					info.Author = content
				}
				if info.Author == "" && name == "author" {
					info.Author = content
				}
			} else if n.Data == "title" && info.Title == "" {
				if n.FirstChild != nil {
					info.Title = n.FirstChild.Data
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return info, nil
}

func parseJSONLD(data string, info *Info) {
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return
	}

	process := func(obj map[string]interface{}) {
		typeVal, _ := obj["@type"].(string)

		getString := func(key string) string {
			if s, ok := obj[key].(string); ok {
				return s
			}
			return ""
		}

		getAuthor := func() string {
			val := obj["author"]
			if s, ok := val.(string); ok {
				return s
			}
			if m, ok := val.(map[string]interface{}); ok {
				if name, ok := m["name"].(string); ok {
					return name
				}
			}
			if s, ok := val.([]interface{}); ok && len(s) > 0 {
				if m, ok := s[0].(map[string]interface{}); ok {
					if name, ok := m["name"].(string); ok {
						return name
					}
				}
			}
			return ""
		}

		// Prioritize VideoObject, but could extract generic info too if empty
		if typeVal == "VideoObject" || strings.EqualFold(typeVal, "VideoObject") {
			// We prioritize JSON-LD over meta tags, so overwrite or only set if empty?
			// The request says "Prioritize JSON-LD". So we should overwrite if we found it here.
			// However, parseJSONLD is called during traversal.

			if t := getString("name"); t != "" {
				info.Title = t
			}
			if d := getString("description"); d != "" {
				info.Description = d
			}
			if dur := getString("duration"); dur != "" {
				info.Duration = dur
			}
			if ud := getString("uploadDate"); ud != "" {
				info.UploadDate = ud
			}
			if auth := getAuthor(); auth != "" {
				info.Author = auth
			}

			if img := getString("thumbnailUrl"); img != "" {
				info.Image = img
			} else if img := getString("image"); img != "" {
				info.Image = img
			}
		}
	}

	switch val := v.(type) {
	case map[string]interface{}:
		process(val)
	case []interface{}:
		for _, item := range val {
			if m, ok := item.(map[string]interface{}); ok {
				process(m)
			}
		}
	}
}
