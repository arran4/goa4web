package blogs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arran4/goa4web/handlers/handlertest"
	"github.com/stretchr/testify/assert"
)

func TestBlogAddPage_AccessDenied(t *testing.T) {
	req := httptest.NewRequest("GET", "/blogs/add", nil)
	// Create CoreData with default settings (no admin, no grants)
	req, cd, _ := handlertest.RequestWithCoreData(t, req)
	cd.AdminMode = false // Ensure explicit non-admin mode

	w := httptest.NewRecorder()

	BlogAddPage(w, req)

	// Check response status
	assert.Equal(t, http.StatusForbidden, w.Code)
}
