package imagebbs

import (
	"database/sql"
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

func AdminFilesPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	cd.PageTitle = "Image Files"
	type Entry struct {
		Name     string
		Path     string
		Size     int64
		IsDir    bool
		Username string
		Board    string
		Posted   time.Time
		URL      string
	}
	type Data struct {
		*common.CoreData
		Path    string
		Parent  string
		Entries []Entry
	}

	base := cd.Config.ImageUploadDir
	reqPath := r.URL.Query().Get("path")
	cleaned := filepath.Clean("/" + reqPath)
	abs := filepath.Join(base, cleaned)
	if rel, err := filepath.Rel(base, abs); err != nil || rel == ".." || strings.HasPrefix(rel, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	f, err := os.ReadDir(abs)
	if err != nil {
		log.Printf("readdir: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := Data{
		CoreData: r.Context().Value(consts.KeyCoreData).(*common.CoreData),
		Path:     cleaned,
	}
	if cleaned != "/" {
		data.Parent = filepath.Dir(cleaned)
	}

	ttlStr := r.URL.Query().Get("ttl")
	ttl := 24 * time.Hour
	if ttlStr != "" {
		if d, err := time.ParseDuration(ttlStr); err == nil {
			ttl = d
		}
	}

	for _, e := range f {
		fi, _ := e.Info()
		ent := Entry{
			Name:  e.Name(),
			Path:  filepath.Join(cleaned, e.Name()),
			Size:  fi.Size(),
			IsDir: e.IsDir(),
		}
		if !e.IsDir() {
			dbPath := path.Join("/imagebbs/images", ent.Path)
			q := cd.Queries()
			row, err := q.GetImagePostInfoByPath(r.Context(), db.GetImagePostInfoByPathParams{
				Fullimage: sql.NullString{Valid: true, String: dbPath},
				Thumbnail: sql.NullString{Valid: true, String: dbPath},
			})
			if err == nil && row != nil {
				ent.Username = row.Username.String
				ent.Board = row.Title.String
				if row.Posted.Valid {
					ent.Posted = row.Posted.Time
				}
			}
			if cd.ImageSigner != nil {
				id := filepath.Base(ent.Path)
				ent.URL = cd.ImageSigner.SignedURLTTL(id, ttl)
			}
		}
		data.Entries = append(data.Entries, ent)
	}
	sort.Slice(data.Entries, func(i, j int) bool { return data.Entries[i].Name < data.Entries[j].Name })

	handlers.TemplateHandler(w, r, "adminFilesPage.gohtml", data)
}
