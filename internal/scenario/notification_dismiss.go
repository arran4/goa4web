package scenario

import (
	"context"
	"fmt"
	"strings"

	notif "github.com/arran4/goa4web/internal/notifications"
)

// refTypeReplyActivity binds a logical reply independently of its comment ID.
// A comment can contain multiple append activities.
const refTypeReplyActivity RefType = "reply-activity"

// NotificationDismissData identifies a generated notification by recipient and reply.
type NotificationDismissData struct {
	Actor string
	Reply string
}

func (d *NotificationDismissData) Op() string { return "notification.dismiss" }

// NotificationDismissOp marks one generated reply notification read.
type NotificationDismissOp struct{}

func (o *NotificationDismissOp) OpName() string { return "notification.dismiss" }

func (o *NotificationDismissOp) AllowedHeaders() []string {
	return []string{"Op", "Actor", "Reply", "At"}
}

func (o *NotificationDismissOp) RequiredHeaders() []string { return []string{"Actor", "Reply"} }

func (o *NotificationDismissOp) DeclaredRef(*Event) (RefType, string, bool) {
	return "", "", false
}

func (o *NotificationDismissOp) ReferencedSymbols(evt *Event) []SymbolRef {
	return []SymbolRef{
		{Type: RefTypeUser, Symbol: strings.TrimSpace(evt.Headers.Get("Actor")), Field: "Actor"},
		{Type: RefTypePost, Symbol: strings.TrimSpace(evt.Headers.Get("Reply")), Field: "Reply"},
	}
}

func (o *NotificationDismissOp) AssetPaths(*Event) []string { return nil }

func (o *NotificationDismissOp) Parse(evt *Event) (OperationData, error) {
	return &NotificationDismissData{
		Actor: strings.TrimSpace(evt.Headers.Get("Actor")),
		Reply: strings.TrimSpace(evt.Headers.Get("Reply")),
	}, nil
}

func (r *Runner) applyNotificationDismiss(ctx context.Context, data *NotificationDismissData) error {
	actorID, ok := r.refRegistry.ResolveUser(data.Actor)
	if !ok {
		return fmt.Errorf("cannot resolve actor %q", data.Actor)
	}
	value, _ := r.refRegistry.Resolve(refTypeReplyActivity, data.Reply)
	activityID, ok := value.(string)
	if !ok || activityID == "" {
		return fmt.Errorf("cannot resolve reply activity %q", data.Reply)
	}
	actor := r.coreData.ForUser(actorID)
	id, err := actor.NotificationIDByActivity(ctx, notif.ForumReplyActivityKey(activityID))
	if err != nil {
		return fmt.Errorf("find actor's notification for reply %q: %w", data.Reply, err)
	}
	return actor.DismissNotification(ctx, id)
}
