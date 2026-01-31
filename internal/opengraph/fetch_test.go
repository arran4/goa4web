package opengraph

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		wantInfo Info
	}{
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
				w.Write([]byte(tt.html))
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
		})
	}
}
