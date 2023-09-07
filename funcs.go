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
)

func NewFuncs(r *http.Request) template.FuncMap {
	return map[string]any{
		//"getPermissionsByUserIdAndSectionAndSectionAll":
		"now": func() time.Time { return time.Now() },
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
			type Post struct {
				*GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
				ShowReply bool
				ShowEdit  bool
				Editing   bool
			}
			var result []*Post
			queries := r.Context().Value(ContextValues("queries")).(*Queries)

			posts, err := queries.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(r.Context(), GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
				Limit:  15,
				Offset: 0,
			})
			if err != nil {
				switch {
				case errors.Is(err, sql.ErrNoRows):
				default:
					return nil, fmt.Errorf("getNewsPostsWithWriterUsernameAndThreadCommentCountDescending: %w", err)
				}
			}

			editingId, _ := strconv.Atoi(r.URL.Query().Get("reply"))

			for _, post := range posts {
				result = append(result, &Post{
					GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow: post,
					ShowReply: true, // TODO
					ShowEdit:  true, // TODO
					Editing:   editingId == int(post.Idsitenews),
				})
			}

			return result, nil
		},
	}
}
