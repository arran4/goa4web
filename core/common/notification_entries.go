package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// NotificationEntry is an authorized inbox row or a display summary of unread
// replies. Reply groups contain the IDs visible in that database snapshot.
type NotificationEntry struct {
	db.Notification
	ForumReply   bool
	TopicTitle   string
	ReplyAuthor  string
	ReplyCount   int64
	Snapshot     string
	TotalEntries int64
	memberIDs    []int32
}

// DisplayMessage renders a reply summary using only current metadata.
func (n NotificationEntry) DisplayMessage() string {
	if !n.ForumReply || n.ReadAt.Valid {
		return n.Message.String
	}
	reply := "replies"
	if n.ReplyCount == 1 {
		reply = "reply"
	}
	title := strings.TrimSpace(n.TopicTitle)
	if title == "" {
		title = "a forum thread"
	}
	return fmt.Sprintf("%d new %s in %s", n.ReplyCount, reply, title)
}

// VisibleNotificationEntries lists notifications within the current actor's
// grants. The query scopes account, resource visibility, and pagination together.
func (cd *CoreData) VisibleNotificationEntries(params db.ListVisibleNotificationEntriesForListerParams) ([]NotificationEntry, error) {
	if cd.queries == nil || cd.UserID <= 0 {
		return nil, nil
	}
	params.ListerID = cd.UserID
	params.Offset = max(0, params.Offset)
	rows, err := cd.queries.ListVisibleNotificationEntriesForLister(cd.ctx, params)
	if err != nil {
		return nil, err
	}
	entries := make([]NotificationEntry, 0, len(rows))
	for _, row := range rows {
		entry := NotificationEntry{
			Notification: db.Notification{
				ID: row.ID, UsersIdusers: row.UsersIdusers, Link: row.Link, Message: row.Message,
				ActivityKey: row.ActivityKey, Priority: row.Priority, ResourceType: row.ResourceType,
				ResourceID: row.ResourceID, ActivityPosition: row.ActivityPosition,
				CreatedAt: row.CreatedAt, ReadAt: row.ReadAt,
			},
			ForumReply: row.ForumReply != 0, TopicTitle: row.TopicTitle, ReplyAuthor: row.ReplyAuthor,
			ReplyCount: row.ReplyCount, TotalEntries: row.TotalEntries,
		}
		var snapshot string
		switch raw := row.Snapshot.(type) {
		case string:
			snapshot = raw
		case []byte:
			snapshot = string(raw)
		}
		if snapshot != "" {
			var ids []int32
			if err := json.Unmarshal([]byte(snapshot), &ids); err != nil || int64(len(ids)) != row.ReplyCount {
				return nil, fmt.Errorf("invalid notification group snapshot for notification %d", row.ID)
			}
			entry.Snapshot = snapshot
			entry.memberIDs = ids
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// InboxNotificationCount counts visible inbox entries, with unread replies
// contributing one entry per thread.
func (cd *CoreData) InboxNotificationCount(showAll bool) (int64, error) {
	entries, err := cd.VisibleNotificationEntries(db.ListVisibleNotificationEntriesForListerParams{
		GroupUnread: 1, IncludeRead: showAll, Limit: 1,
	})
	if err != nil || len(entries) == 0 {
		return 0, err
	}
	return entries[0].TotalEntries, nil
}

// Notifications lists paginated inbox entries for a page or live fragment.
func (cd *CoreData) Notifications(r *http.Request) ([]NotificationEntry, error) {
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 32)
	return cd.VisibleNotificationEntries(db.ListVisibleNotificationEntriesForListerParams{
		GroupUnread: 1, IncludeRead: r.URL.Query().Get("all") == "1",
		Limit: int32(cd.PageSize()), Offset: int32(offset),
	})
}

// VisibleNotification resolves an individual notification under current grants.
func (cd *CoreData) VisibleNotification(id int32) (*db.Notification, error) {
	if id <= 0 {
		return nil, sql.ErrNoRows
	}
	entries, err := cd.VisibleNotificationEntries(db.ListVisibleNotificationEntriesForListerParams{
		NotificationID: id, IncludeRead: true, Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, sql.ErrNoRows
	}
	return &entries[0].Notification, nil
}

// RenderedNotifications returns transport-ready messages. The stored activity
// and notification IDs are still carried separately for dismissal.
func RenderedNotifications(entries []NotificationEntry) []*db.Notification {
	result := make([]*db.Notification, 0, len(entries))
	for _, entry := range entries {
		note := entry.Notification
		note.Message = sql.NullString{String: entry.DisplayMessage(), Valid: true}
		result = append(result, &note)
	}
	return result
}

// NotificationSnapshotIDs preserves every displayed member, even when a
// transport renders a whole thread as one entry.
func NotificationSnapshotIDs(entries []NotificationEntry) []int32 {
	var ids []int32
	for _, entry := range entries {
		if len(entry.memberIDs) == 0 {
			ids = append(ids, entry.ID)
		} else {
			ids = append(ids, entry.memberIDs...)
		}
	}
	return ids
}
