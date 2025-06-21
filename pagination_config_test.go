package main

import "testing"

func TestLoadPaginationConfigFile(t *testing.T) {
	useMemFS(t)
	file := "pagination.conf"
	content := "PAGE_SIZE_MIN=10\nPAGE_SIZE_MAX=40\n"
	if err := writeFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	cfg, err := loadPaginationConfigFile(file)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Min != 10 || cfg.Max != 40 {
		t.Fatalf("unexpected cfg: %#v", cfg)
	}
}

func TestResolvePaginationConfigPrecedence(t *testing.T) {
	env := PaginationConfig{Min: 5, Max: 30}
	file := PaginationConfig{Min: 8, Max: 20}
	cli := PaginationConfig{Min: 12}
	cfg := resolvePaginationConfig(cli, file, env)
	if cfg.Min != 12 || cfg.Max != 20 {
		t.Fatalf("merged %#v", cfg)
	}
}

func TestLoadPaginationConfigEnvPath(t *testing.T) {
	useMemFS(t)
	file := "pagination.conf"
	if err := writeFile(file, []byte("PAGE_SIZE_MIN=7\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	t.Setenv("PAGINATION_CONFIG_FILE", file)
	paginationConfigFile = ""
	cliPaginationConfig = PaginationConfig{}
	cfg := loadPaginationConfig()
	if cfg.Min != 7 {
		t.Fatalf("want 7 got %d", cfg.Min)
	}
}
