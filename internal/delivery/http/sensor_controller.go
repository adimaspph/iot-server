package http

import (
	"iot-subscriber/internal/model"
	"iot-subscriber/internal/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type SensorController struct {
	UseCase *usecase.SensorUsecase
	Log     *logrus.Logger
}

func NewSensorController(useCase *usecase.SensorUsecase, log *logrus.Logger) *SensorController {
	return &SensorController{
		UseCase: useCase,
		Log:     log,
	}
}

func (c SensorController) CreateSensor(ctx echo.Context) error {
	var request model.CreateSensorRequest

	err := ctx.Bind(&request)
	if err != nil {
		c.Log.WithError(err).Error("failed to bind request")
		return err
	}

	response, err := c.UseCase.Create(ctx.Request().Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("failed to create sensor record")
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.SensorResponse]{Data: response})
}

func (c SensorController) SearchByCombinedId(ctx echo.Context) error {
	var request model.SensorSearchByIdRequest

	err := ctx.Bind(&request)
	if err != nil {
		c.Log.WithError(err).Error("failed to bind request")
		return err
	}

	// Defaults value
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = 20
	}

	response, metadata, err := c.UseCase.SearchByIdCombination(ctx.Request().Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("failed to search sensor record")
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.SensorResponse]{
		Data:   response,
		Paging: metadata,
	})
}

func (c SensorController) SearchByTimeRange(ctx echo.Context) error {
	var request model.SensorSearchByTimeRangeRequest

	err := ctx.Bind(&request)
	if err != nil {
		c.Log.WithError(err).Error("failed to bind request")
		return err
	}

	// Defaults value
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = 20
	}

	response, metadata, err := c.UseCase.SearchByTimeRange(ctx.Request().Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("failed to search sensor record")
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[[]model.SensorResponse]{
		Data:   response,
		Paging: metadata,
	})
}
