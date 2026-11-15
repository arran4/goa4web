package searchworker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/jobs"
)

// RunDurableWorker processes durable search_index jobs using search_dirty_targets revisions
// so concurrent edits never leave stale search tokens indexed.
func RunDurableWorker(ctx context.Context, jobSvc *jobs.Service, q db.Querier, workerID string) {
	if jobSvc == nil || q == nil {
		return
	}
	jobSvc.RunWorkerLoop(ctx, workerID, 0, []string{jobs.JobTypeSearchIndex}, func(ctx context.Context, job *db.DurableJob) error {
		if job == nil {
			return nil
		}
		targetType := job.TargetType
		targetID := int32(job.TargetID)
		if targetType == "" || targetID == 0 {
			return nil
		}
		dirty, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
			TargetType: targetType,
			TargetID:   targetID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		text, found, err := loadSearchableText(ctx, q, targetType, targetID)
		if err != nil {
			return err
		}
		if found && text != "" {
			if err := index(ctx, q, IndexEventData{Type: targetType, ID: targetID, Text: text}); err != nil {
				return err
			}
		}
		if _, err := q.SystemMarkSearchTargetIndexed(ctx, db.SystemMarkSearchTargetIndexedParams{
			IndexedRevision:   dirty.Revision,
			TargetType:        targetType,
			TargetID:          targetID,
			IndexedRevision_2: dirty.Revision,
		}); err != nil {
			return err
		}
		after, err := q.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
			TargetType: targetType,
			TargetID:   targetID,
		})
		if err == nil && after.Revision > dirty.Revision {
			_, _ = jobSvc.Enqueue(ctx, jobs.EnqueueRequest{
				JobType:        jobs.JobTypeSearchIndex,
				TargetType:     targetType,
				TargetID:       int64(targetID),
				TargetKey:      fmt.Sprintf("%s:%d", targetType, targetID),
				TargetRevision: after.Revision,
			})
		}
		return nil
	})
}

func loadSearchableText(ctx context.Context, q db.Querier, targetType string, targetID int32) (string, bool, error) {
	switch targetType {
	case TypeComment:
		row, err := q.SystemGetCommentTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		if err != nil {
			return "", false, err
		}
		return row.Text.String, true, nil
	case TypeWriting:
		row, err := q.SystemGetWritingTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		if err != nil {
			return "", false, err
		}
		return strings.TrimSpace(row.Title.String + "\n" + row.Abstract.String + "\n" + row.Writing.String), true, nil
	case TypeLinker:
		row, err := q.SystemGetLinkerTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		if err != nil {
			return "", false, err
		}
		return strings.TrimSpace(row.Title.String + "\n" + row.Url.String + "\n" + row.Description.String), true, nil
	case TypeImage:
		row, err := q.SystemGetImagePostTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		if err != nil {
			return "", false, err
		}
		return row.Description.String, true, nil
	default:
		return "", false, nil
	}
}

// EventKey is the map key used for search index events.
const EventKey = "search_index"

// Index types handled by the worker.
const (
	TypeComment = "comment"
	TypeWriting = "writing"
	TypeLinker  = "linker"
	TypeImage   = "image"
)

// IndexEventData describes content to index.
type IndexEventData struct {
	Type string
	ID   int32
	Text string
}

// Worker listens for index events and updates the search tables.
func Worker(ctx context.Context, bus *eventbus.Bus, q db.Querier) {
	if bus == nil || q == nil {
		return
	}
	ch := bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case env, ok := <-ch:
			if !ok {
				return
			}
			evt, ok := env.Msg.(eventbus.TaskEvent)
			if !ok {
				env.Ack()
				continue
			}
			// The server keeps the worker context alive while accepted events drain,
			// then cancels it at the shutdown deadline.
			evtCtx := ctx
			if data, ok := evt.Data[EventKey].(IndexEventData); ok {
				if err := index(evtCtx, q, data); err != nil {
					log.Printf("index error: %v", err)
				}
			} else {
				processEvent(evtCtx, evt, q)
			}
			env.Ack()
		case <-ctx.Done():
			return
		}
	}
}

func index(ctx context.Context, q db.Querier, data IndexEventData) error {
	counts := map[string]int32{}
	for _, w := range BreakupTextToWords(data.Text) {
		counts[strings.ToLower(w)]++
	}
	for word, count := range counts {
		id, err := q.SystemCreateSearchWord(ctx, strings.ToLower(word))
		if err != nil {
			return err
		}
		switch data.Type {
		case TypeComment:
			if err := q.SystemAddToForumCommentSearch(ctx, db.SystemAddToForumCommentSearchParams{CommentID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeWriting:
			if err := q.SystemAddToForumWritingSearch(ctx, db.SystemAddToForumWritingSearchParams{WritingID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeLinker:
			if err := q.SystemAddToLinkerSearch(ctx, db.SystemAddToLinkerSearchParams{LinkerID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		case TypeImage:
			if err := q.SystemAddToImagePostSearch(ctx, db.SystemAddToImagePostSearchParams{ImagePostID: data.ID, SearchwordlistIdsearchwordlist: int32(id), WordCount: count}); err != nil {
				return err
			}
		}
	}
	switch data.Type {
	case TypeComment:
		return q.SystemSetCommentLastIndex(ctx, data.ID)
	case TypeWriting:
		return q.SystemSetWritingLastIndex(ctx, data.ID)
	case TypeLinker:
		return q.SystemSetLinkerLastIndex(ctx, data.ID)
	case TypeImage:
		return q.SystemSetImagePostLastIndex(ctx, data.ID)
	}
	return nil
}

// Index synchronously updates the search index for one item.
func Index(ctx context.Context, q db.Querier, data IndexEventData) error {
	return index(ctx, q, data)
}
