package images

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers"
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
	ext := strings.ToLower(filepath.Ext(header.Filename))
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
