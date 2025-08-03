package templates_test

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

type commentForTest struct {
	Written            struct{ Time time.Time }
	Text               struct{ String string }
	Posterusername     struct{ String string }
	Idcomments         int32
	Writing            struct{ Idwriting int32 }
	EditCommentID      int32
	CanReply           bool
	Languages          []struct{}
	SelectedLanguageId int32
	IsOwner            bool
}

// Test that the comment template shows the edit form when Editing is true.
func TestCommentTemplateEditing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	tmpl := templates.GetCompiledSiteTemplates(cd.Funcs(r))

	c := commentForTest{}
	c.Written.Time = time.Now()
	c.Text.String = "hello"
	c.Posterusername.String = "user"
	c.Idcomments = 1
	c.Writing.Idwriting = 2
	c.CanReply = true
	c.IsOwner = true
	c.EditCommentID = c.Idcomments

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "comment", c); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, "Edit Reply") {
		t.Errorf("edit form not rendered when expected")
	}

	c.EditCommentID = 0
	buf.Reset()
	if err := tmpl.ExecuteTemplate(&buf, "comment", c); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(buf.String(), "Edit Reply") {
		t.Errorf("edit form rendered unexpectedly")
	}
}
