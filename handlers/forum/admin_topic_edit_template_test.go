package forum

import (
	"bytes"
	"context"
	"database/sql"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestAdminTopicEditTemplateDeleteTaskValue(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	req := httptest.NewRequest("GET", "/admin/forum/topics/topic/12/edit", nil)

	data := struct {
		Topic       *db.GetForumTopicByIdForUserRow
		Categories  []*db.Forumcategory
		Roles       []*db.Role
		Restriction any
	}{
		Topic: &db.GetForumTopicByIdForUserRow{
			Idforumtopic:                 12,
			Title:                        sql.NullString{String: "topic", Valid: true},
			Description:                  sql.NullString{String: "desc", Valid: true},
			ForumcategoryIdforumcategory: 1,
		},
		Categories: []*db.Forumcategory{{
			Idforumcategory: 1,
			Title:           sql.NullString{String: "category", Valid: true},
		}},
	}

	var out bytes.Buffer
	if err := cd.ExecuteSiteTemplate(&out, req, "forum/adminTopicEditPage.gohtml", data); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	body := out.String()
	if !strings.Contains(body, "value=\"Forum topic delete\"") {
		t.Fatalf("expected delete task value to match topic delete task, got: %s", body)
	}
	if strings.Contains(body, "value=\"Delete Topic\"") {
		t.Fatalf("unexpected legacy delete task value in template")
	}
}
