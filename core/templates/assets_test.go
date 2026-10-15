package templates_test

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/templates"
)

func TestExternalLinkCardThumbnailStyles(t *testing.T) {
	css, err := templates.Asset("main.css")
	if err != nil {
		t.Fatalf("read embedded stylesheet: %v", err)
	}
	styles := string(css)
	for _, want := range []string{
		".external-link-image-container {",
		"width: 150px;",
		"height: 150px;",
		"padding: 0.5em;",
		"justify-content: center;",
		"object-fit: contain;",
	} {
		if !strings.Contains(styles, want) {
			t.Errorf("embedded stylesheet missing %q", want)
		}
	}
}
