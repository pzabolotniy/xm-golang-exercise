package webapi

import (
	"fmt"
	"net/http"

	"github.com/pzabolotniy/logging/pkg/logging"
)

func WithLogRequestBoundaries() func(next http.Handler) http.Handler {
	httpMw := func(next http.Handler) http.Handler {
		handlerFn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logging.FromContext(ctx)
			requestURI := r.RequestURI
			requestMethod := r.Method
			logRequest := fmt.Sprintf("%s %s", requestMethod, requestURI)
			logger.WithField("request", logRequest).Trace("REQUEST_STARTED")
			next.ServeHTTP(w, r)
			logger.WithField("request", logRequest).Trace("REQUEST_COMPLETED")
		}

		return http.HandlerFunc(handlerFn)
	}

	return httpMw
}
