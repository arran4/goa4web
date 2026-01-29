package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// userCommentsAddCmd implements "user comments add".
type userCommentsAddCmd struct {
	*userCommentsCmd
	fs      *flag.FlagSet
	request adminUserCommentAddRequest
}

// adminUserCommentVisibilityInternal marks internal-only admin user comments.
const adminUserCommentVisibilityInternal = "internal"

// adminUserCommentVisibilityPublic marks admin user comments intended for public visibility.
const adminUserCommentVisibilityPublic = "public"

type adminUserCommentAddRequest struct {
	TargetID       int
	TargetUsername string
	Comment        string
	AdminID        int
	AdminUsername  string
	AsAdmin        bool
	Visibility     string
	Internal       bool
	JSON           bool
}

type adminUserCommentAudit struct {
	UserID        int32  `json:"user_id"`
	Username      string `json:"username,omitempty"`
	CommentID     int32  `json:"comment_id,omitempty"`
	Comment       string `json:"comment"`
	Visibility    string `json:"visibility"`
	Internal      bool   `json:"internal"`
	AdminUserID   int32  `json:"admin_user_id,omitempty"`
	AdminUsername string `json:"admin_username,omitempty"`
	AsAdmin       bool   `json:"as_admin"`
	CreatedAt     string `json:"created_at,omitempty"`
}

func parseUserCommentsAddCmd(parent *userCommentsCmd, args []string) (*userCommentsAddCmd, error) {
	c := &userCommentsAddCmd{userCommentsCmd: parent}
	if err := parseAdminUserCommentAddFlags("add", args, &c.request, &c.fs); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *userCommentsAddCmd) Run() error {
	return runAdminUserCommentAdd(c.rootCmd, c.request)
}

func parseAdminUserCommentAddFlags(name string, args []string, req *adminUserCommentAddRequest, fs **flag.FlagSet) error {
	parsed, _, err := parseFlags(name, args, func(fs *flag.FlagSet) {
		fs.IntVar(&req.TargetID, "id", 0, "user id")
		fs.StringVar(&req.TargetUsername, "username", "", "username")
		fs.StringVar(&req.Comment, "comment", "", "comment text")
		fs.IntVar(&req.AdminID, "admin-id", 0, "administrator user id performing this action")
		fs.StringVar(&req.AdminUsername, "admin-username", "", "administrator username performing this action")
		fs.BoolVar(&req.AsAdmin, "as-admin", false, "skip admin role check")
		fs.StringVar(&req.Visibility, "visibility", "", "comment visibility (internal or public)")
		fs.BoolVar(&req.Internal, "internal", false, "set visibility to internal")
		fs.BoolVar(&req.JSON, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return err
	}
	*fs = parsed
	return nil
}

type userIdentity struct {
	ID       int32
	Username string
}

func normalizeAdminUserCommentVisibility(visibility string, internal bool) (string, error) {
	visibility = strings.ToLower(strings.TrimSpace(visibility))
	if internal {
		if visibility != "" && visibility != adminUserCommentVisibilityInternal {
			return "", fmt.Errorf("visibility %q conflicts with --internal", visibility)
		}
		visibility = adminUserCommentVisibilityInternal
	}
	if visibility == "" {
		visibility = adminUserCommentVisibilityInternal
	}
	switch visibility {
	case adminUserCommentVisibilityInternal, adminUserCommentVisibilityPublic:
		return visibility, nil
	default:
		return "", fmt.Errorf("unsupported visibility %q", visibility)
	}
}

func lookupUserIdentity(ctx context.Context, queries db.Querier, id int, username string) (userIdentity, error) {
	if id > 0 {
		row, err := queries.SystemGetUserByID(ctx, int32(id))
		if err != nil {
			return userIdentity{}, fmt.Errorf("get user: %w", err)
		}
		resolved := userIdentity{ID: row.Idusers}
		if row.Username.Valid {
			resolved.Username = row.Username.String
		}
		if username != "" && resolved.Username != "" && resolved.Username != username {
			return userIdentity{}, fmt.Errorf("username %q does not match user id %d", username, id)
		}
		if resolved.Username == "" {
			resolved.Username = username
		}
		return resolved, nil
	}
	if username == "" {
		return userIdentity{}, fmt.Errorf("id or username required")
	}
	row, err := queries.SystemGetUserByUsername(ctx, sql.NullString{String: username, Valid: true})
	if err != nil {
		return userIdentity{}, fmt.Errorf("get user: %w", err)
	}
	resolved := userIdentity{ID: row.Idusers}
	if row.Username.Valid {
		resolved.Username = row.Username.String
	} else {
		resolved.Username = username
	}
	return resolved, nil
}

func ensureAdminIdentity(ctx context.Context, queries db.Querier, req adminUserCommentAddRequest) (userIdentity, error) {
	if req.AsAdmin {
		if req.AdminID == 0 && req.AdminUsername == "" {
			return userIdentity{}, nil
		}
		return lookupUserIdentity(ctx, queries, req.AdminID, req.AdminUsername)
	}
	if req.AdminID == 0 && req.AdminUsername == "" {
		return userIdentity{}, fmt.Errorf("admin role required; provide --admin-id, --admin-username, or --as-admin")
	}
	admin, err := lookupUserIdentity(ctx, queries, req.AdminID, req.AdminUsername)
	if err != nil {
		return userIdentity{}, err
	}
	if _, err := queries.GetAdministratorUserRole(ctx, admin.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userIdentity{}, fmt.Errorf("user %d is not an administrator", admin.ID)
		}
		return userIdentity{}, fmt.Errorf("check admin role: %w", err)
	}
	return admin, nil
}

func runAdminUserCommentAdd(root *rootCmd, req adminUserCommentAddRequest) error {
	if req.TargetID == 0 && req.TargetUsername == "" {
		return fmt.Errorf("id or username required")
	}
	if strings.TrimSpace(req.Comment) == "" {
		return fmt.Errorf("empty comment")
	}
	visibility, err := normalizeAdminUserCommentVisibility(req.Visibility, req.Internal)
	if err != nil {
		return err
	}
	conn, err := root.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)
	user, err := lookupUserIdentity(ctx, queries, req.TargetID, req.TargetUsername)
	if err != nil {
		return err
	}
	admin, err := ensureAdminIdentity(ctx, queries, req)
	if err != nil {
		return err
	}
	root.Verbosef("adding comment for user %d", user.ID)
	if err := queries.InsertAdminUserComment(ctx, db.InsertAdminUserCommentParams{UsersIdusers: user.ID, Comment: req.Comment}); err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}
	if req.JSON {
		latest, err := queries.LatestAdminUserComment(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("load latest comment: %w", err)
		}
		createdAt := latest.CreatedAt.UTC().Format(time.RFC3339)
		audit := adminUserCommentAudit{
			UserID:        user.ID,
			Username:      user.Username,
			CommentID:     latest.ID,
			Comment:       latest.Comment,
			Visibility:    visibility,
			Internal:      visibility == adminUserCommentVisibilityInternal,
			AdminUserID:   admin.ID,
			AdminUsername: admin.Username,
			AsAdmin:       req.AsAdmin,
			CreatedAt:     createdAt,
		}
		b, _ := json.MarshalIndent(audit, "", "  ")
		fmt.Println(string(b))
		return nil
	}
	root.Infof("added comment for user %d", user.ID)
	return nil
}
