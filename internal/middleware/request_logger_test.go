package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arran4/goa4web/config"
	"github.com/arran4/goa4web/core/common"
	"github.com/arran4/goa4web/core/consts"
	"github.com/gorilla/sessions"
)

func TestRequestLoggerMiddleware(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	// There is no GetOutput, but we can assume it was os.Stderr or we don't care to restore exactly what it was for this test process,
	// but good practice is to restore. We can't easily get the original writer if it wasn't set explicitly via SetOutput before.
	// However, log.SetOutput returns nothing.
	// We will just set it to buf, run tests, and then maybe not worry about restoring since it's a test process that will exit?
	// But other tests might run.
	// We can rely on log.Writer() if available.

	// Note: checks for log.Writer() availability... it is available since Go 1.14.
	orig := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(orig)

	tests := []struct {
		name          string
		logFlags      int
		path          string
		uid           int32
		sessionID     string
		expectLog     bool
		expectSession bool
	}{
		{
			name:      "No logging when flags are zero",
			logFlags:  0,
			path:      "/test",
			expectLog: false,
		},
		{
			name:          "Logging enabled with LogFlagDebug",
			logFlags:      config.LogFlagDebug,
			path:          "/test",
			expectLog:     true,
			expectSession: false,
		},
		{
			name:          "Logging with session",
			logFlags:      config.LogFlagDebug,
			path:          "/test",
			sessionID:     "sess123",
			expectLog:     true,
			expectSession: true,
		},
		{
			name:      "Skip notification polling for uid 0",
			logFlags:  config.LogFlagDebug,
			path:      "/ws/notifications",
			uid:       0,
			expectLog: false,
		},
		{
			name:          "Log notification polling for uid > 0",
			logFlags:      config.LogFlagDebug,
			path:          "/ws/notifications",
			uid:           1,
			expectLog:     true,
			expectSession: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()

			cd := &common.CoreData{
				Config: &config.RuntimeConfig{
					LogFlags: tt.logFlags,
				},
				UserID: tt.uid,
			}

			if tt.sessionID != "" {
				s := sessions.NewSession(nil, "test")
				s.ID = tt.sessionID
				cd.SetSession(s)
			}

			// Handlers
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			handler := RequestLoggerMiddleware(next)

			req := httptest.NewRequest("GET", tt.path, nil)

			// Inject CoreData into context
			ctx := context.WithValue(req.Context(), consts.KeyCoreData, cd)
			req = req.WithContext(ctx)

			handler.ServeHTTP(httptest.NewRecorder(), req)

			output := buf.String()
			if tt.expectLog {
				if output == "" {
					t.Errorf("Expected log output, got none")
				}
				if tt.expectSession {
					if !strings.Contains(output, "session="+tt.sessionID) {
						t.Errorf("Expected session=%s in log, got: %s", tt.sessionID, output)
					}
				} else {
					if strings.Contains(output, "session=") {
						t.Errorf("Expected no session in log, got: %s", output)
					}
				}
			} else {
				if output != "" {
					t.Errorf("Expected no log output, got: %s", output)
				}
			}
		})
	}
}
