package main

import (
	"context"
	"errors"
	"fmt"
	"iot-server/internal/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	viperConfig := config.NewViper()
	log := config.NewLogger(viperConfig)
	db := config.NewDatabase(viperConfig, log)
	validate := config.NewValidator(viperConfig)
	app := config.NewEcho(viperConfig)
	mqttClient := config.NewMqtt(viperConfig, log)

	// Connect MQTT
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	log.Info("MQTT connected")

	// Start Echo server in goroutine
	port := viperConfig.GetInt("APP_PORT")
	serverAddr := fmt.Sprintf(":%d", port)

	go func() {
		log.Infof("Starting server on %s", serverAddr)
		if err := app.Start(serverAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Echo server error: %v", err)
		}
	}()

	config.Bootstrap(&config.BootstrapConfig{
		DB:       db,
		App:      app,
		Log:      log,
		Validate: validate,
		Config:   viperConfig,
		Mqtt:     &mqttClient,
	})

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigCh
	log.Infof("Received signal: %s. Shutting down...", s.String())

	// Shutdown Echo then MQTT
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		log.Errorf("Echo shutdown error: %v", err)
	} else {
		log.Info("HTTP server stopped")
	}

	// Allow in-flight MQTT work to flush
	mqttClient.Disconnect(250)
	log.Info("MQTT disconnected")

	log.Info("Shutdown complete")
}
