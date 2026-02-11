package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
)

type dismissRequestQueries struct {
	db.Querier
	requestID     int32
	request       *db.AdminRequestQueue
	updatedStatus string
	comments      []string
}

func (q *dismissRequestQueries) AdminGetRequestByID(_ context.Context, id int32) (*db.AdminRequestQueue, error) {
	if id != q.requestID {
		return nil, fmt.Errorf("unexpected request id: %d", id)
	}
	return q.request, nil
}

func (q *dismissRequestQueries) AdminUpdateRequestStatus(ctx context.Context, arg db.AdminUpdateRequestStatusParams) error {
	if arg.ID != q.requestID {
		return fmt.Errorf("unexpected request id in update: %d", arg.ID)
	}
	q.updatedStatus = arg.Status
	return nil
}

func (q *dismissRequestQueries) AdminInsertRequestComment(ctx context.Context, arg db.AdminInsertRequestCommentParams) error {
	if arg.RequestID != q.requestID {
		return fmt.Errorf("unexpected request id in comment: %d", arg.RequestID)
	}
	q.comments = append(q.comments, arg.Comment)
	return nil
}

func (q *dismissRequestQueries) InsertAdminUserComment(ctx context.Context, arg db.InsertAdminUserCommentParams) error {
	return nil
}

func (q *dismissRequestQueries) SystemGetUserByID(ctx context.Context, id int32) (*db.SystemGetUserByIDRow, error) {
	return &db.SystemGetUserByIDRow{
		Idusers:  id,
		Username: sql.NullString{String: "admin", Valid: true},
	}, nil
}

func (q *dismissRequestQueries) GetPermissionsByUserID(ctx context.Context, id int32) ([]*db.GetPermissionsByUserIDRow, error) {
	return []*db.GetPermissionsByUserIDRow{
		{
			Name:    "administrator",
			IsAdmin: true,
		},
	}, nil
}

func TestHappyPathDismissRequestTask_Action(t *testing.T) {
	requestID := 10
	queries := &dismissRequestQueries{
		requestID: int32(requestID),
		request: &db.AdminRequestQueue{
			ID:           int32(requestID),
			UsersIdusers: 7,
			Status:       "pending",
			CreatedAt:    time.Now(),
		},
	}

	// Setup request
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/admin/request/%d/dismiss", requestID), nil)
	req = mux.SetURLVars(req, map[string]string{"request": strconv.Itoa(requestID)})
	req.PostForm = make(map[string][]string)
	req.PostForm.Set("task", "Dismiss")
	req.PostForm.Set("comment", "Dismissing as spam")

	// Setup context with CoreData and Admin permissions
	cfg := config.NewRuntimeConfig()
	cd := common.NewCoreData(req.Context(), queries, cfg)
	cd.AdminMode = true
	cd.LoadSelectionsFromRequest(req)

	// Preload permissions to simulate admin user
	cd.UserID = 1

	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	// Execute task
	w := httptest.NewRecorder()

	result := dismissRequestTask.Action(w, req)

	// Verify result is nil (as Action returns nil in our implementation)
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}

	// Verify database updates
	if queries.updatedStatus != "dismissed" {
		t.Errorf("expected status 'dismissed', got '%s'", queries.updatedStatus)
	}

	expectedComment := "status changed to dismissed"
	foundAutoComment := false
	foundUserComment := false
	for _, c := range queries.comments {
		if c == expectedComment {
			foundAutoComment = true
		}
		if c == "Dismissing as spam" {
			foundUserComment = true
		}
	}

	if !foundAutoComment {
		t.Errorf("missing auto comment '%s'", expectedComment)
	}
	if !foundUserComment {
		t.Errorf("missing user comment 'Dismissing as spam'")
	}
}
