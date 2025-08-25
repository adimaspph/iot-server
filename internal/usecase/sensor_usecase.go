package usecase

import (
	"context"
	"database/sql"
	"errors"
	"iot-server/internal/entity"
	"iot-server/internal/model"
	"iot-server/internal/model/converter"
	"iot-server/internal/repository"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type SensorUsecase struct {
	DB               *sql.DB
	Log              *logrus.Logger
	Validate         *validator.Validate
	SensorRepository *repository.SensorRepository
	SensorRecordRepo *repository.SensorRecordRepository
}

func NewSensorUsecase(
	db *sql.DB,
	logger *logrus.Logger,
	validate *validator.Validate,
	sensorRepository *repository.SensorRepository,
	sensorRecordRepo *repository.SensorRecordRepository,
) *SensorUsecase {
	return &SensorUsecase{
		DB:               db,
		Log:              logger,
		Validate:         validate,
		SensorRepository: sensorRepository,
		SensorRecordRepo: sensorRecordRepo,
	}
}

func (u *SensorUsecase) Create(ctx context.Context, request *model.CreateSensorRequest) (*model.SensorResponse, error) {
	// validate
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Error("failed to validate request body")
		return nil, echo.ErrBadRequest
	}

	// Parse with RFC3339
	requestTimestamp, err := time.Parse(time.RFC3339, request.Timestamp)
	if err != nil {
		u.Log.WithError(err).Error("failed to parse timestamp")
		return nil, echo.ErrBadRequest
	}

	// begin tx
	tx, err := u.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		u.Log.WithError(err).Error("failed to begin transaction")
		return nil, echo.ErrInternalServerError
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// find or create sensor
	sensor, err := u.SensorRepository.FindByUnique(ctx, request.ID1, request.ID2, request.SensorType)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			sensor = &entity.Sensor{
				ID1:        request.ID1,
				ID2:        request.ID2,
				SensorType: request.SensorType,
			}
			if err := u.SensorRepository.CreateTx(ctx, tx, sensor); err != nil {
				u.Log.WithError(err).Error("failed to create sensor")
				return nil, echo.ErrInternalServerError
			}
		} else {
			u.Log.WithError(err).Error("failed to find sensor")
			return nil, echo.ErrInternalServerError
		}
	}

	// Create initial sensor record
	record := &entity.SensorRecord{
		SensorID:    sensor.SensorID,
		SensorValue: request.SensorValue,
		Timestamp:   requestTimestamp,
	}
	if err := u.SensorRecordRepo.CreateTx(ctx, tx, record); err != nil {
		u.Log.WithError(err).Error("failed to create sensor record")
		return nil, echo.ErrInternalServerError
	}

	// commit
	if err := tx.Commit(); err != nil {
		u.Log.WithError(err).Error("failed to commit transaction")
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

func (u *SensorUsecase) SearchByIdCombination(ctx context.Context, req *model.SensorSearchByIdRequest) (*model.SensorResponse, *model.PageMetadata, error) {
	// validate
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("failed to validate request")
		return nil, nil, echo.ErrBadRequest
	}

	sensor, meta, err := u.SensorRepository.FindSensorRecordsByIdCombination(ctx, req.ID1, req.ID2, req.Page, req.PageSize)
	if err != nil {
		u.Log.WithError(err).Error("error getting sensor records")
		return nil, nil, echo.ErrInternalServerError
	}

	resp := converter.SensorToResponse(sensor)
	return resp, meta, nil
}

func (u *SensorUsecase) SearchByTimeRange(ctx context.Context, req *model.SensorSearchByTimeRangeRequest) ([]model.SensorResponse, *model.PageMetadata, error) {
	// validate
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("failed to validate request")
		return nil, nil, echo.ErrBadRequest
	}

	sensors, meta, err := u.SensorRepository.FindSensorRecordsByTimeRange(ctx, req.Start, req.End, req.Page, req.PageSize)
	if err != nil {
		u.Log.WithError(err).Error("error getting sensors records")
		return nil, nil, echo.ErrInternalServerError
	}

	resp := converter.SensorRecordsToResponse(sensors)
	return resp, meta, nil
}

func (u *SensorUsecase) SearchByIdAndTimeRange(ctx context.Context, req *model.SensorSearchByIdAndTimeRangeRequest) (*model.SensorResponse, *model.PageMetadata, error) {
	// validate
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("failed to validate request")
		return nil, nil, echo.ErrBadRequest
	}

	sensor, meta, err := u.SensorRepository.FindSensorRecordsByIdAndTimeRange(ctx, req.ID1, req.ID2, req.Start, req.End, req.Page, req.PageSize)
	if err != nil {
		u.Log.WithError(err).Error("error getting sensor records")
		return nil, nil, echo.ErrInternalServerError
	}

	resp := converter.SensorToResponse(sensor)

	return resp, meta, nil
}

func (u *SensorUsecase) DeleteByIdCombination(ctx context.Context, req *model.SensorSearchByIdRequest) (*model.SensorDeleteResponse, error) {
	// validate
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("failed to validate request")
		return nil, echo.ErrBadRequest
	}

	deletedRow, err := u.SensorRepository.DeleteRecordsByIdCombination(ctx, req.ID1, req.ID2)
	if err != nil {
		u.Log.WithError(err).Error("error when deleting sensor records")
		return nil, echo.ErrInternalServerError
	}

	resp := &model.SensorDeleteResponse{
		Deleted: deletedRow,
	}
	return resp, nil
}

func (u *SensorUsecase) DeleteByTimeRange(ctx context.Context, req *model.SensorSearchByTimeRangeRequest) (*model.SensorDeleteResponse, error) {
	// validate
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Error("failed to validate request")
		return nil, echo.ErrBadRequest
	}

	deletedRows, err := u.SensorRepository.DeleteRecordsByTimeRange(ctx, req.Start, req.End)
	if err != nil {
		u.Log.WithError(err).Error("error getting sensors records")
		return nil, echo.ErrInternalServerError
	}

	resp := &model.SensorDeleteResponse{
		Deleted: deletedRows,
	}
	return resp, nil
}
