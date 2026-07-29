package opengraph

import (
	"strings"
	"testing"
)

func TestParse_Keywords(t *testing.T) {
	html := `<html><head>
		<meta name="keywords" content="podcast, science, health, autism">
	</head><body></body></html>`
	info, err := Parse(strings.NewReader(html))
	if err != nil {
		t.Fatalf("Parse() err = %v", err)
	}
	t.Logf("Info: %+v", info)
}
