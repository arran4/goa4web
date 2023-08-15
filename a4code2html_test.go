package main

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
		{
			name:  "Causes crash",
			input: "Word of mouth. But that doesn't explain it's \"fans.\"\n\nhttp://en.wikipedia.org/wiki/Robotech\n[q Carl Macek attempted another sequel with the development of Robotech 3000. This all-CGI series would have been set a millennium in the future of the Robotech universe and feature none of the old series' characters. In the three-minute trailer, an expedition is sent to check on a non-responsive mining outpost and is attacked by \"infected\" Veritech mecha. Again, the idea was abandoned midway into production after negative reception within the company, [b negative fan reactions] at FanimeCon and San Diego Comic-Con in 2000, and financial difficulties within Netter Digital who was animating the show. It now exists only in trailer form on the official Robotech website.]\n\nEmphasis mine.\n\nAnyway, for something to achieve a fan base it should be rather significant. However that really doesn't [a http://en.wikipedia.org/wiki/List_of_movie_clich",
			want:  "",
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
