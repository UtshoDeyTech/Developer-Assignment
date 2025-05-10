package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/database"
	"github.com/tahsin005/affpilot-auth/internal/http/handlers"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
)

func RegisterUserRoutes(r *mux.Router, secretKey string) {
	users := r.PathPrefix("/users").Subrouter()
	users.Use(middleware.AuthMiddleware(secretKey))

	users.Handle("/", middleware.PermissionMiddleware(database.DB, "user:read:all")(http.HandlerFunc(handlers.UsersListHandler))).Methods(http.MethodGet)

	users.Handle("/{user_id}", middleware.SelfOrAuthorizedMiddleware(database.DB, "user:read:all")(http.HandlerFunc(handlers.UserDetailsHandler))).Methods(http.MethodGet)

	users.Handle("/{user_id}", middleware.SelfOrAuthorizedMiddleware(database.DB, "user:update:all")(http.HandlerFunc(handlers.UserUpdateDetailsHandler))).Methods(http.MethodPut)

	users.Handle("/{user_id}/request-deletion", middleware.SelfOnlyMiddleware()(http.HandlerFunc(handlers.UserDeletionRequestHandler))).Methods(http.MethodPost)

	users.Handle("/{user_id}", middleware.RoleMiddleware("system_admin", "admin", "moderator")(http.HandlerFunc(handlers.UserDeleteHandler))).Methods(http.MethodDelete)

	users.Handle("/{user_id}/role", middleware.RoleMiddleware("system_admin", "admin")(http.HandlerFunc(handlers.UserRoleChangeHandler))).Methods(http.MethodPost)

	users.Handle("/{user_id}/promote/admin", middleware.PermissionMiddleware(database.DB, "user:promote:admin")(http.HandlerFunc(handlers.UserPromoteAdminHandler))).Methods(http.MethodPost)

	users.Handle("/{user_id}/promote/moderator", middleware.PermissionMiddleware(database.DB, "user:promote:moderator")(http.HandlerFunc(handlers.UserPromoteModeratorHandler))).Methods(http.MethodPost)

	users.Handle("/{user_id}/demote", middleware.PermissionMiddleware(database.DB, "user:demote")(http.HandlerFunc(handlers.UserRoleDemoteHandler))).Methods(http.MethodPost)
	
	r.Handle("/me", middleware.AuthMiddleware(secretKey)(http.HandlerFunc(handlers.GetCurrentUser))).Methods(http.MethodGet)
	r.Handle("/me/permissions", middleware.AuthMiddleware(secretKey)(http.HandlerFunc(handlers.CurrentUserPermissions))).Methods(http.MethodGet)
}
