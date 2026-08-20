package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestTopicLabelTasksPermissionSectionSelection(t *testing.T) {
	taskList := []struct {
		name string
		task tasks.Task
		form url.Values
	}{
		{
			name: "AddTopicPublicLabelTask",
			task: addTopicPublicLabelTask,
			form: url.Values{"label": []string{"Important"}},
		},
		{
			name: "RemoveTopicPublicLabelTask",
			task: removeTopicPublicLabelTask,
			form: url.Values{"label": []string{"Important"}},
		},
		{
			name: "SetTopicLabelsTask",
			task: setTopicLabelsTask,
			form: url.Values{
				"public":  []string{"Important"},
				"private": []string{"Personal"},
			},
		},
	}

	for _, tt := range taskList {
		t.Run(tt.name+" Public Forum uses forum section", func(t *testing.T) {
			topicID := int32(5)
			var checkedSection string

			q := testhelpers.NewQuerierStub()
			q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
				checkedSection = p.Section
				if p.Section == "forum" && p.Item.String == "topic" && p.Action == "label" && p.ItemID.Int32 == topicID {
					return 1, nil
				}
				return 0, sql.ErrNoRows
			}

			cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
			cd.UserID = 42
			cd.ForumBasePath = "/forum"

			req := httptest.NewRequest(http.MethodPost, "/forum/topic/5/labels", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = mux.SetURLVars(req, map[string]string{"topic": "5"})
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

			res := tt.task.Action(httptest.NewRecorder(), req)
			if checkedSection != "forum" {
				t.Fatalf("expected permission section 'forum', got %q", checkedSection)
			}
			if _, ok := res.(handlers.RefreshDirectHandler); !ok {
				t.Fatalf("expected RefreshDirectHandler outcome, got %v", res)
			}
		})

		t.Run(tt.name+" Private Forum uses privateforum section", func(t *testing.T) {
			topicID := int32(5)
			var checkedSection string

			q := testhelpers.NewQuerierStub()
			q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
				checkedSection = p.Section
				// User only has privateforum / topic / label grant, NOT forum / topic / label
				if p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "label" && p.ItemID.Int32 == topicID {
					return 1, nil
				}
				return 0, sql.ErrNoRows
			}

			cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
			cd.UserID = 42
			cd.ForumBasePath = "/private"

			req := httptest.NewRequest(http.MethodPost, "/private/topic/5/labels", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = mux.SetURLVars(req, map[string]string{"topic": "5"})
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

			res := tt.task.Action(httptest.NewRecorder(), req)
			if checkedSection != "privateforum" {
				t.Fatalf("expected permission section 'privateforum', got %q", checkedSection)
			}
			if _, ok := res.(handlers.RefreshDirectHandler); !ok {
				t.Fatalf("expected RefreshDirectHandler outcome, got %v", res)
			}
		})

		t.Run(tt.name+" User with only privateforum label grant is denied on public forum", func(t *testing.T) {
			topicID := int32(5)

			q := testhelpers.NewQuerierStub()
			q.SystemCheckGrantFn = func(p db.SystemCheckGrantParams) (int32, error) {
				if p.Section == "privateforum" && p.Item.String == "topic" && p.Action == "label" && p.ItemID.Int32 == topicID {
					return 1, nil
				}
				return 0, sql.ErrNoRows
			}

			cd := common.NewCoreData(context.Background(), q, config.NewRuntimeConfig())
			cd.UserID = 42
			cd.ForumBasePath = "/forum"

			req := httptest.NewRequest(http.MethodPost, "/forum/topic/5/labels", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req = mux.SetURLVars(req, map[string]string{"topic": "5"})
			req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

			res := tt.task.Action(httptest.NewRecorder(), req)
			err, isErr := res.(error)
			if !isErr || err == nil || !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("expected permission denied error, got %v", res)
			}
		})
	}
}
