package opengraph

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type resolverFunc func(context.Context, string) ([]net.IPAddr, error)

func (f resolverFunc) LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error) {
	return f(ctx, host)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

type contextBody struct {
	ctx    context.Context
	closed atomic.Bool
}

func (b *contextBody) Read([]byte) (int, error) {
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}

func (b *contextBody) Close() error {
	b.closed.Store(true)
	return nil
}

func TestFetch(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantInfo Info
	}{
		{
			name: "Keywords meta",
			html: `<html>
				<head>
					<meta name="keywords" content="podcast, science">
					<meta property="article:tag" content="health">
					<meta itemprop="keywords" content="health, autism">
				</head>
			</html>`,
			wantInfo: Info{
				Keywords: "podcast, science, health, autism",
			},
		},
		{
			name: "JSON-LD VideoObject",
			html: `<html>
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "VideoObject",
					"name": "JSON Title",
					"description": "JSON Desc",
					"thumbnailUrl": "JSON Image",
					"duration": "PT2M35S",
					"uploadDate": "2026-01-09",
					"author": "JSON Author"
				}
				</script>
			</html>`,
			wantInfo: Info{
				Title:       "JSON Title",
				Description: "JSON Desc",
				Image:       "JSON Image",
				Duration:    "PT2M35S",
				UploadDate:  "2026-01-09",
				Author:      "JSON Author",
			},
		},
		{
			name: "Meta Tags Fallback",
			html: `<html>
				<meta property="og:title" content="Meta Title">
				<meta property="og:description" content="Meta Desc">
				<meta property="og:image" content="Meta Image">
				<meta itemprop="duration" content="PT1M">
				<meta itemprop="uploadDate" content="2025-01-01">
				<meta itemprop="author" content="Meta Author">
			</html>`,
			wantInfo: Info{
				Title:       "Meta Title",
				Description: "Meta Desc",
				Image:       "Meta Image",
				Duration:    "PT1M",
				UploadDate:  "2025-01-01",
				Author:      "Meta Author",
			},
		},
		{
			name: "Prioritize JSON-LD",
			html: `<html>
				<meta property="og:title" content="Meta Title">
				<script type="application/ld+json">
				{
					"@context": "https://schema.org",
					"@type": "VideoObject",
					"name": "JSON Title",
                    "duration": "JSON Duration"
				}
				</script>
                <meta itemprop="duration" content="Meta Duration">
			</html>`,
			wantInfo: Info{
				Title:       "JSON Title",
				Description: "",
				Image:       "",
				Duration:    "JSON Duration",
				UploadDate:  "",
				Author:      "",
			},
		},
		{
			name: "Author Object in JSON-LD",
			html: `<html>
                <script type="application/ld+json">
                {
                    "@context": "https://schema.org",
                    "@type": "VideoObject",
                    "author": {
                        "@type": "Person",
                        "name": "Object Author"
                    }
                }
                </script>
            </html>`,
			wantInfo: Info{
				Author: "Object Author",
			},
		},
		{
			name: "Author Array in JSON-LD",
			html: `<html>
                <script type="application/ld+json">
                {
                    "@context": "https://schema.org",
                    "@type": "VideoObject",
                    "author": [
                        {
                            "@type": "Person",
                            "name": "Array Author"
                        }
                    ]
                }
                </script>
            </html>`,
			wantInfo: Info{
				Author: "Array Author",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(tt.html))
			}))
			defer server.Close()

			info, err := Fetch(context.Background(), server.URL, http.DefaultClient)
			if err != nil {
				t.Fatalf("Fetch() error = %v", err)
			}

			if info.Title != tt.wantInfo.Title {
				t.Errorf("Title = %v, want %v", info.Title, tt.wantInfo.Title)
			}
			if info.Description != tt.wantInfo.Description {
				t.Errorf("Description = %v, want %v", info.Description, tt.wantInfo.Description)
			}
			if info.Image != tt.wantInfo.Image {
				t.Errorf("Image = %v, want %v", info.Image, tt.wantInfo.Image)
			}
			if info.Duration != tt.wantInfo.Duration {
				t.Errorf("Duration = %v, want %v", info.Duration, tt.wantInfo.Duration)
			}
			if info.UploadDate != tt.wantInfo.UploadDate {
				t.Errorf("UploadDate = %v, want %v", info.UploadDate, tt.wantInfo.UploadDate)
			}
			if info.Author != tt.wantInfo.Author {
				t.Errorf("Author = %v, want %v", info.Author, tt.wantInfo.Author)
			}
			if info.Keywords != tt.wantInfo.Keywords {
				t.Errorf("Keywords = %v, want %v", info.Keywords, tt.wantInfo.Keywords)
			}
		})
	}
}

func TestFetchCancelledDuringHTTPRequestClosesBody(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	body := &contextBody{ctx: ctx}
	started := make(chan struct{})
	client := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		close(started)
		return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: body}, nil
	})}
	done := make(chan error, 1)
	go func() {
		_, err := Fetch(ctx, "https://example.test", client)
		done <- err
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Fetch() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Fetch did not stop after request cancellation")
	}
	if !body.closed.Load() {
		t.Fatal("response body was not closed")
	}
}

func TestFetchCancellationDuringDNSResolution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	resolver := resolverFunc(func(ctx context.Context, _ string) ([]net.IPAddr, error) {
		close(started)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	done := make(chan error, 1)
	go func() {
		_, err := fetch(ctx, "https://example.test", nil, resolver)
		done <- err
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("fetch() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("fetch did not stop during DNS resolution")
	}
}

func TestFetchCancellationDuringRedirect(t *testing.T) {
	redirected := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/next", http.StatusFound)
	}))
	defer server.Close()
	resolver := resolverFunc(func(ctx context.Context, _ string) ([]net.IPAddr, error) {
		close(redirected)
		<-ctx.Done()
		return nil, ctx.Err()
	})
	client := newSafeClient(resolver)
	client.Transport = server.Client().Transport
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := Fetch(ctx, server.URL, client)
		done <- err
	}()
	<-redirected
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Fetch() error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Fetch did not stop during redirect")
	}
}

func TestFetchCancelledBeforeRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
	_, err := Fetch(ctx, "https://example.test", client)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fetch() error = %v, want context.Canceled", err)
	}
	if called {
		t.Fatal("transport called after context was already cancelled")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantInfo Info
		wantErr  bool
	}{
		{
			name: "Complete og: tags",
			html: `<html>
				<meta property="og:title" content="og Title">
				<meta property="og:description" content="og Desc">
				<meta property="og:image" content="og Image">
			</html>`,
			wantInfo: Info{
				Title:       "og Title",
				Description: "og Desc",
				Image:       "og Image",
			},
		},
		{
			name: "Meta name fallbacks",
			html: `<html>
				<meta name="title" content="name Title">
				<meta name="description" content="name Desc">
				<meta name="uploadDate" content="2025-01-01">
				<meta name="author" content="name Author">
			</html>`,
			wantInfo: Info{
				Title:       "name Title",
				Description: "name Desc",
				UploadDate:  "2025-01-01",
				Author:      "name Author",
			},
		},
		{
			name: "Meta itemprop fallbacks",
			html: `<html>
				<meta itemprop="duration" content="PT1M">
				<meta itemprop="uploadDate" content="2025-01-02">
				<meta itemprop="author" content="itemprop Author">
			</html>`,
			wantInfo: Info{
				Duration:   "PT1M",
				UploadDate: "2025-01-02",
				Author:     "itemprop Author",
			},
		},
		{
			name: "Title tag fallback",
			html: `<html>
				<head><title>Tag Title</title></head>
			</html>`,
			wantInfo: Info{
				Title: "Tag Title",
			},
		},
		{
			name: "JSON-LD VideoObject",
			html: `<html>
				<script type="application/ld+json">
				{
					"@type": "VideoObject",
					"name": "JSON Title",
					"description": "JSON Desc",
					"thumbnailUrl": "JSON Image",
					"duration": "PT2M35S",
					"uploadDate": "2026-01-09",
					"author": "JSON Author"
				}
				</script>
			</html>`,
			wantInfo: Info{
				Title:       "JSON Title",
				Description: "JSON Desc",
				Image:       "JSON Image",
				Duration:    "PT2M35S",
				UploadDate:  "2026-01-09",
				Author:      "JSON Author",
			},
		},
		{
			name: "JSON-LD VideoObject in array",
			html: `<html>
				<script type="application/ld+json">
				[
					{
						"@type": "VideoObject",
						"name": "JSON Array Title"
					}
				]
				</script>
			</html>`,
			wantInfo: Info{
				Title: "JSON Array Title",
			},
		},
		{
			name: "JSON-LD author as object",
			html: `<html>
				<script type="application/ld+json">
				{
					"@type": "VideoObject",
					"author": {
						"name": "Object Author"
					}
				}
				</script>
			</html>`,
			wantInfo: Info{
				Author: "Object Author",
			},
		},
		{
			name: "JSON-LD author as array",
			html: `<html>
				<script type="application/ld+json">
				{
					"@type": "VideoObject",
					"author": [
						{
							"name": "Array Author"
						}
					]
				}
				</script>
			</html>`,
			wantInfo: Info{
				Author: "Array Author",
			},
		},
		{
			name: "Invalid JSON-LD ignored",
			html: `<html>
				<script type="application/ld+json">
				{
					invalid json
				}
				</script>
				<meta property="og:title" content="og Title">
			</html>`,
			wantInfo: Info{
				Title: "og Title",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := Parse(strings.NewReader(tt.html))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if info.Title != tt.wantInfo.Title {
					t.Errorf("Title = %v, want %v", info.Title, tt.wantInfo.Title)
				}
				if info.Description != tt.wantInfo.Description {
					t.Errorf("Description = %v, want %v", info.Description, tt.wantInfo.Description)
				}
				if info.Image != tt.wantInfo.Image {
					t.Errorf("Image = %v, want %v", info.Image, tt.wantInfo.Image)
				}
				if info.Duration != tt.wantInfo.Duration {
					t.Errorf("Duration = %v, want %v", info.Duration, tt.wantInfo.Duration)
				}
				if info.UploadDate != tt.wantInfo.UploadDate {
					t.Errorf("UploadDate = %v, want %v", info.UploadDate, tt.wantInfo.UploadDate)
				}
				if info.Author != tt.wantInfo.Author {
					t.Errorf("Author = %v, want %v", info.Author, tt.wantInfo.Author)
				}
				if info.Keywords != tt.wantInfo.Keywords {
					t.Errorf("Keywords = %v, want %v", info.Keywords, tt.wantInfo.Keywords)
				}
			}
		})
	}
}
