package common

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arran4/goa4web/a4code2html"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/runtimeconfig"
	"github.com/gorilla/csrf"
)

// Version identifies the running build for template display.
var Version = "dev"

// NewFuncs returns template helpers used by multiple handlers.
func NewFuncs(r *http.Request) template.FuncMap {
	var LatestNews any
	return map[string]any{
		"now":       func() time.Time { return time.Now() },
		"csrfField": func() template.HTML { return csrf.TemplateField(r) },
		"version":   func() string { return Version },
		"a4code2html": func(s string) template.HTML {
			c := a4code2html.NewA4Code2HTML()
			c.CodeType = a4code2html.CTHTML
			c.SetInput(s)
			c.Process()
			return template.HTML(c.Output())
		},
		"a4code2string": func(s string) string {
			c := a4code2html.NewA4Code2HTML()
			c.CodeType = a4code2html.CTWordsOnly
			c.SetInput(s)
			c.Process()
			return c.Output()
		},
		"firstline": func(s string) string {
			return strings.Split(s, "\n")[0]
		},
		"left": func(i int, s string) string {
			l := len(s)
			if l > i {
				l = i
			}
			return s[:l]
		},
		"LatestNews": func() (any, error) {
			if LatestNews != nil {
				return LatestNews, nil
			}
			type Post struct {
				*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
				ShowReply    bool
				ShowEdit     bool
				Editing      bool
				Announcement *db.SiteAnnouncement
				IsAdmin      bool
			}
			var result []*Post
			queries := r.Context().Value(ContextKey("queries")).(*db.Queries)

			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

			posts, err := queries.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(r.Context(), db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
				Limit:  15,
				Offset: int32(offset),
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					return nil, fmt.Errorf("getNewsPostsWithWriterUsernameAndThreadCommentCountDescending: %w", err)
				}
			}

			editingId, _ := strconv.Atoi(r.URL.Query().Get("reply"))

			cd := r.Context().Value(ContextKey("coreData")).(*CoreData)
			for _, post := range posts {
				ann, err := queries.GetLatestAnnouncementByNewsID(r.Context(), post.Idsitenews)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return nil, fmt.Errorf("getLatestAnnouncementByNewsID: %w", err)
				}
				result = append(result, &Post{
					GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow: post,
					ShowReply:    cd.UserID != 0,
					ShowEdit:     cd.HasRole("writer"),
					Editing:      editingId == int(post.Idsitenews),
					Announcement: ann,
					IsAdmin:      cd.HasRole("administrator"),
				})
			}
			LatestNews = result
			return result, nil
		},
	}
}

// GetPageSize returns the preferred page size within configured bounds.
func GetPageSize(r *http.Request) int {
	size := runtimeconfig.AppRuntimeConfig.PageSizeDefault
	if pref, _ := r.Context().Value(ContextKey("preference")).(*db.Preference); pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < runtimeconfig.AppRuntimeConfig.PageSizeMin {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMin
	}
	if size > runtimeconfig.AppRuntimeConfig.PageSizeMax {
		size = runtimeconfig.AppRuntimeConfig.PageSizeMax
	}
	return size
}
