package workers

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGroupWaitIsBoundedWhenWorkIgnoresCancellation(t *testing.T) {
	group := NewGroup(context.Background())
	started := make(chan struct{})
	release := make(chan struct{})
	if !group.Go(func(context.Context) {
		close(started)
		<-release
	}) {
		t.Fatal("Group.Go rejected work before shutdown")
	}
	<-started
	group.Cancel()
	waitCtx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := group.Wait(waitCtx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Group.Wait() error = %v, want context.DeadlineExceeded", err)
	}
	close(release)
	finishCtx, finishCancel := context.WithTimeout(context.Background(), time.Second)
	defer finishCancel()
	if err := group.Wait(finishCtx); err != nil {
		t.Fatalf("Group.Wait() after release: %v", err)
	}
}
