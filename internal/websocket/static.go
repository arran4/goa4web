package websocket

import (
	"bytes"
	"net/http"
	"time"

	"github.com/arran4/goa4web/core/templates"
)

// NotificationsJS serves the JavaScript used for WebSocket notification updates.
func NotificationsJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	http.ServeContent(w, r, "notifications.js", time.Time{}, bytes.NewReader(templates.GetNotificationsJSData()))
}
