package handler

import (
	"speedliner-server/src/middleware"

	"github.com/go-chi/chi/v5"
)

func DefineApiRoutes(r chi.Router) {
	// System/Auth
	r.Get("/ping", PingHandler)
	r.Get("/login", LoginHandler)
	r.Get("/callback", CallbackHandler)
	r.Get("/me", MeHandler)
	r.Get("/logout", LogoutHandler)
	r.Get("/role", GetUserRoleHandler)

	// Routes
	r.Get("/routes", RoutesHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Post("/routes", CreateRouteHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Put("/routes/{id}", UpdateRouteHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Delete("/routes/{id}", DeleteRouteHandler)

	// Users/Corps
	r.With(middleware.RoleMiddleware("admin")).Get("/users", ListUsersHandler)
	r.With(middleware.RoleMiddleware("admin")).Put("/users/{charID}/role", UpdateUserRoleHandler)
	r.With(middleware.RoleMiddleware("admin", "provider")).Get("/corps", ListCorpsHandler)

	// Mail
	r.Post("/mail", SendMailHandler)
	r.Post("/express/mail", SendExpressMailFromServiceHandler)
	r.With(middleware.RoleMiddleware("admin")).Get("/express/token-status", ExpressTokenStatusHandler)
}
