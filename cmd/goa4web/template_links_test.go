package main

import (
	"fmt"
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

	// Collect route regexps
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

	// 2. Define roots to scan
	possibleSiteRoots := []string{
		"../../core/templates/site", // from cmd/goa4web
		"core/templates/site",       // from root
	}

	var scanDirs []string

	for _, path := range possibleSiteRoots {
		if _, err := os.Stat(path); err == nil {
			scanDirs = append(scanDirs, path)
			break
		}
	}

	// Add testdata dir
	if _, err := os.Stat("testdata"); err == nil {
		scanDirs = append(scanDirs, "testdata")
	} else if _, err := os.Stat("cmd/goa4web/testdata"); err == nil {
		scanDirs = append(scanDirs, "cmd/goa4web/testdata")
	} else {
        // Create testdata if missing? No, we created it.
        // Assuming we are running in the right dir.
    }

	if len(scanDirs) == 0 {
		t.Skip("Cannot find template directories")
	}

	hrefRegex := regexp.MustCompile(`href="([^"]+)"`)
	tmplRegex := regexp.MustCompile(`\{\{[^}]+\}\}`)

	var validationErrors []string
	foundTestdataError := false

	for _, rootDir := range scanDirs {
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

				if strings.HasPrefix(link, "http") ||
					strings.HasPrefix(link, "mailto:") ||
					strings.HasPrefix(link, "#") ||
					strings.HasPrefix(link, "javascript:") ||
					strings.HasPrefix(link, "data:") {
					continue
				}

				if !strings.HasPrefix(link, "/") {
					continue
				}

				if strings.Contains(link, "{{if ") || strings.Contains(link, "{{range ") || strings.Contains(link, "{{with ") {
					continue
				}

				if strings.HasPrefix(link, "{{") && strings.HasSuffix(link, "}}") && strings.Count(link, "{{") == 1 {
					continue
				}

				urlPath := link
				if idx := strings.Index(urlPath, "?"); idx != -1 {
					urlPath = urlPath[:idx]
				}

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
					sample := tmplRegex.ReplaceAllString(link, "{VAR}")
					msg := fmt.Sprintf("File %s: Invalid link %q (normalized: %s)", path, link, sample)

					if strings.Contains(path, "testdata") && strings.Contains(link, "this-route-does-not-exist") {
						foundTestdataError = true
					} else {
						validationErrors = append(validationErrors, msg)
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	if !foundTestdataError {
		t.Error("Verification failed: Did not catch known invalid link in testdata/bad_link.gohtml")
	}

	if len(validationErrors) > 0 {
		for _, msg := range validationErrors {
			t.Error(msg)
		}
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
