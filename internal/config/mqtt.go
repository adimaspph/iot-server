package config

import (
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewMqtt(config *viper.Viper, log *logrus.Logger) mqtt.Client {
	protocol := config.GetString("MQTT_PROTOCOL")
	host := config.GetString("MQTT_HOST")
	port := config.GetString("MQTT_PORT")
	user := config.GetString("MQTT_USER")
	pass := config.GetString("MQTT_PASS")
	clientID := config.GetString("MQTT_CLIENT_ID")

	broker := fmt.Sprintf("%s://%s:%s", protocol, host, port)

	// MQTT broker config
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(user)
	opts.SetPassword(pass)
	opts.SetKeepAlive(60 * time.Second)
	opts.OnConnect = func(c mqtt.Client) {
		log.Infof("MQTT connected to %s", broker)
	}
	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		log.Errorf("MQTT connection lost: %v", err)
	}
	opts.OnReconnecting = func(c mqtt.Client, co *mqtt.ClientOptions) {
		log.Warn("MQTT reconnecting…")
	}

	// Connect MQTT
	mqttClient := mqtt.NewClient(opts)
	token := mqttClient.Connect()

	// Fail fast (instead of waiting forever)
	if !token.WaitTimeout(15 * time.Second) {
		log.Fatal("MQTT connect timed out")
	}
	if err := token.Error(); err != nil {
		log.Fatalf("MQTT connect error: %v", err)
	}
	log.Info("MQTT connected successfully")

	return mqttClient
}
