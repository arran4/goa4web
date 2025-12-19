package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/arran4/goa4web/internal/db"
)

type roleListAllCmd struct {
	*roleListCmd
	fs *flag.FlagSet
}

func parseRoleListAllCmd(parent *roleListCmd, args []string) (*roleListAllCmd, error) {
	c := &roleListAllCmd{roleListCmd: parent}
	fs := flag.NewFlagSet("all", flag.ContinueOnError)
	c.fs = fs
	fs.SetOutput(parent.fs.Output())
	fs.Usage = c.Usage
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *roleListAllCmd) Run() error {
	// 1. Get DB Roles
	sdb, err := c.getDB()
	var dbRoles []*db.AdminListRolesWithUsersRow
	if err == nil {
		defer closeDB(sdb)
		q := db.New(sdb)
		dbRoles, err = q.AdminListRolesWithUsers(context.Background())
		if err != nil {
			return fmt.Errorf("fetching roles from db: %w", err)
		}
	} else {
		fmt.Fprintf(c.fs.Output(), "Warning: Database connection failed: %v. Showing embedded roles only.\n\n", err)
	}

	// 2. Get Embedded Roles
	embeddedFiles, err := listEmbeddedRoles()
	if err != nil {
		return fmt.Errorf("listing embedded roles: %w", err)
	}

	embeddedRolesMap := make(map[string]string) // filename -> role name inside
	for _, filename := range embeddedFiles {
		name, err := readEmbeddedRoleName(filename)
		if err != nil {
			// Log error but continue? Or fail? failing seems strict but consistent.
			// Let's just note it as error
			name = fmt.Sprintf("<error: %v>", err)
		}
		embeddedRolesMap[filename] = name
	}

	// 3. Merge Data
	// Keys will be the Role Name. However, embedded roles are identified by filename, but have an internal name.
	// DB roles have a name.
	// We want to list by "Identity". Identity is ambiguous.
	// Let's create a map of "Role Name" -> Info.
	// If a filename exists, we also want to show it.
	// If we key by Name, we can match DB name to Seed Name.

	type roleInfo struct {
		SeedFile string
		SeedName string
		DBName   string
		DBID     int32
		DBUsers  string
		InDB     bool
		InSeed   bool
	}

	rolesMap := make(map[string]*roleInfo)

	// Process DB Roles
	for _, r := range dbRoles {
		info := &roleInfo{
			DBName:  r.Name,
			DBID:    r.ID,
			InDB:    true,
			DBUsers: r.Users.String,
		}
		rolesMap[r.Name] = info
	}

	// Process Embedded Roles
	for filename, name := range embeddedRolesMap {
		if info, ok := rolesMap[name]; ok {
			info.SeedFile = filename
			info.SeedName = name
			info.InSeed = true
		} else {
			rolesMap[name] = &roleInfo{
				SeedFile: filename,
				SeedName: name,
				InSeed:   true,
			}
		}
	}

	// 4. Sort
	var roleNames []string
	for name := range rolesMap {
		roleNames = append(roleNames, name)
	}
	sort.Strings(roleNames)

	// 5. Print Table
	w := tabwriter.NewWriter(c.fs.Output(), 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Role Name\tSeed File\tDB ID\tStatus\tUsers (DB)")
	fmt.Fprintln(w, "---------\t---------\t-----\t------\t----------")

	for _, name := range roleNames {
		info := rolesMap[name]
		status := "OK"
		if info.InDB && !info.InSeed {
			status = "DB Only"
		} else if !info.InDB && info.InSeed {
			status = "Seed Only"
		}

		dbIDStr := ""
		if info.InDB {
			dbIDStr = fmt.Sprintf("%d", info.DBID)
		} else {
			dbIDStr = "-"
		}

		seedFile := info.SeedFile
		if seedFile == "" {
			seedFile = "-"
		}

		// Truncate users if too long?
		users := info.DBUsers
		if len(users) > 50 {
			users = users[:47] + "..."
		}
		if users == "" && info.InDB {
			users = "-"
		} else if !info.InDB {
			users = "N/A"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, seedFile, dbIDStr, status, users)
	}

	return w.Flush()
}

func (c *roleListAllCmd) Usage() { executeUsage(c.fs.Output(), "role_list_all_usage.txt", c) }

func (c *roleListAllCmd) FlagGroups() []flagGroup {
	return append(c.roleListCmd.FlagGroups(), flagGroup{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)})
}
