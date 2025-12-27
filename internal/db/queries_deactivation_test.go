package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"testing"
)

type deactivationEntry[T any] struct {
	row      T
	restored bool
}

type deactivationFake struct {
	mu         sync.Mutex
	users      map[int32]*deactivationEntry[AdminListDeactivatedUsersRow]
	blogs      map[int32]*deactivationEntry[AdminListDeactivatedBlogsRow]
	comments   map[int32]*deactivationEntry[AdminListDeactivatedCommentsRow]
	writings   map[int32]*deactivationEntry[AdminListDeactivatedWritingsRow]
	links      map[int32]*deactivationEntry[AdminListDeactivatedLinksRow]
	imageposts map[int32]*deactivationEntry[AdminListDeactivatedImagepostsRow]
}

func newDeactivationFake() *deactivationFake {
	return &deactivationFake{
		users:      make(map[int32]*deactivationEntry[AdminListDeactivatedUsersRow]),
		blogs:      make(map[int32]*deactivationEntry[AdminListDeactivatedBlogsRow]),
		comments:   make(map[int32]*deactivationEntry[AdminListDeactivatedCommentsRow]),
		writings:   make(map[int32]*deactivationEntry[AdminListDeactivatedWritingsRow]),
		links:      make(map[int32]*deactivationEntry[AdminListDeactivatedLinksRow]),
		imageposts: make(map[int32]*deactivationEntry[AdminListDeactivatedImagepostsRow]),
	}
}

func (f *deactivationFake) AdminArchiveUser(_ context.Context, idusers int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[idusers] = &deactivationEntry[AdminListDeactivatedUsersRow]{
		row: AdminListDeactivatedUsersRow{
			Idusers:  idusers,
			Email:    sql.NullString{String: fmt.Sprintf("user-%d@example.com", idusers), Valid: true},
			Username: sql.NullString{String: fmt.Sprintf("user-%d", idusers), Valid: true},
		},
	}
	return nil
}

func (f *deactivationFake) AdminArchiveBlog(_ context.Context, arg AdminArchiveBlogParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blogs[arg.Idblogs] = &deactivationEntry[AdminListDeactivatedBlogsRow]{
		row: AdminListDeactivatedBlogsRow{
			Idblogs: arg.Idblogs,
			Blog:    arg.Blog,
		},
	}
	return nil
}

func (f *deactivationFake) AdminArchiveComment(_ context.Context, arg AdminArchiveCommentParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.comments[arg.Idcomments] = &deactivationEntry[AdminListDeactivatedCommentsRow]{
		row: AdminListDeactivatedCommentsRow{
			Idcomments: arg.Idcomments,
			Text:       arg.Text,
		},
	}
	return nil
}

func (f *deactivationFake) AdminArchiveWriting(_ context.Context, arg AdminArchiveWritingParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.writings[arg.Idwriting] = &deactivationEntry[AdminListDeactivatedWritingsRow]{
		row: AdminListDeactivatedWritingsRow{
			Idwriting: arg.Idwriting,
			Title:     arg.Title,
			Writing:   arg.Writing,
			Abstract:  arg.Abstract,
			Private:   arg.Private,
		},
	}
	return nil
}

func (f *deactivationFake) AdminArchiveLink(_ context.Context, arg AdminArchiveLinkParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.links[arg.ID] = &deactivationEntry[AdminListDeactivatedLinksRow]{
		row: AdminListDeactivatedLinksRow{
			ID:          arg.ID,
			Title:       arg.Title,
			Url:         arg.Url,
			Description: arg.Description,
		},
	}
	return nil
}

func (f *deactivationFake) AdminArchiveImagepost(_ context.Context, arg AdminArchiveImagepostParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.imageposts[arg.Idimagepost] = &deactivationEntry[AdminListDeactivatedImagepostsRow]{
		row: AdminListDeactivatedImagepostsRow{
			Idimagepost: arg.Idimagepost,
			Description: arg.Description,
			Thumbnail:   arg.Thumbnail,
			Fullimage:   arg.Fullimage,
		},
	}
	return nil
}

func (f *deactivationFake) AdminRestoreUser(_ context.Context, idusers int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	entry, ok := f.users[idusers]
	if !ok {
		return fmt.Errorf("user %d not archived", idusers)
	}
	entry.restored = true
	return nil
}

func (f *deactivationFake) AdminMarkBlogRestored(_ context.Context, idblogs int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return markRestored(f.blogs, idblogs)
}

func (f *deactivationFake) AdminMarkCommentRestored(_ context.Context, idcomments int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return markRestored(f.comments, idcomments)
}

func (f *deactivationFake) AdminMarkWritingRestored(_ context.Context, idwriting int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return markRestored(f.writings, idwriting)
}

func (f *deactivationFake) AdminMarkLinkRestored(_ context.Context, id int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return markRestored(f.links, id)
}

func (f *deactivationFake) AdminMarkImagepostRestored(_ context.Context, idimagepost int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	return markRestored(f.imageposts, idimagepost)
}

func (f *deactivationFake) AdminIsUserDeactivated(_ context.Context, idusers int32) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return isDeactivated(f.users, idusers), nil
}

func (f *deactivationFake) AdminIsBlogDeactivated(_ context.Context, idblogs int32) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return isDeactivated(f.blogs, idblogs), nil
}

func (f *deactivationFake) AdminIsCommentDeactivated(_ context.Context, idcomments int32) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return isDeactivated(f.comments, idcomments), nil
}

func (f *deactivationFake) AdminIsWritingDeactivated(_ context.Context, idwriting int32) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return isDeactivated(f.writings, idwriting), nil
}

func (f *deactivationFake) AdminIsLinkDeactivated(_ context.Context, id int32) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return isDeactivated(f.links, id), nil
}

func (f *deactivationFake) AdminIsImagepostDeactivated(_ context.Context, idimagepost int32) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return isDeactivated(f.imageposts, idimagepost), nil
}

func (f *deactivationFake) AdminListDeactivatedUsers(_ context.Context, arg AdminListDeactivatedUsersParams) ([]*AdminListDeactivatedUsersRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listRows(f.users, arg.Limit, arg.Offset), nil
}

func (f *deactivationFake) AdminListDeactivatedBlogs(_ context.Context, arg AdminListDeactivatedBlogsParams) ([]*AdminListDeactivatedBlogsRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listRows(f.blogs, arg.Limit, arg.Offset), nil
}

func (f *deactivationFake) AdminListDeactivatedComments(_ context.Context, arg AdminListDeactivatedCommentsParams) ([]*AdminListDeactivatedCommentsRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listRows(f.comments, arg.Limit, arg.Offset), nil
}

func (f *deactivationFake) AdminListDeactivatedWritings(_ context.Context, arg AdminListDeactivatedWritingsParams) ([]*AdminListDeactivatedWritingsRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listRows(f.writings, arg.Limit, arg.Offset), nil
}

func (f *deactivationFake) AdminListDeactivatedLinks(_ context.Context, arg AdminListDeactivatedLinksParams) ([]*AdminListDeactivatedLinksRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listRows(f.links, arg.Limit, arg.Offset), nil
}

func (f *deactivationFake) AdminListDeactivatedImageposts(_ context.Context, arg AdminListDeactivatedImagepostsParams) ([]*AdminListDeactivatedImagepostsRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return listRows(f.imageposts, arg.Limit, arg.Offset), nil
}

func isDeactivated[T any](entries map[int32]*deactivationEntry[T], id int32) bool {
	entry, ok := entries[id]
	return ok && !entry.restored
}

func markRestored[T any](entries map[int32]*deactivationEntry[T], id int32) error {
	entry, ok := entries[id]
	if !ok {
		return fmt.Errorf("id %d not archived", id)
	}
	entry.restored = true
	return nil
}

func listRows[T any](entries map[int32]*deactivationEntry[T], limit, offset int32) []*T {
	ids := make([]int, 0, len(entries))
	for id, entry := range entries {
		if entry.restored {
			continue
		}
		ids = append(ids, int(id))
	}

	sort.Ints(ids)
	start := int(offset)
	if start >= len(ids) {
		return nil
	}
	end := len(ids)
	if limit > 0 && start+int(limit) < end {
		end = start + int(limit)
	}

	rows := make([]*T, 0, end-start)
	for _, id := range ids[start:end] {
		rowCopy := entries[int32(id)].row
		rows = append(rows, &rowCopy)
	}
	return rows
}

func TestDeactivationFakeUsers(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := newDeactivationFake()

	if is, err := q.AdminIsUserDeactivated(ctx, 1); err != nil || is {
		t.Fatalf("expected user to start active, got is=%v err=%v", is, err)
	}

	if err := q.AdminArchiveUser(ctx, 1); err != nil {
		t.Fatalf("AdminArchiveUser: %v", err)
	}

	if is, err := q.AdminIsUserDeactivated(ctx, 1); err != nil || !is {
		t.Fatalf("expected user deactivated, got is=%v err=%v", is, err)
	}

	users, err := q.AdminListDeactivatedUsers(ctx, AdminListDeactivatedUsersParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedUsers: %v", err)
	}
	if len(users) != 1 || users[0].Idusers != 1 || users[0].Username.String == "" {
		t.Fatalf("unexpected deactivated users: %+v", users)
	}

	if err := q.AdminRestoreUser(ctx, 1); err != nil {
		t.Fatalf("AdminRestoreUser: %v", err)
	}
	if is, _ := q.AdminIsUserDeactivated(ctx, 1); is {
		t.Fatalf("expected user restored")
	}
}

func TestDeactivationFakeBlogs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := newDeactivationFake()

	if err := q.AdminArchiveBlog(ctx, AdminArchiveBlogParams{
		Idblogs: 2,
		Blog:    sql.NullString{String: "first blog", Valid: true},
	}); err != nil {
		t.Fatalf("AdminArchiveBlog: %v", err)
	}
	if is, _ := q.AdminIsBlogDeactivated(ctx, 2); !is {
		t.Fatalf("blog should be deactivated")
	}

	rows, err := q.AdminListDeactivatedBlogs(ctx, AdminListDeactivatedBlogsParams{Limit: 1, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedBlogs: %v", err)
	}
	if len(rows) != 1 || rows[0].Blog.String != "first blog" {
		t.Fatalf("unexpected blog rows: %+v", rows)
	}

	if err := q.AdminMarkBlogRestored(ctx, 2); err != nil {
		t.Fatalf("AdminMarkBlogRestored: %v", err)
	}
	if is, _ := q.AdminIsBlogDeactivated(ctx, 2); is {
		t.Fatalf("expected restored blog to be active")
	}
}

func TestDeactivationFakeCommentsRespectsOffset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := newDeactivationFake()

	for i := int32(1); i <= 2; i++ {
		if err := q.AdminArchiveComment(ctx, AdminArchiveCommentParams{
			Idcomments: i,
			Text:       sql.NullString{String: fmt.Sprintf("comment-%d", i), Valid: true},
		}); err != nil {
			t.Fatalf("AdminArchiveComment %d: %v", i, err)
		}
	}

	rows, err := q.AdminListDeactivatedComments(ctx, AdminListDeactivatedCommentsParams{Limit: 1, Offset: 1})
	if err != nil {
		t.Fatalf("AdminListDeactivatedComments: %v", err)
	}
	if len(rows) != 1 || rows[0].Idcomments != 2 {
		t.Fatalf("expected second comment after offset, got %+v", rows)
	}

	if err := q.AdminMarkCommentRestored(ctx, 2); err != nil {
		t.Fatalf("AdminMarkCommentRestored: %v", err)
	}
	if is, _ := q.AdminIsCommentDeactivated(ctx, 2); is {
		t.Fatalf("expected comment 2 restored")
	}
}

func TestDeactivationFakeWritings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := newDeactivationFake()

	if err := q.AdminArchiveWriting(ctx, AdminArchiveWritingParams{
		Idwriting: 4,
		Title:     sql.NullString{String: "draft", Valid: true},
		Writing:   sql.NullString{String: "body", Valid: true},
		Abstract:  sql.NullString{String: "summary", Valid: true},
		Private:   sql.NullBool{Bool: true, Valid: true},
	}); err != nil {
		t.Fatalf("AdminArchiveWriting: %v", err)
	}

	if is, _ := q.AdminIsWritingDeactivated(ctx, 4); !is {
		t.Fatalf("writing should be deactivated")
	}

	rows, err := q.AdminListDeactivatedWritings(ctx, AdminListDeactivatedWritingsParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedWritings: %v", err)
	}
	if len(rows) != 1 || rows[0].Private.Bool != true {
		t.Fatalf("unexpected writing rows: %+v", rows)
	}

	if err := q.AdminMarkWritingRestored(ctx, 4); err != nil {
		t.Fatalf("AdminMarkWritingRestored: %v", err)
	}
	if is, _ := q.AdminIsWritingDeactivated(ctx, 4); is {
		t.Fatalf("expected writing restored")
	}
}

func TestDeactivationFakeLinks(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := newDeactivationFake()

	if err := q.AdminArchiveLink(ctx, AdminArchiveLinkParams{
		ID:          6,
		Title:       sql.NullString{String: "title", Valid: true},
		Url:         sql.NullString{String: "https://example.com", Valid: true},
		Description: sql.NullString{String: "desc", Valid: true},
	}); err != nil {
		t.Fatalf("AdminArchiveLink: %v", err)
	}

	if is, _ := q.AdminIsLinkDeactivated(ctx, 6); !is {
		t.Fatalf("link should be deactivated")
	}

	rows, err := q.AdminListDeactivatedLinks(ctx, AdminListDeactivatedLinksParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedLinks: %v", err)
	}
	if len(rows) != 1 || rows[0].Url.String == "" {
		t.Fatalf("unexpected link rows: %+v", rows)
	}

	if err := q.AdminMarkLinkRestored(ctx, 6); err != nil {
		t.Fatalf("AdminMarkLinkRestored: %v", err)
	}
	if is, _ := q.AdminIsLinkDeactivated(ctx, 6); is {
		t.Fatalf("expected link restored")
	}
}

func TestDeactivationFakeImageposts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	q := newDeactivationFake()

	if err := q.AdminArchiveImagepost(ctx, AdminArchiveImagepostParams{
		Idimagepost: 5,
		Description: sql.NullString{String: "pic", Valid: true},
		Thumbnail:   sql.NullString{String: "thumb", Valid: true},
		Fullimage:   sql.NullString{String: "full", Valid: true},
	}); err != nil {
		t.Fatalf("AdminArchiveImagepost: %v", err)
	}

	if is, _ := q.AdminIsImagepostDeactivated(ctx, 5); !is {
		t.Fatalf("imagepost should be deactivated")
	}

	rows, err := q.AdminListDeactivatedImageposts(ctx, AdminListDeactivatedImagepostsParams{Limit: 5, Offset: 0})
	if err != nil {
		t.Fatalf("AdminListDeactivatedImageposts: %v", err)
	}
	if len(rows) != 1 || rows[0].Fullimage.String != "full" {
		t.Fatalf("unexpected imagepost rows: %+v", rows)
	}

	if err := q.AdminMarkImagepostRestored(ctx, 5); err != nil {
		t.Fatalf("AdminMarkImagepostRestored: %v", err)
	}
	if is, _ := q.AdminIsImagepostDeactivated(ctx, 5); is {
		t.Fatalf("expected imagepost restored")
	}
}
