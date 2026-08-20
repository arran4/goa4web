//go:build sqlite || sqlite3

package app_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/database"
	_ "github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/handlers/blogs"
	"github.com/arran4/goa4web/handlers/bookmarks"
	"github.com/arran4/goa4web/handlers/externallink"
	"github.com/arran4/goa4web/handlers/faq"
	"github.com/arran4/goa4web/handlers/forum"
	"github.com/arran4/goa4web/handlers/imagebbs"
	"github.com/arran4/goa4web/handlers/images"
	"github.com/arran4/goa4web/handlers/languages"
	"github.com/arran4/goa4web/handlers/linker"
	"github.com/arran4/goa4web/handlers/news"
	"github.com/arran4/goa4web/handlers/privateforum"
	"github.com/arran4/goa4web/handlers/search"
	"github.com/arran4/goa4web/handlers/user"
	"github.com/arran4/goa4web/handlers/writings"
	"github.com/arran4/goa4web/internal/app"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dbdrivers/dbdefaults"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/dlqdefaults"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/email/emaildefaults"
	"github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/sqlutil"
	"github.com/arran4/goa4web/internal/tasks"
	"github.com/arran4/goa4web/internal/upload/uploaddefaults"
	"github.com/arran4/goa4web/migrations"
	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

func setupTestSQLiteDB(t *testing.T, dir string) string {
	t.Helper()
	dbPath := filepath.Join(dir, "e2e_test.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite database: %v", err)
	}
	defer dbConn.Close()

	ctx := context.Background()

	// Apply migrations
	if err := dbstart.Apply(ctx, dbConn, migrations.FS, false, "sqlite3"); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Apply seed data
	seedSQL := database.SeedSQLForDriver("sqlite3")
	if err := sqlutil.RunStatements(ctx, dbConn, strings.NewReader(string(seedSQL))); err != nil {
		t.Fatalf("failed to run seed SQL: %v", err)
	}

	// Insert representative data
	dataStatements := []string{
		`INSERT INTO language (id, nameof) VALUES (1, 'English');`,
		`INSERT INTO users (idusers, username) VALUES
			(1, 'admin'),
			(2, 'testuser'),
			(3, 'writer');`,
		`INSERT INTO passwords (id, users_idusers, passwd, passwd_algorithm, created_at) VALUES
			(1, 1, 'password123', 'plaintext', CURRENT_TIMESTAMP),
			(2, 2, 'password123', 'plaintext', CURRENT_TIMESTAMP),
			(3, 3, 'password123', 'plaintext', CURRENT_TIMESTAMP);`,
		`INSERT INTO user_emails (user_id, email, verified_at, notification_priority) VALUES
			(1, 'admin@example.com', CURRENT_TIMESTAMP, 100),
			(2, 'user@example.com', CURRENT_TIMESTAMP, 100),
			(3, 'writer@example.com', CURRENT_TIMESTAMP, 100);`,
		`INSERT INTO user_roles (users_idusers, role_id) VALUES
			(1, (SELECT id FROM roles WHERE name = 'administrator')),
			(2, (SELECT id FROM roles WHERE name = 'user')),
			(3, (SELECT id FROM roles WHERE name = 'content writer'));`,
		`INSERT INTO user_language (users_idusers, language_id) VALUES
			(1, 1), (2, 1), (3, 1);`,
		`INSERT INTO preferences (language_id, users_idusers, emailforumupdates, page_size, auto_subscribe_replies) VALUES
			(1, 1, 1, 15, 1),
			(1, 2, 1, 15, 1),
			(1, 3, 1, 15, 1);`,
		`INSERT INTO forumcategory (idforumcategory, forumcategory_idforumcategory, title, description, language_id) VALUES
			(1, 0, 'General Discussion', 'Community discussions and chatter', 1),
			(2, 0, 'Technology', 'Technical questions, software, and programming', 1);`,
		`INSERT INTO forumtopic (idforumtopic, lastposter, forumcategory_idforumcategory, title, description, threads, comments, language_id, handler) VALUES
			(1, 1, 1, 'Welcome & Rules', 'Site announcements and guidelines', 1, 2, 1, ''),
			(2, 1, 1, 'General Chit-Chat', 'Chat about anything', 0, 0, 1, ''),
			(3, 2, 2, 'Go Programming', 'Discussions about the Go language and Goa4Web', 1, 1, 1, ''),
			(4, 1, 0, 'Staff Room', 'Private forum for administrators and moderators', 1, 1, 1, 'private');`,
		`INSERT INTO forumthread (idforumthread, firstpost, lastposter, forumtopic_idforumtopic, comments, lastaddition, locked) VALUES
			(1, 1, 2, 1, 2, CURRENT_TIMESTAMP, 0),
			(2, 3, 2, 3, 1, CURRENT_TIMESTAMP, 0),
			(3, 4, 1, 4, 1, CURRENT_TIMESTAMP, 0);`,
		`INSERT INTO comments (idcomments, forumthread_id, users_idusers, language_id, written, text) VALUES
			(1, 1, 1, 1, CURRENT_TIMESTAMP, 'Welcome to the Goa4Web community platform!'),
			(2, 1, 2, 1, CURRENT_TIMESTAMP, 'Thanks! Excited to be part of the community.'),
			(3, 2, 2, 1, CURRENT_TIMESTAMP, 'How do I run Goa4Web with the SQLite backend?'),
			(4, 3, 1, 1, CURRENT_TIMESTAMP, 'Confidential staff notes for internal review.');`,
		`INSERT INTO site_news (idsiteNews, forumthread_id, language_id, users_idusers, news, occurred) VALUES
			(1, 0, 1, 1, 'Goa4Web now supports SQLite backend out of the box!', CURRENT_TIMESTAMP);`,
		`INSERT INTO blogs (idblogs, forumthread_id, users_idusers, language_id, blog, written) VALUES
			(1, 0, 1, 1, 'Welcome to the official developer blog for Goa4Web.', CURRENT_TIMESTAMP);`,
		`INSERT INTO linker_category (id, title, position, sortorder) VALUES
			(1, 'Useful Resources', 1, 1);`,
		`INSERT INTO linker (id, language_id, author_id, category_id, thread_id, title, url, description, listed) VALUES
			(1, 1, 1, 1, 0, 'Official Go Website', 'https://go.dev', 'The home of the Go programming language.', CURRENT_TIMESTAMP);`,
		`INSERT INTO faq_categories (id, name, parent_category_id, language_id, priority) VALUES
			(1, 'General FAQ', NULL, 1, 1);`,
		`INSERT INTO faq (id, category_id, author_id, language_id, question, answer, priority, description) VALUES
			(1, 1, 1, 1, 'What is Goa4Web?', 'Goa4Web is an open source community platform written in Go.', 1, 'Overview FAQ');`,
		`INSERT INTO grants (created_at, user_id, role_id, section, item, rule_type, item_id, action, active) VALUES
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'forum', 'topic', 'allow', NULL, 'see', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'forum', 'topic', 'allow', NULL, 'view', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'forum', 'thread', 'allow', NULL, 'see', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'forum', 'thread', 'allow', NULL, 'view', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'user'), 'forum', 'topic', 'allow', NULL, 'post', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'user'), 'forum', 'topic', 'allow', NULL, 'reply', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'user'), 'forum', 'thread', 'allow', NULL, 'reply', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'news', NULL, 'allow', NULL, 'see', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'news', NULL, 'allow', NULL, 'view', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'faq', NULL, 'allow', NULL, 'see', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'faq', NULL, 'allow', NULL, 'view', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'linker', NULL, 'allow', NULL, 'see', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'linker', NULL, 'allow', NULL, 'view', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'blogs', NULL, 'allow', NULL, 'see', 1),
			(CURRENT_TIMESTAMP, NULL, (SELECT id FROM roles WHERE name = 'anyone'), 'blogs', NULL, 'allow', NULL, 'view', 1),
			(CURRENT_TIMESTAMP, 1, NULL, 'privateforum', 'topic', 'allow', 4, 'see', 1),
			(CURRENT_TIMESTAMP, 1, NULL, 'privateforum', 'topic', 'allow', 4, 'view', 1),
			(CURRENT_TIMESTAMP, 1, NULL, 'privateforum', 'topic', 'allow', 4, 'post', 1),
			(CURRENT_TIMESTAMP, 1, NULL, 'privateforum', 'topic', 'allow', 4, 'reply', 1),
			(CURRENT_TIMESTAMP, 1, NULL, 'privateforum_thread', 'thread', 'allow', 3, 'view', 1),
			(CURRENT_TIMESTAMP, 1, NULL, 'privateforum_thread', 'thread', 'allow', 3, 'reply', 1);`,
		`INSERT INTO bookmarks (idbookmarks, users_idusers, list) VALUES
			(1, 2, '{"thread_ids":[1,2]}');`,
	}

	for _, stmt := range dataStatements {
		if _, err := dbConn.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("failed to insert seed statement: %v\nSQL: %s", err, stmt)
		}
	}

	if os.Getenv("UPDATE_SQLITE_SEED_DUMP") == "1" {
		dumpOut, err := exec.Command("sqlite3", dbPath, ".dump").Output()
		if err == nil && len(dumpOut) > 0 {
			dumpPath := filepath.Join("..", "..", "testdata", "schema", "testing_seed.sqlite.sql")
			_ = os.WriteFile(dumpPath, dumpOut, 0644)
		}
	}

	return dbPath
}

func registerAllModules(reg *router.Registry, ah *admin.Handlers) {
	ah.Register(reg)
	auth.Register(reg)
	blogs.Register(reg)
	bookmarks.Register(reg)
	faq.Register(reg)
	forum.Register(reg)
	imagebbs.Register(reg)
	languages.Register(reg)
	linker.Register(reg)
	news.Register(reg)
	privateforum.Register(reg)
	search.Register(reg)
	images.Register(reg)
	externallink.Register(reg)
	user.Register(reg)
	writings.Register(reg)
}

func TestServerEndToEndWithSQLite(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := setupTestSQLiteDB(t, tempDir)

	dbReg := dbdrivers.NewRegistry()
	emailReg := email.NewRegistry()
	dlqReg := dlq.NewRegistry()
	tasksReg := tasks.NewRegistry()
	routerReg := router.NewRegistry()
	adminHandlers := admin.New()

	registerAllModules(routerReg, adminHandlers)
	emaildefaults.Register(emailReg)
	dlqdefaults.RegisterDefaults(dlqReg, emailReg)
	dbdefaults.Register(dbReg)
	uploaddefaults.Register()

	cfg := config.NewRuntimeConfig()
	cfg.DBDriver = "sqlite3"
	cfg.DBConn = dbPath
	cfg.ImageUploadProvider = "local"
	cfg.ImageUploadDir = filepath.Join(tempDir, "uploads")
	cfg.ImageCacheProvider = "local"
	cfg.ImageCacheDir = filepath.Join(tempDir, "cache")
	cfg.SessionSecret = "01234567890123456789012345678901"
	cfg.CSRFEnabled = false
	cfg.DefaultLanguage = "English"
	cfg.AutoMigrate = false

	if err := os.MkdirAll(cfg.ImageUploadDir, 0755); err != nil {
		t.Fatalf("failed to create upload dir: %v", err)
	}
	if err := os.MkdirAll(cfg.ImageCacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	origSessionName := core.SessionName
	core.SessionName = "test-session"
	defer func() { core.SessionName = origSessionName }()

	ctx := context.Background()
	srv, err := app.NewServer(
		ctx,
		cfg,
		adminHandlers,
		app.WithDBRegistry(dbReg),
		app.WithDLQRegistry(dlqReg),
		app.WithEmailRegistry(emailReg),
		app.WithTasksRegistry(tasksReg),
		app.WithRouterRegistry(routerReg),
		app.WithSessionSecret(cfg.SessionSecret),
	)
	if err != nil {
		t.Fatalf("failed to create server with sqlite: %v", err)
	}

	origHandle := tasks.Handle
	tasks.Handle = func(w http.ResponseWriter, r *http.Request, p tasks.Template, data any) error {
		err := origHandle(w, r, p, data)
		if err != nil {
			t.Logf("tasks.Handle error for template %s: %v", p, err)
		}
		return err
	}
	defer func() { tasks.Handle = origHandle }()

	q := db.New(srv.DB)
	cats, err := q.GetAllForumCategories(ctx, db.GetAllForumCategoriesParams{ViewerID: 0})
	if err != nil {
		t.Fatalf("direct GetAllForumCategories failed: %v", err)
	}
	t.Logf("direct GetAllForumCategories returned %d categories", len(cats))

	topics, err := q.GetForumTopicsByCategoryId(ctx, db.GetForumTopicsByCategoryIdParams{CategoryID: 1, ViewerID: 0})
	if err != nil {
		t.Fatalf("direct GetForumTopicsByCategoryId failed: %v", err)
	}
	t.Logf("direct GetForumTopicsByCategoryId returned %d topics", len(topics))

	req := httptest.NewRequest(http.MethodGet, "/forum", nil)
	w := httptest.NewRecorder()
	cd, req := srv.GetCoreData(w, req)
	if cd == nil {
		t.Fatalf("GetCoreData returned nil")
	}
	catRows, err := cd.ForumCategories()
	if err != nil {
		t.Fatalf("cd.ForumCategories failed: %v", err)
	}
	t.Logf("cd.ForumCategories returned %d categories", len(catRows))
	topRows, err := cd.ForumTopics(0)
	if err != nil {
		t.Fatalf("cd.ForumTopics(0) failed: %v", err)
	}
	t.Logf("cd.ForumTopics(0) returned %d topics", len(topRows))
	for _, tr := range topRows {
		pub, _, err := cd.ThreadPublicLabels(tr.Idforumtopic)
		if err != nil {
			t.Fatalf("cd.ThreadPublicLabels(%d) failed: %v", tr.Idforumtopic, err)
		}
		t.Logf("cd.ThreadPublicLabels(%d) returned %v", tr.Idforumtopic, pub)
	}
	var topicRows []*forum.ForumtopicPlus
	for _, tr := range topRows {
		topicRows = append(topicRows, &forum.ForumtopicPlus{
			Idforumtopic:                 tr.Idforumtopic,
			Lastposter:                   tr.Lastposter,
			ForumcategoryIdforumcategory: tr.ForumcategoryIdforumcategory,
			Title:                        tr.Title,
			Description:                  tr.Description,
			Threads:                      tr.Threads,
			Comments:                     tr.Comments,
			Lastaddition:                 tr.Lastaddition,
			Lastposterusername:           tr.Lastposterusername,
		})
	}
	categoryTree := forum.NewCategoryTree(catRows, topicRows)
	type ForumData struct {
		Categories              []*forum.ForumcategoryPlus
		Admin                   bool
		CopyDataToSubCategories func(rootCategory *forum.ForumcategoryPlus) *ForumData
		Category                *forum.ForumcategoryPlus
		Back                    bool
	}
	var fd ForumData
	fd.Categories = categoryTree.CategoryChildrenLookup[0]
	fd.CopyDataToSubCategories = func(rootCategory *forum.ForumcategoryPlus) *ForumData {
		d := fd
		d.Categories = rootCategory.Categories
		d.Category = rootCategory
		d.Back = false
		return &d
	}
	wExec := httptest.NewRecorder()
	execErr := cd.ExecuteSiteTemplate(wExec, req, "domains/forum/page.gohtml", &fd)
	t.Logf("ExecuteSiteTemplate direct error: %v, body len: %d", execErr, wExec.Body.Len())
	if execErr != nil {
		t.Fatalf("ExecuteSiteTemplate direct failed: %v", execErr)
	}

	// Test GET /news first to verify router
	reqNews := httptest.NewRequest(http.MethodGet, "/news", nil)
	wNews := httptest.NewRecorder()
	cdNews, reqNews := srv.GetCoreData(wNews, reqNews)
	if cdNews == nil {
		t.Fatalf("GetCoreData for news returned nil")
	}
	newsErr := cdNews.ExecuteSiteTemplate(wNews, reqNews, "domains/news/page.gohtml", struct{}{})
	t.Logf("ExecuteSiteTemplate news error: %v, body len: %d", newsErr, wNews.Body.Len())
	if newsErr != nil {
		t.Fatalf("ExecuteSiteTemplate news failed: %v", newsErr)
	}

	// 1. Test GET /forum (Forum categories and topics)
	req = httptest.NewRequest(http.MethodGet, "/forum", nil)
	w = httptest.NewRecorder()
	_, req = srv.GetCoreData(w, req)
	forum.Page(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /forum returned status %d, expected 200. Body:\n%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "General Discussion") {
		t.Errorf("GET /forum body does not contain 'General Discussion'")
	}
	if !strings.Contains(body, "Technology") {
		t.Errorf("GET /forum body does not contain 'Technology'")
	}

	// 2. Test GET /forum/topic/1
	req = httptest.NewRequest(http.MethodGet, "/forum/topic/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1"})
	w = httptest.NewRecorder()
	_, req = srv.GetCoreData(w, req)
	forum.TopicsPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /forum/topic/1 returned status %d, expected 200. Body:\n%s", w.Code, w.Body.String())
	}
	body = w.Body.String()
	if !strings.Contains(body, "Welcome &amp; Rules") && !strings.Contains(body, "Welcome & Rules") {
		t.Errorf("GET /forum/topic/1 body does not contain 'Welcome & Rules'")
	}

	// 3. Test GET /forum/topic/1/thread/1
	req = httptest.NewRequest(http.MethodGet, "/forum/topic/1/thread/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	w = httptest.NewRecorder()
	cdThread, req := srv.GetCoreData(w, req)
	cdThread.SetCurrentSection("forum")
	forum.RequireThreadAndTopic(http.HandlerFunc(forum.ThreadPage)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /forum/topic/1/thread/1 returned status %d, expected 200. Body:\n%s", w.Code, w.Body.String())
	}
	body = w.Body.String()
	if !strings.Contains(body, "Welcome to the Goa4Web community platform!") {
		t.Errorf("GET /forum/topic/1/thread/1 body missing comment 1")
	}
	if !strings.Contains(body, "Thanks! Excited to be part of the community.") {
		t.Errorf("GET /forum/topic/1/thread/1 body missing comment 2")
	}

	// 4. Test GET /news
	req = httptest.NewRequest(http.MethodGet, "/news", nil)
	w = httptest.NewRecorder()
	_, req = srv.GetCoreData(w, req)
	news.NewsPageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /news returned status %d, expected 200. Body:\n%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Goa4Web now supports SQLite") {
		t.Errorf("GET /news body missing expected news content")
	}

	// 5. Test GET /faq
	req = httptest.NewRequest(http.MethodGet, "/faq", nil)
	w = httptest.NewRecorder()
	_, req = srv.GetCoreData(w, req)
	faq.Page(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /faq returned status %d, expected 200. Body:\n%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "What is Goa4Web?") {
		t.Errorf("GET /faq body missing expected FAQ question")
	}

	// 6. Test GET /linker
	req = httptest.NewRequest(http.MethodGet, "/linker", nil)
	w = httptest.NewRecorder()
	_, req = srv.GetCoreData(w, req)
	linker.LinkerPage(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /linker returned status %d, expected 200. Body:\n%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Official Go Website") {
		t.Errorf("GET /linker body missing expected linker item")
	}

	// 7. Test write path: Insert a new comment using queries on SQLite
	querier := db.New(srv.DB)
	newCommentID, err := querier.CreateCommentInSectionForCommenter(ctx, db.CreateCommentInSectionForCommenterParams{
		CommenterID:   sql.NullInt32{Int32: 2, Valid: true},
		ForumthreadID: 1,
		LanguageID:    sql.NullInt32{Int32: 1, Valid: true},
		Text:          sql.NullString{String: "This is an automated test comment inserted via SQLite!", Valid: true},
		Written:       sql.NullTime{Time: time.Now(), Valid: true},
		Section:       "forum",
		ItemType:      sql.NullString{String: "thread", Valid: true},
		Action:        "reply",
		ItemID:        sql.NullInt32{Int32: 1, Valid: true},
	})
	if err != nil {
		t.Fatalf("failed to insert comment into SQLite: %v", err)
	}
	if newCommentID <= 0 {
		t.Fatalf("invalid new comment id: %d", newCommentID)
	}

	// Verify the newly written comment is rendered by the server
	req = httptest.NewRequest(http.MethodGet, "/forum/topic/1/thread/1", nil)
	req = mux.SetURLVars(req, map[string]string{"topic": "1", "thread": "1"})
	w = httptest.NewRecorder()
	cdThread2, req := srv.GetCoreData(w, req)
	cdThread2.SetCurrentSection("forum")
	forum.RequireThreadAndTopic(http.HandlerFunc(forum.ThreadPage)).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /forum/topic/1/thread/1 returned status %d, expected 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "This is an automated test comment inserted via SQLite!") {
		t.Errorf("GET /forum/topic/1/thread/1 does not reflect newly inserted comment")
	}
}

func TestRestoreSQLiteTestingSeed(t *testing.T) {
	tempDir := t.TempDir()
	dumpPath := filepath.Join("..", "..", "testdata", "schema", "testing_seed.sqlite.sql")
	dumpSQL, err := os.ReadFile(dumpPath)
	if err != nil {
		t.Fatalf("failed to read testing_seed.sqlite.sql: %v", err)
	}

	dbPath := filepath.Join(tempDir, "restored.db")
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to open restored sqlite database: %v", err)
	}
	defer dbConn.Close()

	if err := sqlutil.RunStatements(context.Background(), dbConn, strings.NewReader(string(dumpSQL))); err != nil {
		t.Fatalf("failed to execute restored dump SQL: %v", err)
	}

	// Verify database can be queried with sqlc
	q := db.New(dbConn)
	cats, err := q.GetAllForumCategories(context.Background(), db.GetAllForumCategoriesParams{})
	if err != nil {
		t.Fatalf("failed to query restored database: %v", err)
	}
	if len(cats) == 0 {
		t.Errorf("restored database contains 0 forum categories, expected >0")
	}
}
