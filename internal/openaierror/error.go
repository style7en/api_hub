package openaierror

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Error detail `json:"error"`
}

type detail struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func Write(w http.ResponseWriter, status int, errorType string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response{Error: detail{Message: message, Type: errorType}})
}
