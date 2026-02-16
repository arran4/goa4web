package common_test

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/core/consts"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/internal/db"
)

func TestFirstLine(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"Single line", "a", "a"},
		{"Multi line", "a\nb\n", "a"},
		{"Empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.FirstLine(tt.in); got != tt.want {
				t.Errorf("FirstLine(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		a, b int
		want int
	}{
		{"Positive numbers", 1, 2, 3},
		{"Negative numbers", -1, -2, -3},
		{"Mixed numbers", -1, 2, 1},
		{"Zero", 0, 0, 0},
		{"Zero and positive", 0, 5, 5},
		{"Zero and negative", 0, -5, -5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.Add(tt.a, tt.b); got != tt.want {
				t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestLeft(t *testing.T) {
	tests := []struct {
		name string
		i    int
		s    string
		want string
	}{
		{"Short string", 3, "hello", "hel"},
		{"Long length", 10, "hi", "hi"},
		{"Empty string", 5, "", ""},
		{"Zero length", 0, "hello", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.Left(tt.i, tt.s); got != tt.want {
				t.Errorf("Left(%d, %q) = %q, want %q", tt.i, tt.s, got, tt.want)
			}
		})
	}
}

func TestTruncateWords(t *testing.T) {
	tests := []struct {
		name string
		i    int
		s    string
		want string
	}{
		{"No truncation needed", 5, "one two three", "one two three"},
		{"Truncation needed", 2, "one two three", "one two..."},
		{"Exact length", 3, "one two three", "one two three"},
		{"Empty string", 5, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.TruncateWords(tt.i, tt.s); got != tt.want {
				t.Errorf("TruncateWords(%d, %q) = %q, want %q", tt.i, tt.s, got, tt.want)
			}
		})
	}
}

func TestToInt32(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want int32
	}{
		{"int", int(10), 10},
		{"int32", int32(20), 20},
		{"int64", int64(30), 30},
		{"string", "40", 40},
		{"string invalid", "abc", 0},
		{"float64 (unsupported default)", 50.5, 0},
		{"nil", nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := common.ToInt32(tt.in); got != tt.want {
				t.Errorf("ToInt32(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestSeq(t *testing.T) {
	tests := []struct {
		name  string
		start int
		end   int
		want  []int
	}{
		{"Ascending", 1, 3, []int{1, 2, 3}},
		{"Single", 5, 5, []int{5}},
		{"Descending (empty)", 3, 1, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := common.Seq(tt.start, tt.end)
			if len(got) != len(tt.want) {
				t.Errorf("Seq(%d, %d) len = %d, want %d", tt.start, tt.end, len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Seq(%d, %d)[%d] = %d, want %d", tt.start, tt.end, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTemplateFuncsCSRFToken(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{}
	funcs := cd.Funcs(r)
	if _, ok := funcs["csrfToken"]; !ok {
		t.Errorf("csrfToken func missing")
	}
	if _, ok := funcs["csrf"]; ok {
		t.Errorf("csrf func should not be present")
	}
}

func TestLatestNewsRespectsPermissions(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)

	now := time.Now()
	rows := sqlmock.NewRows([]string{
		"writerName", "writerId", "idsitenews", "forumthread_id", "language_id",
		"users_idusers", "news", "occurred", "timezone", "comments",
	}).AddRow("w", 1, 1, 0, 1, 1, "a", now, time.Local.String(), 0).AddRow("w", 1, 2, 0, 1, 1, "b", now, time.Local.String(), 0)

	mock.ExpectQuery("SELECT u.username").WithArgs(int32(1), int32(1), int32(1), sql.NullInt32{Int32: 1, Valid: true}, int32(15), int32(0)).WillReturnRows(rows)

	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 1, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	mock.ExpectQuery("SELECT 1 FROM grants").WithArgs(int32(1), "news", sql.NullString{String: "post", Valid: true}, "see", sql.NullInt32{Int32: 2, Valid: true}, sql.NullInt32{Int32: 1, Valid: true}).WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest("GET", "/", nil)
	ctx := req.Context()
	cd := common.NewTestCoreData(t, queries)
	common.WithUserRoles([]string{"user"})(cd)
	cd.UserID = 1
	ctx = context.WithValue(ctx, consts.KeyCoreData, cd)
	_ = req.WithContext(ctx)

	res, err := cd.LatestNews()
	if err != nil {
		t.Fatalf("LatestNews: %v", err)
	}
	if l := len(res); l != 1 {
		t.Fatalf("expected 1 news post, got %d", l)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAddmodeSkipsAdminLinks(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	cd := &common.CoreData{AdminMode: true}
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	funcs := cd.Funcs(req)
	addmode := funcs["addmode"].(func(string) string)

	tests := []struct {
		in   string
		want string
	}{
		{"/admin", "/admin"},
		{"/admin/tools", "/admin/tools"},
		{"/admin/tools?flag=1", "/admin/tools?flag=1"},
		{"http://example.com/admin", "http://example.com/admin"},
		{"/administrator", "/administrator?mode=admin"},
		{"/user", "/user?mode=admin"},
		{"/user?id=1", "/user?id=1&mode=admin"},
	}

	for _, tt := range tests {
		if got := addmode(tt.in); got != tt.want {
			t.Errorf("addmode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestTimeAgo(t *testing.T) {
	now := time.Now()
	tests := []struct {
		d    time.Duration
		want string
	}{
		{time.Second * 30, "post was 30 seconds ago"},
		{time.Second, "post was 1 second ago"},
		{time.Minute + time.Second, "post was 1 minute ago"},
		{time.Minute * 2, "post was 2 minutes ago"},
		{time.Hour + time.Minute, "post was 1 hour ago"},
		{time.Hour * 5, "post was 5 hours ago"},
		{time.Hour * 25, "post was 1 day ago"},
		{time.Hour * 49, "post was 2 days ago"},
	}

	for _, tt := range tests {
		// Use newly exported TimeAgo with explicit 'now'
		got := common.TimeAgo(now.Add(-tt.d), now)
		if got != tt.want {
			t.Errorf("TimeAgo(-%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
