//go:build sqlite

package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/jobs"
	"github.com/arran4/goa4web/workers"
)

const (
	// workerSubprocessEnv points the subprocess test helper at the shared SQLite file.
	workerSubprocessEnv = "GOA4WEB_WORKER_TEST_HELPER"
	// workerSubprocessModeEnv selects subprocess failure or concurrency behavior.
	workerSubprocessModeEnv = "GOA4WEB_WORKER_TEST_MODE"
	// workerSubprocessLeaseMsEnv overrides the job claim lease duration in milliseconds.
	workerSubprocessLeaseMsEnv = "GOA4WEB_WORKER_TEST_LEASE_MS"
	// workerSubprocessEventsEnv sets a log path where workers append execution markers.
	workerSubprocessEventsEnv = "GOA4WEB_WORKER_TEST_EVENTS"
)

func openWALSQLiteWorkerDB(dbPath string) (*sql.DB, error) {
	sdb, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	sdb.SetMaxOpenConns(1)
	return sdb, nil
}

func appendWorkerExecEvent(eventsPath, targetKey, owner string) error {
	if eventsPath == "" {
		return nil
	}
	f, err := os.OpenFile(eventsPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = fmt.Fprintf(f, "%s\t%s\n", targetKey, owner)
	return err
}

func TestWorkerSubprocessHelper(t *testing.T) {
	dbPath := os.Getenv(workerSubprocessEnv)
	if dbPath == "" {
		return
	}
	mode := os.Getenv(workerSubprocessModeEnv)
	eventsPath := os.Getenv(workerSubprocessEventsEnv)
	sdb, err := openWALSQLiteWorkerDB(dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = sdb.Close() }()

	q := db.NewForDriver(sdb, "sqlite")
	svc := jobs.NewConfiguredService(q)
	if msStr := os.Getenv(workerSubprocessLeaseMsEnv); msStr != "" {
		if ms, parseErr := strconv.Atoi(msStr); parseErr == nil && ms > 0 {
			svc.LeaseDuration = time.Duration(ms) * time.Millisecond
		}
	}

	ctx, stop := notifyShutdownContext(context.Background())
	defer stop()

	group := workers.NewGroup(ctx)
	group.Go(func(ctx context.Context) {
		svc.RunWorkerLoop(ctx, "subprocess-worker", 10*time.Millisecond, []string{jobs.JobTypeExternalLink}, func(ctx context.Context, job *db.DurableJob) error {
			switch mode {
			case "second_sigint":
				fmt.Println("claimed")
				<-ctx.Done()
				fmt.Println("cancel-ignored")
				select {}
			case "compete":
				if err := appendWorkerExecEvent(eventsPath, job.TargetKey, "subprocess"); err != nil {
					return err
				}
				time.Sleep(15 * time.Millisecond)
				return nil
			default:
				fmt.Println("claimed")
				<-ctx.Done()
				return ctx.Err()
			}
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

func startWorkerSubprocess(t *testing.T, dbPath, mode string, leaseMs int, extraEnv ...string) (*exec.Cmd, chan string, *bytes.Buffer) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	cmd := exec.Command(exe, "-test.run=^TestWorkerSubprocessHelper$")
	env := append(os.Environ(), workerSubprocessEnv+"="+dbPath)
	if mode != "" {
		env = append(env, workerSubprocessModeEnv+"="+mode)
	}
	if leaseMs > 0 {
		env = append(env, workerSubprocessLeaseMsEnv+"="+strconv.Itoa(leaseMs))
	}
	env = append(env, extraEnv...)
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("StdoutPipe: %v", err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("start worker subprocess: %v", err)
	}

	outCh := make(chan string, 32)
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
	return cmd, outCh, &stderr
}

func waitSubprocessLine(t *testing.T, outCh <-chan string, stderr *bytes.Buffer, want string) {
	t.Helper()
	deadline := time.After(5 * time.Second)
	for {
		select {
		case line, ok := <-outCh:
			if !ok {
				t.Fatalf("subprocess stdout closed waiting for %q (stderr: %s)", want, stderr.String())
			}
			if line == want {
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for %q (stderr: %s)", want, stderr.String())
		}
	}
}

func setupWorkerSubprocessDB(t *testing.T) (string, *sql.DB, db.Querier) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "worker_crash.sqlite")
	sdb, err := openWALSQLiteWorkerDB(dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = sdb.Close() })

	schemaSQL, err := os.ReadFile(filepath.Join("..", "..", "database", "schema.sqlite.sql"))
	if err != nil {
		t.Fatalf("read schema.sqlite.sql: %v", err)
	}
	if _, err := sdb.ExecContext(context.Background(), string(schemaSQL)); err != nil {
		t.Fatalf("apply schema.sqlite.sql: %v", err)
	}
	return dbPath, sdb, db.NewForDriver(sdb, "sqlite")
}

func reclaimAndCompleteAfterExpiry(t *testing.T, q db.Querier, jobID int64) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	replacement := jobs.NewConfiguredService(q)
	replacement.Now = func() time.Time { return time.Now().UTC().Add(5 * time.Second) }

	done := make(chan struct{})
	go func() {
		replacement.RunWorkerLoop(ctx, "replacement-worker", 10*time.Millisecond, []string{jobs.JobTypeExternalLink}, func(_ context.Context, j *db.DurableJob) error {
			if j.ID == jobID {
				close(done)
			}
			return nil
		})
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("replacement worker did not reclaim job %d", jobID)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		job, err := q.SystemGetDurableJobByID(context.Background(), jobID)
		if err == nil && job.Status == jobs.StatusCompleted && job.Attempts == 2 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	job, err := q.SystemGetDurableJobByID(context.Background(), jobID)
	t.Fatalf("reclaimed job state = (%#v, %v), want completed with attempts=2", job, err)
}

func TestWorkerSubprocessCrashRecoveryAndGracefulInterrupt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process interrupt semantics are Unix-specific")
	}
	dbPath, _, q := setupWorkerSubprocessDB(t)
	svc := jobs.NewConfiguredService(q)
	if _, err := svc.Enqueue(context.Background(), jobs.EnqueueRequest{
		JobType:    jobs.JobTypeExternalLink,
		TargetType: jobs.JobTypeExternalLink,
		TargetKey:  "https://example.com/interrupt",
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	cmd, outCh, stderr := startWorkerSubprocess(t, dbPath, "", 0)
	waitSubprocessLine(t, outCh, stderr, "ready")
	waitSubprocessLine(t, outCh, stderr, "claimed")

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send SIGINT: %v", err)
	}
	waitSubprocessLine(t, outCh, stderr, "exited")
	if err := cmd.Wait(); err != nil {
		t.Fatalf("worker subprocess wait: %v (stderr: %s)", err, stderr.String())
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

func TestWorkerSubprocessSIGKILLCrashRecovery(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process kill semantics are Unix-specific")
	}
	dbPath, _, q := setupWorkerSubprocessDB(t)
	svc := jobs.NewConfiguredService(q)
	if _, err := svc.Enqueue(context.Background(), jobs.EnqueueRequest{
		JobType:    jobs.JobTypeExternalLink,
		TargetType: jobs.JobTypeExternalLink,
		TargetKey:  "https://example.com/sigkill",
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	cmd, outCh, stderr := startWorkerSubprocess(t, dbPath, "kill", 300)
	waitSubprocessLine(t, outCh, stderr, "ready")
	waitSubprocessLine(t, outCh, stderr, "claimed")

	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("SIGKILL worker subprocess: %v", err)
	}
	_ = cmd.Wait()

	job, err := q.SystemGetDurableJobByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("SystemGetDurableJobByID: %v", err)
	}
	if job.Status != jobs.StatusRunning || job.Attempts != 1 {
		t.Fatalf("after SIGKILL: status=%q attempts=%d, want running/1", job.Status, job.Attempts)
	}

	reclaimAndCompleteAfterExpiry(t, q, 1)
}

func TestWorkerSubprocessSecondSIGINTForcedTermination(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process interrupt semantics are Unix-specific")
	}
	dbPath, _, q := setupWorkerSubprocessDB(t)
	svc := jobs.NewConfiguredService(q)
	if _, err := svc.Enqueue(context.Background(), jobs.EnqueueRequest{
		JobType:    jobs.JobTypeExternalLink,
		TargetType: jobs.JobTypeExternalLink,
		TargetKey:  "https://example.com/second-sigint",
	}); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	cmd, outCh, stderr := startWorkerSubprocess(t, dbPath, "second_sigint", 300)
	waitSubprocessLine(t, outCh, stderr, "ready")
	waitSubprocessLine(t, outCh, stderr, "claimed")

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send 1st SIGINT: %v", err)
	}
	waitSubprocessLine(t, outCh, stderr, "cancel-ignored")

	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("send 2nd SIGINT: %v", err)
	}
	waitErr := cmd.Wait()
	if waitErr == nil {
		t.Fatal("expected forced exit error after 2nd SIGINT, got nil")
	}
	var exitErr *exec.ExitError
	if !errors.As(waitErr, &exitErr) {
		t.Fatalf("unexpected Wait error after 2nd SIGINT: %v", waitErr)
	}

	job, err := q.SystemGetDurableJobByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("SystemGetDurableJobByID: %v", err)
	}
	if job.Status != jobs.StatusRunning || job.Attempts != 1 {
		t.Fatalf("after 2nd SIGINT: status=%q attempts=%d, want running/1", job.Status, job.Attempts)
	}

	reclaimAndCompleteAfterExpiry(t, q, 1)
}

func TestWorkerSubprocessConcurrentEmbeddedAndStandaloneClaiming(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process interrupt semantics are Unix-specific")
	}
	dbPath, _, q := setupWorkerSubprocessDB(t)
	eventsPath := filepath.Join(t.TempDir(), "exec_events.log")
	svc := jobs.NewConfiguredService(q)

	const totalJobs = 10
	for i := 0; i < totalJobs; i++ {
		if _, err := svc.Enqueue(context.Background(), jobs.EnqueueRequest{
			JobType:    jobs.JobTypeExternalLink,
			TargetType: jobs.JobTypeExternalLink,
			TargetKey:  fmt.Sprintf("https://example.com/compete-%d", i),
		}); err != nil {
			t.Fatalf("Enqueue %d: %v", i, err)
		}
	}

	cmd, outCh, stderr := startWorkerSubprocess(t, dbPath, "compete", 5000, workerSubprocessEventsEnv+"="+eventsPath)
	waitSubprocessLine(t, outCh, stderr, "ready")

	embCtx, embCancel := context.WithCancel(context.Background())
	embDone := make(chan struct{})
	go func() {
		defer close(embDone)
		svc.RunWorkerLoop(embCtx, "embedded-worker", 10*time.Millisecond, []string{jobs.JobTypeExternalLink}, func(_ context.Context, job *db.DurableJob) error {
			if err := appendWorkerExecEvent(eventsPath, job.TargetKey, "embedded"); err != nil {
				return err
			}
			time.Sleep(15 * time.Millisecond)
			return nil
		})
	}()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		completed, err := svc.AdminList(context.Background(), jobs.StatusCompleted, jobs.JobTypeExternalLink, 50, 0)
		if err == nil && len(completed) == totalJobs {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	embCancel()
	<-embDone
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		t.Fatalf("interrupt compete subprocess: %v", err)
	}
	waitSubprocessLine(t, outCh, stderr, "exited")
	if err := cmd.Wait(); err != nil {
		t.Fatalf("compete subprocess wait: %v (stderr: %s)", err, stderr.String())
	}

	allJobs, err := svc.AdminList(context.Background(), "", jobs.JobTypeExternalLink, 50, 0)
	if err != nil || len(allJobs) != totalJobs {
		t.Fatalf("AdminList = (%d, %v), want %d", len(allJobs), err, totalJobs)
	}
	for _, j := range allJobs {
		if j.Status != jobs.StatusCompleted || j.Attempts != 1 {
			t.Fatalf("job %d (%s): status=%q attempts=%d, want completed/1", j.ID, j.TargetKey, j.Status, j.Attempts)
		}
	}

	raw, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("read events log: %v", err)
	}
	counts := make(map[string]int)
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		counts[parts[0]]++
	}
	if len(counts) != totalJobs {
		t.Fatalf("executed %d distinct jobs, want %d: %#v", len(counts), totalJobs, counts)
	}
	for target, c := range counts {
		if c != 1 {
			t.Fatalf("job %s executed %d times across workers, want 1", target, c)
		}
	}
}
