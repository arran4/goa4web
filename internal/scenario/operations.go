package scenario

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/permissions"
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
		&UserGrantOp{},
		&PrivateForumCreateOp{},
		&PrivateForumEditOp{},
		&ForumReplyEditOp{},
		&ForumLabelAddOp{},
		&ForumLabelRemoveOp{},
		&ForumThreadCreateOp{},
		&ForumReplyOp{},
		&ForumThreadReadOp{},
		&ForumSubscribeOp{},
		&ForumUnsubscribeOp{},
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

// --- forum.thread.create ---

// ForumThreadCreateData holds the strongly-typed data for forum.thread.create.
type ForumThreadCreateData struct {
	Ref   string
	Actor string
	Topic string
	Body  string
	At    time.Time
}

func (d *ForumThreadCreateData) Op() string { return "forum.thread.create" }

// ForumThreadCreateOp implements Operation for semantic forum thread creation.
type ForumThreadCreateOp struct{}

func (o *ForumThreadCreateOp) OpName() string { return "forum.thread.create" }

func (o *ForumThreadCreateOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "Topic", "At"}
}

func (o *ForumThreadCreateOp) RequiredHeaders() []string {
	return []string{"Actor", "Topic", "At"}
}

func (o *ForumThreadCreateOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	ref := strings.TrimSpace(evt.Headers.Get("Ref"))
	if ref != "" {
		return RefTypeThread, ref, true
	}
	return "", "", false
}

func (o *ForumThreadCreateOp) ReferencedSymbols(evt *Event) []SymbolRef {
	return []SymbolRef{
		{Type: RefTypeUser, Symbol: strings.TrimSpace(evt.Headers.Get("Actor")), Field: "Actor"},
		{Type: RefTypeForum, Symbol: strings.TrimSpace(evt.Headers.Get("Topic")), Field: "Topic"},
	}
}

func (o *ForumThreadCreateOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumThreadCreateOp) Parse(evt *Event) (OperationData, error) {
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		return nil, fmt.Errorf("forum.thread.create: missing required 'Actor'")
	}
	topic := strings.TrimSpace(evt.Headers.Get("Topic"))
	if topic == "" {
		return nil, fmt.Errorf("forum.thread.create: missing required 'Topic'")
	}
	return &ForumThreadCreateData{
		Ref: strings.TrimSpace(evt.Headers.Get("Ref")), Actor: actor, Topic: topic,
		Body: strings.TrimRight(evt.Body, "\r\n"), At: evt.At,
	}, nil
}

// --- forum.reply ---

// ForumReplyData holds the strongly-typed data for forum.reply.
type ForumReplyData struct {
	Ref    string
	Actor  string
	Thread string
	Body   string
	At     time.Time
}

func (d *ForumReplyData) Op() string { return "forum.reply" }

// ForumReplyOp implements Operation for semantic forum replies.
type ForumReplyOp struct{}

func (o *ForumReplyOp) OpName() string { return "forum.reply" }

func (o *ForumReplyOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "Thread", "At"}
}

func (o *ForumReplyOp) RequiredHeaders() []string {
	return []string{"Actor", "Thread", "At"}
}

func (o *ForumReplyOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	ref := strings.TrimSpace(evt.Headers.Get("Ref"))
	if ref != "" {
		return RefTypePost, ref, true
	}
	return "", "", false
}

func (o *ForumReplyOp) ReferencedSymbols(evt *Event) []SymbolRef {
	return []SymbolRef{
		{Type: RefTypeUser, Symbol: strings.TrimSpace(evt.Headers.Get("Actor")), Field: "Actor"},
		{Type: RefTypeThread, Symbol: strings.TrimSpace(evt.Headers.Get("Thread")), Field: "Thread"},
	}
}

func (o *ForumReplyOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumReplyOp) Parse(evt *Event) (OperationData, error) {
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		return nil, fmt.Errorf("forum.reply: missing required 'Actor'")
	}
	thread := strings.TrimSpace(evt.Headers.Get("Thread"))
	if thread == "" {
		return nil, fmt.Errorf("forum.reply: missing required 'Thread'")
	}
	return &ForumReplyData{
		Ref: strings.TrimSpace(evt.Headers.Get("Ref")), Actor: actor, Thread: thread,
		Body: strings.TrimRight(evt.Body, "\r\n"), At: evt.At,
	}, nil
}

// --- forum.thread.read ---

// ForumThreadReadData holds the strongly-typed data for forum.thread.read.
type ForumThreadReadData struct {
	Actor  string
	Thread string
	At     time.Time
}

func (d *ForumThreadReadData) Op() string { return "forum.thread.read" }

// ForumThreadReadOp implements Operation for explicit mark-read behavior.
type ForumThreadReadOp struct{}

func (o *ForumThreadReadOp) OpName() string { return "forum.thread.read" }

func (o *ForumThreadReadOp) AllowedHeaders() []string {
	return []string{"Op", "Actor", "Thread", "At"}
}

func (o *ForumThreadReadOp) RequiredHeaders() []string {
	return []string{"Actor", "Thread"}
}

func (o *ForumThreadReadOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *ForumThreadReadOp) ReferencedSymbols(evt *Event) []SymbolRef {
	return []SymbolRef{
		{Type: RefTypeUser, Symbol: strings.TrimSpace(evt.Headers.Get("Actor")), Field: "Actor"},
		{Type: RefTypeThread, Symbol: strings.TrimSpace(evt.Headers.Get("Thread")), Field: "Thread"},
	}
}

func (o *ForumThreadReadOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumThreadReadOp) Parse(evt *Event) (OperationData, error) {
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		return nil, fmt.Errorf("forum.thread.read: missing required 'Actor'")
	}
	thread := strings.TrimSpace(evt.Headers.Get("Thread"))
	if thread == "" {
		return nil, fmt.Errorf("forum.thread.read: missing required 'Thread'")
	}
	return &ForumThreadReadData{
		Actor: actor, Thread: thread, At: evt.At,
	}, nil
}

// --- forum.subscribe ---

// ForumSubscribeData holds the strongly-typed data for forum.subscribe.
type ForumSubscribeData struct {
	Actor string
	Topic string

	At time.Time
}

func (d *ForumSubscribeData) Op() string { return "forum.subscribe" }

// ForumSubscribeOp implements Operation for explicit subscribe behavior.
type ForumSubscribeOp struct{}

func (o *ForumSubscribeOp) OpName() string { return "forum.subscribe" }

func (o *ForumSubscribeOp) AllowedHeaders() []string {
	return []string{"Op", "Actor", "Topic", "At"}
}

func (o *ForumSubscribeOp) RequiredHeaders() []string {
	return []string{"Actor", "Topic"}
}

func (o *ForumSubscribeOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *ForumSubscribeOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if topic := strings.TrimSpace(evt.Headers.Get("Topic")); topic != "" {
		refs = append(refs, SymbolRef{Type: RefTypeForum, Symbol: topic, Field: "Topic"})
	}

	return refs
}

func (o *ForumSubscribeOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumSubscribeOp) Parse(evt *Event) (OperationData, error) {
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		return nil, fmt.Errorf("forum.subscribe: missing required 'Actor'")
	}
	topic := strings.TrimSpace(evt.Headers.Get("Topic"))
	if topic == "" {
		return nil, fmt.Errorf("forum.subscribe: missing required 'Topic'")
	}

	return &ForumSubscribeData{
		Actor: actor,
		Topic: topic,
		At:    evt.At,
	}, nil
}

// --- forum.unsubscribe ---

// ForumUnsubscribeData holds the strongly-typed data for forum.unsubscribe.
type ForumUnsubscribeData struct {
	Actor string
	Topic string

	At time.Time
}

func (d *ForumUnsubscribeData) Op() string { return "forum.unsubscribe" }

// ForumUnsubscribeOp implements Operation for explicit unsubscribe behavior.
type ForumUnsubscribeOp struct{}

func (o *ForumUnsubscribeOp) OpName() string { return "forum.unsubscribe" }

func (o *ForumUnsubscribeOp) AllowedHeaders() []string {
	return []string{"Op", "Actor", "Topic", "At"}
}

func (o *ForumUnsubscribeOp) RequiredHeaders() []string {
	return []string{"Actor", "Topic"}
}

func (o *ForumUnsubscribeOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *ForumUnsubscribeOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if topic := strings.TrimSpace(evt.Headers.Get("Topic")); topic != "" {
		refs = append(refs, SymbolRef{Type: RefTypeForum, Symbol: topic, Field: "Topic"})
	}

	return refs
}

func (o *ForumUnsubscribeOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumUnsubscribeOp) Parse(evt *Event) (OperationData, error) {
	actor := strings.TrimSpace(evt.Headers.Get("Actor"))
	if actor == "" {
		return nil, fmt.Errorf("forum.unsubscribe: missing required 'Actor'")
	}
	topic := strings.TrimSpace(evt.Headers.Get("Topic"))
	if topic == "" {
		return nil, fmt.Errorf("forum.unsubscribe: missing required 'Topic'")
	}

	return &ForumUnsubscribeData{
		Actor: actor,
		Topic: topic,
		At:    evt.At,
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

// --- user.grant ---

// UserGrantData holds the strongly-typed data for user.grant.
type UserGrantData struct {
	User    string
	Section string
	Item    string
	ItemRef string
	Action  string
	At      time.Time
}

func (d *UserGrantData) Op() string { return "user.grant" }

// UserGrantOp implements Operation for granting permissions to a specific user.
type UserGrantOp struct{}

func (o *UserGrantOp) OpName() string { return "user.grant" }

func (o *UserGrantOp) AllowedHeaders() []string {
	return []string{"Op", "User", "Section", "Item", "ItemRef", "Action", "At"}
}

func (o *UserGrantOp) RequiredHeaders() []string {
	return []string{"User", "Section", "Action", "At"}
}

func (o *UserGrantOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *UserGrantOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if u := strings.TrimSpace(evt.Headers.Get("User")); u != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: u, Field: "User"})
	}
	section := strings.TrimSpace(evt.Headers.Get("Section"))
	item := strings.TrimSpace(evt.Headers.Get("Item"))
	itemRef := strings.TrimSpace(evt.Headers.Get("ItemRef"))

	if itemRef != "" {
		if section == "privateforum_thread" && item == "thread" {
			refs = append(refs, SymbolRef{Type: RefTypeThread, Symbol: itemRef, Field: "ItemRef"})
		}
	}
	return refs
}

func (o *UserGrantOp) AssetPaths(evt *Event) []string {
	return nil
}

func (o *UserGrantOp) Parse(evt *Event) (OperationData, error) {
	user := strings.TrimSpace(evt.Headers.Get("User"))
	if user == "" {
		return nil, fmt.Errorf("user.grant: missing required 'User'")
	}
	section := strings.TrimSpace(evt.Headers.Get("Section"))
	if section == "" {
		return nil, fmt.Errorf("user.grant: missing required 'Section'")
	}
	action := strings.TrimSpace(evt.Headers.Get("Action"))
	if action == "" {
		return nil, fmt.Errorf("user.grant: missing required 'Action'")
	}
	item := strings.TrimSpace(evt.Headers.Get("Item"))
	itemRef := strings.TrimSpace(evt.Headers.Get("ItemRef"))

	def := permissions.Lookup(section, item, action)
	if def == nil {
		return nil, fmt.Errorf("user.grant: invalid or unsupported permission tuple (%s/%s/%s)", section, item, action)
	}
	if def.RequireItemID {
		if itemRef == "" {
			return nil, fmt.Errorf("user.grant: permission (%s/%s/%s) requires a concrete item ID and cannot be granted globally (ItemRef is required)", section, item, action)
		}
		// Validate that the section/item combo is supported for reference resolution.
		if section == "privateforum_thread" && item == "thread" {
			// Supported
		} else {
			return nil, fmt.Errorf("user.grant: ItemRef resolution is not supported for section %q item %q", section, item)
		}
	} else if itemRef != "" {
		return nil, fmt.Errorf("user.grant: ItemRef provided but permission (%s/%s/%s) does not support or require an item ID", section, item, action)
	}

	return &UserGrantData{
		User:    user,
		Section: section,
		Item:    item,
		ItemRef: itemRef,
		Action:  action,
		At:      evt.At,
	}, nil
}

type PrivateForumEditData struct {
	Actor       string
	ItemRef     string
	Title       string
	Description string
}

func (d *PrivateForumEditData) Op() string { return "private-forum.edit" }

type PrivateForumEditOp struct{}

func (o *PrivateForumEditOp) OpName() string { return "private-forum.edit" }

func (o *PrivateForumEditOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "ItemRef", "Title", "Description", "At"}
}

func (o *PrivateForumEditOp) RequiredHeaders() []string {
	return []string{"Actor", "ItemRef", "At"}
}

func (o *PrivateForumEditOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *PrivateForumEditOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if item := strings.TrimSpace(evt.Headers.Get("ItemRef")); item != "" {
		refs = append(refs, SymbolRef{Type: RefTypeForum, Symbol: item, Field: "ItemRef"})
	}
	return refs
}

func (o *PrivateForumEditOp) AssetPaths(evt *Event) []string { return nil }

func (o *PrivateForumEditOp) Parse(evt *Event) (OperationData, error) {
	return &PrivateForumEditData{
		Actor:       strings.TrimSpace(evt.Headers.Get("Actor")),
		ItemRef:     strings.TrimSpace(evt.Headers.Get("ItemRef")),
		Title:       evt.Headers.Get("Title"),
		Description: evt.Headers.Get("Description"),
	}, nil
}

type ForumReplyEditData struct {
	Actor   string
	ItemRef string
	Text    string
}

func (d *ForumReplyEditData) Op() string { return "forum.reply.edit" }

type ForumReplyEditOp struct{}

func (o *ForumReplyEditOp) OpName() string { return "forum.reply.edit" }

func (o *ForumReplyEditOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "ItemRef", "At"}
}

func (o *ForumReplyEditOp) RequiredHeaders() []string {
	return []string{"Actor", "ItemRef", "At"}
}

func (o *ForumReplyEditOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *ForumReplyEditOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if item := strings.TrimSpace(evt.Headers.Get("ItemRef")); item != "" {
		refs = append(refs, SymbolRef{Type: RefTypePost, Symbol: item, Field: "ItemRef"})
	}
	return refs
}

func (o *ForumReplyEditOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumReplyEditOp) Parse(evt *Event) (OperationData, error) {
	return &ForumReplyEditData{
		Actor:   strings.TrimSpace(evt.Headers.Get("Actor")),
		ItemRef: strings.TrimSpace(evt.Headers.Get("ItemRef")),
		Text:    evt.Body,
	}, nil
}

type ForumLabelData struct {
	Actor   string
	ItemRef string
	Label   string
	Private bool
}

func (d *ForumLabelData) Op() string { return "forum.label" }

type ForumLabelAddOp struct{}

func (o *ForumLabelAddOp) OpName() string { return "forum.label.add" }

func (o *ForumLabelAddOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "ItemRef", "Label", "Private", "At"}
}

func (o *ForumLabelAddOp) RequiredHeaders() []string {
	return []string{"Actor", "ItemRef", "Label", "At"}
}

func (o *ForumLabelAddOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *ForumLabelAddOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if item := strings.TrimSpace(evt.Headers.Get("ItemRef")); item != "" {
		refs = append(refs, SymbolRef{Type: RefTypeThread, Symbol: item, Field: "ItemRef"}) // We assume thread for labels for now based on context, although it can be topic too. Let's resolve as thread for staff-welcome.
	}
	return refs
}

func (o *ForumLabelAddOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumLabelAddOp) Parse(evt *Event) (OperationData, error) {
	d := &ForumLabelData{
		Actor:   strings.TrimSpace(evt.Headers.Get("Actor")),
		ItemRef: strings.TrimSpace(evt.Headers.Get("ItemRef")),
		Label:   strings.TrimSpace(evt.Headers.Get("Label")),
		Private: evt.Headers.Get("Private") == "true",
	}
	// override the Op returned by ForumLabelData so dispatch works
	return &forumLabelDataWithOp{ForumLabelData: *d, op: "forum.label.add"}, nil
}

type ForumLabelRemoveOp struct{}

func (o *ForumLabelRemoveOp) OpName() string { return "forum.label.remove" }

func (o *ForumLabelRemoveOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "ItemRef", "Label", "Private", "At"}
}

func (o *ForumLabelRemoveOp) RequiredHeaders() []string {
	return []string{"Actor", "ItemRef", "Label", "At"}
}

func (o *ForumLabelRemoveOp) DeclaredRef(evt *Event) (RefType, string, bool) {
	return "", "", false
}

func (o *ForumLabelRemoveOp) ReferencedSymbols(evt *Event) []SymbolRef {
	var refs []SymbolRef
	if actor := strings.TrimSpace(evt.Headers.Get("Actor")); actor != "" {
		refs = append(refs, SymbolRef{Type: RefTypeUser, Symbol: actor, Field: "Actor"})
	}
	if item := strings.TrimSpace(evt.Headers.Get("ItemRef")); item != "" {
		refs = append(refs, SymbolRef{Type: RefTypeThread, Symbol: item, Field: "ItemRef"}) 
	}
	return refs
}

func (o *ForumLabelRemoveOp) AssetPaths(evt *Event) []string { return nil }

func (o *ForumLabelRemoveOp) Parse(evt *Event) (OperationData, error) {
	d := &ForumLabelData{
		Actor:   strings.TrimSpace(evt.Headers.Get("Actor")),
		ItemRef: strings.TrimSpace(evt.Headers.Get("ItemRef")),
		Label:   strings.TrimSpace(evt.Headers.Get("Label")),
		Private: evt.Headers.Get("Private") == "true",
	}
	return &forumLabelDataWithOp{ForumLabelData: *d, op: "forum.label.remove"}, nil
}

type forumLabelDataWithOp struct {
	ForumLabelData
	op string
}

func (d *forumLabelDataWithOp) Op() string { return d.op }
