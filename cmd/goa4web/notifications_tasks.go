package main

import (
	_ "embed"
	"flag"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

//go:embed templates/notifications_tasks_usage.txt
var notificationsTasksUsageTemplate string

// notificationsTasksCmd implements "notifications tasks".
type notificationsTasksCmd struct {
	*notificationsCmd
	fs   *flag.FlagSet
	args []string
}

// Usage prints command usage information with examples.
func (c *notificationsTasksCmd) Usage() {
	executeUsage(c.fs.Output(), notificationsTasksUsageTemplate, c.fs, c.notificationsCmd.rootCmd.fs.Name())
}

func parseNotificationsTasksCmd(parent *notificationsCmd, args []string) (*notificationsTasksCmd, error) {
	c := &notificationsTasksCmd{notificationsCmd: parent}
	fs := flag.NewFlagSet("tasks", flag.ContinueOnError)
	c.fs = fs
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	c.args = fs.Args()
	return c, nil
}

func (c *notificationsTasksCmd) Run() error {
	tw := table.NewWriter()
	tw.SetOutputMirror(c.fs.Output())
	tw.AppendHeader(table.Row{"Task", "Self Email", "Self Internal", "Subscribed Email", "Subscribed Internal", "Admin Email", "Admin Internal"})
	for _, info := range taskTemplateInfos() {
		tw.AppendRow(table.Row{
			info.Task,
			strings.Join(info.SelfEmail, ","),
			info.SelfInternal,
			strings.Join(info.SubEmail, ","),
			info.SubInternal,
			strings.Join(info.AdminEmail, ","),
			info.AdminInternal,
		})
	}
	tw.Render()
	return nil
}
