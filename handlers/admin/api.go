package admin

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/arran4/goa4web/handlers"
	"github.com/arran4/goa4web/internal/adminapi"
)

// AdminAPIServerShutdown shuts down the server when provided with a valid signed request.
func (h *Handlers) AdminAPIServerShutdown(w http.ResponseWriter, r *http.Request) {
	const prefix = "Goa4web "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Unauthorized"))
		return
	}
	parts := strings.SplitN(strings.TrimPrefix(auth, prefix), ":", 2)
	if len(parts) != 2 {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Unauthorized"))
		return
	}
	ts, sig := parts[0], parts[1]
	signer := adminapi.NewSigner(AdminAPISecret)
	if !signer.Verify(r.Method, r.URL.Path, ts, sig) {
		handlers.RenderErrorPage(w, r, fmt.Errorf("Unauthorized"))
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := h.Srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()
	w.WriteHeader(http.StatusOK)
}
