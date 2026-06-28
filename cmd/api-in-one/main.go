package main

import (
	"flag"
	"log"
	"net/http"

	"api-in-one/internal/config"
	"api-in-one/internal/server"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	handler := server.New(cfg)
	log.Printf("api-in-one listening on %s", cfg.Server.Address)
	if err := http.ListenAndServe(cfg.Server.Address, handler); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
