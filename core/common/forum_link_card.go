package common

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/a4code"
	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/a4code/ast"
	"github.com/arran4/goa4web/internal/jobs"
)

// ForumLinkCard holds the current display fields. Signed image paths are
// generated for the response and are excluded from change detection.
type ForumLinkCard struct {
	URL         string
	Title       string
	Description string
	Image       string
	ImageCache  string
	Duration    string
	Pending     bool
}

func forumLinkKey(url string) string {
	sum := sha256.Sum256([]byte(url))
	return hex.EncodeToString(sum[:])
}

// ForumLinkCardForViewer resolves the key against a current visible comment.
// Both socket observations and fragment reads use the same authorization path.
func (cd *CoreData) ForumLinkCardForViewer(ctx context.Context, threadID int32, key string) (ForumLinkCard, error) {
	commentKey, urlKey, ok := strings.Cut(key, "-")
	commentID, err := strconv.ParseInt(commentKey, 10, 32)
	if !ok || err != nil || commentID <= 0 || len(urlKey) != sha256.Size*2 {
		return ForumLinkCard{}, sql.ErrNoRows
	}
	if _, err := hex.DecodeString(urlKey); err != nil {
		return ForumLinkCard{}, sql.ErrNoRows
	}
	comment, err := cd.forumCommentForViewer(0, threadID, int32(commentID), false)
	if err != nil || comment == nil || !comment.Text.Valid {
		return ForumLinkCard{}, sql.ErrNoRows
	}
	root, err := a4code.ParseString(comment.Text.String)
	if err != nil {
		return ForumLinkCard{}, sql.ErrNoRows
	}
	var canonicalURL string
	_ = ast.Walk(root, func(node ast.Node) error {
		link, ok := node.(*ast.Link)
		if ok && link.IsImmediateClose() && ast.IsBlockNode(link) {
			candidate := CanonicalizeExternalURL(link.Href)
			if IsHTTPURL(candidate) && forumLinkKey(candidate) == urlKey {
				canonicalURL = candidate
			}
		}
		return nil
	})
	if canonicalURL == "" {
		return ForumLinkCard{}, sql.ErrNoRows
	}
	return cd.StoredLinkCard(ctx, canonicalURL)
}

// StoredLinkCard reads internal preview state for a canonical URL already
// resolved by the server. Request-supplied keys must use ForumLinkCardForViewer.
func (cd *CoreData) StoredLinkCard(ctx context.Context, canonicalURL string) (ForumLinkCard, error) {
	card := ForumLinkCard{URL: canonicalURL}
	if cd.queries == nil {
		return card, nil
	}
	metadata, err := cd.GetExternalLink(ctx, canonicalURL)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return ForumLinkCard{}, err
	}
	if metadata != nil {
		card.Title = metadata.CardTitle.String
		card.Description = metadata.CardDescription.String
		card.Image = metadata.CardImage.String
		card.ImageCache = metadata.CardImageCache.String
		card.Duration = formatDuration(metadata.CardDuration.String)
	}
	active, err := cd.queries.SystemCountActiveDurableJobsByKey(ctx, sql.NullString{
		String: jobs.DedupeKey(jobs.JobTypeExternalLink, canonicalURL), Valid: true,
	})
	if err != nil {
		return ForumLinkCard{}, err
	}
	card.Pending = active != 0
	return card, nil
}

// RenderForumLinkCard serves the same markup for initial display and live refresh.
func (cd *CoreData) RenderForumLinkCard(w io.Writer, r *http.Request, key, threadPath string, card ForumLinkCard) error {
	title := card.Title
	if title == "" {
		title = card.URL
	}
	image := card.Image
	if card.ImageCache != "" {
		image = cd.MapCardImageURL("img", card.ImageCache)
	}
	if _, ok := a4code2html.SanitizeURL(image); !ok {
		image = ""
	}
	data := struct {
		ForumLinkCard
		Key          string
		FragmentURL  string
		TargetURL    string
		DisplayTitle string
		SafeImage    string
	}{
		ForumLinkCard: card,
		Key:           key,
		FragmentURL:   fmt.Sprintf("%s/link-preview?card=%s", threadPath, url.QueryEscape(key)),
		TargetURL:     cd.SignLinkURL(card.URL),
		DisplayTitle:  title,
		SafeImage:     image,
	}
	return cd.ExecuteSiteTemplate(w, r, "domains/forum/linkCard.gohtml", data)
}
