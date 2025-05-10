package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/handlers"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
)

func RegisterPermissionRoutes(r *mux.Router, secretKey string) {
	permissions := r.PathPrefix("/permissions").Subrouter()
	permissions.Use(middleware.AuthMiddleware(secretKey))

	permissions.Handle("/", middleware.PermissionMiddleware(database.DB, "permission:read")(http.HandlerFunc(handlers.PermissionListHandler))).Methods(http.MethodGet)
	permissions.Handle("/{permission_id}", middleware.PermissionMiddleware(database.DB, "permission:read")(http.HandlerFunc(handlers.PermissionDetailsHandler))).Methods(http.MethodGet)
}