package privateforum

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"html/template"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

// APIListTopics handles GET /api/privateforum/topics
func APIListTopics(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := cd.PageSize()
	offset := (page - 1) * pageSize

	// Retrieve topics via CoreData PrivateTopics which resolves access rights
	topics := cd.PrivateTopics()

	// Poor-man's pagination logic for the returned slice
	hasMore := false
	if len(topics) > offset {
		end := offset + pageSize
		if len(topics) > end {
			hasMore = true
		} else {
			end = len(topics)
		}
		topics = topics[offset:end]
	} else {
		topics = nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"topics":   topics,
		"has_more": hasMore,
		"page":     page,
	})
}

// APICreateTopic handles POST /api/privateforum/topics
func APICreateTopic(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	subject := strings.TrimSpace(r.FormValue("subject"))
	text := strings.TrimSpace(r.FormValue("text"))

	if subject == "" || text == "" {
		http.Error(w, "Subject and text are required", http.StatusBadRequest)
		return
	}

	// Add user validation logic from topic_create_task.go
	if err := cd.ValidateCodeImagesForUser(cd.UserID, text); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simplistic create topic logic, ideally invoking task or business logic
	// For API simplicity, we can invoke the task action directly or duplicate core logic

	// Using the existing task logic directly
	res := privateTopicCreateTask.Action(w, r)
	if err, ok := res.(error); ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Topic created",
	})
}

// APIListThreads handles GET /api/privateforum/topic/{topic}/threads
func APIListThreads(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	topicIDStr := mux.Vars(r)["topic"]
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := cd.PageSize()

	rows, err := cd.Queries().GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(r.Context(), db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams{
		TopicID:       int32(topicID),
		ViewerID:      cd.UserID,
		ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"threads":  rows,
		"has_more": hasMore,
		"page":     page,
	})
}

// APIShowComments handles GET /api/privateforum/topic/{topic}/thread/{thread}
func APIShowComments(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	threadIDStr := mux.Vars(r)["thread"]
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := cd.PageSize()

	comments, err := cd.Queries().GetCommentsByThreadIdForUser(r.Context(), db.GetCommentsByThreadIdForUserParams{
		ViewerID: cd.UserID,
		ThreadID: int32(threadID),
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
	})

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	offset := (page - 1) * pageSize
	hasMore := false
	if len(comments) > offset {
		end := offset + pageSize
		if len(comments) > end {
			hasMore = true
		} else {
			end = len(comments)
		}
		comments = comments[offset:end]
	} else {
		comments = nil
	}

	// Format text into HTML for the API response
	type apiComment struct {
		ID       int32  `json:"id"`
		Username string `json:"username"`
		Text     string `json:"text"`
		HTML     string `json:"html"`
		Written  string `json:"written"`
	}

	apiComments := make([]apiComment, 0, len(comments))
	for _, c := range comments {
		html := ""
		if c.Text.Valid {
			html = string(cd.Funcs(r)["a4code2html"].(func(string) template.HTML)(c.Text.String))
		}

		username := ""
		if c.Posterusername.Valid {
			username = c.Posterusername.String
		}

		apiComments = append(apiComments, apiComment{
			ID:       c.Idcomments,
			Username: username,
			Text:     c.Text.String,
			HTML:     html,
			Written:  c.Written.Time.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"comments": apiComments,
		"has_more": hasMore,
		"page":     page,
	})
}

// APIPostComment handles POST /api/privateforum/topic/{topic}/thread/{thread}/reply
func APIPostComment(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	threadIDStr := mux.Vars(r)["thread"]
	threadID, _ := strconv.Atoi(threadIDStr)

	if err := cd.ValidateCodeImagesForThread(cd.UserID, int32(threadID), text); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := forumhandlers.ReplyTaskHandler.Action(w, r)
	if err, ok := res.(error); ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Comment posted",
	})
}
