package config

import (
	"iot-subscriber/internal/delivery/http"
	"iot-subscriber/internal/delivery/http/route"
	"iot-subscriber/internal/repository"
	"iot-subscriber/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type BootstrapConfig struct {
	DB       *gorm.DB
	App      *echo.Echo
	Log      *logrus.Logger
	Validate *validator.Validate
	Config   *viper.Viper
}

func Bootstrap(config *BootstrapConfig) {
	// setup repository
	sensorRepository := repository.NewSensorRepository(config.Log)
	sensorRecordRepository := repository.NewSensorRecordRepository(config.Log)

	// setup use cases
	sensorUseCase := usecase.NewSensorUsecase(config.DB, config.Log, config.Validate, sensorRepository, sensorRecordRepository)

	// setup controller
	sensorController := http.NewSensorController(sensorUseCase, config.Log)

	routeConfig := route.RouteConfig{
		App:              config.App,
		SensorController: sensorController,
	}
	routeConfig.Setup()
}
