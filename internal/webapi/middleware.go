package webapi

import (
	"fmt"
	"net"
	"net/http"

	"github.com/pzabolotniy/logging/pkg/logging"

	"github.com/pzabolotniy/xm-golang-exercise/internal/geoip"
)

func WithLogRequestBoundaries() func(next http.Handler) http.Handler {
	httpMw := func(next http.Handler) http.Handler {
		handlerFn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logging.FromContext(ctx)
			requestURI := r.URL.Path
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

func WithCountryRestriction(
	geoIPService geoip.CountryDetector, allowedCountryName string,
) func(next http.Handler) http.Handler {
	httpMw := func(next http.Handler) http.Handler {
		handlerFn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			logger := logging.FromContext(ctx)
			remoteAddr := r.RemoteAddr
			clientIP, _, err := net.SplitHostPort(remoteAddr)
			if err != nil {
				logger.WithError(err).Error("split remote addr ip failed")
				InternalServerError(ctx, w, "verify client country failed")

				return
			}

			countryName, err := geoIPService.CountryByIP(ctx, clientIP)
			if err != nil {
				logger.WithError(err).Error("detect country failed")
				InternalServerError(ctx, w, "verify client country failed")

				return
			}

			if countryName != allowedCountryName {
				logger.
					WithFields(logging.Fields{
						"client_country_name":  countryName,
						"allowed_country_name": allowedCountryName,
					}).
					Error("country mismatch")
				Forbidden(ctx, w, "access denied")

				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(handlerFn)
	}

	return httpMw
}
