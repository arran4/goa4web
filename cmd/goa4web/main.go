package main

import (
	"database/sql"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/arran4/goa4web/cmd/goa4web/dbhandlers/dbdefaults"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	dbstart "github.com/arran4/goa4web/internal/dbstart"
	dlqreg "github.com/arran4/goa4web/internal/dlq/dlqdefaults"
	"github.com/arran4/goa4web/runtimeconfig"
)

//go:embed templates/root_usage.txt
var rootUsageTemplate string

var version = "dev"

func init() {
	dlqreg.Register()
}

func main() {
	root, err := parseRoot(os.Args)
	if err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
	defer root.Close()
	if err := root.Run(); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
}

// rootCmd is the top-level command state.
type rootCmd struct {
	fs         *flag.FlagSet
	cfg        runtimeconfig.RuntimeConfig
	ConfigFile string
	args       []string
	db         *sql.DB
	Verbosity  int
}

func (r *rootCmd) DB() (*sql.DB, error) {
	if r.db != nil {
		return r.db, nil
	}
	if ue := dbstart.InitDB(r.cfg); ue != nil {
		return nil, fmt.Errorf("init db: %w", ue.Err)
	}
	r.db = dbstart.GetDBPool()
	return r.db, nil
}

func (r *rootCmd) Close() {
	if r.db != nil {
		if err := r.db.Close(); err != nil {
			log.Printf("close db: %v", err)
		}
	}
}

func parseRoot(args []string) (*rootCmd, error) {
	r := &rootCmd{}
	early := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var cfgPath string
	var showVersion bool
	early.StringVar(&cfgPath, "config-file", "", "path to config file")
	early.BoolVar(&showVersion, "version", false, "print version and exit")
	_ = early.Parse(args[1:])
	if cfgPath == "" {
		cfgPath = os.Getenv(config.EnvConfigFile)
	}
	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}
	fileVals, err := config.LoadAppConfigFile(core.OSFS{}, cfgPath)
	if err != nil {
		if errors.Is(err, config.ErrConfigFileNotFound) {
			return nil, fmt.Errorf("config file not found: %s", cfgPath)
		}
		return nil, fmt.Errorf("load config file: %w", err)
	}
	fs := runtimeconfig.NewRuntimeFlagSet(args[0])
	fs.StringVar(&cfgPath, "config-file", cfgPath, "path to config file")
	fs.IntVar(&r.Verbosity, "verbosity", 0, "verbosity level")
	_ = fs.Parse(args[1:])
	r.fs = fs
	r.args = fs.Args()
	r.ConfigFile = cfgPath
	r.cfg = runtimeconfig.GenerateRuntimeConfig(fs, fileVals, os.Getenv)
	return r, nil
}

func (r *rootCmd) Run() error {
	if len(r.args) == 0 {
		r.fs.Usage()
		return fmt.Errorf("no command provided")
	}
	switch r.args[0] {
	case "help", "usage":
		c, err := parseHelpCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("help: %w", err)
		}
		return c.Run()
	case "serve":
		c, err := parseServeCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("serve: %w", err)
		}
		return c.Run()
	case "user":
		c, err := parseUserCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("user: %w", err)
		}
		return c.Run()
	case "email":
		c, err := parseEmailCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("email: %w", err)
		}
		return c.Run()
	case "db":
		c, err := parseDbCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("db: %w", err)
		}
		return c.Run()
	case "perm":
		c, err := parsePermCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("perm: %w", err)
		}
		return c.Run()
	case "board":
		c, err := parseBoardCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("board: %w", err)
		}
		return c.Run()
	case "blog", "blogs":
		c, err := parseBlogCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("blog: %w", err)
		}
		return c.Run()
	case "writing", "writings":
		c, err := parseWritingCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("writing: %w", err)
		}
		return c.Run()
	case "news":
		c, err := parseNewsCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("news: %w", err)
		}
		return c.Run()
	case "faq":
		c, err := parseFaqCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("faq: %w", err)
		}
		return c.Run()
	case "ipban":
		c, err := parseIpBanCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("ipban: %w", err)
		}
		return c.Run()
	case "images":
		c, err := parseImagesCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("images: %w", err)
		}
		return c.Run()
	case "audit":
		c, err := parseAuditCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("audit: %w", err)
		}
		return c.Run()
	case "lang":
		c, err := parseLangCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("lang: %w", err)
		}
		return c.Run()
	case "server":
		c, err := parseServerCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("server: %w", err)
		}
		return c.Run()
	case "config":
		c, err := parseConfigCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		return c.Run()
	default:
		return fmt.Errorf("unknown command %q", r.args[0])
	}
}

// Usage prints command usage information with examples.
func (r *rootCmd) Usage() {
	executeUsage(r.fs.Output(), rootUsageTemplate, r.fs, r.fs.Name())
}
