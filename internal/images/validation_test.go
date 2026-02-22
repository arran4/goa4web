package images

import "testing"

func TestCleanExtension(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{name: "photo.JPG", want: ".jpg"},
		{name: "graphic.png", want: ".png"},
		{name: "animation.gif", want: ".gif"},
		{name: "invalid", wantErr: true},
		{name: "malicious.php", wantErr: true},
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got, err := CleanExtension(tc.name)
			if (err != nil) != tc.wantErr {
				t.Fatalf("CleanExtension(%s) error = %v, wantErr %v", tc.name, err, tc.wantErr)
			}
			if err == nil && got != tc.want {
				t.Fatalf("CleanExtension(%s) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestValidID(t *testing.T) {
	tests := map[string]bool{
		"abcd":            true,
		"1234.png":        true,
		"bad/":            false,
		"...":             false,
		"a":               false,
		"ab":              false,
		"abc":             false,
		"file.name.jpg":   false,
		"1.2.3":           false,
		"valid_thumb.jpg": true,
	}

	for id, want := range tests {
		if got := ValidID(id); got != want {
			t.Fatalf("ValidID(%s)=%v, want %v", id, got, want)
		}
	}
}
