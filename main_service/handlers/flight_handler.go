package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type FlightSearchRequest struct {
	From       string `json:"from"`
	To         string `json:"to"`
	Date       string `json:"date"`
	Passengers int    `json:"passengers"`
}

func NewSearchHandler(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req FlightSearchRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "message": "Invalid request body"})
		}

		searchID := uuid.New().String()
		values := map[string]interface{}{
			"search_id":  searchID,
			"from":       req.From,
			"to":         req.To,
			"date":       req.Date,
			"passengers": req.Passengers,
		}

		_, err := rdb.XAdd(context.Background(), &redis.XAddArgs{
			Stream: "flight.search.requested",
			Values: values,
		}).Result()

		if err != nil {
			log.Println("Failed to publish search request:", err)
			return c.Status(500).JSON(fiber.Map{"success": false, "message": "Failed to publish search request"})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Search request submitted",
			"data": fiber.Map{
				"search_id": searchID,
				"status":    "processing",
			},
		})
	}
}

func NewStreamHandler(rdb *redis.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		searchID := c.Params("search_id")
		if searchID == "" {
			return c.Status(400).SendString("search_id is required")
		}

		log.Printf("new SSE client connected with search_id: %s\n", searchID)

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")

		ctx := context.Background()
		lastID := "0"

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			for {
				streams, err := rdb.XRead(ctx, &redis.XReadArgs{
					Streams: []string{"flight.search.results", lastID},
					Block:   5 * time.Second,
				}).Result()

				if err != nil && err != redis.Nil {
					log.Println("Redis XRead error:", err)
					return
				}

				for _, stream := range streams {
					for _, msg := range stream.Messages {
						if msg.Values["search_id"] == searchID {
							jsonData, _ := json.Marshal(msg.Values)

							w.WriteString("data: " + string(jsonData) + "\n\n")
							w.Flush()

							rdb.XDel(ctx, "flight.search.results", msg.ID)
						}
						lastID = msg.ID
					}
				}
			}
		})

		return nil
	}
}
