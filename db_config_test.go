package main

import (
	"testing"
)

func TestLoadDBConfigFile(t *testing.T) {
	useMemFS(t)
	file := "db.conf"
	content := "DB_USER=u\nDB_PASS=p\nDB_HOST=h\nDB_PORT=3307\nDB_NAME=n\n"
	if err := writeFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := loadDBConfigFile(file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.User != "u" || cfg.Pass != "p" || cfg.Host != "h" || cfg.Port != "3307" || cfg.Name != "n" {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestResolveDBConfigPrecedence(t *testing.T) {
	env := DBConfig{User: "env", Host: "env"}
	file := DBConfig{User: "file", Port: "1"}
	cli := DBConfig{Pass: "cli"}

	cfg := resolveDBConfig(cli, file, env)

	if cfg.User != "file" || cfg.Pass != "cli" || cfg.Host != "env" || cfg.Port != "1" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadDBConfigEnvPath(t *testing.T) {
	useMemFS(t)
	file := "db.conf"
	if err := writeFile(file, []byte("DB_USER=envfile\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	t.Setenv("DB_CONFIG_FILE", file)
	dbConfigFile = ""
	cliDBConfig = DBConfig{}
	cfg := loadDBConfig()
	if cfg.User != "envfile" {
		t.Fatalf("want envfile got %q", cfg.User)
	}
}
