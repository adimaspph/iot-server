package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"iot-server/internal/entity"
	"iot-server/internal/model"
	"iot-server/internal/repository"
	"iot-server/internal/util"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
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

func (u *UserUsecase) Create(ctx context.Context, request *model.RegisterUserRequest) (*model.UserResponse, error) {
	if err := u.Validate.Struct(request); err != nil {
		u.Log.WithError(err).Error("failed to validate request")
		return nil, echo.ErrBadRequest
	}

	// Validate role
	role := entity.UserRole(request.Role)
	if role != entity.RoleAdmin && role != entity.RoleUser {
		u.Log.Error("role is invalid")
		return nil, echo.ErrBadRequest
	}

	// Check existence
	if _, err := u.Repository.FindByID(ctx, request.ID); err == nil {
		u.Log.Warn("user id already exists")
		return nil, echo.ErrConflict
	} else if !errors.Is(err, sql.ErrNoRows) {
		// user not found
		u.Log.WithError(err).Error("failed to find user by id")
		return nil, echo.ErrInternalServerError
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		u.Log.WithError(err).Error("failed to hashing password")
		return nil, echo.ErrInternalServerError
	}

	user := &entity.User{
		ID:        request.ID,
		Name:      request.Name,
		Password:  string(hash),
		Role:      string(role),
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
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

	if err := u.Repository.Create(ctx, user); err != nil {
		u.Log.WithError(err).Error("failed to create user")
		return nil, echo.ErrInternalServerError
	}

	// commit
	if err := tx.Commit(); err != nil {
		u.Log.WithError(err).Error("failed to commit transaction")
		return nil, echo.ErrInternalServerError
	}

	return &model.UserResponse{
		ID:   user.ID,
		Name: user.Name,
	}, nil
}

func (u *UserUsecase) Login(ctx context.Context, req *model.LoginUserRequest) (*model.UserResponse, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Warn("failed to validate request")
		return nil, fmt.Errorf("%w: %v", echo.ErrBadRequest, err)
	}

	user, err := u.Repository.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			u.Log.WithError(err).Warn("user not found")
			return nil, fmt.Errorf("%w: %v", echo.ErrUnauthorized, errors.New("user not found"))
		}
		u.Log.WithError(err).Error("failed to find user by id")
		return nil, echo.ErrInternalServerError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		u.Log.WithError(err).Warn("wrong password")
		return nil, fmt.Errorf("%w: %v", echo.ErrUnauthorized, errors.New("wrong password"))
	}

	// Create JWT (stateless) and store it to redis cache
	token, err := u.TokenUtil.CreateToken(ctx, &model.Auth{ID: user.ID})
	if err != nil {
		u.Log.WithError(err).Error("failed to create token")
		return nil, echo.ErrInternalServerError
	}

	return &model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Token: token,
	}, nil
}

func (u *UserUsecase) Logout(ctx context.Context, req *model.LogoutUserRequest) (bool, error) {
	if err := u.Validate.Struct(req); err != nil {
		u.Log.WithError(err).Warn("failed to validate request")
		return false, fmt.Errorf("%w: %v", echo.ErrBadRequest, err)
	}

	// Ensure user exists
	user, err := u.Repository.FindByID(ctx, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, echo.ErrNotFound
		}
		u.Log.WithError(err).Error("failed to find user by id")
		return false, echo.ErrInternalServerError
	}

	auth := &model.Auth{
		ID: user.ID,
	}

	// Remove token from redis
	err = u.TokenUtil.RemoveToken(ctx, auth)
	if err != nil {
		u.Log.WithError(err).Error("failed to create token")
		return false, echo.ErrInternalServerError
	}

	return true, nil
}
