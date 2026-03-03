package internal

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
	"github.com/spiritov/svelte-go-steam/internal/routes"
)

var (
	api                       huma.API
	sessionCookieSecurityMap  = []map[string][]string{{"Steam": {}}}
	requireSessionMiddlewares huma.Middlewares
	requireModMiddlewares     huma.Middlewares
	requireAdminMiddlewares   huma.Middlewares
	requireDevMiddlewares     huma.Middlewares
)

func setupRouter() *chi.Mux {
	router := chi.NewMux()

	// todo: use strict `AllowedOrigins`
	// todo: use CSRF middleware (?)
	// todo: rate limit (?)
	router.Use(cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowCredentials: true,
		AllowedOrigins:   []string{"*"}, // default value
	}).Handler)

	return router
}

func setupHumaConfig() huma.Config {
	config := huma.DefaultConfig("Site API", "1.0.0")

	// steam security scheme, a JWT with user's OpenID information
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"Steam": {
			Type:        "apiKey",
			In:          "cookie",
			Description: "a session cookie stores the user's session token.",
			Name:        SessionCookieName,
		},
	}
	return config
}

// A readiness endpoint is important - it can be used to inform your infrastructure
// (e.g. fly.io) that the API is available. Readiness checks can help keep your API
// alive, by informing fly on when it should try restarting a machine in case of a
// crash.
func registerHealthCheck(internalApi *huma.Group) {
	type ReadyResponse struct{ OK bool }

	huma.Register(internalApi, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/readyz",
		OperationID: "readyz",
		Summary:     "get readiness",
		Description: "get whether or not the API is ready to process requests",
	}, func(ctx context.Context, _ *struct{}) (*ReadyResponse, error) {
		return &ReadyResponse{OK: true}, nil
	})
}

func ServeAPI(address string) {
	router := setupRouter()
	config := setupHumaConfig()
	api = humachi.New(router, config)

	registerRoutes()

	err := http.ListenAndServe(address, router)
	if err != nil {
		slog.Error("failed to serve api", "error", err)
		log.Fatal()
	}

	slog.Info("serving api", "address", address)
}

func registerRoutes() {
	// create api groups and middlewares
	internalApi := huma.NewGroup(api, "/internal")
	sessionApi := huma.NewGroup(internalApi, "/session")
	devApi := huma.NewGroup(internalApi, "/dev")

	requireSessionMiddlewares = huma.Middlewares{AuthHandler, RequireUserAuthHandler(internalApi)}
	requireDevMiddlewares = huma.Middlewares{AuthHandler, RequireDevHandler(devApi)}

	sessionApi.UseMiddleware(requireSessionMiddlewares...)
	devApi.UseMiddleware(requireDevMiddlewares...)

	// register essential routes
	registerHealthCheck(internalApi)
	registerAuth(internalApi, sessionApi)

	// register all other routes
	routes.RegisterOpenRoutes(internalApi)
	routes.RegisterSessionRoutes(sessionApi)
	routes.RegisterDevRoutes(devApi)
}
