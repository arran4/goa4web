package images

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	intimages "github.com/arran4/goa4web/internal/images"
)

// APIListGallery handles GET /api/images
func APIListGallery(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := cd.PageSize()
	offset := (page - 1) * pageSize

	// List images uploaded by the user
	images, err := cd.Queries().ListUploadedImagesByUserForLister(r.Context(), db.ListUploadedImagesByUserForListerParams{
		ListerID:      cd.UserID,
		UserID:        cd.UserID,
		ListerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:         int32(pageSize + 1),
		Offset:        int32(offset),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hasMore := len(images) > pageSize
	if hasMore {
		images = images[:pageSize]
	}

	// Prepare data
	type apiImage struct {
		ID   int32  `json:"id"`
		Path string `json:"path"`
		URL  string `json:"url"`
	}

	var apiImages []apiImage
	for _, img := range images {
		if img.Path.Valid {
			// Construct a simple relative URL for API users or rely on ImageURLMapper
			apiImages = append(apiImages, apiImage{
				ID:   img.Iduploadedimage,
				Path: img.Path.String,
				URL:  cd.ImageURLMapper("img", "image:"+img.Path.String),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"images":   apiImages,
		"has_more": hasMore,
		"page":     page,
	})
}

// APIUploadImage handles POST /api/images/upload
func APIUploadImage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config

	r.Body = http.MaxBytesReader(w, r.Body, int64(cfg.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(cfg.ImageMaxBytes)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read image data", http.StatusInternalServerError)
		return
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		http.Error(w, "Failed to decode image data", http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256(data)
	id := fmt.Sprintf("%x", hash[:20])
	ext, err := intimages.CleanExtension(header.Filename)
	if err != nil {
		http.Error(w, "Invalid file extension", http.StatusBadRequest)
		return
	}
	if !intimages.ValidID(id) {
		http.Error(w, "Invalid ID generated", http.StatusInternalServerError)
		return
	}

	uid := cd.UserID
	fname, err := cd.StoreImage(common.StoreImageParams{ID: id, Ext: ext, Data: data, Image: img, UploaderID: uid})
	if err != nil {
		http.Error(w, "Failed to store image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  "success",
		"message": "Image uploaded successfully",
		"path":    fname,
		"url":     cd.ImageURLMapper("img", "image:"+fname),
	})
}
