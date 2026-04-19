package config

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedis() *redis.Client {
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")

	if len(redisAddr) >= 8 && (redisAddr[:8] == "redis://" || (len(redisAddr) >= 9 && redisAddr[:9] == "rediss://")) {
		opts, err := redis.ParseURL(redisAddr)
		if err != nil {
			panic("failed to parse redis url: " + err.Error())
		}
		return redis.NewClient(opts)
	}

	return redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: getEnv("REDIS_PASSWORD", ""), // kalau ada password isi disini
		DB:       0,
	})
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
