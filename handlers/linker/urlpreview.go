package linker

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// fetchPageTitle returns the <title> contents of the page at the given URL.
// An empty string is returned if the title cannot be retrieved within the
// timeout.
func fetchPageTitle(ctx context.Context, targetURL string) string {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return ""
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return ""
		case html.StartTagToken:
			t := tokenizer.Token()
			if t.Data == "title" {
				if tokenizer.Next() == html.TextToken {
					text := strings.TrimSpace(tokenizer.Token().Data)
					return text
				}
			}
		}
	}
}
