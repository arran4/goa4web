//go:build sqlite || sqlite3

package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/handlers/auth"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/testdata/scenarios"
)

func TestScenarioServeCmd_ParseCLI(t *testing.T) {
	root := &rootCmd{fs: flag.NewFlagSet("goa4web", flag.ContinueOnError)}
	parent, err := parseScenarioCmd(root, []string{"serve"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	// 1. Missing path error
	serveCmdEmpty, err := parseScenarioServeCmd(parent, []string{})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if err := serveCmdEmpty.Run(); err == nil || !strings.Contains(err.Error(), "scenario path required") {
		t.Fatalf("expected scenario path required error, got: %v", err)
	}

	// 2. Directory path argument
	serveCmdDir, err := parseScenarioServeCmd(parent, []string{"testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if serveCmdDir.Path != "testdata/scenarios/100-private-forum" {
		t.Errorf("expected path testdata/scenarios/100-private-forum, got %q", serveCmdDir.Path)
	}

	// 3. Direct file path argument
	serveCmdFile, err := parseScenarioServeCmd(parent, []string{"testdata/scenarios/100-private-forum/scenario.txtar"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if serveCmdFile.Path != "testdata/scenarios/100-private-forum/scenario.txtar" {
		t.Errorf("expected path testdata/scenarios/100-private-forum/scenario.txtar, got %q", serveCmdFile.Path)
	}

	// 4. Custom listen flag on subcommand
	serveCmdListen, err := parseScenarioServeCmd(parent, []string{"--listen", ":9090", "testdata/scenarios/100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	if serveCmdListen.Listen != ":9090" {
		t.Errorf("expected listen :9090, got %q", serveCmdListen.Listen)
	}
	if serveCmdListen.Path != "testdata/scenarios/100-private-forum" {
		t.Errorf("expected path testdata/scenarios/100-private-forum, got %q", serveCmdListen.Path)
	}
}

func TestScenarioServeCmd_DeriveScenarioBaseURL(t *testing.T) {
	tests := []struct {
		listen string
		want   string
	}{
		{"", "http://localhost:8080"},
		{":8080", "http://localhost:8080"},
		{":9090", "http://localhost:9090"},
		{"0.0.0.0:8080", "http://localhost:8080"},
		{"[::]:8080", "http://localhost:8080"},
		{"127.0.0.1:8080", "http://127.0.0.1:8080"},
		{"http://localhost:8080", "http://localhost:8080"},
		{"https://dev.local:8443/", "https://dev.local:8443"},
	}

	for _, tt := range tests {
		got := deriveScenarioBaseURL(tt.listen)
		if got != tt.want {
			t.Errorf("deriveScenarioBaseURL(%q) = %q, want %q", tt.listen, got, tt.want)
		}
	}
}

func TestScenarioServeCmd_BootstrapAndSameDatabaseInvariant(t *testing.T) {
	ctx := context.Background()

	root, err := parseRoot([]string{"goa4web", "--listen", ":8082", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	defer cleanup()

	// 1. Verify same-database invariant: the database connection in srv is dbConn
	if srv.DB != dbConn {
		t.Errorf("server DB (%p) does not match bootstrap DB (%p)", srv.DB, dbConn)
	}

	// 2. Verify schema version is migrated to 98
	var currentVersion int64
	err = dbConn.QueryRowContext(ctx, "SELECT version_id FROM goose_db_version ORDER BY id DESC LIMIT 1;").Scan(&currentVersion)
	if err != nil {
		t.Fatalf("query goose_db_version: %v", err)
	}
	if currentVersion != 98 {
		t.Errorf("expected migrated schema version 98, got %d", currentVersion)
	}

	// 3. Verify language English is seeded
	var langName string
	err = dbConn.QueryRowContext(ctx, "SELECT nameof FROM language WHERE id = 1;").Scan(&langName)
	if err != nil {
		t.Fatalf("query language: %v", err)
	}
	if langName != "English" {
		t.Errorf("expected language English, got %q", langName)
	}

	// 4. Verify users Alice and Bob exist and have correct passwords
	var aliceUID, bobUID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'alice';").Scan(&aliceUID)
	if err != nil {
		t.Fatalf("query alice idusers: %v", err)
	}
	err = dbConn.QueryRowContext(ctx, "SELECT idusers FROM users WHERE username = 'bob';").Scan(&bobUID)
	if err != nil {
		t.Fatalf("query bob idusers: %v", err)
	}

	var alicePasswd, aliceAlg string
	err = dbConn.QueryRowContext(ctx, "SELECT passwd, passwd_algorithm FROM passwords WHERE users_idusers = ?;", aliceUID).Scan(&alicePasswd, &aliceAlg)
	if err != nil {
		t.Fatalf("query alice password: %v", err)
	}
	if !auth.VerifyPassword("alice-test", alicePasswd, aliceAlg) {
		t.Errorf("alice password verification failed for 'alice-test'")
	}

	var bobPasswd, bobAlg string
	err = dbConn.QueryRowContext(ctx, "SELECT passwd, passwd_algorithm FROM passwords WHERE users_idusers = ?;", bobUID).Scan(&bobPasswd, &bobAlg)
	if err != nil {
		t.Fatalf("query bob password: %v", err)
	}
	if !auth.VerifyPassword("bob-test", bobPasswd, bobAlg) {
		t.Errorf("bob password verification failed for 'bob-test'")
	}

	// 5. Verify private topic "Staff Room" was created
	querier := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, querier, srv.Config)
	aliceCD := cd.ForUser(aliceUID)
	bobCD := cd.ForUser(bobUID)

	// Alice has global see and create; Bob has global see (not create)
	if !aliceCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Error("expected Alice to have global privateforum create permission")
	}
	if !aliceCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Error("expected Alice to have global privateforum see permission")
	}
	if !bobCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Error("expected Bob to have global privateforum see permission")
	}
	if bobCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Error("expected Bob NOT to have global privateforum create permission")
	}

	// Verify topic
	var topicID int32
	err = dbConn.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room';").Scan(&topicID)
	if err != nil {
		t.Fatalf("query Staff Room topic: %v", err)
	}
	topic, err := querier.GetForumTopicById(ctx, topicID)
	if err != nil {
		t.Fatalf("GetForumTopicById: %v", err)
	}
	if topic.Title.String != "Staff Room" {
		t.Errorf("topic title = %q, want 'Staff Room'", topic.Title.String)
	}
	if topic.Description.String != "Private discussion for Alice and Bob" {
		t.Errorf("topic description = %q, want 'Private discussion for Alice and Bob'", topic.Description.String)
	}

	// Verify topic-specific participant grants
	for _, act := range []string{"see", "view", "post", "reply"} {
		if !aliceCD.HasGrant("privateforum", "topic", act, topicID) {
			t.Errorf("expected Alice to have grant %s on private topic %d", act, topicID)
		}
		if !bobCD.HasGrant("privateforum", "topic", act, topicID) {
			t.Errorf("expected Bob to have grant %s on private topic %d", act, topicID)
		}
	}

	// 6. Security boundary: verify an unrelated user has NO global privateforum grants
	initUserRes, err := querier.SystemInsertUser(ctx, sql.NullString{String: "baseline_user", Valid: true})
	if err != nil {
		t.Fatalf("insert baseline user: %v", err)
	}
	baselineUID := int32(initUserRes)
	if err := querier.SystemCreateUserRole(ctx, db.SystemCreateUserRoleParams{
		UsersIdusers: baselineUID,
		Name:         "user",
	}); err != nil {
		t.Fatalf("assign user role: %v", err)
	}
	baselineCD := cd.ForUser(baselineUID)
	if baselineCD.HasGrant("privateforum", "topic", "create", 0) {
		t.Fatal("security violation: baseline user unexpectedly has global privateforum create permission")
	}
	if baselineCD.HasGrant("privateforum", "topic", "see", 0) {
		t.Fatal("security violation: baseline user unexpectedly has global privateforum see permission")
	}
}

func TestScenarioServeCmd_IsolationFromConfiguredProductionSettings(t *testing.T) {
	ctx := context.Background()

	// Configure root with production-like settings across database, URLs, email, uploads, DLQ
	root, err := parseRoot([]string{
		"goa4web",
		"--db-conn", "mysql://invalid-nonexistent-host:9999/dummy",
		"--db-driver", "mysql",
		"--external-url", "https://production.example.test",
		"--hostname", "https://production.example.test",
		"--host", "production.example.test",
		"--email-provider", "smtp",
		"--smtp-host", "smtp.production.test",
		"--smtp-port", "587",
		"--smtp-user", "prod_user",
		"--smtp-pass", "prod_pass",
		"--sendgrid-key", "SG.production_key",
		"scenario", "serve",
		"--listen", ":9090",
		"scenarios/valid",
	})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	// Manually set non-flag production fields on parent config to ensure total override
	root.cfg.BaseURL = "https://production.example.test"
	root.cfg.EmailEnabled = true
	root.cfg.ImageUploadProvider = "s3"
	root.cfg.ImageUploadS3URL = "s3://prod-bucket/uploads"
	root.cfg.ImageCacheProvider = "s3"
	root.cfg.ImageCacheS3URL = "s3://prod-bucket/cache"
	root.cfg.ImageUploadDir = "/var/www/production/uploads"
	root.cfg.ImageCacheDir = "/var/www/production/cache"
	root.cfg.DLQProvider = "file"
	root.cfg.DLQFile = "/var/log/production_dlq.json"

	fsys := fstest.MapFS{
		"scenarios/valid/scenario.txtar": &fstest.MapFile{
			Data: []byte(`-- scenario.meta --
Format: goa4web-scenario/v1
Name: isolation-test

-- 01-alice.event --
Op: user.create
Ref: alice
Username: alice
Email: alice@example.test
Password: pass
At: 2026-08-01T09:00:00Z
`),
		},
	}

	parent, err := parseScenarioCmd(root, []string{"serve", "--listen", ":9090", "scenarios/valid"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{"--listen", ":9090", "scenarios/valid"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = fsys

	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap should succeed in isolation from production settings, but failed: %v", err)
	}

	// 1. Database isolation
	if srv.Config.DBDriver != "sqlite3" {
		t.Errorf("expected DBDriver sqlite3, got %q", srv.Config.DBDriver)
	}
	if srv.Config.DBConn != "" {
		t.Errorf("expected DBConn cleared, got %q", srv.Config.DBConn)
	}

	// 2. BaseURL isolation: derived from :9090, not inherited from production
	if srv.Config.BaseURL != "http://localhost:9090" {
		t.Errorf("expected local BaseURL http://localhost:9090, got %q", srv.Config.BaseURL)
	}
	if srv.Config.HTTPHostname != "" {
		t.Errorf("expected HTTPHostname cleared, got %q", srv.Config.HTTPHostname)
	}
	if srv.Config.ExternalURL != "" {
		t.Errorf("expected ExternalURL cleared, got %q", srv.Config.ExternalURL)
	}

	// 3. Email isolation
	if srv.Config.EmailEnabled != false {
		t.Errorf("expected EmailEnabled false, got %v", srv.Config.EmailEnabled)
	}
	if srv.Config.EmailProvider != "log" {
		t.Errorf("expected EmailProvider 'log', got %q", srv.Config.EmailProvider)
	}
	if srv.Config.EmailSMTPHost != "" {
		t.Errorf("expected EmailSMTPHost cleared, got %q", srv.Config.EmailSMTPHost)
	}
	if srv.Config.EmailSendGridKey != "" {
		t.Errorf("expected EmailSendGridKey cleared, got %q", srv.Config.EmailSendGridKey)
	}

	// 4. Image upload / cache isolation
	if srv.Config.ImageUploadProvider != "local" {
		t.Errorf("expected ImageUploadProvider local, got %q", srv.Config.ImageUploadProvider)
	}
	if srv.Config.ImageCacheProvider != "local" {
		t.Errorf("expected ImageCacheProvider local, got %q", srv.Config.ImageCacheProvider)
	}
	if srv.Config.ImageUploadS3URL != "" {
		t.Errorf("expected ImageUploadS3URL cleared, got %q", srv.Config.ImageUploadS3URL)
	}
	if srv.Config.ImageCacheS3URL != "" {
		t.Errorf("expected ImageCacheS3URL cleared, got %q", srv.Config.ImageCacheS3URL)
	}
	if srv.Config.ImageUploadDir == "/var/www/production/uploads" {
		t.Error("expected ImageUploadDir to be overridden with temporary directory")
	}

	// 5. DLQ isolation
	if srv.Config.DLQProvider != "db" {
		t.Errorf("expected DLQProvider 'db', got %q", srv.Config.DLQProvider)
	}
	if srv.Config.DLQFile != "" {
		t.Errorf("expected DLQFile cleared, got %q", srv.Config.DLQFile)
	}

	// 6. Temporary directory exists during run
	uploadDir := srv.Config.ImageUploadDir
	tempDir := filepath.Dir(uploadDir)
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatalf("expected temporary scenario directory %s to exist during run", tempDir)
	}

	// 7. Alice exists in the ephemeral SQLite database
	var aliceCount int
	if err := dbConn.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE username = 'alice';").Scan(&aliceCount); err != nil {
		t.Fatalf("query alice in ephemeral sqlite DB: %v", err)
	}
	if aliceCount != 1 {
		t.Errorf("expected 1 alice user in ephemeral DB, got %d", aliceCount)
	}

	// 8. Cleanup removes the temporary directory
	cleanup()
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("expected temporary scenario directory %s to be removed on cleanup, stat err: %v", tempDir, err)
	}
}

func TestScenarioServeCmd_HTTPSmokeTest(t *testing.T) {
	ctx := context.Background()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	srv, _, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	defer cleanup()

	// Test 1: HTTP GET / returns 200 OK
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET / returned status %d, want %d", rec.Code, http.StatusOK)
	}

	// Test 2: HTTP GET /login returns 200 OK
	loginReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRec := httptest.NewRecorder()
	srv.Router.ServeHTTP(loginRec, loginReq)

	if loginRec.Code != http.StatusOK {
		t.Errorf("GET /login returned status %d, want %d", loginRec.Code, http.StatusOK)
	}

	// Test 3: HTTP GET /privateforum/ (requires login, redirects or returns 200/403/302 appropriately)
	pfReq := httptest.NewRequest(http.MethodGet, "/privateforum/", nil)
	pfRec := httptest.NewRecorder()
	srv.Router.ServeHTTP(pfRec, pfReq)

	// An unauthenticated user accessing private forum gets redirected to login or denied
	if pfRec.Code != http.StatusOK && pfRec.Code != http.StatusFound && pfRec.Code != http.StatusForbidden && pfRec.Code != http.StatusSeeOther {
		t.Errorf("GET /privateforum/ returned unexpected status %d", pfRec.Code)
	}
}

func TestScenarioServeCmd_PrivateForumCreateThreadHTTP(t *testing.T) {
	ctx := context.Background()

	root, err := parseRoot([]string{"goa4web", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()
	parent, err := parseScenarioCmd(root, []string{"serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}
	serveCmd, err := parseScenarioServeCmd(parent, []string{"100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	srv, dbConn, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap: %v", err)
	}
	defer cleanup()

	httpServer := httptest.NewServer(srv.Router)
	defer httpServer.Close()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar.New: %v", err)
	}
	client := &http.Client{Jar: jar}

	loginPage := scenarioHTTPGet(t, client, httpServer.URL+"/login")
	loginToken := scenarioCSRFToken(t, loginPage)
	loginForm := url.Values{
		"username":           {"alice"},
		"password":           {"alice-test"},
		"task":               {"Login"},
		"gorilla.csrf.Token": {loginToken},
	}
	loginResponse := scenarioHTTPPostForm(t, client, httpServer.URL+"/login", loginForm)
	if strings.Contains(loginResponse, "Invalid username or password") {
		t.Fatal("Alice's scenario credentials were rejected")
	}

	privatePage := scenarioHTTPGet(t, client, httpServer.URL+"/private")
	for _, visible := range []string{"Staff Room", "Coordination"} {
		if !strings.Contains(privatePage, visible) {
			t.Fatalf("authenticated private page does not contain %q", visible)
		}
	}
	if strings.Contains(privatePage, "Project Room") {
		t.Fatal("authenticated Alice page exposed Project Room")
	}

	var staffTopicID int32
	if err := dbConn.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Staff Room'").Scan(&staffTopicID); err != nil {
		t.Fatalf("query Staff Room: %v", err)
	}
	createURL := fmt.Sprintf("%s/private/topic/%d/thread", httpServer.URL, staffTopicID)
	createPage := scenarioHTTPGet(t, client, createURL)
	createToken := scenarioCSRFToken(t, createPage)
	const openingText = "Alice created this thread through the normal private-forum HTTP form."
	createForm := url.Values{
		"replytext":          {openingText},
		"task":               {"Create Thread"},
		"gorilla.csrf.Token": {createToken},
	}
	createResponse := scenarioHTTPPostForm(t, client, createURL, createForm)
	if strings.Contains(createResponse, "error-message") || strings.Contains(createResponse, "Access Denied") {
		t.Fatalf("create-thread response reported an error: %s", createResponse)
	}

	var threadID int32
	var languageID sql.NullInt32
	if err := dbConn.QueryRowContext(ctx, `
		SELECT forumthread_id, language_id
		FROM comments
		WHERE text = ?`, openingText).Scan(&threadID, &languageID); err != nil {
		t.Fatalf("query HTTP-created opening comment: %v", err)
	}
	if languageID.Valid {
		t.Fatalf("opening comment language_id = %d, want NULL when the form omits language", languageID.Int32)
	}

	threadURL := fmt.Sprintf("%s/private/topic/%d/thread/%d", httpServer.URL, staffTopicID, threadID)
	threadPage := scenarioHTTPGet(t, client, threadURL)
	if !strings.Contains(threadPage, openingText) {
		t.Fatal("HTTP-created thread is not visible on its normal private thread page")
	}

	bobJar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create Bob cookie jar: %v", err)
	}
	bobClient := &http.Client{Jar: bobJar}
	bobLoginPage := scenarioHTTPGet(t, bobClient, httpServer.URL+"/login")
	bobLoginForm := url.Values{
		"username":           {"bob"},
		"password":           {"bob-test"},
		"task":               {"Login"},
		"gorilla.csrf.Token": {scenarioCSRFToken(t, bobLoginPage)},
	}
	scenarioHTTPPostForm(t, bobClient, httpServer.URL+"/login", bobLoginForm)

	var projectTopicID, projectThreadID int32
	if err := dbConn.QueryRowContext(ctx, "SELECT idforumtopic FROM forumtopic WHERE title = 'Project Room'").Scan(&projectTopicID); err != nil {
		t.Fatalf("query Project Room: %v", err)
	}
	if err := dbConn.QueryRowContext(ctx,
		"SELECT forumthread_id FROM comments WHERE text = 'Project Room kickoff for Carol and Dave.'",
	).Scan(&projectThreadID); err != nil {
		t.Fatalf("query Project Room thread: %v", err)
	}
	projectURL := fmt.Sprintf("%s/private/topic/%d/thread/%d", httpServer.URL, projectTopicID, projectThreadID)
	projectResponse, err := bobClient.Get(projectURL)
	if err != nil {
		t.Fatalf("Bob GET inaccessible project thread: %v", err)
	}
	defer projectResponse.Body.Close()
	projectBody, err := io.ReadAll(projectResponse.Body)
	if err != nil {
		t.Fatalf("read inaccessible project thread response: %v", err)
	}
	if projectResponse.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("inaccessible project thread status = %d", projectResponse.StatusCode)
	}
	for _, secretText := range []string{"Project Room kickoff for Carol and Dave.", "Dave here. I can see the project conversation and reply."} {
		if strings.Contains(string(projectBody), secretText) {
			t.Fatalf("Bob's direct URL response leaked inaccessible content %q", secretText)
		}
	}
}

func scenarioHTTPGet(t *testing.T, client *http.Client, target string) string {
	t.Helper()
	response, err := client.Get(target)
	if err != nil {
		t.Fatalf("GET %s: %v", target, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read GET %s response: %v", target, err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET %s status = %d, body: %s", target, response.StatusCode, body)
	}
	return string(body)
}

func scenarioHTTPPostForm(t *testing.T, client *http.Client, target string, form url.Values) string {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("create POST %s: %v", target, err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Referer", target)
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("POST %s: %v", target, err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read POST %s response: %v", target, err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("POST %s status = %d, body: %s", target, response.StatusCode, body)
	}
	return string(body)
}

func scenarioCSRFToken(t *testing.T, body string) string {
	t.Helper()
	match := regexp.MustCompile(`name="gorilla\.csrf\.Token" value="([^"]+)"`).FindStringSubmatch(body)
	if len(match) != 2 {
		t.Fatal("response does not contain a CSRF token")
	}
	return match[1]
}

func TestScenarioServeCmd_GracefulShutdown(t *testing.T) {
	root, err := parseRoot([]string{"goa4web", "--listen", "127.0.0.1:0", "scenario", "serve", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseRoot: %v", err)
	}
	defer root.Close()

	parent, err := parseScenarioCmd(root, []string{"serve", "--listen", "127.0.0.1:0", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioCmd: %v", err)
	}

	serveCmd, err := parseScenarioServeCmd(parent, []string{"--listen", "127.0.0.1:0", "100-private-forum"})
	if err != nil {
		t.Fatalf("parseScenarioServeCmd: %v", err)
	}
	serveCmd.fsys = scenarios.FS

	ctx, cancel := context.WithCancel(context.Background())

	srv, _, cleanup, err := serveCmd.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("Bootstrap failed: %v", err)
	}
	defer cleanup()

	done := make(chan error, 1)
	go func() {
		done <- srv.RunContext(ctx)
	}()

	// Allow server to start then trigger graceful shutdown
	cancel()

	err = <-done
	if err != nil {
		t.Errorf("RunContext returned unexpected error on shutdown: %v", err)
	}
}
