package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arran4/goa4web"
	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/runtimeconfig"
)

func main() {
	early := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var cfgPath string
	early.StringVar(&cfgPath, "config-file", "", "path to application configuration file")
	_ = early.Parse(os.Args[1:])
	if cfgPath == "" {
		cfgPath = os.Getenv(config.EnvConfigFile)
	}

	fileVals := goa4web.LoadAppConfigFile(core.OSFS{}, cfgPath)

	fs := runtimeconfig.NewRuntimeFlagSet(os.Args[0])
	var (
		sessionSecret     string
		sessionSecretFile string
	)
	fs.StringVar(&sessionSecret, "session-secret", "", "session secret key")
	fs.StringVar(&sessionSecretFile, "session-secret-file", "", "path to session secret file")
	fs.StringVar(&cfgPath, "config-file", cfgPath, "path to application configuration file")

	_ = fs.Parse(os.Args[1:])

	goa4web.ConfigFile = cfgPath

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

	if err := goa4web.RunWithConfig(ctx, cfg, secret); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}
}
