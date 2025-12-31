package auth

import (
	"os"
	"testing"

	"github.com/arran4/goa4web/core/templates"
)

func TestMain(m *testing.M) {
	templates.SetDir("../../core/templates")
	os.Exit(m.Run())
}
