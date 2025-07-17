package imagebbs

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	common "github.com/arran4/goa4web/core/common"

	handlers "github.com/arran4/goa4web/handlers"
	db "github.com/arran4/goa4web/internal/db"
	searchutil "github.com/arran4/goa4web/internal/searchworker"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/core/templates"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/upload"
)

// UploadImageTask handles uploading an image to a board.
type UploadImageTask struct{ tasks.TaskString }

var uploadImageTask = &UploadImageTask{TaskString: TaskUploadImage}

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
		CoreData:    r.Context().Value(common.KeyCoreData).(*common.CoreData),
		IsSubBoard:  bid != 0,
		BoardNumber: bid,
	}

	if !data.CoreData.HasGrant("imagebbs", "board", "view", int32(bid)) {
		_ = templates.GetCompiledSiteTemplates(r.Context().Value(common.KeyCoreData).(*common.CoreData).Funcs(r)).ExecuteTemplate(w, "noAccessPage.gohtml", data.CoreData)
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

func (UploadImageTask) Action(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("text")

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])

	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(common.KeyQueries).(*db.Queries)

	board, err := queries.GetImageBoardById(r.Context(), int32(bid))
	if err != nil {
		log.Printf("GetImageBoardById Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(config.AppRuntimeConfig.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(config.AppRuntimeConfig.ImageMaxBytes)); err != nil {
		http.Error(w, "bad upload", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	h := sha1.New()
	size, err := io.Copy(io.MultiWriter(&buf, h), file)
	if err != nil {
		log.Printf("copy upload error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	shaHex := fmt.Sprintf("%x", h.Sum(nil))
	ext := strings.ToLower(filepath.Ext(header.Filename))
	sub1, sub2 := shaHex[:2], shaHex[2:4]
	data := buf.Bytes()
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		log.Printf("decode image error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fname := shaHex + ext
	if p := upload.ProviderFromConfig(config.AppRuntimeConfig); p != nil {
		if err := p.Write(r.Context(), path.Join(sub1, sub2, fname), data); err != nil {
			log.Printf("upload write: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		thumb := imaging.Thumbnail(img, 200, 200, imaging.Lanczos)
		var buf bytes.Buffer
		if err := imaging.Encode(&buf, thumb, imaging.PNG); err != nil {
			log.Printf("encode thumb: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		thumbName := shaHex + "_thumb" + ext
		if err := p.Write(r.Context(), path.Join(sub1, sub2, thumbName), buf.Bytes()); err != nil {
			log.Printf("thumb write: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
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
		log.Printf("printTopicRestrictions Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	wordIds, done := searchutil.SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if searchutil.InsertWordsToImageSearch(w, r, wordIds, queries, pid) {
		return
	}

	handlers.TaskDoneAutoRefreshPage(w, r)
}
