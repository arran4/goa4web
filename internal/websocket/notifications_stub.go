//go:build !websocket

package websocket

import "github.com/arran4/goa4web/internal/eventbus"

// SetBus is a no-op when websocket build tags are disabled.
func SetBus(_ *eventbus.Bus) {}
