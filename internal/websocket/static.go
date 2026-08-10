package websocket

import (
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
)

// NotificationsJS serves the JavaScript used for WebSocket notification updates.
var NotificationsJS = handlers.StaticAssetHandler("notifications.js", "application/javascript", templates.GetNotificationsJSData)
