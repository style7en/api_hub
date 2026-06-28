package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"api_hub/internal/config"
	"api_hub/internal/gui"
	"api_hub/internal/server"
)

func main() {
	mode := flag.String("mode", "gui", "run mode: gui or api")
	configPath := flag.String("config", "config.yaml", "path to config file")
	guiListen := flag.String("gui-listen", "127.0.0.1:8090", "GUI listen address")
	openBrowser := flag.Bool("open-browser", true, "open browser on startup")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	switch *mode {
	case "api":
		runAPI(cfg)
	default:
		runGUI(cfg, *configPath, *guiListen, *openBrowser)
	}
}

func runAPI(cfg *config.Config) {
	handler := server.New(cfg)
	log.Printf("API Hub listening on %s", cfg.Server.Address)
	if err := http.ListenAndServe(cfg.Server.Address, handler); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func runGUI(cfg *config.Config, configPath, guiListen string, openBrowser bool) {
	manager := gui.NewRuntimeManager(configPath, cfg)
	handler := gui.NewServer(configPath, cfg, manager)
	guiURL := fmt.Sprintf("http://%s", guiListen)
	log.Printf("API Hub GUI listening on %s", guiURL)
	if openBrowser {
		openURL(guiURL)
	}
	if err := http.ListenAndServe(guiListen, handler); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func openURL(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("failed to open browser: %v", err)
	}
}
