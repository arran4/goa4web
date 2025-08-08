package privateforum

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/tasks"
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
	topicID, threadID, err := cd.CreatePrivateTopic(common.CreatePrivateTopicParams{CreatorID: creator, ParticipantIDs: uids, Body: body})
	if err != nil {
		return fmt.Errorf("create private topic %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/forum/topic/%d/thread/%d", topicID, threadID)}
}
