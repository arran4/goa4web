package goa4web

import (
	corecommon "github.com/arran4/goa4web/core/common"
	common "github.com/arran4/goa4web/handlers/common"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/runtimeconfig"
)

func imagebbsAdminFilesPage(w http.ResponseWriter, r *http.Request) {
	type Entry struct {
		Name  string
		Path  string
		Size  int64
		IsDir bool
	}
	type Data struct {
		*CoreData
		Path    string
		Parent  string
		Entries []Entry
	}

	base := runtimeconfig.AppRuntimeConfig.ImageUploadDir
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
		CoreData: r.Context().Value(common.KeyCoreData).(*CoreData),
		Path:     cleaned,
	}
	if cleaned != "/" {
		data.Parent = filepath.Dir(cleaned)
	}

	for _, e := range f {
		fi, _ := e.Info()
		data.Entries = append(data.Entries, Entry{
			Name:  e.Name(),
			Path:  filepath.Join(cleaned, e.Name()),
			Size:  fi.Size(),
			IsDir: e.IsDir(),
		})
	}
	sort.Slice(data.Entries, func(i, j int) bool { return data.Entries[i].Name < data.Entries[j].Name })

	if err := templates.RenderTemplate(w, "adminFilesPage.gohtml", data, corecommon.NewFuncs(r)); err != nil {
		log.Printf("template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
