package main

import (
	"io/fs"
	"testing"
	"testing/fstest"
)

func TestLoadHTTPConfigFile(t *testing.T) {
	fsys := fstest.MapFS{
		"http.conf": {Data: []byte("LISTEN=1.2.3.4:80\n")},
	}
	old := httpReadFile
	httpReadFile = func(name string) ([]byte, error) { return fs.ReadFile(fsys, name) }
	defer func() { httpReadFile = old }()
	cfg, err := loadHTTPConfigFile("http.conf")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Listen != "1.2.3.4:80" {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestResolveHTTPConfigPrecedence(t *testing.T) {
	env := HTTPConfig{Listen: ":1"}
	file := HTTPConfig{Listen: ":2"}
	cli := HTTPConfig{Listen: ":3"}

	cfg := resolveHTTPConfig(cli, file, env)

	if cfg.Listen != ":3" {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadHTTPConfigEnvPath(t *testing.T) {
	fsys := fstest.MapFS{
		"http.conf": {Data: []byte("LISTEN=:9\n")},
	}
	old := httpReadFile
	httpReadFile = func(name string) ([]byte, error) { return fs.ReadFile(fsys, name) }
	defer func() { httpReadFile = old }()
	t.Setenv("HTTP_CONFIG_FILE", "http.conf")
	httpConfigFile = ""
	cliHTTPConfig = HTTPConfig{}
	cfg := loadHTTPConfig()
	if cfg.Listen != ":9" {
		t.Fatalf("want :9 got %q", cfg.Listen)
	}
}
