package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
)

//go:embed frontend/*
var frontendEmbed embed.FS

func SetupRoutes() {
	// Serve static files from the embedded frontend folder
	subFS, err := fs.Sub(frontendEmbed, "frontend")
	if err != nil {
		log.Fatalf("Failed to sub embedded frontend filesystem: %v", err)
	}
	http.Handle("/", http.FileServer(http.FS(subFS)))

	apiRouter := http.NewServeMux()
	
	// API Endpoints
	apiRouter.HandleFunc("/api/config", handleConfig)
	apiRouter.HandleFunc("/api/ha/entities", handleHAEntities)

	// Combine into main mux
	http.Handle("/api/", apiRouter)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		configMut.RLock()
		defer configMut.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AppConfig)
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut {
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		configMut.Lock()
		AppConfig = newConfig
		configMut.Unlock()

		if err := SaveConfig(); err != nil {
			http.Error(w, "Failed to save config: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Re-initialize MQTT with new settings
		InitMQTT()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleHAEntities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entities, err := getHAEntities()
	if err != nil {
		http.Error(w, "Failed to get HA entities: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entities)
}
