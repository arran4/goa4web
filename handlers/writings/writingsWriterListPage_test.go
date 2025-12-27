package writings

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

func TestWriterListPage_List(t *testing.T) {
	t.Helper()
	q := &db.QuerierStub{
		ListWritersForListerReturns: []*db.ListWritersForListerRow{
			{Username: sql.NullString{String: "alice", Valid: true}, Count: 3},
			{Username: sql.NullString{String: "bob", Valid: true}, Count: 2},
			{Username: sql.NullString{String: "carol", Valid: true}, Count: 1},
		},
	}

	req := httptest.NewRequest("GET", "/writings/writers", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.Config.PageSizeDefault = 2
	cd.Config.PageSizeMin = 1
	cd.Config.PageSizeMax = 10
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	WriterListPage(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if len(q.ListWritersForListerCalls) != 1 {
		t.Fatalf("calls=%d", len(q.ListWritersForListerCalls))
	}
	call := q.ListWritersForListerCalls[0]
	if call.Limit != int32(cd.PageSize()+1) || call.Offset != 0 {
		t.Fatalf("unexpected args %#v", call)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "alice") || !strings.Contains(body, "bob") {
		t.Fatalf("expected writers in body: %s", body)
	}
	if strings.Contains(body, "carol") {
		t.Fatalf("expected page to cap results: %s", body)
	}
	if cd.NextLink != "/writings/writers?offset=2" {
		t.Fatalf("next link=%s", cd.NextLink)
	}
	if cd.PrevLink != "" {
		t.Fatalf("prev link=%s", cd.PrevLink)
	}
}

func TestWriterListPage_Search(t *testing.T) {
	t.Helper()
	q := &db.QuerierStub{
		ListWritersSearchForListerReturns: []*db.ListWritersSearchForListerRow{
			{Username: sql.NullString{String: "bob", Valid: true}, Count: 2},
			{Username: sql.NullString{String: "bobby", Valid: true}, Count: 1},
			{Username: sql.NullString{String: "robert", Valid: true}, Count: 4},
		},
	}

	req := httptest.NewRequest("GET", "/writings/writers?search=bob&offset=2", nil)
	ctx := req.Context()
	cd := common.NewCoreData(ctx, q, config.NewRuntimeConfig())
	cd.Config.PageSizeDefault = 2
	cd.Config.PageSizeMin = 1
	cd.Config.PageSizeMax = 10
	cd.UserID = 99
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	WriterListPage(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("status=%d", rr.Result().StatusCode)
	}
	if len(q.ListWritersSearchForListerCalls) != 1 {
		t.Fatalf("search calls=%d", len(q.ListWritersSearchForListerCalls))
	}
	call := q.ListWritersSearchForListerCalls[0]
	if call.Query != "%bob%" || call.Offset != 2 || call.Limit != int32(cd.PageSize()+1) {
		t.Fatalf("unexpected search args %#v", call)
	}
	if len(q.ListWritersForListerCalls) != 0 {
		t.Fatalf("unexpected list calls=%d", len(q.ListWritersForListerCalls))
	}
	body := rr.Body.String()
	if !strings.Contains(body, "bob") || !strings.Contains(body, "bobby") {
		t.Fatalf("expected search writers in body: %s", body)
	}
	if strings.Contains(body, "robert") {
		t.Fatalf("expected search results capped to page size: %s", body)
	}
	if cd.NextLink != "/writings/writers?search=bob&offset=4" {
		t.Fatalf("next link=%s", cd.NextLink)
	}
	if cd.PrevLink != "/writings/writers?search=bob&offset=0" {
		t.Fatalf("prev link=%s", cd.PrevLink)
	}
}
