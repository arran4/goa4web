package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
)

// AddUserSubscriptionTask creates a subscription for a user.
type AddUserSubscriptionTask struct{ tasks.TaskString }

var addUserSubscriptionTask = &AddUserSubscriptionTask{TaskString: TaskAdd}

var _ tasks.Task = (*AddUserSubscriptionTask)(nil)

func (AddUserSubscriptionTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	uid := cd.CurrentProfileUserID()
	if uid == 0 {
		return fmt.Errorf("user not found %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	pattern := r.PostFormValue("pattern")
	method := r.PostFormValue("method")
	if pattern == "" || method == "" {
		return fmt.Errorf("missing pattern or method %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if err := cd.Queries().InsertSubscription(r.Context(), db.InsertSubscriptionParams{
		UsersIdusers: uid,
		Pattern:      pattern,
		Method:       method,
	}); err != nil {
		return fmt.Errorf("insert subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d/subscriptions", uid)}
}

// UpdateUserSubscriptionTask modifies a user's subscription.
type UpdateUserSubscriptionTask struct{ tasks.TaskString }

var updateUserSubscriptionTask = &UpdateUserSubscriptionTask{TaskString: TaskUpdate}

var _ tasks.Task = (*UpdateUserSubscriptionTask)(nil)

func (UpdateUserSubscriptionTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	uid := cd.CurrentProfileUserID()
	if uid == 0 {
		return fmt.Errorf("user not found %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	idStr := r.PostFormValue("id")
	pattern := r.PostFormValue("pattern")
	method := r.PostFormValue("method")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if pattern == "" || method == "" {
		return fmt.Errorf("missing pattern or method %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if err := cd.Queries().UpdateSubscriptionByIDForSubscriber(r.Context(), db.UpdateSubscriptionByIDForSubscriberParams{
		Pattern:      pattern,
		Method:       method,
		SubscriberID: uid,
		ID:           int32(id),
	}); err != nil {
		return fmt.Errorf("update subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d/subscriptions", uid)}
}

// DeleteUserSubscriptionTask removes a user's subscription.
type DeleteUserSubscriptionTask struct{ tasks.TaskString }

var deleteUserSubscriptionTask = &DeleteUserSubscriptionTask{TaskString: TaskDelete}

var _ tasks.Task = (*DeleteUserSubscriptionTask)(nil)

func (DeleteUserSubscriptionTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.LoadSelectionsFromRequest(r)
	uid := cd.CurrentProfileUserID()
	if uid == 0 {
		return fmt.Errorf("user not found %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("")))
	}
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	idStr := r.PostFormValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("id parse fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if err := cd.Queries().DeleteSubscriptionByIDForSubscriber(r.Context(), db.DeleteSubscriptionByIDForSubscriberParams{
		SubscriberID: uid,
		ID:           int32(id),
	}); err != nil {
		return fmt.Errorf("delete subscription fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	return handlers.RefreshDirectHandler{TargetURL: fmt.Sprintf("/admin/user/%d/subscriptions", uid)}
}
