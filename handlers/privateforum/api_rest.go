package privateforum

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"html/template"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
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

	// Retrieve topics via CoreData PrivateForumTopics which resolves access rights
	topics, err := cd.PrivateForumTopics()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

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
	_ = json.NewEncoder(w).Encode(map[string]any{
		"topics":   topics,
		"has_more": hasMore,
		"page":     page,
	})
}

// APICreateTopic handles POST /api/privateforum/topics
func APICreateTopic(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	participantsStr := r.Form["participants"] // Allow multiple participants

	if title == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "title is required",
		})
		return
	}

	// Resolve participants by username
	var participants []common.PrivateTopicParticipant
	var invalidUsers []string

	// Make sure the creator is included
	participants = append(participants, common.PrivateTopicParticipant{ID: cd.UserID})

	for _, username := range participantsStr {
		username = strings.TrimSpace(username)
		if username == "" {
			continue
		}
		userRow, err := cd.Queries().SystemGetUserByUsername(r.Context(), sql.NullString{String: username, Valid: true})
		if err != nil || userRow.Idusers == cd.UserID {
			if err != nil {
				invalidUsers = append(invalidUsers, username)
			}
			continue
		}
		participants = append(participants, common.PrivateTopicParticipant{ID: userRow.Idusers})
	}

	if len(invalidUsers) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "Invalid participants: " + strings.Join(invalidUsers, ", "),
		})
		return
	}

	hasGrant := cd.HasGrant("privateforum", "topic", "create", 0)

	if !hasGrant {
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": "permission denied",
		})
		return
	}

	topicID, err := cd.CreatePrivateTopic(common.CreatePrivateTopicParams{
		CreatorID:    cd.UserID,
		Participants: participants,
		Title:        title,
		Description:  description,
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Topic created",
		"topicID": topicID,
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

	rows, err := cd.ForumThreads(int32(topicID))

	if err != nil && err != sql.ErrNoRows {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"threads":  rows,
		"has_more": hasMore,
		"page":     page,
	})
}

// APIShowComments handles GET /api/privateforum/topic/{topic}/thread/{thread}
func APIShowComments(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	topicIDStr := mux.Vars(r)["topic"]
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil || topicID <= 0 {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	threadIDStr := mux.Vars(r)["thread"]
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil || threadID <= 0 {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Enforce topic/thread relationship and basic visibility before reading
	thread, err := cd.ForumThreadByID(int32(threadID))
	if err != nil || thread == nil {
		http.Error(w, "Thread not found", http.StatusNotFound)
		return
	}
	if thread.ForumtopicIdforumtopic != int32(topicID) {
		// Visible thread belongs to a different topic, reject it
		http.Error(w, "Thread not found", http.StatusNotFound)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := cd.PageSize()

	comments, err := cd.ThreadComments(int32(threadID))

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
	_ = json.NewEncoder(w).Encode(map[string]any{
		"comments": apiComments,
		"has_more": hasMore,
		"page":     page,
	})
}

func mapAPIError(w http.ResponseWriter, err error) {
	var notFoundErr common.ForumResourceNotFoundError
	var forbiddenErr common.ForumOperationForbiddenError
	var mismatchErr common.ForumHandlerMismatchError

	switch {
	case errors.As(err, &notFoundErr):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.As(err, &forbiddenErr):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.As(err, &mismatchErr):
		http.Error(w, err.Error(), http.StatusNotFound) // fail closed for private API
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	topicIDStr := mux.Vars(r)["topic"]
	topicID, err := strconv.Atoi(topicIDStr)
	if err != nil || topicID <= 0 {
		http.Error(w, "Invalid topic ID", http.StatusBadRequest)
		return
	}

	threadIDStr := mux.Vars(r)["thread"]
	threadID, err := strconv.Atoi(threadIDStr)
	if err != nil || threadID <= 0 {
		http.Error(w, "Invalid thread ID", http.StatusBadRequest)
		return
	}

	// Enforce topic/thread relationship and basic visibility before mutation
	thread, err := cd.ForumThreadByID(int32(threadID))
	if err != nil || thread == nil {
		http.Error(w, "Thread not found", http.StatusNotFound)
		return
	}
	if thread.ForumtopicIdforumtopic != int32(topicID) {
		// Visible thread belongs to a different topic, reject it
		http.Error(w, "Thread not found", http.StatusNotFound)
		return
	}

	// Resolve language
	langID := cd.PreferredLanguageID("")
	if langID == 0 {
		langID = 1
	}

	result, err := cd.ReplyForumThread(r.Context(), common.ReplyForumThreadParams{
		ActorID:        cd.UserID,
		ThreadID:       int32(threadID),
		LanguageID:     int32(langID),
		Text:           text,
		Private:        true,
		EnforceHandler: true,
		BasePath:       "/private",
	})

	if err != nil {
		mapAPIError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":     "success",
		"thread_id":  result.ThreadID,
		"topic_id":   result.TopicID,
		"comment_id": result.CommentID,
		"url":        result.URL,
		"appended":   result.Appended,
	})
}
