package main

import (
	"github.com/joho/godotenv"
	"log"
	"provider_service/consumer"
	redis2 "provider_service/redis"
)

func main() {
	_ = godotenv.Load()
	redisClient := redis2.NewRedisClient()

	log.Println("Starting provider service...")
	consumer.StartFlightConsumer(redisClient)
}
