package websocket

import (
	"encoding/json"
	"fmt"
	corecommon "github.com/arran4/goa4web/core/common"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	coreconsts "github.com/arran4/goa4web/core/consts"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/db"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/navigation"
	routerpkg "github.com/arran4/goa4web/internal/router"
	"github.com/arran4/goa4web/internal/tasks"
)

// Module bundles the event bus for websocket handlers.
type Module struct {
	Bus    *eventbus.Bus
	Config *config.RuntimeConfig
}

// NotificationsHandler provides a websocket endpoint streaming bus events.
type NotificationsHandler struct {
	Bus      *eventbus.Bus      // event source
	Upgrader websocket.Upgrader // websocket upgrader
	Config   *config.RuntimeConfig
}

// NewModule returns a websocket module using bus for events.
func NewModule(bus *eventbus.Bus, cfg *config.RuntimeConfig) *Module {
	return &Module{Bus: bus, Config: cfg}
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
func parseHosts(s string) []string {
	var hosts []string
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if u, err := url.Parse(part); err == nil && u.Host != "" {
			hosts = append(hosts, u.Host)
		} else {
			hosts = append(hosts, part)
		}
	}
	return hosts
}

func NewNotificationsHandler(bus *eventbus.Bus, cfg *config.RuntimeConfig) *NotificationsHandler {
	h := &NotificationsHandler{Bus: bus, Config: cfg}
	cfgHosts := parseHosts(cfg.BaseURL)
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		o, err := url.Parse(origin)
		if err != nil {
			return false
		}
		for _, allowed := range cfgHosts {
			if strings.EqualFold(o.Host, allowed) {
				return true
			}
		}
		return strings.EqualFold(o.Host, r.Host)
	}
	h.Upgrader = upgrader
	return h
}

// ServeHTTP upgrades the connection and streams events as JSON.

func (h *NotificationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sess, err := core.GetSession(r)
	if err != nil {
		core.SessionError(w, r, err)
		w.WriteHeader(http.StatusUnauthorized)
		handlers.RenderErrorPage(w, r, fmt.Errorf("invalid session"))
		return
	}
	uid, _ := sess.Values["UID"].(int32)
	if uid == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		handlers.RenderErrorPage(w, r, fmt.Errorf("authentication required"))
		return
	}

	queries := r.Context().Value(coreconsts.KeyCoreData).(*corecommon.CoreData).Queries()
	if queries == nil {
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, fmt.Errorf("db unavailable"))
		return
	}

	ctx := r.Context()

	loadSubs := func() ([]*db.ListSubscriptionsByUserRow, map[string]bool, error) {
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
		w.WriteHeader(http.StatusInternalServerError)
		handlers.RenderErrorPage(w, r, corecommon.ErrInternalServerError)
		return
	}
	if h.Config.LogFlags&config.LogFlagDebug != 0 {
		log.Printf("subscriptions loaded: %d entries", len(subsRows))
	}

	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	ch := h.Bus.Subscribe(eventbus.TaskMessageType)
	for {
		select {
		case env, ok := <-ch:
			if !ok {
				return
			}
			shouldReturn := func() bool {
				defer env.Ack()
				evt, ok := env.Msg.(eventbus.TaskEvent)
				if !ok {
					return false
				}
				if evt.UserID == uid && strings.HasPrefix(evt.Path, "/usr/subscriptions") {
					if n, ok := evt.Task.(tasks.Name); ok {
						if name := n.Name(); name == "Update" || name == "Delete" {
							var err error
							subsRows, patterns, err = loadSubs()
							if err != nil {
								log.Printf("refresh subscriptions: %v", err)
							} else if h.Config.LogFlags&config.LogFlagDebug != 0 {
								log.Printf("subscriptions updated: %d entries", len(subsRows))
							}
							return false
						}
					}
				}
				if evt.UserID == uid {
					return false
				}
				allowed := false
				name := ""
				if n, ok := evt.Task.(tasks.Name); ok {
					name = n.Name()
				}
				for _, p := range buildPatterns(name, evt.Path) {
					if patterns[p] {
						allowed = true
						break
					}
				}
				if !allowed {
					return false
				}
				data, _ := json.Marshal(evt)
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					return true
				}
				return false
			}()
			if shouldReturn {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// registerRoutes attaches the websocket handler to r.
func (m *Module) registerRoutes(r *mux.Router, cfg *config.RuntimeConfig) []navigation.RouterOptions {
	h := NewNotificationsHandler(m.Bus, cfg)
	r.Handle("/ws/notifications", h).Methods(http.MethodGet)
	r.HandleFunc("/websocket/notifications.js", NotificationsJS(cfg)).Methods(http.MethodGet)
	return nil
}

// Register registers the websocket router module.
func (m *Module) Register(reg *routerpkg.Registry) {
	reg.RegisterModule("websocket", nil, m.registerRoutes)
}
