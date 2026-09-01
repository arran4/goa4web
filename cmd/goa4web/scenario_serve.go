//go:build sqlite || sqlite3

package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "modernc.org/sqlite"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/database"
	"github.com/arran4/goa4web/internal/app"
	"github.com/arran4/goa4web/internal/app/dbstart"
	"github.com/arran4/goa4web/internal/app/server"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/scenario"
	"github.com/arran4/goa4web/internal/sqlutil"
	"github.com/arran4/goa4web/migrations"
)

// scenarioServeCmd starts a disposable Goa4Web HTTP server populated from a scenario.
type scenarioServeCmd struct {
	*scenarioCmd
	fs     *flag.FlagSet
	Path   string
	Listen string
	fsys   fs.FS
	dbConn *sql.DB
}

func parseScenarioServeCmd(parent *scenarioCmd, args []string) (*scenarioServeCmd, error) {
	c := &scenarioServeCmd{scenarioCmd: parent}
	c.fs = newFlagSet("serve")
	c.fs.Usage = c.Usage
	c.fs.StringVar(&c.Listen, "listen", "", "The address and port for the HTTP server to listen on.")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	if len(c.fs.Args()) > 0 {
		c.Path = c.fs.Arg(0)
	}
	return c, nil
}

func generateEphemeralSecret(explicitSecret string) (string, error) {
	if explicitSecret != "" {
		return explicitSecret, nil
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// Bootstrap initializes an ephemeral SQLite database, applies migrations, seeds,
// runs scenario preflight, applies the scenario, and creates the *server.Server.
// It returns the server, database connection, cleanup function, and any error encountered.
func (c *scenarioServeCmd) Bootstrap(ctx context.Context) (*server.Server, *sql.DB, func(), error) {
	if c.Path == "" {
		c.fs.Usage()
		return nil, nil, nil, fmt.Errorf("scenario path required")
	}

	sc, err := loadScenarioPath(c.fsys, c.Path)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := scenario.Validate(sc); err != nil {
		return nil, nil, nil, fmt.Errorf("validate scenario: %w", err)
	}
	c.Infof("scenario %q valid", sc.Meta.Name)

	// Safety requirement: Deliberately isolate from configured production/development databases.
	// We build an isolated RuntimeConfig that ignores DB_CONN and uses an ephemeral SQLite database.
	var serveCfg config.RuntimeConfig
	if c.cfg != nil {
		serveCfg = *c.cfg
	}
	serveCfg.DBDriver = "sqlite3"
	serveCfg.DBConn = ""
	serveCfg.DBHost = ""
	serveCfg.DBUser = ""
	serveCfg.DBPass = ""
	serveCfg.DBName = ""
	if c.Listen != "" {
		serveCfg.HTTPListen = c.Listen
	}
	if serveCfg.HTTPListen == "" {
		serveCfg.HTTPListen = ":8080"
	}

	var dbConn *sql.DB
	var cleanupDB func()
	if c.dbConn != nil {
		dbConn = c.dbConn
		cleanupDB = func() {}
	} else {
		dsn := fmt.Sprintf("file:scenario_serve_%d?mode=memory&cache=shared", time.Now().UnixNano())
		openedDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("open ephemeral sqlite database: %w", err)
		}
		dbConn = openedDB
		cleanupDB = func() {
			_ = openedDB.Close()
		}
	}

	// 1. Apply schema migrations in-memory
	if err := dbstart.Apply(ctx, dbConn, migrations.FS, false, "sqlite3"); err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("apply migrations: %w", err)
	}

	// 2. Apply database seed SQL (roles, role grants, etc.)
	seedSQL := database.SeedSQLForDriver("sqlite3")
	if err := sqlutil.RunStatements(ctx, dbConn, strings.NewReader(string(seedSQL))); err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("run seed SQL: %w", err)
	}

	// 3. Seed default language
	if _, err := dbConn.ExecContext(ctx, `INSERT INTO language (id, nameof) VALUES (1, 'English') ON CONFLICT DO NOTHING;`); err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("insert default language: %w", err)
	}

	c.Infof("ephemeral SQLite database initialized")

	querier := db.NewForDriver(dbConn, "sqlite3")
	cd := common.NewCoreData(ctx, querier, &serveCfg)
	runner := scenario.NewRunner(cd)

	res, err := runner.Apply(ctx, sc)
	if err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("apply scenario: %w", err)
	}
	c.Infof("scenario %q applied successfully (%d events)", res.ScenarioName, res.EventsApplied)

	// Process-local generated session and signing secrets
	secret, err := generateEphemeralSecret(serveCfg.SessionSecret)
	if err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("session secret: %w", err)
	}
	signKey, err := generateEphemeralSecret(serveCfg.ImageSignSecret)
	if err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("image sign secret: %w", err)
	}
	linkKey, err := generateEphemeralSecret(serveCfg.LinkSignSecret)
	if err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("link sign secret: %w", err)
	}
	shareKey, err := generateEphemeralSecret(serveCfg.ShareSignSecret)
	if err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("share sign secret: %w", err)
	}
	apiKey, err := generateEphemeralSecret(serveCfg.AdminAPISecret)
	if err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("admin api secret: %w", err)
	}

	srv, err := app.NewServer(ctx, &serveCfg, c.adminHandlers,
		app.WithSessionSecret(secret),
		app.WithImageSignSecret(signKey),
		app.WithLinkSignSecret(linkKey),
		app.WithShareSignSecret(shareKey),
		app.WithDB(dbConn),
		app.WithQuerier(querier),
		app.WithEmailRegistry(c.emailReg),
		app.WithDLQRegistry(c.dlqReg),
		app.WithTasksRegistry(c.tasksReg),
		app.WithAPISecret(apiKey),
		app.WithRouterRegistry(c.routerReg),
	)
	if err != nil {
		cleanupDB()
		return nil, nil, nil, fmt.Errorf("create server: %w", err)
	}

	listenAddr := serveCfg.HTTPListen
	if strings.HasPrefix(listenAddr, ":") {
		listenAddr = "http://localhost" + listenAddr
	} else if !strings.HasPrefix(listenAddr, "http://") && !strings.HasPrefix(listenAddr, "https://") {
		listenAddr = "http://" + listenAddr
	}
	c.Infof("scenario environment listening on %s", listenAddr)
	c.Infof("environment is disposable and will be removed on exit")

	cleanup := func() {
		srv.Close()
		cleanupDB()
	}

	return srv, dbConn, cleanup, nil
}

func (c *scenarioServeCmd) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv, _, cleanup, err := c.Bootstrap(ctx)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := srv.RunContext(ctx); err != nil {
		return err
	}
	return nil
}

func (c *scenarioServeCmd) Usage() {
	_ = executeUsage(c.fs.Output(), "scenario_serve_usage.txt", c)
}

func (c *scenarioServeCmd) FlagGroups() []flagGroup {
	return append(c.scenarioCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*scenarioServeCmd)(nil)
