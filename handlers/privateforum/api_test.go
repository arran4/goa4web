package privateforum

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type QuerierProxier struct {
	db.Querier
	SystemGetUserByUsernameFunc func(ctx context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error)
}

func (p *QuerierProxier) SystemGetUserByUsername(ctx context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
	return p.SystemGetUserByUsernameFunc(ctx, username)
}

func TestUserExistsAPI(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		mockFunc       func(username sql.NullString) (*db.SystemGetUserByUsernameRow, error)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing username",
			username:       "",
			mockFunc:       nil,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "username is required\n",
		},
		{
			name:     "user exists",
			username: "testuser",
			mockFunc: func(username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
				return &db.SystemGetUserByUsernameRow{Username: sql.NullString{String: "testuser", Valid: true}}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"exists":true}` + "\n",
		},
		{
			name:     "user does not exist",
			username: "nonexistent",
			mockFunc: func(username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
				return nil, sql.ErrNoRows
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"exists":false}` + "\n",
		},
		{
			name:     "database error",
			username: "testuser",
			mockFunc: func(username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "database error\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockQuerier := &QuerierProxier{}
			if tt.mockFunc != nil {
				mockQuerier.SystemGetUserByUsernameFunc = func(ctx context.Context, username sql.NullString) (*db.SystemGetUserByUsernameRow, error) {
					return tt.mockFunc(username)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/private/api/user-exists", strings.NewReader(url.Values{"username": {tt.username}}.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			cd := common.NewCoreData(context.Background(), mockQuerier, nil)
			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)

			UserExistsAPI(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}
