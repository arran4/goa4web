package opengraph

import (
	"time"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

			info, err := Fetch(server.URL, http.DefaultClient)
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

func TestNewSafeClient(t *testing.T) {
	client := NewSafeClient()
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.Timeout != 2*time.Second {
		t.Errorf("Expected Timeout of 2s, got %v", client.Timeout)
	}
	if client.CheckRedirect == nil {
		t.Error("Expected CheckRedirect to be set")
	}
}
