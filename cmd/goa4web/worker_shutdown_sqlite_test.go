//go:build sqlite

package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/workers"
)

const workerSubprocessEnv = "GOA4WEB_WORKER_TEST_HELPER"

func TestWorkerSubprocessHelper(t *testing.T) {
	dbPath := os.Getenv(workerSubprocessEnv)
	if dbPath == "" {
		return
	}
	sdb, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sdb.SetMaxOpenConns(1)
	defer func() { _ = sdb.Close() }()

	q := db.NewForDriver(sdb, "sqlite")
	svc := jobs.NewConfiguredService(q)

	ctx, stop := notifyShutdownContext(context.Background())
	defer stop()

	group := workers.NewGroup(ctx)
	group.Go(func(ctx context.Context) {
		svc.RunWorkerLoop(ctx, "subprocess-worker", 10*time.Millisecond, []string{jobs.JobTypeExternalLink}, func(ctx context.Context, _ *db.DurableJob) error {
			fmt.Println("claimed")
			<-ctx.Done()
			return ctx.Err()
		})
	})

	fmt.Println("ready")
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	group.Cancel()
	if err := group.Wait(shutdownCtx); err != nil {
		t.Fatalf("worker shutdown wait: %v", err)
	}
	fmt.Println("exited")
}

func TestWorkerSubprocessCrashRecoveryAndGracefulInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process interrupt semantics are Unix-specific")
	}
	dbPath := filepath.Join(t.TempDir(), "worker_crash.sqlite")
	sdb, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(5000)")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sdb.SetMaxOpenConns(1)
	defer func() { _ = sdb.Close() }()

	schemaSQL, err := os.ReadFile(filepath.Join("..", "..", "database", "schema.sqlite.sql"))
	if err != nil {
		t.Fatalf("read schema.sqlite.sql: %v", err)
	}
	if _, err := sdb.ExecContext(context.Background(), string(schemaSQL)); err != nil {
		t.Fatalf("apply schema.sqlite.sql: %v", err)
	}

	q := db.NewForDriver(sdb, "sqlite")
	svc := jobs.NewConfiguredService(q)
	if _, err := svc.Enqueue(context.Background(), jobs.EnqueueRequest{
		JobType:    jobs.JobTypeExternalLink,
		TargetType: jobs.JobTypeExternalLink,
		TargetKey:  "https://example.com/interrupt",
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	cmd := exec.Command(exe, "-test.run=^TestWorkerSubprocessHelper$")
	cmd.Env = append(os.Environ(), workerSubprocessEnv+"="+dbPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe: %v", err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start worker subprocess: %v", err)
	}

	outCh := make(chan string, 10)
	go func() {
		buf := make([]byte, 4096)
		var acc string
		for {
			n, readErr := stdout.Read(buf)
			if n > 0 {
				acc += string(buf[:n])
				for {
					idx := strings.IndexByte(acc, '\n')
					if idx < 0 {
						break
					}
					outCh <- strings.TrimSpace(acc[:idx])
					acc = acc[idx+1:]
				}
			}
			if readErr != nil {
				close(outCh)
				return
			}
		}
	}()

	waitOutput := func(want string) {
		t.Helper()
		deadline := time.After(5 * time.Second)
		for {
			select {
			case line, ok := <-outCh:
				if !ok {
					t.Fatalf("subprocess stdout closed waiting for %q", want)
				}
				if line == want {
					return
				}
			case <-deadline:
				t.Fatalf("timed out waiting for %q", want)
			}
		}
	}

	waitOutput("ready")
	waitOutput("claimed")

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send SIGINT: %v", err)
	}
	waitOutput("exited")
	if err := cmd.Wait(); err != nil {
		t.Fatalf("worker subprocess wait: %v", err)
	}

	// Verify the interrupted job was cleanly released back to pending without charging an attempt.
	job, err := q.SystemGetDurableJobByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("SystemGetDurableJobByID: %v", err)
	}
	if job.Status != jobs.StatusPending || job.Attempts != 0 {
		t.Fatalf("after graceful interrupt: status=%q attempts=%d, want pending/0", job.Status, job.Attempts)
	}
	_ = config.RuntimeConfig{}
}
