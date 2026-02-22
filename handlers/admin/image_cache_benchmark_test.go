package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkListImageCacheEntries(b *testing.B) {
	// Setup: create a temp dir with many files
	dir, err := os.MkdirTemp("", "image_cache_bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create 1000 files
	for i := 0; i < 1000; i++ {
		filename := fmt.Sprintf("image_%d.jpg", i)
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte("dummy content"), 0644); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := listImageCacheEntries(dir)
		if err != nil {
			b.Fatal(err)
		}
	}
}
