package news

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
	"unicode"

	hcommon "github.com/arran4/goa4web/handlers/common"
	db "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/email"
)

func processCommentFullQuote(username, text string) string {
	var out bytes.Buffer
	var quote bytes.Buffer
	var it, bc, nlc int

	for it < len(text) {
		switch text[it] {
		case ']':
			bc--
		case '[':
			bc++
		case '\\':
			if it+1 < len(text) {
				if text[it+1] == '[' || text[it+1] == ']' {
					out.WriteByte(text[it+1])
					it++
				}
			}
		case '\n':
			if bc == 0 && nlc == 1 {
				quote.WriteString(processCommentQuote(username, out.String()))
				out.Reset()
			}
			nlc++
			it++
			continue
		case '\r':
			it++
			continue
		case ' ':
			fallthrough
		default:
			if nlc != 0 {
				if out.Len() > 0 {
					out.WriteByte('\n')
				}
				nlc = 0
			}
			out.WriteByte(text[it])
		}
		it++
	}
	quote.WriteString(processCommentQuote(username, out.String()))
	return quote.String()
}

func processCommentQuote(username string, text string) string {
	return fmt.Sprintf("[quoteof \"%s\" %s]\n", username, text)
}

func notifyChange(ctx context.Context, provider email.Provider, emailAddr string, page string) error {
	if emailAddr == "" {
		return fmt.Errorf("no email specified")
	}
	from := email.SourceEmail

	type EmailContent struct {
		To      string
		From    string
		Subject string
		URL     string
	}

	content := EmailContent{
		To:      emailAddr,
		From:    from,
		Subject: "Website Update Notification",
		URL:     page,
	}

	var notification bytes.Buffer
	tmpl, err := template.New("email").Parse(getUpdateEmailText(ctx))
	if err != nil {
		return fmt.Errorf("parse email template: %w", err)
	}
	if err := tmpl.Execute(&notification, content); err != nil {
		return fmt.Errorf("execute email template: %w", err)
	}

	if q, ok := ctx.Value(hcommon.KeyQueries).(*db.Queries); ok {
		if err := q.InsertPendingEmail(ctx, db.InsertPendingEmailParams{ToEmail: emailAddr, Subject: content.Subject, Body: notification.String()}); err != nil {
			return err
		}
	} else if provider != nil {
		if err := provider.Send(ctx, emailAddr, content.Subject, notification.String()); err != nil {
			return fmt.Errorf("send email: %w", err)
		}
	}
	return nil
}

func breakupTextToWords(input string) []string {
	var tokens []string
	startIndex := -1

	for i, char := range input {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || strings.ContainsRune("'-", char) {
			if startIndex == -1 {
				startIndex = i
			}
		} else if startIndex != -1 {
			tokens = append(tokens, input[startIndex:i])
			startIndex = -1
		}
	}
	if startIndex != -1 {
		tokens = append(tokens, input[startIndex:])
	}
	return tokens
}

func SearchWordIdsFromTextLocal(w http.ResponseWriter, r *http.Request, text string, queries *db.Queries) ([]int64, bool) {
	words := map[string]int32{}
	for _, word := range breakupTextToWords(text) {
		words[strings.ToLower(word)] = 0
	}
	wordIds := make([]int64, 0, len(words))
	for word := range words {
		id, err := queries.CreateSearchWord(r.Context(), strings.ToLower(word))
		if err != nil {
			log.Printf("Error: createSearchWord: %s", err)
			http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
			return nil, true
		}
		wordIds = append(wordIds, id)
	}
	return wordIds, false
}

func InsertWordsToForumSearchLocal(w http.ResponseWriter, r *http.Request, wordIds []int64, queries *db.Queries, cid int64) bool {
	for _, wid := range wordIds {
		if err := queries.AddToForumCommentSearch(r.Context(), db.AddToForumCommentSearchParams{
			CommentsIdcomments:             int32(cid),
			SearchwordlistIdsearchwordlist: int32(wid),
		}); err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("Error: addToForumCommentSearch: %s", err)
				http.Redirect(w, r, "?error="+err.Error(), http.StatusTemporaryRedirect)
				return true
			}
		}
	}
	return false
}

func PostUpdateLocal(ctx context.Context, q *db.Queries, threadID, topicID int32) error {
	if err := q.RecalculateForumThreadByIdMetaData(ctx, threadID); err != nil {
		return fmt.Errorf("recalc thread metadata: %w", err)
	}
	if err := q.RebuildForumTopicByIdMetaColumns(ctx, topicID); err != nil {
		return fmt.Errorf("rebuild topic metadata: %w", err)
	}
	return nil
}

func getUpdateEmailText(ctx context.Context) string {
	// simplified; in original package it reads from templates. Here we just return a basic text.
	return "A page you follow has been updated: {{.URL}}"
}
