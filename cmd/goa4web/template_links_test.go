package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/handlers/admin"
	nav "github.com/arran4/goa4web/internal/navigation"
	"github.com/arran4/goa4web/internal/router"
	"github.com/gorilla/mux"
)

func TestTemplateLinks(t *testing.T) {
	// 1. Setup Router
	reg := router.NewRegistry()
	ah := admin.New()
	registerModules(reg, ah)

	r := mux.NewRouter()
	cfg := &config.RuntimeConfig{
		SessionName: "session",
	}
	navReg := nav.NewRegistry()

	reg.InitModules(r, cfg, navReg)
	router.RegisterRoutes(r, reg, cfg, navReg)

	// Collect route regexps to bypass MatcherFunc checks (e.g. auth)
	var routeRegexps []*regexp.Regexp
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		str, err := route.GetPathRegexp()
		if err != nil {
			return nil
		}
		re, err := regexp.Compile("^" + str + "$")
		if err == nil {
			routeRegexps = append(routeRegexps, re)
		}
		return nil
	})

	// 2. Find all templates
	possibleRoots := []string{
		"../../core/templates/site", // When running from cmd/goa4web
		"core/templates/site",       // When running from root
	}

	var rootDir string
	for _, path := range possibleRoots {
		if _, err := os.Stat(path); err == nil {
			rootDir = path
			break
		}
	}

	if rootDir == "" {
		t.Skip("Cannot find template directory")
	}

	hrefRegex := regexp.MustCompile(`href="([^"]+)"`)
	tmplRegex := regexp.MustCompile(`\{\{[^}]+\}\}`)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".gohtml") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		matches := hrefRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			link := match[1]

			// Validation Logic

			// 1. Skip external, anchor, javascript, etc.
			if strings.HasPrefix(link, "http") ||
				strings.HasPrefix(link, "mailto:") ||
				strings.HasPrefix(link, "#") ||
				strings.HasPrefix(link, "javascript:") ||
				strings.HasPrefix(link, "data:") {
				continue
			}

			// 2. Skip relative links (not starting with /)
			if !strings.HasPrefix(link, "/") {
				continue
			}

			// 3. Skip complex control structures
			if strings.Contains(link, "{{if ") || strings.Contains(link, "{{range ") || strings.Contains(link, "{{with ") {
				continue
			}

			// 4. Skip purely dynamic links (e.g. "{{.Url}}")
			if strings.HasPrefix(link, "{{") && strings.HasSuffix(link, "}}") && strings.Count(link, "{{") == 1 {
				continue
			}

			urlPath := link
			if idx := strings.Index(urlPath, "?"); idx != -1 {
				urlPath = urlPath[:idx]
			}

			// 5. Normalize and Match
			// Try variants for variable substitution
			variants := []string{"1", "dummy", "a", "user"}

			matched := false
			for _, v := range variants {
				normalized := tmplRegex.ReplaceAllString(urlPath, v)
				if matchAny(routeRegexps, normalized) {
					matched = true
					break
				}
			}

			if !matched {
				// Generate a sample normalization for the error message
				sample := tmplRegex.ReplaceAllString(link, "{VAR}")
				t.Errorf("File %s: Invalid link %q (normalized: %s)", path, link, sample)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func matchAny(regexps []*regexp.Regexp, urlStr string) bool {
	for _, re := range regexps {
		if re.MatchString(urlStr) {
			return true
		}
	}
	return false
}
