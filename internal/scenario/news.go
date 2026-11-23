package scenario

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// NewsPostCreateData represents a published News item authored by a scenario user.
type NewsPostCreateData struct {
	Ref, Actor string
	LanguageID int32
	Body       string
	At         time.Time
}

func (d *NewsPostCreateData) Op() string { return "news.post.create" }

// NewsPostCreateOp parses published News creation events.
type NewsPostCreateOp struct{}

func (*NewsPostCreateOp) OpName() string { return "news.post.create" }
func (*NewsPostCreateOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "Language", "At"}
}
func (*NewsPostCreateOp) RequiredHeaders() []string { return []string{"Ref", "Actor", "At"} }
func (*NewsPostCreateOp) DeclaredRef(e *Event) (RefType, string, bool) {
	r := strings.TrimSpace(e.Headers.Get("Ref"))
	return RefTypeNews, r, r != ""
}
func (*NewsPostCreateOp) ReferencedSymbols(e *Event) []SymbolRef {
	return []SymbolRef{{Type: RefTypeUser, Symbol: strings.TrimSpace(e.Headers.Get("Actor")), Field: "Actor"}}
}
func (*NewsPostCreateOp) AssetPaths(*Event) []string { return nil }

func (*NewsPostCreateOp) Parse(e *Event) (OperationData, error) {
	languageID, err := parseNewsLanguage(e)
	if err != nil {
		return nil, err
	}
	body := strings.TrimRight(e.Body, "\r\n")
	if strings.TrimSpace(body) == "" {
		return nil, fmt.Errorf("news.post.create: body is required")
	}
	return &NewsPostCreateData{
		Ref:        strings.TrimSpace(e.Headers.Get("Ref")),
		Actor:      strings.TrimSpace(e.Headers.Get("Actor")),
		LanguageID: languageID,
		Body:       body,
		At:         e.At,
	}, nil
}

// NewsReplyData represents a comment using the News discussion bridge.
type NewsReplyData struct {
	Ref, Actor, News string
	LanguageID       int32
	Body             string
	At               time.Time
}

func (d *NewsReplyData) Op() string { return "news.reply" }

// NewsReplyOp parses comments created through the News discussion bridge.
type NewsReplyOp struct{}

func (*NewsReplyOp) OpName() string { return "news.reply" }
func (*NewsReplyOp) AllowedHeaders() []string {
	return []string{"Op", "Ref", "Actor", "News", "Language", "At"}
}
func (*NewsReplyOp) RequiredHeaders() []string { return []string{"Actor", "News", "At"} }
func (*NewsReplyOp) DeclaredRef(e *Event) (RefType, string, bool) {
	r := strings.TrimSpace(e.Headers.Get("Ref"))
	return RefTypePost, r, r != ""
}
func (*NewsReplyOp) ReferencedSymbols(e *Event) []SymbolRef {
	return []SymbolRef{{Type: RefTypeUser, Symbol: strings.TrimSpace(e.Headers.Get("Actor")), Field: "Actor"}, {Type: RefTypeNews, Symbol: strings.TrimSpace(e.Headers.Get("News")), Field: "News"}}
}
func (*NewsReplyOp) AssetPaths(*Event) []string { return nil }

func (*NewsReplyOp) Parse(e *Event) (OperationData, error) {
	languageID, err := parseNewsLanguage(e)
	if err != nil {
		return nil, err
	}
	return &NewsReplyData{
		Ref:        strings.TrimSpace(e.Headers.Get("Ref")),
		Actor:      strings.TrimSpace(e.Headers.Get("Actor")),
		News:       strings.TrimSpace(e.Headers.Get("News")),
		LanguageID: languageID,
		Body:       strings.TrimRight(e.Body, "\r\n"),
		At:         e.At,
	}, nil
}

func parseNewsLanguage(e *Event) (int32, error) {
	raw := strings.TrimSpace(e.Headers.Get("Language"))
	if raw == "" {
		return 1, nil
	}
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid Language %q", raw)
	}
	return int32(id), nil
}
