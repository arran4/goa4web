package scheduler

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

type Handler func(ctx context.Context, t time.Time) error

type Task struct {
	Name    string
	Handler Handler
}

type Scheduler struct {
	Queries db.Querier
	Tasks   []Task
}

func New(q db.Querier) *Scheduler {
	return &Scheduler{
		Queries: q,
	}
}

func (s *Scheduler) Register(name string, handler Handler) {
	s.Tasks = append(s.Tasks, Task{Name: name, Handler: handler})
}

// Run starts the scheduler loop.
func (s *Scheduler) Run(ctx context.Context, interval time.Duration) {
	if s.Queries == nil {
		return
	}

	ticker := time.NewTicker(interval)
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
	for _, task := range s.Tasks {
		s.processTask(ctx, task)
	}
}

func (s *Scheduler) processTask(ctx context.Context, task Task) {
	state, err := s.Queries.GetSchedulerState(ctx, task.Name)
	var lastRun time.Time
	if err != nil {
		if err == sql.ErrNoRows {
			lastRun = time.Now().UTC().Add(-1 * time.Hour)
		} else {
			log.Printf("Scheduler GetSchedulerState(%s) error: %v", task.Name, err)
			return
		}
	} else {
		if state.LastRunAt.Valid {
			lastRun = state.LastRunAt.Time
		} else {
			lastRun = time.Now().UTC().Add(-1 * time.Hour)
		}
	}

	now := time.Now().UTC()
	currentHour := now.Truncate(time.Hour)
	lastRunTrunc := lastRun.Truncate(time.Hour)

	if !currentHour.After(lastRunTrunc) {
		return
	}

	// Loop from lastRun + 1 hour to currentHour
	for t := lastRunTrunc.Add(time.Hour); !t.After(currentHour); t = t.Add(time.Hour) {
		if err := task.Handler(ctx, t); err != nil {
			log.Printf("Scheduler task %s failed for %v: %v", task.Name, t, err)
			// Decide on retry policy. For now, we log and continue,
			// effectively skipping this hour if we proceed to update state.
			// Ideally we shouldn't update state if failed, but partial failure in a loop is tricky.
			// Current implementation updates state at the end, so if one fails, we retry ALL next time?
			// No, let's return here so we don't update state, and retry this hour next tick.
			return
		}
	}

	err = s.Queries.UpsertSchedulerState(ctx, db.UpsertSchedulerStateParams{
		TaskName:  task.Name,
		LastRunAt: sql.NullTime{Time: currentHour, Valid: true},
	})
	if err != nil {
		log.Printf("Scheduler UpsertSchedulerState(%s) error: %v", task.Name, err)
	}
}
