package a4code

import (
	"testing"
)

func TestToText_Code(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Inline code",
			input: "Start [code func main() {}] End",
			want:  "Start func main() {} End",
		},
		{
			name:  "Block code",
			input: "[code\nfunc main() {}\n]",
			want:  "func main() {}\n",
		},
		{
			name:  "CodeIn",
			input: "[codein \"go\" func main() {}]",
			want:  "func main() {}",
		},
        {
            name: "Code with brackets",
            input: "[code [b]bold[/b]]",
            want: "[b]bold[/b]",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString error: %v", err)
			}
			got := ToText(root)
			if got != tt.want {
				t.Errorf("ToText() = %q, want %q", got, tt.want)
			}

            // Also check ToCleanText
            clean := ToCleanText(root)
            if clean != tt.want {
                t.Errorf("ToCleanText() = %q, want %q", clean, tt.want)
            }
		})
	}
}
