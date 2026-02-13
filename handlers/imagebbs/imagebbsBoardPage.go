package imagebbs

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/workers/searchworker"
	"golang.org/x/image/draw"
)

// UploadImageTask handles uploading an image to a board.
type UploadImageTask struct{ tasks.TaskString }

var uploadImageTask = &UploadImageTask{TaskString: TaskUploadImage}

// UploadImageTask participates in generic task handling
var _ tasks.Task = (*UploadImageTask)(nil)
var _ tasks.AuditableTask = (*UploadImageTask)(nil)

func (UploadImageTask) IndexType() string { return searchworker.TypeImage }

func (UploadImageTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = UploadImageTask{}

func ImagebbsBoardPage(w http.ResponseWriter, r *http.Request) {
	t := NewImagebbsBoardTask().(*imagebbsBoardTask)
	t.Get(w, r)
}

func (UploadImageTask) Action(w http.ResponseWriter, r *http.Request) any {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	text := r.PostFormValue("text")

	vars := mux.Vars(r)
	bidStr := vars["board"]
	if bidStr == "" {
		bidStr = vars["boardno"]
	}
	bid, _ := strconv.Atoi(bidStr)

	session := cd.GetSession()
	uid, _ := session.Values["UID"].(int32)

	queries := cd.Queries()

	board, err := queries.GetImageBoardById(r.Context(), int32(bid))
	if err != nil {
		return fmt.Errorf("get image board fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	if !cd.HasGrant("imagebbs", "board", "post", int32(bid)) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}
	if !cd.HasGrant("images", "upload", "post", 0) {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlers.RenderErrorPage(w, r, handlers.ErrForbidden)
		})
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(cd.Config.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(cd.Config.ImageMaxBytes)); err != nil {
		return fmt.Errorf("parse form fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		return fmt.Errorf("image required %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	defer file.Close()

	var buf bytes.Buffer
	h := sha1.New()
	size, err := io.Copy(io.MultiWriter(&buf, h), file)
	if err != nil {
		return fmt.Errorf("copy upload error %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	shaHex := fmt.Sprintf("%x", h.Sum(nil))
	ext, err := imagesign.CleanExtension(header.Filename)
	if err != nil {
		return fmt.Errorf("invalid extension %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	sub1, sub2 := shaHex[:2], shaHex[2:4]
	data := buf.Bytes()
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decode image error %w", handlers.ErrRedirectOnSamePageHandler(err))
	}
	fname := shaHex + ext
	if p := upload.ProviderFromConfig(cd.Config); p != nil {
		if err := p.Write(r.Context(), path.Join(sub1, sub2, fname), data); err != nil {
			return fmt.Errorf("upload write fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		src := img.Bounds()
		var crop image.Rectangle
		if src.Dx() > src.Dy() {
			side := src.Dy()
			x0 := src.Min.X + (src.Dx()-side)/2
			crop = image.Rect(x0, src.Min.Y, x0+side, src.Min.Y+side)
		} else {
			side := src.Dx()
			y0 := src.Min.Y + (src.Dy()-side)/2
			crop = image.Rect(src.Min.X, y0, src.Min.X+side, y0+side)
		}
		thumb := image.NewRGBA(image.Rect(0, 0, 200, 200))
		draw.CatmullRom.Scale(thumb, thumb.Bounds(), img, crop, draw.Over, nil)
		var buf bytes.Buffer
		enc, err := imagesign.EncoderByExtension(ext)
		if err != nil {
			return fmt.Errorf("encoder fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		if err := enc(&buf, thumb); err != nil {
			return fmt.Errorf("encode thumb fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
		thumbName := shaHex + "_thumb" + ext
		if err := p.Write(r.Context(), path.Join(sub1, sub2, thumbName), buf.Bytes()); err != nil {
			return fmt.Errorf("thumb write fail %w", handlers.ErrRedirectOnSamePageHandler(err))
		}
	}

	relBase := path.Join("/imagebbs/images", sub1, sub2)
	relFull := path.Join(relBase, fname)
	thumbName := shaHex + "_thumb" + ext
	relThumb := path.Join(relBase, thumbName)

	if _, err := queries.CreateUploadedImageForUploader(r.Context(), db.CreateUploadedImageForUploaderParams{
		UploaderID: uid,
		Path:       sql.NullString{String: relFull, Valid: true},
		Width:      sql.NullInt32{Int32: int32(img.Bounds().Dx()), Valid: true},
		Height:     sql.NullInt32{Int32: int32(img.Bounds().Dy()), Valid: true},
		FileSize:   int32(size),
	}); err != nil {
		return fmt.Errorf("record uploaded image %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	approved := !board.ApprovalRequired

	pid, err := queries.CreateImagePostForPoster(r.Context(), db.CreateImagePostForPosterParams{
		ImageboardID: sql.NullInt32{Int32: int32(bid), Valid: true},
		Thumbnail:    sql.NullString{Valid: true, String: relThumb},
		Fullimage:    sql.NullString{Valid: true, String: relFull},
		PosterID:     uid,
		Description:  sql.NullString{Valid: true, String: text},
		Posted:       sql.NullTime{Time: time.Now().UTC(), Valid: true},
		Timezone:     sql.NullString{String: cd.Location().String(), Valid: true},
		Approved:     approved,
		FileSize:     int32(size),
		GrantBoardID: sql.NullInt32{Int32: int32(bid), Valid: true},
		GranteeID:    sql.NullInt32{Int32: uid, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("create image post fail %w", handlers.ErrRedirectOnSamePageHandler(err))
	}

	if cd, ok := r.Context().Value(consts.KeyCoreData).(*common.CoreData); ok {
		if evt := cd.Event(); evt != nil {
			if evt.Data == nil {
				evt.Data = map[string]any{}
			}
			evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeImage, ID: int32(pid), Text: text}
			evt.Data["ImagePostID"] = int32(pid)
			evt.Data["BoardID"] = int32(bid)
		}
	}

	return nil
}

func (UploadImageTask) AuditRecord(data map[string]any) string {
	if id, ok := data["ImagePostID"].(int32); ok {
		return fmt.Sprintf("uploaded image %d", id)
	}
	return "uploaded image"
}
