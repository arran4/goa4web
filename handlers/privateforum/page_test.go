package privateforum

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

func TestPage_NoAccess(t *testing.T) {
	cd := common.NewCoreData(context.Background(), nil, config.NewRuntimeConfig())
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	if body := w.Body.String(); !strings.Contains(body, "Forbidden") {
		t.Fatalf("expected no access message, got %q", body)
	}
}

func TestPage_Access(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()

	queries := db.New(conn)
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig(), common.WithUserRoles([]string{"administrator"}))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
	mock.ExpectQuery("SELECT 1 FROM grants").WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "Private Topics") {
		t.Fatalf("expected private topics page, got %q", body)
	}
	if !strings.Contains(body, "<form id=\"private-form\"") {
		t.Fatalf("expected create form, got %q", body)
	}
}

func TestPage_SeeNoCreate(t *testing.T) {
	callCount := 0
	mockQueries := &db.QuerierProxier{
		OverwrittenSystemCheckGrant: func(ctx context.Context, arg db.SystemCheckGrantParams) (int32, error) {
			callCount++
			if callCount == 1 {
				// First call, permission granted
				return 1, nil
			}
			// Second call, permission denied
			return 0, sql.ErrNoRows
		},
		OverwrittenGetPermissionsByUserID: func(ctx context.Context, usersIdusers int32) ([]*db.GetPermissionsByUserIDRow, error) {
			return []*db.GetPermissionsByUserIDRow{}, nil
		},
		OverwrittenSystemCheckRoleGrant: func(ctx context.Context, arg db.SystemCheckRoleGrantParams) (int32, error) {
			return 1, nil
		},
		OverwrittenSystemGetUserByID: func(ctx context.Context, idusers int32) (*db.SystemGetUserByIDRow, error) {
			return &db.SystemGetUserByIDRow{Username: sql.NullString{String: "testuser", Valid: true}}, nil
		},
		OverwrittenListContentPublicLabels: func(ctx context.Context, arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
			return []*db.ListContentPublicLabelsRow{}, nil
		},
		OverwrittenListContentPrivateLabels: func(ctx context.Context, arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
			return []*db.ListContentPrivateLabelsRow{}, nil
		},
		OverwrittenListPrivateTopicParticipantsByTopicIDForUser: func(ctx context.Context, arg db.ListPrivateTopicParticipantsByTopicIDForUserParams) ([]*db.ListPrivateTopicParticipantsByTopicIDForUserRow, error) {
			return []*db.ListPrivateTopicParticipantsByTopicIDForUserRow{}, nil
		},
		OverwrittenListPrivateTopicsByUserID: func(ctx context.Context, userID sql.NullInt32) ([]*db.ListPrivateTopicsByUserIDRow, error) {
			return []*db.ListPrivateTopicsByUserIDRow{}, nil
		},
		OverwrittenGetPreferenceForLister: func(ctx context.Context, listerID int32) (*db.Preference, error) {
			return &db.Preference{Timezone: sql.NullString{String: "UTC", Valid: true}}, nil
		},
	}
	cd := common.NewCoreData(context.Background(), mockQueries, config.NewRuntimeConfig())
	cd.UserID = 1

	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))

	w := httptest.NewRecorder()
	PrivateForumPage(w, req)

	body := w.Body.String()
	if strings.Contains(body, "Start conversation") {
		t.Fatalf("unexpected create form, got %q", body)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls to HasGrant, got %d", callCount)
	}
}
