package images

import (
	"context"
	"crypto/sha256"
	"fmt"
	"path"
	"strings"

	"github.com/arran4/goa4web/internal/upload"
)

// imageBoardThumbnailSize retains the square size used by existing image board URLs.
const imageBoardThumbnailSize = 200

// WriteImageBoardThumbnail regenerates a thumbnail from its immutable upload.
func WriteImageBoardThumbnail(ctx context.Context, provider upload.Provider, filename string) error {
	ext := path.Ext(filename)
	digest := strings.TrimSuffix(filename, ext)
	if !ValidCacheDigest(digest) || !AllowedExtension(ext) || ext != strings.ToLower(ext) {
		return fmt.Errorf("invalid image board upload name")
	}
	if provider == nil {
		return fmt.Errorf("image board upload provider missing")
	}

	dir := path.Join(digest[:2], digest[2:4])
	data, err := provider.Read(ctx, path.Join(dir, filename))
	if err != nil {
		return fmt.Errorf("read image board upload: %w", err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
		return fmt.Errorf("image board upload digest mismatch")
	}
	img, err := DecodeForResize(data)
	if err != nil {
		return fmt.Errorf("decode image board upload: %w", err)
	}
	thumb, err := GenerateThumbnail(img, ext, "draw", imageBoardThumbnailSize)
	if err != nil {
		return fmt.Errorf("generate image board thumbnail: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	// Every writer for this key derives the same bytes from the same upload.
	if err := provider.Write(ctx, path.Join(dir, digest+"_thumb"+ext), thumb); err != nil {
		return fmt.Errorf("write image board thumbnail: %w", err)
	}
	return nil
}
