package a4code2html

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"io"
	"testing"
)

func TestA4code2html_Process(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Test case 1 faq",
			input: "Really just \\[b text\\] for bold, i for italic and u for under line. Links and URL's are also possible:\n\\[image <url>\\] and \\[link <url> <name>\\]",
			want:  "Really just [b text] for bold, i for italic and u for under line. Links and URL's are also possible:<br />\n[image &lt;url&gt;] and [link &lt;url&gt; &lt;name&gt;]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewA4Code2HTML()
			c.input = tt.input
			c.Process()
			got := c.Output()
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("preprocessBookmarks() = diff\n%s", diff)
			}
		})
	}
}

func TestA4code2htmlEscape(t *testing.T) {
	c := NewA4Code2HTML()
	if got := c.Escape('&'); got != "&amp;" {
		t.Errorf("amp %q", got)
	}
	if got := c.Escape('<'); got != "&lt;" {
		t.Errorf("lt %q", got)
	}
	if got := c.Escape('>'); got != "&gt;" {
		t.Errorf("gt %q", got)
	}
	if got := c.Escape('\n'); got != "<br />\n" {
		t.Errorf("newline %q", got)
	}
	c.CodeType = CTTagStrip
	if got := c.Escape('\n'); got != "\n" {
		t.Errorf("tagstrip %q", got)
	}
	c.CodeType = CTWordsOnly
	if got := c.Escape('&'); got != " " {
		t.Errorf("wordsonly %q", got)
	}
}

func TestGetNext(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "abc]"
	got := c.getNext(false)
	if got != "abc" {
		t.Fatalf("got %q", got)
	}
	if c.input != "" {
		t.Fatalf("remaining %q", c.input)
	}
}

func TestDirectOutput(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "fooENDbar"
	c.directOutput("END")
	if got := c.output.String(); got != "foo" {
		t.Fatalf("got %q", got)
	}
	if c.input != "bar" {
		t.Fatalf("input %q", c.input)
	}
}

func TestNextcommSimple(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[*]"
	c.Process()
	if got := c.Output(); got != "<strong></strong>" {
		t.Fatalf("output %q", got)
	}
}

func TestGetNextSpecialChars(t *testing.T) {
	specials := []byte{'*', '/', '_'}
	for _, ch := range specials {
		c := NewA4Code2HTML()
		c.input = string(ch) + "]"
		if got := c.getNext(true); got != string(ch) || c.input != "" {
			t.Fatalf("char %q got %q remaining %q", ch, got, c.input)
		}

		c = NewA4Code2HTML()
		c.input = "\\" + string(ch) + "]"
		if got := c.getNext(false); got != string(ch) || c.input != "" {
			t.Fatalf("escape %q got %q remaining %q", ch, got, c.input)
		}
	}
}

func TestProcessSpecialCommands(t *testing.T) {
	tests := []struct {
		cmd  string
		want string
	}{
		{"*", "<strong>text</strong>"},
		{"/", "<i>text</i>"},
		{"_", "<u>text</u>"},
	}
	for _, tt := range tests {
		c := NewA4Code2HTML()
		c.input = "[" + tt.cmd + " text]"
		c.Process()
		if got := c.Output(); got != tt.want {
			t.Fatalf("cmd %q got %q", tt.cmd, got)
		}
	}
}

func TestProcessEscapedSpecialChars(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "start \\* mid \\/ end \\_"
	c.Process()
	if got := c.Output(); got != "start * mid / end _" {
		t.Fatalf("got %q", got)
	}
}

func TestA4code2htmlComplex(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[b Bold [i Italic]] plain [link http://x example]"
	c.Process()
	want := "<strong>Bold <i>Italic</i></strong> plain <a href=\"http://x\" target=\"_BLANK\">example</a>"
	if got := c.Output(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestA4code2htmlUnclosed(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[b bold"
	c.Process()
	want := "<strong>bold</strong>"
	if got := c.Output(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestA4code2htmlBadURL(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[link javascript:alert(1) example]"
	c.Process()
	want := "javascript:alert(1)example"
	if got := c.Output(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestSpoiler(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[Spoiler secret]"
	c.Process()
	want := "<span onmouseover=\"this.style.color='#FFFFFF';\" onmouseout=\"this.style.color='#000000';\" style=\"color:#000000;background:#000000;\">secret</span>"
	if got := c.Output(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestCodeSlashClose(t *testing.T) {
	c := NewA4Code2HTML()
	c.input = "[code]foo[/code]"
	c.Process()
	want := "<table width=90% align=center bgcolor=lightblue><tr><th>Code: <tr><td><pre>foo</pre></table>"
	if got := c.Output(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestProcessReader(t *testing.T) {
	in := bytes.NewBufferString("[*]")
	out := new(bytes.Buffer)
	c := NewA4Code2HTML()
	if err := c.ProcessReader(in, out); err != nil {
		t.Fatalf("ProcessReader error: %v", err)
	}
	if got := out.String(); got != "<strong></strong>" {
		t.Fatalf("got %q", got)
	}
}

type slowReader struct {
	data string
	i    int
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.i >= len(s.data) {
		return 0, io.EOF
	}
	p[0] = s.data[s.i]
	s.i++
	return 1, nil
}

func TestProcessReaderStreaming(t *testing.T) {
	sr := &slowReader{data: "[*]"}
	out := new(bytes.Buffer)
	c := NewA4Code2HTML()
	if err := c.ProcessReader(sr, out); err != nil {
		t.Fatalf("ProcessReader error: %v", err)
	}
	if got := out.String(); got != "<strong></strong>" {
		t.Fatalf("got %q", got)
	}
}
