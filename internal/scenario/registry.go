package scenario

import (
	"fmt"
	"sync"
)

// ErrDuplicateRef indicates that a symbol was declared more than once for a given reference type.
type ErrDuplicateRef struct {
	Type   RefType
	Symbol string
}

func (e ErrDuplicateRef) Error() string {
	return fmt.Sprintf("duplicate %s reference %q", e.Type, e.Symbol)
}

// ErrUnresolvedRef indicates that a symbol could not be resolved.
type ErrUnresolvedRef struct {
	Type   RefType
	Symbol string
	Field  string
}

func (e ErrUnresolvedRef) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("unresolved %s reference %q in field %q", e.Type, e.Symbol, e.Field)
	}
	return fmt.Sprintf("unresolved %s reference %q", e.Type, e.Symbol)
}

// ErrRefTypeMismatch indicates that a symbol was declared as one type but referenced as another.
type ErrRefTypeMismatch struct {
	Symbol       string
	ExpectedType RefType
	ActualType   RefType
	Field        string
}

func (e ErrRefTypeMismatch) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("reference %q in field %q is a %s, expected %s", e.Symbol, e.Field, e.ActualType, e.ExpectedType)
	}
	return fmt.Sprintf("reference %q is a %s, expected %s", e.Symbol, e.ActualType, e.ExpectedType)
}

type refKey struct {
	Type   RefType
	Symbol string
}

// RefRegistry manages declared and bound typed symbolic references.
type RefRegistry struct {
	mu          sync.RWMutex
	declared    map[refKey]bool
	symbolTypes map[string]RefType // tracks declared type for each symbol name
	values      map[refKey]any
}

// NewRefRegistry creates a new empty RefRegistry.
func NewRefRegistry() *RefRegistry {
	return &RefRegistry{
		declared:    make(map[refKey]bool),
		symbolTypes: make(map[string]RefType),
		values:      make(map[refKey]any),
	}
}

// Declare registers a symbol name for a given RefType during scenario validation or execution.
// Returns ErrDuplicateRef if the same symbol is declared multiple times for the same type.
func (r *RefRegistry) Declare(typ RefType, symbol string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	k := refKey{Type: typ, Symbol: symbol}
	if r.declared[k] {
		return ErrDuplicateRef{Type: typ, Symbol: symbol}
	}
	r.declared[k] = true
	if _, ok := r.symbolTypes[symbol]; !ok {
		r.symbolTypes[symbol] = typ
	}
	return nil
}

// HasDeclared returns true if the symbol is declared for the given type.
func (r *RefRegistry) HasDeclared(typ RefType, symbol string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Special built-in user actors (e.g. admin, system) are implicitly available as RefTypeUser.
	if typ == RefTypeUser && (symbol == "admin" || symbol == "system") {
		return true
	}

	return r.declared[refKey{Type: typ, Symbol: symbol}]
}

// LookupDeclaredType returns the RefType for which a symbol was declared, if any.
func (r *RefRegistry) LookupDeclaredType(symbol string) (RefType, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if symbol == "admin" || symbol == "system" {
		return RefTypeUser, true
	}

	t, ok := r.symbolTypes[symbol]
	return t, ok
}

// Bind assigns a runtime value (e.g. database ID) to a typed symbol.
func (r *RefRegistry) Bind(typ RefType, symbol string, val any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	k := refKey{Type: typ, Symbol: symbol}
	r.declared[k] = true
	if _, ok := r.symbolTypes[symbol]; !ok {
		r.symbolTypes[symbol] = typ
	}
	r.values[k] = val
	return nil
}

// Resolve returns the runtime value bound to a typed symbol, or false if not bound.
func (r *RefRegistry) Resolve(typ RefType, symbol string) (any, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	val, ok := r.values[refKey{Type: typ, Symbol: symbol}]
	return val, ok
}

// ResolveUser returns the int32 user ID bound to a user symbol.
func (r *RefRegistry) ResolveUser(symbol string) (int32, bool) {
	val, ok := r.Resolve(RefTypeUser, symbol)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case int32:
		return v, true
	case int:
		return int32(v), true
	case int64:
		return int32(v), true
	default:
		return 0, false
	}
}
