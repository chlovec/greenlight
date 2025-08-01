package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// Use the requirePermission() middleware on each of the /v1/movies** endpoints,
	// passing in the required permission code as the first parameter.
	router.HandlerFunc(
		http.MethodGet,
		"/v1/movies",
		app.requirePermission("movies:read", app.listMoviesHandler),
	)
	router.HandlerFunc(
		http.MethodPost,
		"/v1/movies",
		app.requirePermission("movies:write", app.createMovieHandler),
	)
	router.HandlerFunc(
		http.MethodGet,
		"/v1/movies/:id",
		app.requirePermission("movies:read", app.showMovieHandler),
	)
	router.HandlerFunc(
		http.MethodPost,
		"/v1/movies/:id",
		app.requirePermission("movies:write", app.updateMovieHandler),
	)
	router.HandlerFunc(
		http.MethodPatch,
		"/v1/movies/:id",
		app.requirePermission("movies:write", app.patchMovieHandler),
	)
	router.HandlerFunc(
		http.MethodDelete,
		"/v1/movies/:id",
		app.requirePermission("movies:write", app.deleteMovieHandler),
	)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(
		http.MethodPost,
		"/v1/tokens/authentication",
		app.createAuthenticationTokenHandler,
	)

	// Register a new GET /debug/vars endpoint pointing to the expvar handler.
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// Use the new metrics() middleware at the start of the chain.
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
