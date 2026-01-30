package a4code2html

import (
	"bytes"
	"io"
	"testing"

	"github.com/arran4/goa4web/a4code"
	"github.com/google/go-cmp/cmp"
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
			c := New()
			c.SetInput(tt.input)
			gotBytes, _ := io.ReadAll(c.Process())
			got := string(gotBytes)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("preprocessBookmarks() = diff\n%s", diff)
			}
		})
	}
}

func TestA4code2htmlEscape(t *testing.T) {
	c := New()
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
		c := New()
		c.SetInput("[" + tt.cmd + " text]")
		got, _ := io.ReadAll(c.Process())
		if string(got) != tt.want {
			t.Fatalf("cmd %q got %q", tt.cmd, got)
		}
	}
}

func TestProcessEscapedSpecialChars(t *testing.T) {
	c := New()
	c.SetInput("start \\* mid \\/ end \\_")
	got, _ := io.ReadAll(c.Process())
	if string(got) != "start * mid / end _" {
		t.Fatalf("got %q", string(got))
	}
}

func TestA4code2htmlComplex(t *testing.T) {
	c := New()
	c.SetInput("[b Bold [i Italic]] plain [link http://x example]")
	got, _ := io.ReadAll(c.Process())
	want := "<strong>Bold <i>Italic</i></strong> plain <a href=\"http://x\" target=\"_blank\"> example</a>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestA4code2htmlUnclosed(t *testing.T) {
	c := New()
	c.SetInput("[b bold")
	got, _ := io.ReadAll(c.Process())
	want := "<strong>bold</strong>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestA4code2htmlBadURL(t *testing.T) {
	c := New()
	c.SetInput("[link javascript:alert(1) example]")
	got, _ := io.ReadAll(c.Process())
	want := "javascript:alert(1) example"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestSpoiler(t *testing.T) {
	c := New()
	c.SetInput("[Spoiler secret]")
	got, _ := io.ReadAll(c.Process())
	want := "<span class=\"spoiler\">secret</span>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestCodeSlashClose(t *testing.T) {
	c := New()
	c.SetInput("[code]foo[/code]")
	got, _ := io.ReadAll(c.Process())
	want := "<div class=\"a4code-block a4code-code-wrapper\"><div class=\"code-header\">Code</div><pre class=\"a4code-code-body\">]foo</pre></div>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestQuoteMarkup(t *testing.T) {
	c := New()
	c.SetInput("[quote hi]")
	got, _ := io.ReadAll(c.Process())
	want := "<blockquote class=\"a4code-block a4code-quote quote-color-0\"><div class=\"quote-header\">Quote:</div><div class=\"quote-body\">hi</div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestQuoteOfMarkup(t *testing.T) {
	c := New()
	c.SetInput("[quoteof bob hi]")
	got, _ := io.ReadAll(c.Process())
	want := "<blockquote class=\"a4code-block a4code-quoteof quote-color-0\"><div class=\"quote-header\">Quote of bob:</div><div class=\"quote-body\"> hi</div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestQuoteOfColorMapping(t *testing.T) {
	colorMap := func(name string) string {
		return "mapped-color"
	}
	c := New(colorMap)
	c.SetInput("[quoteof bob hi]")
	got, _ := io.ReadAll(c.Process())
	want := "<blockquote class=\"a4code-block a4code-quoteof mapped-color quote-color-0\"><div class=\"quote-header\">Quote of bob:</div><div class=\"quote-body\"> hi</div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestNestedQuotes(t *testing.T) {
	c := New()
	c.SetInput("[quote 0 [quote 1 [quote 2]]]")
	got, _ := io.ReadAll(c.Process())
	want := "<blockquote class=\"a4code-block a4code-quote quote-color-0\"><div class=\"quote-header\">Quote:</div><div class=\"quote-body\">0 <blockquote class=\"a4code-block a4code-quote quote-color-1\"><div class=\"quote-header\">Quote:</div><div class=\"quote-body\">1 <blockquote class=\"a4code-block a4code-quote quote-color-2\"><div class=\"quote-header\">Quote:</div><div class=\"quote-body\">2</div></blockquote></div></blockquote></div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestIndentMarkup(t *testing.T) {
	c := New()
	c.SetInput("[indent hi]")
	got, _ := io.ReadAll(c.Process())
	want := "<div class=\"a4code-block a4code-indent\"><div>hi</div></div>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestProcessReader(t *testing.T) {
	in := bytes.NewBufferString("[*]")
	out := new(bytes.Buffer)
	c := New()
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
	c := New()
	if err := c.ProcessReader(sr, out); err != nil {
		t.Fatalf("ProcessReader error: %v", err)
	}
	if got := out.String(); got != "<strong></strong>" {
		t.Fatalf("got %q", got)
	}
}

func TestImageURLMapper(t *testing.T) {
	mapper := func(tag, val string) string {
		return "map:" + tag + ":" + val
	}
	c := New(mapper)
	c.SetInput("[img=image:abc]")
	got, _ := io.ReadAll(c.Process())
	want := "<img class=\"a4code-image\" src=\"map:img:image:abc\" />"
	if string(got) != want {
		t.Fatalf("img map got %q want %q", got, want)
	}
}

func TestImageClass(t *testing.T) {
	c := New()
	c.SetInput("[img http://example.com/foo.jpg]")
	got, _ := io.ReadAll(c.Process())
	want := "<img class=\"a4code-image\" src=\"http://example.com/foo.jpg\" />"
	if string(got) != want {
		t.Fatalf("img got %q want %q", got, want)
	}
}

func TestHrTag(t *testing.T) {
	tests := []struct {
		name     string
		codeType CodeType
		want     string
	}{
		{"HTML", CTHTML, "<hr /><br />\n"},
		{"TagStrip", CTTagStrip, "\n"},
		{"WordsOnly", CTWordsOnly, " "},
		{"TableOfContents", CTTableOfContents, "<br />\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New()
			c.CodeType = tt.codeType
			c.SetInput("[hr]\n")
			got, _ := io.ReadAll(c.Process())
			if string(got) != tt.want {
				t.Errorf("got %q want %q", string(got), tt.want)
			}
		})
	}
}

func TestExternalLinkCard(t *testing.T) {
	provider := func(u string) *a4code.LinkMetadata {
		if u == "http://example.com/card" {
			return &a4code.LinkMetadata{
				Title:       "Example Title",
				Description: "Example Description",
				ImageURL:    "http://example.com/image.jpg",
			}
		}
		if u == "http://example.com/simple" {
			return &a4code.LinkMetadata{
				Title:       "Simple Title",
				Description: "Simple Description",
				ImageURL:    "http://example.com/image.jpg",
			}
		}
		return nil
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Inline with metadata (Title injection)",
			input: "See [link http://example.com/card]",
			want:  "See <a href=\"http://example.com/card\" target=\"_blank\" title=\"Example Description\">Example Title</a>",
		},
		{
			name:  "Inline without metadata",
			input: "See [link http://example.com/none]",
			want:  "See <a href=\"http://example.com/none\" target=\"_blank\">http://example.com/none</a>",
		},
		{
			name:  "Inline with explicit title (ignore metadata title)",
			input: "See [link http://example.com/card Explicit]",
			want:  "See <a href=\"http://example.com/card\" target=\"_blank\" title=\"Example Title - Example Description\"> Explicit</a>",
		},
		{
			name:  "Block link [link url] with metadata (Complex Card)",
			input: "[link http://example.com/card]\n",
			want:  "<div class=\"external-link-card\"><a href=\"http://example.com/card\" target=\"_blank\" class=\"external-link-card-inner\"><img src=\"http://example.com/image.jpg\" class=\"external-link-image\" /><div class=\"external-link-content\"><div class=\"external-link-title\">Example Title</div><div class=\"external-link-description\">Example Description</div></div></a></div>",
		},
		{
			name:  "Block link [link url] without metadata -> Inline",
			input: "[link http://example.com/none]\n",
			want:  "<a href=\"http://example.com/none\" target=\"_blank\">http://example.com/none</a><br />\n",
		},
		{
			name:  "Block link [link url Title] with metadata -> Inline with Tooltip",
			input: "[link http://example.com/simple Simple!]\n",
			want:  "<a href=\"http://example.com/simple\" target=\"_blank\" title=\"Simple Title - Simple Description\"> Simple!</a><br />\n",
		},
		{
			name:  "Consecutive Block Links",
			input: "[link http://example.com/card]\n[link http://example.com/card]\n",
			want: "<div class=\"external-link-card\"><a href=\"http://example.com/card\" target=\"_blank\" class=\"external-link-card-inner\"><img src=\"http://example.com/image.jpg\" class=\"external-link-image\" /><div class=\"external-link-content\"><div class=\"external-link-title\">Example Title</div><div class=\"external-link-description\">Example Description</div></div></a></div>" +
				"<div class=\"external-link-card\"><a href=\"http://example.com/card\" target=\"_blank\" class=\"external-link-card-inner\"><img src=\"http://example.com/image.jpg\" class=\"external-link-image\" /><div class=\"external-link-content\"><div class=\"external-link-title\">Example Title</div><div class=\"external-link-description\">Example Description</div></div></a></div>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(a4code.LinkMetadataProvider(provider))
			c.SetInput(tt.input)
			gotBytes, _ := io.ReadAll(c.Process())
			got := string(gotBytes)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("%s: diff\n%s", tt.name, diff)
			}
		})
	}
}
