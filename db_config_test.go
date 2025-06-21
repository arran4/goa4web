package main

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestLoadDBConfigFile(t *testing.T) {
	fsys := fstest.MapFS{
		"db.conf": {Data: []byte("DB_USER=u\nDB_PASS=p\nDB_HOST=h\nDB_PORT=3307\nDB_NAME=n\n")},
	}
	old := dbReadFile
	dbReadFile = func(name string) ([]byte, error) { return fs.ReadFile(fsys, name) }
	defer func() { dbReadFile = old }()
	cfg, err := loadDBConfigFile("db.conf")
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
	fsys := fstest.MapFS{
		"db.conf": {Data: []byte("DB_USER=envfile\n")},
	}
	old := dbReadFile
	dbReadFile = func(name string) ([]byte, error) { return fs.ReadFile(fsys, name) }
	defer func() { dbReadFile = old }()
	t.Setenv("DB_CONFIG_FILE", "db.conf")
	dbConfigFile = ""
	cliDBConfig = DBConfig{}
	cfg := loadDBConfig()
	if cfg.User != "envfile" {
		t.Fatalf("want envfile got %q", cfg.User)
	}
}
