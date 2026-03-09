package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// A config fájl elérési útja (HA addonoknál általában a /data mappában érdemes menteni)
const configFilePath = "/data/lockd2_config.json"
// Helyi teszteléshez
const localConfigFilePath = "lockd2_config.json"

type LockEntity struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	EntityID      string `json:"entity_id"`
	BatteryEntity string `json:"battery_entity,omitempty"`
	TopicSuffix   string `json:"topic_suffix"`
	Enabled       bool   `json:"enabled"`
	Mode          string `json:"mode,omitempty"`     // toggle vagy pulse (kapcsolóknál)
	PulseDuration int    `json:"pulse_duration,omitempty"` // másodpercben
}

type Config struct {
	MqttHost string       `json:"mqtt_host"`
	MqttPort int          `json:"mqtt_port"`
	MqttUser string       `json:"mqtt_user"`
	MqttPass string       `json:"mqtt_pass"`
	MqttSSL  bool         `json:"mqtt_ssl"`
	Locks    []LockEntity `json:"locks"`
}

var (
	AppConfig Config
	configMut sync.RWMutex
)

// LoadConfig betölti a beállításokat a fájlból
func LoadConfig() {
	configMut.Lock()
	defer configMut.Unlock()

	path := configFilePath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = localConfigFilePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("No existing config found, starting fresh.")
		AppConfig = Config{Locks: []LockEntity{}}
		return
	}

	if err := json.Unmarshal(data, &AppConfig); err != nil {
		log.Printf("Error parsing config: %v", err)
	}
}

// SaveConfig elmenti a beállításokat a fájlba
func SaveConfig() error {
	configMut.RLock()
	defer configMut.RUnlock()

	data, err := json.MarshalIndent(AppConfig, "", "  ")
	if err != nil {
		return err
	}

	path := configFilePath
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		path = localConfigFilePath // fallback lokalba
	}

	return os.WriteFile(path, data, 0644)
}
