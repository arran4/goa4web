package common

import (
	"context"
	"errors"
	"fmt"

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
	_ = ctx

	var errs []error
	if cd != nil {
		if event.ClearUnreadForOthers {
			if err := cd.ClearThreadUnreadForOthers(event.ThreadID); err != nil {
				errs = append(errs, fmt.Errorf("clear unread labels: %w", err))
			}
		}
		if event.MarkThreadRead {
			if err := cd.SetThreadReadMarker(event.ThreadID, event.CommentID); err != nil {
				errs = append(errs, fmt.Errorf("set read marker: %w", err))
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
		if event.TopicTitle != "" {
			evt.Data["TopicTitle"] = event.TopicTitle
		}
		evt.Data["Username"] = event.Username
		evt.Data["Author"] = event.Author
		if event.Thread != nil {
			evt.Data["Thread"] = event.Thread
		}
		if event.ThreadID != 0 {
			evt.Data["ThreadID"] = event.ThreadID
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
