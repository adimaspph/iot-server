package config

import (
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewMqtt(config *viper.Viper, log *logrus.Logger) mqtt.Client {
	broker := config.GetString("MQTT_BROKER")
	user := config.GetString("MQTT_USER")
	pass := config.GetString("MQTT_PASS")
	clientID := config.GetString("MQTT_CLIENT_ID")

	// MQTT broker config
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(user)
	opts.SetPassword(pass)
	opts.SetKeepAlive(60 * time.Second)

	return mqtt.NewClient(opts)
}
