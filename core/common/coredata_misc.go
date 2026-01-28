package common

import (
	"database/sql"
	"fmt"
	"image"
	"log"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
)

// CreatePrivateTopicParams groups input for CreatePrivateTopic.
type CreatePrivateTopicParams struct {
	CreatorID    int32
	Participants []PrivateTopicParticipant
	Title        string
	Description  string
}

// PrivateTopicParticipant pairs a participant ID with an optional username.
type PrivateTopicParticipant struct {
	ID       int32
	Username string
}

// CreatePrivateTopic creates a new private topic and assigns grants and the initial comment.
func (cd *CoreData) CreatePrivateTopic(p CreatePrivateTopicParams) (topicID int32, err error) {
	if cd == nil || cd.queries == nil {
		return 0, fmt.Errorf("no queries")
	}
	if !cd.HasGrant("privateforum", "topic", "create", 0) {
		log.Printf("private topic create denied: user=%d", p.CreatorID)
		return 0, fmt.Errorf("permission denied")
	}
	usernames := make([]string, 0, len(p.Participants))
	for _, participant := range p.Participants {
		name := participant.Username
		if name == "" {
			if u := cd.UserByID(participant.ID); u != nil {
				name = u.Username.String
			} else {
				return 0, fmt.Errorf("unknown user %d", participant.ID)
			}
		}
		usernames = append(usernames, name)
	}
	title := p.Title
	description := p.Description
	if title == "" {
		title = fmt.Sprintf("Private chat with %s", strings.Join(usernames, ", "))
		if description == "" {
			description = title
		}
	}
	tid, err := cd.queries.CreateForumTopicForPoster(cd.ctx, db.CreateForumTopicForPosterParams{
		PosterID:        p.CreatorID,
		ForumcategoryID: PrivateForumCategoryID,
		ForumLang:       sql.NullInt32{},
		Title:           sql.NullString{String: title, Valid: true},
		Description:     sql.NullString{String: description, Valid: true},
		Handler:         "private",
		Section:         "privateforum",
		GrantCategoryID: sql.NullInt32{Int32: PrivateForumCategoryID, Valid: true},
		GranteeID:       sql.NullInt32{Int32: p.CreatorID, Valid: p.CreatorID != 0},
	})
	if err != nil {
		return 0, fmt.Errorf("create topic %w", err)
	}
	if tid == 0 {
		return 0, fmt.Errorf("create topic returned 0")
	}
	topicID = int32(tid)
	for _, participant := range p.Participants {
		uid := participant.ID
		for _, act := range []string{"see", "view", "post", "reply", "edit"} {
			if _, err := cd.GrantPrivateForumTopic(topicID, sql.NullInt32{Int32: uid, Valid: true}, sql.NullInt32{}, act); err != nil {
				return 0, fmt.Errorf("create %s grant %w", act, err)
			}
		}
	}
	return topicID, nil
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
	if !imagesign.ValidID(p.ID) {
		return "", fmt.Errorf("invalid id")
	}
	if !imagesign.AllowedExtension(p.Ext) {
		return "", fmt.Errorf("unsupported image extension: %s", p.Ext)
	}
	if !cd.HasGrant("images", "upload", "post", 0) {
		return "", fmt.Errorf("permission denied")
	}
	return imagesign.StoreSystemImage(cd.ctx, cd.queries, cd.Config, imagesign.StoreImageParams{
		ID:         p.ID,
		Ext:        p.Ext,
		Data:       p.Data,
		Image:      p.Image,
		UploaderID: p.UploaderID,
	})
}

// StoreSystemImage stores the image bytes as a system upload (bypassing user grant checks).
func (cd *CoreData) StoreSystemImage(p StoreImageParams) (string, error) {
	if cd == nil || cd.queries == nil {
		return "", fmt.Errorf("no queries")
	}
	if !imagesign.ValidID(p.ID) {
		return "", fmt.Errorf("invalid id")
	}
	if !imagesign.AllowedExtension(p.Ext) {
		return "", fmt.Errorf("unsupported image extension: %s", p.Ext)
	}
	// System upload: no grant check needed, but ensure uploader is system/admin or 0
	return imagesign.StoreSystemImage(cd.ctx, cd.queries, cd.Config, imagesign.StoreImageParams{
		ID:         p.ID,
		Ext:        p.Ext,
		Data:       p.Data,
		Image:      p.Image,
		UploaderID: p.UploaderID,
	})
}
