package user

import (
	"github.com/arran4/goa4web/core/consts"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/core/common"

	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
)

type galleryImage struct {
	Thumb  string
	Full   string
	A4Code string
}

func userGalleryPage(w http.ResponseWriter, r *http.Request) {
	session, ok := core.GetSessionOrFail(w, r)
	if !ok {
		return
	}
	uid, _ := session.Values["UID"].(int32)
	queries := r.Context().Value(consts.KeyCoreData).(*common.CoreData).Queries()

	pageStr := r.URL.Query().Get("p")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)
	size := cd.Config.PageSizeDefault
	if pref, _ := cd.Preference(); pref != nil {
		size = int(pref.PageSize)
	}

	offset := (page - 1) * size

	rows, err := queries.ListUploadedImagesByUser(r.Context(), db.ListUploadedImagesByUserParams{
		UsersIdusers: uid,
		Limit:        int32(size + 1),
		Offset:       int32(offset),
	})
	if err != nil {
		log.Printf("list images: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	hasMore := len(rows) > size
	if hasMore {
		rows = rows[:size]
	}

	var imgs []galleryImage
	for _, img := range rows {
		if !img.Path.Valid {
			continue
		}
		fname := path.Base(img.Path.String)
		ext := filepath.Ext(fname)
		id := strings.TrimSuffix(fname, ext)
		thumb := id + "_thumb" + ext
		imgs = append(imgs, galleryImage{
			Thumb:  cd.ImageSigner.SignedCacheURL(thumb),
			Full:   cd.ImageSigner.SignedURL("image:" + fname),
			A4Code: "[img=image:" + fname + "]",
		})
	}

	base := "/usr/notifications/gallery"
	var nextLink, prevLink string
	if hasMore {
		nextLink = base + "?p=" + strconv.Itoa(page+1)
	}
	if page > 1 {
		prevLink = base + "?p=" + strconv.Itoa(page-1)
	}

	data := struct {
		*common.CoreData
		Images   []galleryImage
		NextLink string
		PrevLink string
		PageSize int
	}{
		CoreData: cd,
		Images:   imgs,
		NextLink: nextLink,
		PrevLink: prevLink,
		PageSize: size,
	}

	handlers.TemplateHandler(w, r, "gallery.gohtml", data)
}
