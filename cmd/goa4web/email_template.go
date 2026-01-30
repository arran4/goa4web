package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	adminhandlers "github.com/arran4/goa4web/handlers/admin"
	"github.com/arran4/goa4web/internal/db"
)

// emailTemplateCmd handles `email template` subcommands.
type emailTemplateCmd struct {
	*emailCmd
	fs *flag.FlagSet
}

func parseEmailTemplateCmd(parent *emailCmd, args []string) (*emailTemplateCmd, error) {
	c := &emailTemplateCmd{emailCmd: parent}
	c.fs = newFlagSet("template")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailTemplateCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing email template command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "get":
		cmd, err := parseEmailTemplateGetCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("get: %w", err)
		}
		return cmd.Run()
	case "set":
		cmd, err := parseEmailTemplateSetCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("set: %w", err)
		}
		return cmd.Run()
	case "test":
		cmd, err := parseEmailTemplateTestCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("test: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown email template command %q", args[0])
	}
}

// Usage prints command usage information with examples.
func (c *emailTemplateCmd) Usage() {
	executeUsage(c.fs.Output(), "email_template_usage.txt", c)
}

func (c *emailTemplateCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*emailTemplateCmd)(nil)

type emailTemplateGetCmd struct {
	*emailTemplateCmd
	fs     *flag.FlagSet
	name   string
	output string
}

func parseEmailTemplateGetCmd(parent *emailTemplateCmd, args []string) (*emailTemplateGetCmd, error) {
	c := &emailTemplateGetCmd{emailTemplateCmd: parent}
	c.fs = newFlagSet("get")
	c.fs.StringVar(&c.name, "name", "", "template name (optional; shows list when empty)")
	c.fs.StringVar(&c.output, "output", "", "output file path (default: stdout)")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailTemplateGetCmd) Run() error {
	cd, err := c.templateCoreData(c.rootCmd.Querier)
	if err != nil {
		return err
	}
	reqURL := "http://localhost/admin/email/template"
	if c.name != "" {
		reqURL += "?name=" + url.QueryEscape(c.name)
	}
	req := httptest.NewRequest(http.MethodGet, reqURL, nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()
	adminhandlers.AdminEmailTemplatePage(rec, req)
	writer, closeFn, err := outputWriter(c.output)
	if err != nil {
		return err
	}
	defer func() { _ = closeFn() }()
	if _, err := writer.Write(rec.Body.Bytes()); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func (c *emailTemplateGetCmd) Usage() {
	executeUsage(c.fs.Output(), "email_template_usage.txt", c)
}

func (c *emailTemplateGetCmd) FlagGroups() []flagGroup {
	return append(c.emailTemplateCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*emailTemplateGetCmd)(nil)

type emailTemplateSetCmd struct {
	*emailTemplateCmd
	fs     *flag.FlagSet
	name   string
	body   string
	file   string
	output string
	dryRun bool
}

func parseEmailTemplateSetCmd(parent *emailTemplateCmd, args []string) (*emailTemplateSetCmd, error) {
	c := &emailTemplateSetCmd{emailTemplateCmd: parent}
	c.fs = newFlagSet("set")
	c.fs.StringVar(&c.name, "name", "", "template name to update")
	c.fs.StringVar(&c.body, "body", "", "template body (use --file or stdin for large content)")
	c.fs.StringVar(&c.file, "file", "", "file path containing the template body")
	c.fs.StringVar(&c.output, "output", "", "output file path (default: stdout)")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "preview changes without writing to the database")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailTemplateSetCmd) Run() error {
	if c.name == "" {
		return fmt.Errorf("name flag is required")
	}
	body, err := c.readBody()
	if err != nil {
		return err
	}
	q, cleanup, err := c.templateQuerier(c.dryRun)
	if err != nil {
		return err
	}
	defer func() { _ = cleanup() }()
	cd, err := c.templateCoreData(func() (db.Querier, error) { return q, nil })
	if err != nil {
		return err
	}
	form := url.Values{}
	form.Set("name", c.name)
	form.Set("body", body)
	req := httptest.NewRequest(http.MethodPost, "http://localhost/admin/email/template", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()
	task := adminhandlers.SaveTemplateTask{TaskString: adminhandlers.TaskUpdate}
	if err := handleTaskResult(task.Action(rec, req)); err != nil {
		return err
	}
	if c.output != "" || c.dryRun {
		writer, closeFn, err := outputWriter(c.output)
		if err != nil {
			return err
		}
		defer func() { _ = closeFn() }()
		if c.dryRun {
			if _, err := fmt.Fprintf(writer, "Dry run: would set template %q with %d bytes\n", c.name, len(body)); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
		} else {
			if _, err := fmt.Fprintf(writer, "Updated template %q\n", c.name); err != nil {
				return fmt.Errorf("write output: %w", err)
			}
		}
	}
	return nil
}

func (c *emailTemplateSetCmd) readBody() (string, error) {
	if c.file != "" {
		b, err := os.ReadFile(c.file)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}
		return string(b), nil
	}
	if c.body != "" {
		return c.body, nil
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	if len(b) == 0 {
		return "", fmt.Errorf("template body is required")
	}
	return string(b), nil
}

func (c *emailTemplateSetCmd) Usage() {
	executeUsage(c.fs.Output(), "email_template_usage.txt", c)
}

func (c *emailTemplateSetCmd) FlagGroups() []flagGroup {
	return append(c.emailTemplateCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*emailTemplateSetCmd)(nil)

type emailTemplateTestCmd struct {
	*emailTemplateCmd
	fs     *flag.FlagSet
	userID int
	output string
	dryRun bool
	host   string
}

func parseEmailTemplateTestCmd(parent *emailTemplateCmd, args []string) (*emailTemplateTestCmd, error) {
	c := &emailTemplateTestCmd{emailTemplateCmd: parent}
	c.fs = newFlagSet("test")
	c.fs.IntVar(&c.userID, "user-id", 0, "user ID to send the preview email to")
	c.fs.StringVar(&c.host, "host", "localhost", "host to use when building preview URLs")
	c.fs.StringVar(&c.output, "output", "", "output file path (default: stdout)")
	c.fs.BoolVar(&c.dryRun, "dry-run", false, "preview email generation without inserting into the queue")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *emailTemplateTestCmd) Run() error {
	if c.userID == 0 {
		return fmt.Errorf("user-id flag is required")
	}
	cfg, err := c.rootCmd.RuntimeConfig()
	if err != nil {
		return err
	}
	provider, err := c.rootCmd.emailReg.ProviderFromConfig(cfg)
	if err != nil || provider == nil {
		if err != nil {
			return fmt.Errorf("email provider: %w", err)
		}
		return fmt.Errorf("email provider not configured")
	}
	q, cleanup, err := c.templateQuerier(c.dryRun)
	if err != nil {
		return err
	}
	defer func() { _ = cleanup() }()
	cd, err := c.templateCoreData(func() (db.Querier, error) { return q, nil },
		common.WithEmailProvider(provider),
		common.WithEmailRegistry(c.rootCmd.emailReg),
	)
	if err != nil {
		return err
	}
	cd.UserID = int32(c.userID)
	req := httptest.NewRequest(http.MethodPost, "http://"+c.host+"/admin/email/template", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	rec := httptest.NewRecorder()
	task := adminhandlers.TestTemplateTask{TaskString: adminhandlers.TaskTestMail}
	if err := handleTaskResult(task.Action(rec, req)); err != nil {
		return err
	}
	writer, closeFn, err := outputWriter(c.output)
	if err != nil {
		return err
	}
	defer func() { _ = closeFn() }()
	if c.dryRun {
		if _, err := fmt.Fprintf(writer, "Dry run: would queue preview email for user ID %d\n", c.userID); err != nil {
			return fmt.Errorf("write output: %w", err)
		}
		return nil
	}
	if _, err := fmt.Fprintf(writer, "Queued preview email for user ID %d\n", c.userID); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}

func (c *emailTemplateTestCmd) Usage() {
	executeUsage(c.fs.Output(), "email_template_usage.txt", c)
}

func (c *emailTemplateTestCmd) FlagGroups() []flagGroup {
	return append(c.emailTemplateCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}

var _ usageData = (*emailTemplateTestCmd)(nil)

func (c *emailTemplateCmd) templateCoreData(queries func() (db.Querier, error), opts ...common.CoreOption) (*common.CoreData, error) {
	cfg, err := c.rootCmd.RuntimeConfig()
	if err != nil {
		return nil, err
	}
	q, err := queries()
	if err != nil {
		return nil, err
	}
	modules := c.rootCmd.routerReg.Names()
	coreOpts := append([]common.CoreOption{
		common.WithTasksRegistry(c.rootCmd.tasksReg),
		common.WithRouterModules(modules),
	}, opts...)
	return common.NewCoreData(c.rootCmd.Context(), q, cfg, coreOpts...), nil
}

func (c *emailTemplateCmd) templateQuerier(dryRun bool) (db.Querier, func() error, error) {
	if !dryRun {
		q, err := c.rootCmd.Querier()
		return q, func() error { return nil }, err
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return nil, func() error { return nil }, err
	}
	tx, err := conn.BeginTx(c.rootCmd.Context(), nil)
	if err != nil {
		return nil, func() error { return nil }, err
	}
	q := db.New(tx)
	return q, tx.Rollback, nil
}

func handleTaskResult(result any) error {
	if result == nil {
		return nil
	}
	if err, ok := result.(error); ok {
		return err
	}
	return nil
}

func outputWriter(path string) (io.Writer, func() error, error) {
	if path == "" {
		return os.Stdout, func() error { return nil }, nil
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, func() error { return nil }, fmt.Errorf("create output file: %w", err)
	}
	return f, f.Close, nil
}
