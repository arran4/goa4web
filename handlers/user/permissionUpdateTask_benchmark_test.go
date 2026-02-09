package user

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/arran4/goa4web/internal/db"
)

func BenchmarkRoleInfoByPermID(b *testing.B) {
	targetID := int32(5000)

	// Mock the single row return for the optimized query
	row := &db.GetPermissionByIDRow{
		IduserRoles:  targetID,
		UsersIdusers: int32(targetID * 10),
		Name:         fmt.Sprintf("role-%d", targetID),
		Username:     sql.NullString{String: fmt.Sprintf("user-%d", targetID), Valid: true},
	}

	q := &db.QuerierStub{
		GetPermissionByIDRow: row,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, err := roleInfoByPermID(ctx, q, targetID)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
