package consumer

import (
	"context"
	"encoding/json"
	"log"
	"provider_service/internal/mockapi"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	streamName   = "flight.search.requested"
	groupName    = "flight-search-group"
	consumerID   = "provider-1"
	resultStream = "flight.search.results"
)

func StartFlightConsumer(rdb *redis.Client) {
	ctx := context.Background()

	// create group and skip if ald exists
	err := rdb.XGroupCreateMkStream(ctx, streamName, groupName, "0").Err()
	if err != nil && !isGroupExistsError(err) {
		log.Fatalf("Failed to create consumer group: %v", err)
	}
	log.Println("consumer group ready to use")

	for {
		streams, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    groupName,
			Consumer: consumerID,
			Streams:  []string{streamName, ">"},
			Block:    5 * time.Second,
			Count:    1,
		}).Result()

		if err != nil && err != redis.Nil {
			log.Printf("XReadGroup error: %v", err)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				go handleFlightRequest(ctx, rdb, msg)
				// acknowledge message
				rdb.XAck(ctx, streamName, groupName, msg.ID)
			}
		}
	}
}

func handleFlightRequest(ctx context.Context, rdb *redis.Client, msg redis.XMessage) {
	log.Println("processing flight request:", msg.ID)

	searchID := getString(msg.Values["search_id"])
	from := getString(msg.Values["from"])
	to := getString(msg.Values["to"])
	date := getString(msg.Values["date"])

	// publish processing ke redis
	err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: resultStream,
		Values: map[string]interface{}{
			"search_id": searchID,
			"status":    "processing",
			"results":   "[]",
		},
	}).Err()

	if err != nil {
		log.Printf("Failed to publish results: %v", err)
	}

	time.Sleep(2 * time.Second)
	results, err := mockapi.SearchFlights(from, to, date)
	if err != nil {
		log.Printf("Error searching flights: %v", err)
		return
	}
	// publish completed ke redis without total
	err = rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: resultStream,
		Values: map[string]interface{}{
			"search_id": searchID,
			"status":    "completed",
			"results":   toJSON(results),
		},
	}).Err()

	if err != nil {
		log.Printf("Failed to publish results: %v", err)
	}

	time.Sleep(2 * time.Second)

	// publish completed ke redis with total
	err = rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: resultStream,
		Values: map[string]interface{}{
			"search_id":     searchID,
			"status":        "completed",
			"results":       toJSON(results),
			"total_results": len(results),
		},
	}).Err()

	if err != nil {
		log.Printf("Failed to publish results: %v", err)
	}
}

func getString(val interface{}) string {
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func toJSON(data interface{}) string {
	bytes, err := json.Marshal(data)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}

func isGroupExistsError(err error) bool {
	return err != nil && (err.Error() == "BUSYGROUP Consumer Group name already exists")
}
