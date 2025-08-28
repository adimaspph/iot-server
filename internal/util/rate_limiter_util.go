package util

import (
	"context"
	"iot-server/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RateLimiterUtil struct {
	Redis      *redis.Client
	Log        *logrus.Logger
	MaxRequest int64
	Duration   time.Duration
}

func NewRateLimiterUtil(redis *redis.Client, log *logrus.Logger, maxRequest int64, duration int) *RateLimiterUtil {
	return &RateLimiterUtil{
		Redis:      redis,
		Log:        log,
		MaxRequest: maxRequest,
		Duration:   time.Duration(duration) * time.Second,
	}
}

func (u RateLimiterUtil) IsAllowed(ctx context.Context, auth *model.Auth) bool {
	key := "rate-limit-" + auth.ID

	increment, err := u.Redis.Incr(ctx, key).Result()
	if err != nil {
		u.Log.Errorln("Error incrementing rate limit value:", err)
		return false
	}

	if increment == 1 {
		err := u.Redis.Expire(ctx, key, u.Duration).Err()
		if err != nil {
			return false
		}
	}

	return increment <= u.MaxRequest
}
