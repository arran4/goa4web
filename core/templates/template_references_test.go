package templates_test

import (
	"fmt"
	"html/template"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	ttext "text/template"
	"text/template/parse"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
)

func TestAllTemplateReferencesAreSatisfied(t *testing.T) {
	cd := common.CoreData{}
	r := httptest.NewRequest("GET", "/", nil)
	f := cd.Funcs(r)
	// Add notification helper function "lower" which is normally provided by the notifier
	f["lower"] = func(s string) string { return s }

	t.Run("site html", func(t *testing.T) {
		tpl := templates.GetCompiledSiteTemplates(f, templates.WithSilence(true))
		assertAllRefsSatisfiedHTML(t, tpl)
	})

	t.Run("email html", func(t *testing.T) {
		tpl := templates.GetCompiledEmailHtmlTemplates(f, templates.WithSilence(true))
		assertAllRefsSatisfiedHTML(t, tpl)
	})

	t.Run("notifications text", func(t *testing.T) {
		tpl := templates.GetCompiledNotificationTemplates(f, templates.WithSilence(true))
		assertAllRefsSatisfiedText(t, tpl)
	})

	t.Run("email text", func(t *testing.T) {
		tpl := templates.GetCompiledEmailTextTemplates(f, templates.WithSilence(true))
		assertAllRefsSatisfiedText(t, tpl)
	})
}

func assertAllRefsSatisfiedHTML(t *testing.T, root *template.Template) {
	t.Helper()
	assertAllRefsSatisfiedGeneric(t, root.Templates(), func(tt *template.Template) *parse.Tree {
		return tt.Tree
	})
}

func assertAllRefsSatisfiedText(t *testing.T, root *ttext.Template) {
	t.Helper()
	assertAllRefsSatisfiedGeneric(t, root.Templates(), func(tt *ttext.Template) *parse.Tree {
		return tt.Tree
	})
}

func assertAllRefsSatisfiedGeneric[T any](
	t *testing.T,
	templates []T,
	treeOf func(T) *parse.Tree,
) {
	t.Helper()

	defined := map[string]struct{}{}
	for _, tt := range templates {
		name := getName(tt)
		defined[name] = struct{}{}
	}

	type miss struct {
		From string
		Ref  string
	}

	var missing []miss

	for _, tt := range templates {
		fromName := getName(tt)
		tr := treeOf(tt)
		if tr == nil || tr.Root == nil {
			continue
		}
		refs := collectTemplateRefs(tr.Root)
		for ref := range refs {
			if _, ok := defined[ref]; !ok {
				missing = append(missing, miss{From: fromName, Ref: ref})
			}
		}
	}

	if len(missing) == 0 {
		return
	}

	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Ref == missing[j].Ref {
			return missing[i].From < missing[j].From
		}
		return missing[i].Ref < missing[j].Ref
	})

	msg := "unresolved template references:\n"
	for _, m := range missing {
		msg += fmt.Sprintf("  %q references missing %q\n", m.From, m.Ref)
	}
	t.Fatal(msg)
}

func getName[T any](tt T) string {
	// Both html/template.Template and text/template.Template have Name() string
	v := reflect.ValueOf(tt)
	m := v.MethodByName("Name")
	if !m.IsValid() {
		return "<unknown>"
	}
	out := m.Call(nil)
	if len(out) != 1 {
		return "<unknown>"
	}
	if s, ok := out[0].Interface().(string); ok {
		return s
	}
	return "<unknown>"
}

// collectTemplateRefs returns a set of referenced template names found via {{template "x"}} and {{block "x"}}.
func collectTemplateRefs(n parse.Node) map[string]struct{} {
	refs := map[string]struct{}{}
	walkNode(n, func(node parse.Node) {
		switch nn := node.(type) {
		case *parse.TemplateNode:
			// {{template "name" .}} becomes TemplateNode with Name="name"
			if nn != nil && nn.Name != "" {
				refs[nn.Name] = struct{}{}
			}
		default:
			// Handle {{block "name" .}} ... {{end}} across Go versions without directly depending on parse.BlockNode.
			// If it's a BlockNode-like value, read its Name field via reflection.
			rv := reflect.ValueOf(node)
			if rv.IsValid() && rv.Kind() == reflect.Ptr && !rv.IsNil() {
				rt := rv.Elem().Type()
				if rt.Name() == "BlockNode" && rt.PkgPath() == "text/template/parse" {
					nameField := rv.Elem().FieldByName("Name")
					if nameField.IsValid() && nameField.Kind() == reflect.String {
						name := nameField.String()
						if name != "" {
							refs[name] = struct{}{}
						}
					}
				}
			}
		}
	})
	return refs
}

func walkNode(n parse.Node, visit func(parse.Node)) {
	if n == nil {
		return
	}
	visit(n)

	switch nn := n.(type) {
	case *parse.ListNode:
		if nn == nil {
			return
		}
		for _, child := range nn.Nodes {
			walkNode(child, visit)
		}

	case *parse.ActionNode:
		// pipes don’t contain template refs; nothing to recurse into

	case *parse.RangeNode:
		if nn == nil {
			return
		}
		walkNode(nn.List, visit)
		walkNode(nn.ElseList, visit)

	case *parse.IfNode:
		if nn == nil {
			return
		}
		walkNode(nn.List, visit)
		walkNode(nn.ElseList, visit)

	case *parse.WithNode:
		if nn == nil {
			return
		}
		walkNode(nn.List, visit)
		walkNode(nn.ElseList, visit)

	default:
		// Reflective recursion for BlockNode body (so refs inside the block default body are also checked).
		rv := reflect.ValueOf(n)
		if rv.IsValid() && rv.Kind() == reflect.Ptr && !rv.IsNil() {
			ev := rv.Elem()
			et := ev.Type()
			if et.Name() == "BlockNode" && et.PkgPath() == "text/template/parse" {
				list := ev.FieldByName("List")
				elseList := ev.FieldByName("ElseList")
				if list.IsValid() && list.CanInterface() {
					if ln, ok := list.Interface().(*parse.ListNode); ok {
						walkNode(ln, visit)
					}
				}
				if elseList.IsValid() && elseList.CanInterface() {
					if ln, ok := elseList.Interface().(*parse.ListNode); ok {
						walkNode(ln, visit)
					}
				}
			}
		}
	}
}
