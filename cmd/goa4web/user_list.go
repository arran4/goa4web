package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/arran4/goa4web/internal/db"
)

// userListCmd implements "user list".
type userListCmd struct {
	*userCmd
	fs          *flag.FlagSet
	showAdmin   bool
	showCreated bool
	jsonOut     bool
}

func parseUserListCmd(parent *userCmd, args []string) (*userListCmd, error) {
	c := &userListCmd{userCmd: parent}
	fs, _, err := parseFlags("list", args, func(fs *flag.FlagSet) {
		fs.BoolVar(&c.showAdmin, "admin", false, "include admin status")
		fs.BoolVar(&c.showCreated, "created", false, "include creation date")
		fs.BoolVar(&c.jsonOut, "json", false, "machine-readable JSON output")
	})
	if err != nil {
		return nil, err
	}
	c.fs = fs
	return c, nil
}

type UserEmailStatus struct {
	Verified   []string
	Unverified []string
}

func (c *userListCmd) Usage() {
	executeUsage(c.fs.Output(), "user_list_usage.txt", c)
}

func (c *userListCmd) FlagGroups() []flagGroup {
	return []flagGroup{{Title: c.fs.Name() + " flags", Flags: flagInfos(c.fs)}}
}

var _ usageData = (*userListCmd)(nil)

func (c *userListCmd) Run() error {
	conn, err := c.rootCmd.DB()
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	ctx := context.Background()
	queries := db.New(conn)

	// Fetch all users
	rows, err := queries.SystemListAllUsers(ctx)
	if err != nil {
		return fmt.Errorf("list users: %w", err)
	}

	// Fetch roles
	roleRows, err := queries.ListUsersWithRoles(ctx)
	if err != nil {
		return fmt.Errorf("list user roles: %w", err)
	}
	rolesByUser := make(map[int32]string)
	for _, r := range roleRows {
		rolesByUser[r.Idusers] = toString(r.Roles)
	}

	// Fetch all emails
	emailRows, err := queries.SystemListAllUserEmails(ctx)
	if err != nil {
		return fmt.Errorf("list user emails: %w", err)
	}
	emailsByUser := make(map[int32]*UserEmailStatus)
	for _, e := range emailRows {
		if _, ok := emailsByUser[e.UserID]; !ok {
			emailsByUser[e.UserID] = &UserEmailStatus{}
		}
		// e.VerifiedAt is sql.NullTime
		if e.VerifiedAt.Valid {
			emailsByUser[e.UserID].Verified = append(emailsByUser[e.UserID].Verified, e.Email)
		} else {
			emailsByUser[e.UserID].Unverified = append(emailsByUser[e.UserID].Unverified, e.Email)
		}
	}

	if c.jsonOut {
		out := make([]map[string]interface{}, 0, len(rows))
		for _, u := range rows {
			es := emailsByUser[u.Idusers]
			if es == nil {
				es = &UserEmailStatus{}
			}
			status := "Active"
			if u.DeletedAt.Valid {
				status = "Inactive"
			}

			item := map[string]interface{}{
				"id":                u.Idusers,
				"username":          u.Username.String,
				"status":            status,
				"roles":             rolesByUser[u.Idusers],
				"verified_emails":   es.Verified,
				"unverified_emails": es.Unverified,
			}
			// Maintain backward compatibility with old fields if possible or just add them
			if c.showAdmin || c.jsonOut {
				item["admin"] = u.Admin
			}
			if c.showCreated || c.jsonOut {
				if t, ok := u.CreatedAt.(sql.NullTime); ok && t.Valid {
					item["created_at"] = t.Time.Format(time.RFC3339)
				}
			}
			out = append(out, item)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	// Header
	fmt.Fprint(w, "ID\tUsername\tStatus\tRoles\tVerified Emails\tUnverified Emails")
	if c.showAdmin {
		fmt.Fprint(w, "\tAdmin")
	}
	if c.showCreated {
		fmt.Fprint(w, "\tCreated")
	}
	fmt.Fprintln(w)

	for _, u := range rows {
		es := emailsByUser[u.Idusers]
		if es == nil {
			es = &UserEmailStatus{}
		}
		status := "Active"
		if u.DeletedAt.Valid {
			status = "Inactive"
		}

		roles := rolesByUser[u.Idusers]

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s",
			u.Idusers,
			u.Username.String,
			status,
			roles,
			strings.Join(es.Verified, ", "),
			strings.Join(es.Unverified, ", "),
		)

		if c.showAdmin {
			fmt.Fprintf(w, "\t%t", u.Admin)
		}
		if c.showCreated {
			if t, ok := u.CreatedAt.(sql.NullTime); ok && t.Valid {
				fmt.Fprintf(w, "\t%s", t.Time.Format(time.RFC3339))
			} else {
				fmt.Fprintf(w, "\t-")
			}
		}
		fmt.Fprintln(w)
	}
	w.Flush()
	return nil
}

func toString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case sql.NullString:
		if t.Valid {
			return t.String
		}
		return ""
	default:
		return fmt.Sprintf("%v", t)
	}
}
