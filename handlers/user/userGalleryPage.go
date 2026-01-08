package user

import (
	"database/sql"
	"fmt"
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
	cd.PageTitle = "Gallery"
	size := cd.Config.PageSizeDefault
	if pref, _ := cd.Preference(); pref != nil {
		size = int(pref.PageSize)
	}

	offset := (page - 1) * size

	rows, err := queries.ListUploadedImagesByUserForLister(r.Context(), db.ListUploadedImagesByUserForListerParams{
		ListerID:      uid,
		UserID:        uid,
		ListerMatchID: sql.NullInt32{Int32: uid, Valid: uid != 0},
		Limit:         int32(size + 1),
		Offset:        int32(offset),
	})
	if err != nil {
		log.Printf("list images: %v", err)
		handlers.RenderErrorPage(w, r, fmt.Errorf("Internal Server Error"))
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
		imgPath := img.Path.String
		if strings.HasPrefix(imgPath, "/uploads/") {
			fname := path.Base(imgPath)
			ext := filepath.Ext(fname)
			id := strings.TrimSuffix(fname, ext)
			thumb := id + "_thumb" + ext
			full := imgPath
			thumbURL := thumb
			if cd.ImageSigner != nil {
				full = cd.ImageSigner.SignedURL("image:" + fname)
				thumbURL = cd.ImageSigner.SignedCacheURL(thumb)
			}
			imgs = append(imgs, galleryImage{
				Thumb:  thumbURL,
				Full:   full,
				A4Code: "[img=image:" + fname + "]",
			})
			continue
		}
		ext := filepath.Ext(imgPath)
		base := strings.TrimSuffix(imgPath, ext)
		thumb := base + "_thumb" + ext
		imgs = append(imgs, galleryImage{
			Thumb:  thumb,
			Full:   imgPath,
			A4Code: "[img " + imgPath + "]",
		})
	}

	base := "/usr/notifications/gallery"
	if hasMore {
		cd.NextLink = base + "?p=" + strconv.Itoa(page+1)
	}
	if page > 1 {
		cd.PrevLink = base + "?p=" + strconv.Itoa(page-1)
	}

	data := struct {
		Images   []galleryImage
		PageSize int
	}{
		Images:   imgs,
		PageSize: size,
	}

	handlers.TemplateHandler(w, r, "user/gallery.gohtml", data)
}
