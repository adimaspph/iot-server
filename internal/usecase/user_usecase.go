package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iot-server/internal/model"
	"iot-server/internal/repository"
	"iot-server/internal/util"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type UserUsecase struct {
	DB         *sql.DB
	Log        *logrus.Logger
	Validate   *validator.Validate
	Repository *repository.UserRepository
	TokenUtil  *util.TokenUtil
}

func NewUserUsecase(
	db *sql.DB,
	log *logrus.Logger,
	validate *validator.Validate,
	repo *repository.UserRepository,
	tokenUtil *util.TokenUtil,
) *UserUsecase {
	return &UserUsecase{
		DB:         db,
		Log:        log,
		Validate:   validate,
		Repository: repo,
		TokenUtil:  tokenUtil,
	}
}

func (u *UserUsecase) Verify(ctx context.Context, request *model.VerifyUserRequest) (*model.Auth, error) {
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Warn("failed to validate request")
		return nil, fmt.Errorf("%w: %v", echo.ErrBadRequest, err)
	}

	user, err := u.Repository.FindByToken(ctx, request.Token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.Log.WithError(err).Warn("token doesn't exist")
			return nil, echo.ErrNotFound
		}
		u.Log.WithError(err).Error("failed to find user by token")
		return nil, echo.ErrInternalServerError
	}

	return &model.Auth{ID: user.ID}, nil
}
