package notifications

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/tasks"
	"log"
	"strings"
	"text/template"

	handlers "github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	dbdlq "github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"time"
)

type namedTask struct{ name string }

func (n namedTask) TaskName() string { return n.name }

func dlqRecordAndNotify(ctx context.Context, q dlq.DLQ, n Notifier, msg string) error {
	if q == nil {
		return fmt.Errorf("no dlq provider")
	}
	if err := q.Record(ctx, msg); err == nil {
		if dbq, ok := q.(dbdlq.DLQ); ok {
			if count, err := dbq.Queries.CountDeadLetters(ctx); err == nil {
				if isPow10(count) {
					n.NotifyAdmins(ctx, "/admin/dlq")
				}
			}
		}
	}
	return nil
}

func isPow10(n int64) bool {
	if n < 1 {
		return false
	}
	for n%10 == 0 {
		n /= 10
	}
	return n == 1
}

// buildPatterns expands the task/path pair into all matching subscription patterns.
func buildPatterns(task tasks.NamedTask, path string) []string {
	name := strings.ToLower(task.TaskName())
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{fmt.Sprintf("%s:/*", name)}
	}
	parts := strings.Split(path, "/")
	patterns := []string{fmt.Sprintf("%s:/%s", name, path)}
	for i := len(parts) - 1; i >= 1; i-- {
		prefix := strings.Join(parts[:i], "/")
		patterns = append(patterns, fmt.Sprintf("%s:/%s/*", name, prefix))
	}
	patterns = append(patterns, fmt.Sprintf("%s:/*", name))
	return patterns
}

// parseEvent identifies a subscription target from the request path.
// It returns the item type and id if recognised.
func parseEvent(evt eventbus.Event) (string, int32, bool) {
	if evt.Data == nil {
		return "", 0, false
	}
	if v, ok := evt.Data["target"]; ok {
		if t, ok := v.(SubscriptionTarget); ok {
			typ, id := t.SubscriptionTarget()
			return typ, id, true
		}
	}
	return "", 0, false
}

func BusWorker(ctx context.Context, bus *eventbus.Bus, n Notifier, q dlq.DLQ) {
	if bus == nil || n.Queries == nil {
		return
	}
	ch := bus.Subscribe()
	for {
		select {
		case evt := <-ch:
			processEvent(ctx, evt, n, q)
		case <-ctx.Done():
			return
		}
	}
}

func processEvent(ctx context.Context, evt eventbus.Event, n Notifier, q dlq.DLQ) {
	if !handlers.NotificationsEnabled() {
		return
	}

	if evt.Task == nil {
		return
	}

	emailHtmlTemplates := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	emailTextTemplates := templates.GetCompiledEmailTextTemplates(map[string]any{})
	notificationTemplates := templates.GetCompiledNotificationTemplates(map[string]any{})

	if tp, ok := evt.Task.(AdminEmailTemplateProvider); ok {
		if err := notifyAdmins(ctx, evt, n, tp, emailHtmlTemplates, emailTextTemplates, notificationTemplates); err != nil {
			dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("admin notify: %v", err))
		}
	}

	if tp, ok := evt.Task.(SelfNotificationTemplateProvider); ok {
		if err := notifySelf(ctx, evt, n, tp); err != nil {
			dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("deliver self to %d: %v", evt.UserID, err))
		}

	}

	if tp, ok := evt.Task.(SubscribersNotificationTemplateProvider); ok {
		if err := notifySubscribers(ctx, evt, n); err != nil {
			dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("notify subscribers: %v", err))
		}

	}

	if tp, ok := evt.Task.(AutoSubscribeProvider); ok {
		handleAutoSubscribe(ctx, evt, n)

	}
}

func notifySelf(ctx context.Context, evt eventbus.Event, n Notifier, tp SelfNotificationTemplateProvider) error {
	user, err := n.Queries.GetUserById(ctx, evt.UserID)
	if err != nil || !user.Email.Valid || user.Email.String == "" {
		notifyMissingEmail(ctx, n.Queries, evt.UserID)
	} else {
		if err := CreateEmailTemplateAndQueue(ctx, n.Queries, evt.UserID, user.Email.String, evt.Path, evt.Task, evt.Data); err != nil {
			return err
		}
	}
	msg := renderMessage(ctx, n.Queries, evt.Task, evt.Path, evt.Data)
	if msg != "" {
		if err := n.Queries.InsertNotification(ctx, dbpkg.InsertNotificationParams{
			UsersIdusers: evt.UserID,
			Link:         sql.NullString{String: evt.Path, Valid: true},
			Message:      sql.NullString{String: msg, Valid: true},
		}); err != nil {
			return err
		}
	}
	return nil
}

func notifySubscribers(ctx context.Context, evt eventbus.Event, n Notifier) error {
	patterns := buildPatterns(namedTask{evt.Task}, evt.Path)
	subs := map[int32]map[string]func(context.Context) error{}
	msg := renderMessage(ctx, n.Queries, evt.Task, evt.Path, evt.Data)

	for _, p := range patterns {
		for _, method := range []string{"email", "internal"} {
			ids, err := n.Queries.ListSubscribersForPattern(ctx, dbpkg.ListSubscribersForPatternParams{Pattern: p, Method: method})
			if err != nil {
				return fmt.Errorf("list subscribers: %w", err)
			}
			for _, id := range ids {
				if id == evt.UserID {
					continue
				}
				if subs[id] == nil {
					subs[id] = map[string]func(context.Context) error{}
				}
				if method == "email" {
					uid := id
					subs[id][method] = func(c context.Context) error {
						user, err := n.Queries.GetUserById(c, uid)
						if err != nil || !user.Email.Valid || user.Email.String == "" {
							notifyMissingEmail(c, n.Queries, uid)
							return err
						}
						return CreateEmailTemplateAndQueue(c, n.Queries, uid, user.Email.String, evt.Path, evt.Task, evt.Data)
					}
				} else if method == "internal" && msg != "" {
					uid := id
					subs[id][method] = func(c context.Context) error {
						return n.Queries.InsertNotification(c, dbpkg.InsertNotificationParams{
							UsersIdusers: uid,
							Link:         sql.NullString{String: evt.Path, Valid: true},
							Message:      sql.NullString{String: msg, Valid: true},
						})
					}
				}
			}
		}
	}

	for id, methods := range subs {
		for typ, fn := range methods {
			if err := fn(ctx); err != nil {
				return fmt.Errorf("deliver %s to %d: %w", typ, id, err)
			}
		}
	}
	return nil
}

func handleAutoSubscribe(ctx context.Context, evt eventbus.Event, n Notifier) {
	auto := true
	email := false
	if pref, err := n.Queries.GetPreferenceByUserID(ctx, evt.UserID); err == nil {
		auto = pref.AutoSubscribeReplies
		if pref.Emailforumupdates.Valid {
			email = pref.Emailforumupdates.Bool
		}
	}
	if auto {
		pattern := buildPatterns(namedTask{evt.Task}, evt.Path)[0]
		ensureSubscription(ctx, n.Queries, evt.UserID, pattern, "internal")
		if email {
			ensureSubscription(ctx, n.Queries, evt.UserID, pattern, "email")
		}
	}
}

func notifyAdmins(ctx context.Context, evt eventbus.Event, n Notifier, tp AdminEmailTemplateProvider, htmlTmpls, textTmpls, noteTmpls *template.Template) error {
	if !config.AdminNotificationsEnabled() {
		return nil
	}
	for _, addr := range config.GetAdminEmails(ctx, n.Queries) {
		var uid int32
		if n.Queries != nil {
			if u, err := n.Queries.UserByEmail(ctx, addr); err == nil {
				uid = u.Idusers
			} else {
				log.Printf("user by email %s: %v", addr, err)
			}
		}
		if et := tp.AdminEmailTemplate(); et != nil {
			if err := CreateEmailTemplateAndQueue(ctx, n.Queries, uid, addr, evt.Path, evt.Task, evt.Data); err != nil {
				return err
			}
		}
		if nt := tp.AdminInternalNotificationTemplate(); nt != nil && n.Queries != nil {
			var buf bytes.Buffer
			if err := noteTmpls.ExecuteTemplate(&buf, *nt, evt.Data); err != nil {
				return err
			}
			if err := n.Queries.InsertNotification(ctx, dbpkg.InsertNotificationParams{
				UsersIdusers: uid,
				Link:         sql.NullString{String: evt.Path, Valid: evt.Path != ""},
				Message:      sql.NullString{String: buf.String(), Valid: buf.Len() > 0},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureSubscription(ctx context.Context, q *dbpkg.Queries, userID int32, pattern, method string) {
	if q == nil || userID == 0 {
		return
	}
	ids, err := q.ListSubscribersForPattern(ctx, dbpkg.ListSubscribersForPatternParams{Pattern: pattern, Method: method})
	if err == nil {
		for _, id := range ids {
			if id == userID {
				return
			}
		}
	}
	if err := q.InsertSubscription(ctx, dbpkg.InsertSubscriptionParams{UsersIdusers: userID, Pattern: pattern, Method: method}); err != nil {
		log.Printf("insert subscription: %v", err)
	}
}

func notifyMissingEmail(ctx context.Context, q *dbpkg.Queries, userID int32) {
	if q == nil || userID == 0 {
		return
	}
	last, err := q.LastNotificationByMessage(ctx, dbpkg.LastNotificationByMessageParams{UsersIdusers: userID, Message: sql.NullString{String: "missing email address", Valid: true}})
	if err == nil && time.Since(last.CreatedAt) < 7*24*time.Hour {
		return
	}
	_ = q.InsertNotification(ctx, dbpkg.InsertNotificationParams{UsersIdusers: userID, Message: sql.NullString{String: "missing email address", Valid: true}})
}
