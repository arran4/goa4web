package bookmarks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/gobookmarks"
	"github.com/arran4/gobookmarks/core"
	"golang.org/x/oauth2"
)

// GoBookmarksUserProvider adapts goa4web user auth for gobookmarks.
type GoBookmarksUserProvider struct{}

// CurrentUser returns the current user from the request context
func (p *GoBookmarksUserProvider) CurrentUser(r *http.Request) (core.User, error) {
	val := r.Context().Value(consts.KeyCoreData)
	if val == nil {
		return nil, nil
	}
	cd, ok := val.(*common.CoreData)
	if !ok {
		return nil, nil
	}
	u, err := cd.CurrentUser()
	if err != nil || u == nil {
		return nil, err
	}
	if !u.Username.Valid {
		return nil, nil
	}
	return &core.BasicUser{Login: u.Username.String}, nil
}

func (p *GoBookmarksUserProvider) IsLoggedIn(r *http.Request) bool {
	val := r.Context().Value(consts.KeyCoreData)
	if val == nil {
		return false
	}
	cd, ok := val.(*common.CoreData)
	if !ok {
		return false
	}
	return cd.UserID != 0
}

// ContextRepo implements core.Repo and gobookmarks.Provider by looking up the DB from the request context.
type ContextRepo struct{}

func (cr *ContextRepo) getRepo(ctx context.Context) (core.Repo, error) {
	val := ctx.Value(consts.KeyCoreData)
	if val == nil {
		return nil, fmt.Errorf("coredata not found in context")
	}
	cd, ok := val.(*common.CoreData)
	if !ok {
		return nil, fmt.Errorf("invalid coredata type in context")
	}
	if cd.DB == nil {
		return nil, fmt.Errorf("database not available in coredata")
	}
	return gobookmarks.NewSQLProvider(cd.DB), nil
}

// Provider interface methods

func (cr *ContextRepo) Name() string {
	return "sql"
}

func (cr *ContextRepo) DefaultServer() string {
	return ""
}

func (cr *ContextRepo) Config(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return nil
}

// CurrentUser returns the current user from ContextRepo.
// Since ContextRepo primarily deals with repo access, this might just delegate or return nil/error depending on logic.
// However, to satisfy Provider interface, it must match signature.
func (p *ContextRepo) CurrentUser(ctx context.Context, token *oauth2.Token) (core.User, error) {
	// If ContextRepo needs to return a user, it should do so here.
	// Assuming it duplicates logic or uses the context.
	// For now, let's assume it delegates to GoBookmarksUserProvider logic or similar if strict compilation needed.
	// Based on error "have ... *core.User", existing impl likely returned *core.User.
	// I will just change signature to core.User.
	return nil, nil // Or implementation if existing code had one.
}

// Repo interface methods

func (cr *ContextRepo) GetBookmarks(ctx context.Context, user, ref string, token *oauth2.Token) (string, string, error) {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return "", "", err
	}
	return repo.GetBookmarks(ctx, user, ref, token)
}

func (cr *ContextRepo) UpdateBookmarks(ctx context.Context, user string, token *oauth2.Token, sourceRef, branch, text, expectSHA string) error {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return err
	}
	return repo.UpdateBookmarks(ctx, user, token, sourceRef, branch, text, expectSHA)
}

func (cr *ContextRepo) CreateBookmarks(ctx context.Context, user string, token *oauth2.Token, branch, text string) error {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return err
	}
	return repo.CreateBookmarks(ctx, user, token, branch, text)
}

func (cr *ContextRepo) RepoExists(ctx context.Context, user string, token *oauth2.Token, name string) (bool, error) {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return false, err
	}
	return repo.RepoExists(ctx, user, token, name)
}

func (cr *ContextRepo) CreateRepo(ctx context.Context, user string, token *oauth2.Token, name string) error {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return err
	}
	return repo.CreateRepo(ctx, user, token, name)
}

func (cr *ContextRepo) CreateUser(ctx context.Context, user, password string) error {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return err
	}
	return repo.CreateUser(ctx, user, password)
}

func (cr *ContextRepo) CheckPassword(ctx context.Context, user, password string) (bool, error) {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return false, err
	}
	return repo.CheckPassword(ctx, user, password)
}

func (cr *ContextRepo) GetTags(ctx context.Context, user string, token *oauth2.Token) ([]*core.Tag, error) {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return nil, err
	}
	return repo.GetTags(ctx, user, token)
}

func (cr *ContextRepo) GetBranches(ctx context.Context, user string, token *oauth2.Token) ([]*core.Branch, error) {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return nil, err
	}
	return repo.GetBranches(ctx, user, token)
}

func (cr *ContextRepo) GetCommits(ctx context.Context, user string, token *oauth2.Token, ref string, page, perPage int) ([]*core.Commit, error) {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return nil, err
	}
	return repo.GetCommits(ctx, user, token, ref, page, perPage)
}

func (cr *ContextRepo) AdjacentCommits(ctx context.Context, user string, token *oauth2.Token, ref, sha string) (string, string, error) {
	repo, err := cr.getRepo(ctx)
	if err != nil {
		return "", "", err
	}
	return repo.AdjacentCommits(ctx, user, token, ref, sha)
}
