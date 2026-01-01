package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// faqCategoryCmd handles category subcommands.
type faqCategoryCmd struct {
	*faqCmd
	fs *flag.FlagSet
}

func parseFaqCategoryCmd(parent *faqCmd, args []string) (*faqCategoryCmd, error) {
	c := &faqCategoryCmd{faqCmd: parent}
	c.fs = newFlagSet("faq category")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCategoryCmd) Run() error {
	args := c.fs.Args()
	if len(args) == 0 {
		c.fs.Usage()
		return fmt.Errorf("missing category subcommand (create, list, update, delete)")
	}
	if err := usageIfHelp(c.fs, args); err != nil {
		return err
	}
	switch args[0] {
	case "create":
		cmd, err := parseFaqCategoryCreateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return cmd.Run()
	case "list":
		cmd, err := parseFaqCategoryListCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		return cmd.Run()
	case "update":
		cmd, err := parseFaqCategoryUpdateCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return cmd.Run()
	case "delete":
		cmd, err := parseFaqCategoryDeleteCmd(c, args[1:])
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}
		return cmd.Run()
	default:
		c.fs.Usage()
		return fmt.Errorf("unknown category command %q", args[0])
	}
}

func (c *faqCategoryCmd) Usage() {
	executeUsage(c.fs.Output(), "faq_category_usage.txt", c)
}

func (c *faqCategoryCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*faqCategoryCmd)(nil)

// Create
type faqCategoryCreateCmd struct {
	*faqCategoryCmd
	fs         *flag.FlagSet
	parentID   int
	languageID int
}

func parseFaqCategoryCreateCmd(parent *faqCategoryCmd, args []string) (*faqCategoryCreateCmd, error) {
	c := &faqCategoryCreateCmd{faqCategoryCmd: parent}
	c.fs = newFlagSet("faq category create")
	c.fs.IntVar(&c.parentID, "parent", 0, "Parent Category ID")
	c.fs.IntVar(&c.languageID, "language", 0, "Language ID")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCategoryCreateCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 1 {
		c.fs.Usage()
		return fmt.Errorf("usage: faq category create [flags] <name>")
	}
	name := strings.Join(args, " ")

	conn, err := c.rootCmd.getDB()
	if err != nil {
		return err
	}
	defer closeDB(conn)
	d := db.New(conn)

	pid := sql.NullInt32{}
	if c.parentID != 0 {
		pid = sql.NullInt32{Int32: int32(c.parentID), Valid: true}
	}
	lid := sql.NullInt32{}
	if c.languageID != 0 {
		lid = sql.NullInt32{Int32: int32(c.languageID), Valid: true}
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

func (c *faqCategoryCreateCmd) Usage() {
	// Standard flag usage for now
	c.fs.Usage()
}

func (c *faqCategoryCreateCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*faqCategoryCreateCmd)(nil)

// List
type faqCategoryListCmd struct {
	*faqCategoryCmd
	fs *flag.FlagSet
}

func parseFaqCategoryListCmd(parent *faqCategoryCmd, args []string) (*faqCategoryListCmd, error) {
	c := &faqCategoryListCmd{faqCategoryCmd: parent}
	c.fs = newFlagSet("faq category list")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCategoryListCmd) Run() error {
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

	categoryMap := make(map[int32]*db.FaqCategory)
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

func (c *faqCategoryListCmd) Usage() {
	c.fs.Usage()
}

func (c *faqCategoryListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*faqCategoryListCmd)(nil)

// Update
type faqCategoryUpdateCmd struct {
	*faqCategoryCmd
	fs         *flag.FlagSet
	parentID   int
	languageID int
	name       string
}

func parseFaqCategoryUpdateCmd(parent *faqCategoryCmd, args []string) (*faqCategoryUpdateCmd, error) {
	c := &faqCategoryUpdateCmd{faqCategoryCmd: parent}
	c.fs = newFlagSet("faq category update")
	c.fs.IntVar(&c.parentID, "parent", -1, "New Parent Category ID (0 for root, -1 to keep current)")
	c.fs.IntVar(&c.languageID, "language", -1, "New Language ID (0 for none, -1 to keep current)")
	c.fs.StringVar(&c.name, "name", "", "New Name")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCategoryUpdateCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 1 {
		c.fs.Usage()
		return fmt.Errorf("usage: faq category update [flags] <id>")
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

	if c.name != "" {
		params.Name = sql.NullString{String: c.name, Valid: true}
	}
	if c.parentID != -1 {
		if c.parentID == 0 {
			params.ParentCategoryID = sql.NullInt32{Valid: false}
		} else {
			params.ParentCategoryID = sql.NullInt32{Int32: int32(c.parentID), Valid: true}
		}
	}
	if c.languageID != -1 {
		if c.languageID == 0 {
			params.LanguageID = sql.NullInt32{Valid: false}
		} else {
			params.LanguageID = sql.NullInt32{Int32: int32(c.languageID), Valid: true}
		}
	}

	err = d.AdminUpdateFAQCategory(context.Background(), params)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	fmt.Printf("Updated FAQ Category %d\n", id)
	return nil
}

func (c *faqCategoryUpdateCmd) Usage() {
	c.fs.Usage()
}

func (c *faqCategoryUpdateCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*faqCategoryUpdateCmd)(nil)

// Delete
type faqCategoryDeleteCmd struct {
	*faqCategoryCmd
	fs        *flag.FlagSet
	migrateTo int
}

func parseFaqCategoryDeleteCmd(parent *faqCategoryCmd, args []string) (*faqCategoryDeleteCmd, error) {
	c := &faqCategoryDeleteCmd{faqCategoryCmd: parent}
	c.fs = newFlagSet("faq category delete")
	c.fs.IntVar(&c.migrateTo, "migrate-to", 0, "Migrate content and subcategories to this Category ID before deletion (0 to error if not empty)")
	c.fs.Usage = c.Usage
	if err := c.fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *faqCategoryDeleteCmd) Run() error {
	args := c.fs.Args()
	if len(args) < 1 {
		c.fs.Usage()
		return fmt.Errorf("usage: faq category delete [flags] <id>")
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

	// Check content
	qCount, err := d.AdminGetFAQCategoryWithQuestionCountByID(context.Background(), id)
	if err != nil {
		// Category might not exist or error
		return fmt.Errorf("get category info: %w", err)
	}

	// Check for children
	cats, err := d.AdminListFAQCategories(context.Background())
	if err != nil {
		return err
	}
	hasChildren := false
	for _, cat := range cats {
		if cat.ParentCategoryID.Valid && cat.ParentCategoryID.Int32 == id {
			hasChildren = true
			break
		}
	}

	if qCount.Questioncount > 0 || hasChildren {
		if c.migrateTo == 0 {
			return fmt.Errorf("category is not empty (has %d questions or subcategories). use -migrate-to <id> to move content.", qCount.Questioncount)
		}
		// Migrate
		fmt.Printf("Migrating content to Category %d...\n", c.migrateTo)
		targetID := sql.NullInt32{}
		if c.migrateTo > 0 {
			targetID = sql.NullInt32{Int32: int32(c.migrateTo), Valid: true}
		} else {
			// migrate to root? or literally ID 0?
			// The flag description implies ID.
			// If user passes -migrate-to 0, it means error if not empty.
			// But if they want to migrate to root, they can't via flag 0?
			// Let's assume -migrate-to targetID. If targetID is valid category.
			// Wait, my flag default is 0.
			// If they pass -migrate-to 5, migrateTo=5.
			// If they pass -migrate-to -1 (root)?
			// I'll assume standard IDs. Root is usually NULL parent, but content must have category?
			// `faq` table `category_id` is nullable.
			// So `migrate-to -1` could mean make content uncategorized (root)?
			// I'll support `migrate-to` as ID.
		}

		// Move content
		// AdminMoveFAQContent: UPDATE faq SET category_id = new WHERE category_id = old
		err = d.AdminMoveFAQContent(context.Background(), db.AdminMoveFAQContentParams{
			NewCategoryID: targetID,
			OldCategoryID: sql.NullInt32{Int32: id, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("migrate content: %w", err)
		}

		// Move children
		// AdminMoveFAQChildren: UPDATE faq_categories SET parent = new WHERE parent = old
		err = d.AdminMoveFAQChildren(context.Background(), db.AdminMoveFAQChildrenParams{
			NewParentID: targetID,
			OldParentID: sql.NullInt32{Int32: id, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("migrate children: %w", err)
		}
	}

	err = d.AdminDeleteFAQCategory(context.Background(), id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	fmt.Printf("Deleted FAQ Category %d\n", id)
	return nil
}

func (c *faqCategoryDeleteCmd) Usage() {
	c.fs.Usage()
}

func (c *faqCategoryDeleteCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*faqCategoryDeleteCmd)(nil)
