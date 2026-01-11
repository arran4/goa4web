package news

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/share"
	"github.com/arran4/goa4web/internal/sharesign"
	"github.com/gorilla/mux"
)

// SharedPreviewPage renders an OpenGraph preview for a news post.
func SharedPreviewPage(w http.ResponseWriter, r *http.Request) {
	cd := r.Context().Value(consts.KeyCoreData).(*common.CoreData)

	signer := cd.ShareSigner
	if signer == nil {
		signer = sharesign.NewSigner(cd.Config, cd.Config.ShareSignSecret)
	}

	// Verify signature
	if share.VerifyAndGetPath(r, signer) == "" {
		handlers.RenderErrorPage(w, r, handlers.WrapForbidden(fmt.Errorf("invalid signature")))
		return
	}

	vars := mux.Vars(r)
	newsID, _ := strconv.Atoi(vars["news"])

	foundPost, err := cd.SystemGetNewsPost(int32(newsID))
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}
		return
	}

	titleLine := strings.Split(foundPost.News.String, "\n")[0]
	ogTitle := a4code.Snip(titleLine, 100)
	ogDescription := a4code.Snip(foundPost.News.String, 128)

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}

	tsStr := r.URL.Query().Get("ts")
	ts, _ := strconv.ParseInt(tsStr, 10, 64)
	exp := time.Now().Add(24 * time.Hour)
	if ts > 0 {
		exp = time.Unix(ts, 0)
	}

	ogData := share.OpenGraphData{
		Title:       ogTitle,
		Description: ogDescription,
		ImageURL:    template.URL(share.MakeImageURL(cd.AbsoluteURL(""), ogTitle, signer, exp)),
		ContentURL:  template.URL(cd.AbsoluteURL(r.URL.RequestURI())),
		ImageWidth:  cd.Config.OGImageWidth,
		ImageHeight: cd.Config.OGImageHeight,
		TwitterSite: cd.Config.TwitterSite,
	}

	if err := share.RenderOpenGraph(w, r, ogData); err != nil {
		handlers.RenderErrorPage(w, r, err)
	}
}
