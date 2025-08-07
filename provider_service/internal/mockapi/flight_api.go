package mockapi

import (
	"encoding/json"
	"fmt"
	"os"
	"provider_service/models"
	"strings"
)

func LoadFlightData() ([]models.Flight, error) {
	data, err := os.ReadFile("data/sample.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read flight data: %w", err)
	}

	var flights []models.Flight
	if err := json.Unmarshal(data, &flights); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flight data: %w", err)
	}
	return flights, nil
}

func SearchFlights(from, to, date string) ([]models.Flight, error) {
	allFlights, err := LoadFlightData()
	if err != nil {
		return nil, err
	}

	var results []models.Flight
	for _, f := range allFlights {
		if strings.EqualFold(f.From, from) && strings.EqualFold(f.To, to) &&
			strings.HasPrefix(f.DepartureTime, date) {
			results = append(results, f)
		}
	}

	return results, nil
}
