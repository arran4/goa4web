package scenario

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
)

// Clock provides current time abstraction.
type Clock interface {
	Now() time.Time
}

// RealClock uses system time.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// FixedClock provides a fixed time for testing.
type FixedClock struct {
	T time.Time
}

func (f FixedClock) Now() time.Time { return f.T }

// ApplyResult contains the outcome of applying a scenario.
type ApplyResult struct {
	ScenarioName  string
	EventsApplied int
	Registry      *RefRegistry
}

// Runner applies validated scenario events to a goa4web database instance.
type Runner struct {
	queries  db.Querier
	clock    Clock
	registry *RefRegistry
}

// Option configures a Runner.
type Option func(*Runner)

// WithClock sets a custom Clock on the Runner.
func WithClock(c Clock) Option {
	return func(r *Runner) {
		r.clock = c
	}
}

// WithRefRegistry sets a custom RefRegistry on the Runner.
func WithRefRegistry(reg *RefRegistry) Option {
	return func(r *Runner) {
		r.registry = reg
	}
}

// NewRunner creates a new Runner with the provided Querier and options.
func NewRunner(q db.Querier, opts ...Option) *Runner {
	r := &Runner{
		queries:  q,
		clock:    RealClock{},
		registry: NewRefRegistry(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Registry returns the runner's RefRegistry.
func (r *Runner) Registry() *RefRegistry {
	return r.registry
}

// Apply validates and executes all events in a scenario in order.
func (r *Runner) Apply(ctx context.Context, s *Scenario) (*ApplyResult, error) {
	if err := Validate(s); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	applied := 0
	for _, evt := range s.Events {
		if err := r.applyEvent(ctx, evt); err != nil {
			return nil, fmt.Errorf("apply event %s (%s): %w", evt.File, evt.Op, err)
		}
		applied++
	}

	return &ApplyResult{
		ScenarioName:  s.Meta.Name,
		EventsApplied: applied,
		Registry:      r.registry,
	}, nil
}

func (r *Runner) applyEvent(ctx context.Context, evt *Event) error {
	switch evt.Op {
	case "user.create":
		data, ok := evt.OpData.(*UserCreateData)
		if !ok {
			return fmt.Errorf("invalid operation data for user.create")
		}
		return r.applyUserCreate(ctx, data)

	case "user.enable":
		data, ok := evt.OpData.(*UserEnableData)
		if !ok {
			return fmt.Errorf("invalid operation data for user.enable")
		}
		return r.applyUserEnable(ctx, data)

	default:
		return fmt.Errorf("operation %q apply handler not yet implemented in this version", evt.Op)
	}
}

func (r *Runner) applyUserCreate(ctx context.Context, data *UserCreateData) error {
	// Generate or hash password
	pw := data.Password
	if pw == "" {
		// Provide default scenario password if unspecified
		pw = "scenario-default-password"
	}
	hash, alg, err := auth.HashPassword(pw)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	// Insert user
	id, err := r.queries.SystemInsertUser(ctx, sql.NullString{String: data.Username, Valid: true})
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("user %q already exists: %w", data.Username, err)
		}
		return fmt.Errorf("insert user: %w", err)
	}
	uid := int32(id)

	// Insert email if specified
	if data.Email != "" {
		verifiedAt := sql.NullTime{Time: data.At, Valid: true}
		if err := r.queries.InsertUserEmail(ctx, db.InsertUserEmailParams{
			UserID:               uid,
			Email:                data.Email,
			VerifiedAt:           verifiedAt,
			LastVerificationCode: sql.NullString{},
			NotificationPriority: 100,
		}); err != nil {
			return fmt.Errorf("insert user email: %w", err)
		}
	}

	// Insert password
	if err := r.queries.InsertPassword(ctx, db.InsertPasswordParams{
		UsersIdusers:    uid,
		Passwd:          hash,
		PasswdAlgorithm: sql.NullString{String: alg, Valid: true},
	}); err != nil {
		return fmt.Errorf("insert password: %w", err)
	}

	// Bind reference if declared
	if data.Ref != "" {
		if err := r.registry.Bind(RefTypeUser, data.Ref, uid); err != nil {
			return fmt.Errorf("bind user ref %q: %w", data.Ref, err)
		}
	}

	return nil
}

func (r *Runner) applyUserEnable(ctx context.Context, data *UserEnableData) error {
	uid, ok := r.registry.ResolveUser(data.User)
	if !ok {
		return fmt.Errorf("cannot resolve user %q", data.User)
	}

	if err := r.queries.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{
		UsersIdusers: uid,
		Name:         "user",
	}); err != nil {
		return fmt.Errorf("grant user role: %w", err)
	}

	return nil
}
