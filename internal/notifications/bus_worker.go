package notifications

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/tasks"
)

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

// collectSubscribers returns a set of user IDs subscribed to any of the
// patterns using the specified delivery method.
func collectSubscribers(ctx context.Context, q *dbpkg.Queries, patterns []string, method string) (map[int32]struct{}, error) {
	subs := map[int32]struct{}{}
	ids, err := q.ListSubscribersForPatterns(ctx, dbpkg.ListSubscribersForPatternsParams{Patterns: patterns, Method: method})
	if err != nil {
		return nil, fmt.Errorf("list subscribers: %w", err)
	}
	for _, id := range ids {
		subs[id] = struct{}{}
	}
	return subs, nil
}

func (n *Notifier) BusWorker(ctx context.Context, bus *eventbus.Bus, q dlq.DLQ) {
	if bus == nil || n.Queries == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case msg := <-ch:
			evt, ok := msg.(eventbus.TaskEvent)
			if !ok {
				continue
			}
			if err := n.processEvent(ctx, evt, q); err != nil {
				log.Printf("process event: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (n *Notifier) processEvent(ctx context.Context, evt eventbus.TaskEvent, q dlq.DLQ) error {
	if !handlers.NotificationsEnabled() {
		return nil
	}

	if evt.Task == nil {
		return nil
	}

	if tp, ok := evt.Task.(AdminEmailTemplateProvider); ok {
		if err := n.notifyAdmins(ctx, evt, tp); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("admin notify: %v", err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}
	}

	if tp, ok := evt.Task.(SelfNotificationTemplateProvider); ok {
		if err := n.notifySelf(ctx, evt, tp); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("deliver self to %d: %v", evt.UserID, err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}

	}

	if tp, ok := evt.Task.(DirectEmailNotificationTemplateProvider); ok {
		if err := n.notifyDirectEmail(ctx, evt, tp); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("direct email notify: %v", err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}

	}

	if tp, ok := evt.Task.(TargetUsersNotificationProvider); ok {
		if err := n.notifyTargetUsers(ctx, evt, tp); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("notify target users: %v", err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}

	}

	if tp, ok := evt.Task.(SubscribersNotificationTemplateProvider); ok {
		if err := n.notifySubscribers(ctx, evt, tp); err != nil {
			if dlqErr := dlqRecordAndNotify(ctx, q, n, fmt.Sprintf("notify subscribers: %v", err)); dlqErr != nil {
				return dlqErr
			}
			return err
		}

	}

	if tp, ok := evt.Task.(AutoSubscribeProvider); ok {
		n.handleAutoSubscribe(ctx, evt, tp)

	}

	return nil
}

func (n *Notifier) notifySelf(ctx context.Context, evt eventbus.TaskEvent, tp SelfNotificationTemplateProvider) error {
	if et := tp.SelfEmailTemplate(); et != nil {
		if b, ok := evt.Task.(SelfEmailBroadcaster); ok && b.SelfEmailBroadcast() {
			emails, err := n.Queries.ListVerifiedEmailsByUserID(ctx, evt.UserID)
			if err == nil {
				for _, e := range emails {
					if err := n.renderAndQueueEmailFromTemplates(ctx, evt.UserID, e.Email, et, evt.Data, false); err != nil {
						return err
					}
				}
			}
		} else {
			ue, err := n.Queries.GetNotificationEmailByUserID(ctx, evt.UserID)
			if err != nil {
				notifyMissingEmail(ctx, n.Queries, evt.UserID)
			} else {
				if err := n.renderAndQueueEmailFromTemplates(ctx, evt.UserID, ue.Email, et, evt.Data, false); err != nil {
					return err
				}
			}
		}
	}
	if nt := tp.SelfInternalNotificationTemplate(); nt != nil {
		data := struct {
			eventbus.TaskEvent
			Item interface{}
		}{TaskEvent: evt, Item: evt.Data}
		msg, err := n.renderNotification(ctx, *nt, data)
		if err != nil {
			return err
		}
		if len(msg) > 0 {
			if err := n.Queries.InsertNotification(ctx, dbpkg.InsertNotificationParams{
				UsersIdusers: evt.UserID,
				Link:         sql.NullString{String: evt.Path, Valid: true},
				Message:      sql.NullString{String: string(msg), Valid: true},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Notifier) notifyDirectEmail(ctx context.Context, evt eventbus.TaskEvent, tp DirectEmailNotificationTemplateProvider) error {
	addr := tp.DirectEmailAddress(evt)
	if addr == "" {
		return nil
	}
	if et := tp.DirectEmailTemplate(); et != nil {
		if err := n.renderAndQueueEmailFromTemplates(ctx, 0, addr, et, evt.Data, true); err != nil {
			return err
		}
	}
	return nil
}

func (n *Notifier) notifyTargetUsers(ctx context.Context, evt eventbus.TaskEvent, tp TargetUsersNotificationProvider) error {
	for _, id := range tp.TargetUserIDs(evt) {
		user, err := n.Queries.GetUserById(ctx, id)
		if err != nil || !user.Email.Valid || user.Email.String == "" {
			notifyMissingEmail(ctx, n.Queries, id)
		} else {
			if et := tp.TargetEmailTemplate(); et != nil {
				if err := n.renderAndQueueEmailFromTemplates(ctx, id, user.Email.String, et, evt.Data, false); err != nil {
					return err
				}
			}
		}
		if nt := tp.TargetInternalNotificationTemplate(); nt != nil {
			data := struct {
				eventbus.TaskEvent
				Item interface{}
			}{TaskEvent: evt, Item: evt.Data}
			msg, err := n.renderNotification(ctx, *nt, data)
			if err != nil {
				return err
			}
			if len(msg) > 0 {
				if err := n.Queries.InsertNotification(ctx, dbpkg.InsertNotificationParams{
					UsersIdusers: id,
					Link:         sql.NullString{String: evt.Path, Valid: true},
					Message:      sql.NullString{String: string(msg), Valid: true},
				}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n *Notifier) notifySubscribers(ctx context.Context, evt eventbus.TaskEvent, tp SubscribersNotificationTemplateProvider) error {
	name := ""
	if tn, ok := evt.Task.(tasks.Name); ok {
		name = tn.Name()
	}
	patterns := buildPatterns(tasks.TaskString(name), evt.Path)

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

	if gp, ok := evt.Task.(GrantsRequiredProvider); ok {
		reqs := gp.GrantsRequired(evt)
		if len(reqs) != 0 {
			filterSubs := func(m map[int32]struct{}) {
				for id := range m {
					for _, g := range reqs {
						if _, err := n.Queries.CheckGrant(ctx, dbpkg.CheckGrantParams{
							ViewerID: id,
							Section:  g.Section,
							Item:     sql.NullString{String: g.Item, Valid: g.Item != ""},
							Action:   g.Action,
							ItemID:   sql.NullInt32{Int32: g.ItemID, Valid: g.ItemID != 0},
							UserID:   sql.NullInt32{Int32: id, Valid: id != 0},
						}); err != nil {
							delete(m, id)
							break
						}
					}
				}
			}
			filterSubs(emailSubs)
			filterSubs(internalSubs)
		}
	}

	var msg []byte
	data := struct {
		eventbus.TaskEvent
		Item interface{}
	}{TaskEvent: evt, Item: evt.Data}
	if nt := tp.SubscribedInternalNotificationTemplate(); nt != nil {
		var err error
		msg, err = n.renderNotification(ctx, *nt, data)
		if err != nil {
			log.Printf("render subscriber notification: %v", err)
			return fmt.Errorf("render notification: %w", err)
		}
	}

	et := tp.SubscribedEmailTemplate()
	for id := range emailSubs {
		if err := n.sendSubscriberEmail(ctx, id, evt, et); err != nil {
			return fmt.Errorf("deliver email to %d: %w", id, err)
		}
	}

	if len(msg) != 0 {
		for id := range internalSubs {
			if err := sendInternalNotification(ctx, n.Queries, id, evt.Path, string(msg)); err != nil {
				return fmt.Errorf("deliver internal to %d: %w", id, err)
			}
		}
	}

	return nil
}

func (n *Notifier) handleAutoSubscribe(ctx context.Context, evt eventbus.TaskEvent, tp AutoSubscribeProvider) {
	auto := true
	email := false
	if pref, err := n.Queries.GetPreferenceByUserID(ctx, evt.UserID); err == nil {
		auto = pref.AutoSubscribeReplies
		if pref.Emailforumupdates.Valid {
			email = pref.Emailforumupdates.Bool
		}
	}
	if auto {
		task, path := tp.AutoSubscribePath(evt)
		pattern := buildPatterns(tasks.TaskString(task), path)[0]
		if config.AppRuntimeConfig.NotificationsEnabled {
			ensureSubscription(ctx, n.Queries, evt.UserID, pattern, "internal")
		}
		if email {
			ensureSubscription(ctx, n.Queries, evt.UserID, pattern, "email")
		}
	}
}

func (n *Notifier) notifyAdmins(ctx context.Context, evt eventbus.TaskEvent, tp AdminEmailTemplateProvider) error {
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
			if err := n.renderAndQueueEmailFromTemplates(ctx, uid, addr, et, evt.Data, false); err != nil {
				return err
			}
		}
		if nt := tp.AdminInternalNotificationTemplate(); nt != nil && n.Queries != nil {
			data := struct {
				eventbus.TaskEvent
				Item interface{}
			}{TaskEvent: evt, Item: evt.Data}
			msg, err := n.renderNotification(ctx, *nt, data)
			if err != nil {
				return err
			}
			if err := n.Queries.InsertNotification(ctx, dbpkg.InsertNotificationParams{
				UsersIdusers: uid,
				Link:         sql.NullString{String: evt.Path, Valid: evt.Path != ""},
				Message:      sql.NullString{String: string(msg), Valid: len(msg) > 0},
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
