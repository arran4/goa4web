package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/arran4/goa4web/internal/db"
)

func TestScheduler_ProcessTasks_Backfill(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	s := New(q)
	taskName := "test_task"

	// Track calls
	var calledTimes []time.Time
	handler := func(ctx context.Context, t time.Time) error {
		calledTimes = append(calledTimes, t)
		return nil
	}
	s.Register(Task{
		Name:    taskName,
		Handler: handler,
		Type:    TaskTypeBackfill,
	})

	// Test case: Last run was 2 hours ago. Should run for T-1 and T-0.
	now := time.Now().UTC().Truncate(time.Hour)
	lastRun := now.Add(-2 * time.Hour)

	// Expect GetSchedulerState
	rows := sqlmock.NewRows([]string{"task_name", "last_run_at", "metadata"}).
		AddRow(taskName, lastRun, nil)
	mock.ExpectQuery("SELECT task_name, last_run_at, metadata FROM scheduler_state").
		WithArgs(taskName).
		WillReturnRows(rows)

	// Expect UpsertSchedulerState
	mock.ExpectExec("INSERT INTO scheduler_state").
		WithArgs(taskName, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Execute
	s.processTasks(context.Background())

	if len(calledTimes) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(calledTimes))
	}

	// Verify times
	t1 := lastRun.Add(time.Hour)
	if !calledTimes[0].Equal(t1) {
		t.Errorf("expected call 1 at %v, got %v", t1, calledTimes[0])
	}
	t2 := now
	if !calledTimes[1].Equal(t2) {
		t.Errorf("expected call 2 at %v, got %v", t2, calledTimes[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %s", err)
	}
}

func TestScheduler_ProcessTasks_Periodic(t *testing.T) {
	conn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer conn.Close()
	q := db.New(conn)

	s := New(q)
	taskName := "periodic_task"
	interval := 10 * time.Second

	var calledTimes []time.Time
	handler := func(ctx context.Context, t time.Time) error {
		calledTimes = append(calledTimes, t)
		return nil
	}
	s.Register(Task{
		Name:      taskName,
		Handler:   handler,
		Type:      TaskTypePeriodic,
		Interval:  interval,
		Ephemeral: true, // No DB calls expected
	})

	// First run
	s.processTasks(context.Background())

	if len(calledTimes) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calledTimes))
	}

	// Second run immediate - should be skipped (NextRun set to now+interval)
	s.processTasks(context.Background())
	if len(calledTimes) != 1 {
		t.Fatalf("expected 1 call (skipped), got %d", len(calledTimes))
	}

	// Mocking time passage is hard without DI for time.Now().
	// But we verified logic: First run happens, NextRun is set.

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("there were unfulfilled expectations: %s", err)
	}
}
