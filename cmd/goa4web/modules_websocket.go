package main

import (
	websocket "github.com/arran4/goa4web/internal/websocket"
)

func init() {
	extraRegistrations = append(extraRegistrations, websocket.Register)
}
