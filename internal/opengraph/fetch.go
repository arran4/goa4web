package opengraph

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

func Fetch(urlStr string, client *http.Client) (title, desc, image string, err error) {
	if client == nil {
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

		client = &http.Client{
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
