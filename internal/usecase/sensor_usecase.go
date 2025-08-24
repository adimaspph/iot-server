package usecase

import (
	"context"
	"errors"
	"iot-subscriber/internal/entity"
	"iot-subscriber/internal/model"
	"iot-subscriber/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SensorUsecase struct {
	DB               *gorm.DB
	Log              *logrus.Logger
	Validate         *validator.Validate
	SensorRepository *repository.SensorRepository
	SensorRecordRepo *repository.SensorRecordRepository
}

func NewSensorUsecase(db *gorm.DB, logger *logrus.Logger, validate *validator.Validate, sensorRepository *repository.SensorRepository, sensorRecordRepo *repository.SensorRecordRepository) *SensorUsecase {
	return &SensorUsecase{
		DB:               db,
		Log:              logger,
		Validate:         validate,
		SensorRepository: sensorRepository,
		SensorRecordRepo: sensorRecordRepo,
	}
}

func (c *SensorUsecase) Create(ctx context.Context, request *model.CreateSensorRequest) (*model.SensorResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("failed to validate request body")
		return nil, echo.ErrBadRequest
	}

	// Parse with RFC3339
	requestTimestamp, err := time.Parse(time.RFC3339, request.Timestamp)
	if err != nil {
		c.Log.WithError(err).Error("failed to parse timestamp")
		return nil, echo.ErrBadRequest
	}

	// Check if sensor exists
	var sensor *entity.Sensor
	sensor, err = c.SensorRepository.FindByUnique(tx, request.ID1, request.ID2, request.SensorType)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Sensor does not exist, create new
		sensor = &entity.Sensor{
			ID1:        request.ID1,
			ID2:        request.ID2,
			SensorType: request.SensorType,
		}
		if err := c.SensorRepository.Create(tx, sensor); err != nil {
			c.Log.WithError(err).Error("failed to create sensor")
			return nil, echo.ErrInternalServerError
		}
	}

	// Create initial sensor record
	record := &entity.SensorRecord{
		SensorID:    sensor.SensorID,
		SensorValue: request.SensorValue,
		Timestamp:   requestTimestamp,
	}

	// Insert new sensor
	if err := c.SensorRecordRepo.Create(tx, record); err != nil {
		c.Log.WithError(err).Error("failed to create sensor record")
		return nil, echo.ErrInternalServerError
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("failed to commit transaction")
		return nil, echo.ErrInternalServerError
	}

	// Build response
	resp := &model.SensorResponse{
		ID1:        sensor.ID1,
		ID2:        sensor.ID2,
		SensorType: sensor.SensorType,
		Sensors: []model.Sensor{
			{
				SensorValue: record.SensorValue,
				Timestamp:   record.Timestamp,
			},
		},
	}

	return resp, nil
}
