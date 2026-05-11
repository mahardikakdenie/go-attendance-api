package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx         = context.Background()
	redisClient *redis.Client
)

func NewRedis() *redis.Client {
	if redisClient != nil {
		return redisClient
	}

	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	log.Printf("📡 Connecting to Redis at %s...\n", redisAddr)

	if len(redisAddr) >= 8 && (redisAddr[:8] == "redis://" || (len(redisAddr) >= 9 && redisAddr[:9] == "rediss://")) {
		opts, err := redis.ParseURL(redisAddr)
		if err != nil {
			panic("failed to parse redis url: " + err.Error())
		}
		redisClient = redis.NewClient(opts)
		return redisClient
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: getEnv("REDIS_PASSWORD", ""), // kalau ada password isi disini
		DB:       0,
	})

	return redisClient
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
