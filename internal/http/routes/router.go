package routes

import (
	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/config"
)

func RegisterRoutes() *mux.Router {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	v1 := api.PathPrefix("/v1").Subrouter()

	cfg := config.LoadConfig()

	RegisterCheckHealthRoutes(v1)
	RegisterAuthRoutes(v1, cfg.JWT_SECRET)
	RegisterUserRoutes(v1, cfg.JWT_SECRET)
	RegisterRoleRoutes(v1, cfg.JWT_SECRET)
	RegisterPermissionRoutes(v1, cfg.JWT_SECRET)

	return r
}