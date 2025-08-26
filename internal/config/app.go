package config

import (
	"database/sql"
	"iot-server/internal/delivery/http"
	"iot-server/internal/delivery/http/route"
	"iot-server/internal/delivery/messaging"
	"iot-server/internal/repository"
	"iot-server/internal/usecase"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type BootstrapConfig struct {
	DB       *sql.DB
	App      *echo.Echo
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
	Mqtt     *mqtt.Client
}

func Bootstrap(config *BootstrapConfig) {
	// setup repository
	sensorRepository := repository.NewSensorRepository(config.DB, config.Log)
	sensorRecordRepository := repository.NewSensorRecordRepository(config.Log)

	// setup use cases
	sensorUseCase := usecase.NewSensorUsecase(config.DB, config.Log, config.Validate, sensorRepository, sensorRecordRepository)

	// setup MQTT broker
	sensorConsumer := messaging.NewSensorConsumer(sensorUseCase, config.Log)
	mqttClient := *config.Mqtt
	mqttClient.Subscribe(config.Config.GetString("MQTT_TOPIC"), 0, sensorConsumer.SensorMQTTHandler)
	config.Log.Info(mqttClient.IsConnected())

	// setup controller
	sensorController := http.NewSensorController(sensorUseCase, config.Log)

	routeConfig := route.RouteConfig{
		App:              config.App,
		SensorController: sensorController,
	}
	routeConfig.Setup()
}
