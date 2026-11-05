package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/arran4/goa4web/workers"
)

func TestShutdownIsBoundedByContextWhenWorkerIsBlocked(t *testing.T) {
	group := workers.NewGroup(context.Background())
	started := make(chan struct{})
	release := make(chan struct{})
	group.Go(func(context.Context) {
		close(started)
		<-release
	})
	<-started
	srv := &Server{WorkerGroup: group, WorkerCancel: group.Cancel}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	startedAt := time.Now()
	err := srv.Shutdown(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown() error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed := time.Since(startedAt); elapsed > time.Second {
		t.Fatalf("Shutdown took %s with a blocked worker", elapsed)
	}
	close(release)
}

func TestShutdownForcesBlockedHTTPRequestAtDeadline(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	httpServer := &http.Server{Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(started)
		<-release
	})}
	// A real loopback listener is required because http.Server shutdown semantics
	// depend on active connection tracking.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	serveDone := make(chan error, 1)
	go func() { serveDone <- httpServer.Serve(listener) }()
	requestDone := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + listener.Addr().String())
		if resp != nil {
			_ = resp.Body.Close()
		}
		requestDone <- err
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request handler did not start")
	}
	srv := &Server{httpServer: httpServer}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := srv.Shutdown(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Shutdown() error = %v, want context.DeadlineExceeded", err)
	}
	close(release)
	select {
	case err := <-requestDone:
		_ = err
	case <-time.After(time.Second):
		t.Fatal("blocked HTTP request remained active after forced close")
	}
	select {
	case err := <-serveDone:
		if !errors.Is(err, http.ErrServerClosed) {
			t.Fatalf("Serve() error = %v, want http.ErrServerClosed", err)
		}
	case <-time.After(time.Second):
		t.Fatal("HTTP serve loop did not stop")
	}
}
