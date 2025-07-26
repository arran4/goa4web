package imagebbs

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"github.com/arran4/goa4web/core/consts"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/workers/searchworker"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/gorilla/mux"
	"golang.org/x/image/draw"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/upload"
)

// UploadImageTask handles uploading an image to a board.
type UploadImageTask struct{ tasks.TaskString }

var uploadImageTask = &UploadImageTask{TaskString: TaskUploadImage}

// UploadImageTask participates in generic task handling
var _ tasks.Task = (*UploadImageTask)(nil)

func (UploadImageTask) IndexType() string { return searchworker.TypeImage }

func (UploadImageTask) IndexData(data map[string]any) []searchworker.IndexEventData {
	if v, ok := data[searchworker.EventKey].(searchworker.IndexEventData); ok {
		return []searchworker.IndexEventData{v}
	}
	return nil
}

var _ searchworker.IndexedTask = UploadImageTask{}

func BoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*common.CoreData
		Boards      []*db.Imageboard
		IsSubBoard  bool
		BoardNumber int
		Posts       []*db.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountForUserRow
	}

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])

	data := Data{
		CoreData:    r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		IsSubBoard:  bid != 0,
		BoardNumber: bid,
	}

	if !data.CoreData.HasGrant("imagebbs", "board", "view", int32(bid)) {
		_ = templates.GetCompiledSiteTemplates(r.Context().Value(consts.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
		return
	}

	boards, err := data.CoreData.SubImageBoards(int32(bid))
	if err != nil {
		log.Printf("imageboards: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Boards = boards

	posts, err := data.CoreData.ImageBoardPosts(int32(bid))
	if err != nil {
		log.Printf("imageboard posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data.Posts = posts

	handlers.TemplateHandler(w, r, "boardPage.gohtml", data)
}

func (UploadImageTask) Action(w http.ResponseWriter, r *http.Request) any {
	text := r.PostFormValue("text")

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return handlers.SessionFetchFail{}
	}
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	board, err := queries.GetImageBoardById(r.Context(), int32(bid))
	if err != nil {
		return fmt.Errorf("get image board fail %w", handlers.ErrRedirectOnSamePageHandler(err))
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
	ext := strings.ToLower(filepath.Ext(header.Filename))
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

	approved := !board.ApprovalRequired

	pid, err := queries.CreateImagePost(r.Context(), db.CreateImagePostParams{
		ImageboardIdimageboard: int32(bid),
		Thumbnail:              sql.NullString{Valid: true, String: relThumb},
		Fullimage:              sql.NullString{Valid: true, String: relFull},
		UsersIdusers:           uid,
		Description:            sql.NullString{Valid: true, String: text},
		Approved:               approved,
		FileSize:               int32(size),
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
		}
	}

	return nil
}
