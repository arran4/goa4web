package forum

import (
	"context"
	"errors"
	"testing"
)

type postUpdateQuerierStub struct {
	threadID int32
	topicID  int32
	err      error
}

func (q *postUpdateQuerierStub) AdminRecalculateForumThreadByIdMetaData(_ context.Context, threadID int32) error {
	q.threadID = threadID
	return q.err
}

func (q *postUpdateQuerierStub) SystemRebuildForumTopicMetaByID(_ context.Context, topicID int32) error {
	q.topicID = topicID
	return q.err
}

func TestApplyPostUpdate(t *testing.T) {
	t.Run("Happy Path", func(t *testing.T) {
		q := &postUpdateQuerierStub{}
		if err := ApplyPostUpdate(context.Background(), q, 10, 20); err != nil {
			t.Fatalf("ApplyPostUpdate: %v", err)
		}
		if q.threadID != 10 || q.topicID != 20 {
			t.Fatalf("calls = thread %d topic %d, want thread 10 topic 20", q.threadID, q.topicID)
		}
	})

	t.Run("Unhappy Path - thread update error", func(t *testing.T) {
		want := errors.New("thread metadata")
		q := &postUpdateQuerierStub{err: want}
		if err := ApplyPostUpdate(context.Background(), q, 10, 20); !errors.Is(err, want) {
			t.Fatalf("ApplyPostUpdate error = %v, want wrapped %v", err, want)
		}
	})
}
