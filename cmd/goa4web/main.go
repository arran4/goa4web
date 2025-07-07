package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	dbstart "github.com/arran4/goa4web/internal/dbstart"
	dlqreg "github.com/arran4/goa4web/internal/dlq/register/defaults"
	"github.com/arran4/goa4web/runtimeconfig"
)

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
	case "blog":
		c, err := parseBlogCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("blog: %w", err)
		}
		return c.Run()
	case "ipban":
		c, err := parseIpBanCmd(r, r.args[1:])
		if err != nil {
			return fmt.Errorf("ipban: %w", err)
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
	w := r.fs.Output()
	fmt.Fprintf(w, "Usage:\n  %s [flags] <command> [<args>]\n", r.fs.Name())
	fmt.Fprintln(w, "\nCommands:")
	fmt.Fprintln(w, "  serve\trun the web server")
	fmt.Fprintln(w, "  user\tmanage users")
	fmt.Fprintln(w, "  perm\tmanage permissions")
	fmt.Fprintln(w, "  config\treload or update configuration")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s serve\n", r.fs.Name())
	fmt.Fprintf(w, "  %s user add -username alice -password secret\n", r.fs.Name())
	fmt.Fprintf(w, "  %s perm list\n", r.fs.Name())
	fmt.Fprintf(w, "  %s config reload\n\n", r.fs.Name())
	fmt.Fprintln(w, "  board\tmanage image boards")
	fmt.Fprintln(w, "  blog\tmanage blog entries")
	fmt.Fprintln(w, "\nExamples:")
	fmt.Fprintf(w, "  %s board list\n", r.fs.Name())
	fmt.Fprintln(w, "  ipban\tmanage IP bans")
	fmt.Fprintf(w, "  %s ipban list\n\n", r.fs.Name())
	fmt.Fprintln(w, "  audit\tshow recent audit log entries")
	fmt.Fprintln(w, "  db\tmanage database")
	fmt.Fprintf(w, "  %s db migrate\n", r.fs.Name())
	fmt.Fprintln(w, "  lang\tmanage languages")
	fmt.Fprintf(w, "  %s lang list\n\n", r.fs.Name())
	fmt.Fprintln(w, "  server\tmanage the running server")
	fmt.Fprintln(w, "  email\tmanage emails")
	fmt.Fprintln(w, "  config\tmanage configuration")
	fmt.Fprintf(w, "  %s config show\n\n", r.fs.Name())
	r.fs.PrintDefaults()
}
