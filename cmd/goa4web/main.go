package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/internal/app"
	"github.com/arran4/goa4web/runtimeconfig"
)

var version = "dev"

func main() {
	early := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var cfgPath string
	var showVersion bool
	early.StringVar(&cfgPath, "config-file", "", "path to application configuration file")
	early.BoolVar(&showVersion, "version", false, "print version and exit")
	_ = early.Parse(os.Args[1:])
	if cfgPath == "" {
		cfgPath = os.Getenv(config.EnvConfigFile)
	}

	if showVersion {
		fmt.Println(version)
		return
	}

	fileVals := config.LoadAppConfigFile(core.OSFS{}, cfgPath)

	fs := runtimeconfig.NewRuntimeFlagSet(os.Args[0])
	var (
		sessionSecret     string
		sessionSecretFile string
	)
	fs.StringVar(&sessionSecret, "session-secret", "", "session secret key")
	fs.StringVar(&sessionSecretFile, "session-secret-file", "", "path to session secret file")
	fs.StringVar(&cfgPath, "config-file", cfgPath, "path to application configuration file")

	_ = fs.Parse(os.Args[1:])

	app.ConfigFile = cfgPath

	secretPath := sessionSecretFile
	if secretPath == "" {
		if v, ok := fileVals["SESSION_SECRET_FILE"]; ok {
			secretPath = v
		}
	}
	secret, err := core.LoadSessionSecret(core.OSFS{}, sessionSecret, secretPath)
	if err != nil {
		log.Fatalf("session secret: %v", err)
	}

	cfg := runtimeconfig.GenerateRuntimeConfig(fs, fileVals, os.Getenv)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.RunWithConfig(ctx, cfg, secret); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
}
