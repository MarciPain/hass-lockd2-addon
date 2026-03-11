package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	log.Println("Lockd2 Addon Backend starting...")

	// Inicializáljuk a beállítások betöltését
	LoadConfig()

	// Ingress port, a config.yaml alapján a 8099-es portot kértük
	port := os.Getenv("PORT")
	if port == "" {
		port = "8099"
	}

	// Útvonalak beállítása a beágyazott Frontendhez
	SetupRoutes()

	// HA API inicializálása
	InitHAAPI()
	StartHAWebSocket()

	// Ha van beállított MQTT, akkor csatlakozunk
	if AppConfig.MqttHost != "" {
		InitMQTT()
	}

	log.Printf("Listening on :%s (Ingress)...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
