package main

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	. "github.com/arran4/gorillamuxlogic"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/arran4/goa4web/config"
)

var configFile string

var (
	//	// Replace these with your Google OAuth2 credentials
	//	clientID     = ""
	//	clientSecret = ""
	//	redirectURL  = "http://localhost:8080/callback"
	//
	//	// Change this to your desired session key
	sessionName           = "my-session"
	sessionSecretFlag     = flag.String("session-secret", "", "session secret key")
	sessionSecretFileFlag = flag.String("session-secret-file", "", "path to session secret file")
	//sessionKey  = "authenticated"
	store *sessions.CookieStore

	configFileFlag = flag.String("config-file", "", "path to application configuration file")

	emailCfgPath      = flag.String("email-config", "", "path to email configuration file")
	emailProviderFlag = flag.String("email-provider", "", "email provider")
	smtpHostFlag      = flag.String("smtp-host", "", "SMTP host")
	smtpPortFlag      = flag.String("smtp-port", "", "SMTP port")
	smtpUserFlag      = flag.String("smtp-user", "", "SMTP user")
	smtpPassFlag      = flag.String("smtp-pass", "", "SMTP pass")
	awsRegionFlag     = flag.String("aws-region", "", "AWS region")
	jmapEndpointFlag  = flag.String("jmap-endpoint", "", "JMAP endpoint")
	jmapAccountFlag   = flag.String("jmap-account", "", "JMAP account")
	jmapIdentityFlag  = flag.String("jmap-identity", "", "JMAP identity")
	jmapUserFlag      = flag.String("jmap-user", "", "JMAP user")
	jmapPassFlag      = flag.String("jmap-pass", "", "JMAP pass")
	sendGridKeyFlag   = flag.String("sendgrid-key", "", "SendGrid API key")

	dbCfgPath          = flag.String("db-config", "", "path to database configuration file")
	dbUserFlag         = flag.String("db-user", "", "database user")
	dbPassFlag         = flag.String("db-pass", "", "database password")
	dbHostFlag         = flag.String("db-host", "", "database host")
	dbPortFlag         = flag.String("db-port", "", "database port")
	dbNameFlag         = flag.String("db-name", "", "database name")
	dbLogVerbosityFlag = flag.Int("db-log-verbosity", 0, "database logging verbosity")

	listenFlag          = flag.String("listen", ":8080", "server listen address")
	hostnameFlag        = flag.String("hostname", "", "server base URL")
	httpCfgPath         = flag.String("http-config", "", "path to HTTP configuration file")
	feedsEnabledFlag    = flag.String("feeds-enabled", "", "enable or disable feeds")
	statsStartYearFlag  = flag.String("stats-start-year", "", "start year for usage stats")
	pageSizeMinFlag     = flag.Int("page-size-min", 0, "minimum allowed page size")
	pageSizeMaxFlag     = flag.Int("page-size-max", 0, "maximum allowed page size")
	pageSizeDefaultFlag = flag.Int("page-size-default", 0, "default page size")
	paginationCfgPath   = flag.String("pagination-config", "", "path to pagination configuration file")
	listenFlagSet       bool
	hostnameFlagSet     bool
	feedsFlagSet        bool

	srv *Server
	//
	//	oauth2Config = oauth2.Config{
	//		ClientID:     clientID,
	//		ClientSecret: clientSecret,
	//		RedirectURL:  redirectURL,
	//		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	//		Endpoint:     endpoints.Google,
	//	}

	version = "dev"
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func run() error {
	early := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	var cfgPath string
	early.StringVar(&cfgPath, "config-file", "", "path to application configuration file")
	_ = early.Parse(os.Args[1:])
	if cfgPath == "" {
		cfgPath = os.Getenv(config.EnvConfigFile)
	}
	appCfg := loadAppConfigFile(cfgPath)

	flag.Parse()

	configFile = *configFileFlag
	if configFile == "" {
		configFile = cfgPath
	}

	flag.CommandLine.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "listen":
			listenFlagSet = true
		case "hostname":
			hostnameFlagSet = true
		case "feeds-enabled":
			feedsFlagSet = true
		}
	})

	sessionSecretPath := *sessionSecretFileFlag
	if sessionSecretPath == "" {
		if v, ok := appCfg["SESSION_SECRET_FILE"]; ok {
			sessionSecretPath = v
		}
	}
	sessionSecret, err := loadSessionSecret(*sessionSecretFlag, sessionSecretPath)
	if err != nil {
		return fmt.Errorf("session secret: %w", err)
	}
	store = sessions.NewCookieStore([]byte(sessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	cliDBConfig = DBConfig{
		User:         *dbUserFlag,
		Pass:         *dbPassFlag,
		Host:         *dbHostFlag,
		Port:         *dbPortFlag,
		Name:         *dbNameFlag,
		LogVerbosity: *dbLogVerbosityFlag,
	}
	dbConfigFile = *dbCfgPath
	if dbConfigFile == "" {
		if v, ok := appCfg["DB_CONFIG_FILE"]; ok {
			dbConfigFile = v
		}
	}

	cliEmailConfig = EmailConfig{
		Provider:     *emailProviderFlag,
		SMTPHost:     *smtpHostFlag,
		SMTPPort:     *smtpPortFlag,
		SMTPUser:     *smtpUserFlag,
		SMTPPass:     *smtpPassFlag,
		AWSRegion:    *awsRegionFlag,
		JMAPEndpoint: *jmapEndpointFlag,
		JMAPAccount:  *jmapAccountFlag,
		JMAPIdentity: *jmapIdentityFlag,
		JMAPUser:     *jmapUserFlag,
		JMAPPass:     *jmapPassFlag,
		SendGridKey:  *sendGridKeyFlag,
	}
	emailConfigFile = *emailCfgPath
	if emailConfigFile == "" {
		if v, ok := appCfg["EMAIL_CONFIG_FILE"]; ok {
			emailConfigFile = v
		}
	}

	if listenFlagSet {
		cliHTTPConfig.Listen = *listenFlag
	}
	if hostnameFlagSet {
		cliHTTPConfig.Hostname = *hostnameFlag
	}
	httpConfigFile = *httpCfgPath
	if httpConfigFile == "" {
		if v, ok := appCfg["HTTP_CONFIG_FILE"]; ok {
			httpConfigFile = v
		}
	}

	cliPaginationConfig = PaginationConfig{
		Min:     *pageSizeMinFlag,
		Max:     *pageSizeMaxFlag,
		Default: *pageSizeDefaultFlag,
	}
	paginationConfigFile = *paginationCfgPath
	if paginationConfigFile == "" {
		if v, ok := appCfg["PAGINATION_CONFIG_FILE"]; ok {
			paginationConfigFile = v
		}
	}

	httpCfg := loadHTTPConfig()

	var handler http.Handler

	var cliFeeds string
	if feedsFlagSet {
		cliFeeds = *feedsEnabledFlag
	}
	loadFeedsEnabled(cliFeeds, appCfg)
	loadStatsStartYear(*statsStartYearFlag, appCfg)

	dbCfg := loadDBConfig()
	emailCfg := loadEmailConfig()

	if err := performStartupChecks(dbCfg); err != nil {
		return fmt.Errorf("startup checks: %w", err)
	}

	if dbPool != nil {
		defer func() {
			if err := dbPool.Close(); err != nil {
				log.Printf("DB close error: %v", err)
			}
		}()
	}

	r := mux.NewRouter()

	// TODO consider adsense / adwords / etc

	r.HandleFunc("/main.css", mainCSSHandler).Methods("GET")

	// News
	r.Handle("/", AddNewsIndex(http.HandlerFunc(runTemplate("newsPage.gohtml")))).Methods("GET")
	r.HandleFunc("/", taskDoneAutoRefreshPage).Methods("POST")
	nr := r.PathPrefix("/news").Subrouter()
	nr.Use(AddNewsIndex)
	nr.HandleFunc(".rss", newsRssPage).Methods("GET")
	nr.HandleFunc("", runTemplate("newsPage.gohtml")).Methods("GET")
	nr.HandleFunc("", taskDoneAutoRefreshPage).Methods("POST")
	//TODO nr.HandleFunc("/news/{id:[0-9]+}", newsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", newsPostPage).Methods("GET")
	nr.HandleFunc("/news/{post}", newsPostReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskReply))
	nr.HandleFunc("/news/{post}", newsPostEditActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskEdit))
	nr.HandleFunc("/news/{post}", newsPostNewActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskNewPost))
	nr.HandleFunc("/news/{post}/announcement", newsAnnouncementActivateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskAdd))
	nr.HandleFunc("/news/{post}/announcement", newsAnnouncementDeactivateActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskDelete))
	nr.HandleFunc("/news/{post}", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	nr.HandleFunc("/news/{post}", taskDoneAutoRefreshPage).Methods("POST")
	nr.HandleFunc("/user/permissions", newsUserPermissionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	nr.HandleFunc("/users/permissions", newsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Allow"))
	nr.HandleFunc("/users/permissions", newsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher("User Disallow"))

	faqr := r.PathPrefix("/faq").Subrouter()
	faqr.HandleFunc("", faqPage).Methods("GET", "POST")
	faqr.HandleFunc("/ask", faqAskPage).Methods("GET")
	faqr.HandleFunc("/ask", faqAskActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAsk))

	br := r.PathPrefix("/blogs").Subrouter()
	br.HandleFunc("/rss", blogsRssPage).Methods("GET")
	br.HandleFunc("/atom", blogsAtomPage).Methods("GET")
	br.HandleFunc("", blogsPage).Methods("GET")
	br.HandleFunc("/", blogsPage).Methods("GET")
	br.HandleFunc("/add", blogsBlogAddPage).Methods("GET").MatcherFunc(RequiredAccess("writer", "administrator"))
	br.HandleFunc("/add", blogsBlogAddActionPage).Methods("POST").MatcherFunc(RequiredAccess("writer", "administrator")).MatcherFunc(TaskMatcher(TaskAdd))
	br.HandleFunc("/bloggers", blogsBloggersPage).Methods("GET")
	br.HandleFunc("/blogger/{username}", blogsBloggerPage).Methods("GET")
	br.HandleFunc("/blogger/{username}/", blogsBloggerPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", blogsBlogPage).Methods("GET")
	br.HandleFunc("/blog/{blog}", taskDoneAutoRefreshPage).Methods("POST")
	br.HandleFunc("/blog/{blog}/comments", blogsCommentPage).Methods("GET", "POST")
	br.HandleFunc("/blog/{blog}/reply", blogsBlogReplyPostPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogsCommentEditPostPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher(TaskEditReply))
	br.HandleFunc("/blog/{blog}/comment/{comment}", blogsCommentEditPostCancelPage).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor())).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	br.HandleFunc("/blog/{blog}/edit", blogsBlogEditPage).Methods("GET").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor())))
	br.HandleFunc("/blog/{blog}/edit", blogsBlogEditActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("administrator"), And(RequiredAccess("writer"), BlogAuthor()))).MatcherFunc(TaskMatcher(TaskEdit))
	br.HandleFunc("/blog/{blog}/edit", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))

	// TODO a matcher check to ensure topics and threads align.
	fr := r.PathPrefix("/forum").Subrouter()
	fr.HandleFunc("/topic/{topic}.rss", forumTopicRssPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}.atom", forumTopicAtomPage).Methods("GET")
	fr.HandleFunc("", forumPage).Methods("GET")
	fr.HandleFunc("/category/{category}", forumPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}", forumTopicsPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewPage).Methods("GET")
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateThread))
	fr.HandleFunc("/topic/{topic}/thread", forumThreadNewCancelPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}", forumThreadPage).Methods("GET").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(GetThreadAndTopic())
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumTopicThreadReplyPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskReply))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/reply", forumTopicThreadReplyCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskCancel))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forumTopicThreadCommentEditActionPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskEditReply)).MatcherFunc(Or(RequiredAccess("administrator"), CommentAuthor()))
	fr.HandleFunc("/topic/{topic}/thread/{thread}/comment/{comment}", forumTopicThreadCommentEditActionCancelPage).Methods("POST").MatcherFunc(GetThreadAndTopic()).MatcherFunc(TaskMatcher(TaskCancel))

	lr := r.PathPrefix("/linker").Subrouter()
	lr.HandleFunc("/rss", linkerRssPage).Methods("GET")
	lr.HandleFunc("/atom", linkerAtomPage).Methods("GET")
	lr.HandleFunc("", linkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}", linkerLinkerPage).Methods("GET")
	lr.HandleFunc("/linker/{username}/", linkerLinkerPage).Methods("GET")
	lr.HandleFunc("/categories", linkerCategoriesPage).Methods("GET")
	lr.HandleFunc("/category/{category}", linkerCategoryPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linkerCommentsPage).Methods("GET")
	lr.HandleFunc("/comments/{link}", linkerCommentsReplyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	lr.HandleFunc("/show/{link}", linkerShowPage).Methods("GET")
	lr.HandleFunc("/show/{link}", linkerShowReplyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	lr.HandleFunc("/suggest", linkerSuggestPage).Methods("GET")
	lr.HandleFunc("/suggest", linkerSuggestActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSuggest))

	bmr := r.PathPrefix("/bookmarks").Subrouter()
	bmr.HandleFunc("", bookmarksPage).Methods("GET")
	bmr.HandleFunc("/mine", bookmarksMinePage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	bmr.HandleFunc("/edit", bookmarksEditSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSave))
	bmr.HandleFunc("/edit", bookmarksEditCreateActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskCreate))
	bmr.HandleFunc("/edit", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiresAnAccount())

	ibr := r.PathPrefix("/imagebbs").Subrouter()
	ibr.HandleFunc(".rss", imagebbsRssPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.rss", imagebbsBoardRssPage).Methods("GET")
	ibr.HandleFunc(".atom", imagebbsAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno:[0-9]+}.atom", imagebbsBoardAtomPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbsBoardPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}", imagebbsBoardPostImageActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskAddOffsiteImage))
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbsBoardThreadPage).Methods("GET")
	ibr.HandleFunc("/board/{boardno}/thread/{thread}", imagebbsBoardThreadReplyActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskReply))
	ibr.HandleFunc("", imagebbsPage).Methods("GET")
	ibr.HandleFunc("/", imagebbsPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}", imagebbsPosterPage).Methods("GET")
	ibr.HandleFunc("/poster/{username}/", imagebbsPosterPage).Methods("GET")

	sr := r.PathPrefix("/search").Subrouter()
	sr.HandleFunc("", searchPage).Methods("GET")
	sr.HandleFunc("", searchResultForumActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchForum))
	sr.HandleFunc("", searchResultNewsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchNews))
	sr.HandleFunc("", searchResultLinkerActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchLinker))
	sr.HandleFunc("", searchResultBlogsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchBlogs))
	sr.HandleFunc("", searchResultWritingsActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskSearchWritings))

	r.PathPrefix("/writing").HandlerFunc(redirectPermanentPrefix("/writing", "/writings"))
	r.PathPrefix("/links").HandlerFunc(redirectPermanentPrefix("/links", "/linker"))

	wr := r.PathPrefix("/writings").Subrouter()
	wr.HandleFunc("/rss", writingsRssPage).Methods("GET")
	wr.HandleFunc("/atom", writingsAtomPage).Methods("GET")
	wr.HandleFunc("", writingsPage).Methods("GET")
	wr.HandleFunc("/", writingsPage).Methods("GET")
	wr.HandleFunc("/writer/{username}", writingsWriterPage).Methods("GET")
	wr.HandleFunc("/writer/{username}/", writingsWriterPage).Methods("GET")
	wr.HandleFunc("/user/permissions", writingsUserPermissionsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
	wr.HandleFunc("/users/permissions", writingsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserAllow))
	wr.HandleFunc("/users/permissions", writingsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserDisallow))
	wr.HandleFunc("/article/{article}", writingsArticlePage).Methods("GET")
	wr.HandleFunc("/article/{article}", writingsArticleReplyActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskReply))
	wr.HandleFunc("/article/{article}/edit", writingsArticleEditPage).Methods("GET").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator")))
	wr.HandleFunc("/article/{article}/edit", writingsArticleEditActionPage).Methods("POST").MatcherFunc(Or(And(RequiredAccess("writer"), WritingAuthor()), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher(TaskUpdateWriting))
	wr.HandleFunc("/categories", writingsCategoriesPage).Methods("GET")
	wr.HandleFunc("/categories", writingsCategoriesPage).Methods("GET")
	wr.HandleFunc("/category/{category}", writingsCategoryPage).Methods("GET")
	wr.HandleFunc("/category/{category}/add", writingsArticleAddPage).Methods("GET").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator")))
	wr.HandleFunc("/category/{category}/add", writingsArticleAddActionPage).Methods("POST").MatcherFunc(Or(RequiredAccess("writer"), RequiredAccess("administrator"))).MatcherFunc(TaskMatcher(TaskSubmitWriting))

	ir := r.PathPrefix("/information").Subrouter()
	ir.HandleFunc("", informationPage).Methods("GET")

	ur := r.PathPrefix("/usr").Subrouter()
	ur.HandleFunc("", userPage).Methods("GET")
	ur.HandleFunc("/logout", userLogoutPage).Methods("GET")
	ur.HandleFunc("/lang", userLangPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/lang", userLangSaveLanguagesActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveLanguages))
	ur.HandleFunc("/lang", userLangSaveLanguagePreferenceActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveLanguage))
	ur.HandleFunc("/lang", userLangSaveAllActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/email", userEmailSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/email", userEmailTestActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskTestMail))
	ur.HandleFunc("/paging", userPagingPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/paging", userPagingSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/page-size", userPageSizePage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/page-size", userPageSizeSaveActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskSaveAll))
	ur.HandleFunc("/notifications", userNotificationsPage).Methods("GET").MatcherFunc(RequiresAnAccount())
	ur.HandleFunc("/notifications/dismiss", userNotificationsDismissActionPage).Methods("POST").MatcherFunc(RequiresAnAccount()).MatcherFunc(TaskMatcher(TaskDismiss))
	ur.HandleFunc("/notifications/rss", notificationsRssPage).Methods("GET").MatcherFunc(RequiresAnAccount())

	// Redirect legacy paths to the updated usr endpoints.
	r.HandleFunc("/user/lang", redirectPermanent("/usr/lang"))
	r.HandleFunc("/user/email", redirectPermanent("/usr/email"))

	rr := r.PathPrefix("/register").Subrouter()
	rr.HandleFunc("", registerPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	rr.HandleFunc("", registerActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(TaskMatcher(TaskRegister))

	ulr := r.PathPrefix("/login").Subrouter()
	ulr.HandleFunc("", loginUserPassPage).Methods("GET").MatcherFunc(Not(RequiresAnAccount()))
	ulr.HandleFunc("", loginActionPage).Methods("POST").MatcherFunc(Not(RequiresAnAccount())).MatcherFunc(TaskMatcher(TaskLogin))

	ar := r.PathPrefix("/admin").Subrouter()
	ar.Use(AdminCheckerMiddleware)
	ar.HandleFunc("", adminPage).Methods("GET")
	ar.HandleFunc("/", adminPage).Methods("GET")
	ar.HandleFunc("/forum", adminForumPage).Methods("GET")
	ar.HandleFunc("/forum", adminForumRemakeForumThreadPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeStatisticInformationOnForumthread))
	ar.HandleFunc("/forum", adminForumRemakeForumTopicPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeStatisticInformationOnForumtopic))
	ar.HandleFunc("/forum/flagged", adminForumFlaggedPostsPage).Methods("GET")
	ar.HandleFunc("/forum/logs", adminForumModeratorLogsPage).Methods("GET")
	ar.HandleFunc("/forum/list", adminForumWordListPage).Methods("GET")
	ar.HandleFunc("/forum/flagged", adminForumFlaggedPostsPage).Methods("GET")
	ar.HandleFunc("/forum/modlog", adminForumModeratorLogsPage).Methods("GET")
	ar.HandleFunc("/users", adminUsersPage).Methods("GET")
	ar.HandleFunc("/users/disable", adminUserDisablePage).Methods("POST")
	ar.HandleFunc("/users/reset", adminUserResetPasswordPage).Methods("POST")
	ar.HandleFunc("/users/edit", adminUserEditFormPage).Methods("GET")
	ar.HandleFunc("/users/edit", adminUserEditSavePage).Methods("POST")
	ar.HandleFunc("/users/permissions", adminUsersPermissionsPage).Methods("GET")
	ar.HandleFunc("/users/permissions", adminUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	ar.HandleFunc("/users/permissions", adminUsersPermissionsUpdatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdatePermission))
	ar.HandleFunc("/languages", adminLanguagesPage).Methods("GET")
	ar.HandleFunc("/language", adminLanguageRedirect).Methods("GET")
	ar.HandleFunc("/languages", adminLanguagesRenamePage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameLanguage))
	ar.HandleFunc("/languages", adminLanguagesDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteLanguage))
	ar.HandleFunc("/languages", adminLanguagesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateLanguage))
	ar.HandleFunc("/categories", adminCategoriesPage).Methods("GET")
	ar.HandleFunc("/permissions/sections", adminPermissionsSectionPage).Methods("GET")
	ar.HandleFunc("/permissions/sections/view", adminPermissionsSectionViewPage).Methods("GET")
	ar.HandleFunc("/permissions/sections", adminPermissionsSectionRenamePage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameSection))
	ar.HandleFunc("/email/queue", adminEmailQueuePage).Methods("GET")
	ar.HandleFunc("/email/queue", adminEmailQueueResendActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskResend))
	ar.HandleFunc("/email/queue", adminEmailQueueDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	ar.HandleFunc("/notifications", adminNotificationsPage).Methods("GET")
	ar.HandleFunc("/announcements", adminAnnouncementsPage).Methods("GET")
	ar.HandleFunc("/announcements", adminAnnouncementsAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAdd))
	ar.HandleFunc("/announcements", adminAnnouncementsDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	ar.HandleFunc("/sessions", adminSessionsPage).Methods("GET")
	ar.HandleFunc("/sessions/delete", adminSessionsDeletePage).Methods("POST")
	ar.HandleFunc("/login/attempts", adminLoginAttemptsPage).Methods("GET")
	ar.HandleFunc("/settings", adminSiteSettingsPage).Methods("GET", "POST")
	ar.HandleFunc("/stats", adminServerStatsPage).Methods("GET")
	ar.HandleFunc("/usage", adminUsageStatsPage).Methods("GET")
	ar.HandleFunc("/search", adminSearchPage).Methods("GET")
	ar.HandleFunc("/search", adminSearchRemakeCommentsSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeCommentsSearch))
	ar.HandleFunc("/search", adminSearchRemakeNewsSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeNewsSearch))
	ar.HandleFunc("/search", adminSearchRemakeBlogSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeBlogSearch))
	ar.HandleFunc("/search", adminSearchRemakeLinkerSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeLinkerSearch))
	ar.HandleFunc("/search", adminSearchRemakeWritingSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeWritingSearch))
	ar.HandleFunc("/search", adminSearchRemakeImageSearchPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemakeImageSearch))
	ar.HandleFunc("/search/list", adminSearchWordListPage).Methods("GET")
	ar.HandleFunc("/search/list.txt", adminSearchWordListDownloadPage).Methods("GET")
	ar.HandleFunc("/linker/categories", linkerAdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/linker/categories", linkerAdminCategoriesUpdatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdate))
	ar.HandleFunc("/linker/categories", linkerAdminCategoriesRenamePage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameCategory))
	ar.HandleFunc("/linker/categories", linkerAdminCategoriesDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteCategory))
	ar.HandleFunc("/linker/categories", linkerAdminCategoriesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateCategory))
	ar.HandleFunc("/linker/add", linkerAdminAddPage).Methods("GET")
	ar.HandleFunc("/linker/add", linkerAdminAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAdd))
	ar.HandleFunc("/linker/queue", linkerAdminQueuePage).Methods("GET")
	ar.HandleFunc("/linker/queue", linkerAdminQueueDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDelete))
	ar.HandleFunc("/linker/queue", linkerAdminQueueApproveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskApprove))
	ar.HandleFunc("/linker/queue", linkerAdminQueueUpdateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdate))
	ar.HandleFunc("/linker/queue", linkerAdminQueueBulkApproveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskBulkApprove))
	ar.HandleFunc("/linker/queue", linkerAdminQueueBulkDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskBulkDelete))
	ar.HandleFunc("/linker/users/levels", linkerAdminUserLevelsPage).Methods("GET")
	ar.HandleFunc("/linker/users/levels", linkerAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	ar.HandleFunc("/linker/users/levels", linkerAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	ar.HandleFunc("/faq/answer", faqAdminAnswerPage).Methods("GET", "POST").MatcherFunc(NoTask())
	ar.HandleFunc("/faq/answer", faqAnswerAnswerActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAnswer))
	ar.HandleFunc("/faq/answer", faqAnswerRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemoveRemove))
	ar.HandleFunc("/faq/categories", faqAdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/faq/categories", faqCategoriesRenameActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRenameCategory))
	ar.HandleFunc("/faq/categories", faqCategoriesDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteCategory))
	ar.HandleFunc("/faq/categories", faqCategoriesCreateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreateCategory))
	ar.HandleFunc("/faq/questions", faqAdminQuestionsPage).Methods("GET", "POST").MatcherFunc(NoTask())
	ar.HandleFunc("/faq/questions", faqQuestionsEditActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskEdit))
	ar.HandleFunc("/faq/questions", faqQuestionsDeleteActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemoveRemove))
	ar.HandleFunc("/faq/questions", faqQuestionsCreateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCreate))
	ar.HandleFunc("/blogs/user/permissions", getPermissionsByUserIdAndSectionBlogsPage).Methods("GET")
	ar.HandleFunc("/blogs/users/permissions", blogsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	ar.HandleFunc("/blogs/users/permissions", blogsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	ar.HandleFunc("/blogs/users/permissions", blogsUsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUsersAllow))
	ar.HandleFunc("/blogs/users/permissions", blogsUsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUsersDisallow))
	ar.HandleFunc("/imagebbs", imagebbsAdminPage).Methods("GET")
	ar.HandleFunc("/imagebbs/boards", imagebbsAdminBoardsPage).Methods("GET")
	ar.HandleFunc("/imagebbs/boards", taskDoneAutoRefreshPage).Methods("POST")
	ar.HandleFunc("/imagebbs/board", imagebbsAdminNewBoardPage).Methods("GET")
	ar.HandleFunc("/imagebbs/board", imagebbsAdminNewBoardMakePage).Methods("POST").MatcherFunc(TaskMatcher(TaskNewBoard))
	ar.HandleFunc("/imagebbs/board", taskDoneAutoRefreshPage).Methods("POST")
	ar.HandleFunc("/imagebbs/board/{board}", imagebbsAdminBoardModifyBoardActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskModifyBoard))
	ar.HandleFunc("/forum/categories", forumAdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/forum/categories", taskDoneAutoRefreshPage).Methods("POST")
	ar.HandleFunc("/forum/category/{category}", forumAdminCategoryEditPage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumCategoryChange))
	ar.HandleFunc("/forum/category", forumAdminCategoryCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumCategoryCreate))
	ar.HandleFunc("/forum/category/delete", forumAdminCategoryDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteCategory))
	ar.HandleFunc("/forum/topics", forumAdminTopicsPage).Methods("GET")
	ar.HandleFunc("/forum/topics", taskDoneAutoRefreshPage).Methods("POST")
	ar.HandleFunc("/forum/conversations", forumAdminThreadsPage).Methods("GET")
	ar.HandleFunc("/forum/thread/{thread}/delete", forumAdminThreadDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumThreadDelete))
	ar.HandleFunc("/forum/topic/{topic}/edit", forumAdminTopicEditPage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicChange))
	ar.HandleFunc("/forum/topic/{topic}/delete", forumAdminTopicDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicDelete))
	ar.HandleFunc("/forum/topic", forumTopicCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskForumTopicCreate))
	ar.HandleFunc("/forum/topic/{topic}/levels", forumAdminTopicRestrictionLevelPage).Methods("GET")
	ar.HandleFunc("/forum/topic/{topic}/levels", forumAdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateTopicRestriction))
	ar.HandleFunc("/forum/topic/{topic}/levels", forumAdminTopicRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskSetTopicRestriction))
	ar.HandleFunc("/forum/topic/{topic}/levels", forumAdminTopicRestrictionLevelDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteTopicRestriction))
	ar.HandleFunc("/forum/topic/{topic}/levels", forumAdminTopicRestrictionLevelCopyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCopyTopicRestriction))
	ar.HandleFunc("/forum/users", forumAdminUserPage).Methods("GET")
	ar.HandleFunc("/forum/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskSetUserLevel))
	ar.HandleFunc("/forum/user/{user}/levels", forumAdminUserLevelUpdatePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskUpdateUserLevel))
	ar.HandleFunc("/forum/user/{user}/levels", forumAdminUserLevelDeletePage).Methods("GET", "POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher(TaskDeleteUserLevel))
	ar.HandleFunc("/forum/user/{user}/levels", forumAdminUserLevelPage).Methods("GET")
	ar.HandleFunc("/forum/restrictions/users", forumAdminUsersRestrictionsDeletePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel())).MatcherFunc(TaskMatcher(TaskDeleteUserLevel))
	ar.HandleFunc("/forum/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskUpdateUserLevel))
	ar.HandleFunc("/forum/restrictions/users", forumAdminUsersRestrictionsUpdatePage).Methods("POST").MatcherFunc(And(AdminUsersMaxLevelNotLowerThanTargetLevel(), TargetUsersLevelNotHigherThanAdminsMax())).MatcherFunc(TaskMatcher(TaskSetUserLevel))
	ar.HandleFunc("/forum/restrictions/users", forumAdminUsersRestrictionsPage).Methods("GET")
	ar.HandleFunc("/forum/restrictions/topics", forumAdminTopicsRestrictionLevelPage).Methods("GET")
	ar.HandleFunc("/forum/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateTopicRestriction))
	ar.HandleFunc("/forum/restrictions/topics", forumAdminTopicsRestrictionLevelDeletePage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteTopicRestriction))
	ar.HandleFunc("/forum/restrictions/topics", forumAdminTopicsRestrictionLevelChangePage).Methods("POST").MatcherFunc(TaskMatcher(TaskSetTopicRestriction))
	ar.HandleFunc("/forum/restrictions/topics", forumAdminTopicsRestrictionLevelCopyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskCopyTopicRestriction))
	ar.HandleFunc("/writings/user/permissions", writingsUserPermissionsPage).Methods("GET")
	ar.HandleFunc("/writings/users/permissions", writingsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	ar.HandleFunc("/writings/users/permissions", writingsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	ar.HandleFunc("/writings/users/levels", writingsAdminUserLevelsPage).Methods("GET")
	ar.HandleFunc("/writings/users/levels", writingsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	ar.HandleFunc("/writings/users/levels", writingsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	ar.HandleFunc("/writings/users/access", writingsAdminUserAccessPage).Methods("GET")
	ar.HandleFunc("/writings/users/access", writingsAdminUserAccessAddActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAddApproval))
	ar.HandleFunc("/writings/users/access", writingsAdminUserAccessUpdateActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUpdateUserApproval))
	ar.HandleFunc("/writings/users/access", writingsAdminUserAccessRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskDeleteUserApproval))
	ar.HandleFunc("/writings/category/{category}/permissions", writingsCategoryPermissionsPage).Methods("GET")
	ar.HandleFunc("/writings/category/{category}/permissions", writingsCategoryPermissionsAllowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserAllow))
	ar.HandleFunc("/writings/category/{category}/permissions/delete", writingsCategoryPermissionsDisallowPage).Methods("POST").MatcherFunc(TaskMatcher(TaskUserDisallow))
	ar.HandleFunc("/writings/categories", writingsAdminCategoriesPage).Methods("GET")
	ar.HandleFunc("/writings/categories", writingsAdminCategoriesModifyPage).Methods("POST").MatcherFunc(TaskMatcher(TaskWritingCategoryChange))
	ar.HandleFunc("/writings/categories", writingsAdminCategoriesCreatePage).Methods("POST").MatcherFunc(TaskMatcher(TaskWritingCategoryCreate))
	ar.HandleFunc("/news/users/levels", newsAdminUserLevelsPage).Methods("GET")
	ar.HandleFunc("/news/users/levels", newsAdminUserLevelsAllowActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskAllow))
	ar.HandleFunc("/news/users/levels", newsAdminUserLevelsRemoveActionPage).Methods("POST").MatcherFunc(TaskMatcher(TaskRemoveLower))
	ar.HandleFunc("/reload", adminReloadConfigPage).Methods("POST")
	ar.HandleFunc("/shutdown", adminShutdownPage).Methods("POST")

	/* TODO (agent) reintegrate these admin pages
	  	br.HandleFunc("/user/permissions", getPermissionsByUserIdAndSectionBlogsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
		br.HandleFunc("/users/permissions", blogsUsersPermissionsPermissionUserAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserAllow))
		br.HandleFunc("/users/permissions", blogsUsersPermissionsDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUserDisallow))
		br.HandleFunc("/users/permissions", blogsUsersPermissionsBulkAllowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUsersAllow))
		br.HandleFunc("/users/permissions", blogsUsersPermissionsBulkDisallowPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskUsersDisallow))
		ibr.HandleFunc("/admin", imagebbsAdminPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
		ibr.HandleFunc("/admin/boards", imagebbsAdminBoardsPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
		ibr.HandleFunc("/admin/boards", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
		ibr.HandleFunc("/admin/board", imagebbsAdminNewBoardPage).Methods("GET").MatcherFunc(RequiredAccess("administrator"))
		ibr.HandleFunc("/admin/board", imagebbsAdminNewBoardMakePage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskNewBoard))
		ibr.HandleFunc("/admin/board", taskDoneAutoRefreshPage).Methods("POST").MatcherFunc(RequiredAccess("administrator"))
		ibr.HandleFunc("/admin/board/{board}", imagebbsAdminBoardModifyBoardActionPage).Methods("POST").MatcherFunc(RequiredAccess("administrator")).MatcherFunc(TaskMatcher(TaskModifyBoard))
	*/

	// oauth shit
	//r.HandleFunc("/login", loginPage)
	//r.HandleFunc("/callback", callbackHandler)
	//r.HandleFunc("/logout", logoutHandler)

	handler = newMiddlewareChain(
		DBAdderMiddleware,
		UserAdderMiddleware,
		CoreAdderMiddleware,
		RequestLoggerMiddleware,
		SecurityHeadersMiddleware,
	).Wrap(r)
	if csrfEnabled() {
		handler = newCSRFMiddleware(sessionSecret, httpCfg, version).Wrap(handler)
	}

	srv = newServer(handler, store, dbPool, dbCfg, emailCfg)
	loadPaginationConfig()

	log.Printf("Getting email parser")
	provider := providerFromConfig(emailCfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startWorkers(ctx, dbPool, provider)

	log.Printf("Starting web server")
	if err := runServer(ctx, srv, httpCfg.Listen); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
}

func runTemplate(template string) func(http.ResponseWriter, *http.Request) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type Data struct {
			*CoreData
		}

		data := Data{
			CoreData: r.Context().Value(ContextValues("coreData")).(*CoreData),
		}

		CustomNewsIndex(data.CoreData, r)

		log.Printf("rendering template %s", template)

		if err := renderTemplate(w, r, template, data); err != nil {
			log.Printf("Template Error: %s", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})
}

func AddNewsIndex(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cd := r.Context().Value(ContextValues("coreData")).(*CoreData)
		CustomNewsIndex(cd, r)
		handler.ServeHTTP(w, r)
	})
}

// safeGo runs fn in a goroutine and terminates the program if a panic occurs.
func safeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("goroutine panic: %v", r)
				os.Exit(1)
			}
		}()
		fn()
	}()
}

// mainCSSHandler serves the site's stylesheet.
func mainCSSHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "main.css", time.Time{}, bytes.NewReader(getMainCSSData()))
}

// redirectPermanent returns a handler that redirects to the provided path using
// StatusPermanentRedirect to preserve the request method.
func redirectPermanent(to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to, http.StatusPermanentRedirect)
	}
}

// redirectPermanentPrefix redirects any path starting with the given prefix to
// the same path under a new prefix while preserving query parameters.
func redirectPermanentPrefix(from, to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest := strings.TrimPrefix(r.URL.Path, from)
		if !strings.HasPrefix(rest, "/") && rest != "" {
			rest = "/" + rest
		}
		target := to + rest
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, target, http.StatusPermanentRedirect)
	}
}

// TODO we could do better
func TargetUsersLevelNotHigherThanAdminsMax() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		session, err := GetSession(r)
		if err != nil {
			return false
		}
		adminUid, _ := session.Values["UID"].(int32)

		targetUid, err := strconv.Atoi(r.PostFormValue("uid"))
		if err != nil {
			return false
		}

		tid, err := strconv.Atoi(r.PostFormValue("tid"))
		if err != nil {
			return false
		}

		queries := r.Context().Value(ContextValues("queries")).(*Queries)

		targetUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), GetUsersTopicLevelByUserIdAndThreadIdParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(targetUid),
		})
		if err != nil {
			return false
		}

		adminUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), GetUsersTopicLevelByUserIdAndThreadIdParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(adminUid),
		})
		if err != nil {
			return false
		}

		return adminUser.Invitemax.Int32 >= targetUser.Level.Int32
	}
}

// TODO we could do better
func AdminUsersMaxLevelNotLowerThanTargetLevel() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		session, err := GetSession(r)
		if err != nil {
			return false
		}
		adminUid, _ := session.Values["UID"].(int32)

		inviteMax, err := strconv.Atoi(r.PostFormValue("inviteMax"))
		if err != nil {
			return false
		}
		level, err := strconv.Atoi(r.PostFormValue("level"))
		if err != nil {
			return false
		}
		tid, err := strconv.Atoi(r.PostFormValue("tid"))
		if err != nil {
			return false
		}
		queries := r.Context().Value(ContextValues("queries")).(*Queries)

		adminUser, err := queries.GetUsersTopicLevelByUserIdAndThreadId(r.Context(), GetUsersTopicLevelByUserIdAndThreadIdParams{
			ForumtopicIdforumtopic: int32(tid),
			UsersIdusers:           int32(adminUid),
		})
		if err != nil {
			return false
		}

		return int(adminUser.Invitemax.Int32) >= level && int(adminUser.Invitemax.Int32) >= inviteMax
	}
}

func RequiredAccess(accessLevels ...string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return roleAllowed(request, accessLevels...)
	}
}

func RequiresAnAccount() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)
		return uid != 0
	}
}

func NewsPostAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		newsPostId, _ := strconv.Atoi(vars["post"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetForumThreadIdByNewsPostId(request.Context(), int32(newsPostId))
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.Idusers.Int32 == uid
	}
}

func BlogAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		blogId, _ := strconv.Atoi(vars["blog"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetBlogEntryForUserById(request.Context(), int32(blogId))
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
			default:
				log.Printf("Error: %s", err)
				return false
			}
		}

		return row.UsersIdusers == uid
	}
}

func WritingAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		writingId, _ := strconv.Atoi(vars["writing"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetWritingByIdForUserDescendingByPublishedDate(request.Context(), GetWritingByIdForUserDescendingByPublishedDateParams{
			Userid:    uid,
			Idwriting: int32(writingId),
		})
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.UsersIdusers == uid
	}
}

func CommentAuthor() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(request)
		commentId, _ := strconv.Atoi(vars["comment"])
		queries := request.Context().Value(ContextValues("queries")).(*Queries)
		session, err := GetSession(request)
		if err != nil {
			return false
		}
		uid, _ := session.Values["UID"].(int32)

		row, err := queries.GetCommentByIdForUser(request.Context(), GetCommentByIdForUserParams{
			UsersIdusers: uid,
			Idcomments:   int32(commentId),
		})
		if err != nil {
			log.Printf("Error: %s", err)
			return false
		}

		return row.UsersIdusers == uid
	}
}

func GetThreadAndTopic() mux.MatcherFunc {
	return func(r *http.Request, match *mux.RouteMatch) bool {
		vars := mux.Vars(r)
		topicID, err := strconv.Atoi(vars["topic"])
		if err != nil {
			return false
		}
		threadID, err := strconv.Atoi(vars["thread"])
		if err != nil {
			return false
		}

		queries := r.Context().Value(ContextValues("queries")).(*Queries)

		session, _ := GetSession(r)
		var uid int32
		if session != nil {
			uid, _ = session.Values["UID"].(int32)
		}

		threadRow, err := queries.GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissions(r.Context(), GetThreadByIdForUserByIdWithLastPoserUserNameAndPermissionsParams{
			UsersIdusers:  uid,
			Idforumthread: int32(threadID),
		})
		if err != nil {
			log.Printf("GetThreadAndTopic thread: %v", err)
			return false
		}

		topicRow, err := queries.GetForumTopicByIdForUser(r.Context(), GetForumTopicByIdForUserParams{
			UsersIdusers: uid,
			Idforumtopic: threadRow.ForumtopicIdforumtopic,
		})
		if err != nil {
			log.Printf("GetThreadAndTopic topic: %v", err)
			return false
		}

		if int(topicRow.Idforumtopic) != topicID {
			return false
		}

		ctx := context.WithValue(r.Context(), ContextValues("thread"), threadRow)
		ctx = context.WithValue(ctx, ContextValues("topic"), topicRow)
		*r = *r.WithContext(ctx)
		return true
	}
}

func TaskMatcher(taskName string) mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return request.PostFormValue("task") == taskName
	}
}

func NoTask() mux.MatcherFunc {
	return func(request *http.Request, match *mux.RouteMatch) bool {
		return request.PostFormValue("task") == ""
	}
}

//func oauthHomeHandler(w http.ResponseWriter, r *http.Request) {
//	// Check if user is authenticated
//	if !isAuthenticated(r) {
//		http.Redirect(w, r, "/login", http.StatusFound)
//		return
//	}
//
//	tmpl := `
//		<!DOCTYPE html>
//		<html>
//		<head>
//			<title>Home Page</title>
//		</head>
//		<body>
//			<h1>Welcome, {{ .Email }}</h1>
//			<a href="/logout">Logout</a>
//		</body>
//		</html>
//	`
//
//	t := template.Must(template.New("home").Parse(tmpl))
//	data := map[string]string{"Email": getEmail(r)}
//
//	t.Execute(w, data)
//}
//
//func loginHandler(w http.ResponseWriter, r *http.Request) {
//	// Generate the URL to redirect the user to Google's consent page
//	url := oauth2Config.AuthCodeURL("", oauth2.AccessTypeOffline)
//	http.Redirect(w, r, url, http.StatusFound)
//}
//
//func callbackHandler(w http.ResponseWriter, r *http.Request) {
//	code := r.FormValue("code")
//	if code == "" {
//		http.Error(w, "Failed to get authorization code", http.StatusInternalServerError)
//		return
//	}
//
//	// Exchange the authorization code for an access token
//	token, err := oauth2Config.Exchange(context.Background(), code)
//	if err != nil {
//		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
//		return
//	}
//
//	// Store the access token in the session
//	session, _ := store.Get(r, sessionName)
//	session.Values[sessionKey] = token.AccessToken
//	session.Save(r, w)
//
//	http.Redirect(w, r, "/", http.StatusFound)
//}
//
//func logoutHandler(w http.ResponseWriter, r *http.Request) {
//	// Clear the session and log the user out
//	session, _ := store.Get(r, sessionName)
//	session.Values[sessionKey] = nil
//	session.Save(r, w)
//
//	http.Redirect(w, r, "/", http.StatusFound)
//}
//
//func isAuthenticated(r *http.Request) bool {
//	session, _ := store.Get(r, sessionName)
//	accessToken, ok := session.Values[sessionKey]
//	if !ok {
//		return false
//	}
//
//	return accessToken != nil
//}
//
//func getEmail(r *http.Request) string {
//	// Fetch user's email using the access token from the session
//	session, _ := store.Get(r, sessionName)
//	_, ok := session.Values[sessionKey]
//	if !ok {
//		return ""
//	}
//
//	// Here, you can use the access token to fetch the user's email from the Google API
//	// For simplicity, we just return a dummy email
//	return "example@example.com"
//}
