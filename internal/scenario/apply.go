package scenario

import (
	"context"
	"fmt"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
)

// ErrUnsupportedOperation indicates that an operation is valid in format but not yet supported by this runner.
type ErrUnsupportedOperation struct {
	Op        string
	EventFile string
}

func (e ErrUnsupportedOperation) Error() string {
	if e.EventFile != "" {
		return fmt.Sprintf("event %s: operation %q is not supported for application in this runner", e.EventFile, e.Op)
	}
	return fmt.Sprintf("operation %q is not supported for application in this runner", e.Op)
}

// ApplyResult contains the outcome of applying a scenario.
type ApplyResult struct {
	ScenarioName  string
	EventsApplied int
	Registry      *RefRegistry
}

// Runner applies validated scenario events to a goa4web database instance.
type Runner struct {
	coreData     *common.CoreData
	opRegistry   *Registry
	refRegistry  *RefRegistry
	supportedOps map[string]bool
}

// Option configures a Runner.
type Option func(*Runner)

// WithOpRegistry sets a custom Operation Registry on the Runner.
func WithOpRegistry(reg *Registry) Option {
	return func(r *Runner) {
		r.opRegistry = reg
	}
}

// WithRefRegistry sets a custom RefRegistry on the Runner.
func WithRefRegistry(reg *RefRegistry) Option {
	return func(r *Runner) {
		r.refRegistry = reg
	}
}

// NewRunner creates a new Runner with CoreData.
func NewRunner(cd *common.CoreData, opts ...Option) *Runner {
	r := &Runner{
		coreData:    cd,
		opRegistry:  DefaultRegistry(),
		refRegistry: NewRefRegistry(),
		supportedOps: map[string]bool{
			"user.create":          true,
			"user.enable":          true,
			"private-forum.create": true,
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// NewRunnerWithQuerier creates a new Runner using a db.Querier.
func NewRunnerWithQuerier(ctx context.Context, q db.Querier, opts ...Option) *Runner {
	cd := common.NewCoreData(ctx, q, &config.RuntimeConfig{})
	return NewRunner(cd, opts...)
}

// Registry returns the runner's RefRegistry.
func (r *Runner) Registry() *RefRegistry {
	return r.refRegistry
}

// Preflight checks if the scenario is valid and if ALL events are supported by this runner
// before any database modifications occur.
func (r *Runner) Preflight(s *Scenario) error {
	if err := ValidateWithRegistry(s, r.opRegistry); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	for _, evt := range s.Events {
		if !r.supportedOps[evt.Op] {
			return ErrUnsupportedOperation{Op: evt.Op, EventFile: evt.File}
		}
	}
	return nil
}

// Apply validates, preflights, and executes all events in a scenario in order.
func (r *Runner) Apply(ctx context.Context, s *Scenario) (*ApplyResult, error) {
	if err := r.Preflight(s); err != nil {
		return nil, err
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
		Registry:      r.refRegistry,
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

	case "private-forum.create":
		data, ok := evt.OpData.(*PrivateForumCreateData)
		if !ok {
			return fmt.Errorf("invalid operation data for private-forum.create")
		}
		return r.applyPrivateForumCreate(ctx, data)

	default:
		return ErrUnsupportedOperation{Op: evt.Op, EventFile: evt.File}
	}
}

func (r *Runner) applyUserCreate(ctx context.Context, data *UserCreateData) error {
	hash, alg, err := auth.HashPassword(data.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	uid, err := r.coreData.CreateUserWithEmail(data.Username, data.Email, hash, alg)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("user %q already exists: %w", data.Username, err)
		}
		return fmt.Errorf("create user: %w", err)
	}

	if data.Ref != "" {
		if err := r.refRegistry.Bind(RefTypeUser, data.Ref, uid); err != nil {
			return fmt.Errorf("bind user ref %q: %w", data.Ref, err)
		}
	}

	return nil
}

func (r *Runner) applyUserEnable(ctx context.Context, data *UserEnableData) error {
	uid, ok := r.refRegistry.ResolveUser(data.User)
	if !ok {
		return fmt.Errorf("cannot resolve user %q", data.User)
	}

	if err := r.coreData.ApproveUser(uid); err != nil {
		return fmt.Errorf("approve user: %w", err)
	}

	return nil
}

func (r *Runner) applyPrivateForumCreate(ctx context.Context, data *PrivateForumCreateData) error {
	actorUID, ok := r.refRegistry.ResolveUser(data.Actor)
	if !ok {
		return fmt.Errorf("cannot resolve actor %q", data.Actor)
	}

	participants := make([]common.PrivateTopicParticipant, 0, len(data.Participants))
	for _, p := range data.Participants {
		uid, ok := r.refRegistry.ResolveUser(p)
		if !ok {
			return fmt.Errorf("cannot resolve participant %q", p)
		}
		participants = append(participants, common.PrivateTopicParticipant{
			ID:       uid,
			Username: p,
		})
	}

	actorCD := r.coreData
	if actorUID != 0 && r.coreData.UserID != actorUID {
		actorCD = r.coreData.ForUser(actorUID)
	}

	topicID, err := actorCD.CreatePrivateTopic(common.CreatePrivateTopicParams{
		CreatorID:    actorUID,
		Participants: participants,
		Title:        data.Title,
		Description:  data.Description,
	})
	if err != nil {
		return fmt.Errorf("create private topic: %w", err)
	}

	if data.Ref != "" {
		if err := r.refRegistry.Bind(RefTypeForum, data.Ref, topicID); err != nil {
			return fmt.Errorf("bind forum ref %q: %w", data.Ref, err)
		}
	}

	return nil
}
