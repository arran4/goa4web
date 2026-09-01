package common

import (
	"database/sql"
	"fmt"
	"image"
	"log"
	"path"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/internal/db"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/upload"
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

// ErrInvalidParticipants indicates that one or more invited participants are ineligible
// for private forum topics (e.g. they lack global privateforum:topic:see permission).
type ErrInvalidParticipants struct {
	UserIDs   []int32
	Usernames []string
}

func (e *ErrInvalidParticipants) Error() string {
	if len(e.Usernames) > 0 {
		return fmt.Sprintf("ineligible participants: %s", strings.Join(e.Usernames, ", "))
	}
	ids := make([]string, len(e.UserIDs))
	for i, id := range e.UserIDs {
		ids[i] = fmt.Sprintf("%d", id)
	}
	return fmt.Sprintf("ineligible participant user IDs: %s", strings.Join(ids, ", "))
}

// CanSeePrivateForum reports whether the given user has permission to see private forum topics.
func (cd *CoreData) CanSeePrivateForum(userID int32) bool {
	if cd == nil || cd.queries == nil || userID == 0 {
		return false
	}
	userCD := cd.ForUser(userID)
	return userCD.HasGrant("privateforum", "topic", "see", 0)
}

// CreatePrivateTopic creates a new private topic and assigns grants and subscriptions.
func (cd *CoreData) CreatePrivateTopic(p CreatePrivateTopicParams) (topicID int32, err error) {
	if cd == nil || cd.queries == nil {
		return 0, fmt.Errorf("no queries")
	}
	actorCD := cd
	if p.CreatorID != 0 && cd.UserID != p.CreatorID {
		actorCD = cd.ForUser(p.CreatorID)
	}
	if !actorCD.HasGrant("privateforum", "topic", "create", 0) {
		log.Printf("private topic create denied: user=%d", p.CreatorID)
		return 0, fmt.Errorf("permission denied")
	}

	participants := p.Participants
	if p.CreatorID != 0 {
		hasCreator := false
		for _, pt := range participants {
			if pt.ID == p.CreatorID {
				hasCreator = true
				break
			}
		}
		if !hasCreator {
			participants = append(participants, PrivateTopicParticipant{ID: p.CreatorID})
		}
	}

	hasOtherMember := false
	for _, pt := range participants {
		if pt.ID != p.CreatorID {
			hasOtherMember = true
			break
		}
	}
	if !hasOtherMember {
		return 0, fmt.Errorf("at least one other participant is required")
	}

	var invalidIDs []int32
	var invalidNames []string
	for _, pt := range participants {
		if pt.ID == p.CreatorID {
			continue
		}
		if !actorCD.CanSeePrivateForum(pt.ID) {
			invalidIDs = append(invalidIDs, pt.ID)
			name := pt.Username
			if name == "" {
				if u := actorCD.UserByID(pt.ID); u != nil {
					name = u.Username.String
				}
			}
			if name != "" {
				invalidNames = append(invalidNames, name)
			} else {
				invalidNames = append(invalidNames, fmt.Sprintf("user %d", pt.ID))
			}
		}
	}
	if len(invalidIDs) > 0 {
		return 0, &ErrInvalidParticipants{UserIDs: invalidIDs, Usernames: invalidNames}
	}

	title := p.Title
	description := p.Description
	if title == "" {
		usernames := make([]string, 0, len(participants))
		for _, participant := range participants {
			name := participant.Username
			if name == "" {
				if u := actorCD.UserByID(participant.ID); u != nil {
					name = u.Username.String
				} else {
					return 0, fmt.Errorf("unknown user %d", participant.ID)
				}
			}
			usernames = append(usernames, name)
		}
		title = fmt.Sprintf("%s%s", PrivateTopicDefaultTitlePrefix, strings.Join(usernames, ", "))
		if description == "" {
			description = title
		}
	}
	langID := actorCD.PreferredLanguageID("")
	if langID == 0 {
		langID = 1
	}
	tid, err := actorCD.queries.CreateForumTopicForPoster(actorCD.ctx, db.CreateForumTopicForPosterParams{
		PosterID:        p.CreatorID,
		ForumcategoryID: PrivateForumCategoryID,
		ForumLang:       sql.NullInt32{Int32: langID, Valid: true},
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
	for _, participant := range participants {
		uid := participant.ID
		for _, act := range []string{"see", "view", "post", "reply"} {
			if _, err := actorCD.GrantPrivateForumTopic(topicID, sql.NullInt32{Int32: uid, Valid: true}, sql.NullInt32{}, act); err != nil {
				return 0, fmt.Errorf("create %s grant %w", act, err)
			}
		}
		if err := actorCD.SubscribeTopic(uid, topicID); err != nil {
			return 0, fmt.Errorf("subscribe topic %w", err)
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

func thumbnailFilename(imageID, ext string, size config.ThumbnailSize) string {
	return imageID + "_thumb_" + strconv.Itoa(size.Width) + "x" + strconv.Itoa(size.Height) + ext
}

// UploadedImageByImageID returns an uploaded image using its file identifier.
func (cd *CoreData) UploadedImageByImageID(imageID string) (*db.UploadedImage, error) {
	if cd == nil || cd.queries == nil {
		return nil, nil
	}
	imagePath, err := imageIDToUploadPath(imageID)
	if err != nil {
		return nil, err
	}
	return cd.queries.GetUploadedImageByPath(cd.ctx, sql.NullString{String: imagePath, Valid: true})
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
	return cd.storeImageInternal(p)
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
	return cd.storeImageInternal(p)
}

func (cd *CoreData) storeImageInternal(p StoreImageParams) (string, error) {
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

	generator := "bild"
	size := config.ThumbnailSize{Width: config.DefaultImageThumbnailWidth, Height: config.DefaultImageThumbnailHeight}
	if cfg != nil {
		if cfg.ImageThumbnailGenerator != "" {
			generator = cfg.ImageThumbnailGenerator
		}
		size = cfg.ThumbnailSizes()[0]
	}
	thumbBytes, err := imagesign.GenerateThumbnailWithinBounds(p.Image, p.Ext, generator, size.Height, size.Width)
	if err != nil {
		return "", fmt.Errorf("generate thumbnail %w", err)
	}

	thumbName := thumbnailFilename(p.ID, p.Ext, size)
	url := path.Join("/", sub1, sub2, fname)
	uploadedImageID, err := cd.queries.CreateUploadedImageForUploader(cd.ctx, db.CreateUploadedImageForUploaderParams{
		UploaderID: p.UploaderID,
		Path:       sql.NullString{String: url, Valid: true},
		Width:      sql.NullInt32{Int32: int32(width), Valid: true},
		Height:     sql.NullInt32{Int32: int32(height), Valid: true},
		FileSize:   int32(len(p.Data)),
	})
	if err != nil {
		return "", fmt.Errorf("create uploaded image %w", err)
	}
	if cp := upload.CacheProviderFromConfig(cfg); cp != nil {
		if err := cp.Write(cd.ctx, path.Join(sub1, sub2, thumbName), thumbBytes); err != nil {
			log.Printf("cache write: %v", err)
			return "", fmt.Errorf("cache write %w", err)
		}
		source := &db.UploadedImage{
			Iduploadedimage: int32(uploadedImageID),
			Path:            sql.NullString{String: url, Valid: true},
		}
		thumbnailHeight, thumbnailWidth, err := imagesign.DimensionsWithinBounds(p.Image, size.Height, size.Width)
		if err != nil {
			return "", fmt.Errorf("thumbnail dimensions %w", err)
		}
		if err := cd.RecordUploadedImageThumbnail(cd.ctx, thumbName, source, thumbBytes, thumbnailHeight, thumbnailWidth); err != nil {
			return "", fmt.Errorf("record image cache entry %w", err)
		}
		if ccp, ok := cp.(upload.CacheProvider); ok {
			if err := ccp.Cleanup(cd.ctx, int64(cfg.ImageCacheMaxBytes)); err != nil {
				log.Printf("cache cleanup: %v", err)
			}
		}
	}
	return fname, nil
}

// GrantRole creates an administrative grant for a named role.
func (cd *CoreData) GrantRole(roleName, section, item, action string) error {
	role, err := cd.queries.GetRoleByName(cd.ctx, roleName)
	if err != nil {
		return fmt.Errorf("lookup role %q: %w", roleName, err)
	}
	_, err = cd.queries.AdminCreateGrant(cd.ctx, db.AdminCreateGrantParams{
		RoleID:   sql.NullInt32{Int32: role.ID, Valid: true},
		Section:  section,
		Item:     sql.NullString{String: item, Valid: item != ""},
		RuleType: "allow",
		ItemID:   sql.NullInt32{},
		ItemRule: sql.NullString{},
		Action:   action,
		Extra:    sql.NullString{},
	})
	if err != nil {
		return fmt.Errorf("create grant for role %q (%s/%s/%s): %w", roleName, section, item, action, err)
	}
	return nil
}
