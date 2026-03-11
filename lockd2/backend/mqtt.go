package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var MqttClient mqtt.Client

func InitMQTT() {
	configMut.RLock()
	host := AppConfig.MqttHost
	port := AppConfig.MqttPort
	user := AppConfig.MqttUser
	pass := AppConfig.MqttPass
	useSSL := AppConfig.MqttSSL
	configMut.RUnlock()

	if host == "" {
		log.Println("MQTT settings missing, skipping connection.")
		return
	}

	protocol := "tcp"
	if useSSL {
		protocol = "tls"
	}
	brokerURL := fmt.Sprintf("%s://%s:%d", protocol, host, port)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(fmt.Sprintf("lockd2-addon-%d", time.Now().Unix()))
	opts.SetUsername(user)
	opts.SetPassword(pass)

	if useSSL {
		opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}

	opts.OnConnect = func(c mqtt.Client) {
		log.Printf("Connected to MQTT broker at %s", brokerURL)
		SubscribeAll()
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Printf("Lost connecton to MQTT broker: %v", err)
	}

    if MqttClient != nil && MqttClient.IsConnected() {
        MqttClient.Disconnect(250)
    }

	MqttClient = mqtt.NewClient(opts)
	go func() {
		if token := MqttClient.Connect(); token.Wait() && token.Error() != nil {
			log.Printf("Failed to connect to MQTT in background: %v", token.Error())
		}
	}()
}

func SubscribeAll() {
	if MqttClient == nil || !MqttClient.IsConnected() {
		return
	}

	configMut.RLock()
	defer configMut.RUnlock()

	// Iratkozzunk fel minden olyan topic-ra amit konfiguráltunk
	for _, lock := range AppConfig.Locks {
		if !lock.Enabled {
			continue
		}
		
		topicSuffix := lock.TopicSuffix
		cmdTopic := fmt.Sprintf("locks/%s/cmd", topicSuffix)

		// Command handler
		MqttClient.Subscribe(cmdTopic, 1, handleMQTTMessage(lock))

		// Küldjük ki az "online" állapotot bekapcsoláskor (Zárva/Nyitva helyett, mint az ESPHome on_boot)
		PublishState(topicSuffix, "online")

		// Lekérdezzük a kezdeti állapotot HA-ból
		go FetchAndPublishState(lock)
	}
}

func handleMQTTMessage(lock LockEntity) mqtt.MessageHandler {
	return func(c mqtt.Client, m mqtt.Message) {
		cmd := strings.ToUpper(strings.TrimSpace(string(m.Payload())))
		log.Printf("Received cmd '%s' on %s", cmd, m.Topic())

		domain := "lock"
		if strings.HasPrefix(lock.EntityID, "switch.") {
			domain = "switch"
		}

		switch cmd {
		case "LOCK", "ON":
			if domain == "switch" {
				CallHAService("switch", "turn_on", lock.EntityID)
			} else {
				CallHAService("lock", "lock", lock.EntityID)
			}
		case "UNLOCK", "OFF":
			if domain == "switch" {
				CallHAService("switch", "turn_off", lock.EntityID)
			} else {
				CallHAService("lock", "unlock", lock.EntityID)
			}
		case "STATUS":
			PublishStatusAck(lock.TopicSuffix)
			FetchAndPublishState(lock)
		}
	}
}

func PublishState(topicSuffix, state string) {
	if MqttClient == nil || !MqttClient.IsConnected() {
		return
	}
	topic := fmt.Sprintf("locks/%s/state", topicSuffix)
	MqttClient.Publish(topic, 1, true, state) // retain = true
}

func PublishBatt(topicSuffix, batt string) {
	if MqttClient == nil || !MqttClient.IsConnected() {
		return
	}
	topic := fmt.Sprintf("locks/%s/batt", topicSuffix)
	MqttClient.Publish(topic, 1, true, batt) // retain = true
}

func PublishStatusAck(topicSuffix string) {
	if MqttClient == nil || !MqttClient.IsConnected() {
		return
	}
	topic := fmt.Sprintf("locks/%s/status_ack", topicSuffix)
	payload := fmt.Sprintf("ACK %d", time.Now().Unix())
	MqttClient.Publish(topic, 1, false, payload) // retain = false
}
