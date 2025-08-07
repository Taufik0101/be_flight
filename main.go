package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type SearchResponse struct {
	Data struct {
		SearchID string `json:"search_id"`
	} `json:"data"`
}

func randomDateBetween(start, end time.Time) string {
	diff := end.Sub(start)
	randomDuration := time.Duration(rand.Int63n(int64(diff)))
	randomTime := start.Add(randomDuration)
	return randomTime.Format("2006-01-02")
}

var airports = []string{"CGK", "DPS", "SUB"}

func randomFromTo() (string, string) {
	from := airports[rand.Intn(len(airports))]
	to := from
	for to == from {
		to = airports[rand.Intn(len(airports))]
	}
	return from, to
}

func main() {
	var wg sync.WaitGroup
	requests := 10
	startDate := time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 7, 17, 0, 0, 0, 0, time.UTC)
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			fmt.Printf("[Req %d] Sending flight search...\n", i)

			from, to := randomFromTo()
			date := randomDateBetween(startDate, endDate)

			payload := map[string]interface{}{
				"from":       from,
				"to":         to,
				"date":       date,
				"passengers": 1,
			}
			body, _ := json.Marshal(payload)
			fmt.Printf("[Req %d] Payload: %v\n", i, payload)

			resp, err := http.Post("http://localhost:3000/api/flights/search", "application/json", bytes.NewBuffer(body))
			if err != nil {
				fmt.Printf("[Req %d] Error sending request: %v\n", i, err)
				return
			}
			defer resp.Body.Close()

			var result SearchResponse
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				fmt.Printf("[Req %d] Error decoding response: %v\n", i, err)
				return
			}

			searchID := result.Data.SearchID
			if searchID == "" {
				fmt.Printf("[Req %d] No search_id received\n", i)
				return
			}

			fmt.Printf("[Req %d] Got search_id: %s\n", i, searchID)

			streamURL := fmt.Sprintf("http://localhost:3000/api/flights/search/%s/stream", searchID)
			req, _ := http.NewRequest("GET", streamURL, nil)
			client := &http.Client{Timeout: 15 * time.Second}
			streamResp, err := client.Do(req)
			if err != nil {
				fmt.Printf("[Req %d] SSE error: %v\n", i, err)
				return
			}
			defer streamResp.Body.Close()

			scanner := bufio.NewScanner(streamResp.Body)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "data:") {
					fmt.Printf("[Req %d] %s\n", i, strings.TrimPrefix(line, "data: "))
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("✅ All requests completed.")
}
