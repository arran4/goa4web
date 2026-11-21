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
	"github.com/arran4/goa4web/internal/searchindex"
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
		transactor, ok := q.(db.JobTransactor)
		if !ok {
			return fmt.Errorf("durable search transactions unavailable")
		}
		return transactor.WithJobTransaction(ctx, func(tx db.Querier) error {
			// Take the source, revision and job locks in the same order as
			// the producer that writes the source and records its work.
			text, err := loadSearchableText(ctx, tx, targetType, targetID)
			if err != nil {
				return err
			}
			dirty, err := tx.SystemGetSearchDirtyTarget(ctx, db.SystemGetSearchDirtyTargetParams{
				TargetType: targetType, TargetID: targetID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			if err != nil {
				return err
			}
			if dirty.IndexedRevision == dirty.Revision {
				return nil
			}

			// Lock the revision being published. A concurrent producer that has
			// already advanced the target leaves this attempt with nothing to do.
			marked, err := tx.SystemMarkSearchTargetIndexed(ctx, db.SystemMarkSearchTargetIndexedParams{
				IndexedRevision: dirty.Revision, TargetType: targetType, TargetID: targetID,
			})
			if err != nil || marked == 0 {
				return err
			}
			if err := jobSvc.LockActiveLease(ctx, tx, job); err != nil {
				return err
			}
			return searchindex.Replace(ctx, tx, IndexEventData{Type: targetType, ID: targetID, Text: text})
		})
	})
}

func loadSearchableText(ctx context.Context, q db.Querier, targetType string, targetID int32) (string, error) {
	switch targetType {
	case TypeComment:
		row, err := q.SystemGetCommentTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return row.Text.String, nil
	case TypeWriting:
		row, err := q.SystemGetWritingTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(row.Title.String + "\n" + row.Abstract.String + "\n" + row.Writing.String), nil
	case TypeLinker:
		row, err := q.SystemGetLinkerTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(row.Title.String + "\n" + row.Url.String + "\n" + row.Description.String), nil
	case TypeImage:
		row, err := q.SystemGetImagePostTextForSearchIndex(ctx, targetID)
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return row.Description.String, nil
	default:
		return "", nil
	}
}

// EventKey is retained as a compatibility alias for the searchindex event key.
const EventKey = searchindex.IndexEventKey

// Index types are retained as compatibility aliases for the searchindex contract.
const (
	TypeComment = searchindex.TypeComment
	TypeWriting = searchindex.TypeWriting
	TypeLinker  = searchindex.TypeLinker
	TypeImage   = searchindex.TypeImage
)

// IndexEventData is retained as a compatibility alias for the searchindex contract.
type IndexEventData = searchindex.IndexEventData

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
	return searchindex.Index(ctx, q, data)
}

// Index synchronously updates the search index for one item.
func Index(ctx context.Context, q db.Querier, data IndexEventData) error {
	return index(ctx, q, data)
}
