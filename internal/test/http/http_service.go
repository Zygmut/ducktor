package main

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{Status: "healthy"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.ListenAndServe(":8001", nil)
}
