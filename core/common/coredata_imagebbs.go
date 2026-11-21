package common

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/feeds"

	"github.com/arran4/goa4web/a4code/a4code2html"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/searchindex"
)

// ImageBBSFeed constructs an RSS/Atom feed for the provided image posts.
func (cd *CoreData) ImageBBSFeed(r *http.Request, title string, boardID int, rows []*db.ListImagePostsByBoardForListerRow) *feeds.Feed {
	feedTitle := title
	if cd.SiteTitle != "" {
		feedTitle = fmt.Sprintf("%s - %s", cd.SiteTitle, title)
	}
	feed := &feeds.Feed{
		Title:       feedTitle,
		Link:        &feeds.Link{Href: r.URL.Path},
		Description: fmt.Sprintf("Latest posts for %s", title),
		Created:     time.Now(),
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Posted.Time.After(rows[j].Posted.Time)
	})
	if len(rows) > 15 {
		rows = rows[:15]
	}
	for _, row := range rows {
		if !row.Description.Valid {
			continue
		}
		desc := row.Description.String
		conv := a4code2html.New(cd.ImageURLMapper, a4code2html.FullImageURLMapper(cd.MapFullImageURL))
		conv.CodeType = a4code2html.CTTagStrip
		conv.SetInput(desc)
		out, _ := io.ReadAll(conv.Process())
		i := min(len(desc), 255)
		item := &feeds.Item{
			Title:   desc[:i],
			Link:    &feeds.Link{Href: fmt.Sprintf("/imagebbs/board/%d/thread/%d", boardID, row.ForumthreadID)},
			Created: time.Now(),
			Description: fmt.Sprintf("%s\n-\n%s", string(out), func() string {
				if row.Username.Valid {
					return row.Username.String
				}
				return ""
			}()),
		}
		if row.Posted.Valid {
			item.Created = row.Posted.Time
		}
		feed.Items = append(feed.Items, item)
	}
	return feed
}

// ImageBBSPoster lists posts made by the specified username.
func (cd *CoreData) ImageBBSPoster(username string, offset int32) (*ImageBBSPoster, error) {
	if cd.queries == nil {
		return &ImageBBSPoster{Username: username}, nil
	}
	u, err := cd.queries.SystemGetUserByUsername(cd.ctx, sql.NullString{String: username, Valid: true})
	if err != nil {
		return nil, err
	}
	rows, err := cd.queries.ListImagePostsByPosterForLister(cd.ctx, db.ListImagePostsByPosterForListerParams{
		ListerID:     cd.UserID,
		PosterID:     u.Idusers,
		ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:        15,
		Offset:       offset,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return &ImageBBSPoster{Posts: rows, Username: username, IsOffset: offset != 0}, nil
}

// ImageBBSPoster holds data for the poster page template.
type ImageBBSPoster struct {
	Posts    []*db.ListImagePostsByPosterForListerRow
	Username string
	IsOffset bool
}

// ImageBBSBoard gathers posts for the given board.
func (cd *CoreData) ImageBBSBoard(boardID int32) (*ImageBBSBoard, error) {
	rows, err := cd.ImageBoardPosts(boardID)
	if err != nil {
		return nil, err
	}
	return &ImageBBSBoard{BoardID: boardID, Posts: rows}, nil
}

// ImageBBSBoard represents board information with its posts.
type ImageBBSBoard struct {
	BoardID int32
	Posts   []*db.ListImagePostsByBoardForListerRow
}

// ImageBBSThread loads thread and image post details for the given board and thread.
func (cd *CoreData) ImageBBSThread(boardID, threadID int32) (*ImageBBSThread, error) {
	cd.SetCurrentThreadAndTopic(threadID, 0)
	thread, err := cd.SelectedThread()
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	post, err := cd.ImagePostByID(boardID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	return &ImageBBSThread{BoardID: int(boardID), ForumThreadID: int(threadID), ImagePost: post, Thread: thread}, nil
}

// ImageBBSThread encapsulates thread and post data for templates.
type ImageBBSThread struct {
	BoardID       int
	ForumThreadID int
	ImagePost     *db.GetImagePostByIDForListerRow
	Thread        *db.GetThreadLastPosterAndPermsForUserRow
}

// ImageBBSThreadPosts retrieves comment rows for the currently selected thread.
func (cd *CoreData) ImageBBSThreadPosts() (ImageBBSThreadPosts, error) {
	rows, err := cd.SelectedSectionThreadComments()
	if err != nil {
		return nil, err
	}
	return ImageBBSThreadPosts(rows), nil
}

// ImageBBSThreadPosts represents comments within a thread.
type ImageBBSThreadPosts []*db.GetCommentsByThreadIdForUserRow

// CreateImageBoardUpload records a post, upload metadata and its background work.
func (cd *CoreData) CreateImageBoardUpload(uploaded db.CreateUploadedImageForUploaderParams, post db.CreateImagePostForPosterParams, storedFilename string) (int64, error) {
	return cd.withSearchWork(searchindex.TypeImage, func(q db.Querier) (int64, error) {
		if _, err := q.CreateUploadedImageForUploader(cd.ctx, uploaded); err != nil {
			return 0, fmt.Errorf("record uploaded image: %w", err)
		}
		result, err := q.CreateImagePostForPoster(cd.ctx, post)
		if err != nil {
			return 0, err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		if rows == 0 {
			return 0, sql.ErrNoRows
		}
		id, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		if cd.jobService != nil && storedFilename != "" {
			if err := cd.jobService.RecordImageBoardThumbnail(cd.ctx, q, storedFilename, uploaded.UploaderID); err != nil {
				return 0, err
			}
		}
		return id, nil
	})
}
