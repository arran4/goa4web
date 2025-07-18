package notifications

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	handlers "github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	dbdlq "github.com/arran4/goa4web/internal/dlq/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

func dlqRecordAndNotify(ctx context.Context, q dlq.DLQ, n Notifier, msg string) error {
	if q == nil {
		return fmt.Errorf("no dlq provider")
	}
	if err := q.Record(ctx, msg); err == nil {
		if dbq, ok := q.(dbdlq.DLQ); ok {
			if count, err := dbq.Queries.CountDeadLetters(ctx); err == nil {
				if isPow10(count) {
					NotifyAdmins(ctx, n, "/admin/dlq")
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
func buildPatterns(task tasks.Name, path string) []string {
	name := strings.ToLower(task.Name())
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

// renderMessage loads the template for the action and populates it with data.
func renderMessage(ctx context.Context, q *dbpkg.Queries, tmpls *template.Template, tmplName, path string, item interface{}, uid int32) (string, error) {
	var t *template.Template
	if q != nil {
		if body, err := q.GetTemplateOverride(ctx, tmplName); err == nil && body != "" {
			parsed, perr := template.New(tmplName).Parse(body)
			if perr != nil {
				return "", fmt.Errorf("parse template %s: %w", tmplName, perr)
			}
			t = parsed
		}
	}
	if t == nil {
		if tmpls == nil || tmpls.Lookup(tmplName) == nil {
			return "", nil
		}
		t = tmpls
	}
	var buf bytes.Buffer
	data := struct {
		Path   string
		Item   interface{}
		UserID int32
	}{Path: path, Item: item, UserID: uid}
	if err := t.ExecuteTemplate(&buf, tmplName, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", tmplName, err)
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// collectSubscribers returns a set of user IDs subscribed to any of the
// patterns using the specified delivery method.
func collectSubscribers(ctx context.Context, q *dbpkg.Queries, patterns []string, method string) (map[int32]struct{}, error) {
	subs := map[int32]struct{}{}
	for _, p := range patterns {
		ids, err := q.ListSubscribersForPattern(ctx, dbpkg.ListSubscribersForPatternParams{Pattern: p, Method: method})
		if err != nil {
			return nil, fmt.Errorf("list subscribers: %w", err)
		}
		for _, id := range ids {
			subs[id] = struct{}{}
		}
	}
	return subs, nil
}

// sendSubscriberEmail queues an email notification for a subscriber.
func sendSubscriberEmail(ctx context.Context, n Notifier, userID int32, evt eventbus.Event) error {
	user, err := n.Queries.GetUserById(ctx, userID)
	if err != nil || !user.Email.Valid || user.Email.String == "" {
		notifyMissingEmail(ctx, n.Queries, userID)
		return err
	}
	name, _ := evt.Task.(tasks.Name)
	return CreateEmailTemplateAndQueue(ctx, n.Queries, userID, user.Email.String, evt.Path, name.Name(), evt.Data)
}

// sendInternalNotification stores an internal notification for the user.
func sendInternalNotification(ctx context.Context, q *dbpkg.Queries, userID int32, path, msg string) error {
	return q.InsertNotification(ctx, dbpkg.InsertNotificationParams{
		UsersIdusers: userID,
		Link:         sql.NullString{String: path, Valid: path != ""},
		Message:      sql.NullString{String: msg, Valid: msg != ""},
	})
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
			if err := processEvent(ctx, evt, n, q); err != nil {
				log.Printf("process event: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func processEvent(ctx context.Context, evt eventbus.Event, n Notifier, q dlq.DLQ) error {
	if !handlers.NotificationsEnabled() {
		return nil
	}

	if evt.Task == nil {
		return nil
	}

	emailHtmlTemplates := templates.GetCompiledEmailHtmlTemplates(map[string]any{})
	emailTextTemplates := templates.GetCompiledEmailTextTemplates(map[string]any{})
	notificationTemplates := templates.GetCompiledNotificationTemplates(map[string]any{})

	if tp, ok := evt.Task.(AdminEmailTemplateProvider); ok {
		if err := notifyAdmins(ctx, evt, n, tp, emailHtmlTemplates, emailTextTemplates, notificationTemplates); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("admin notify: %v", err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}
	}

	if tp, ok := evt.Task.(SelfNotificationTemplateProvider); ok {
		if err := notifySelf(ctx, evt, n, tp, notificationTemplates); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("deliver self to %d: %v", evt.UserID, err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}

	}

	if tp, ok := evt.Task.(SubscribersNotificationTemplateProvider); ok {
		if err := notifySubscribers(ctx, evt, n, tp, notificationTemplates); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("notify subscribers: %v", err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}

	}

	if tp, ok := evt.Task.(AutoSubscribeProvider); ok {
		handleAutoSubscribe(ctx, evt, n, tp)

	}

	return nil
}

func notifySelf(ctx context.Context, evt eventbus.Event, n Notifier, tp SelfNotificationTemplateProvider, noteTmpls *template.Template) error {
	user, err := n.Queries.GetUserById(ctx, evt.UserID)
	if err != nil || !user.Email.Valid || user.Email.String == "" {
		notifyMissingEmail(ctx, n.Queries, evt.UserID)
	} else {
		name, _ := evt.Task.(tasks.Name)
		if err := CreateEmailTemplateAndQueue(ctx, n.Queries, evt.UserID, user.Email.String, evt.Path, name.Name(), evt.Data); err != nil {
			return err
		}
	}
	if nt := tp.SelfInternalNotificationTemplate(); nt != nil {
		name, _ := evt.Task.(tasks.Name)
		msg, err := renderMessage(ctx, n.Queries, noteTmpls, *nt, evt.Path, evt.Data, evt.UserID)
		if err != nil {
			return err
		}
		if msg != "" {
			if err := n.Queries.InsertNotification(ctx, dbpkg.InsertNotificationParams{
				UsersIdusers: evt.UserID,
				Link:         sql.NullString{String: evt.Path, Valid: true},
				Message:      sql.NullString{String: msg, Valid: true},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func notifySubscribers(ctx context.Context, evt eventbus.Event, n Notifier, tp SubscribersNotificationTemplateProvider, noteTmpls *template.Template) error {
	named, ok := evt.Task.(tasks.Name)
	if !ok {
		return nil
	}
	patterns := buildPatterns(named, evt.Path)

	emailSubs, err := collectSubscribers(ctx, n.Queries, patterns, "email")
	if err != nil {
		return err
	}
	internalSubs, err := collectSubscribers(ctx, n.Queries, patterns, "internal")
	if err != nil {
		return err
	}

	delete(emailSubs, evt.UserID)
	delete(internalSubs, evt.UserID)

	var msg string
	if nt := tp.SubscribedInternalNotificationTemplate(); nt != nil {
		name, _ := evt.Task.(tasks.Name)
		var err error
		msg, err = renderMessage(ctx, n.Queries, noteTmpls, *nt, evt.Path, evt.Data, evt.UserID)
		if err != nil {
			return err
		}
	}

	for id := range emailSubs {
		if err := sendSubscriberEmail(ctx, n, id, evt); err != nil {
			return fmt.Errorf("deliver email to %d: %w", id, err)
		}
	}

	if msg != "" {
		for id := range internalSubs {
			if err := sendInternalNotification(ctx, n.Queries, id, evt.Path, msg); err != nil {
				return fmt.Errorf("deliver internal to %d: %w", id, err)
			}
		}
	}

	return nil
}

func handleAutoSubscribe(ctx context.Context, evt eventbus.Event, n Notifier, tp AutoSubscribeProvider) {
	auto := true
	email := false
	if pref, err := n.Queries.GetPreferenceByUserID(ctx, evt.UserID); err == nil {
		auto = pref.AutoSubscribeReplies
		if pref.Emailforumupdates.Valid {
			email = pref.Emailforumupdates.Bool
		}
	}
	if auto {
		task, path := tp.AutoSubscribePath()
		pattern := buildPatterns(task, path)[0]
		if internalNotification {
			ensureSubscription(ctx, n.Queries, evt.UserID, pattern, "internal")
		}
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
			name, _ := evt.Task.(tasks.Name)
			if err := CreateEmailTemplateAndQueue(ctx, n.Queries, uid, addr, evt.Path, name.Name(), evt.Data); err != nil {
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
