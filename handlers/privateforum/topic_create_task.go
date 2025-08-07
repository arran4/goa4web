package privateforum

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

// PrivateTopicCreateTask creates a new private conversation and assigns grants.
type PrivateTopicCreateTask struct{ tasks.TaskString }

var privateTopicCreateTask = &PrivateTopicCreateTask{TaskString: TaskPrivateTopicCreate}

var _ tasks.Task = (*PrivateTopicCreateTask)(nil)

// Action creates a new private topic and assigns view permissions to participants.
func (PrivateTopicCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	parts := strings.Split(r.PostFormValue("participants"), ",")
	body := strings.TrimSpace(r.PostFormValue("body"))
	var uids []int32
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: p, Valid: true})
		if err != nil {
			continue
		}
		uids = append(uids, u.Idusers)
	}
	creator := cd.UserID
	seen := false
	for _, id := range uids {
		if id == creator {
			seen = true
			break
		}
	}
	if creator != 0 && !seen {
		uids = append(uids, creator)
	}
	if !cd.HasGrant("privateforum", "topic", "post", 0) {
		log.Printf("private topic create denied: user=%d", creator)
		err := handlers.ErrForbidden
		return fmt.Errorf("UserCanCreateTopic deny %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	tid, err := queries.CreateForumTopicForPoster(r.Context(), db.CreateForumTopicForPosterParams{
		PosterID:        creator,
		ForumcategoryID: common.PrivateForumCategoryID,
		LanguageID:      0,
		Title:           sql.NullString{},
		Description:     sql.NullString{},
		GrantCategoryID: sql.NullInt32{Int32: common.PrivateForumCategoryID, Valid: true},
		GranteeID:       sql.NullInt32{Int32: creator, Valid: creator != 0},
	})
	if err != nil {
		return fmt.Errorf("create topic %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if tid == 0 {
		err := handlers.ErrForbidden
		return fmt.Errorf("create topic %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	topicID := int32(tid)
	if err := queries.SystemSetForumTopicHandlerByID(r.Context(), db.SystemSetForumTopicHandlerByIDParams{Handler: "private", ID: topicID}); err != nil {
		return fmt.Errorf("set handler %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	thid, err := queries.SystemCreateThread(r.Context(), topicID)
	if err != nil {
		return fmt.Errorf("create thread %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	threadID := int32(thid)
	for _, uid := range uids {
		for _, act := range []string{"see", "view"} {
			if _, err := queries.SystemCreateGrant(r.Context(), db.SystemCreateGrantParams{
				UserID:   sql.NullInt32{Int32: uid, Valid: true},
				RoleID:   sql.NullInt32{},
				Section:  "forum",
				Item:     sql.NullString{String: "topic", Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				ItemRule: sql.NullString{},
				Action:   act,
				Extra:    sql.NullString{},
			}); err != nil {
				return fmt.Errorf("create %s grant %w", act, handlers.ErrRedirectOnSamePageHandler(err))
			}
		}
		if _, err := queries.SystemCreateGrant(r.Context(), db.SystemCreateGrantParams{
			UserID:   sql.NullInt32{Int32: uid, Valid: true},
			RoleID:   sql.NullInt32{},
			Section:  "forum",
			Item:     sql.NullString{String: "thread", Valid: true},
			RuleType: "allow",
			ItemID:   sql.NullInt32{Int32: threadID, Valid: true},
			ItemRule: sql.NullString{},
			Action:   "reply",
			Extra:    sql.NullString{},
		}); err != nil {
			return fmt.Errorf("create reply grant %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	cid, err := queries.CreateCommentForCommenter(r.Context(), db.CreateCommentForCommenterParams{
		LanguageID:         0,
		CommenterID:        creator,
		ForumthreadID:      threadID,
		Text:               sql.NullString{String: body, Valid: body != ""},
		GrantForumthreadID: sql.NullInt32{Int32: threadID, Valid: true},
		GranteeID:          sql.NullInt32{Int32: creator, Valid: creator != 0},
	})
	if err != nil {
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cid == 0 {
		err := handlers.ErrForbidden
		return fmt.Errorf("create comment %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(cid), ThreadID: threadID, TopicID: topicID}
		evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: body}
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, threadID)}
}
