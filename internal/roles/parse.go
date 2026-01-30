package roles

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/arran4/goa4web/internal/db"
)

// ParseRoleName extracts the role name from role SQL content.
func ParseRoleName(data []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(strings.ToLower(line), "-- role:") {
			name := strings.TrimSpace(line[len("-- role:"):])
			if name != "" {
				return name, nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan role comments: %w", err)
	}

	nameRegexp := regexp.MustCompile(`(?is)insert\s+into\s+roles\s*\([^)]*name[^)]*\)\s*values\s*\(\s*'([^']+)'`)
	if matches := nameRegexp.FindSubmatch(data); len(matches) > 1 {
		return string(matches[1]), nil
	}
	return "", fmt.Errorf("role name not found")
}

// ParseRoleGrants extracts grant definitions from role SQL content.
func ParseRoleGrants(data []byte) ([]*db.Grant, error) {
	grantRegexp := regexp.MustCompile(`(?is)insert\s+into\s+grants\s*\([^)]*section[^)]*item[^)]*rule_type[^)]*action[^)]*active[^)]*\)\s*select\s+now\(\)\s*,\s*r\.id\s*,\s*'([^']*)'\s*,\s*(null|'([^']*)')\s*,\s*'([^']*)'\s*,\s*'([^']*)'\s*,\s*([0-9]+)`)
	matches := grantRegexp.FindAllSubmatch(data, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no grants found")
	}
	grants := make([]*db.Grant, 0, len(matches))
	for i, match := range matches {
		section := string(match[1])
		item := strings.TrimSpace(string(match[3]))
		ruleType := string(match[4])
		action := string(match[5])
		activeInt, err := strconv.Atoi(string(match[6]))
		if err != nil {
			return nil, fmt.Errorf("grant %d active flag: %w", i+1, err)
		}
		grants = append(grants, &db.Grant{
			ID:       int32(i + 1),
			Section:  section,
			Item:     sql.NullString{String: item, Valid: item != ""},
			RuleType: ruleType,
			Action:   action,
			Active:   activeInt != 0,
		})
	}
	return grants, nil
}
