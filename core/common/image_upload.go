package common

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	imagesign "github.com/arran4/goa4web/internal/images"
)

// PreparedImageUpload contains image data validated for normal upload storage.
type PreparedImageUpload struct {
	ID    string
	Ext   string
	Data  []byte
	Image image.Image
}

// PrepareImageUpload decodes an image and derives its deterministic upload identity.
func PrepareImageUpload(data []byte, filename string) (*PreparedImageUpload, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	ext, err := imagesign.CleanExtension(filename)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	id := fmt.Sprintf("%x", hash[:20])
	if !imagesign.ValidID(id) {
		return nil, fmt.Errorf("invalid generated image id")
	}
	return &PreparedImageUpload{ID: id, Ext: ext, Data: data, Image: img}, nil
}

// StoreUploadedImage validates and stores one normal user image upload.
func (cd *CoreData) StoreUploadedImage(data []byte, filename string, uploaderID int32) (string, error) {
	prepared, err := PrepareImageUpload(data, filename)
	if err != nil {
		return "", err
	}
	return cd.StorePreparedImage(prepared, uploaderID)
}

// StorePreparedImage stores a previously validated normal user image upload.
func (cd *CoreData) StorePreparedImage(prepared *PreparedImageUpload, uploaderID int32) (string, error) {
	if prepared == nil {
		return "", fmt.Errorf("prepared image upload is nil")
	}
	return cd.StoreImage(StoreImageParams{
		ID:         prepared.ID,
		Ext:        prepared.Ext,
		Data:       prepared.Data,
		Image:      prepared.Image,
		UploaderID: uploaderID,
	})
}
