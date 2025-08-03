package common

// TODO: sort CoreData struct fields and related methods alphabetically.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"github.com/gorilla/sessions"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/eventbus"
	imagesign "github.com/arran4/goa4web/internal/images"
	"github.com/arran4/goa4web/internal/lazy"
	linksign "github.com/arran4/goa4web/internal/linksign"
	"github.com/arran4/goa4web/internal/tasks"
)

// IndexItem represents a navigation item linking to site sections.
type IndexItem struct {
	Name string
	Link string
}

// AdminSection groups admin navigation links under a section heading.
type AdminSection struct {
	Name  string
	Links []IndexItem
}

// SessionManager defines optional hooks for storing and removing session
// information. Implementations may persist session metadata in a database or
// other storage while exposing a storage-agnostic API to CoreData.
type SessionManager interface {
	InsertSession(ctx context.Context, sessionID string, userID int32) error
	DeleteSessionByID(ctx context.Context, sessionID string) error
}

// MailProvider defines the interface required by CoreData for sending emails.
type MailProvider interface {
	Send(ctx context.Context, to mail.Address, rawEmailMessage []byte) error
}

// NavigationProvider exposes index and admin navigation links.
type NavigationProvider interface {
	IndexItems() []IndexItem
	AdminLinks() []IndexItem
	AdminSections() []AdminSection
}

// No package-level pagination constants as runtime config provides these values.

// NewsPost describes a news entry with access metadata.
type NewsPost struct {
	*db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow
	ShowReply    bool
	ShowEdit     bool
	Editing      bool
	Announcement *db.SiteAnnouncement
}

type CoreData struct {
	IndexItems       []IndexItem
	CustomIndexItems []IndexItem
	UserID           int32
	// PageTitle holds the title of the current page.
	PageTitle    string
	Title        string
	AutoRefresh  string
	FeedsEnabled bool
	RSSFeedUrl   string
	AtomFeedUrl  string
	// AdminMode indicates whether admin-only UI elements should be displayed.
	AdminMode         bool
	NotificationCount int32
	Config            *config.RuntimeConfig
	ImageSigner       *imagesign.Signer
	LinkSigner        *linksign.Signer
	Nav               NavigationProvider
	mapMu             sync.Mutex
	TasksReg          *tasks.Registry
	a4codeMapper      func(tag, val string) string

	session        *sessions.Session
	sessionManager SessionManager

	ctx           context.Context
	queries       db.Querier
	customQueries db.CustomQueries
	emailProvider lazy.Value[MailProvider]

	allRoles                 lazy.Value[[]*db.Role]
	announcement             lazy.Value[*db.GetActiveAnnouncementWithNewsForListerRow]
	annMu                    sync.Mutex
	bloggers                 lazy.Value[[]*db.ListBloggersForListerRow]
	bookmarks                lazy.Value[*db.GetBookmarksForUserRow]
	event                    *eventbus.TaskEvent
	forumCategories          lazy.Value[[]*db.Forumcategory]
	forumThreads             map[int32]*lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]
	forumTopicLists          map[int32]*lazy.Value[[]*db.Forumtopic]
	forumTopics              map[int32]*lazy.Value[*db.GetForumTopicByIdForUserRow]
	forumThreadRows          map[int32]*lazy.Value[*db.GetThreadLastPosterAndPermsRow]
	forumComments            map[int32]*lazy.Value[*db.GetCommentByIdForUserRow]
	forumThreadComments      map[int32]*lazy.Value[[]*db.GetCommentsByThreadIdForUserRow]
	newsPosts                map[int32]*lazy.Value[*db.GetForumThreadIdByNewsPostIdRow]
	threadComments           map[int32]*lazy.Value[[]*db.GetCommentsByThreadIdForUserRow]
	currentThreadID          int32
	currentTopicID           int32
	currentCommentID         int32
	currentNewsPostID        int32
	currentBoardID           int32
	currentImagePostID       int32
	imageBoardPosts          map[int32]*lazy.Value[[]*db.ListImagePostsByBoardForListerRow]
	imageBoards              lazy.Value[[]*db.Imageboard]
	imagePostRows            map[int32]*lazy.Value[*db.GetImagePostByIDForListerRow]
	languagesAll             lazy.Value[[]*db.Language]
	langs                    lazy.Value[[]*db.Language]
	latestNews               lazy.Value[[]*NewsPost]
	latestWritings           lazy.Value[[]*db.Writing]
	adminLinkerItemRows      map[int32]*lazy.Value[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow]
	linkerCategories         lazy.Value[[]*db.GetLinkerCategoryLinkCountsRow]
	linkerCategoryLinks      map[int32]*lazy.Value[[]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow]
	linkerCategoryRows       map[int32]*lazy.Value[*db.LinkerCategory]
	linkerCatsAll            lazy.Value[[]*db.LinkerCategory]
	linkerCatsForUser        lazy.Value[[]*db.LinkerCategory]
	externalLinks            map[string]*lazy.Value[*db.ExternalLink]
	newsAnnouncements        map[int32]*lazy.Value[*db.SiteAnnouncement]
	notifCount               lazy.Value[int32]
	notifications            map[string]*lazy.Value[[]*db.Notification]
	perms                    lazy.Value[[]*db.GetPermissionsByUserIDRow]
	pref                     lazy.Value[*db.Preference]
	preferredLanguageID      lazy.Value[int32]
	publicWritings           map[string]*lazy.Value[[]*db.ListPublicWritingsInCategoryForListerRow]
	subImageBoards           map[int32]*lazy.Value[[]*db.Imageboard]
	unreadCount              lazy.Value[int64]
	subscriptions            lazy.Value[map[string]bool]
	subscriptionRows         lazy.Value[[]*db.ListSubscriptionsByUserRow]
	userSubscriptions        lazy.Value[[]*db.ListSubscriptionsByUserRow]
	user                     lazy.Value[*db.User]
	userRoles                lazy.Value[[]string]
	visibleWritingCategories lazy.Value[[]*db.WritingCategory]
	writerWritings           map[int32]*lazy.Value[[]*db.ListPublicWritingsByUserForListerRow]
	writers                  lazy.Value[[]*db.ListWritersForListerRow]
	writingCategories        lazy.Value[[]*db.WritingCategory]
	currentWritingID         int32
	writingRows              map[int32]*lazy.Value[*db.GetWritingForListerByIDRow]
	currentBlogID            int32
	blogEntries              map[int32]*lazy.Value[*db.GetBlogEntryForListerByIDRow]
	blogListOffset           int
	blogListRows             lazy.Value[[]*db.ListBlogEntriesByAuthorForListerRow]
	blogListUID              int32

	absoluteURLBase lazy.Value[string]
	dbRegistry      *dbdrivers.Registry
	// marks records which template sections have been rendered to avoid
	// duplicate output when re-rendering after an error.
	marks map[string]struct{}
}

// SetRoles preloads the current user roles.
func (cd *CoreData) SetRoles(r []string) { cd.userRoles.Set(r) }

// Marked returns true the first time it is called with key. Subsequent
// calls return false. It is used to avoid re-rendering template sections
// when streaming pages after an error.
func (cd *CoreData) Marked(key string) bool {
	if cd.marks == nil {
		cd.marks = map[string]struct{}{}
	}
	_, marked := cd.marks[key]
	cd.marks[key] = struct{}{}
	return !marked
}

// CoreOption configures a new CoreData instance.
type CoreOption func(*CoreData)

// WithImageURLMapper sets the a4code image mapper option.
func WithImageURLMapper(fn func(tag, val string) string) CoreOption {
	return func(cd *CoreData) { cd.a4codeMapper = fn }
}

func (cd *CoreData) composeMapper() {
	var fns []func(tag, val string) string
	if cd.ImageSigner != nil {
		fns = append(fns, cd.ImageSigner.MapURL)
	}
	if cd.LinkSigner != nil {
		fns = append(fns, cd.LinkSigner.MapURL)
	}
	if len(fns) == 0 {
		cd.a4codeMapper = nil
		return
	}
	cd.a4codeMapper = func(tag, val string) string {
		for _, fn := range fns {
			newVal := fn(tag, val)
			if newVal != val {
				return newVal
			}
			val = newVal
		}
		return val
	}
}

// WithSession stores the gorilla session on the CoreData object.
func WithSession(s *sessions.Session) CoreOption {
	return func(cd *CoreData) { cd.session = s }
}

// WithSessionManager sets the session manager used by CoreData.
func WithSessionManager(sm SessionManager) CoreOption {
	return func(cd *CoreData) { cd.sessionManager = sm }
}

// WithEvent links an event to the CoreData object.
func WithEvent(evt *eventbus.TaskEvent) CoreOption { return func(cd *CoreData) { cd.event = evt } }

// WithAbsoluteURLBase sets the base URL used to build absolute links.
func WithAbsoluteURLBase(base string) CoreOption {
	return func(cd *CoreData) { cd.absoluteURLBase.Set(strings.TrimRight(base, "/")) }
}

// WithPreference preloads the user preference object.
func WithPreference(p *db.Preference) CoreOption {
	return func(cd *CoreData) { cd.pref.Set(p) }
}

// WithConfig sets the runtime config for this CoreData.
func WithConfig(cfg *config.RuntimeConfig) CoreOption {
	return func(cd *CoreData) { cd.Config = cfg }
}

// WithImageSigner registers the image signer and URL mapper on CoreData.
func WithImageSigner(s *imagesign.Signer) CoreOption {
	return func(cd *CoreData) {
		cd.ImageSigner = s
		cd.composeMapper()
	}
}

// WithLinkSigner registers the external link signer on CoreData.
func WithLinkSigner(s *linksign.Signer) CoreOption {
	return func(cd *CoreData) {
		cd.LinkSigner = s
		cd.composeMapper()
	}
}

// WithTasksRegistry registers the task registry on CoreData.
func WithTasksRegistry(r *tasks.Registry) CoreOption {
	return func(cd *CoreData) { cd.TasksReg = r }
}

// WithDBRegistry sets the database driver registry for CoreData.
func WithDBRegistry(r *dbdrivers.Registry) CoreOption {
	return func(cd *CoreData) { cd.dbRegistry = r }
}

// WithNavRegistry registers the navigation registry on CoreData.
func WithNavRegistry(r NavigationProvider) CoreOption {
	return func(cd *CoreData) { cd.Nav = r }
}

// WithCustomQueries sets the db.CustomQueries dependency.
func WithCustomQueries(cq db.CustomQueries) CoreOption {
	return func(cd *CoreData) { cd.customQueries = cq }
}

// WithSelectionsFromRequest extracts integer identifiers from the request and
// stores them on the CoreData instance. It searches path variables, query
// parameters and finally form values.
func WithSelectionsFromRequest(r *http.Request) CoreOption {
	return func(cd *CoreData) {
		setID := func(k, v string) {
			i, err := strconv.Atoi(v)
			if err != nil {
				return
			}
			switch k {
			case "boardno", "board":
				cd.currentBoardID = int32(i)
			case "thread", "replyTo":
				cd.currentThreadID = int32(i)
			case "topic":
				cd.currentTopicID = int32(i)
			case "comment":
				cd.currentCommentID = int32(i)
			case "news":
				cd.currentNewsPostID = int32(i)
			case "post":
				cd.currentImagePostID = int32(i)
			case "writing":
				cd.currentWritingID = int32(i)
			case "blog":
				cd.currentBlogID = int32(i)
			}
		}
		for k, v := range mux.Vars(r) {
			setID(k, v)
		}
		q := r.URL.Query()
		for k, v := range q {
			if len(v) > 0 {
				setID(k, v[0])
			}
		}
		if err := r.ParseForm(); err == nil {
			for k, v := range r.Form {
				if len(v) > 0 {
					setID(k, v[0])
				}
			}
		}
	}
}

// NewCoreData creates a CoreData with context and queries applied.
func NewCoreData(ctx context.Context, q db.Querier, cfg *config.RuntimeConfig, opts ...CoreOption) *CoreData {
	cd := &CoreData{
		ctx:               ctx,
		queries:           q,
		newsAnnouncements: map[int32]*lazy.Value[*db.SiteAnnouncement]{},
		Config:            cfg,
	}
	if cq, ok := q.(db.CustomQueries); ok {
		cd.customQueries = cq
	}
	for _, o := range opts {
		o(cd)
	}
	return cd
}

// Queries returns the db.Queries instance associated with this CoreData.
func (cd *CoreData) Queries() db.Querier { return cd.queries }

// CustomQueries returns the db.CustomQueries instance associated with this CoreData.
func (cd *CoreData) CustomQueries() db.CustomQueries { return cd.customQueries }

// SelectedBoard returns the selected board identifier.
func (cd *CoreData) SelectedBoard() int { return int(cd.currentBoardID) }

// SelectedThread returns the selected thread identifier.
func (cd *CoreData) SelectedThread() int { return int(cd.currentThreadID) }

// SelectedImagePost returns the selected image post identifier.
func (cd *CoreData) SelectedImagePost() int { return int(cd.currentImagePostID) }

// ImageURLMapper maps image references like "image:" or "cache:" to full URLs.
func (cd *CoreData) ImageURLMapper(tag, val string) string {
	if cd.a4codeMapper != nil {
		return cd.a4codeMapper(tag, val)
	}
	return val
}

// EmailProvider lazily returns the configured email provider.
// WithEmailProvider sets the email provider used by CoreData.
func WithEmailProvider(p MailProvider) CoreOption {
	return func(cd *CoreData) { cd.emailProvider.Set(p) }
}

// EmailProvider returns the configured email provider.
func (cd *CoreData) EmailProvider() MailProvider {
	p, err := cd.emailProvider.Load(func() (MailProvider, error) { return nil, nil })
	if err != nil {
		log.Printf("load email provider: %v", err)
	}
	return p
}

// HasRole reports whether the current user explicitly has the named role.
func (cd *CoreData) HasRole(role string) bool {
	for _, r := range cd.UserRoles() {
		if r == role {
			return true
		}
	}
	if cd.queries != nil {
		for _, r := range cd.UserRoles() {
			if _, err := cd.queries.SystemCheckRoleGrant(cd.ctx, db.SystemCheckRoleGrantParams{Name: r, Action: role}); err == nil {
				return true
			}
		}
	} else {
		for _, r := range cd.UserRoles() {
			switch r {
			case "administrator":
				if role == "moderator" || role == "content writer" || role == "user" {
					return true
				}
			case "moderator":
				if role == "user" {
					return true
				}
			case "content writer":
				if role == "user" {
					return true
				}
			}
		}
	}
	return false
}

// ContainsItem returns true if items includes an entry with the given name.
func ContainsItem(items []IndexItem, name string) bool {
	for _, it := range items {
		if it.Name == name {
			return true
		}
	}
	return false
}

// UserRoles returns the user roles loaded lazily.
func (cd *CoreData) UserRoles() []string {
	roles, err := cd.userRoles.Load(func() ([]string, error) {
		rs := []string{"anonymous"}
		if cd.UserID == 0 || cd.queries == nil {
			return rs, nil
		}
		rs = append(rs, "user")
		perms, err := cd.queries.GetPermissionsByUserID(cd.ctx, cd.UserID)
		if err != nil {
			return rs, nil
		}
		for _, p := range perms {
			if p.Name != "" {
				rs = append(rs, p.Name)
			}
		}
		return rs, nil
	})
	if err != nil {
		log.Printf("load user roles: %v", err)
	}
	return roles
}

// Role returns the first loaded role or "anonymous" when none.
func (cd *CoreData) Role() string {
	roles := cd.UserRoles()
	if len(roles) == 0 {
		return "anonymous"
	}
	return roles[0]
}

// SetSession stores s on cd for later retrieval.
func (cd *CoreData) SetSession(s *sessions.Session) { cd.session = s }

// Session returns the request session if available.
func (cd *CoreData) Session() *sessions.Session { return cd.session }

// SessionManager returns the configured session manager, if any.
func (cd *CoreData) SessionManager() SessionManager { return cd.sessionManager }

// DBRegistry returns the database driver registry associated with this request.
func (cd *CoreData) DBRegistry() *dbdrivers.Registry { return cd.dbRegistry }

// SetEvent stores evt on cd for handler access.
func (cd *CoreData) SetEvent(evt *eventbus.TaskEvent) { cd.event = evt }

// SetEventTask records the task associated with the current request event.
func (cd *CoreData) SetEventTask(t tasks.Task) {
	if cd.event != nil {
		cd.event.Task = t
	}
}

// SetPageTitle updates the Title field used by templates.
func (cd *CoreData) SetPageTitle(title string) {
	cd.Title = title
}

// AbsoluteURL returns an absolute URL by combining the configured hostname or
// the request host with path. The base value is cached per request.
func (cd *CoreData) AbsoluteURL(path string) string {
	base, err := cd.absoluteURLBase.Load(func() (string, error) { return "", nil })
	if err != nil {
		log.Printf("load absolute URL base: %v", err)
	}
	return base + path
}

// Event returns the event associated with the request, if any.
func (cd *CoreData) Event() *eventbus.TaskEvent { return cd.event }

// CurrentUser returns the logged in user's record loaded on demand.
func (cd *CoreData) CurrentUser() (*db.User, error) {
	return cd.user.Load(func() (*db.User, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.SystemGetUserByID(cd.ctx, cd.UserID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			return nil, nil
		}
		return &db.User{Idusers: row.Idusers, Username: row.Username}, nil
	})
}

// CurrentUserLoaded returns the cached current user without triggering a database lookup.
func (cd *CoreData) CurrentUserLoaded() *db.User {
	u, ok := cd.user.Peek()
	if !ok {
		return nil
	}
	return u
}

// SetCurrentThreadAndTopic stores the requested thread and topic IDs.
func (cd *CoreData) SetCurrentThreadAndTopic(threadID, topicID int32) {
	cd.currentThreadID = threadID
	cd.currentTopicID = topicID
}

// CurrentThread returns the currently requested thread lazily loaded.
func (cd *CoreData) CurrentThread(ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	if cd.currentThreadID == 0 {
		return nil, nil
	}
	return cd.ForumThreadByID(cd.currentThreadID, ops...)
}

// CurrentThreadLoaded returns the cached current thread without database access.
func (cd *CoreData) CurrentThreadLoaded() *db.GetThreadLastPosterAndPermsRow {
	if cd.forumThreadRows == nil {
		return nil
	}
	lv, ok := cd.forumThreadRows[cd.currentThreadID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// CurrentTopic returns the currently requested topic lazily loaded.
func (cd *CoreData) CurrentTopic(ops ...lazy.Option[*db.GetForumTopicByIdForUserRow]) (*db.GetForumTopicByIdForUserRow, error) {
	if cd.currentTopicID == 0 {
		return nil, nil
	}
	return cd.ForumTopicByID(cd.currentTopicID, ops...)
}

// CurrentTopicLoaded returns the cached current topic without database access.
func (cd *CoreData) CurrentTopicLoaded() *db.GetForumTopicByIdForUserRow {
	if cd.forumTopics == nil {
		return nil
	}
	lv, ok := cd.forumTopics[cd.currentTopicID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// SetCurrentWriting stores the requested writing ID.
func (cd *CoreData) SetCurrentWriting(id int32) { cd.currentWritingID = id }

// CurrentWriting returns the currently requested writing lazily loaded.
func (cd *CoreData) CurrentWriting(ops ...lazy.Option[*db.GetWritingForListerByIDRow]) (*db.GetWritingForListerByIDRow, error) {
	if cd.currentWritingID == 0 {
		return nil, nil
	}
	return cd.WritingByID(cd.currentWritingID, ops...)
}

// CurrentWritingLoaded returns the cached current writing without database access.
func (cd *CoreData) CurrentWritingLoaded() *db.GetWritingForListerByIDRow {
	if cd.writingRows == nil {
		return nil
	}
	lv, ok := cd.writingRows[cd.currentWritingID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// SetCurrentBlog stores the requested blog entry ID.
func (cd *CoreData) SetCurrentBlog(id int32) { cd.currentBlogID = id }

// CurrentBlog returns the currently requested blog entry lazily loaded.
func (cd *CoreData) CurrentBlog(ops ...lazy.Option[*db.GetBlogEntryForListerByIDRow]) (*db.GetBlogEntryForListerByIDRow, error) {
	if cd.currentBlogID == 0 {
		return nil, nil
	}
	return cd.BlogEntryByID(cd.currentBlogID, ops...)
}

// CurrentBlogLoaded returns the cached current blog entry without database access.
func (cd *CoreData) CurrentBlogLoaded() *db.GetBlogEntryForListerByIDRow {
	if cd.blogEntries == nil {
		return nil
	}
	lv, ok := cd.blogEntries[cd.currentBlogID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// Permissions returns the user's permissions loaded on demand.
func (cd *CoreData) Permissions() ([]*db.GetPermissionsByUserIDRow, error) {
	return cd.perms.Load(func() ([]*db.GetPermissionsByUserIDRow, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPermissionsByUserID(cd.ctx, cd.UserID)
	})
}

// Preference returns the user's preferences loaded on demand.
func (cd *CoreData) Preference() (*db.Preference, error) {
	return cd.pref.Load(func() (*db.Preference, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetPreferenceForLister(cd.ctx, cd.UserID)
	})
}

// PageSize returns the preferred page size within configured limits.
func (cd *CoreData) PageSize() int {
	size := cd.Config.PageSizeDefault
	if pref, err := cd.Preference(); err == nil && pref != nil && pref.PageSize != 0 {
		size = int(pref.PageSize)
	}
	if size < cd.Config.PageSizeMin {
		size = cd.Config.PageSizeMin
	}
	if size > cd.Config.PageSizeMax {
		size = cd.Config.PageSizeMax
	}
	return size
}

// Languages returns the list of available languages loaded on demand.
func (cd *CoreData) Languages() ([]*db.Language, error) {
	return cd.langs.Load(func() ([]*db.Language, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.SystemListLanguages(cd.ctx)
	})
}

// AllLanguages returns all languages cached once.
func (cd *CoreData) AllLanguages() ([]*db.Language, error) {
	return cd.languagesAll.Load(func() ([]*db.Language, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.SystemListLanguages(cd.ctx)
	})
}

// PreferredLanguageID returns the user's preferred language ID if set,
// otherwise it resolves the site's default language name to an ID.
func (cd *CoreData) PreferredLanguageID(siteDefault string) int32 {
	id, err := cd.preferredLanguageID.Load(func() (int32, error) {
		if pref, err := cd.Preference(); err == nil && pref != nil {
			if pref.LanguageIdlanguage != 0 {
				return pref.LanguageIdlanguage, nil
			}
		}
		if cd.queries == nil || siteDefault == "" {
			return 0, nil
		}
		langID, err := cd.queries.SystemGetLanguageIDByName(cd.ctx, sql.NullString{String: siteDefault, Valid: true})
		if err != nil {
			return 0, nil
		}
		return langID, nil
	})
	if err != nil {
		log.Printf("load preferred language id: %v", err)
	}
	return id
}

// AllRoles returns every defined role loaded once from the database.
func (cd *CoreData) AllRoles() ([]*db.Role, error) {
	return cd.allRoles.Load(func() ([]*db.Role, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.AdminListRoles(cd.ctx)
	})
}

// Announcement returns the active announcement row loaded lazily.
func (cd *CoreData) Announcement() *db.GetActiveAnnouncementWithNewsForListerRow {
	ann, err := cd.announcement.Load(func() (*db.GetActiveAnnouncementWithNewsForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetActiveAnnouncementWithNewsForLister(cd.ctx, db.GetActiveAnnouncementWithNewsForListerParams{
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			return nil, err
		}
		return row, nil
	})
	if err != nil {
		log.Printf("load announcement: %v", err)
	}
	return ann
}

// AnnouncementLoaded returns the cached active announcement without querying the database.
func (cd *CoreData) AnnouncementLoaded() *db.GetActiveAnnouncementWithNewsForListerRow {
	ann, ok := cd.announcement.Peek()
	if !ok {
		return nil
	}
	return ann
}

// AnnouncementForNews fetches the latest announcement for the given news post
// only once.
func (cd *CoreData) AnnouncementForNews(id int32) (*db.SiteAnnouncement, error) {
	if cd.newsAnnouncements == nil {
		cd.newsAnnouncements = map[int32]*lazy.Value[*db.SiteAnnouncement]{}
	}
	lv, ok := cd.newsAnnouncements[id]
	if !ok {
		lv = &lazy.Value[*db.SiteAnnouncement]{}
		cd.newsAnnouncements[id] = lv
	}
	return lv.Load(func() (*db.SiteAnnouncement, error) {
		if cd.queries == nil {
			return nil, nil
		}
		ann, err := cd.queries.GetLatestAnnouncementByNewsID(cd.ctx, id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return ann, err
	})
}

// NewsAnnouncement returns the latest announcement for the given news post. The
// result is cached so repeated lookups for the same id hit the database only
// once.
func (cd *CoreData) NewsAnnouncement(id int32) (*db.SiteAnnouncement, error) {
	cd.annMu.Lock()
	lv, ok := cd.newsAnnouncements[id]
	if !ok {
		lv = &lazy.Value[*db.SiteAnnouncement]{}
		cd.newsAnnouncements[id] = lv
	}
	cd.annMu.Unlock()

	return lv.Load(func() (*db.SiteAnnouncement, error) {
		if cd.queries == nil {
			return nil, nil
		}
		ann, err := cd.queries.GetLatestAnnouncementByNewsID(cd.ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return ann, nil
	})
}

// ForumCategories loads all forum categories once.
func (cd *CoreData) ForumCategories() ([]*db.Forumcategory, error) {
	return cd.forumCategories.Load(func() ([]*db.Forumcategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetAllForumCategories(cd.ctx)
	})
}

// ForumTopics loads forum topics for a given category once per category.
func (cd *CoreData) ForumTopics(categoryID int32) ([]*db.Forumtopic, error) {
	if cd.forumTopicLists == nil {
		cd.forumTopicLists = make(map[int32]*lazy.Value[[]*db.Forumtopic])
	}
	lv, ok := cd.forumTopicLists[categoryID]
	if !ok {
		lv = &lazy.Value[[]*db.Forumtopic]{}
		cd.forumTopicLists[categoryID] = lv
	}
	return lv.Load(func() ([]*db.Forumtopic, error) {
		if cd.queries == nil {
			return nil, nil
		}
		if categoryID == 0 {
			return cd.queries.GetAllForumTopics(cd.ctx)
		}
		return cd.queries.GetForumTopicsByCategoryId(cd.ctx, categoryID)
	})
}

// AdminForumTopics returns all forum topics without category filtering.
func (cd *CoreData) AdminForumTopics() ([]*db.Forumtopic, error) {
	return cd.ForumTopics(0)
}

// ForumThreads loads the threads for a forum topic once per topic.
func (cd *CoreData) ForumThreads(topicID int32) ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
	if cd.forumThreads == nil {
		cd.forumThreads = make(map[int32]*lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow])
	}
	lv, ok := cd.forumThreads[topicID]
	if !ok {
		lv = &lazy.Value[[]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow]{}
		cd.forumThreads[topicID] = lv
	}
	return lv.Load(func() ([]*db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostText(cd.ctx, db.GetForumThreadsByForumTopicIdForUserWithFirstAndLastPosterAndFirstPostTextParams{
			ViewerID:      cd.UserID,
			TopicID:       topicID,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	})
}

// LatestNews returns recent news posts with permission data.
func (cd *CoreData) LatestNews(r *http.Request) ([]*NewsPost, error) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	replyID, _ := strconv.Atoi(r.URL.Query().Get("reply"))
	return cd.latestNews.Load(func() ([]*NewsPost, error) {
		return cd.fetchLatestNews(int32(offset), 15, replyID)
	})
}

// LatestNewsList returns recent news posts without needing an HTTP request.
func (cd *CoreData) LatestNewsList(offset, limit int32) ([]*NewsPost, error) {
	return cd.fetchLatestNews(offset, limit, 0)
}

// fetchLatestNews loads news posts from the database with permission data.
func (cd *CoreData) fetchLatestNews(offset, limit int32, replyID int) ([]*NewsPost, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescending(cd.ctx, db.GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingParams{
		ViewerID: cd.UserID,
		UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var posts []*NewsPost
	for _, row := range rows {
		if !cd.HasGrant("news", "post", "see", row.Idsitenews) {
			continue
		}
		ann, err := cd.queries.GetLatestAnnouncementByNewsID(cd.ctx, row.Idsitenews)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if !cd.HasGrant("news", "post", "see", row.Idsitenews) {
			continue
		}
		posts = append(posts, &NewsPost{
			GetNewsPostsWithWriterUsernameAndThreadCommentCountDescendingRow: row,
			ShowReply:    cd.UserID != 0,
			ShowEdit:     cd.HasGrant("news", "post", "edit", row.Idsitenews) && (cd.AdminMode || cd.UserID != 0),
			Editing:      replyID == int(row.Idsitenews),
			Announcement: ann,
		})
	}
	return posts, nil
}

// LatestWritings returns recent public writings with permission data.
type LatestWritingsOption func(*db.GetPublicWritingsParams)

// WithWritingsOffset sets the query offset.
func WithWritingsOffset(o int32) LatestWritingsOption {
	return func(p *db.GetPublicWritingsParams) { p.Offset = o }
}

// WithWritingsLimit sets the query limit.
func WithWritingsLimit(l int32) LatestWritingsOption {
	return func(p *db.GetPublicWritingsParams) { p.Limit = l }
}

func (cd *CoreData) LatestWritings(opts ...LatestWritingsOption) ([]*db.Writing, error) {
	return cd.latestWritings.Load(func() ([]*db.Writing, error) {
		if cd.queries == nil {
			return nil, nil
		}
		params := db.GetPublicWritingsParams{Limit: 15}
		for _, o := range opts {
			o(&params)
		}
		rows, err := cd.queries.GetPublicWritings(cd.ctx, params)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		var writings []*db.Writing
		for _, row := range rows {
			if !cd.HasGrant("writing", "article", "see", row.Idwriting) {
				continue
			}
			writings = append(writings, row)
		}
		return writings, nil
	})
}

// WritingCategories returns the visible writing categories for userID.
func (cd *CoreData) VisibleWritingCategories(userID int32) ([]*db.WritingCategory, error) {
	return cd.visibleWritingCategories.Load(func() ([]*db.WritingCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.ListWritingCategoriesForLister(cd.ctx, db.ListWritingCategoriesForListerParams{
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: userID, Valid: userID != 0},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var cats []*db.WritingCategory
		for _, row := range rows {
			if cd.HasGrant("writing", "category", "see", row.Idwritingcategory) {
				cats = append(cats, row)
			}
		}
		return cats, nil
	})
}

// CurrentUserVisibleWritingCategories returns writing categories visible to the current user.
func (cd *CoreData) CurrentUserVisibleWritingCategories() ([]*db.WritingCategory, error) {
	return cd.VisibleWritingCategories(cd.UserID)
}

// WritingCategories returns all writing categories cached once.
func (cd *CoreData) WritingCategories() ([]*db.WritingCategory, error) {
	return cd.writingCategories.Load(func() ([]*db.WritingCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.SystemListWritingCategories(cd.ctx, db.SystemListWritingCategoriesParams{Limit: math.MaxInt32, Offset: 0})
	})
}

// PublicWritings returns public writings in a category, cached per category and offset.
func (cd *CoreData) PublicWritings(categoryID int32, r *http.Request) ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
	if cd.publicWritings == nil {
		cd.publicWritings = map[string]*lazy.Value[[]*db.ListPublicWritingsInCategoryForListerRow]{}
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	key := fmt.Sprintf("%d:%d", categoryID, offset)
	lv, ok := cd.publicWritings[key]
	if !ok {
		lv = &lazy.Value[[]*db.ListPublicWritingsInCategoryForListerRow]{}
		cd.publicWritings[key] = lv
	}
	return lv.Load(func() ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.ListPublicWritingsInCategoryForLister(cd.ctx, db.ListPublicWritingsInCategoryForListerParams{
			ListerID:          cd.UserID,
			WritingCategoryID: categoryID,
			UserID:            sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:             15,
			Offset:            int32(offset),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var res []*db.ListPublicWritingsInCategoryForListerRow
		for _, row := range rows {
			if cd.HasGrant("writing", "article", "see", row.Idwriting) {
				res = append(res, row)
			}
		}
		return res, nil
	})
}

// SelectedCategoryPublicWritings returns public writings for the given category.
func (cd *CoreData) SelectedCategoryPublicWritings(categoryID int32, r *http.Request) ([]*db.ListPublicWritingsInCategoryForListerRow, error) {
	return cd.PublicWritings(categoryID, r)
}

// Bloggers returns bloggers ordered by username with post counts.
func (cd *CoreData) Bloggers(r *http.Request) ([]*db.ListBloggersForListerRow, error) {
	return cd.bloggers.Load(func() ([]*db.ListBloggersForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		ps := cd.PageSize()
		search := r.URL.Query().Get("search")
		if search != "" {
			like := "%" + search + "%"
			rows, err := cd.queries.ListBloggersSearchForLister(cd.ctx, db.ListBloggersSearchForListerParams{
				ListerID: cd.UserID,
				Query:    like,
				UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
				Limit:    int32(ps + 1),
				Offset:   int32(offset),
			})
			if err != nil {
				return nil, err
			}
			items := make([]*db.ListBloggersForListerRow, 0, len(rows))
			for _, r := range rows {
				items = append(items, &db.ListBloggersForListerRow{Username: r.Username, Count: r.Count})
			}
			return items, nil
		}
		return cd.queries.ListBloggersForLister(cd.ctx, db.ListBloggersForListerParams{
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    int32(ps + 1),
			Offset:   int32(offset),
		})
	})
}

// Writers returns writers ordered by username with article counts.
func (cd *CoreData) Writers(r *http.Request) ([]*db.ListWritersForListerRow, error) {
	return cd.writers.Load(func() ([]*db.ListWritersForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		ps := cd.PageSize()
		search := r.URL.Query().Get("search")
		if search != "" {
			like := "%" + search + "%"
			rows, err := cd.queries.ListWritersSearchForLister(cd.ctx, db.ListWritersSearchForListerParams{
				ListerID: cd.UserID,
				Query:    like,
				UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
				Limit:    int32(ps + 1),
				Offset:   int32(offset),
			})
			if err != nil {
				return nil, err
			}
			items := make([]*db.ListWritersForListerRow, 0, len(rows))
			for _, r := range rows {
				items = append(items, &db.ListWritersForListerRow{Username: r.Username, Count: r.Count})
			}
			return items, nil
		}
		return cd.queries.ListWritersForLister(cd.ctx, db.ListWritersForListerParams{
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    int32(ps + 1),
			Offset:   int32(offset),
		})
	})
}

// ForumTopicByID loads a forum topic once per ID using caching.
func (cd *CoreData) ForumTopicByID(id int32, ops ...lazy.Option[*db.GetForumTopicByIdForUserRow]) (*db.GetForumTopicByIdForUserRow, error) {
	fetch := func(i int32) (*db.GetForumTopicByIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumTopicByIdForUser(cd.ctx, db.GetForumTopicByIdForUserParams{
			ViewerID:      cd.UserID,
			Idforumtopic:  i,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumTopics, &cd.mapMu, id, fetch, ops...)
}

// ForumThreadByID returns a single forum thread lazily loading it once per ID.
func (cd *CoreData) ForumThreadByID(id int32, ops ...lazy.Option[*db.GetThreadLastPosterAndPermsRow]) (*db.GetThreadLastPosterAndPermsRow, error) {
	fetch := func(i int32) (*db.GetThreadLastPosterAndPermsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetThreadLastPosterAndPerms(cd.ctx, db.GetThreadLastPosterAndPermsParams{
			ViewerID:      cd.UserID,
			ThreadID:      i,
			ViewerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumThreadRows, &cd.mapMu, id, fetch, ops...)
}

// ThreadComments returns comments for the thread lazily loading once per thread ID.
func (cd *CoreData) ThreadComments(id int32, ops ...lazy.Option[[]*db.GetCommentsByThreadIdForUserRow]) ([]*db.GetCommentsByThreadIdForUserRow, error) {
	fetch := func(i int32) ([]*db.GetCommentsByThreadIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetCommentsByThreadIdForUser(cd.ctx, db.GetCommentsByThreadIdForUserParams{
			ViewerID: cd.UserID,
			ThreadID: i,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumThreadComments, &cd.mapMu, id, fetch, ops...)
}

// WritingByID returns a single writing lazily loading it once per ID.
func (cd *CoreData) WritingByID(id int32, ops ...lazy.Option[*db.GetWritingForListerByIDRow]) (*db.GetWritingForListerByIDRow, error) {
	fetch := func(i int32) (*db.GetWritingForListerByIDRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetWritingForListerByID(cd.ctx, db.GetWritingForListerByIDParams{
			ListerID:      cd.UserID,
			Idwriting:     i,
			ListerMatchID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.writingRows, &cd.mapMu, id, fetch, ops...)
}

// BlogEntryByID returns a blog entry lazily loading it once per ID.
func (cd *CoreData) BlogEntryByID(id int32, ops ...lazy.Option[*db.GetBlogEntryForListerByIDRow]) (*db.GetBlogEntryForListerByIDRow, error) {
	fetch := func(i int32) (*db.GetBlogEntryForListerByIDRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetBlogEntryForListerByID(cd.ctx, db.GetBlogEntryForListerByIDParams{
			ListerID: cd.UserID,
			ID:       i,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.blogEntries, &cd.mapMu, id, fetch, ops...)
}

// SetBlogListParams stores parameters for listing blogs.
func (cd *CoreData) SetBlogListParams(uid int32, offset int) {
	cd.blogListUID = uid
	cd.blogListOffset = offset
}

// BlogList returns blog entries for the current parameters.
func (cd *CoreData) BlogList() ([]*db.ListBlogEntriesByAuthorForListerRow, error) {
	return cd.blogListRows.Load(func() ([]*db.ListBlogEntriesByAuthorForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.ListBlogEntriesByAuthorForLister(cd.ctx, db.ListBlogEntriesByAuthorForListerParams{
			AuthorID: cd.blogListUID,
			ListerID: cd.UserID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    15,
			Offset:   int32(cd.blogListOffset),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		var list []*db.ListBlogEntriesByAuthorForListerRow
		for _, row := range rows {
			if !cd.HasGrant("blogs", "entry", "see", row.Idblogs) {
				continue
			}
			list = append(list, row)
		}
		return list, nil
	})
}

// BlogListUID returns the user ID parameter for the blog list.
func (cd *CoreData) BlogListUID() int32 { return cd.blogListUID }

// BlogListOffset returns the offset parameter for the blog list.
func (cd *CoreData) BlogListOffset() int { return cd.blogListOffset }

// CommentByID returns a forum comment lazily loading it once per ID.
func (cd *CoreData) CommentByID(id int32, ops ...lazy.Option[*db.GetCommentByIdForUserRow]) (*db.GetCommentByIdForUserRow, error) {
	fetch := func(i int32) (*db.GetCommentByIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetCommentByIdForUser(cd.ctx, db.GetCommentByIdForUserParams{
			ViewerID: cd.UserID,
			ID:       i,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.forumComments, &cd.mapMu, id, fetch, ops...)
}

// CurrentComment returns the current comment lazily loaded.
func (cd *CoreData) CurrentComment(r *http.Request, ops ...lazy.Option[*db.GetCommentByIdForUserRow]) (*db.GetCommentByIdForUserRow, error) {
	if cd.currentCommentID == 0 {
		if r != nil {
			idStr := r.URL.Query().Get("comment")
			if idStr == "" {
				if vars := mux.Vars(r); vars != nil {
					idStr = vars["comment"]
				}
			}
			if idStr != "" {
				id, err := strconv.Atoi(idStr)
				if err != nil {
					return nil, fmt.Errorf("invalid comment id: %w", err)
				}
				cd.currentCommentID = int32(id)
			}
		}
		if cd.currentCommentID == 0 {
			return nil, nil
		}
	}
	return cd.CommentByID(cd.currentCommentID, ops...)
}

// CurrentCommentLoaded returns the cached current comment if available.
func (cd *CoreData) CurrentCommentLoaded() *db.GetCommentByIdForUserRow {
	if cd.forumComments == nil {
		return nil
	}
	lv, ok := cd.forumComments[cd.currentCommentID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// ThreadComments returns comments for a thread lazily loading them once per thread ID.
func (cd *CoreData) ThreadComments(threadID int32) ([]*db.GetCommentsByThreadIdForUserRow, error) {
	if cd.threadComments == nil {
		cd.threadComments = make(map[int32]*lazy.Value[[]*db.GetCommentsByThreadIdForUserRow])
	}
	lv, ok := cd.threadComments[threadID]
	if !ok {
		lv = &lazy.Value[[]*db.GetCommentsByThreadIdForUserRow]{}
		cd.threadComments[threadID] = lv
	}
	return lv.Load(func() ([]*db.GetCommentsByThreadIdForUserRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetCommentsByThreadIdForUser(cd.ctx, db.GetCommentsByThreadIdForUserParams{
			ViewerID: cd.UserID,
			ThreadID: threadID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return rows, nil
	})
}

// NewsPostByID returns the news post lazily loading it once per ID.
func (cd *CoreData) NewsPostByID(id int32, ops ...lazy.Option[*db.GetForumThreadIdByNewsPostIdRow]) (*db.GetForumThreadIdByNewsPostIdRow, error) {
	fetch := func(i int32) (*db.GetForumThreadIdByNewsPostIdRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetForumThreadIdByNewsPostId(cd.ctx, i)
	}
	return lazy.Map(&cd.newsPosts, &cd.mapMu, id, fetch, ops...)
}

// SetCurrentNewsPost stores the current news post ID.
func (cd *CoreData) SetCurrentNewsPost(id int32) { cd.currentNewsPostID = id }

// CurrentNewsPost returns the current news post lazily loaded.
func (cd *CoreData) CurrentNewsPost(ops ...lazy.Option[*db.GetForumThreadIdByNewsPostIdRow]) (*db.GetForumThreadIdByNewsPostIdRow, error) {
	if cd.currentNewsPostID == 0 {
		return nil, nil
	}
	return cd.NewsPostByID(cd.currentNewsPostID, ops...)
}

// CurrentNewsPostLoaded returns the cached current news post if available.
func (cd *CoreData) CurrentNewsPostLoaded() *db.GetForumThreadIdByNewsPostIdRow {
	if cd.newsPosts == nil {
		return nil
	}
	lv, ok := cd.newsPosts[cd.currentNewsPostID]
	if !ok {
		return nil
	}
	v, ok := lv.Peek()
	if !ok {
		return nil
	}
	return v
}

// WriterWritings returns public writings for the specified author respecting cd's permissions.
func (cd *CoreData) WriterWritings(userID int32, r *http.Request) ([]*db.ListPublicWritingsByUserForListerRow, error) {
	if cd.writerWritings == nil {
		cd.writerWritings = map[int32]*lazy.Value[[]*db.ListPublicWritingsByUserForListerRow]{}
	}
	lv, ok := cd.writerWritings[userID]
	if !ok {
		lv = &lazy.Value[[]*db.ListPublicWritingsByUserForListerRow]{}
		cd.writerWritings[userID] = lv
	}
	return lv.Load(func() ([]*db.ListPublicWritingsByUserForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		rows, err := cd.queries.ListPublicWritingsByUserForLister(cd.ctx, db.ListPublicWritingsByUserForListerParams{
			ListerID: cd.UserID,
			AuthorID: userID,
			UserID:   sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:    15,
			Offset:   int32(offset),
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		var list []*db.ListPublicWritingsByUserForListerRow
		for _, row := range rows {
			if !cd.HasGrant("writing", "article", "see", row.Idwriting) {
				continue
			}
			list = append(list, row)
		}
		return list, nil
	})
}

// Bookmarks returns the user's bookmark list loaded lazily.
func (cd *CoreData) Bookmarks() (*db.GetBookmarksForUserRow, error) {
	return cd.bookmarks.Load(func() (*db.GetBookmarksForUserRow, error) {
		if cd.UserID == 0 || cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetBookmarksForUser(cd.ctx, cd.UserID)
	})
}

// CanEditAny reports whether cd is in admin mode with administrator role.
func (cd *CoreData) CanEditAny() bool {
	return cd.HasRole("administrator") && cd.AdminMode
}

// ImageBoards retrieves sub-boards under parentID lazily.
func (cd *CoreData) SubImageBoards(parentID int32) ([]*db.Imageboard, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.subImageBoards == nil {
		cd.subImageBoards = make(map[int32]*lazy.Value[[]*db.Imageboard])
	}
	lv, ok := cd.subImageBoards[parentID]
	if !ok {
		lv = &lazy.Value[[]*db.Imageboard]{}
		cd.subImageBoards[parentID] = lv
	}
	return lv.Load(func() ([]*db.Imageboard, error) {
		return cd.queries.ListBoardsByParentIDForLister(cd.ctx, db.ListBoardsByParentIDForListerParams{
			ListerID:     cd.UserID,
			ParentID:     parentID,
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:        200,
			Offset:       0,
		})
	})
}

// ImageBoards returns all image boards cached once.
func (cd *CoreData) ImageBoards() ([]*db.Imageboard, error) {
	return cd.imageBoards.Load(func() ([]*db.Imageboard, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.AdminListBoards(cd.ctx, db.AdminListBoardsParams{Limit: 200, Offset: 0})
	})
}

// ImageBoardPosts retrieves approved posts for the board lazily.
func (cd *CoreData) ImageBoardPosts(boardID int32) ([]*db.ListImagePostsByBoardForListerRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	if cd.imageBoardPosts == nil {
		cd.imageBoardPosts = make(map[int32]*lazy.Value[[]*db.ListImagePostsByBoardForListerRow])
	}
	lv, ok := cd.imageBoardPosts[boardID]
	if !ok {
		lv = &lazy.Value[[]*db.ListImagePostsByBoardForListerRow]{}
		cd.imageBoardPosts[boardID] = lv
	}
	return lv.Load(func() ([]*db.ListImagePostsByBoardForListerRow, error) {
		return cd.queries.ListImagePostsByBoardForLister(cd.ctx, db.ListImagePostsByBoardForListerParams{
			ListerID:     cd.UserID,
			BoardID:      boardID,
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
			Limit:        200,
			Offset:       0,
		})
	})
}

// ImagePostByID returns an image post once per ID using caching.
func (cd *CoreData) ImagePostByID(id int32, ops ...lazy.Option[*db.GetImagePostByIDForListerRow]) (*db.GetImagePostByIDForListerRow, error) {
	fetch := func(i int32) (*db.GetImagePostByIDForListerRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		return cd.queries.GetImagePostByIDForLister(cd.ctx, db.GetImagePostByIDForListerParams{
			ListerID:     cd.UserID,
			ID:           i,
			ListerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
	}
	return lazy.Map(&cd.imagePostRows, &cd.mapMu, id, fetch, ops...)
}

// UnreadNotificationCount returns the number of unread notifications for the
// current user. The value is fetched lazily on the first call and cached for
// subsequent calls.
func (cd *CoreData) UnreadNotificationCount() int64 {
	count, err := cd.unreadCount.Load(func() (int64, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return 0, nil
		}
		return cd.queries.GetUnreadNotificationCountForLister(cd.ctx, cd.UserID)
	})
	if err != nil {
		log.Printf("load unread notification count: %v", err)
	}
	return count
}

// Notifications returns the notifications for the current user using query
// parameters to control pagination. Results are cached per offset and filter
// combination.
func (cd *CoreData) Notifications(r *http.Request) ([]*db.Notification, error) {
	if cd.notifications == nil {
		cd.notifications = map[string]*lazy.Value[[]*db.Notification]{}
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	showAll := r.URL.Query().Get("all") == "1"
	key := fmt.Sprintf("%t:%d", showAll, offset)
	lv, ok := cd.notifications[key]
	if !ok {
		lv = &lazy.Value[[]*db.Notification]{}
		cd.notifications[key] = lv
	}
	return lv.Load(func() ([]*db.Notification, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return nil, nil
		}
		limit := int32(cd.Config.PageSizeDefault)
		if showAll {
			return cd.queries.ListNotificationsForLister(cd.ctx, db.ListNotificationsForListerParams{
				ListerID: cd.UserID,
				Limit:    limit,
				Offset:   int32(offset),
			})
		}
		return cd.queries.ListUnreadNotificationsForLister(cd.ctx, db.ListUnreadNotificationsForListerParams{
			ListerID: cd.UserID,
			Limit:    limit,
			Offset:   int32(offset),
		})
	})
}

// subscriptionMap loads the current user's subscriptions once.
func (cd *CoreData) subscriptionMap() (map[string]bool, error) {
	return cd.subscriptions.Load(func() (map[string]bool, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return map[string]bool{}, nil
		}
		rows, err := cd.UserSubscriptions()
		if err != nil {
			return nil, err
		}
		m := make(map[string]bool)
		for _, row := range rows {
			key := row.Pattern + "|" + row.Method
			m[key] = true
			if row.Method == "internal" {
				m[row.Pattern] = true
			}
		}
		return m, nil
	})
}

// Subscribed reports whether the user has a subscription matching pattern and method.
func (cd *CoreData) Subscribed(pattern, method string) bool {
	m, _ := cd.subscriptionMap()
	return m[pattern+"|"+method]
}

// UserSubscriptions returns the current user's subscriptions loaded lazily.
func (cd *CoreData) UserSubscriptions() ([]*db.ListSubscriptionsByUserRow, error) {
	return cd.userSubscriptions.Load(func() ([]*db.ListSubscriptionsByUserRow, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return nil, nil
		}
		return cd.queries.ListSubscriptionsByUser(cd.ctx, cd.UserID)
	})
}

// HasSubscription reports whether the user has subscribed to pattern with method.
func (cd *CoreData) HasSubscription(pattern, method string) bool {
	m, _ := cd.subscriptionMap()
	return m[pattern+"|"+method]
}

// Subscriptions returns the current user's subscriptions.
func (cd *CoreData) Subscriptions() ([]*db.ListSubscriptionsByUserRow, error) {
	return cd.subscriptionRows.Load(func() ([]*db.ListSubscriptionsByUserRow, error) {
		if cd.queries == nil || cd.UserID == 0 {
			return nil, nil
		}
		return cd.queries.ListSubscriptionsByUser(cd.ctx, cd.UserID)
	})
}

// AdminLinkerItemByID returns a single linker item lazily loading it once per ID.
func (cd *CoreData) AdminLinkerItemByID(id int32, ops ...lazy.Option[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow]) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, error) {
	fetch := func(i int32) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		row, err := cd.queries.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescending(cd.ctx, i)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return row, nil
	}
	return lazy.Map(&cd.adminLinkerItemRows, &cd.mapMu, id, fetch, ops...)
}

// LinkerCategories returns all linker categories.
func (cd *CoreData) LinkerCategories() ([]*db.LinkerCategory, error) {
	return cd.linkerCatsAll.Load(func() ([]*db.LinkerCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetAllLinkerCategories(cd.ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	})
}

// LinkerCategoriesForUser returns linker categories the viewer can access.
func (cd *CoreData) LinkerCategoriesForUser() ([]*db.LinkerCategory, error) {
	return cd.linkerCatsForUser.Load(func() ([]*db.LinkerCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetAllLinkerCategoriesForUser(cd.ctx, db.GetAllLinkerCategoriesForUserParams{
			ViewerID:     cd.UserID,
			ViewerUserID: sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	})
}

// LinkerCategoryByID returns a linker category lazily loading it once per ID.
func (cd *CoreData) LinkerCategoryByID(id int32, ops ...lazy.Option[*db.LinkerCategory]) (*db.LinkerCategory, error) {
	fetch := func(i int32) (*db.LinkerCategory, error) {
		if cd.queries == nil {
			return nil, nil
		}
		cat, err := cd.queries.GetLinkerCategoryById(cd.ctx, i)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return cat, nil
	}
	return lazy.Map(&cd.linkerCategoryRows, &cd.mapMu, id, fetch, ops...)
}

// LinkerCategoryCounts lazily loads linker category statistics.
func (cd *CoreData) LinkerCategoryCounts() ([]*db.GetLinkerCategoryLinkCountsRow, error) {
	return cd.linkerCategories.Load(func() ([]*db.GetLinkerCategoryLinkCountsRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetLinkerCategoryLinkCounts(cd.ctx)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	})
}

// LinkerItemsForUser returns linker items for the given category and offset respecting viewer permissions.
func (cd *CoreData) LinkerItemsForUser(catID, offset int32) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow, error) {
	if cd.queries == nil {
		return nil, nil
	}
	rows, err := cd.queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginated(cd.ctx, db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedParams{
		ViewerID:         cd.UserID,
		Idlinkercategory: catID,
		ViewerUserID:     sql.NullInt32{Int32: cd.UserID, Valid: cd.UserID != 0},
		Limit:            15,
		Offset:           offset,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	var out []*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow
	for _, row := range rows {
		if cd.HasGrant("linker", "link", "see", row.Idlinker) {
			out = append(out, row)
		}
	}
	return out, nil
}

// LinkerLinksByCategoryID returns the links for a category lazily loading them once per ID.
func (cd *CoreData) LinkerLinksByCategoryID(id int32, ops ...lazy.Option[[]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow]) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow, error) {
	fetch := func(i int32) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingRow, error) {
		if cd.queries == nil {
			return nil, nil
		}
		rows, err := cd.queries.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescending(cd.ctx, db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingParams{Idlinkercategory: i})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return rows, nil
	}
	return lazy.Map(&cd.linkerCategoryLinks, &cd.mapMu, id, fetch, ops...)
}

// SelectedAdminLinkerItem returns the linker item for the ID found in the request.
func (cd *CoreData) SelectedAdminLinkerItem(r *http.Request, ops ...lazy.Option[*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow]) (*db.GetLinkerItemByIdWithPosterUsernameAndCategoryTitleDescendingRow, int32, error) {
	id, err := cd.SelectedAdminLinkerItemID(r)
	if err != nil {
		return nil, 0, err
	}
	link, err := cd.AdminLinkerItemByID(id, ops...)
	if err != nil {
		return nil, id, err
	}
	return link, id, nil
}

// SelectedAdminLinkerItemID extracts the linker item ID from URL vars, form values or query parameters.
func (cd *CoreData) SelectedAdminLinkerItemID(r *http.Request) (int32, error) {
	var idStr string
	if v, ok := mux.Vars(r)["link"]; ok {
		idStr = v
	} else if v := r.PostFormValue("link"); v != "" {
		idStr = v
	} else {
		idStr = r.URL.Query().Get("link")
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id == 0 {
		return 0, sql.ErrNoRows
	}
	return int32(id), nil
}

// SelectedLinkerCategory returns the linker category for the given ID.
func (cd *CoreData) SelectedLinkerCategory(id int32, ops ...lazy.Option[*db.LinkerCategory]) (*db.LinkerCategory, error) {
	return cd.LinkerCategoryByID(id, ops...)
}

// SelectedLinkerItemsForCurrentUser returns linker items for the given category
// and offset for the current user and ensures the category is cached.
func (cd *CoreData) SelectedLinkerItemsForCurrentUser(catID, offset int32) ([]*db.GetAllLinkerItemsByCategoryIdWitherPosterUsernameAndCategoryTitleDescendingForUserPaginatedRow, error) {
	if catID != 0 {
		if _, err := cd.SelectedLinkerCategory(catID); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	}
	return cd.LinkerItemsForUser(catID, offset)
}

// ExternalLink lazily resolves metadata for url.
func (cd *CoreData) ExternalLink(url string) *db.ExternalLink {
	if cd.queries == nil {
		return nil
	}
	if cd.externalLinks == nil {
		cd.externalLinks = make(map[string]*lazy.Value[*db.ExternalLink])
	}
	lv, ok := cd.externalLinks[url]
	if !ok {
		lv = &lazy.Value[*db.ExternalLink]{}
		cd.externalLinks[url] = lv
	}
	link, err := lv.Load(func() (*db.ExternalLink, error) {
		l, err := cd.queries.GetExternalLink(cd.ctx, url)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		return l, nil
	})
	if err != nil {
		log.Printf("load external link: %v", err)
	}
	return link
}

// RegisterExternalLinkClick records click statistics for url.
func (cd *CoreData) RegisterExternalLinkClick(url string) {
	if cd.queries == nil {
		return
	}
	if err := cd.queries.RegisterExternalLinkClick(cd.ctx, url); err != nil {
		log.Printf("record external link click: %v", err)
	}
}

// HasAdminRole reports whether the current user has the administrator role.
func (cd *CoreData) HasAdminRole() bool {
	return cd.HasRole("administrator")
}

// HasContentWriterRole reports whether the current user has the content writer role.
func (cd *CoreData) HasContentWriterRole() bool {
	return cd.HasRole("content writer")
}

// ExecuteSiteTemplate renders the named site template using cd's helper
// functions. It wraps templates.GetCompiledSiteTemplates(cd.Funcs(r)).
func (cd *CoreData) ExecuteSiteTemplate(w io.Writer, r *http.Request, name string, data any) error {
	return templates.GetCompiledSiteTemplates(cd.Funcs(r)).ExecuteTemplate(w, name, data)
}
