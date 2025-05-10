package handlers

import (
	"encoding/json"
	"net/http"
)

func CheckHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string {
		"message": "Hurrah! API Seems to be connected... over and out",
	})
}