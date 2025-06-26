package goa4web

import (
	"context"
	"fmt"
	"log"
	"net/http"

	common "github.com/arran4/goa4web/core/common"
	hcommon "github.com/arran4/goa4web/handlers/common"
)

type CoreData = common.CoreData

// DBAdderMiddleware adds database handles and query helpers to the request context.
func DBAdderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if dbPool == nil {
			ue := common.UserError{Err: fmt.Errorf("db not initialized"), ErrorMessage: "database unavailable"}
			log.Printf("%s: %v", ue.ErrorMessage, ue.Err)
			http.Error(writer, ue.ErrorMessage, http.StatusInternalServerError)
			return
		}
		if dbLogVerbosity > 0 {
			log.Printf("db pool stats: %+v", dbPool.Stats())
		}
		ctx := request.Context()
		ctx = context.WithValue(ctx, hcommon.KeySQLDB, dbPool)
		ctx = context.WithValue(ctx, hcommon.KeyQueries, New(dbPool))
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
