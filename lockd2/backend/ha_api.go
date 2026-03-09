package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var haToken string

func InitHAAPI() {
	haToken = os.Getenv("SUPERVISOR_TOKEN")
	if haToken == "" {
		log.Println("WARNING: SUPERVISOR_TOKEN is not set. HA Entity discovery will not work outside of addon container.")
	}
}

// HAEntity represents a simple Home Assistant state entity
type HAEntity struct {
	EntityID string `json:"entity_id"`
	State    string `json:"state"`
	Attributes struct {
		FriendlyName string `json:"friendly_name,omitempty"`
	} `json:"attributes,omitempty"`
}

func getHAEntities() ([]HAEntity, error) {
	if haToken == "" {
		// Mock data for local testing outside container
		return []HAEntity{
			{EntityID: "lock.front_door", Attributes: struct{FriendlyName string `json:"friendly_name,omitempty"`}{FriendlyName: "Front Door"}},
			{EntityID: "switch.gate", Attributes: struct{FriendlyName string `json:"friendly_name,omitempty"`}{FriendlyName: "Gate Relay"}},
			{EntityID: "sensor.front_door_battery", Attributes: struct{FriendlyName string `json:"friendly_name,omitempty"`}{FriendlyName: "Front Door Battery"}},
		}, nil
	}

	url := "http://supervisor/core/api/states"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+haToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Superivsor API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var entities []HAEntity
	if err := json.Unmarshal(body, &entities); err != nil {
		return nil, err
	}

	// Filter down to relevant entities for our UI to make payload smaller
	var filtered []HAEntity
	for _, e := range entities {
		if hasPrefix(e.EntityID, "lock.") || hasPrefix(e.EntityID, "switch.") || hasPrefix(e.EntityID, "sensor.") {
			filtered = append(filtered, e)
		}
	}

	return filtered, nil
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
