package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http/httptest"
	"os"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/internal/db"
)

// testVerificationTemplateCmd implements the "template" subcommand under "test verification".
type testVerificationTemplateCmd struct {
	*testVerificationCmd
	fs       *flag.FlagSet
	Template string
	DataFile string
}

func parseTestVerificationTemplateCmd(parent *testVerificationCmd, args []string) (*testVerificationTemplateCmd, error) {
	c := &testVerificationTemplateCmd{testVerificationCmd: parent}
	c.fs = newFlagSet("template")
	c.fs.StringVar(&c.Template, "template", "", "Path to the template file (e.g. news/postPage.gohtml)")
	c.fs.StringVar(&c.DataFile, "data", "", "Path to the JSON data file")
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

	// Mock DB
	qs := &db.QuerierStub{}
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
	templates.SetDir("core/templates")

	// Verify template existence
	if !templates.TemplateExists(c.Template) {
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

	// Execute
	funcs := cd.Funcs(req)

	// Render
	tmpl := templates.GetCompiledSiteTemplates(funcs)
	if err := tmpl.ExecuteTemplate(os.Stdout, c.Template, dot); err != nil {
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
		// Convert float64 to int32 if it's a whole number, as Go templates are strict about types
		// and most IDs in DB are int32.
		if m == float64(int32(m)) {
			return int32(m)
		}
		return m
	case string:
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
