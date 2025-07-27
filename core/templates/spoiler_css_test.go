package templates_test

import (
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/templates"
)

func TestSpoilerCSS(t *testing.T) {
	css := string(templates.GetMainCSSData())
	if !strings.Contains(css, ".spoiler:hover") {
		t.Errorf("spoiler CSS rule missing")
	}
}
