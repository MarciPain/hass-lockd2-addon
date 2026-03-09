package main

import (
	"crypto/tls"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
	apiRouter.HandleFunc("/api/mqtt/test", handleMQTTTest)

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

func handleMQTTTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var testConfig Config
	if err := json.NewDecoder(r.Body).Decode(&testConfig); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if testConfig.MqttHost == "" {
		http.Error(w, "Host is required", http.StatusBadRequest)
		return
	}


	brokerURL := "tcp://" + testConfig.MqttHost + ":" + fmt.Sprint(testConfig.MqttPort)
	if testConfig.MqttSSL {
		brokerURL = "tls://" + testConfig.MqttHost + ":" + fmt.Sprint(testConfig.MqttPort)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID("lockd2-addon-test")
	opts.SetUsername(testConfig.MqttUser)
	opts.SetPassword(testConfig.MqttPass)
	// Rövid timeout a teszthez
	opts.SetConnectTimeout(3 * time.Second)

	if testConfig.MqttSSL {
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"status": "error", "message": "` + token.Error().Error() + `"}`))
		return
	}

	// Siker, bontjuk a kapcsolatot
	client.Disconnect(250)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}
