package admin

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	adminapi "github.com/arran4/goa4web/internal/adminapi"
)

// AdminAPIServerShutdown shuts down the server when provided with a valid signed request.
func AdminAPIServerShutdown(w http.ResponseWriter, r *http.Request) {
	const prefix = "Goa4web "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	parts := strings.SplitN(strings.TrimPrefix(auth, prefix), ":", 2)
	if len(parts) != 2 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	ts, sig := parts[0], parts[1]
	signer := adminapi.NewSigner(AdminAPISecret)
	if !signer.Verify(r.Method, r.URL.Path, ts, sig) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := Srv.Shutdown(ctx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()
	w.WriteHeader(http.StatusOK)
}
