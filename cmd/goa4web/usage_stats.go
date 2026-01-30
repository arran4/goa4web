package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

const (
	// usageStatsFormatTable is the default table output format for usage stats.
	usageStatsFormatTable = "table"
	// usageStatsFormatJSON is the JSON output format for usage stats.
	usageStatsFormatJSON = "json"
	// usageStatsDateLayout is the YYYY-MM-DD date layout for reporting windows.
	usageStatsDateLayout = "2006-01-02"
	// usageStatsTimeout caps the time spent collecting usage statistics.
	usageStatsTimeout = 5 * time.Minute
)

// usageCmd handles usage reporting subcommands.
type usageCmd struct {
	*rootCmd
	fs *flag.FlagSet
}

func parseUsageCmd(parent *rootCmd, args []string) (*usageCmd, error) {
	c := &usageCmd{rootCmd: parent}
	c.fs = newFlagSet("usage")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *usageCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing usage command")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "stats":
		cmd, err := parseUsageStatsCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("stats: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown usage command %q", args[0])
	}
}

func (c *usageCmd) Usage() {
	executeUsage(c.fs.Output(), "usage_usage.txt", c)
}

func (c *usageCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*usageCmd)(nil)

// usageStatsCmd implements "usage stats".
type usageStatsCmd struct {
	*usageCmd
	fs     *flag.FlagSet
	Since  string
	Until  string
	Format string
}

func parseUsageStatsCmd(parent *usageCmd, args []string) (*usageStatsCmd, error) {
	c := &usageStatsCmd{usageCmd: parent}
	fs, _, err := parseFlags("stats", args, func(fs *flag.FlagSet) {
		fs.StringVar(&c.Since, "since", "", "reporting window start (RFC3339 or YYYY-MM-DD)")
		fs.StringVar(&c.Until, "until", "", "reporting window end (RFC3339 or YYYY-MM-DD)")
		fs.StringVar(&c.Format, "format", usageStatsFormatTable, "output format (table, json)")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	c.fs.Usage = c.Usage
	return c, nil
}

type usageStatsData struct {
	Errors            []string
	ForumTopics       []*db.AdminForumTopicThreadCountsRow
	ForumHandlers     []*db.AdminForumHandlerThreadCountsRow
	ForumCategories   []*db.AdminForumCategoryThreadCountsRow
	WritingCategories []*db.AdminWritingCategoryCountsRow
	LinkerCategories  []*db.GetLinkerCategoryLinkCountsRow
	Imageboards       []*db.AdminImageboardPostCountsRow
	Users             []*db.AdminUserPostCountsRow
	Monthly           []*db.MonthlyUsageRow
	UserMonthly       []*db.UserMonthlyUsageRow
}

type usageStatsWindow struct {
	Since string `json:"since,omitempty"`
	Until string `json:"until,omitempty"`
}

type usageStatsReport struct {
	Window            usageStatsWindow        `json:"window"`
	StartYear         int                     `json:"start_year"`
	Errors            []string                `json:"errors,omitempty"`
	ForumTopics       []usageForumTopic       `json:"forum_topics"`
	ForumHandlers     []usageForumHandler     `json:"forum_handlers"`
	ForumCategories   []usageForumCategory    `json:"forum_categories"`
	WritingCategories []usageWritingCategory  `json:"writing_categories"`
	LinkerCategories  []usageLinkerCategory   `json:"linker_categories"`
	Imageboards       []usageImageboard       `json:"imageboards"`
	Users             []usageUserPosts        `json:"users"`
	Monthly           []usageMonthlyStats     `json:"monthly"`
	UserMonthly       []usageUserMonthlyStats `json:"user_monthly"`
}

type usageForumTopic struct {
	ID       int32  `json:"id"`
	Title    string `json:"title"`
	Handler  string `json:"handler"`
	Threads  int64  `json:"threads"`
	Comments int64  `json:"comments"`
}

type usageForumHandler struct {
	Handler  string `json:"handler"`
	Threads  int64  `json:"threads"`
	Comments int64  `json:"comments"`
}

type usageForumCategory struct {
	ID       int32  `json:"id"`
	Title    string `json:"title"`
	Threads  int64  `json:"threads"`
	Comments int64  `json:"comments"`
}

type usageWritingCategory struct {
	ID    int32  `json:"id"`
	Title string `json:"title"`
	Count int64  `json:"count"`
}

type usageLinkerCategory struct {
	ID    int32  `json:"id"`
	Title string `json:"title"`
	Count int64  `json:"count"`
}

type usageImageboard struct {
	ID    int32  `json:"id"`
	Title string `json:"title"`
	Count int64  `json:"count"`
}

type usageUserPosts struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Blogs    int64  `json:"blogs"`
	News     int64  `json:"news"`
	Comments int64  `json:"comments"`
	Images   int64  `json:"images"`
	Links    int64  `json:"links"`
	Writings int64  `json:"writings"`
}

type usageMonthlyStats struct {
	Year     int32 `json:"year"`
	Month    int32 `json:"month"`
	Blogs    int64 `json:"blogs"`
	News     int64 `json:"news"`
	Comments int64 `json:"comments"`
	Images   int64 `json:"images"`
	Links    int64 `json:"links"`
	Writings int64 `json:"writings"`
}

type usageUserMonthlyStats struct {
	UserID   int32  `json:"user_id"`
	Username string `json:"username"`
	Year     int32  `json:"year"`
	Month    int32  `json:"month"`
	Blogs    int64  `json:"blogs"`
	News     int64  `json:"news"`
	Comments int64  `json:"comments"`
	Images   int64  `json:"images"`
	Links    int64  `json:"links"`
	Writings int64  `json:"writings"`
}

func (c *usageStatsCmd) Run() error {
	format := strings.ToLower(strings.TrimSpace(c.Format))
	if format == "" {
		format = usageStatsFormatTable
	}
	if format != usageStatsFormatTable && format != usageStatsFormatJSON {
		return fmt.Errorf("invalid format %q (expected %s or %s)", c.Format, usageStatsFormatTable, usageStatsFormatJSON)
	}

	since, sinceSet, err := parseUsageStatsWindow(c.Since)
	if err != nil {
		return err
	}
	until, untilSet, err := parseUsageStatsWindow(c.Until)
	if err != nil {
		return err
	}
	if sinceSet && untilSet && usageStatsMonthKey(since) > usageStatsMonthKey(until) {
		return fmt.Errorf("since is after until")
	}

	cfg, err := c.rootCmd.RuntimeConfig()
	if err != nil {
		return fmt.Errorf("runtime config: %w", err)
	}
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	queries := db.New(conn)
	data := usageStatsData{}
	ctx, cancel := context.WithTimeout(c.rootCmd.Context(), usageStatsTimeout)
	defer cancel()

	addErr := func(name string, err error) {
		data.Errors = append(data.Errors, fmt.Errorf("%s: %w", name, err).Error())
	}

	if rows, err := queries.AdminForumTopicThreadCounts(ctx); err == nil {
		data.ForumTopics = rows
	} else {
		addErr("forum topic counts", err)
	}

	if rows, err := queries.AdminForumHandlerThreadCounts(ctx); err == nil {
		data.ForumHandlers = rows
	} else {
		addErr("forum handler counts", err)
	}

	if rows, err := queries.AdminForumCategoryThreadCounts(ctx); err == nil {
		data.ForumCategories = rows
	} else {
		addErr("forum category counts", err)
	}

	if rows, err := queries.AdminImageboardPostCounts(ctx); err == nil {
		data.Imageboards = rows
	} else {
		addErr("imageboard post counts", err)
	}

	if rows, err := queries.AdminUserPostCounts(ctx); err == nil {
		data.Users = rows
	} else {
		addErr("user post counts", err)
	}

	if rows, err := queries.AdminWritingCategoryCounts(ctx); err == nil {
		data.WritingCategories = rows
	} else {
		addErr("writing category counts", err)
	}

	if rows, err := queries.GetLinkerCategoryLinkCounts(ctx); err == nil {
		data.LinkerCategories = rows
	} else {
		addErr("linker category counts", err)
	}

	if rows, err := queries.MonthlyUsageCounts(ctx, int32(cfg.StatsStartYear)); err == nil {
		data.Monthly = filterUsageMonthly(rows, since, sinceSet, until, untilSet)
	} else {
		addErr("monthly usage counts", err)
	}

	if rows, err := queries.UserMonthlyUsageCounts(ctx, int32(cfg.StatsStartYear)); err == nil {
		data.UserMonthly = filterUsageUserMonthly(rows, since, sinceSet, until, untilSet)
	} else {
		addErr("user monthly usage counts", err)
	}

	data.ForumHandlers = ensureUsageHandlers(data.ForumHandlers)
	sort.Slice(data.ForumHandlers, func(i, j int) bool {
		return data.ForumHandlers[i].Handler < data.ForumHandlers[j].Handler
	})

	report := usageStatsReportFromData(data, cfg.StatsStartYear, since, sinceSet, until, untilSet)
	switch format {
	case usageStatsFormatJSON:
		return c.printJSON(report)
	default:
		return c.printTable(report)
	}
}

func (c *usageStatsCmd) Usage() {
	executeUsage(c.fs.Output(), "usage_stats_usage.txt", c)
}

func (c *usageStatsCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

func (c *usageStatsCmd) Prog() string { return c.rootCmd.fs.Name() }

var _ usageData = (*usageStatsCmd)(nil)

func parseUsageStatsWindow(raw string) (time.Time, bool, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false, nil
	}
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, true, nil
	}
	t, err := time.Parse(usageStatsDateLayout, raw)
	if err != nil {
		return time.Time{}, false, fmt.Errorf("invalid date %q (expected RFC3339 or %s)", raw, usageStatsDateLayout)
	}
	return t, true, nil
}

func usageStatsMonthKey(t time.Time) int {
	return t.Year()*12 + int(t.Month())
}

func filterUsageMonthly(rows []*db.MonthlyUsageRow, since time.Time, sinceSet bool, until time.Time, untilSet bool) []*db.MonthlyUsageRow {
	if !sinceSet && !untilSet {
		return rows
	}
	sinceKey := usageStatsMonthKey(since)
	untilKey := usageStatsMonthKey(until)
	filtered := make([]*db.MonthlyUsageRow, 0, len(rows))
	for _, row := range rows {
		key := int(row.Year)*12 + int(row.Month)
		if sinceSet && key < sinceKey {
			continue
		}
		if untilSet && key > untilKey {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func filterUsageUserMonthly(rows []*db.UserMonthlyUsageRow, since time.Time, sinceSet bool, until time.Time, untilSet bool) []*db.UserMonthlyUsageRow {
	if !sinceSet && !untilSet {
		return rows
	}
	sinceKey := usageStatsMonthKey(since)
	untilKey := usageStatsMonthKey(until)
	filtered := make([]*db.UserMonthlyUsageRow, 0, len(rows))
	for _, row := range rows {
		key := int(row.Year)*12 + int(row.Month)
		if sinceSet && key < sinceKey {
			continue
		}
		if untilSet && key > untilKey {
			continue
		}
		filtered = append(filtered, row)
	}
	return filtered
}

func ensureUsageHandlers(rows []*db.AdminForumHandlerThreadCountsRow) []*db.AdminForumHandlerThreadCountsRow {
	ensure := func(handler string) {
		for _, row := range rows {
			if row.Handler == handler {
				return
			}
		}
		rows = append(rows, &db.AdminForumHandlerThreadCountsRow{Handler: handler, Threads: 0, Comments: 0})
	}
	ensure("private")
	ensure("all")
	return rows
}

func usageStatsReportFromData(data usageStatsData, startYear int, since time.Time, sinceSet bool, until time.Time, untilSet bool) usageStatsReport {
	report := usageStatsReport{
		Window: usageStatsWindow{
			Since: formatUsageStatsWindow(since, sinceSet),
			Until: formatUsageStatsWindow(until, untilSet),
		},
		StartYear: startYear,
		Errors:    data.Errors,
	}

	for _, row := range data.ForumTopics {
		report.ForumTopics = append(report.ForumTopics, usageForumTopic{
			ID:       row.Idforumtopic,
			Title:    nullStringValue(row.Title),
			Handler:  row.Handler,
			Threads:  row.Threads,
			Comments: row.Comments,
		})
	}
	for _, row := range data.ForumHandlers {
		report.ForumHandlers = append(report.ForumHandlers, usageForumHandler{
			Handler:  row.Handler,
			Threads:  row.Threads,
			Comments: row.Comments,
		})
	}
	for _, row := range data.ForumCategories {
		report.ForumCategories = append(report.ForumCategories, usageForumCategory{
			ID:       row.Idforumcategory,
			Title:    nullStringValue(row.Title),
			Threads:  row.Threads,
			Comments: row.Comments,
		})
	}
	for _, row := range data.WritingCategories {
		report.WritingCategories = append(report.WritingCategories, usageWritingCategory{
			ID:    row.Idwritingcategory,
			Title: nullStringValue(row.Title),
			Count: row.Count,
		})
	}
	for _, row := range data.LinkerCategories {
		report.LinkerCategories = append(report.LinkerCategories, usageLinkerCategory{
			ID:    row.ID,
			Title: nullStringValue(row.Title),
			Count: row.Linkcount,
		})
	}
	for _, row := range data.Imageboards {
		report.Imageboards = append(report.Imageboards, usageImageboard{
			ID:    row.Idimageboard,
			Title: nullStringValue(row.Title),
			Count: row.Count,
		})
	}
	for _, row := range data.Users {
		report.Users = append(report.Users, usageUserPosts{
			ID:       row.Idusers,
			Username: nullStringValue(row.Username),
			Blogs:    row.Blogs,
			News:     row.News,
			Comments: row.Comments,
			Images:   row.Images,
			Links:    row.Links,
			Writings: row.Writings,
		})
	}
	for _, row := range data.Monthly {
		report.Monthly = append(report.Monthly, usageMonthlyStats{
			Year:     row.Year,
			Month:    row.Month,
			Blogs:    row.Blogs,
			News:     row.News,
			Comments: row.Comments,
			Images:   row.Images,
			Links:    row.Links,
			Writings: row.Writings,
		})
	}
	for _, row := range data.UserMonthly {
		report.UserMonthly = append(report.UserMonthly, usageUserMonthlyStats{
			UserID:   row.Idusers,
			Username: nullStringValue(row.Username),
			Year:     row.Year,
			Month:    row.Month,
			Blogs:    row.Blogs,
			News:     row.News,
			Comments: row.Comments,
			Images:   row.Images,
			Links:    row.Links,
			Writings: row.Writings,
		})
	}
	return report
}

func formatUsageStatsWindow(t time.Time, set bool) string {
	if !set {
		return ""
	}
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
		return t.Format(usageStatsDateLayout)
	}
	return t.Format(time.RFC3339)
}

func nullStringValue(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func (c *usageStatsCmd) printJSON(report usageStatsReport) error {
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}
	if _, err := fmt.Fprintln(c.fs.Output(), string(payload)); err != nil {
		return fmt.Errorf("write json: %w", err)
	}
	return nil
}

func (c *usageStatsCmd) printTable(report usageStatsReport) error {
	w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Usage stats (start year %d)\n", report.StartYear)
	if report.Window.Since != "" || report.Window.Until != "" {
		fmt.Fprintf(w, "Window\tSince\tUntil\n")
		fmt.Fprintf(w, "\t%s\t%s\n", report.Window.Since, report.Window.Until)
	}

	writeSectionHeader(w, "Forum Topics")
	fmt.Fprintln(w, "ID\tTitle\tHandler\tThreads\tComments")
	for _, row := range report.ForumTopics {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%d\n", row.ID, row.Title, row.Handler, row.Threads, row.Comments)
	}

	writeSectionHeader(w, "Forum Handlers")
	fmt.Fprintln(w, "Handler\tThreads\tComments")
	for _, row := range report.ForumHandlers {
		fmt.Fprintf(w, "%s\t%d\t%d\n", row.Handler, row.Threads, row.Comments)
	}

	writeSectionHeader(w, "Forum Categories")
	fmt.Fprintln(w, "ID\tTitle\tThreads\tComments")
	for _, row := range report.ForumCategories {
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\n", row.ID, row.Title, row.Threads, row.Comments)
	}

	writeSectionHeader(w, "Writing Categories")
	fmt.Fprintln(w, "ID\tTitle\tCount")
	for _, row := range report.WritingCategories {
		fmt.Fprintf(w, "%d\t%s\t%d\n", row.ID, row.Title, row.Count)
	}

	writeSectionHeader(w, "Linker Categories")
	fmt.Fprintln(w, "ID\tTitle\tCount")
	for _, row := range report.LinkerCategories {
		fmt.Fprintf(w, "%d\t%s\t%d\n", row.ID, row.Title, row.Count)
	}

	writeSectionHeader(w, "Imageboards")
	fmt.Fprintln(w, "ID\tTitle\tCount")
	for _, row := range report.Imageboards {
		fmt.Fprintf(w, "%d\t%s\t%d\n", row.ID, row.Title, row.Count)
	}

	writeSectionHeader(w, "User Posts")
	fmt.Fprintln(w, "ID\tUsername\tBlogs\tNews\tComments\tImages\tLinks\tWritings")
	for _, row := range report.Users {
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%d\t%d\t%d\t%d\n", row.ID, row.Username, row.Blogs, row.News, row.Comments, row.Images, row.Links, row.Writings)
	}

	writeSectionHeader(w, "Monthly Usage")
	fmt.Fprintln(w, "Year\tMonth\tBlogs\tNews\tComments\tImages\tLinks\tWritings")
	for _, row := range report.Monthly {
		fmt.Fprintf(w, "%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n", row.Year, row.Month, row.Blogs, row.News, row.Comments, row.Images, row.Links, row.Writings)
	}

	writeSectionHeader(w, "User Monthly Usage")
	fmt.Fprintln(w, "UserID\tUsername\tYear\tMonth\tBlogs\tNews\tComments\tImages\tLinks\tWritings")
	for _, row := range report.UserMonthly {
		fmt.Fprintf(w, "%d\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n",
			row.UserID,
			row.Username,
			row.Year,
			row.Month,
			row.Blogs,
			row.News,
			row.Comments,
			row.Images,
			row.Links,
			row.Writings,
		)
	}

	if len(report.Errors) > 0 {
		writeSectionHeader(w, "Errors")
		fmt.Fprintln(w, "Message")
		for _, msg := range report.Errors {
			fmt.Fprintf(w, "%s\n", msg)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush table: %w", err)
	}
	return nil
}

func writeSectionHeader(w *tabwriter.Writer, title string) {
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "%s\n", title)
}
