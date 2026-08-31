package scenario

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnknownOp indicates that an event specified an unrecognized operation name.
type ErrUnknownOp struct {
	Op string
}

func (e ErrUnknownOp) Error() string {
	return fmt.Sprintf("unknown operation %q", e.Op)
}

// ErrUnknownField indicates that an event or meta section contains an unrecognized header.
type ErrUnknownField struct {
	Section string
	Field   string
}

func (e ErrUnknownField) Error() string {
	if e.Section != "" {
		return fmt.Sprintf("unknown field %q in %s", e.Field, e.Section)
	}
	return fmt.Sprintf("unknown field %q", e.Field)
}

// ErrMissingRequiredField indicates that a mandatory header was omitted.
type ErrMissingRequiredField struct {
	Section string
	Field   string
}

func (e ErrMissingRequiredField) Error() string {
	if e.Section != "" {
		return fmt.Sprintf("missing required field %q in %s", e.Field, e.Section)
	}
	return fmt.Sprintf("missing required field %q", e.Field)
}

// Validate performs strict structural, referential, and asset validation on a Scenario.
func Validate(s *Scenario) error {
	if s == nil {
		return errors.New("nil scenario")
	}

	// 1. Validate scenario.meta
	if s.Meta.Format == "" {
		return ErrMissingRequiredField{Section: MetaFile, Field: "Format"}
	}
	if s.Meta.Format != FormatV1 {
		return ErrInvalidFormat{Format: s.Meta.Format}
	}
	if s.Meta.Name == "" {
		return ErrMissingRequiredField{Section: MetaFile, Field: "Name"}
	}

	// Validate allowed meta headers
	allowedMeta := map[string]bool{
		"format":      true,
		"name":        true,
		"description": true,
	}
	for _, k := range s.Meta.Headers.Keys() {
		if !allowedMeta[strings.ToLower(k)] {
			return ErrUnknownField{Section: MetaFile, Field: k}
		}
	}

	// 2. Validate events sequentially
	reg := NewRefRegistry()

	for idx, evt := range s.Events {
		sectionName := evt.File
		if sectionName == "" {
			sectionName = fmt.Sprintf("event[%d]", idx)
		}

		if evt.Op == "" {
			return ErrMissingRequiredField{Section: sectionName, Field: "Op"}
		}

		op, ok := LookupOperation(evt.Op)
		if !ok {
			return fmt.Errorf("%s: %w", sectionName, ErrUnknownOp{Op: evt.Op})
		}

		// Check timestamp
		if evt.At.IsZero() {
			return fmt.Errorf("%s (%s): %w", sectionName, evt.Op, ErrMissingRequiredField{Section: sectionName, Field: "At"})
		}

		// Check unknown headers
		allowedMap := make(map[string]bool)
		for _, ah := range op.AllowedHeaders() {
			allowedMap[strings.ToLower(ah)] = true
		}
		for _, k := range evt.Headers.Keys() {
			if !allowedMap[strings.ToLower(k)] {
				return fmt.Errorf("%s (%s): %w", sectionName, evt.Op, ErrUnknownField{Section: sectionName, Field: k})
			}
		}

		// Check required headers
		for _, req := range op.RequiredHeaders() {
			if !evt.Headers.Has(req) || strings.TrimSpace(evt.Headers.Get(req)) == "" {
				return fmt.Errorf("%s (%s): %w", sectionName, evt.Op, ErrMissingRequiredField{Section: sectionName, Field: req})
			}
		}

		// Check referenced symbols (ordering matters: must be declared before use)
		for _, ref := range op.ReferencedSymbols(evt) {
			if !reg.HasDeclared(ref.Type, ref.Symbol) {
				// Check if declared under a different type
				if actualType, ok := reg.LookupDeclaredType(ref.Symbol); ok && actualType != ref.Type {
					return fmt.Errorf("%s (%s): %w", sectionName, evt.Op, ErrRefTypeMismatch{
						Symbol:       ref.Symbol,
						ExpectedType: ref.Type,
						ActualType:   actualType,
						Field:        ref.Field,
					})
				}
				return fmt.Errorf("%s (%s): %w", sectionName, evt.Op, ErrUnresolvedRef{
					Type:   ref.Type,
					Symbol: ref.Symbol,
					Field:  ref.Field,
				})
			}
		}

		// Check declared ref
		if typ, symbol, hasDecl := op.DeclaredRef(evt); hasDecl {
			if err := reg.Declare(typ, symbol); err != nil {
				return fmt.Errorf("%s (%s): %w", sectionName, evt.Op, err)
			}
		}

		// Check asset paths
		for _, assetPath := range op.AssetPaths(evt) {
			if _, err := ValidateAsset(s.FS, assetPath); err != nil {
				return fmt.Errorf("%s (%s): %w", sectionName, evt.Op, err)
			}
		}

		// Parse operation data
		opData, err := op.Parse(evt)
		if err != nil {
			return fmt.Errorf("%s (%s): parse: %w", sectionName, evt.Op, err)
		}
		evt.OpData = opData
	}

	return nil
}
