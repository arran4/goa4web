package writings

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
)

func TestWritingCategoryChangeTask(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.ListWritingCategoriesForListerReturns = []*db.WritingCategory{
		writingCategoryRow(1, 0, "a", ""),
	}
	queries.SystemListWritingCategoriesReturns = queries.ListWritingCategoriesForListerReturns
	queries.AdminUpdateWritingCategoryErr = nil

	form := url.Values{"name": {"A"}, "desc": {"B"}, "pcid": {"0"}, "cid": {"1"}}
	req := httptest.NewRequest("POST", "/admin/writings/categories/category/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if v := writingCategoryChangeTask.Action(nil, req); v != nil {
		t.Fatalf("action returned %v", v)
	}
}

func TestWritingCategoryWouldLoop(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.ListWritingCategoriesForListerReturns = []*db.WritingCategory{
		writingCategoryRow(1, 0, "a", ""),
		writingCategoryRow(2, 1, "b", ""),
	}
	queries.SystemListWritingCategoriesReturns = queries.ListWritingCategoriesForListerReturns

	_, loop, err := writingCategoryWouldLoop(context.Background(), queries, 1, 2)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
}

func TestWritingCategoryWouldLoopSelfRef(t *testing.T) {
	_, loop, err := writingCategoryWouldLoop(context.Background(), nil, 3, 3)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
}

func TestWritingCategoryWouldLoopHeadToTail(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.ListWritingCategoriesForListerReturns = []*db.WritingCategory{
		writingCategoryRow(1, 0, "a", ""),
		writingCategoryRow(2, 1, "b", ""),
		writingCategoryRow(3, 2, "c", ""),
		writingCategoryRow(4, 3, "d", ""),
	}
	queries.SystemListWritingCategoriesReturns = queries.ListWritingCategoriesForListerReturns

	_, loop, err := writingCategoryWouldLoop(context.Background(), queries, 1, 4)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
}

func TestWritingCategoryWouldLoopAfterNode(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.ListWritingCategoriesForListerReturns = []*db.WritingCategory{
		writingCategoryRow(1, 0, "a", ""),
		writingCategoryRow(2, 3, "b", ""),
		writingCategoryRow(3, 2, "c", ""),
	}
	queries.SystemListWritingCategoriesReturns = queries.ListWritingCategoriesForListerReturns

	_, loop, err := writingCategoryWouldLoop(context.Background(), queries, 1, 2)
	if err != nil {
		t.Fatalf("writingCategoryWouldLoop: %v", err)
	}
	if !loop {
		t.Fatalf("expected loop")
	}
}

func TestWritingCategoryChangeTaskLoop(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	queries.ListWritingCategoriesForListerReturns = []*db.WritingCategory{
		writingCategoryRow(1, 2, "a", ""),
		writingCategoryRow(2, 1, "b", ""),
	}
	queries.SystemListWritingCategoriesReturns = queries.ListWritingCategoriesForListerReturns

	form := url.Values{"name": {"A"}, "desc": {"B"}, "pcid": {"2"}, "cid": {"1"}}
	req := httptest.NewRequest("POST", "/admin/writings/categories/category/1/edit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cd := common.NewCoreData(req.Context(), queries, config.NewRuntimeConfig())
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	if v := writingCategoryChangeTask.Action(nil, req); v == nil {
		t.Fatalf("expected error")
	} else if ue, ok := v.(common.UserError); !ok {
		t.Fatalf("expected user error got %T", v)
	} else if !strings.HasPrefix(ue.UserErrorMessage(), "invalid parent category: loop") {
		t.Fatalf("unexpected error message %q", ue.UserErrorMessage())
	}
}

func writingCategoryRow(id int32, parentID int32, title string, description string) *db.WritingCategory {
	parent := sql.NullInt32{Valid: false}
	if parentID != 0 {
		parent = sql.NullInt32{Int32: parentID, Valid: true}
	}
	return &db.WritingCategory{
		Idwritingcategory: id,
		WritingCategoryID: parent,
		Title:             sql.NullString{String: title, Valid: title != ""},
		Description:       sql.NullString{String: description, Valid: description != ""},
	}
}
