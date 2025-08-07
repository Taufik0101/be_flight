package redis

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis at", addr)
	return client
}
