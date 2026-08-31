package scenario

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// ErrAssetEscape indicates that an asset path attempts to escape the scenario root directory.
var ErrAssetEscape = errors.New("asset path escapes scenario directory")

// ErrAssetNotFound indicates that a referenced asset file does not exist.
type ErrAssetNotFound struct {
	Path string
	Err  error
}

func (e ErrAssetNotFound) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("asset not found: %s: %v", e.Path, e.Err)
	}
	return fmt.Sprintf("asset not found: %s", e.Path)
}

func (e ErrAssetNotFound) Unwrap() error {
	return e.Err
}

// CleanAssetPath verifies that an asset path is relative and does not escape the scenario root.
// It converts Windows-style backslashes to forward slashes and returns the cleaned relative path.
func CleanAssetPath(assetPath string) (string, error) {
	if assetPath == "" {
		return "", errors.New("empty asset path")
	}

	// Normalize backslashes to forward slashes.
	p := strings.ReplaceAll(assetPath, "\\", "/")

	// Reject absolute paths.
	if strings.HasPrefix(p, "/") || (len(p) > 1 && p[1] == ':') {
		return "", fmt.Errorf("%w: %s (absolute paths not allowed)", ErrAssetEscape, assetPath)
	}

	cleaned := path.Clean(p)
	if cleaned == "." || cleaned == "" {
		return "", errors.New("invalid asset path")
	}

	// Reject path traversal escaping root.
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("%w: %s", ErrAssetEscape, assetPath)
	}

	return cleaned, nil
}

// ValidateAsset verifies that the asset path is safe and exists within the given fs.FS.
func ValidateAsset(fsys fs.FS, assetPath string) (string, error) {
	cleanPath, err := CleanAssetPath(assetPath)
	if err != nil {
		return "", err
	}

	if fsys == nil {
		return cleanPath, nil
	}

	info, err := fs.Stat(fsys, cleanPath)
	if err != nil {
		return "", ErrAssetNotFound{Path: cleanPath, Err: err}
	}
	if info.IsDir() {
		return "", fmt.Errorf("asset path %q is a directory, expected file", cleanPath)
	}

	return cleanPath, nil
}
