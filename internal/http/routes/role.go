package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/handlers"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
)

func RegisterRoleRoutes(r *mux.Router, secretKey string) {
	roles := r.PathPrefix("/roles").Subrouter()
	roles.Use(middleware.AuthMiddleware(secretKey))

	roles.Handle("/", middleware.PermissionMiddleware(database.DB, "role:read")(http.HandlerFunc(handlers.RoleListHandler))).Methods(http.MethodGet)
	roles.Handle("/{role_id}", middleware.PermissionMiddleware(database.DB, "role:read")(http.HandlerFunc(handlers.RoleDetailsHandler))).Methods(http.MethodGet)
	roles.Handle("/", middleware.PermissionMiddleware(database.DB, "role:create")(http.HandlerFunc(handlers.RoleCreateHandler))).Methods(http.MethodPost)
	roles.Handle("/{role_id}", middleware.PermissionMiddleware(database.DB, "role:update")(http.HandlerFunc(handlers.RoleUpdateHandler))).Methods(http.MethodPut)
	roles.Handle("/{role_id}", middleware.PermissionMiddleware(database.DB, "role:delete")(http.HandlerFunc(handlers.RoleDeleteHandler))).Methods(http.MethodDelete)
}