package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"log"
	"main_service/handlers"
	"main_service/redis"
	"os"
)

func main() {
	app := fiber.New()
	_ = godotenv.Load()

	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowHeaders:  "Origin, Content-Type, Accept",
		AllowMethods:  "GET, POST, OPTIONS",
		ExposeHeaders: "Content-Type",
	}))

	redisClient := redis.NewRedisClient()
	app.Static("/", "./public")

	app.Post("/api/flights/search", handlers.NewSearchHandler(redisClient))
	app.Get("/api/flights/search/:search_id/stream", handlers.NewStreamHandler(redisClient))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Main service listening on port %s\n", port)
	log.Fatal(app.Listen(":" + port))
}
