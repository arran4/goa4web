package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// FAQQuerier implements FAQ-related queries for tests.
type FAQQuerier struct {
	*BaseQuerier
	FAQ            *db.Faq
	Updated        []db.AdminUpdateFAQQuestionAnswerParams
	Revisions      []db.InsertFAQRevisionForUserParams
	CreatedFAQArgs []db.CreateFAQQuestionForWriterParams
	DeletedIDs     []int32
}

// NewFAQQuerier returns a FAQ querier stub.
func NewFAQQuerier(t testing.TB) *FAQQuerier {
	t.Helper()
	return &FAQQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *FAQQuerier) AdminGetFAQByID(ctx context.Context, id int32) (*db.Faq, error) {
	return q.FAQ, nil
}

func (q *FAQQuerier) AdminDeleteFAQ(ctx context.Context, id int32) error {
	q.DeletedIDs = append(q.DeletedIDs, id)
	return nil
}

func (q *FAQQuerier) AdminUpdateFAQQuestionAnswer(ctx context.Context, arg db.AdminUpdateFAQQuestionAnswerParams) error {
	q.Updated = append(q.Updated, arg)
	return nil
}

func (q *FAQQuerier) InsertFAQRevisionForUser(ctx context.Context, arg db.InsertFAQRevisionForUserParams) error {
	q.Revisions = append(q.Revisions, arg)
	return nil
}

func (q *FAQQuerier) CreateFAQQuestionForWriter(ctx context.Context, arg db.CreateFAQQuestionForWriterParams) error {
	q.CreatedFAQArgs = append(q.CreatedFAQArgs, arg)
	return nil
}
