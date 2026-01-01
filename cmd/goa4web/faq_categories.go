package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// faqCategoryCmd handles category subcommands.
type faqCategoryCmd struct {
	*faqCmd
}

func parseFaqCategoryCmd(parent *faqCmd, args []string) (*faqCategoryCmd, error) {
	c := &faqCategoryCmd{faqCmd: parent}
	// We do not define flags here as we dispatch to subcommands
	c.fs = newFlagSet("faq category")
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCategoryCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		return fmt.Errorf("missing category subcommand (create, list, update, delete)")
	}
	switch args[0] {
	case "create":
		return c.runCreate(args[1:])
	case "list":
		return c.runList(args[1:])
	case "update":
		return c.runUpdate(args[1:])
	case "delete":
		return c.runDelete(args[1:])
	default:
		return fmt.Errorf("unknown category command %q", args[0])
	}
}

func (c *faqCategoryCmd) runCreate(args []string) error {
	fs := newFlagSet("faq category create")
	parentID := fs.Int("parent", 0, "Parent Category ID")
	languageID := fs.Int("language", 0, "Language ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: faq category create <name>")
	}
	name := strings.Join(fs.Args(), " ")

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	pid := sql.NullInt32{}
	if *parentID != 0 {
		pid = sql.NullInt32{Int32: int32(*parentID), Valid: true}
	}
	lid := sql.NullInt32{}
	if *languageID != 0 {
		lid = sql.NullInt32{Int32: int32(*languageID), Valid: true}
	}

	res, err := d.AdminCreateFAQCategory(context.Background(), db.AdminCreateFAQCategoryParams{
		Name:             sql.NullString{String: name, Valid: true},
		ParentCategoryID: pid,
		LanguageID:       lid,
	})
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}
	id, _ := res.LastInsertId()
	fmt.Printf("Created FAQ Category %d: %s\n", id, name)
	return nil
}

func (c *faqCategoryCmd) runList(args []string) error {
	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	cats, err := d.AdminListFAQCategories(context.Background())
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}

	// Build a map for hierarchy display if needed, but simple list first
	// Or try to indent based on parent_id
	categoryMap := make(map[int32]db.FaqCategory)
	childrenMap := make(map[int32][]int32)
	var rootIds []int32

	for _, cat := range cats {
		categoryMap[cat.ID] = cat
		if cat.ParentCategoryID.Valid {
			childrenMap[cat.ParentCategoryID.Int32] = append(childrenMap[cat.ParentCategoryID.Int32], cat.ID)
		} else {
			rootIds = append(rootIds, cat.ID)
		}
	}

	var printTree func(id int32, level int)
	printTree = func(id int32, level int) {
		cat := categoryMap[id]
		indent := strings.Repeat("  ", level)
		fmt.Printf("%s%d: %s\n", indent, cat.ID, cat.Name.String)
		for _, childID := range childrenMap[id] {
			printTree(childID, level+1)
		}
	}

	fmt.Println("FAQ Categories Hierarchy:")
	for _, id := range rootIds {
		printTree(id, 0)
	}
	return nil
}

func (c *faqCategoryCmd) runUpdate(args []string) error {
	fs := newFlagSet("faq category update")
	parentID := fs.Int("parent", -1, "New Parent Category ID (0 for root, -1 to keep current)")
	languageID := fs.Int("language", -1, "New Language ID (0 for none, -1 to keep current)")
	name := fs.String("name", "", "New Name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: faq category update [flags] <id>")
	}
	idVal, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}
	id := int32(idVal)

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	// Fetch existing
	current, err := d.AdminGetFAQCategory(context.Background(), id)
	if err != nil {
		return fmt.Errorf("get category: %w", err)
	}

	params := db.AdminUpdateFAQCategoryParams{
		ID:               id,
		Name:             current.Name,
		ParentCategoryID: current.ParentCategoryID,
		LanguageID:       current.LanguageID,
	}

	if *name != "" {
		params.Name = sql.NullString{String: *name, Valid: true}
	}
	if *parentID != -1 {
		if *parentID == 0 {
			params.ParentCategoryID = sql.NullInt32{Valid: false}
		} else {
			params.ParentCategoryID = sql.NullInt32{Int32: int32(*parentID), Valid: true}
		}
	}
	if *languageID != -1 {
		if *languageID == 0 {
			params.LanguageID = sql.NullInt32{Valid: false}
		} else {
			params.LanguageID = sql.NullInt32{Int32: int32(*languageID), Valid: true}
		}
	}

	err = d.AdminUpdateFAQCategory(context.Background(), params)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	fmt.Printf("Updated FAQ Category %d\n", id)
	return nil
}

func (c *faqCategoryCmd) runDelete(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: faq category delete <id>")
	}
	idVal, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}
	id := int32(idVal)

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	err = d.AdminDeleteFAQCategory(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	fmt.Printf("Deleted FAQ Category %d\n", id)
	return nil
}
