package webapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/pzabolotniy/logging/pkg/logging"
	loggingMW "github.com/pzabolotniy/logging/pkg/middlewares"

	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
	"github.com/pzabolotniy/xm-golang-exercise/internal/geoip"
)

func CreateRouter(
	logger logging.Logger, handler *HandlerEnv,
	countryDetector geoip.CountryDetector, geoIPConf *config.GeoIP,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(
		loggingMW.WithLogger(logger),
		WithLogRequestBoundaries(),
		WithCountryRestriction(countryDetector, geoIPConf.AllowedCountryName),
	)

	router.Route("/api/v1", func(apiV1Router chi.Router) {
		apiV1Router.Post("/companies", handler.PostCompanies)
	})

	return router
}
