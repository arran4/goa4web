package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/faq_templates"
)

// Helper to parse the template format
func parseTemplateContent(content string) (string, string, error) {
	// Normalize CRLF to LF
	content = strings.ReplaceAll(content, "\r\n", "\n")
	parts := strings.SplitN(content, "\n===\n", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid template format: missing '===' separator")
	}
	question := strings.TrimSpace(parts[0])
	answer := strings.TrimSpace(parts[1])
	return question, answer, nil
}

type faqAddFromTemplateCmd struct {
	*faqCmd
	categoryID int
	languageID int
	authorID   int
}

func parseFaqAddFromTemplateCmd(parent *faqCmd, args []string) (*faqAddFromTemplateCmd, error) {
	c := &faqAddFromTemplateCmd{faqCmd: parent}
	c.fs = newFlagSet("faq add-from-template")
	c.fs.IntVar(&c.categoryID, "category", 0, "Category ID")
	c.fs.IntVar(&c.languageID, "language", 0, "Language ID")
	c.fs.IntVar(&c.authorID, "author", 0, "Author ID")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqAddFromTemplateCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 1 {
		return fmt.Errorf("missing template name")
	}
	templateName := args[0]
	content, err := faq_templates.Get(templateName)
	if err != nil {
		return fmt.Errorf("template %q not found: %w", templateName, err)
	}

	question, answer, err := parseTemplateContent(content)
	if err != nil {
		return err
	}

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	lid := sql.NullInt32{}
	if c.languageID != 0 {
		lid = sql.NullInt32{Int32: int32(c.languageID), Valid: true}
	}
	cid := sql.NullInt32{}
	if c.categoryID != 0 {
		cid = sql.NullInt32{Int32: int32(c.categoryID), Valid: true}
	}

	params := db.AdminCreateFAQParams{
		Question:   sql.NullString{String: question, Valid: true},
		Answer:     sql.NullString{String: answer, Valid: true},
		CategoryID: cid,
		AuthorID:   int32(c.authorID),
		LanguageID: lid,
		Priority:   0,
	}

	res, err := d.AdminCreateFAQ(context.Background(), params)
	if err != nil {
		return fmt.Errorf("create faq: %w", err)
	}
	id, _ := res.LastInsertId()
	fmt.Printf("Created FAQ %d\n", id)
	return nil
}

type faqListTemplatesCmd struct {
	*faqCmd
}

func parseFaqListTemplatesCmd(parent *faqCmd, args []string) (*faqListTemplatesCmd, error) {
	c := &faqListTemplatesCmd{faqCmd: parent}
	c.fs = newFlagSet("faq list-templates")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqListTemplatesCmd) Run() error {
	names, err := faq_templates.List()
	if err != nil {
		return err
	}
	for _, name := range names {
		fmt.Println(name)
	}
	return nil
}

type faqDumpCmd struct {
	*faqCmd
}

func parseFaqDumpCmd(parent *faqCmd, args []string) (*faqDumpCmd, error) {
	c := &faqDumpCmd{faqCmd: parent}
	c.fs = newFlagSet("faq dump")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqDumpCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 1 {
		return fmt.Errorf("missing faq ID")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	f, err := d.AdminGetFAQByID(context.Background(), int32(id))
	if err != nil {
		return fmt.Errorf("get faq: %w", err)
	}

	fmt.Printf("%s\n===\n%s\n", f.Question.String, f.Answer.String)
	return nil
}

type faqUpdateCmd struct {
	*faqCmd
	file string
}

func parseFaqUpdateCmd(parent *faqCmd, args []string) (*faqUpdateCmd, error) {
	c := &faqUpdateCmd{faqCmd: parent}
	c.fs = newFlagSet("faq update")
	c.fs.StringVar(&c.file, "file", "", "File to read from (defaults to stdin)")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqUpdateCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 1 {
		return fmt.Errorf("missing faq ID")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	var content []byte
	if c.file != "" {
		content, err = os.ReadFile(c.file)
		if err != nil {
			return err
		}
	} else {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	}

	question, answer, err := parseTemplateContent(string(content))
	if err != nil {
		return err
	}

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	existing, err := d.AdminGetFAQByID(context.Background(), int32(id))
	if err != nil {
		return fmt.Errorf("get faq: %w", err)
	}

	params := db.AdminUpdateFAQQuestionAnswerParams{
		Question:   sql.NullString{String: question, Valid: true},
		Answer:     sql.NullString{String: answer, Valid: true},
		CategoryID: existing.CategoryID,
		ID:         int32(id),
	}

	err = d.AdminUpdateFAQQuestionAnswer(context.Background(), params)
	if err != nil {
		return fmt.Errorf("update faq: %w", err)
	}
	fmt.Printf("Updated FAQ %d\n", id)
	return nil
}

type faqReorderCmd struct {
	*faqCmd
}

func parseFaqReorderCmd(parent *faqCmd, args []string) (*faqReorderCmd, error) {
	c := &faqReorderCmd{faqCmd: parent}
	c.fs = newFlagSet("faq reorder")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqReorderCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 2 {
		return fmt.Errorf("usage: faq reorder <id> <priority>")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}
	priority, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid priority: %w", err)
	}

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	err = d.AdminUpdateFAQPriority(context.Background(), db.AdminUpdateFAQPriorityParams{
		Priority: int32(priority),
		ID:       int32(id),
	})
	if err != nil {
		return fmt.Errorf("update priority: %w", err)
	}
	fmt.Printf("Updated priority for FAQ %d to %d\n", id, priority)
	return nil
}

type faqListCmd struct {
	*faqCmd
}

func parseFaqListCmd(parent *faqCmd, args []string) (*faqListCmd, error) {
	c := &faqListCmd{faqCmd: parent}
	c.fs = newFlagSet("faq list")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqListCmd) Run() error {
	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	list, err := d.SystemGetFAQQuestions(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("%-5s %-5s %-30s %s\n", "ID", "Prio", "Question", "Answer Snippet")
	for _, f := range list {
		ans := f.Answer.String
		if len(ans) > 50 {
			ans = ans[:47] + "..."
		}
		ans = strings.ReplaceAll(ans, "\n", " ")
		fmt.Printf("%-5d %-5d %-30s %s\n", f.ID, f.Priority, f.Question.String, ans)
	}
	return nil
}

type faqDeleteCmd struct {
	*faqCmd
}

func parseFaqDeleteCmd(parent *faqCmd, args []string) (*faqDeleteCmd, error) {
	c := &faqDeleteCmd{faqCmd: parent}
	c.fs = newFlagSet("faq delete")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqDeleteCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 1 {
		return fmt.Errorf("missing faq ID")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	err = d.AdminDeleteFAQ(context.Background(), int32(id))
	if err != nil {
		return fmt.Errorf("delete faq: %w", err)
	}
	fmt.Printf("Deleted FAQ %d\n", id)
	return nil
}

type faqCreateCmd struct {
	*faqCmd
	categoryID int
	languageID int
	authorID   int
	file       string
}

func parseFaqCreateCmd(parent *faqCmd, args []string) (*faqCreateCmd, error) {
	c := &faqCreateCmd{faqCmd: parent}
	c.fs = newFlagSet("faq create")
	c.fs.IntVar(&c.categoryID, "category", 0, "Category ID")
	c.fs.IntVar(&c.languageID, "language", 0, "Language ID")
	c.fs.IntVar(&c.authorID, "author", 0, "Author ID")
	c.fs.StringVar(&c.file, "file", "", "File to read from (defaults to stdin)")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCreateCmd) Run() error {
	var content []byte
	var err error
	if c.file != "" {
		content, err = os.ReadFile(c.file)
		if err != nil {
			return err
		}
	} else {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	}

	question, answer, err := parseTemplateContent(string(content))
	if err != nil {
		return err
	}

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	lid := sql.NullInt32{}
	if c.languageID != 0 {
		lid = sql.NullInt32{Int32: int32(c.languageID), Valid: true}
	}
	cid := sql.NullInt32{}
	if c.categoryID != 0 {
		cid = sql.NullInt32{Int32: int32(c.categoryID), Valid: true}
	}

	params := db.AdminCreateFAQParams{
		Question:   sql.NullString{String: question, Valid: true},
		Answer:     sql.NullString{String: answer, Valid: true},
		CategoryID: cid,
		AuthorID:   int32(c.authorID),
		LanguageID: lid,
		Priority:   0,
	}

	res, err := d.AdminCreateFAQ(context.Background(), params)
	if err != nil {
		return fmt.Errorf("create faq: %w", err)
	}
	id, _ := res.LastInsertId()
	fmt.Printf("Created FAQ %d\n", id)
	return nil
}
