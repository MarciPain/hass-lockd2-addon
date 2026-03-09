package main

import (
	"crypto/tls"
	"fmt"
	"log"
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
		topic := fmt.Sprintf("locks/%s/+", lock.TopicSuffix)
		// Itt a jövőben tudjuk kezelni a lockd-go2-től érkező visszajelzéseket (statuszokat)
		MqttClient.Subscribe(topic, 1, func(c mqtt.Client, m mqtt.Message) {
			log.Printf("Received message on topic %s: %s", m.Topic(), string(m.Payload()))
		})
	}
}
