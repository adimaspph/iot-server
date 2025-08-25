package usecase

import (
	"context"
	"errors"
	"iot-subscriber/internal/entity"
	"iot-subscriber/internal/model"
	"iot-subscriber/internal/model/converter"
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
		SensorsRecords: []model.SensorRecord{
			{
				SensorValue: record.SensorValue,
				Timestamp:   record.Timestamp,
			},
		},
	}

	return resp, nil
}

func (c SensorUsecase) SearchByIdCombination(ctx context.Context, request *model.SensorSearchByIdRequest) (*model.SensorResponse, *model.PageMetadata, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// validate request
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("failed to validate request body")
		return nil, nil, echo.ErrBadRequest
	}

	sensor, metadata, err := c.SensorRepository.FindSensorRecordsByIdCombination(tx, request.ID1, request.ID2, request.Page, request.PageSize)

	if err != nil {
		c.Log.WithError(err).Error("error getting sensor records")
		return nil, nil, echo.ErrNotFound
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, echo.ErrInternalServerError
	}

	// Build response
	resp := converter.SensorToResponse(sensor)

	return resp, metadata, nil
}

func (u SensorUsecase) SearchByTimeRange(ctx context.Context, request *model.SensorSearchByTimeRangeRequest) ([]model.SensorResponse, *model.PageMetadata, error) {
	tx := u.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// validate request
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Error("failed to validate request body")
		return nil, nil, echo.ErrBadRequest
	}

	sensors, metadata, err := u.SensorRepository.FindSensorRecordsByTimeRange(tx, request.Start, request.End, request.Page, request.PageSize)

	if err != nil {
		u.Log.WithError(err).Error("error getting sensors records")
		return nil, nil, echo.ErrNotFound
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		u.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, echo.ErrInternalServerError
	}

	// Build response
	resp := converter.SensorRecordsToResponse(sensors)

	return resp, metadata, nil
}

func (c SensorUsecase) SearchByIdAndTimeRange(ctx context.Context, request *model.SensorSearchByIdAndTimeRangeRequest) (*model.SensorResponse, *model.PageMetadata, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	// validate request
	if err := c.Validate.Struct(request); err != nil {
		c.Log.WithError(err).Error("failed to validate request body")
		return nil, nil, echo.ErrBadRequest
	}

	sensor, metadata, err := c.SensorRepository.FindSensorRecordsByIdAndTimeRange(tx, request.ID1, request.ID2, request.Start, request.End, request.Page, request.PageSize)

	if err != nil {
		c.Log.WithError(err).Error("error getting sensor records")
		return nil, nil, echo.ErrNotFound
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.Log.WithError(err).Error("failed to commit transaction")
		return nil, nil, echo.ErrInternalServerError
	}

	// Build response
	resp := converter.SensorToResponse(sensor)

	return resp, metadata, nil
}
