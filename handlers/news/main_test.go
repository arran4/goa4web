package news

import (
	"github.com/arran4/goa4web/core/templates"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	templates.SetDir("../../core/templates")
	os.Exit(m.Run())
}
