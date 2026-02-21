package a4code

import (
	"strings"

	"github.com/arran4/goa4web/a4code/ast"
)

// Substring extracts a substring based on visible text length, preserving markup.
// start and end are indices into the *visible text*.
func Substring(s string, start, end int) (string, error) {
	root, err := ParseString(s)
	if err != nil {
		// Fallback to simple string slicing if parsing fails, though imperfect.
		if start >= len(s) {
			return "", err
		}
		if end > len(s) {
			end = len(s)
		}
		return s[start:end], err
	}

	newRoot := &ast.Root{}
	pos := 0
	// filterNodes expects a slice of nodes and returns a slice of nodes.
	// It updates pos as it traverses text.
	newRoot.Children = filterNodes(root.Children, start, end, &pos)

	res := ToCode(newRoot)
	return res, nil
}

func filterNodes(nodes []ast.Node, start, end int, pos *int) []ast.Node {
	var kept []ast.Node

	for _, n := range nodes {
		if *pos >= end {
			break
		}

		switch t := n.(type) {
		case *ast.Text:
			l := len(t.Value)
			// Range of this node is [*pos, *pos+l)
			// Intersect with [start, end)
			s := max(*pos, start)
			e := min(*pos+l, end)

			if s < e {
				// Visible part
				relS := s - *pos
				relE := e - *pos
				// Clone node to avoid mutating original AST if reused (though here it's fresh)
				newText := &ast.Text{}
				newText.Value = t.Value[relS:relE]
				kept = append(kept, newText)
			}
			*pos += l

		case ast.Container:
			// Recurse into children
			children := t.GetChildren()
			newChildren := filterNodes(children, start, end, pos)

			if len(newChildren) > 0 {
				// We need to create a new container of the same type or mutate the existing one.
				// Since ParseString returns a fresh tree, we can mutate 't' directly,
				// but we must be careful if we are sharing nodes (we aren't).
				// However, 't' in the loop is a pointer to the node in the original tree.
				// If we modify it, we modify the original tree. That's fine here.
				// But simpler is to reuse 't' and set its children.
				setChildren(t, newChildren)
				kept = append(kept, t)
			} else {
				// Debug why empty?
				// fmt.Printf("Container %T empty after filter\n", t)
			}

		case *ast.Image:
			// Images are 0-width text-wise usually, or treated as such.
			// If we are strictly within the range (start <= pos < end), we keep it.
			// Or should we keep it if we just passed start?
			if *pos >= start && *pos < end {
				kept = append(kept, t)
			}

		case *ast.Code:
			// Code block treated as text for length purposes?
			// Usually code blocks have visible content.
			l := len(t.Value)
			s := max(*pos, start)
			e := min(*pos+l, end)

			if s < e {
				relS := s - *pos
				relE := e - *pos
				t.Value = t.Value[relS:relE]
				kept = append(kept, t)
			}
			*pos += l

		case *ast.HR:
			if *pos >= start && *pos < end {
				kept = append(kept, t)
			}

		default:
			// For other nodes, if they are not containers and not text,
			// we assume they are 0-width visible items?
			// If we don't know them, safe to skip or keep?
			// Best effort: skip unknown.
		}
	}
	return kept
}

func setChildren(c ast.Container, children []ast.Node) {
	switch t := c.(type) {
	case *ast.Root:
		t.Children = children
	case *ast.Bold:
		t.Children = children
	case *ast.Italic:
		t.Children = children
	case *ast.Underline:
		t.Children = children
	case *ast.Sup:
		t.Children = children
	case *ast.Sub:
		t.Children = children
	case *ast.Link:
		t.Children = children
	case *ast.Quote:
		t.Children = children
	case *ast.QuoteOf:
		t.Children = children
	case *ast.Spoiler:
		t.Children = children
	case *ast.Indent:
		t.Children = children
	case *ast.Custom:
		t.Children = children
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// normaliseSimpleBB is kept for backward compatibility if used elsewhere,
// though it looks like it was helper for the old substring.
func normaliseSimpleBB(in string) string {
	if len(in) == 0 {
		return in
	}
	var out strings.Builder
	for i := 0; i < len(in); {
		if in[i] == '\\' && i+1 < len(in) {
			out.WriteByte(in[i])
			out.WriteByte(in[i+1])
			i += 2
			continue
		}
		if strings.HasPrefix(in[i:], "[b]") {
			out.WriteString("[b")
			i += 3
			continue
		}
		if strings.HasPrefix(in[i:], "[/b]") {
			out.WriteByte(']')
			i += 4
			continue
		}
		out.WriteByte(in[i])
		i++
	}
	return out.String()
}
