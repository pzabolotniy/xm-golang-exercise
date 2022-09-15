package webapi

import (
	"github.com/go-chi/chi/v5"
	"github.com/pzabolotniy/logging/pkg/logging"
	loggingMW "github.com/pzabolotniy/logging/pkg/middlewares"
)

func CreateRouter(logger logging.Logger, handler *HandlerEnv) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		loggingMW.WithLogger(logger),
		WithLogRequestBoundaries(),
	)

	router.Route("/api/v1", func(apiV1Router chi.Router) {
		apiV1Router.Post("/companies", handler.PostCompanies)
	})

	return router
}
