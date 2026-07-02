package forum

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

func TestQuoteApi(t *testing.T) {
	type testCase struct {
		name           string
		commentID      string
		replyType      string
		mockComment    *db.GetCommentByIdForUserRow
		expectedStatus int
		expectedBody   string
	}

	happyTests := []testCase{
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
			expectedBody:   `{"text":"[quoteof \"testuser\" hello]\n\n\n\n[quoteof \"testuser\" world]\n"}`,
		},
		{
			name:      "Selected Text",
			commentID: "1",
			replyType: "selected&start=2&end=8",
			mockComment: &db.GetCommentByIdForUserRow{
				Text: sql.NullString{String: "hello [b world]", Valid: true},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"text":"[quoteof \"\" llo [b wo]]\n"}`,
		},
		{
			name:           "Comment Not Found", // Treated as happy path (200 OK)
			commentID:      "2",
			mockComment:    nil,
			expectedStatus: http.StatusOK,
			expectedBody:   `{"text":""}`,
		},
	}

	unhappyTests := []testCase{
		{
			name:           "Invalid Comment ID",
			commentID:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid comment ID\n",
		},
	}

	runTest := func(t *testing.T, tt testCase) {
		req, err := http.NewRequest("GET", "/api/forum/quote/"+tt.commentID+"?type="+tt.replyType, nil)
		if err != nil {
			t.Fatal(err)
		}

		q := testhelpers.NewQuerierStub()
		q.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
			if tt.mockComment != nil && arg.ID == 1 {
				return tt.mockComment, nil
			}
			return nil, nil
		}

		cd := common.NewCoreData(context.Background(), q, nil)
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
	}

	t.Run("Happy Path", func(t *testing.T) {
		for _, tt := range happyTests {
			t.Run(tt.name, func(t *testing.T) {
				runTest(t, tt)
			})
		}
	})

	t.Run("Unhappy Path", func(t *testing.T) {
		for _, tt := range unhappyTests {
			t.Run(tt.name, func(t *testing.T) {
				runTest(t, tt)
			})
		}
	})
}

func TestQuoteSelectionApiGroupsConsecutiveSelectionsByUser(t *testing.T) {
	body, err := json.Marshal(quoteSelectionRequest{
		Ranges: []quoteSelectionRange{
			{CommentID: 1, Start: 0, End: 5},
			{CommentID: 2, Start: 0, End: 1},
			{CommentID: 3, Start: 0, End: 4},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/forum/quote-selection", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	comments := map[int32]*db.GetCommentByIdForUserRow{
		1: {
			Username: sql.NullString{String: "arran", Valid: true},
			Text:     sql.NullString{String: "hello world", Valid: true},
		},
		2: {
			Username: sql.NullString{String: "arran", Valid: true},
			Text:     sql.NullString{String: "[img=image.jpg]", Valid: true},
		},
		3: {
			Username: sql.NullString{String: "casey", Valid: true},
			Text:     sql.NullString{String: "done", Valid: true},
		},
	}
	q := testhelpers.NewQuerierStub()
	q.GetCommentByIdForUserFn = func(ctx context.Context, arg db.GetCommentByIdForUserParams) (*db.GetCommentByIdForUserRow, error) {
		return comments[arg.ID], nil
	}

	cd := common.NewCoreData(context.Background(), q, nil)
	ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	QuoteSelectionApi(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	want := "[quoteof \"arran\" hello\n\n[img=image.jpg]]\n\n\n[quoteof \"casey\" done]\n"
	if got["text"] != want {
		t.Fatalf("handler returned unexpected text: got %q want %q", got["text"], want)
	}
}
