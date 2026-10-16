package websocket

import (
	"github.com/arran4/goa4web/core/templates"
	"github.com/arran4/goa4web/handlers"
)

// LiveEventsJS serves the JavaScript used for WebSocket live-event updates.
var LiveEventsJS = handlers.StaticAssetHandler("live-events.js", "application/javascript", templates.GetLiveEventsJSData)

// NotificationsJS serves the legacy JavaScript endpoint for notification updates.
var NotificationsJS = handlers.StaticAssetHandler("notifications.js", "application/javascript", templates.GetNotificationsJSData)
