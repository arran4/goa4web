package goa4web

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func imagebbsBoardPage(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		*CoreData
		Boards      []*Imageboard
		IsSubBoard  bool
		BoardNumber int
		Posts       []*GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCountRow
	}

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])

	data := Data{
		CoreData:    r.Context().Value(ContextValues("coreData")).(*CoreData),
		IsSubBoard:  bid != 0,
		BoardNumber: bid,
	}

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	subBoardRows, err := queries.GetAllBoardsByParentBoardId(r.Context(), int32(bid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Boards = subBoardRows

	posts, err := queries.GetAllImagePostsByBoardIdWithAuthorUsernameAndThreadCommentCount(r.Context(), int32(bid))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
		default:
			log.Printf("getAllBoardsByParentBoardId Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	data.Posts = posts

	CustomImageBBSIndex(data.CoreData, r)

	if err := renderTemplate(w, r, "boardPage.gohtml", data); err != nil {
		log.Printf("Template Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func imagebbsBoardPostImageActionPage(w http.ResponseWriter, r *http.Request) {
	text := r.PostFormValue("text")

	vars := mux.Vars(r)
	bid, _ := strconv.Atoi(vars["boardno"])

	session, ok := GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)

	queries := r.Context().Value(ContextValues("queries")).(*Queries)

	board, err := queries.GetImageBoardById(r.Context(), int32(bid))
	if err != nil {
		log.Printf("GetImageBoardById Error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if strings.HasPrefix(appRuntimeConfig.ImageUploadDir, "s3://") {
		// TODO: upload to S3 instead of the local filesystem
		http.Error(w, "s3 uploads not implemented", http.StatusNotImplemented)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, int64(appRuntimeConfig.ImageMaxBytes))
	if err := r.ParseMultipartForm(int64(appRuntimeConfig.ImageMaxBytes)); err != nil {
		http.Error(w, "bad upload", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "image required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	info, err := os.Stat(appRuntimeConfig.ImageUploadDir)
	if err != nil || !info.IsDir() {
		log.Printf("invalid upload dir: %v", err)
		http.Error(w, "Uploads disabled", http.StatusInternalServerError)
		return
	}

	tmp, err := os.CreateTemp(appRuntimeConfig.ImageUploadDir, "upload-")
	if err != nil {
		log.Printf("tempfile error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())

	h := sha1.New()
	size, err := io.Copy(io.MultiWriter(tmp, h), file)
	if err != nil {
		tmp.Close()
		log.Printf("copy upload error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	tmp.Close()

	shaHex := fmt.Sprintf("%x", h.Sum(nil))
	ext := strings.ToLower(filepath.Ext(header.Filename))
	sub1, sub2 := shaHex[:2], shaHex[2:4]
	destDir := filepath.Join(appRuntimeConfig.ImageUploadDir, sub1, sub2)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Printf("mkdir error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	fname := shaHex + ext
	fullPath := filepath.Join(destDir, fname)
	if err := os.Rename(tmp.Name(), fullPath); err != nil {
		log.Printf("rename error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	img, err := imaging.Open(fullPath)
	if err != nil {
		log.Printf("decode image error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	thumb := imaging.Thumbnail(img, 200, 200, imaging.Lanczos)
	thumbName := shaHex + "_thumb" + ext
	thumbPath := filepath.Join(destDir, thumbName)
	if err := imaging.Save(thumb, thumbPath); err != nil {
		log.Printf("save thumbnail error: %s", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	relBase := path.Join("/imagebbs/images", sub1, sub2)
	relFull := path.Join(relBase, fname)
	relThumb := path.Join(relBase, thumbName)

	approved := !board.ApprovalRequired

	pid, err := queries.CreateImagePost(r.Context(), CreateImagePostParams{
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

	wordIds, done := SearchWordIdsFromText(w, r, text, queries)
	if done {
		return
	}

	if InsertWordsToImageSearch(w, r, wordIds, queries, pid) {
		return
	}

	taskDoneAutoRefreshPage(w, r)
}
