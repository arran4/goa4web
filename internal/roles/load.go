package roles

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/goa4web/internal/sqlutil"
)

// ReadRoleSQL loads role SQL content from an explicit file or embedded resources.
func ReadRoleSQL(roleName string, explicitFile string) ([]byte, error) {
	if explicitFile != "" {
		p := explicitFile
		if !strings.HasSuffix(strings.ToLower(p), ".sql") {
			p = p + ".sql"
		}
		abs, _ := filepath.Abs(p)
		log.Printf("Loading role %q from file %s", roleName, abs)
		b, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("failed to read role file: %w", err)
		}
		return b, nil
	}

	log.Printf("Loading role %q from embedded roles", roleName)
	b, err := ReadEmbeddedRole(roleName)
	if err != nil {
		available, listErr := ListEmbeddedRoles()
		if listErr == nil && len(available) > 0 {
			return nil, fmt.Errorf("failed to read embedded role %q: %w. Available roles: %s", roleName, err, strings.Join(available, ", "))
		}
		return nil, fmt.Errorf("failed to read embedded role %q: %w", roleName, err)
	}
	return b, nil
}

// ApplyRoleSQL runs role SQL statements against the database.
func ApplyRoleSQL(ctx context.Context, roleName string, data []byte, sdb *sql.DB) error {
	if err := sqlutil.RunStatements(ctx, sdb, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("failed to apply role %q: %w", roleName, err)
	}
	log.Printf("Role %q loaded successfully.", roleName)
	return nil
}

// LoadRole loads a role from embedded SQL or an explicit file into the database.
func LoadRole(ctx context.Context, roleName string, explicitFile string, sdb *sql.DB) error {
	data, err := ReadRoleSQL(roleName, explicitFile)
	if err != nil {
		return err
	}
	return ApplyRoleSQL(ctx, roleName, data, sdb)
}
