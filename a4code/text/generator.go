package text

import (
	"bytes"
	"io"

	"github.com/arran4/goa4web/a4code/ast"
)

type lineTracker interface {
	isStartOfLine() bool
}

type SmartWriter struct {
	w        io.Writer
	lastByte byte
}

func (sw *SmartWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	n, err = sw.w.Write(p)
	if n > 0 {
		sw.lastByte = p[n-1]
	}
	return
}

func (sw *SmartWriter) isStartOfLine() bool {
	return sw.lastByte == '\n'
}

type PrefixWriter struct {
	w         io.Writer
	prefix    []byte
	startLine bool
}

func (pw *PrefixWriter) isStartOfLine() bool {
	return pw.startLine
}

func (pw *PrefixWriter) Write(p []byte) (int, error) {
	written := 0
	for len(p) > 0 {
		if pw.startLine {
			if _, err := pw.w.Write(pw.prefix); err != nil {
				return written, err
			}
			pw.startLine = false
		}

		idx := bytes.IndexByte(p, '\n')
		var toWrite []byte
		if idx == -1 {
			toWrite = p
		} else {
			toWrite = p[:idx+1]
		}

		n, err := pw.w.Write(toWrite)
		written += n
		if err != nil {
			return written, err
		}

		if idx != -1 {
			pw.startLine = true
			p = p[idx+1:]
		} else {
			p = nil
		}
	}
	return written, nil
}

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Root(w io.Writer, n *ast.Root) error {
	sw := &SmartWriter{w: w, lastByte: '\n'}
	for _, c := range n.Children {
		if err := ast.Generate(sw, c, g); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) Text(w io.Writer, t *ast.Text) error {
	io.WriteString(w, t.Value)
	return nil
}

func (g *Generator) visitChildren(w io.Writer, children []ast.Node) error {
	for _, c := range children {
		if err := ast.Generate(w, c, g); err != nil {
			return err
		}
	}
	return nil
}

func (g *Generator) Bold(w io.Writer, n *ast.Bold) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Italic(w io.Writer, n *ast.Italic) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Underline(w io.Writer, n *ast.Underline) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Sup(w io.Writer, n *ast.Sup) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Sub(w io.Writer, n *ast.Sub) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Link(w io.Writer, n *ast.Link) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Image(w io.Writer, n *ast.Image) error {
	return nil
}

func (g *Generator) Code(w io.Writer, n *ast.Code) error {
	return nil
}

func (g *Generator) CodeIn(w io.Writer, n *ast.CodeIn) error {
	return nil
}

func (g *Generator) Quote(w io.Writer, n *ast.Quote) error {
	if lt, ok := w.(lineTracker); ok {
		if !lt.isStartOfLine() {
			io.WriteString(w, "\n")
		}
	}
	pw := &PrefixWriter{w: w, prefix: []byte("> "), startLine: true}
	return g.visitChildren(pw, n.Children)
}

func (g *Generator) QuoteOf(w io.Writer, n *ast.QuoteOf) error {
	if lt, ok := w.(lineTracker); ok {
		if !lt.isStartOfLine() {
			io.WriteString(w, "\n")
		}
	}
	io.WriteString(w, "> "+n.Name+" wrote:\n")
	pw := &PrefixWriter{w: w, prefix: []byte("> "), startLine: true}
	return g.visitChildren(pw, n.Children)
}

func (g *Generator) Spoiler(w io.Writer, n *ast.Spoiler) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) Indent(w io.Writer, n *ast.Indent) error {
	return g.visitChildren(w, n.Children)
}

func (g *Generator) HR(w io.Writer, n *ast.HR) error {
	if lt, ok := w.(lineTracker); ok {
		if !lt.isStartOfLine() {
			io.WriteString(w, "\n")
		}
	}
	io.WriteString(w, "---\n")
	return nil
}

func (g *Generator) Custom(w io.Writer, n *ast.Custom) error {
	return g.visitChildren(w, n.Children)
}
