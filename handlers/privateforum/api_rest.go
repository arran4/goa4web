package privateforum

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/db"
	"github.com/gorilla/mux"
)

type apiError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type apiTopic struct {
	ID                int32    `json:"id"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	Participants      []string `json:"participants"`
	TotalParticipants int      `json:"total_participants"`
	Threads           int32    `json:"threads"`
	Comments          int32    `json:"comments"`
}

func writeAPIJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeAPIError(w http.ResponseWriter, status int, message string) {
	writeAPIJSON(w, status, apiError{Status: "error", Message: message})
}

// APIListTopics handles GET /api/privateforum/topics
func APIListTopics(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	// Retrieve topics via CoreData PrivateForumTopics which resolves access rights
	topics, err := cd.PrivateForumTopics()
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "unable to list topics")
		return
	}
	apiTopics := make([]apiTopic, 0, len(topics))
	for _, topic := range topics {
		if topic == nil || topic.ListPrivateTopicsByUserIDRow == nil {
			continue
		}
		apiTopics = append(apiTopics, apiTopic{
			ID:                topic.Idforumtopic,
			Title:             topic.DisplayTitle,
			Description:       topic.Description.String,
			Participants:      append([]string(nil), topic.Participants...),
			TotalParticipants: topic.TotalParticipants,
			Threads:           topic.Threads.Int32,
			Comments:          topic.Comments.Int32,
		})
	}

	startIdx, endIdx, hasMore, err := privateAPIPageOffset(r.URL.Query().Get("page"), cd.PageSize(), len(apiTopics))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, e := strconv.Atoi(pageStr); e == nil {
			page = p
		}
	}

	apiTopics = apiTopics[startIdx:endIdx]
	if len(apiTopics) == 0 {
		apiTopics = []apiTopic{}
	}

	writeAPIJSON(w, http.StatusOK, map[string]any{
		"topics":   apiTopics,
		"has_more": hasMore,
		"page":     page,
	})
}

// parsePrivateAPIID safely extracts and validates an ID for private forum API resources,
// rejecting zeroes, negatives, malformed input, and values exceeding 32-bit storage limits.
func parsePrivateAPIID(raw string) (int32, error) {
	parsed, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || parsed <= 0 {
		return 0, errors.New("invalid ID")
	}
	return int32(parsed), nil
}

// privateAPIPageOffset validates a requested pagination page, computes its safe zero-indexed
// slice offset bounds for a given slice length, and indicates whether more pages exist.
// Returns boundaries (start, end) bounded safely to avoid slice panic.
func privateAPIPageOffset(pageStr string, pageSize int, collectionLength int) (int, int, bool, error) {
	page := 1
	if pageStr != "" {
		parsed, err := strconv.ParseInt(pageStr, 10, 32)
		if err != nil || parsed <= 0 {
			return 0, 0, false, errors.New("invalid page")
		}
		page = int(parsed)
	}
	if pageSize <= 0 {
		return 0, 0, false, errors.New("invalid page size")
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	hasMore := collectionLength > end

	if start >= collectionLength {
		return collectionLength, collectionLength, false, nil
	}
	if end > collectionLength {
		end = collectionLength
	}
	return start, end, hasMore, nil
}

// APICreateTopic handles POST /api/privateforum/topics
func APICreateTopic(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid form")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	participantsStr := r.Form["participants"] // Allow multiple participants

	if title == "" {
		writeAPIError(w, http.StatusBadRequest, "title is required")
		return
	}

	// Pre-flight permission check to prevent username enumeration by unauthorized actors.
	if err := cd.CheckPrivateTopicCreatePermission(cd.UserID); err != nil {
		mapAPIError(w, err)
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
		writeAPIError(w, http.StatusBadRequest, "invalid participants: "+strings.Join(invalidUsers, ", "))
		return
	}

	cd.SetEventTask(privateTopicCreateTask)

	topicID, err := cd.CreatePrivateTopic(common.CreatePrivateTopicParams{
		CreatorID:    cd.UserID,
		Participants: participants,
		Title:        title,
		Description:  description,
	})

	if err != nil {
		mapAPIError(w, err)
		return
	}
	cd.SetEventData(topicCreateEventKey, topicCreateEventData{TopicID: topicID, BasePath: "/private"})

	writeAPIJSON(w, http.StatusOK, map[string]any{
		"status":  "success",
		"message": "Topic created",
		"topicID": topicID,
	})
}

// APIListThreads handles GET /api/privateforum/topic/{topic}/threads
func APIListThreads(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	topicID, err := parsePrivateAPIID(mux.Vars(r)["topic"])
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid topic ID")
		return
	}

	topic, err := cd.ForumTopicByID(topicID)
	if err != nil || topic == nil || topic.Handler != "private" {
		// Do not disclose whether an inaccessible ID exists.
		writeAPIError(w, http.StatusNotFound, "topic not found")
		return
	}

	rows, err := cd.ForumThreads(topicID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeAPIError(w, http.StatusNotFound, "topic not found")
		} else {
			writeAPIError(w, http.StatusInternalServerError, "unable to list threads")
		}
		return
	}

	startIdx, endIdx, hasMore, err := privateAPIPageOffset(r.URL.Query().Get("page"), cd.PageSize(), len(rows))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, e := strconv.Atoi(pageStr); e == nil {
			page = p
		}
	}

	rows = rows[startIdx:endIdx]
	if len(rows) == 0 {
		rows = []*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow{}
	}

	writeAPIJSON(w, http.StatusOK, map[string]any{
		"threads":  rows,
		"has_more": hasMore,
		"page":     page,
	})
}

// APIShowComments handles GET /api/privateforum/topic/{topic}/thread/{thread}
func APIShowComments(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	topicID, err := parsePrivateAPIID(mux.Vars(r)["topic"])
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid topic ID")
		return
	}

	threadID, err := parsePrivateAPIID(mux.Vars(r)["thread"])
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid thread ID")
		return
	}

	// Enforce topic/thread relationship and basic visibility before reading
	thread, err := cd.ForumThreadByID(threadID)
	if err != nil || thread == nil {
		writeAPIError(w, http.StatusNotFound, "thread not found")
		return
	}
	if thread.ForumtopicIdforumtopic != topicID {
		// Visible thread belongs to a different topic, reject it
		writeAPIError(w, http.StatusNotFound, "thread not found")
		return
	}

	comments, err := cd.ThreadComments(threadID)

	if err != nil && err != sql.ErrNoRows {
		writeAPIError(w, http.StatusInternalServerError, "unable to list comments")
		return
	}

	startIdx, endIdx, hasMore, err := privateAPIPageOffset(r.URL.Query().Get("page"), cd.PageSize(), len(comments))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, e := strconv.Atoi(pageStr); e == nil {
			page = p
		}
	}

	comments = comments[startIdx:endIdx]
	if len(comments) == 0 {
		comments = []*db.GetCommentsByThreadIdForUserRow{}
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

	writeAPIJSON(w, http.StatusOK, map[string]any{
		"comments": apiComments,
		"has_more": hasMore,
		"page":     page,
	})
}

func mapAPIError(w http.ResponseWriter, err error) {
	var notFoundErr common.ForumResourceNotFoundError
	var forbiddenErr common.ForumOperationForbiddenError
	var mismatchErr common.ForumHandlerMismatchError
	var invalidParticipantsErr *common.ErrInvalidParticipants
	var participantRequiredErr common.ErrPrivateTopicParticipantRequired

	switch {
	case errors.As(err, &notFoundErr):
		writeAPIError(w, http.StatusNotFound, "resource not found")
	case errors.As(err, &forbiddenErr):
		writeAPIError(w, http.StatusForbidden, "permission denied")
	case errors.As(err, &mismatchErr):
		writeAPIError(w, http.StatusNotFound, "resource not found") // fail closed for private API
	case errors.As(err, &invalidParticipantsErr), errors.As(err, &participantRequiredErr):
		writeAPIError(w, http.StatusBadRequest, err.Error())
	default:
		writeAPIError(w, http.StatusInternalServerError, "internal server error")
	}
}

// APIPostComment handles POST /api/privateforum/topic/{topic}/thread/{thread}/reply
func APIPostComment(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	if err := r.ParseForm(); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid form")
		return
	}

	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		writeAPIError(w, http.StatusBadRequest, "text is required")
		return
	}

	topicID, err := parsePrivateAPIID(mux.Vars(r)["topic"])
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid topic ID")
		return
	}

	threadID, err := parsePrivateAPIID(mux.Vars(r)["thread"])
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid thread ID")
		return
	}

	cd.SetEventTask(forumhandlers.ReplyTaskHandler)

	// Enforce topic/thread relationship and basic visibility before mutation
	thread, err := cd.ForumThreadByID(threadID)
	if err != nil || thread == nil {
		writeAPIError(w, http.StatusNotFound, "thread not found")
		return
	}
	if thread.ForumtopicIdforumtopic != topicID {
		// Visible thread belongs to a different topic, reject it
		writeAPIError(w, http.StatusNotFound, "thread not found")
		return
	}

	// Resolve language
	langID := cd.PreferredLanguageID("")
	if langID == 0 {
		langID = 1
	}

	result, err := cd.ReplyForumThread(r.Context(), common.ReplyForumThreadParams{
		ActorID:        cd.UserID,
		ThreadID:       threadID,
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

	writeAPIJSON(w, http.StatusOK, map[string]any{
		"status":     "success",
		"thread_id":  result.ThreadID,
		"topic_id":   result.TopicID,
		"comment_id": result.CommentID,
		"url":        result.URL,
		"appended":   result.Appended,
	})
}
