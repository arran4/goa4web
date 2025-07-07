package templates

import "strings"
import "testing"

func TestSpoilerCSS(t *testing.T) {
	css := string(GetMainCSSData())
	if !strings.Contains(css, ".spoiler:hover") {
		t.Errorf("spoiler CSS rule missing")
	}
}
