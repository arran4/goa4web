package scenario

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// OperationData is a marker interface for parsed, strongly-typed operation payloads.
type OperationData interface {
	Op() string
}

// Operation defines the specification and validator/parser for an application operation.
type Operation interface {
	// OpName returns the unique string identifier for the operation (e.g. "user.create").
	OpName() string
	// AllowedHeaders returns the list of valid header names for this operation.
	AllowedHeaders() []string
	// RequiredHeaders returns the list of mandatory header names for this operation.
	RequiredHeaders() []string
	// DeclaredRef returns the RefType and symbol declared by this event, if any.
	DeclaredRef(evt *Event) (RefType, string, bool)
	// ReferencedSymbols returns all symbolic references required by this event.
	ReferencedSymbols(evt *Event) []SymbolRef
	// AssetPaths returns all relative asset file paths declared in this event's headers.
	AssetPaths(evt *Event) []string
	// Parse parses and validates the event headers/body into a typed OperationData struct.
	Parse(evt *Event) (OperationData, error)
}

// Registry manages known scenario operations without global state.
type Registry struct {
	ops map[string]Operation
}

// NewRegistry creates a new Operation Registry with the given operations.
func NewRegistry(ops ...Operation) *Registry {
	r := &Registry{
		ops: make(map[string]Operation, len(ops)),
	}
	for _, op := range ops {
		r.ops[op.OpName()] = op
	}
	return r
}

// DefaultRegistry returns a standard Registry populated with the default set of known operations.
func DefaultRegistry() *Registry {
	return NewRegistry(
		&UserCreateOp{},
		&UserEnableOp{},
		&RoleGrantOp{},
		&PrivateForumCreateOp{},
		&ForumPostOp{},
	)
}

// Register registers an Operation in the registry.
func (r *Registry) Register(op Operation) {
	if r.ops == nil {
		r.ops = make(map[string]Operation)
	}
	r.ops[op.OpName()] = op
}

// Lookup retrieves an Operation from the registry by name.
func (r *Registry) Lookup(name string) (Operation, bool) {
	if r == nil || r.ops == nil {
		return nil, false
	}
	op, ok := r.ops[name]
	return op, ok
}

// RegisteredOperations returns the sorted list of registered operation names.
func (r *Registry) RegisteredOperations() []string {
	if r == nil || len(r.ops) == 0 {
		return nil
	}
	names := make([]string, 0, len(r.ops))
	for name := range r.ops {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// --- user.create ---

// UserCreateData holds the strongly-typed data for user.create.
type UserCreateData struct {
	Ref      string
	Username string
	Email    string
	Password string
	At       time.Time
}

func (d *UserCreateData) Op() string { return "user.create" }

// UserCreateOp implements Operation for user.create.
type UserCreateOp struct{}

func (o *UserCreateOp) OpName() string { return "user.create" }

func (o *UserCreateOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Username", "Email", "Password", "At"}
}

func (o *UserCreateOp) RequiredHeaders() []string {
	return []string{"Username", "Email", "Password", "At"}
}

func (o *UserCreateOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	ref := strings.TrimSpace(evt.Headers.Get("Ref"))
	if ref != "" {
		return RefTypeUser, ref, true
	}
	return "", "", false
}

func (o *UserCreateOp) ReferencedSymbols(evt *Event) []SymbolRef {
	return nil
}

func (o *UserCreateOp) AssetPaths(evt *Event) []string {
	return nil
}

func (o *UserCreateOp) Parse(evt *Event) (OperationData, error) {
	username := strings.TrimSpace(evt.Headers.Get("Username"))
	if username == "" {
		return nil, fmt.Errorf("user.create: missing required 'Username'")
	}
	email := strings.TrimSpace(evt.Headers.Get("Email"))
	if email == "" {
		return nil, fmt.Errorf("user.create: missing required 'Email'")
	}
	password := strings.TrimSpace(evt.Headers.Get("Password"))
	if password == "" {
		return nil, fmt.Errorf("user.create: missing required 'Password'")
	}
	return &UserCreateData{
		Ref:      strings.TrimSpace(evt.Headers.Get("Ref")),
		Username: username,
		Email:    email,
		Password: password,
		At:       evt.At,
	}, nil
}

// --- user.enable ---

// UserEnableData holds the strongly-typed data for user.enable.
type UserEnableData struct {
	Actor string
	User  string
	At    time.Time
}

func (d *UserEnableData) Op() string { return "user.enable" }

// UserEnableOp implements Operation for user.enable.
type UserEnableOp struct{}

func (o *UserEnableOp) OpName() string { return "user.enable" }

func (o *UserEnableOp) AllowedHeaders() []string {
	return []string{"Op", "Actor", "User", "At"}
}

func (o *UserEnableOp) RequiredHeaders() []string {
	return []string{"User", "At"}
}

func (o *UserEnableOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *UserEnableOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" && actor != "admin" && actor != "system" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if user := strings.TrimSpace(evt.Headers.Get("User")); user != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: user, Field: "User"})
	}
	return refs
}

func (o *UserEnableOp) AssetPaths(evt *Event) []string {
	return nil
}

func (o *UserEnableOp) Parse(evt *Event) (OperationData, error) {
	user := strings.TrimSpace(evt.Headers.Get("User"))
	if user == "" {
		return nil, fmt.Errorf("user.enable: missing required 'User'")
	}
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		actor = "admin"
	}
	return &UserEnableData{
		Actor: actor,
		User:  user,
		At:    evt.At,
	}, nil
}

// --- private-forum.create ---

// PrivateForumCreateData holds the strongly-typed data for private-forum.create.
type PrivateForumCreateData struct {
	Ref          string
	Actor        string
	Participants []string
	Title        string
	Description  string
	At           time.Time
}

func (d *PrivateForumCreateData) Op() string { return "private-forum.create" }

// PrivateForumCreateOp implements Operation for private-forum.create.
type PrivateForumCreateOp struct{}

func (o *PrivateForumCreateOp) OpName() string { return "private-forum.create" }

func (o *PrivateForumCreateOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "Participant", "Title", "Description", "At"}
}

func (o *PrivateForumCreateOp) RequiredHeaders() []string {
	return []string{"Actor", "Participant", "Title", "At"}
}

func (o *PrivateForumCreateOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	ref := strings.TrimSpace(evt.Headers.Get("Ref"))
	if ref != "" {
		return RefTypeForum, ref, true
	}
	return "", "", false
}

func (o *PrivateForumCreateOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	for _, p := range evt.Headers.Values("Participant") {
		if p = strings.TrimSpace(p); p != "" {
			refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: p, Field: "Participant"})
		}
	}
	return refs
}

func (o *PrivateForumCreateOp) AssetPaths(evt *Event) []string {
	return nil
}

func (o *PrivateForumCreateOp) Parse(evt *Event) (OperationData, error) {
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		return nil, fmt.Errorf("private-forum.create: missing required 'Actor'")
	}
	title := strings.TrimSpace(evt.Headers.Get("Title"))
	if title == "" {
		return nil, fmt.Errorf("private-forum.create: missing required 'Title'")
	}

	participantsRaw := evt.Headers.Values("Participant")
	if len(participantsRaw) == 0 {
		return nil, fmt.Errorf("private-forum.create: missing required 'Participant'")
	}

	seen := make(map[string]bool)
	var participants []string
	for _, p := range participantsRaw {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if p == actor {
			return nil, fmt.Errorf("private-forum.create: cannot specify Actor %q as Participant", actor)
		}
		if seen[p] {
			return nil, fmt.Errorf("private-forum.create: duplicate participant %q", p)
		}
		seen[p] = true
		participants = append(participants, p)
	}

	if len(participants) == 0 {
		return nil, fmt.Errorf("private-forum.create: at least one participant other than Actor is required")
	}

	return &PrivateForumCreateData{
		Ref:          strings.TrimSpace(evt.Headers.Get("Ref")),
		Actor:        actor,
		Participants: participants,
		Title:        title,
		Description:  strings.TrimSpace(evt.Headers.Get("Description")),
		At:           evt.At,
	}, nil
}

// --- forum.post ---

// ForumPostData holds the strongly-typed data for forum.post.
type ForumPostData struct {
	Ref         string
	Actor       string
	Forum       string
	Topic       string
	Thread      string
	Attachments []string
	Body        string
	At          time.Time
}

func (d *ForumPostData) Op() string { return "forum.post" }

// ForumPostOp implements Operation for forum.post.
type ForumPostOp struct{}

func (o *ForumPostOp) OpName() string { return "forum.post" }

func (o *ForumPostOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "Forum", "Topic", "Thread", "Attachment", "At"}
}

func (o *ForumPostOp) RequiredHeaders() []string {
	return []string{"Actor", "At"}
}

func (o *ForumPostOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	ref := strings.TrimSpace(evt.Headers.Get("Ref"))
	if ref != "" {
		return RefTypePost, ref, true
	}
	return "", "", false
}

func (o *ForumPostOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if forum := strings.TrimSpace(evt.Headers.Get("Forum")); forum != "" {
		refs = append(refs, SymbolRef{Type: RefTypeForum, Symbol: forum, Field: "Forum"})
	}
	if topic := strings.TrimSpace(evt.Headers.Get("Topic")); topic != "" {
		refs = append(refs, SymbolRef{Type: RefTypeTopic, Symbol: topic, Field: "Topic"})
	}
	if thread := strings.TrimSpace(evt.Headers.Get("Thread")); thread != "" {
		refs = append(refs, SymbolRef{Type: RefTypeThread, Symbol: thread, Field: "Thread"})
	}
	return refs
}

func (o *ForumPostOp) AssetPaths(evt *Event) []string {
	return evt.Headers.Values("Attachment")
}

func (o *ForumPostOp) Parse(evt *Event) (OperationData, error) {
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		return nil, fmt.Errorf("forum.post: missing required 'Actor'")
	}
	forum := strings.TrimSpace(evt.Headers.Get("Forum"))
	topic := strings.TrimSpace(evt.Headers.Get("Topic"))
	thread := strings.TrimSpace(evt.Headers.Get("Thread"))
	if forum == "" && topic == "" && thread == "" {
		return nil, fmt.Errorf("forum.post: one of 'Forum', 'Topic', or 'Thread' is required")
	}

	return &ForumPostData{
		Ref:         strings.TrimSpace(evt.Headers.Get("Ref")),
		Actor:       actor,
		Forum:       forum,
		Topic:       topic,
		Thread:      thread,
		Attachments: evt.Headers.Values("Attachment"),
		Body:        evt.Body,
		At:          evt.At,
	}, nil
}

// --- role.grant ---

// RoleGrantData holds the strongly-typed data for role.grant.
type RoleGrantData struct {
	Role    string
	Section string
	Item    string
	Action  string
	At      time.Time
}

func (d *RoleGrantData) Op() string { return "role.grant" }

// RoleGrantOp implements Operation for granting permissions to a named role.
type RoleGrantOp struct{}

func (o *RoleGrantOp) OpName() string { return "role.grant" }

func (o *RoleGrantOp) AllowedHeaders() []string {
	return []string{"Op", "Role", "Section", "Item", "Action", "At"}
}

func (o *RoleGrantOp) RequiredHeaders() []string {
	return []string{"Role", "Section", "Action", "At"}
}

func (o *RoleGrantOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *RoleGrantOp) ReferencedSymbols(evt *Event) []SymbolRef {
	return nil
}

func (o *RoleGrantOp) AssetPaths(evt *Event) []string {
	return nil
}

func (o *RoleGrantOp) Parse(evt *Event) (OperationData, error) {
	role := strings.TrimSpace(evt.Headers.Get("Role"))
	if role == "" {
		return nil, fmt.Errorf("role.grant: missing required 'Role'")
	}
	section := strings.TrimSpace(evt.Headers.Get("Section"))
	if section == "" {
		return nil, fmt.Errorf("role.grant: missing required 'Section'")
	}
	action := strings.TrimSpace(evt.Headers.Get("Action"))
	if action == "" {
		return nil, fmt.Errorf("role.grant: missing required 'Action'")
	}
	item := strings.TrimSpace(evt.Headers.Get("Item"))

	return &RoleGrantData{
		Role:    role,
		Section: section,
		Item:    item,
		Action:  action,
		At:      evt.At,
	}, nil
}
