package a4code2html

import (
	"bytes"
	"fmt"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/google/go-cmp/cmp"
	"io"
	"strings"
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
			c := New()
			c.SetInput(tt.input)
			gotBytes := testhelpers.Must(io.ReadAll(c.Process()))
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
		got := testhelpers.Must(io.ReadAll(c.Process()))
		if string(got) != tt.want {
			t.Fatalf("cmd %q got %q", tt.cmd, got)
		}
	}
}

func TestProcessEscapedSpecialChars(t *testing.T) {
	c := New()
	c.SetInput("start \\* mid \\/ end \\_")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	if string(got) != "start * mid / end _" {
		t.Fatalf("got %q", string(got))
	}
}

func TestA4code2htmlComplex(t *testing.T) {
	c := New()
	c.SetInput("[b Bold [i Italic]] plain [link http://x example]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<strong>Bold <i>Italic</i></strong> plain <a href=\"http://x\" target=\"_blank\"> example</a>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestA4code2htmlUnclosed(t *testing.T) {
	c := New()
	c.SetInput("[b bold")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<strong>bold</strong>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestA4code2htmlBadURL(t *testing.T) {
	c := New()
	c.SetInput("[link javascript:alert(1) example]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "javascript:alert(1) example"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestSpoiler(t *testing.T) {
	c := New()
	c.SetInput("[Spoiler secret]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<span class=\"spoiler\">secret</span>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestQuoteMarkup(t *testing.T) {
	c := New()
	c.SetInput("[quote hi]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<blockquote class=\"a4code-block a4code-quote quote-color-0\"><div class=\"quote-header\">Quote:</div><div class=\"quote-body\">hi</div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestQuoteOfMarkup(t *testing.T) {
	c := New()
	c.SetInput("[quoteof bob hi]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
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
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<blockquote class=\"a4code-block a4code-quoteof mapped-color quote-color-0\"><div class=\"quote-header\">Quote of bob:</div><div class=\"quote-body\"> hi</div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestNestedQuotes(t *testing.T) {
	c := New()
	c.SetInput("[quote 0 [quote 1 [quote 2]]]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	// [quote 0] is at line start -> Block.
	// [quote 1] and [quote 2] are not at line start -> Inline <q>.
	want := "<blockquote class=\"a4code-block a4code-quote quote-color-0\"><div class=\"quote-header\">Quote:</div><div class=\"quote-body\">0 <q class=\"a4code-inline a4code-quote\">1 <q class=\"a4code-inline a4code-quote\">2</q></q></div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestIndentMarkup(t *testing.T) {
	c := New()
	c.SetInput("[indent hi]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
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
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<img class=\"a4code-image\" src=\"map:img:image:abc\" />"
	if string(got) != want {
		t.Fatalf("img map got %q want %q", got, want)
	}
}

func TestImageClass(t *testing.T) {
	c := New()
	c.SetInput("[img http://example.com/foo.jpg]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
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
			got := testhelpers.Must(io.ReadAll(c.Process()))
			if string(got) != tt.want {
				t.Errorf("got %q want %q", string(got), tt.want)
			}
		})
	}
}

// legacyTestLinkProvider implements LinkProvider to simulate the old MetadataProvider logic for tests
type legacyTestLinkProvider struct {
	metadata map[string]struct {
		Title       string
		Description string
		ImageURL    string
	}
}

func (p *legacyTestLinkProvider) RenderLink(url string, isBlock bool, isImmediateClose bool) (string, string, bool) {
	safe, ok := SanitizeURL(url)
	if !ok {
		return url, "", false
	}
	meta, hasMeta := p.metadata[url]
	if isBlock && hasMeta {
		// Render Card
		imageHTML := ""
		if meta.ImageURL != "" {
			safeImg, imgOk := SanitizeURL(meta.ImageURL)
			if imgOk {
				imageHTML = fmt.Sprintf("<img src=\"%s\" class=\"external-link-image\" />", safeImg)
			}
		}

		if isImmediateClose {
			// Complex Card: Title and Description from Metadata
			return fmt.Sprintf(
				"<div class=\"external-link-card\"><a href=\"%s\" target=\"_blank\" class=\"external-link-card-inner\">%s<div class=\"external-link-content\"><div class=\"external-link-title\">%s</div><div class=\"external-link-description\">%s</div></div></a></div>",
				safe, imageHTML, meta.Title, meta.Description), "", true
		} else {
			// Simple Card: Title from user provided text (consumed later)
			return fmt.Sprintf(
				"<div class=\"external-link-card\"><a href=\"%s\" target=\"_blank\" class=\"external-link-card-inner\">%s<div class=\"external-link-content\"><div class=\"external-link-title\">",
				safe, imageHTML), "</div></div></a></div>", false
		}
	}

	// Inline Link or Block fallback (no metadata)
	if isImmediateClose {
		text := url
		if hasMeta && meta.Title != "" {
			text = meta.Title
		}
		// Return consumeImmediate=false so the parser handles the closing ] and subsequent newline
		return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s", safe, text), "</a>", false
	}

	return fmt.Sprintf("<a href=\"%s\" target=\"_blank\">", safe), "</a>", false
}

func (p *legacyTestLinkProvider) MapImageURL(tag, val string) string {
	return val
}

func TestExternalLinkCard(t *testing.T) {
	provider := &legacyTestLinkProvider{
		metadata: map[string]struct {
			Title       string
			Description string
			ImageURL    string
		}{
			"http://example.com/card": {
				Title:       "Example Title",
				Description: "Example Description",
				ImageURL:    "http://example.com/image.jpg",
			},
			"http://example.com/simple": {
				Title:       "Simple Title",
				Description: "Simple Description",
				ImageURL:    "http://example.com/image.jpg",
			},
		},
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Inline with metadata (Title injection)",
			input: "See [link http://example.com/card]",
			want:  "See <a href=\"http://example.com/card\" target=\"_blank\">Example Title</a>",
		},
		{
			name:  "Inline without metadata",
			input: "See [link http://example.com/none]",
			want:  "See <a href=\"http://example.com/none\" target=\"_blank\">http://example.com/none</a>",
		},
		{
			name:  "Inline with explicit title (ignore metadata title)",
			input: "See [link http://example.com/card Explicit]",
			want:  "See <a href=\"http://example.com/card\" target=\"_blank\"> Explicit</a>",
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
			name:  "Block link [link url Title] with metadata -> Simple Card",
			input: "[link http://example.com/simple Simple!]\n",
			want:  "<div class=\"external-link-card\"><a href=\"http://example.com/simple\" target=\"_blank\" class=\"external-link-card-inner\"><img src=\"http://example.com/image.jpg\" class=\"external-link-image\" /><div class=\"external-link-content\"><div class=\"external-link-title\"> Simple!</div></div></a></div><br />\n",
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
			c := New(LinkProvider(provider))
			c.SetInput(tt.input)
			gotBytes := testhelpers.Must(io.ReadAll(c.Process()))
			got := string(gotBytes)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("%s: diff\n%s", tt.name, diff)
			}
		})
	}
}

type mockProvider struct{}

func (m *mockProvider) RenderLink(url string, isBlock bool, isImmediateClose bool) (string, string, bool) {
	if url == "http://custom.com" {
		return "<custom-link>", "", true
	}
	if isBlock {
		return "<block-link>", "</block-link>", false
	}
	return "<inline-link>", "</inline-link>", false
}

func (m *mockProvider) MapImageURL(tag, val string) string {
	return "mapped:" + val
}

func TestLinkProvider(t *testing.T) {
	c := New(&mockProvider{})

	// Case 1: Custom handling (consumed immediate)
	c.SetInput("[link http://custom.com]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<custom-link>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}

	// Case 2: Block link
	c.SetInput("[link http://other.com]\n")
	got = testhelpers.Must(io.ReadAll(c.Process()))
	want = "<block-link></block-link><br />\n"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}

	// Case 3: Image mapping
	c.SetInput("[img foo.jpg]")
	got = testhelpers.Must(io.ReadAll(c.Process()))
	want = "<img class=\"a4code-image\" src=\"mapped:foo.jpg\" />"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestA4code2htmlInlineCode(t *testing.T) {
	c := New()
	c.SetInput("text [code inline] text")
	gotBytes, _ := io.ReadAll(c.Process())
	got := string(gotBytes)
	want := "text <code class=\"a4code-inline a4code-code\">inline</code> text"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestA4code2htmlBlockCode(t *testing.T) {
	c := New()
	c.SetInput("[code \nblock\n]")
	gotBytes, _ := io.ReadAll(c.Process())
	got := string(gotBytes)
	// Check for block wrapper
	if !strings.Contains(got, "a4code-block a4code-code-wrapper") {
		t.Errorf("expected block wrapper, got %q", got)
	}
}

func TestA4code2htmlInlineCodeWithBrackets(t *testing.T) {
	c := New()
	c.SetInput("please use [code [quote\\]] so I know.")
	gotBytes, _ := io.ReadAll(c.Process())
	got := string(gotBytes)
	want := "please use <code class=\"a4code-inline a4code-code\">[quote]</code> so I know."
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestA4code2htmlInlineQuote(t *testing.T) {
	c := New()
	c.SetInput("text [quote inline] text")
	gotBytes, _ := io.ReadAll(c.Process())
	got := string(gotBytes)
	if !strings.Contains(got, "<q class=\"a4code-inline a4code-quote\">") {
		t.Errorf("expected inline quote <q>, got %q", got)
	}
}

func TestA4code2htmlBlockQuote(t *testing.T) {
	c := New()
	c.SetInput("[quote \nblock\n]")
	gotBytes, _ := io.ReadAll(c.Process())
	got := string(gotBytes)
	if !strings.Contains(got, "<blockquote class=\"a4code-block a4code-quote") {
		t.Errorf("expected block quote <blockquote>, got %q", got)
	}
}

func TestCodeIn(t *testing.T) {
	c := New()
	c.SetInput("[codein \"go\" func main() {}]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<div class=\"a4code-block a4code-code-wrapper a4code-language-go\"><div class=\"code-header\">Code (go)</div><pre class=\"a4code-code-body\"><code class=\"language-go\">func main() {}</code></pre></div>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestQuoteOfWithQuotes(t *testing.T) {
	c := New()
	c.SetInput("[quoteof \"bob jones\" hi]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	// Should strip quotes from name
	want := "<blockquote class=\"a4code-block a4code-quoteof quote-color-0\"><div class=\"quote-header\">Quote of bob jones:</div><div class=\"quote-body\"> hi</div></blockquote>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}

func TestCodeInWithQuotesAndSpaces(t *testing.T) {
	c := New()
	c.SetInput("[codein \"c++\" int main() {}]")
	got := testhelpers.Must(io.ReadAll(c.Process()))
	want := "<div class=\"a4code-block a4code-code-wrapper a4code-language-c++\"><div class=\"code-header\">Code (c++)</div><pre class=\"a4code-code-body\"><code class=\"language-c++\">int main() {}</code></pre></div>"
	if string(got) != want {
		t.Errorf("got %q want %q", string(got), want)
	}
}
