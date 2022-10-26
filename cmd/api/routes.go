package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"
)

// routes defines all the routes and hanlders in our application
func (app *Application) routes() http.Handler {
	router := chi.NewRouter()

	router.NotFound(http.HandlerFunc(app.NotFoundResponse))
	router.MethodNotAllowed(http.HandlerFunc(app.MethodNotAllowedResponse))

	router.Use(app.RecoverPanic)
	// router.Use(app.HTTPMetrics(app.Config.ServiceName))
	router.Use(otelchi.Middleware(app.Config.ServiceName, otelchi.WithChiRoutes(router)))
	router.Use(app.LogRequest)
	router.Use(app.SecureHeaders)

	router.Get("/healthcheck", app.healthCheckHandler)

	router.Route("/items", func(r chi.Router) {
		r.Use(app.Authenticate(app.UsersRepository, app.Config.RSA.PublicKey))

		r.With(app.RequirePermission(app.UsersRepository, "inventory:read")).Get("/", app.getInventoryItemsHandler)
		r.With(app.RequirePermission(app.UsersRepository, "inventory:write")).Post("/", app.grantItemsHandler)
	})

	router.Get("/metrics", promhttp.Handler().ServeHTTP)

	return router
}
