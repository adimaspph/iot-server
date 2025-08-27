package config

import (
	"context"
	"database/sql"
	"fmt"
	"iot-server/internal/delivery/http"
	"iot-server/internal/delivery/http/middleware"
	"iot-server/internal/delivery/http/route"
	"iot-server/internal/delivery/messaging"
	"iot-server/internal/entity"
	"iot-server/internal/repository"
	"iot-server/internal/usecase"
	"iot-server/internal/util"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
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

	// setup util
	redisClient := config.Redis
	tokenUtil := util.NewTokenUtil(config.Config.GetString("AUTH_SECRET"), redisClient)
	maxRequest := config.Config.GetInt64("RATE_LIMIT_MAX_REQUEST")
	duration := config.Config.GetInt("RATE_LIMIT_DURATION")
	rateLimitUtil := util.NewRateLimiterUtil(redisClient, config.Log, maxRequest, duration)

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
	authMiddleware := middleware.NewAuth(userUsecase, tokenUtil, rateLimitUtil)

	routeConfig := route.RouteConfig{
		App:              config.App,
		SensorController: sensorController,
		UserController:   userController,
		AuthMiddleware:   authMiddleware,
	}
	routeConfig.Setup()

	// Seed admin user
	ctx := context.Background()
	if err := seedAdmin(ctx, config); err != nil {
		config.Log.WithError(err).Fatalf("failed to seedAdmin")
	}

}

func seedAdmin(ctx context.Context, config *BootstrapConfig) error {
	adminID := config.Config.GetString("ADMIN_ID")
	adminName := config.Config.GetString("ADMIN_NAME")
	adminPass := config.Config.GetString("ADMIN_PASSWORD")

	if adminPass == "" {
		return fmt.Errorf("ADMIN_PASSWORD is required to seed admin")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	now := time.Now().UnixMilli()

	const q = `
		INSERT IGNORE INTO users (id, name, password, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = config.DB.ExecContext(ctx, q, adminID, adminName, string(hash), entity.RoleAdmin, now, now)
	return err
}
