package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/http/handlers"
)

func RegisterCheckHealthRoutes(r *mux.Router) {
	r.HandleFunc("/health", handlers.CheckHealth).Methods(http.MethodGet)
}