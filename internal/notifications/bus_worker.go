package notifications

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"text/template"

	hcommon "github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	dbdlq "github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/utils/emailutil"
	"time"
)

type namedTask struct{ name string }

func (n namedTask) TaskName() string { return n.name }

func recordAndNotify(ctx context.Context, q dlq.DLQ, n Notifier, msg string) {
	if q != nil {
		_ = q.Record(ctx, msg)
		if dbq, ok := q.(dbdlq.DLQ); ok {
			if count, err := dbq.Queries.CountDeadLetters(ctx); err == nil {
				if isPow10(count) {
					n.NotifyAdmins(ctx, "/admin/dlq")
				}
			}
		}
	} else {
		log.Print(msg)
	}
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

// BusWorker listens for events on the bus and sends notifications.
func shouldNotify(ctx context.Context, q *dbpkg.Queries, task eventbus.NamedTask) bool {
	name := strings.ToLower(task.TaskName())
	if _, ok := defaultTemplates[name]; ok {
		return true
	}
	if q != nil {
		tmpl := fmt.Sprintf("notify_%s", name)
		if body, err := q.GetTemplateOverride(ctx, tmpl); err == nil && body != "" {
			return true
		}
	}
	return false
}

// buildPatterns expands the task/path pair into all matching subscription patterns.
func buildPatterns(task eventbus.NamedTask, path string) []string {
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

func renderMessage(ctx context.Context, q *dbpkg.Queries, action, path string, item interface{}) string {
	name := fmt.Sprintf("notify_%s", strings.ToLower(action))
	tmplText := ""
	if q != nil {
		if body, err := q.GetTemplateOverride(ctx, name); err == nil && body != "" {
			tmplText = body
		}
	}
	if tmplText == "" {
		if d, ok := defaultTemplates[strings.ToLower(action)]; ok {
			tmplText = d
		}
	}
	if tmplText == "" {
		return ""
	}
	t, err := template.New("msg").Parse(tmplText)
	if err != nil {
		log.Printf("parse template %s: %v", name, err)
		return ""
	}
	var buf bytes.Buffer
	_ = t.Execute(&buf, struct {
		Action string
		Path   string
		Item   interface{}
	}{Action: action, Path: path, Item: item})
	msg := buf.String()
	msg = strings.TrimSuffix(msg, "\n")
	return msg
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
	if !shouldNotify(ctx, n.Queries, namedTask{evt.Task}) || evt.UserID == 0 || evt.Path == "" {
		return
	}
	if !hcommon.NotificationsEnabled() {
		return
	}

	if evt.Admin {
		n.NotifyAdmins(ctx, evt.Path)
	}

	if evt.Task == hcommon.TaskReply && n.Queries != nil {
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

	itemType, targetID, ok := parseEvent(evt)
	if ok {
		switch itemType {
		case "thread":
			n.NotifyThreadSubscribers(ctx, targetID, evt.UserID, evt.Path)
		case "writing":
			n.NotifyWritingSubscribers(ctx, targetID, evt.UserID, evt.Path)
		}
	}

	patterns := buildPatterns(namedTask{evt.Task}, evt.Path)
	subs := map[int32]map[string]func(context.Context) error{}
	msg := renderMessage(ctx, n.Queries, evt.Task, evt.Path, evt.Data)
	for _, p := range patterns {
		for _, method := range []string{"email", "internal"} {
			ids, err := n.Queries.ListSubscribersForPattern(ctx, dbpkg.ListSubscribersForPatternParams{Pattern: p, Method: method})
			if err != nil {
				recordAndNotify(ctx, q, n, fmt.Sprintf("list subscribers: %v", err))
				continue
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
						return emailutil.CreateEmailTemplateAndQueue(c, n.Queries, uid, user.Email.String, evt.Path, evt.Task, evt.Data)
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
				recordAndNotify(ctx, q, n, fmt.Sprintf("deliver %s to %d: %v", typ, id, err))
			}
		}
	}
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
