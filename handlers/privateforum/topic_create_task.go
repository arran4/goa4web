package privateforum

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	notif "github.com/arran4/goa4web/internal/notifications"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/postcountworker"
)

// PrivateTopicCreateTask creates a new private conversation and assigns grants.
type PrivateTopicCreateTask struct{ tasks.TaskString }

var privateTopicCreateTask = &PrivateTopicCreateTask{TaskString: TaskPrivateTopicCreate}

var (
	_ tasks.Task                  = (*PrivateTopicCreateTask)(nil)
	_ notif.AutoSubscribeProvider = (*PrivateTopicCreateTask)(nil)
)

// Action creates a new private topic and assigns view permissions to participants.
func (PrivateTopicCreateTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	queries := cd.Queries()
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	participants := r.PostFormValue("participants")
	parts := strings.Split(participants, ",")
	title := strings.TrimSpace(r.PostFormValue("title"))
	description := strings.TrimSpace(r.PostFormValue("description"))
	var uids []int32 // TODO make map
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		u, err := queries.SystemGetUserByUsername(r.Context(), sql.NullString{String: p, Valid: true})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				cd.SetCurrentError(fmt.Sprintf("unknown user %q", p))
				forumhandlers.CreateTopicPageWithPostTask(w, r, TaskPrivateTopicCreate, &forumhandlers.CreateTopicPageForm{
					Participants: participants,
					Title:        title,
					Description:  description,
				})
				return nil
			}
			return fmt.Errorf("unknown error %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		// Use SystemCheckGrant so role-based grants are considered for the participant.
		// We pass the participant ID as both ViewerID and UserID to test their own access.
		if _, err := queries.SystemCheckGrant(r.Context(), db.SystemCheckGrantParams{
			ViewerID: u.Idusers,
			Section:  "privateforum",
			Item:     sql.NullString{String: "topic", Valid: true},
			Action:   "see",
			ItemID:   sql.NullInt32{Valid: false},
			UserID:   sql.NullInt32{Int32: u.Idusers, Valid: true},
		}); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("checking user grant: %w", handlers.ErrRedirectOnSamePageHandler(err))
			}
			// NoRows means no matching grant (direct or via role)
			cd.SetCurrentError(fmt.Sprintf("user %q does not have permission to access private forums", p))
			forumhandlers.CreateTopicPageWithPostTask(w, r, TaskPrivateTopicCreate, &forumhandlers.CreateTopicPageForm{
				Participants: participants,
				Title:        title,
				Description:  description,
			})
			return nil
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
	topicID, err := cd.CreatePrivateTopic(common.CreatePrivateTopicParams{CreatorID: creator, ParticipantIDs: uids, Title: title, Description: description})
	if err != nil {
		return fmt.Errorf("create private topic %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	for _, uid := range uids {
		if err := cd.SubscribeTopic(uid, topicID); err != nil {
			return fmt.Errorf("subscribe topic for user %d: %w", uid, handlers.ErrRedirectOnSamePageHandler(err))
		}
	}
	base := cd.ForumBasePath
	if base == "" {
		base = "/forum"
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("%s/topic/%d", base, topicID)}
}

// AutoSubscribePath ensures conversation creators follow replies and future threads.
// AutoSubscribePath implements notif.AutoSubscribeProvider. When postcountworker
// context is available it subscribes to the created thread so authors receive
// updates on subsequent comments.
func (PrivateTopicCreateTask) AutoSubscribePath(evt eventbus.TaskEvent) (string, string, error) {
	if data, ok := evt.Data[postcountworker.EventKey].(postcountworker.UpdateEventData); ok {
		base := "/forum"
		if idx := strings.Index(evt.Path, "/topic/"); idx > 0 {
			base = evt.Path[:idx]
		}
		return string(TaskPrivateTopicCreate), fmt.Sprintf("%s/topic/%d", base, data.TopicID), nil
	}
	return string(TaskPrivateTopicCreate), evt.Path, nil
}
