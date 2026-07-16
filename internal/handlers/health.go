package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthResponce struct {
	Status string `json:""status"`
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthResponce{
		Status: "ok",
	}

	json.NewEncoder(w).Encode(response)
}
