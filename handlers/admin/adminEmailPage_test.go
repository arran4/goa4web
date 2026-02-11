package admin

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestAdminEmailPage_PopulatesUsers(t *testing.T) {
	queries := testhelpers.NewQuerierStub()
	cd := common.NewCoreData(context.Background(), queries, config.NewRuntimeConfig())

	// Mock AdminCountUnsentPendingEmails
	queries.AdminCountUnsentPendingEmailsReturns = 1

	// Mock AdminListUnsentPendingEmails
	queries.AdminListUnsentPendingEmailsReturns = []*db.AdminListUnsentPendingEmailsRow{
		{
			ID:          1,
			ToUserID:    sql.NullInt32{Int32: 123, Valid: true},
			Body:        "body",
			CreatedAt:   time.Now(),
			DirectEmail: false,
		},
	}

	// Mock SystemGetUsersByIDs
	called := false
	queries.SystemGetUsersByIDsFn = func(ctx context.Context, ids []int32) ([]*db.SystemGetUsersByIDsRow, error) {
		called = true
		assert.Equal(t, []int32{123}, ids)
		return []*db.SystemGetUsersByIDsRow{
			{
				Idusers:  123,
				Username: sql.NullString{String: "testuser", Valid: true},
				Email:    sql.NullString{String: "testuser@example.com", Valid: true},
			},
		}, nil
	}

	req := httptest.NewRequest("GET", "/admin/email/queue", nil)
	req = req.WithContext(context.WithValue(req.Context(), consts.KeyCoreData, cd))
	w := httptest.NewRecorder()

	AdminEmailPage(w, req)

	assert.True(t, called, "SystemGetUsersByIDs should be called")
}
