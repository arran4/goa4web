package common

import (
	"database/sql"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/searchindex"
)

// AdminApproveQueuedLink publishes a queued link and records its indexing work.
func (cd *CoreData) AdminApproveQueuedLink(queueID int32) (int64, error) {
	return cd.withSearchWork(searchindex.TypeLinker, func(q db.Querier) (int64, error) {
		result, err := q.AdminInsertQueuedLinkFromQueue(cd.ctx, queueID)
		if err != nil {
			return 0, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		if rows == 0 {
			return 0, sql.ErrNoRows
		}
		return result.LastInsertId()
	})
}
