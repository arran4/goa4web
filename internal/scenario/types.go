package scenario

import (
	"fmt"
	"io/fs"
	"strings"
	"time"
)

// FormatV1 is the format identifier for version 1 of goa4web scenarios.
const FormatV1 = "goa4web-scenario/v1"

// MetaFile is the standard name of the scenario metadata file inside a TXTAR archive.
const MetaFile = "scenario.meta"

// RefType identifies the entity type for a symbolic reference.
type RefType string

const (
	// RefTypeUser identifies a user reference.
	RefTypeUser RefType = "user"
	// RefTypeForum identifies a forum reference.
	RefTypeForum RefType = "forum"
	// RefTypeTopic identifies a forum topic reference.
	RefTypeTopic RefType = "topic"
	// RefTypeThread identifies a forum thread reference.
	RefTypeThread RefType = "thread"
	// RefTypePost identifies a post/comment reference.
	RefTypePost RefType = "post"
	// RefTypeBlog identifies a blog entry reference.
	RefTypeBlog RefType = "blog"
	// RefTypeWriting identifies a writing/article reference.
	RefTypeWriting RefType = "writing"
	// RefTypeLink identifies an external link reference.
	RefTypeLink RefType = "link"
	// RefTypeImage identifies an image reference.
	RefTypeImage RefType = "image"
)

// Meta contains the metadata parsed from scenario.meta.
type Meta struct {
	Format      string
	Name        string
	Description string
	Headers     Header
}

// Header represents RFC822/HTTP-style headers with support for repeated keys.
type Header struct {
	entries []HeaderEntry
}

// HeaderEntry represents a single key-value header entry.
type HeaderEntry struct {
	Key   string
	Value string
}

// NewHeader constructs a new Header from entries.
func NewHeader() Header {
	return Header{entries: nil}
}

// Add adds a key/value pair.
func (h *Header) Add(key, val string) {
	h.entries = append(h.entries, HeaderEntry{Key: key, Value: val})
}

// Set sets a key/value pair, replacing existing entries for that key.
func (h *Header) Set(key, val string) {
	h.Del(key)
	h.Add(key, val)
}

// Get returns the first value associated with the given key (case-insensitive), or "" if not found.
func (h Header) Get(key string) string {
	for _, e := range h.entries {
		if strings.EqualFold(e.Key, key) {
			return e.Value
		}
	}
	return ""
}

// Values returns all values associated with the given key (case-insensitive).
func (h Header) Values(key string) []string {
	var vals []string
	for _, e := range h.entries {
		if strings.EqualFold(e.Key, key) {
			vals = append(vals, e.Value)
		}
	}
	return vals
}

// Has checks if a key exists (case-insensitive).
func (h Header) Has(key string) bool {
	for _, e := range h.entries {
		if strings.EqualFold(e.Key, key) {
			return true
		}
	}
	return false
}

// Del deletes all entries matching key (case-insensitive).
func (h *Header) Del(key string) {
	var filtered []HeaderEntry
	for _, e := range h.entries {
		if !strings.EqualFold(e.Key, key) {
			filtered = append(filtered, e)
		}
	}
	h.entries = filtered
}

// Keys returns all unique header keys in their original casing.
func (h Header) Keys() []string {
	var keys []string
	seen := make(map[string]bool)
	for _, e := range h.entries {
		lower := strings.ToLower(e.Key)
		if !seen[lower] {
			seen[lower] = true
			keys = append(keys, e.Key)
		}
	}
	return keys
}

// Entries returns all raw header entries in declaration order.
func (h Header) Entries() []HeaderEntry {
	return h.entries
}

// Event represents a single parsed scenario event from a .event member.
type Event struct {
	File    string        // Source filename in txtar archive (e.g. "010-alice.event")
	Op      string        // Operation name (e.g. "user.create")
	Ref     string        // Declared symbolic reference, if any (from "Ref:" header)
	At      time.Time     // Event timestamp (from "At:" header)
	Headers Header        // All headers
	Body    string        // Optional multiline body
	OpData  OperationData // Parsed operation data, if available
}

// Scenario represents a complete parsed scenario with its metadata and ordered events.
type Scenario struct {
	Meta   Meta
	Events []*Event
	FS     fs.FS // Adjacent filesystem context for asset resolution
}

// SymbolRef identifies a reference to an entity used within an event.
type SymbolRef struct {
	Type   RefType // Expected reference type (e.g. RefTypeUser)
	Symbol string  // Reference name (e.g. "alice")
	Field  string  // Header field where the reference appeared (e.g. "Actor", "User", "Forum")
}

func (s SymbolRef) String() string {
	return fmt.Sprintf("%s:%s (in %s)", s.Type, s.Symbol, s.Field)
}
