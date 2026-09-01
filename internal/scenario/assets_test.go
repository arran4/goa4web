package scenario

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

func TestValidateAsset(t *testing.T) {
	fsys := fstest.MapFS{
		"assets/welcome.jpg":    &fstest.MapFile{Data: []byte("jpg")},
		"assets/nested/map.png": &fstest.MapFile{Data: []byte("png")},
	}

	tests := []struct {
		name      string
		path      string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "valid relative asset",
			path:    "assets/welcome.jpg",
			wantErr: false,
		},
		{
			name:    "valid nested asset",
			path:    "assets/nested/map.png",
			wantErr: false,
		},
		{
			name:      "missing asset file",
			path:      "assets/missing.jpg",
			wantErr:   true,
			errSubstr: "asset not found",
		},
		{
			name:      "escape via parent traversal",
			path:      "../secret.txt",
			wantErr:   true,
			errSubstr: "escapes scenario directory",
		},
		{
			name:      "escape via nested parent traversal",
			path:      "assets/../../secret.txt",
			wantErr:   true,
			errSubstr: "escapes scenario directory",
		},
		{
			name:      "absolute path rejected",
			path:      "/etc/passwd",
			wantErr:   true,
			errSubstr: "escapes scenario directory",
		},
		{
			name:      "empty path rejected",
			path:      "",
			wantErr:   true,
			errSubstr: "empty asset path",
		},
		{
			name:      "directory path rejected",
			path:      "assets",
			wantErr:   true,
			errSubstr: "is a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateAsset(fsys, tt.path)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateAsset(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
			if tt.wantErr && tt.errSubstr != "" {
				if err == nil || !errors.Is(err, ErrAssetEscape) && !containsStr(err.Error(), tt.errSubstr) {
					t.Fatalf("ValidateAsset(%q) error = %v, want substring %q", tt.path, err, tt.errSubstr)
				}
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return strings.Contains(s, substr)
}
