package a4code

import (
	"bytes"
	"strings"
)

func substring(s string, start, end int) string {
	if start < 0 || end < start {
		return ""
	}

	root, err := ParseString(s)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	var current int
	var stack []Node

	var traverse func(n Node)
	traverse = func(n Node) {
		if current >= end {
			return
		}

		switch n := n.(type) {
		case *Text:
			if current < end && start < current+len(n.Value) {
				s := max(0, start-current)
				e := min(len(n.Value), end-current)
				if s < e {
					for _, p := range stack {
						buf.WriteString(p.Tag())
					}
					buf.WriteString(n.Value[s:e])
					for i := len(stack) - 1; i >= 0; i-- {
						buf.WriteString(stack[i].EndTag())
					}
				}
			}
			current += len(n.Value)
		case parent:
			stack = append(stack, n)
			for _, child := range n.children() {
				traverse(child)
			}
			stack = stack[:len(stack)-1]
		}
	}

	for _, child := range root.Children {
		traverse(child)
	}

	return strings.TrimSpace(buf.String())
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
