package admin

import (
	"fmt"
	"github.com/arran4/goa4web/internal/tasks"
	"net/http"

	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
)

const AdminPageTmpl tasks.Template = "admin/page.gohtml"
