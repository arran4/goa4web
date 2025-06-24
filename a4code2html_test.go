package goa4web

import (
	"github.com/google/go-cmp/cmp"
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
