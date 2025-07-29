# Notifications

Goa4Web exposes a WebSocket endpoint at `/ws/notifications`. Clients must
include their session cookie when connecting. Events published on the
server's event bus are delivered as JSON only when the connected user is
subscribed to the matching event pattern.
