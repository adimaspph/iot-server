package http

import (
	"iot-server/internal/model"
	"iot-server/internal/usecase"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserController struct {
	Log     *logrus.Logger
	UseCase *usecase.UserUsecase
}

func NewUserController(useCase *usecase.UserUsecase, logger *logrus.Logger) *UserController {
	return &UserController{
		Log:     logger,
		UseCase: useCase,
	}
}

func (c *UserController) Register(ctx echo.Context) error {
	var request model.RegisterUserRequest

	err := ctx.Bind(&request)
	if err != nil {
		c.Log.WithError(err).Error("failed to bind request")
		return err
	}

	response, err := c.UseCase.Create(ctx.Request().Context(), &request)
	if err != nil {
		c.Log.WithError(err).Error("failed to register user")
		return err
	}

	return ctx.JSON(http.StatusOK, model.WebResponse[*model.UserResponse]{Data: response})
}
