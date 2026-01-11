package images

import (
	"fmt"
	"path/filepath"
	"strings"
)

// minImageIDLen enforces the minimum length required to build upload paths safely.
const minImageIDLen = 4

var allowedExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
}

// AllowedExtension reports whether ext is a permitted image extension.
func AllowedExtension(ext string) bool {
	_, ok := allowedExtensions[strings.ToLower(ext)]
	return ok
}

// CleanExtension extracts and validates the extension from name.
func CleanExtension(name string) (string, error) {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "", fmt.Errorf("missing extension")
	}
	if !AllowedExtension(ext) {
		return "", fmt.Errorf("unsupported image extension: %s", ext)
	}
	return ext, nil
}

// ValidID reports whether s is long enough and contains only safe characters.
func ValidID(s string) bool {
	if len(s) < minImageIDLen || s == "." || s == ".." {
		return false
	}
	dotCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || c == '-') {
			if c == '.' && dotCount == 0 {
				dotCount++
			} else {
				return false
			}
		}
	}
	return true
}
