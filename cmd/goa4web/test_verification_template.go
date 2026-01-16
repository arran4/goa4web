package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/gorilla/mux"
)

// testVerificationTemplateCmd implements the "template" subcommand under "test verification".
type testVerificationTemplateCmd struct {
	*testVerificationCmd
	fs       *flag.FlagSet
	Template string
	DataFile string
	Output   string
	Listen   string
}

func parseTestVerificationTemplateCmd(parent *testVerificationCmd, args []string) (*testVerificationTemplateCmd, error) {
	c := &testVerificationTemplateCmd{testVerificationCmd: parent}
	c.fs = newFlagSet("template")
	c.fs.StringVar(&c.Template, "template", "", "Path to the template file (e.g. news/postPage.gohtml)")
	c.fs.StringVar(&c.DataFile, "data", "", "Path to the JSON data file")
	c.fs.StringVar(&c.Output, "output", "", "Output file path (default: stdout)")
	c.fs.StringVar(&c.Listen, "listen", "", "Address to listen on (e.g. :8080)")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

type TemplateVerificationData struct {
	Dot    json.RawMessage       `json:"Dot"`
	Config *config.RuntimeConfig `json:"Config"`
	URL    string                `json:"URL"`
	User   *db.User              `json:"User"`
}

func (c *testVerificationTemplateCmd) Run() error {
	if c.Template == "" {
		return fmt.Errorf("template flag is required")
	}

	var data TemplateVerificationData
	if c.DataFile != "" {
		b, err := os.ReadFile(c.DataFile)
		if err != nil {
			return fmt.Errorf("read data file: %w", err)
		}
		if err := json.Unmarshal(b, &data); err != nil {
			return fmt.Errorf("unmarshal data: %w", err)
		}
	} else {
		// Default empty config if not provided
		data.Config = &config.RuntimeConfig{}
	}

	// Override templates dir for verification if empty (assume dev mode like old code did)
	if data.Config.TemplatesDir == "" {
		data.Config.TemplatesDir = "core/templates"
	}

	// Mock DB
	qs := testhelpers.NewQuerierStub()
	if data.User != nil {
		qs.SystemGetUserByIDRow = &db.SystemGetUserByIDRow{
			Idusers:  data.User.Idusers,
			Username: data.User.Username,
		}
	}

	// Initialize CoreData
	ctx := context.Background()
	cd := common.NewCoreData(ctx, qs, data.Config)
	if data.User != nil {
		cd.UserID = data.User.Idusers
	}

	// Setup Request
	u := data.URL
	if u == "" {
		u = "http://localhost/"
	}
	req := httptest.NewRequest("GET", u, nil)

	// Load template

	// Verify template existence
	if !templates.TemplateExists(c.Template, templates.WithDir(data.Config.TemplatesDir)) {
		return fmt.Errorf("template %q not found", c.Template)
	}

	// Decode Dot data
	var dot map[string]any
	if len(data.Dot) > 0 {
		if err := json.Unmarshal(data.Dot, &dot); err != nil {
			return fmt.Errorf("unmarshal Dot: %w", err)
		}
	}

	// Fix time and number fields in Dot
	fixDataFields(dot)

	// Prepare execution
	funcs := cd.Funcs(req)
	tmpl := templates.GetCompiledSiteTemplates(funcs, templates.WithDir(data.Config.TemplatesDir))

	// Define handler for rendering
	renderHandler := func(w http.ResponseWriter, r *http.Request) {
		// Re-initialize funcs for the current request (important for CSRF, URL etc if we were fully using them)
		// But here we reuse the initial CD which might be enough for static verification.
		// NOTE: In a real server, CD is created per request. Here we reuse `cd` for simplicity.

		if err := tmpl.ExecuteTemplate(w, c.Template, dot); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

	if c.Listen != "" {
		r := mux.NewRouter()

		cfg := data.Config
		// Static assets
		r.HandleFunc("/main.css", handlers.MainCSS(cfg)).Methods("GET")
		r.HandleFunc("/favicon.svg", handlers.Favicon(cfg)).Methods("GET")
		r.HandleFunc("/static/site.js", handlers.SiteJS(cfg)).Methods("GET")
		r.HandleFunc("/static/a4code.js", handlers.A4CodeJS(cfg)).Methods("GET")

		// Module specific assets that are often hardcoded in templates
		// These paths should ideally be dynamic but for verification we map common ones
		r.HandleFunc("/forum/topic_labels.js", handlers.TopicLabelsJS(cfg)).Methods("GET")
		r.HandleFunc("/private/topic_labels.js", handlers.TopicLabelsJS(cfg)).Methods("GET")
		r.HandleFunc("/news/topic_labels.js", handlers.TopicLabelsJS(cfg)).Methods("GET") // Guessed path
		// Add more if discovered

		// Render page
		r.HandleFunc("/", renderHandler).Methods("GET")

		fmt.Printf("Listening on %s...\n", c.Listen)
		return http.ListenAndServe(c.Listen, r)
	}

	// Single shot render
	var out = os.Stdout
	if c.Output != "" {
		f, err := os.Create(c.Output)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	// We use a response recorder to capture output to handle http.ResponseWriter interface if needed,
	// but ExecuteTemplate writes to io.Writer.
	if err := tmpl.ExecuteTemplate(out, c.Template, dot); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}

func fixDataFields(v any) any {
	switch m := v.(type) {
	case map[string]any:
		for k, val := range m {
			m[k] = fixDataFields(val)
		}
		return m
	case []any:
		for i, val := range m {
			m[i] = fixDataFields(val)
		}
		return m
	case float64:
		// Convert float64 to int32 if it's a whole number
		if m == float64(int32(m)) {
			return int32(m)
		}
		return m
	case string:
		// Try to parse as time
		if t, err := time.Parse(time.RFC3339, m); err == nil {
			return t
		}
		return m
	default:
		return m
	}
}

func (c *testVerificationTemplateCmd) Usage() {
	executeUsage(c.fs.Output(), "test_verification_template_usage.txt", c)
}

func (c *testVerificationTemplateCmd) FlagGroups() []flagGroup {
	return append(c.testVerificationCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*testVerificationTemplateCmd)(nil)
