package templates

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	corecommon "github.com/arran4/goa4web/core/common"
)

type commentForTest struct {
	Written            struct{ Time time.Time }
	Text               struct{ String string }
	Posterusername     struct{ String string }
	Idcomments         int32
	ShowReply          bool
	EditUrl            string
	EditSaveUrl        string
	Editing            bool
	Languages          []struct{}
	SelectedLanguageId int32
}

// Test that the comment template shows the edit form when Editing is true.
func TestCommentTemplateEditing(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	tmpl := GetCompiledTemplates(corecommon.NewFuncs(r))

	c := commentForTest{}
	c.Written.Time = time.Now()
	c.Text.String = "hello"
	c.Posterusername.String = "user"
	c.Idcomments = 1
	c.ShowReply = true
	c.EditUrl = "/edit"
	c.EditSaveUrl = "/save"
	c.Editing = true

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "comment", c); err != nil {
		t.Fatalf("render: %v", err)
	}
	html := buf.String()
	if !strings.Contains(html, "Edit Reply") {
		t.Errorf("edit form not rendered when expected")
	}

	c.Editing = false
	buf.Reset()
	if err := tmpl.ExecuteTemplate(&buf, "comment", c); err != nil {
		t.Fatalf("render: %v", err)
	}
	if strings.Contains(buf.String(), "Edit Reply") {
		t.Errorf("edit form rendered unexpectedly")
	}
}
