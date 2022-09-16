package webapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/pzabolotniy/logging/pkg/logging"
	loggingMW "github.com/pzabolotniy/logging/pkg/middlewares"

	"github.com/pzabolotniy/xm-golang-exercise/internal/authn"
	"github.com/pzabolotniy/xm-golang-exercise/internal/config"
	"github.com/pzabolotniy/xm-golang-exercise/internal/geoip"
)

type RouterParams struct {
	Logger          logging.Logger
	Handler         *HandlerEnv
	CountryDetector geoip.CountryDetector
	GeoIPConf       *config.GeoIP
	TokenService    authn.TokenValidator
}

func CreateRouter(params *RouterParams) *chi.Mux {
	logger := params.Logger
	tokenService := params.TokenService
	countryDetector := params.CountryDetector
	geoIPConf := params.GeoIPConf
	handler := params.Handler

	router := chi.NewRouter()
	router.Use(
		loggingMW.WithLogger(logger),
		WithXRequestID,
		WithLogRequestBoundaries(),
	)

	router.Route("/api/v1", func(apiV1Router chi.Router) {
		apiV1Router.Route("/companies", func(companiesRouter chi.Router) {
			companiesRouter.Group(func(restrictedRouter chi.Router) {
				restrictedRouter.Use(
					WithAuthN(tokenService),
					WithCountryRestriction(countryDetector, geoIPConf.AllowedCountryName),
				)
				restrictedRouter.Post("/", handler.PostCompanies)
				restrictedRouter.Delete("/{companyID}", handler.DeleteCompany)
			})
		})
	})

	return router
}
