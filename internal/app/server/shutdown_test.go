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
	requestDone := make(chan struct{})
	go func() {
		resp, err := http.Get("http://" + listener.Addr().String())
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
		close(requestDone)
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
	case <-requestDone:
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

func TestStartAfterShutdownDoesNotBind(t *testing.T) {
	srv := &Server{Router: http.NotFoundHandler()}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	// An invalid address makes any attempt to bind after shutdown observable
	// without creating a listener or racing a goroutine.
	if err := srv.Start("invalid-listen-address"); err != nil {
		t.Fatalf("start after shutdown attempted to listen: %v", err)
	}
	if addr := srv.Addr(); addr != "" {
		t.Fatalf("start after shutdown published address %q", addr)
	}
}

func TestStartAndShutdownCanOverlap(t *testing.T) {
	srv := &Server{Router: http.NotFoundHandler()}
	start := make(chan struct{})
	serveDone := make(chan error, 1)
	go func() {
		<-start
		// A real loopback listener exercises the handoff to http.Server.Serve.
		serveDone <- srv.Start("127.0.0.1:0")
	}()
	close(start)
	_ = srv.Addr()
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server started accepting after shutdown")
	}
}
