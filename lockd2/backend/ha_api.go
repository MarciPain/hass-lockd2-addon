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
		haToken = os.Getenv("HASSIO_TOKEN")
	}

	if haToken == "" {
		log.Println("WARNING: Supervisor token is not set. HA Entity discovery will not work.")
	} else {
		// Log just the first few chars for safety if we really need to check if it's there
		log.Printf("HA API initialized with Supervisor Token (starts with: %s...)", haToken[:min(5, len(haToken))])
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
	// Re-check token in case it was set late
	if haToken == "" {
		haToken = os.Getenv("SUPERVISOR_TOKEN")
		if haToken == "" {
			haToken = os.Getenv("HASSIO_TOKEN")
		}
	}

	if haToken == "" {
		// Mock data for local testing...
		return []HAEntity{
			{EntityID: "lock.front_door", Attributes: struct{FriendlyName string `json:"friendly_name,omitempty"`}{FriendlyName: "Front Door"}},
			{EntityID: "switch.gate", Attributes: struct{FriendlyName string `json:"friendly_name,omitempty"`}{FriendlyName: "Gate Relay"}},
			{EntityID: "sensor.front_door_battery", Attributes: struct{FriendlyName string `json:"friendly_name,omitempty"`}{FriendlyName: "Front Door Battery"}},
		}, nil
	}

	// Probláljuk több útvonalon, hátha a Supervisor proxy máshogy viselkedik
	endpoints := []string{
		"http://supervisor/core/api/states",
		"http://supervisor/api/states",
	}

	var lastErr error
	for _, url := range endpoints {
		log.Printf("Calling HA API: %s", url)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		// Csak a legszükségesebb fejléc, néha a túl sok fejléc megzavarja a proxyt
		req.Header.Set("Authorization", "Bearer "+haToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error calling HA API (%s): %v", url, err)
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				lastErr = err
				continue
			}

			var entities []HAEntity
			if err := json.Unmarshal(body, &entities); err != nil {
				lastErr = err
				continue
			}
			
			log.Printf("Successfully fetched %d entities from %s", len(entities), url)
			return filterEntities(entities), nil
		}

		body, _ := io.ReadAll(resp.Body)
		log.Printf("HA API (%s) returned non-OK status: %s, body: %s", url, resp.Status, string(body))
		lastErr = fmt.Errorf("Supervisor API (%s) returned status: %s", url, resp.Status)
	}

	return nil, lastErr
}

func filterEntities(entities []HAEntity) []HAEntity {
	// Filter down to relevant entities for our UI to make payload smaller
	var filtered []HAEntity
	for _, e := range entities {
		if hasPrefix(e.EntityID, "lock.") || hasPrefix(e.EntityID, "switch.") || hasPrefix(e.EntityID, "sensor.") {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}
