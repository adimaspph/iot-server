package config

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewRedis(config *viper.Viper, log *logrus.Logger) *redis.Client {
	ctx := context.Background()

	host := config.GetString("REDIS_HOST")
	port := config.GetString("REDIS_PORT")
	database := config.GetInt("REDIS_DB")
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%v:%v", host, port),
		DB:   database,
	})

	if redisClient == nil {
		log.Fatal("Failed to connect to redis")
		panic("Failed to connect to redis")
	}

	// Ping Redis
	for i := 0; i < 10; i++ {
		_, err := redisClient.Ping(ctx).Result()
		if err == nil {
			break
		}
		log.Warn("Waiting for redis...")
		time.Sleep(2 * time.Second)
	}

	msg, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
		panic(err)
	}

	log.Infof("Redis connected successfully: %v", msg)

	return redisClient
}
