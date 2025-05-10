package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tahsin005/affpilot-auth/internal/http/handlers"
	"github.com/tahsin005/affpilot-auth/internal/http/middleware"
)

func RegisterAuthRoutes(r *mux.Router, secretKey string) {
	auth := r.PathPrefix("/auth").Subrouter()

	auth.HandleFunc("/register", handlers.UserRegisterHandler).Methods(http.MethodPost)
	auth.HandleFunc("/login", handlers.UserLoginHandler).Methods(http.MethodPost)
	auth.HandleFunc("/verify/{verification_token}", handlers.VerifyEmailHandler).Methods(http.MethodGet)
	auth.HandleFunc("/resend-verification", handlers.ResendVerificationEmail).Methods(http.MethodPost)
	auth.HandleFunc("/password-reset", handlers.CheckHealth).Methods(http.MethodPost)

	auth.Handle("/logout", middleware.AuthMiddleware(secretKey)(http.HandlerFunc(handlers.UserLogoutHandler))).Methods(http.MethodGet)
}
