package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/csrf"
)

func NewFuncs(r *http.Request) template.FuncMap {
	var LatestNews any
	return map[string]any{
		//"getPermissionsByUserIdAndSectionAndSectionAll":
		"now":       func() time.Time { return time.Now() },
		"csrfField": func() template.HTML { return csrf.TemplateField(r) },
		"version":   func() string { return version },
		"a4code2html": func(s string) template.HTML {
			c := NewA4Code2HTML()
			c.codeType = ct_html
			c.input = s
			c.Process()
			return template.HTML(c.output.String())
		},
		"a4code2string": func(s string) string {
			c := NewA4Code2HTML()
			c.codeType = ct_wordsonly
			c.input = s
			c.Process()
			return c.output.String()
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
				*GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
				ShowReply    bool
				ShowEdit     bool
				Editing      bool
				Announcement *SiteAnnouncement
				IsAdmin      bool
			}
			var result []*Post
			queries := r.Context().Value(ContextValues("queries")).(*Queries)

			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

			posts, err := queries.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(r.Context(), GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
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

			cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
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
