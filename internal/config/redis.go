package config

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "127.0.0.1:6379"),
		Password: "", // kalau ada password isi disini
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
