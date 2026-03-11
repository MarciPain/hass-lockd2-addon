package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func StartHAWebSocket() {
	if haToken == "" {
		log.Println("Cannot start HA WebSocket: no supervisor token available")
		return
	}

	go func() {
		for {
			connectAndListenWS()
			log.Println("HA WebSocket disconnected, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}()
}

func connectAndListenWS() {
	urlStr := "ws://supervisor/core/api/websocket"
	log.Printf("Connecting to HA WebSocket: %s", urlStr)

	c, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		log.Printf("WebSocket dial error: %v", err)
		return
	}
	defer c.Close()

	var msg map[string]interface{}
	err = c.ReadJSON(&msg)
	if err != nil || msg["type"] != "auth_required" {
		log.Printf("Expected auth_required, got: %v (err: %v)", msg, err)
		return
	}

	authMsg := map[string]string{
		"type":         "auth",
		"access_token": haToken,
	}
	if err := c.WriteJSON(authMsg); err != nil {
		log.Printf("Auth write error: %v", err)
		return
	}

	err = c.ReadJSON(&msg)
	if err != nil || msg["type"] != "auth_ok" {
		log.Printf("Expected auth_ok, got: %v (err: %v)", msg, err)
		return
	}
	log.Println("HA WebSocket authenticated successfully")

	subMsg := map[string]interface{}{
		"id":         1,
		"type":       "subscribe_events",
		"event_type": "state_changed",
	}
	if err := c.WriteJSON(subMsg); err != nil {
		log.Printf("Subscribe write error: %v", err)
		return
	}

	for {
		err := c.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		if msg["type"] == "event" {
			event, ok := msg["event"].(map[string]interface{})
			if !ok {
				continue
			}

			data, ok := event["data"].(map[string]interface{})
			if !ok {
				continue
			}

			entityID, ok := data["entity_id"].(string)
			if !ok {
				continue
			}

			newState, ok := data["new_state"].(map[string]interface{})
			if !ok || newState == nil {
				continue
			}

			stateStr, ok := newState["state"].(string)
			if !ok {
				continue
			}

			configMut.RLock()
			for _, lock := range AppConfig.Locks {
				if !lock.Enabled {
					continue
				}

				if lock.EntityID == entityID {
					isSwitch := false
					// Checking dynamically string length doesn't matter here, go string operation is fast. Or check strings prefix:
					if len(entityID) >= 7 && entityID[0:7] == "switch." {
						isSwitch = true
					}
					PublishState(lock.TopicSuffix, mapHAStateToHu(stateStr, isSwitch))
				} else if lock.BatteryEntity == entityID {
					PublishBatt(lock.TopicSuffix, stateStr) 
					// Megjegyzés: itt float format is mehetne, 
					// de egyszerűsítve meghagyjuk ahogy a REST API adná, 
					// mivel a valós HA state stringként adja vissza. 
					// ESPHome str_sprintf("%.0f") is stringként publikálta.
				}
			}
			configMut.RUnlock()
		}
	}
}
