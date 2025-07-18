//go:build websocket

package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/arran4/goa4web/core"
	corecommon "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
	dbpkg "github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	routerpkg "github.com/arran4/goa4web/internal/router"
)

// NotificationsHandler provides a websocket endpoint streaming bus events.
type NotificationsHandler struct {
	Bus      *eventbus.Bus      // event source
	Upgrader websocket.Upgrader // websocket upgrader
}

func buildPatterns(task, path string) []string {
	name := strings.ToLower(task)
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{fmt.Sprintf("%s:/*", name)}
	}
	parts := strings.Split(path, "/")
	patterns := []string{fmt.Sprintf("%s:/%s", name, path)}
	for i := len(parts) - 1; i >= 1; i-- {
		prefix := strings.Join(parts[:i], "/")
		patterns = append(patterns, fmt.Sprintf("%s:/%s/*", name, prefix))
	}
	patterns = append(patterns, fmt.Sprintf("%s:/*", name))
	return patterns
}

// NewNotificationsHandler returns a handler using bus for events.
func NewNotificationsHandler(bus *eventbus.Bus) *NotificationsHandler {
	return &NotificationsHandler{
		Bus:      bus,
		Upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
}

// ServeHTTP upgrades the connection and streams events as JSON.

func (h *NotificationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess, err := core.GetSession(r)
	if err != nil {
		core.SessionError(w, r, err)
		http.Error(w, "invalid session", http.StatusUnauthorized)
		return
	}
	uid, _ := sess.Values["UID"].(int32)
	if uid == 0 {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}

	queries, ok := r.Context().Value(corecommon.KeyQueries).(*dbpkg.Queries)
	if !ok || queries == nil {
		http.Error(w, "db unavailable", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	loadSubs := func() ([]*dbpkg.ListSubscriptionsByUserRow, map[string]bool, error) {
		rows, err := queries.ListSubscriptionsByUser(ctx, uid)
		if err != nil {
			return nil, nil, err
		}
		p := make(map[string]bool)
		for _, row := range rows {
			if row.Method == "internal" {
				p[row.Pattern] = true
			}
		}
		return rows, p, nil
	}

	subsRows, patterns, err := loadSubs()
	if err != nil {
		log.Printf("list subscriptions: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("subscriptions loaded: %d entries", len(subsRows))

	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	ch := h.Bus.Subscribe()
	for {
		select {
		case evt := <-ch:
			if evt.UserID == uid && strings.HasPrefix(evt.Path, "/usr/subscriptions") &&
				(evt.Task == hcommon.TaskUpdate || evt.Task == hcommon.TaskDelete) {
				var err error
				subsRows, patterns, err = loadSubs()
				if err != nil {
					log.Printf("refresh subscriptions: %v", err)
				} else {
					log.Printf("subscriptions updated: %d entries", len(subsRows))
				}
				continue
			}
			if evt.UserID == uid {
				continue
			}
			allowed := false
			for _, p := range buildPatterns(evt.Task, evt.Path) {
				if patterns[p] {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
			data, _ := json.Marshal(evt)
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// RegisterRoutes attaches the websocket handler to r.
func RegisterRoutes(r *mux.Router) {
	h := NewNotificationsHandler(eventbus.DefaultBus)
	r.Handle("/ws/notifications", h).Methods(http.MethodGet)
	r.HandleFunc("/notifications.js", NotificationsJS).Methods(http.MethodGet)
}

// Register registers the websocket router module.
func Register() {
	routerpkg.RegisterModule("websocket", nil, RegisterRoutes)
}
