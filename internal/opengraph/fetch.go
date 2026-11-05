package opengraph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// IPResolver resolves host names for SSRF validation and cancellable dialing.
type IPResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

// Info contains extracted metadata.
type Info struct {
	Title       string
	Description string
	Image       string
	Duration    string
	UploadDate  string
	Author      string
	Keywords    string
}

// NewSafeClient returns an http.Client configured to block internal IP addresses.
func NewSafeClient() *http.Client {
	return newSafeClient(net.DefaultResolver)
}

func newSafeClient(resolver IPResolver) *http.Client {
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	dialer := &net.Dialer{}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		ips, err := lookupPublicIPs(ctx, resolver, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, ip := range ips {
			conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if err == nil {
				return conn, nil
			}
			lastErr = err
		}
		return nil, lastErr
	}
	return &http.Client{
		Timeout:   2 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			if err := validateFetchURL(req.Context(), req.URL, resolver); err != nil {
				return fmt.Errorf("redirect: %w", err)
			}
			return nil
		},
	}
}

// Fetch retrieves and parses Open Graph metadata. All network work is tied to ctx.
func Fetch(ctx context.Context, urlStr string, client *http.Client) (*Info, error) {
	return fetch(ctx, urlStr, client, net.DefaultResolver)
}

func fetch(ctx context.Context, urlStr string, client *http.Client, resolver IPResolver) (*Info, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if client == nil {
		u, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		if err := validateFetchURL(ctx, u, resolver); err != nil {
			return nil, err
		}
		client = newSafeClient(resolver)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "goa4web/1.0 (+https://github.com/arran4/goa4web)")
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	return Parse(io.LimitReader(resp.Body, 5*1024*1024))
}

func validateFetchURL(ctx context.Context, u *url.URL, resolver IPResolver) error {
	if u == nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("unsupported URL scheme")
	}
	if u.Hostname() == "" {
		return fmt.Errorf("URL host required")
	}
	_, err := lookupPublicIPs(ctx, resolver, u.Hostname())
	return err
}

func lookupPublicIPs(ctx context.Context, resolver IPResolver, host string) ([]net.IPAddr, error) {
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("no addresses for host %q", host)
	}
	for _, addr := range ips {
		ip := addr.IP
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return nil, fmt.Errorf("blocked internal ip: %s", ip)
		}
	}
	return ips, nil
}

func Parse(r io.Reader) (*Info, error) {
	doc, err := html.Parse(r)
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

				var scriptBody strings.Builder
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.TextNode {
						scriptBody.WriteString(c.Data)
					}
				}
				if scriptStr := scriptBody.String(); scriptStr != "" {
					if jsonText, ok := extractAssignedJSONObject(scriptStr, "ytInitialPlayerResponse"); ok {
						var pr struct {
							VideoDetails struct {
								LengthSeconds string `json:"lengthSeconds"`
								Title         string `json:"title"`
							} `json:"videoDetails"`
						}
						if err := json.Unmarshal([]byte(jsonText), &pr); err == nil {
							if pr.VideoDetails.LengthSeconds != "" && info.Duration == "" {
								info.Duration = pr.VideoDetails.LengthSeconds
							}
							if pr.VideoDetails.Title != "" && info.Title == "" {
								info.Title = pr.VideoDetails.Title
							}
						}
					}
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
				case "og:title", "twitter:title":
					if info.Title == "" {
						info.Title = content
					}
				case "og:description", "twitter:description":
					if info.Description == "" {
						info.Description = content
					}
				case "og:image", "twitter:image":
					if info.Image == "" {
						info.Image = content
					}
				}

				if name == "twitter:title" && info.Title == "" {
					info.Title = content
				}
				if name == "twitter:description" && info.Description == "" {
					info.Description = content
				}
				if name == "twitter:image" && info.Image == "" {
					info.Image = content
				}

				if info.Title == "" && (name == "title" || itemprop == "title") {
					info.Title = content
				}
				if info.Description == "" && (name == "description" || itemprop == "description") {
					info.Description = content
				}
				if info.Image == "" && (name == "image" || itemprop == "image") {
					info.Image = content
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

				if name == "keywords" || itemprop == "keywords" || prop == "article:tag" {
					parts := strings.SplitSeq(content, ",")
					for p := range parts {
						p = strings.TrimSpace(p)
						if p == "" {
							continue
						}
						if info.Keywords == "" {
							info.Keywords = p
						} else {
							existing := strings.Split(info.Keywords, ", ")
							found := slices.Contains(existing, p)
							if !found {
								info.Keywords += ", " + p
							}
						}
					}
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
	var v any
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return
	}

	process := func(obj map[string]any) {
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
			if m, ok := val.(map[string]any); ok {
				if name, ok := m["name"].(string); ok {
					return name
				}
			}
			if s, ok := val.([]any); ok && len(s) > 0 {
				if m, ok := s[0].(map[string]any); ok {
					if name, ok := m["name"].(string); ok {
						return name
					}
				}
			}
			return ""
		}

		// Broaden JSON-LD parsing to general types, as they can also hold generic info (WebPage, PodcastEpisode, Article, etc.)
		if typeVal == "VideoObject" || strings.EqualFold(typeVal, "VideoObject") ||
			typeVal == "WebPage" || strings.EqualFold(typeVal, "WebPage") ||
			typeVal == "PodcastEpisode" || strings.EqualFold(typeVal, "PodcastEpisode") ||
			typeVal == "Article" || strings.EqualFold(typeVal, "Article") ||
			typeVal == "NewsArticle" || strings.EqualFold(typeVal, "NewsArticle") || typeVal != "" {
			// We prioritize JSON-LD over meta tags, so overwrite or only set if empty?
			// The request says "Prioritize JSON-LD". So we should overwrite if we found it here.
			// However, parseJSONLD is called during traversal.

			if t := getString("name"); t != "" && (typeVal == "VideoObject" || info.Title == "") {
				info.Title = t
			}
			if t := getString("headline"); t != "" && info.Title == "" {
				info.Title = t
			}
			if d := getString("description"); d != "" && (typeVal == "VideoObject" || info.Description == "") {
				info.Description = d
			}
			if dur := getString("duration"); dur != "" && (typeVal == "VideoObject" || info.Duration == "") {
				info.Duration = dur
			}
			if ud := getString("uploadDate"); ud != "" && (typeVal == "VideoObject" || info.UploadDate == "") {
				info.UploadDate = ud
			}
			if ud := getString("datePublished"); ud != "" && info.UploadDate == "" {
				info.UploadDate = ud
			}
			if auth := getAuthor(); auth != "" && (typeVal == "VideoObject" || info.Author == "") {
				info.Author = auth
			}

			if img := getString("thumbnailUrl"); img != "" && (typeVal == "VideoObject" || info.Image == "") {
				info.Image = img
			} else if img := getString("image"); img != "" && (typeVal == "VideoObject" || info.Image == "") {
				info.Image = img
			}

			if kw := getString("keywords"); kw != "" {
				parts := strings.SplitSeq(kw, ",")
				for p := range parts {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}
					if info.Keywords == "" {
						info.Keywords = p
					} else {
						existing := strings.Split(info.Keywords, ", ")
						found := slices.Contains(existing, p)
						if !found {
							info.Keywords += ", " + p
						}
					}
				}
			}
		}
	}

	switch val := v.(type) {
	case map[string]any:
		process(val)
	case []any:
		for _, item := range val {
			if m, ok := item.(map[string]any); ok {
				process(m)
			}
		}
	}
}

func extractAssignedJSONObject(script string, variableName string) (string, bool) {
	pos := strings.Index(script, variableName)
	if pos < 0 {
		return "", false
	}

	eq := strings.IndexByte(script[pos:], '=')
	if eq < 0 {
		return "", false
	}

	startSearch := pos + eq + 1

	start := strings.IndexByte(script[startSearch:], '{')
	if start < 0 {
		return "", false
	}

	start += startSearch

	end, ok := findMatchingJSONBrace(script, start)
	if !ok {
		return "", false
	}

	return script[start : end+1], true
}

func findMatchingJSONBrace(s string, start int) (int, bool) {
	if start < 0 || start >= len(s) || s[start] != '{' {
		return 0, false
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		ch := s[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}

			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}

			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}

	return 0, false
}
