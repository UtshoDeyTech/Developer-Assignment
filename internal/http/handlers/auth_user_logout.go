package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)


func UserLogoutHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    http.SetCookie(w, &http.Cookie{
        Name:     "jwt_token",
        Value:    "",
        Expires:  time.Now().Add(-time.Hour),
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteStrictMode,
        Path:     "/",
    })
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Logout successful",
    })
}