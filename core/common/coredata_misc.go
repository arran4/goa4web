package common

import (
	"bytes"
	"database/sql"
	"fmt"
	"image"
	"log"
	"path"

	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/upload"
	"github.com/arran4/goa4web/workers/postcountworker"
	"github.com/arran4/goa4web/workers/searchworker"
	"golang.org/x/image/draw"
)

// CreatePrivateTopicParams groups input for CreatePrivateTopic.
type CreatePrivateTopicParams struct {
	CreatorID      int32
	ParticipantIDs []int32
	Body           string
}

// CreatePrivateTopic creates a new private topic and assigns grants and the initial comment.
func (cd *CoreData) CreatePrivateTopic(p CreatePrivateTopicParams) (topicID, threadID int32, err error) {
	if cd == nil || cd.queries == nil {
		return 0, 0, fmt.Errorf("no queries")
	}
	if !cd.HasGrant("privateforum", "topic", "create", 0) {
		log.Printf("private topic create denied: user=%d", p.CreatorID)
		return 0, 0, fmt.Errorf("permission denied")
	}
	tid, err := cd.queries.CreateForumTopicForPoster(cd.ctx, db.CreateForumTopicForPosterParams{
		PosterID:        p.CreatorID,
		ForumcategoryID: PrivateForumCategoryID,
		LanguageID:      0,
		Title:           sql.NullString{},
		Description:     sql.NullString{},
		Handler:         "private",
		Section:         "privateforum",
		GrantCategoryID: sql.NullInt32{Int32: PrivateForumCategoryID, Valid: true},
		GranteeID:       sql.NullInt32{Int32: p.CreatorID, Valid: p.CreatorID != 0},
	})
	if err != nil {
		return 0, 0, fmt.Errorf("create topic %w", err)
	}
	if tid == 0 {
		return 0, 0, fmt.Errorf("create topic returned 0")
	}
	topicID = int32(tid)
	thid, err := cd.queries.SystemCreateThread(cd.ctx, topicID)
	if err != nil {
		return 0, 0, fmt.Errorf("create thread %w", err)
	}
	threadID = int32(thid)
	for _, uid := range p.ParticipantIDs {
		for _, act := range []string{"see", "view", "post", "reply"} {
			if _, err := cd.queries.SystemCreateGrant(cd.ctx, db.SystemCreateGrantParams{
				UserID:   sql.NullInt32{Int32: uid, Valid: true},
				RoleID:   sql.NullInt32{},
				Section:  "forum",
				Item:     sql.NullString{String: "topic", Valid: true},
				RuleType: "allow",
				ItemID:   sql.NullInt32{Int32: topicID, Valid: true},
				ItemRule: sql.NullString{},
				Action:   act,
				Extra:    sql.NullString{},
			}); err != nil {
				return 0, 0, fmt.Errorf("create %s grant %w", act, err)
			}
		}
	}
	cid, err := cd.CreateForumCommentForCommenter(p.CreatorID, threadID, topicID, 0, p.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("create comment %w", err)
	}
	if evt := cd.Event(); evt != nil {
		if evt.Data == nil {
			evt.Data = map[string]any{}
		}
		evt.Data[postcountworker.EventKey] = postcountworker.UpdateEventData{CommentID: int32(cid), ThreadID: threadID, TopicID: topicID}
		evt.Data[searchworker.EventKey] = searchworker.IndexEventData{Type: searchworker.TypeComment, ID: int32(cid), Text: p.Body}
	}
	return topicID, threadID, nil
}

// StoreImageParams groups input for StoreImage.
type StoreImageParams struct {
	ID         string
	Ext        string
	Data       []byte
	Image      image.Image
	UploaderID int32
}

// StoreImage stores the image bytes, generates thumbnails and records metadata.
func (cd *CoreData) StoreImage(p StoreImageParams) (string, error) {
	if cd == nil || cd.queries == nil {
		return "", fmt.Errorf("no queries")
	}
	cfg := cd.Config
	sub1, sub2 := p.ID[:2], p.ID[2:4]
	fname := p.ID + p.Ext
	if prov := upload.ProviderFromConfig(cfg); prov != nil {
		if err := prov.Write(cd.ctx, path.Join(sub1, sub2, fname), p.Data); err != nil {
			log.Printf("upload write: %v", err)
			return "", fmt.Errorf("upload write %w", err)
		}
	}
	width := p.Image.Bounds().Dx()
	height := p.Image.Bounds().Dy()
	src := p.Image.Bounds()
	var crop image.Rectangle
	if src.Dx() > src.Dy() {
		side := src.Dy()
		x0 := src.Min.X + (src.Dx()-side)/2
		crop = image.Rect(x0, src.Min.Y, x0+side, src.Min.Y+side)
	} else {
		side := src.Dx()
		y0 := src.Min.Y + (src.Dy()-side)/2
		crop = image.Rect(src.Min.X, y0, src.Min.X+side, src.Min.Y+side)
	}
	thumbName := p.ID + "_thumb" + p.Ext
	var tbuf bytes.Buffer
	thumb := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), p.Image, crop, draw.Over, nil)
	enc, err := imagesign.EncoderByExtension(p.Ext)
	if err != nil {
		return "", fmt.Errorf("encoder ext %w", err)
	}
	if err := enc(&tbuf, thumb); err != nil {
		return "", fmt.Errorf("thumb encode %w", err)
	}
	if cp := upload.CacheProviderFromConfig(cfg); cp != nil {
		if err := cp.Write(cd.ctx, path.Join(sub1, sub2, thumbName), tbuf.Bytes()); err != nil {
			log.Printf("cache write: %v", err)
			return "", fmt.Errorf("cache write %w", err)
		}
		if ccp, ok := cp.(upload.CacheProvider); ok {
			if err := ccp.Cleanup(cd.ctx, int64(cfg.ImageCacheMaxBytes)); err != nil {
				log.Printf("cache cleanup: %v", err)
			}
		}
	}
	url := path.Join("/uploads", sub1, sub2, fname)
	_, err = cd.queries.CreateUploadedImageForUploader(cd.ctx, db.CreateUploadedImageForUploaderParams{
		UploaderID: p.UploaderID,
		Path:       sql.NullString{String: url, Valid: true},
		Width:      sql.NullInt32{Int32: int32(width), Valid: true},
		Height:     sql.NullInt32{Int32: int32(height), Valid: true},
		FileSize:   int32(len(p.Data)),
		GranteeID:  sql.NullInt32{Int32: p.UploaderID, Valid: p.UploaderID != 0},
	})
	if err != nil {
		return "", fmt.Errorf("create uploaded image %w", err)
	}
	return fname, nil
}
