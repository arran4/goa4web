package main

import (
	"flag"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

// notificationsTasksCmd implements "notifications tasks".
type notificationsTasksCmd struct {
	*notificationsCmd
	fs *flag.FlagSet
}

// Usage prints command usage information with examples.
func (c *notificationsTasksCmd) Usage() {
	executeUsage(c.fs.Output(), "notifications_tasks_usage.txt", c)
}

func (c *notificationsTasksCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*notificationsTasksCmd)(nil)

func parseNotificationsTasksCmd(parent *notificationsCmd, args []string) (*notificationsTasksCmd, error) {
	c := &notificationsTasksCmd{notificationsCmd: parent}
	c.fs = newFlagSet("tasks")

	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *notificationsTasksCmd) Run() error {
	tw := table.NewWriter()
	tw.SetOutputMirror(c.fs.Output())
	tw.AppendHeader(table.Row{"Section", "Task", "Self Email", "Self Internal", "Direct Email", "Target Email", "Target Internal", "Subscribed Email", "Subscribed Internal", "Admin Email", "Admin Internal"})
	for _, info := range taskTemplateInfos(c.notificationsCmd.rootCmd.tasksReg) {
		tw.AppendRow(table.Row{
			info.Section,
			info.Task,
			strings.Join(info.SelfEmail, ","),
			info.SelfInternal,
			strings.Join(info.DirectEmail, ","),
			strings.Join(info.TargetEmail, ","),
			info.TargetInternal,
			strings.Join(info.SubEmail, ","),
			info.SubInternal,
			strings.Join(info.AdminEmail, ","),
			info.AdminInternal,
		})
	}
	tw.Render()
	return nil
}
