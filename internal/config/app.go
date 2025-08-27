package config

import (
	"database/sql"
	"iot-server/internal/delivery/http"
	"iot-server/internal/delivery/http/middleware"
	"iot-server/internal/delivery/http/route"
	"iot-server/internal/delivery/messaging"
	"iot-server/internal/repository"
	"iot-server/internal/usecase"
	"iot-server/internal/util"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
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
	Redis    *redis.Client
}

func Bootstrap(config *BootstrapConfig) {
	// setup repository
	sensorRepository := repository.NewSensorRepository(config.DB, config.Log)
	sensorRecordRepository := repository.NewSensorRecordRepository(config.Log)
	userRepository := repository.NewUserRepository(config.DB, config.Log)

	redisClient := config.Redis
	tokenUtil := util.NewTokenUtil(config.Config.GetString("AUTH_SECRET"), redisClient)

	// setup use cases
	sensorUseCase := usecase.NewSensorUsecase(config.DB, config.Log, config.Validate, sensorRepository, sensorRecordRepository)
	userUsecase := usecase.NewUserUsecase(config.DB, config.Log, config.Validate, userRepository, tokenUtil)

	// setup MQTT broker
	sensorConsumer := messaging.NewSensorConsumer(sensorUseCase, config.Log)
	mqttClient := *config.Mqtt
	mqttClient.Subscribe(config.Config.GetString("MQTT_TOPIC"), 0, sensorConsumer.SensorMQTTHandler)

	// setup controller
	sensorController := http.NewSensorController(sensorUseCase, config.Log)
	userController := http.NewUserController(userUsecase, config.Log)

	// setup middleware
	authMiddleware := middleware.NewAuth(userUsecase, tokenUtil)

	routeConfig := route.RouteConfig{
		App:              config.App,
		SensorController: sensorController,
		UserController:   userController,
		AuthMiddleware:   authMiddleware,
	}
	routeConfig.Setup()
}
