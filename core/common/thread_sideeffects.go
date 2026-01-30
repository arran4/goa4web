package common

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
)

// ThreadUpdatedEvent captures side effects tied to thread updates.
type ThreadUpdatedEvent struct {
	Event *eventbus.TaskEvent

	ThreadID   int32
	TopicID    int32
	CommentID  int32
	Thread     any
	TopicTitle string
	Username   string
	Author     string

	LabelItem   string
	LabelItemID int32

	CommentText string
	CommentURL  string
	PostURL     string
	ThreadURL   string

	ClearUnreadForOthers bool
	MarkThreadRead       bool
	IncludePostCount     bool
	IncludeSearch        bool

	AdditionalData map[string]any
}

// HandleThreadUpdated applies notification, email, and thread-update side effects.
func (cd *CoreData) HandleThreadUpdated(ctx context.Context, event ThreadUpdatedEvent) error {
	var errs []error
	if cd != nil {
		if event.ClearUnreadForOthers {
			if err := cd.ClearThreadUnreadForOthers(event.ThreadID); err != nil {
				errs = append(errs, fmt.Errorf("clear unread labels: %w", err))
			}
			if event.LabelItem != "" && event.LabelItemID != 0 {
				if err := cd.ClearUnreadForOthers(event.LabelItem, event.LabelItemID); err != nil {
					errs = append(errs, fmt.Errorf("clear item unread labels: %w", err))
				}
			}
		}
		if event.MarkThreadRead {
			if err := cd.SetThreadReadMarker(event.ThreadID, event.CommentID); err != nil {
				errs = append(errs, fmt.Errorf("set read marker: %w", err))
			}
			// When marking a thread as read, we also want to ensure the "new" and "unread" labels
			// are cleared for the current user (author/viewer).
			// Passing false, false means: Not New, Not Unread.
			if err := cd.SetThreadPrivateLabelStatus(event.ThreadID, false, false); err != nil {
				errs = append(errs, fmt.Errorf("set private label status: %w", err))
			}
			if event.LabelItem != "" && event.LabelItemID != 0 {
				if err := cd.SetPrivateLabelStatus(event.LabelItem, event.LabelItemID, false, false); err != nil {
					errs = append(errs, fmt.Errorf("set item private label status: %w", err))
				}
			}
		}
	}

	evt := event.Event
	if evt == nil && cd != nil {
		evt = cd.Event()
	}
	if evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		if _, ok := evt.Data["Time"]; !ok {
			evt.Data["Time"] = time.Now().UTC()
		}
		if event.TopicTitle != "" {
			evt.Data["TopicTitle"] = event.TopicTitle
		} else if event.TopicID != 0 && cd != nil && cd.queries != nil {
			if topic, err := cd.queries.GetForumTopicById(ctx, event.TopicID); err == nil && topic != nil && topic.Title.Valid {
				evt.Data["TopicTitle"] = topic.Title.String
			}
		}
		evt.Data["Username"] = event.Username
		evt.Data["Author"] = event.Author
		if event.Thread != nil {
			evt.Data["Thread"] = event.Thread
		} else if event.ThreadID != 0 && cd != nil && cd.queries != nil {
			thread, err := cd.queries.GetThreadLastPosterAndPerms(ctx, db.GetThreadLastPosterAndPermsParams{
				ViewerID:      cd.UserID,
				ThreadID:      event.ThreadID,
				ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			})
			if err == nil && thread != nil {
				evt.Data["Thread"] = thread
			}
		}
		if event.ThreadID != 0 {
			evt.Data["ThreadID"] = event.ThreadID
		}
		if event.CommentURL != "" {
			evt.Data["URL"] = event.CommentURL
		} else if event.PostURL != "" {
			evt.Data["URL"] = event.PostURL
		} else if event.ThreadURL != "" {
			evt.Data["URL"] = event.ThreadURL
		}
		if event.CommentURL != "" {
			evt.Data["CommentURL"] = event.CommentURL
		}
		evt.Data["Body"] = event.CommentText
		if event.PostURL != "" {
			evt.Data["PostURL"] = event.PostURL
		}
		if event.ThreadURL != "" {
			evt.Data["ThreadURL"] = event.ThreadURL
		}
		for key, value := range event.AdditionalData {
			evt.Data[key] = value
		}
		if event.IncludePostCount {
			evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{
				CommentID: event.CommentID,
				ThreadID:  event.ThreadID,
				TopicID:   event.TopicID,
			}
		}
		if event.IncludeSearch {
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{
				Type: searchworker.TypeComment,
				ID:   event.CommentID,
				Text: event.CommentText,
			}
		}
	}

	return errors.Join(errs...)
}
