package common_test

import (
	"context"
	"testing"

	"github.com/arran4/goa4web/core/common"
	"github.com/gorilla/sessions"
	"github.com/arran4/goa4web/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestSessionManagerInterface(t *testing.T) {
	// Ensure SessionProxy implements SessionManager.
	var _ common.SessionManager = (*db.SessionProxy)(nil)
}

func TestCoreDataSession(t *testing.T) {
	store := sessions.NewCookieStore([]byte("secret"))
	session := sessions.NewSession(store, "test-session")
	session.Values["UID"] = int32(123)

	t.Run("WithSession", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), nil, nil, common.WithSession(session))
		assert.Equal(t, session, cd.Session())
		assert.Equal(t, int32(123), cd.UserID)
		assert.Equal(t, session, cd.GetSession())
	})

	t.Run("SetSession", func(t *testing.T) {
		cd := common.NewCoreData(context.Background(), nil, nil)
		assert.Nil(t, cd.Session())
		assert.Nil(t, cd.GetSession())

		cd.SetSession(session)
		assert.Equal(t, session, cd.Session())
		assert.Equal(t, session, cd.GetSession())
	})
}
