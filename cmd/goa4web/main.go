package main

import (
	"database/sql"

	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

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
)

var version = "dev"

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
	fs            *flag.FlagSet
	cfg           *config.RuntimeConfig
	ConfigFile    string
	db            *sql.DB
	Verbosity     int
	tasksReg      *tasks.Registry
	dbReg         *dbdrivers.Registry
	emailReg      *email.Registry
	dlqReg        *dlq.Registry
	routerReg     *router.Registry
	adminHandlers *adminhandlers.Handlers
}

func (r *rootCmd) DB() (*sql.DB, error) {
	if r.db != nil {
		return r.db, nil
	}
	dbPool, ue := dbstart.InitDB(r.cfg, r.dbReg)
	if ue != nil {
		return nil, fmt.Errorf("init db: %w", ue.Err)
	}
	r.db = dbPool
	return r.db, nil
}

func (r *rootCmd) Close() {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}
}

func (r *rootCmd) Infof(format string, args ...any) {
	fmt.Printf(format, args...)
	if !strings.HasSuffix(format, "\n") {
		fmt.Println()
	}
}

func (r *rootCmd) Verbosef(format string, args ...any) {
	if r.Verbosity > 0 {
		fmt.Printf(format, args...)
		if !strings.HasSuffix(format, "\n") {
			fmt.Println()
		}
	}
}

func parseRoot(args []string) (*rootCmd, error) {
	r := &rootCmd{
		tasksReg:      tasks.NewRegistry(),
		dbReg:         dbdrivers.NewRegistry(),
		emailReg:      email.NewRegistry(),
		dlqReg:        dlq.NewRegistry(),
		routerReg:     router.NewRegistry(),
		adminHandlers: adminhandlers.New(),
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

	if err := r.fs.Parse(rest); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			r.fs.Usage()
			return r, flag.ErrHelp
		}
		return nil, err
	}

	r.ConfigFile = cfgPath
	r.cfg = config.NewRuntimeConfig(
		config.WithFlagSet(r.fs),
		config.WithFileValues(fileVals),
		config.WithGetenv(os.Getenv),
	)
	return r, nil
}

func (r *rootCmd) Run() error {
	args := r.fs.Args()
	if len(args) == 0 {
		r.fs.Usage()
		return fmt.Errorf("no command provided")
	}
	switch args[0] {
	case "help", "usage":
		c, err := parseHelpCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("help: %w", err)
		}
		return c.Run()
	case "serve":
		c, err := parseServeCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("serve: %w", err)
		}
		return c.Run()
	case "user":
		c, err := parseUserCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("user: %w", err)
		}
		return c.Run()
	case "email":
		c, err := parseEmailCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("email: %w", err)
		}
		return c.Run()
	case "db":
		c, err := parseDbCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
		return c.Run()
	case "perm":
		c, err := parsePermCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("perm: %w", err)
		}
		return c.Run()
	case "role":
		c, err := parseRoleCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("role: %w", err)
		}
		return c.Run()
	case "grant":
		c, err := parseGrantCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("grant: %w", err)
		}
		return c.Run()
	case "board":
		c, err := parseBoardCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("board: %w", err)
		}
		return c.Run()
	case "blog", "blogs":
		c, err := parseBlogCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("blog: %w", err)
		}
		return c.Run()
	case "writing":
		c, err := parseWritingCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("writing: %w", err)
		}
		return c.Run()
	case "news":
		c, err := parseNewsCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("news: %w", err)
		}
		return c.Run()
	case "faq":
		c, err := parseFaqCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("faq: %w", err)
		}
		return c.Run()
	case "ipban":
		c, err := parseIpBanCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("ipban: %w", err)
		}
		return c.Run()
	case "images":
		c, err := parseImagesCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("images: %w", err)
		}
		return c.Run()
	case "links":
		c, err := parseLinksCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("links: %w", err)
		}
		return c.Run()
	case "comment", "comments":
		c, err := parseCommentCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("comment: %w", err)
		}
		return c.Run()
	case "audit":
		c, err := parseAuditCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("audit: %w", err)
		}
		return c.Run()
	case "notifications":
		c, err := parseNotificationsCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("notifications: %w", err)
		}
		return c.Run()
	case "repl":
		c, err := parseReplCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("repl: %w", err)
		}
		return c.Run()
	case "lang":
		c, err := parseLangCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("lang: %w", err)
		}
		return c.Run()
	case "server":
		c, err := parseServerCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("server: %w", err)
		}
		return c.Run()
	case "config":
		c, err := parseConfigCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		return c.Run()
	case "test":
		c, err := parseTestCmd(r, args[1:])
		if err != nil {
			return fmt.Errorf("test: %w", err)
		}
		return c.Run()
	default:
		r.fs.Usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
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
