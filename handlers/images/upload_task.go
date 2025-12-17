package images

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
	intimages "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
)

// UploadImageTask processes authenticated image uploads.
type UploadImageTask struct{ tasks.TaskString }

var uploadImageTask = &UploadImageTask{TaskString: TaskUploadImage}

// ensure UploadImageTask conforms to tasks.Task
var _ tasks.Task = (*UploadImageTask)(nil)

func (UploadImageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cfg := cd.Config

	r.Body = http.MaxBytesReader(w, r.Body, int64(cfg.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(cfg.ImageMaxBytes)); err != nil {
		return fmt.Errorf("bad upload %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		return fmt.Errorf("image required %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read file %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decode image %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	id := r.FormValue("id")
	if id == "" {
		id = fmt.Sprintf("%x", sha1.Sum(data))
	}
	ext, err := intimages.CleanExtension(header.Filename)
	if err != nil {
		return fmt.Errorf("invalid extension %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if !intimages.ValidID(id) {
		return fmt.Errorf("invalid id %w", handlers.ErrRedirectOnSamePageHandler(fmt.Errorf("bad id")))
	}
	uid := cd.UserID
	fname, err := cd.StoreImage(common.StoreImageParams{ID: id, Ext: ext, Data: data, Image: img, UploaderID: uid})
	if err != nil {
		return fmt.Errorf("store image %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if cd.ImageSigner != nil {
		signed := cd.ImageSigner.SignedRef("image:" + fname)
		return handlers.TextByteWriter([]byte(signed))
	}
	return handlers.TextByteWriter([]byte("image:" + fname))
}
