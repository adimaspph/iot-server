package util

import (
	"context"
	"iot-server/internal/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type TokenUtil struct {
	SecretKey string
	Redis     *redis.Client
}

func NewTokenUtil(secretKey string, redisClient *redis.Client) *TokenUtil {
	return &TokenUtil{
		SecretKey: secretKey,
		Redis:     redisClient,
	}
}

func (t TokenUtil) CreateToken(ctx context.Context, auth *model.Auth) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":     auth.ID,
		"role":   auth.Role,
		"expire": time.Now().Add(time.Hour * 24 * 30).UnixMilli(),
	})

	jwtToken, err := token.SignedString([]byte(t.SecretKey))
	if err != nil {
		return "", err
	}

	_, err = t.Redis.SetEx(ctx, auth.ID, jwtToken, time.Hour*25*30).Result()
	if err != nil {
		return "", err
	}

	return jwtToken, nil
}

func (t TokenUtil) ParseToken(ctx context.Context, jwtToken string) (*model.Auth, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.SecretKey), nil
	})
	if err != nil {
		return nil, echo.ErrUnauthorized
	}

	claims := token.Claims.(jwt.MapClaims)

	expire := claims["expire"].(float64)
	if int64(expire) < time.Now().UnixMilli() {
		return nil, echo.ErrUnauthorized
	}

	result, err := t.Redis.Get(ctx, claims["id"].(string)).Result()
	if err != nil {
		return nil, err
	}

	if result != jwtToken {
		return nil, echo.ErrUnauthorized
	}

	auth := &model.Auth{
		ID:   claims["id"].(string),
		Role: claims["role"].(string),
	}
	return auth, nil
}

func (t TokenUtil) RemoveToken(ctx context.Context, auth *model.Auth) error {
	// Delete the token key from Redis
	_, err := t.Redis.Del(ctx, auth.ID).Result()
	if err != nil {
		return err
	}

	return nil
}
