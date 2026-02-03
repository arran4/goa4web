package scheduler

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

type Handler func(ctx context.Context, t time.Time) error

type TaskType int

const (
	TaskTypeBackfill TaskType = iota
	TaskTypePeriodic
)

type Task struct {
	Name      string
	Handler   Handler
	Type      TaskType
	Interval  time.Duration
	Ephemeral bool
}

type runtimeTask struct {
	Task
	LastRun time.Time
	NextRun time.Time
}

type Scheduler struct {
	Queries db.Querier
	tasks   []*runtimeTask
}

func New(q db.Querier) *Scheduler {
	return &Scheduler{
		Queries: q,
	}
}

func (s *Scheduler) Register(task Task) {
	s.tasks = append(s.tasks, &runtimeTask{
		Task: task,
	})
}

// Run starts the scheduler loop.
// tickRate determines how often the scheduler checks for tasks to run.
func (s *Scheduler) Run(ctx context.Context, tickRate time.Duration) {
	if s.Queries == nil {
		return
	}

	ticker := time.NewTicker(tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processTasks(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) processTasks(ctx context.Context) {
	now := time.Now().UTC()
	for _, task := range s.tasks {
		if now.Before(task.NextRun) {
			continue
		}
		s.processTask(ctx, task, now)
	}
}

func (s *Scheduler) processTask(ctx context.Context, rt *runtimeTask, now time.Time) {
	switch rt.Type {
	case TaskTypeBackfill:
		s.processBackfillTask(ctx, rt, now)
	case TaskTypePeriodic:
		s.processPeriodicTask(ctx, rt, now)
	}
}

func (s *Scheduler) processBackfillTask(ctx context.Context, rt *runtimeTask, now time.Time) {
	// Backfill tasks are checked at most once per minute to avoid DB spam,
	// unless we are way behind?
	// The original logic checked roughly every 30 mins (interval of Run).
	// Let's stick to 1 minute check interval for DB state.
	rt.NextRun = now.Add(1 * time.Minute)

	state, err := s.Queries.GetSchedulerState(ctx, rt.Name)
	var lastRun time.Time
	if err != nil {
		if err == sql.ErrNoRows {
			lastRun = now.Add(-1 * time.Hour)
		} else {
			log.Printf("Scheduler GetSchedulerState(%s) error: %v", rt.Name, err)
			return
		}
	} else {
		if state.LastRunAt.Valid {
			lastRun = state.LastRunAt.Time
		} else {
			lastRun = now.Add(-1 * time.Hour)
		}
	}

	currentHour := now.Truncate(time.Hour)
	lastRunTrunc := lastRun.Truncate(time.Hour)

	if !currentHour.After(lastRunTrunc) {
		return
	}

	// Loop from lastRun + 1 hour to currentHour
	for t := lastRunTrunc.Add(time.Hour); !t.After(currentHour); t = t.Add(time.Hour) {
		if err := rt.Handler(ctx, t); err != nil {
			log.Printf("Scheduler task %s failed for %v: %v", rt.Name, t, err)
			return
		}
	}

	err = s.Queries.UpsertSchedulerState(ctx, db.UpsertSchedulerStateParams{
		TaskName:  rt.Name,
		LastRunAt: sql.NullTime{Time: currentHour, Valid: true},
	})
	if err != nil {
		log.Printf("Scheduler UpsertSchedulerState(%s) error: %v", rt.Name, err)
	}
}

func (s *Scheduler) processPeriodicTask(ctx context.Context, rt *runtimeTask, now time.Time) {
	// For periodic tasks, we just run them if it's time.
	// If it's persistent, we might want to check DB, but optimizing for now:
	// We rely on in-memory LastRun for scheduling.
	// If it's the first run (LastRun is zero), we might want to check DB for persistent tasks.

	if rt.LastRun.IsZero() && !rt.Ephemeral {
		state, err := s.Queries.GetSchedulerState(ctx, rt.Name)
		if err == nil && state.LastRunAt.Valid {
			rt.LastRun = state.LastRunAt.Time
			// If we recovered LastRun from DB, check if we need to wait
			if now.Before(rt.LastRun.Add(rt.Interval)) {
				rt.NextRun = rt.LastRun.Add(rt.Interval)
				return
			}
		}
	}

	if err := rt.Handler(ctx, now); err != nil {
		log.Printf("Scheduler periodic task %s failed: %v", rt.Name, err)
		// We still update LastRun/NextRun so we don't retry immediately?
		// Or do we retry?
		// Original EmailQueueWorker waits 'delay' regardless of success/failure.
		// So we proceed.
	}

	rt.LastRun = now
	rt.NextRun = now.Add(rt.Interval)

	if !rt.Ephemeral {
		err := s.Queries.UpsertSchedulerState(ctx, db.UpsertSchedulerStateParams{
			TaskName:  rt.Name,
			LastRunAt: sql.NullTime{Time: now, Valid: true},
		})
		if err != nil {
			log.Printf("Scheduler UpsertSchedulerState(%s) error: %v", rt.Name, err)
		}
	}
}
