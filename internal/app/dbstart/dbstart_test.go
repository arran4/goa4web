package dbstart

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckUploadDir(t *testing.T) {
	t.Run("S3 Valid", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ImageUploadDir: "s3://bucket/prefix",
		}
		err := CheckUploadDir(cfg)
		assert.Nil(t, err)
	})

	t.Run("S3 Valid Root", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ImageUploadDir: "s3://bucket",
		}
		err := CheckUploadDir(cfg)
		assert.Nil(t, err)
	})

	t.Run("S3 Invalid Missing Bucket", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ImageUploadDir: "s3://",
		}
		err := CheckUploadDir(cfg)
		require.NotNil(t, err)
		assert.Contains(t, err.ErrorMessage, "image upload directory invalid")
	})

	t.Run("S3 Invalid Empty Bucket", func(t *testing.T) {
		cfg := &config.RuntimeConfig{
			ImageUploadDir: "s3:///prefix",
		}
		err := CheckUploadDir(cfg)
		require.NotNil(t, err)
		assert.Contains(t, err.ErrorMessage, "image upload directory invalid")
	})

	t.Run("Filesystem Valid Existing", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := &config.RuntimeConfig{
			ImageUploadDir: tmpDir,
		}
		err := CheckUploadDir(cfg)
		assert.Nil(t, err)
	})

	t.Run("Filesystem Invalid Non-Existent NoCreate", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistent := filepath.Join(tmpDir, "nonexistent")
		cfg := &config.RuntimeConfig{
			ImageUploadDir: nonExistent,
			CreateDirs:     false,
		}
		err := CheckUploadDir(cfg)
		require.NotNil(t, err)
		assert.Contains(t, err.ErrorMessage, "image upload directory invalid")
	})

	t.Run("Filesystem Valid Non-Existent Create", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistent := filepath.Join(tmpDir, "created")
		cfg := &config.RuntimeConfig{
			ImageUploadDir: nonExistent,
			CreateDirs:     true,
		}
		err := CheckUploadDir(cfg)
		assert.Nil(t, err)
		info, statErr := os.Stat(nonExistent)
		assert.Nil(t, statErr)
		assert.True(t, info.IsDir())
	})

	t.Run("ImageCacheDir Valid Existing", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := t.TempDir()
		cfg := &config.RuntimeConfig{
			ImageUploadDir: tmpDir,
			ImageCacheDir:  cacheDir,
		}
		err := CheckUploadDir(cfg)
		assert.Nil(t, err)
	})

	t.Run("ImageCacheDir Invalid Non-Existent NoCreate", func(t *testing.T) {
		tmpDir := t.TempDir()
		cacheDir := filepath.Join(tmpDir, "nonexistent_cache")
		cfg := &config.RuntimeConfig{
			ImageUploadDir: tmpDir,
			ImageCacheDir:  cacheDir,
			CreateDirs:     false,
		}
		err := CheckUploadDir(cfg)
		require.NotNil(t, err)
		assert.Contains(t, err.ErrorMessage, "image cache directory invalid")
	})
}
