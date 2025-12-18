package forum

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

type mockQuerierAPI struct {
	db.Querier
	GetCommentByIDFunc func(ctx context.Context, id int32) (*db.GetCommentByIdForUserRow, error)
}

func (m *mockQuerierAPI) GetCommentByID(ctx context.Context, id int32) (*db.GetCommentByIdForUserRow, error) {
	if m.GetCommentByIDFunc != nil {
		return m.GetCommentByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockQuerierAPI) GetCommentByIdForUser(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
	if m.GetCommentByIDFunc != nil {
		return m.GetCommentByIDFunc(ctx, arg.ID)
	}
	return nil, nil
}

func TestQuoteApi(t *testing.T) {
	tests := []struct {
		name           string
		commentID      string
		replyType      string
		mockComment    *db.GetCommentByIdForUserRow
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "Full Quote",
			commentID: "1",
			replyType: "full",
			mockComment: &db.GetCommentByIdForUserRow{
				Username: sql.NullString{String: "testuser", Valid: true},
				Text:     sql.NullString{String: "hello world", Valid: true},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"text":"[quoteof \"testuser\" hello world]\n"}`,
		},
		{
			name:      "Paragraph Quote",
			commentID: "1",
			replyType: "paragraph",
			mockComment: &db.GetCommentByIdForUserRow{
				Username: sql.NullString{String: "testuser", Valid: true},
				Text:     sql.NullString{String: "hello\n\nworld", Valid: true},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"text":"[quoteof \"testuser\" hello]\n[quoteof \"testuser\" world]\n"}`,
		},
		{
			name:           "Invalid Comment ID",
			commentID:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid comment ID\n",
		},
		{
			name:           "Comment Not Found",
			commentID:      "2",
			mockComment:    nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"text":""}`,
		},
		{
			name:      "Selected Text",
			commentID: "1",
			replyType: "selected&start=2&end=8",
			mockComment: &db.GetCommentByIdForUserRow{
				Text: sql.NullString{String: "hello [b]world[/b]", Valid: true},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"text":"[quoteof \"\" llo [b]wo[/b]]\n"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/forum/quote/"+tt.commentID+"?type="+tt.replyType, nil)
			if err != nil {
				t.Fatal(err)
			}

			cd := common.NewCoreData(context.Background(), &mockQuerierAPI{
				GetCommentByIDFunc: func(ctx context.Context, id int32) (*db.GetCommentByIdForUserRow, error) {
					if tt.mockComment != nil && id == 1 {
						return tt.mockComment, nil
					}
					return nil, nil
				},
			}, nil)
			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/api/forum/quote/{commentid}", QuoteApi)
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
