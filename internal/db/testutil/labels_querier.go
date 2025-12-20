package testutil

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

// LabelsQuerier implements label-related db.Querier methods for tests.
type LabelsQuerier struct {
	*BaseQuerier
	PublicLabels  []*db.ListContentPublicLabelsRow
	LabelStatus   []*db.ListContentLabelStatusRow
	PrivateLabels []*db.ListContentPrivateLabelsRow

	AddedPublic    []string
	RemovedPublic  []string
	AddedPrivate   []string
	RemovedPrivate []string
	AddedStatus    []string
	RemovedStatus  []string
}

// NewLabelsQuerier returns a labels querier stub.
func NewLabelsQuerier(t testing.TB) *LabelsQuerier {
	t.Helper()
	return &LabelsQuerier{
		BaseQuerier: NewBaseQuerier(t),
	}
}

func (q *LabelsQuerier) ListContentPublicLabels(ctx context.Context, arg db.ListContentPublicLabelsParams) ([]*db.ListContentPublicLabelsRow, error) {
	return q.PublicLabels, nil
}

func (q *LabelsQuerier) ListContentLabelStatus(ctx context.Context, arg db.ListContentLabelStatusParams) ([]*db.ListContentLabelStatusRow, error) {
	return q.LabelStatus, nil
}

func (q *LabelsQuerier) ListContentPrivateLabels(ctx context.Context, arg db.ListContentPrivateLabelsParams) ([]*db.ListContentPrivateLabelsRow, error) {
	return q.PrivateLabels, nil
}

func (q *LabelsQuerier) AddContentPublicLabel(ctx context.Context, arg db.AddContentPublicLabelParams) error {
	q.AddedPublic = append(q.AddedPublic, arg.Label)
	return nil
}

func (q *LabelsQuerier) RemoveContentPublicLabel(ctx context.Context, arg db.RemoveContentPublicLabelParams) error {
	q.RemovedPublic = append(q.RemovedPublic, arg.Label)
	return nil
}

func (q *LabelsQuerier) AddContentPrivateLabel(ctx context.Context, arg db.AddContentPrivateLabelParams) error {
	q.AddedPrivate = append(q.AddedPrivate, arg.Label)
	return nil
}

func (q *LabelsQuerier) RemoveContentPrivateLabel(ctx context.Context, arg db.RemoveContentPrivateLabelParams) error {
	q.RemovedPrivate = append(q.RemovedPrivate, arg.Label)
	return nil
}

func (q *LabelsQuerier) SystemClearContentPrivateLabel(ctx context.Context, arg db.SystemClearContentPrivateLabelParams) error {
	q.RemovedPrivate = append(q.RemovedPrivate, arg.Label)
	return nil
}

func (q *LabelsQuerier) AddContentLabelStatus(ctx context.Context, arg db.AddContentLabelStatusParams) error {
	q.AddedStatus = append(q.AddedStatus, arg.Label)
	return nil
}

func (q *LabelsQuerier) RemoveContentLabelStatus(ctx context.Context, arg db.RemoveContentLabelStatusParams) error {
	q.RemovedStatus = append(q.RemovedStatus, arg.Label)
	return nil
}
