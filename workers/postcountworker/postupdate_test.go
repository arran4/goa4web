package postcountworker

import (
	"context"
	"testing"
)

type stubPostUpdateQuerier struct {
	threadCalls []int32
	topicCalls  []int32
	threadErr   error
	topicErr    error
}

func (s *stubPostUpdateQuerier) AdminRecalculateForumThreadByIdMetaData(_ context.Context, idforumthread int32) error {
	s.threadCalls = append(s.threadCalls, idforumthread)
	return s.threadErr
}

func (s *stubPostUpdateQuerier) SystemRebuildForumTopicMetaByID(_ context.Context, idforumtopic int32) error {
	s.topicCalls = append(s.topicCalls, idforumtopic)
	return s.topicErr
}

func TestPostUpdate(t *testing.T) {
	q := &stubPostUpdateQuerier{}

	if err := PostUpdate(context.Background(), q, 1, 2); err != nil {
		t.Fatalf("PostUpdate: %v", err)
	}

	if got := q.threadCalls; len(got) != 1 || got[0] != 1 {
		t.Fatalf("AdminRecalculateForumThreadByIdMetaData called with %v, want [1]", got)
	}
	if got := q.topicCalls; len(got) != 1 || got[0] != 2 {
		t.Fatalf("SystemRebuildForumTopicMetaByID called with %v, want [2]", got)
	}
}
