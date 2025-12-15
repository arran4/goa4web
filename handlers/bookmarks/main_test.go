package bookmarks

import (
	"testing"

	"github.com/arran4/goa4web/core/templates"
)

func TestMain(m *testing.M) {
	templates.SetDir("../../core/templates")
	m.Run()
}
