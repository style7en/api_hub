package models

import (
	"encoding/json"
	"net/http"
	"sort"

	"api-in-one/internal/config"
)

type listResponse struct {
	Object string      `json:"object"`
	Data   []modelInfo `json:"data"`
}

type modelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

func Handler(cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := make([]modelInfo, 0)
		providerNames := make([]string, 0, len(cfg.Providers))
		for name := range cfg.Providers {
			providerNames = append(providerNames, name)
		}
		sort.Strings(providerNames)
		for _, providerName := range providerNames {
			provider := cfg.Providers[providerName]
			models := append([]string(nil), provider.Models...)
			sort.Strings(models)
			for _, model := range models {
				data = append(data, modelInfo{ID: providerName + "/" + model, Object: "model", OwnedBy: providerName})
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listResponse{Object: "list", Data: data})
	})
}
