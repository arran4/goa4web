package ast

import (
	"strings"
)

// DetermineLinkProperties traverses the AST and sets IsBlock and IsImmediateClose on Link nodes.
func DetermineLinkProperties(n Node) {
	if n == nil {
		return
	}
	if p, ok := n.(parent); ok {
		children := *p.childrenPtr()

		isBlockContext := false
		switch n.(type) {
		case *Root, *Quote, *QuoteOf, *Spoiler, *Indent:
			isBlockContext = true
		}

		for i, c := range children {
			if l, ok := c.(*Link); ok {
				if len(l.Children) == 0 {
					l.IsImmediateClose = true
				}

				if isBlockContext {
					// Check previous sibling
					prevIsNewline := false
					if i == 0 {
						prevIsNewline = true
					} else {
						if txt, ok := children[i-1].(*Text); ok {
							if strings.HasSuffix(txt.Value, "\n") {
								prevIsNewline = true
							}
						}
					}

					// Check next sibling
					nextIsNewline := false
					if i == len(children)-1 {
						nextIsNewline = true
					} else {
						if txt, ok := children[i+1].(*Text); ok {
							if strings.HasPrefix(txt.Value, "\n") {
								nextIsNewline = true
							}
						}
					}

					if prevIsNewline && nextIsNewline {
						l.IsBlock = true
					}
				}
			}
			DetermineLinkProperties(c)
		}
	}
}
