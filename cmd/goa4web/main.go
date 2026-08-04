package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	_ "time/tzdata"

	"github.com/arran4/goa4web"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	authhandlers "github.com/arran4/goa4web/handlers/auth"
	bloghandlers "github.com/arran4/goa4web/handlers/blogs"
	bookmarkhandlers "github.com/arran4/goa4web/handlers/bookmarks"
	faqhandlers "github.com/arran4/goa4web/handlers/faq"
	forumhandlers "github.com/arran4/goa4web/handlers/forum"
	imagebbshandlers "github.com/arran4/goa4web/handlers/imagebbs"
	imagehandlers "github.com/arran4/goa4web/handlers/images"
	linkerhandlers "github.com/arran4/goa4web/handlers/linker"
	newshandlers "github.com/arran4/goa4web/handlers/news"
	privateforumhandlers "github.com/arran4/goa4web/handlers/privateforum"
	searchhandlers "github.com/arran4/goa4web/handlers/search"
	userhandlers "github.com/arran4/goa4web/handlers/user"
	writinghandlers "github.com/arran4/goa4web/handlers/writings"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/tasks"

	"github.com/arran4/goa4web/internal/dbdrivers"
	"github.com/arran4/goa4web/internal/dbdrivers/dbdefaults"
	"github.com/arran4/goa4web/internal/dlq"
	"github.com/arran4/goa4web/internal/dlq/dlqdefaults"
	"github.com/arran4/goa4web/internal/email"
	"github.com/arran4/goa4web/internal/email/emaildefaults"

	"github.com/arran4/goa4web/internal/router"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/db"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func registerTasks(reg *tasks.Registry, ah *adminhandlers.Handlers) {
	register := func(section string, ts []tasks.NamedTask) {
		for _, t := range ts {
			reg.Register(section, t)
		}
	}
	register("admin", ah.RegisterTasks())
	register("auth", authhandlers.RegisterTasks())
	register("blogs", bloghandlers.RegisterTasks())
	register("bookmarks", bookmarkhandlers.RegisterTasks())
	register("faq", faqhandlers.RegisterTasks())
	register("forum", forumhandlers.RegisterTasks())
	register("privateforum", privateforumhandlers.RegisterTasks())
	register("images", imagehandlers.RegisterTasks())
	register("imagebbs", imagebbshandlers.RegisterTasks())
	register("linker", linkerhandlers.RegisterTasks())
	register("news", newshandlers.RegisterTasks())
	register("search", searchhandlers.RegisterTasks())
	register("user", userhandlers.RegisterTasks())
	register("writing", writinghandlers.RegisterTasks())
}

func main() {
	goa4web.Version = version
	root, err := parseRoot(os.Args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Printf("%v", err)
		os.Exit(1)
	}
	defer root.Close()
	if err := root.Run(); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Printf("%v", err)
		os.Exit(1)
	}
}

// rootCmd is the top-level command state.
type rootCmd struct {
	fs               *flag.FlagSet
	cfg              *config.RuntimeConfig
	ConfigFile       string
	ConfigFileValues map[string]string
	db               *sql.DB
	querier          db.Querier
	Verbosity        int
	tasksReg         *tasks.Registry
	dbReg            *dbdrivers.Registry
	emailReg         *email.Registry
	dlqReg           *dlq.Registry
	routerReg        *router.Registry
	adminHandlers    *adminhandlers.Handlers
	ctx              context.Context
}

func (r *rootCmd) DB() (*sql.DB, error) {
	if r.db != nil {
		return r.db, nil
	}
	dbPool, ue := dbstart.InitDB(r.cfg, r.dbReg)
	if ue != nil {
		return nil, fmt.Errorf("rootCmd.DB: init: %w", ue.Err)
	}
	r.db = dbPool
	return r.db, nil
}

func (r *rootCmd) Querier() (db.Querier, error) {
	if r.querier != nil {
		return r.querier, nil
	}
	conn, err := r.DB()
	if err != nil {
		return nil, fmt.Errorf("rootCmd.Querier: %w", err)
	}
	return db.New(conn), nil
}

func (r *rootCmd) InitDB(cfg *config.RuntimeConfig) (*sql.DB, error) {
	if r.db != nil {
		return r.db, nil
	}
	dbPool, ue := dbstart.InitDB(cfg, r.dbReg)
	if ue != nil {
		return nil, fmt.Errorf("rootCmd.DB: init: %w", ue.Err)
	}
	r.db = dbPool
	return r.db, nil
}

func (r *rootCmd) Context() context.Context {
	return r.ctx
}

func (r *rootCmd) Close() {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}
}

func (r *rootCmd) Infof(format string, args ...any) {
	_ = log.Output(2, fmt.Sprintf(format, args...))
}

func (r *rootCmd) Verbosef(format string, args ...any) {
	if r.Verbosity > 0 {
		_ = log.Output(2, fmt.Sprintf(format, args...))
	}
}

func (r *rootCmd) RuntimeConfig() (*config.RuntimeConfig, error) {
	if r.cfg == nil {
		return nil, fmt.Errorf("runtime config not initialized")
	}
	return r.cfg, nil
}

func parseRoot(args []string) (*rootCmd, error) {
	r := &rootCmd{
		tasksReg:      tasks.NewRegistry(),
		dbReg:         dbdrivers.NewRegistry(),
		emailReg:      email.NewRegistry(),
		dlqReg:        dlq.NewRegistry(),
		routerReg:     router.NewRegistry(),
		adminHandlers: adminhandlers.New(),
		ctx:           context.Background(),
	}
	registerTasks(r.tasksReg, r.adminHandlers)
	registerModules(r.routerReg, r.adminHandlers)
	emaildefaults.Register(r.emailReg)
	dlqdefaults.RegisterDefaults(r.dlqReg, r.emailReg)
	dbdefaults.Register(r.dbReg)

	early := newFlagSet(args[0])
	early.Usage = func() {}

	var cfgPath string
	var showVersion bool

	early.StringVar(&cfgPath, "config-file", "", "path to config file")
	early.BoolVar(&showVersion, "version", false, "print version and exit")

	earlyErr := early.Parse(args[1:])
	wantHelp := errors.Is(earlyErr, flag.ErrHelp)
	rest := early.Args()

	if cfgPath == "" {
		cfgPath = os.Getenv(config.EnvConfigFile)
	}
	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	r.fs = config.NewRuntimeFlagSet(args[0])
	r.fs.StringVar(&cfgPath, "config-file", cfgPath, "path to config file")
	r.fs.IntVar(&r.Verbosity, "verbosity", 0, "verbosity level")
	r.fs.Usage = r.Usage

	if wantHelp && len(rest) == 0 {
		_ = r.fs.Parse([]string{"-h"})
		r.fs.Usage()
		return r, flag.ErrHelp
	}

	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, cfgPath)
	if err != nil {
		if errors.Is(err, config.ErrConfigFileNotFound) {
			return nil, fmt.Errorf("config file not found: %s", cfgPath)
		}
		return nil, fmt.Errorf("load config file: %w", err)
	}
	loadedConfigFile := cfgPath != ""

	if err := r.fs.Parse(rest); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			r.fs.Usage()
			return r, flag.ErrHelp
		}
		return nil, err
	}
	if loadedConfigFile {
		r.Verbosef("loaded config file %s", cfgPath)
	}

	r.ConfigFile = cfgPath
	r.ConfigFileValues = fileVals
	r.cfg = config.NewRuntimeConfig(
		config.WithFlagSet(r.fs),
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	)

	isTemplateCommand := false
	if len(r.fs.Args()) > 0 {
		switch r.fs.Arg(0) {
		case "serve", "templates":
			isTemplateCommand = true
		}
	}

	if r.cfg.TemplatesDir == "" {
		if isTemplateCommand {
			r.Infof("Embedded Template Mode")
		} else {
			r.Verbosef("Embedded Template Mode")
		}
	} else {
		if isTemplateCommand {
			r.Infof("Live Template Mode: %s", r.cfg.TemplatesDir)
		} else {
			r.Verbosef("Live Template Mode: %s", r.cfg.TemplatesDir)
		}
	}

	for _, name := range r.routerReg.Names() {
		r.Verbosef("Registered module: %s", name)
	}

	return r, nil
}

type runner interface {
	Run() error
}

var cmdParsers = map[string]func(*rootCmd, []string) (runner, error){
	"help":          func(r *rootCmd, args []string) (runner, error) { return parseHelpCmd(r, args) },
	"usage":         func(r *rootCmd, args []string) (runner, error) { return parseUsageCmd(r, args) },
	"serve":         func(r *rootCmd, args []string) (runner, error) { return parseServeCmd(r, args) },
	"user":          func(r *rootCmd, args []string) (runner, error) { return parseUserCmd(r, args) },
	"email":         func(r *rootCmd, args []string) (runner, error) { return parseEmailCmd(r, args) },
	"dlq":           func(r *rootCmd, args []string) (runner, error) { return parseDlqCmd(r, args) },
	"requests":      func(r *rootCmd, args []string) (runner, error) { return parseRequestsCmd(r, args) },
	"db":            func(r *rootCmd, args []string) (runner, error) { return parseDbCmd(r, args) },
	"perm":          func(r *rootCmd, args []string) (runner, error) { return parsePermCmd(r, args) },
	"role":          func(r *rootCmd, args []string) (runner, error) { return parseRoleCmd(r, args) },
	"subscription":  func(r *rootCmd, args []string) (runner, error) { return parseSubscriptionCmd(r, args) },
	"grant":         func(r *rootCmd, args []string) (runner, error) { return parseGrantCmd(r, args) },
	"board":         func(r *rootCmd, args []string) (runner, error) { return parseBoardCmd(r, args) },
	"blog":          func(r *rootCmd, args []string) (runner, error) { return parseBlogCmd(r, args) },
	"blogs":         func(r *rootCmd, args []string) (runner, error) { return parseBlogCmd(r, args) },
	"writing":       func(r *rootCmd, args []string) (runner, error) { return parseWritingCmd(r, args) },
	"news":          func(r *rootCmd, args []string) (runner, error) { return parseNewsCmd(r, args) },
	"announcement":  func(r *rootCmd, args []string) (runner, error) { return parseAnnouncementCmd(r, args) },
	"jmap":          func(r *rootCmd, args []string) (runner, error) { return parseJmapCmd(r, args) },
	"faq":           func(r *rootCmd, args []string) (runner, error) { return parseFaqCmd(r, args) },
	"forum":         func(r *rootCmd, args []string) (runner, error) { return parseForumCmd(r, args) },
	"private-forum": func(r *rootCmd, args []string) (runner, error) { return parsePrivateForumCmd(r, args) },
	"ipban":         func(r *rootCmd, args []string) (runner, error) { return parseIpBanCmd(r, args) },
	"images":        func(r *rootCmd, args []string) (runner, error) { return parseImagesCmd(r, args) },
	"files":         func(r *rootCmd, args []string) (runner, error) { return parseFilesCmd(r, args) },
	"imagebbs":      func(r *rootCmd, args []string) (runner, error) { return parseImagebbsCmd(r, args) },
	"links":         func(r *rootCmd, args []string) (runner, error) { return parseLinksCmd(r, args) },
	"share":         func(r *rootCmd, args []string) (runner, error) { return parseShareCmd(r, args) },
	"comment":       func(r *rootCmd, args []string) (runner, error) { return parseCommentCmd(r, args) },
	"comments":      func(r *rootCmd, args []string) (runner, error) { return parseCommentCmd(r, args) },
	"audit":         func(r *rootCmd, args []string) (runner, error) { return parseAuditCmd(r, args) },
	"notifications": func(r *rootCmd, args []string) (runner, error) { return parseNotificationsCmd(r, args) },
	"repl":          func(r *rootCmd, args []string) (runner, error) { return parseReplCmd(r, args) },
	"lang":          func(r *rootCmd, args []string) (runner, error) { return parseLangCmd(r, args) },
	"maintenance":   func(r *rootCmd, args []string) (runner, error) { return parseMaintenanceCmd(r, args) },
	"server":        func(r *rootCmd, args []string) (runner, error) { return parseServerCmd(r, args) },
	"config":        func(r *rootCmd, args []string) (runner, error) { return parseConfigCmd(r, args) },
	"page-size":     func(r *rootCmd, args []string) (runner, error) { return parsePageSizeCmd(r, args) },
	"templates":     func(r *rootCmd, args []string) (runner, error) { return parseTemplatesCmd(r, args) },
	"test":          func(r *rootCmd, args []string) (runner, error) { return parseTestCmd(r, args) },
}


func (r *rootCmd) Run() error {
	args := r.fs.Args()
	if len(args) == 0 {
		r.fs.Usage()
		return fmt.Errorf("no command provided")
	}

	cmdName := args[0]

	// Special case for usage to handle optional arguments like 'usage <cmd>'
	if cmdName == "usage" && len(args) == 1 {
		c, err := parseHelpCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("rootCmd.Run: usage: %w", err)
		}
		return c.Run()
	}

	parser, ok := cmdParsers[cmdName]
	if !ok {
		r.fs.Usage()
		return fmt.Errorf("rootCmd.Run: unknown command %q", cmdName)
	}

	c, err := parser(r, args[1:])
	if err != nil {
		return fmt.Errorf("rootCmd.Run: %s: %w", cmdName, err)
	}

	return c.Run()
}

// Usage prints command usage information with examples.
func (r *rootCmd) Usage() {
	executeUsage(r.fs.Output(), "root_usage.txt", r)
}

func (r *rootCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: "Global Flags", Flags: flagInfos(r.fs)}}
}

func (r *rootCmd) Prog() string { return r.fs.Name() }

var _ usageData = (*rootCmd)(nil)
