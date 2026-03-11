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

func ingressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ingressPath := r.Header.Get("X-Hass-Ingress-Path")
		if ingressPath != "" {
			// Strip the ingress path from the URL
			http.StripPrefix(ingressPath, next).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SetupRoutes() {
	mux := http.NewServeMux()

	// API Endpoints – ezeket ELŐBB kell regisztrálni, mint a "/" catch-all-t
	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/ha/entities", handleHAEntities)
	mux.HandleFunc("/api/mqtt/test", handleMQTTTest)

	// Serve static files from the embedded frontend folder – "/" catch-all, UTOLJÁRA
	subFS, err := fs.Sub(frontendEmbed, "frontend")
	if err != nil {
		log.Fatalf("Failed to sub embedded frontend filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(subFS)))

	// Wrap the mux with the ingress middleware
	http.Handle("/", ingressMiddleware(mux))
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

		w.Header().Set("Content-Type", "application/json")
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
	opts.SetConnectTimeout(5 * time.Second)

	if testConfig.MqttSSL {
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()

	w.Header().Set("Content-Type", "application/json")

	if token.Error() != nil {
		// Biztonságos JSON encoding – nincs string concatenation!
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": token.Error().Error(),
		})
		return
	}

	// Siker, bontjuk a kapcsolatot
	client.Disconnect(250)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
