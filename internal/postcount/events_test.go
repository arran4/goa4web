package postcount

import (
	"context"
	"errors"
	"testing"
)

type postUpdateQuerierStub struct {
	threadID  int32
	topicID   int32
	threadErr error
	topicErr  error
}

func (q *postUpdateQuerierStub) AdminRecalculateForumThreadByIdMetaData(_ context.Context, threadID int32) error {
	q.threadID = threadID
	return q.threadErr
}

func (q *postUpdateQuerierStub) SystemRebuildForumTopicMetaByID(_ context.Context, topicID int32) error {
	q.topicID = topicID
	return q.topicErr
}

func TestPostCountEventContract(t *testing.T) {
	if PostCountEventKey != "post_count" {
		t.Fatalf("PostCountEventKey = %q, want post_count", PostCountEventKey)
	}
	u := PostCountUpdate{ThreadID: 10, TopicID: 20, CommentID: 30}
	if u.ThreadID != 10 || u.TopicID != 20 || u.CommentID != 30 {
		t.Fatalf("PostCountUpdate = %#v", u)
	}
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
		q := &postUpdateQuerierStub{threadErr: want}
		if err := ApplyPostUpdate(context.Background(), q, 10, 20); !errors.Is(err, want) {
			t.Fatalf("ApplyPostUpdate error = %v, want wrapped %v", err, want)
		}
		if q.topicID != 0 {
			t.Fatalf("topic rebuild should not run after thread recalc error; topicID = %d", q.topicID)
		}
	})

	t.Run("Unhappy Path - topic rebuild error", func(t *testing.T) {
		want := errors.New("topic metadata")
		q := &postUpdateQuerierStub{topicErr: want}
		if err := ApplyPostUpdate(context.Background(), q, 10, 20); !errors.Is(err, want) {
			t.Fatalf("ApplyPostUpdate error = %v, want wrapped %v", err, want)
		}
	})
}
