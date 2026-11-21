package websocket

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core"
	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/eventbus"
	"github.com/arran4/goa4web/internal/navigation"
	routerpkg "github.com/arran4/goa4web/internal/router"
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
}

// LiveEventKindNotification asks the browser to re-fetch its own inbox fragments.
const LiveEventKindNotification = "notification"

// LiveEvent is a browser invalidation signal. The HTML endpoint supplies the data.
type LiveEvent struct {
	Kind string    `json:"kind"`
	Time time.Time `json:"time"`
}

// NewModule returns a websocket module using bus for events.
func NewModule(bus *eventbus.Bus, cfg *config.RuntimeConfig) *Module {
	return &Module{Bus: bus, Config: cfg}
}

func parseHosts(s string) []string {
	var hosts []string
	for part := range strings.SplitSeq(s, ",") {
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
	h := &NotificationsHandler{Bus: bus}
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

	if h.Bus == nil {
		http.Error(w, "live updates unavailable", http.StatusServiceUnavailable)
		return
	}
	sub := h.Bus.SubscribeWithOptions(
		eventbus.WithTypes(eventbus.NotificationsChangedMessageType),
		eventbus.WithBufferSize(1),
		eventbus.WithFilter(func(msg eventbus.Message) bool {
			evt, ok := msg.(eventbus.NotificationsChanged)
			return ok && evt.RecipientID == uid
		}),
	)
	defer sub.Close()

	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade: %v", err)
		return
	}
	defer func() { _ = conn.Close() }()

	// Read control frames so closed clients release their subscriptions.
	disconnected := make(chan struct{})
	conn.SetReadLimit(1024)
	go func() {
		defer close(disconnected)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case env, ok := <-sub.Channel():
			if !ok {
				return
			}
			evt := env.Msg.(eventbus.NotificationsChanged)
			err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			if err == nil {
				err = conn.WriteJSON(LiveEvent{Kind: LiveEventKindNotification, Time: evt.Time})
			}
			env.Ack()
			if err != nil {
				return
			}
		case <-disconnected:
			return
		case <-r.Context().Done():
			return
		}
	}
}

// registerRoutes attaches the websocket handler to r.
func (m *Module) registerRoutes(r *mux.Router, cfg *config.RuntimeConfig) []navigation.RouterOptions {
	h := NewNotificationsHandler(m.Bus, cfg)
	r.Handle("/ws/events", h).Methods(http.MethodGet)
	r.Handle("/ws/notifications", h).Methods(http.MethodGet)
	r.HandleFunc("/websocket/live-events.js", LiveEventsJS(cfg)).Methods(http.MethodGet)
	r.HandleFunc("/websocket/notifications.js", NotificationsJS(cfg)).Methods(http.MethodGet)
	return nil
}

// Register registers the websocket router module.
func (m *Module) Register(reg *routerpkg.Registry) {
	reg.RegisterModule("websocket", nil, m.registerRoutes)
}
