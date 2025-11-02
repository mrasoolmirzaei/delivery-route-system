package server

import (
	"fmt"
	"strings"
	"strconv"
)

type Location string

func (l Location) String() string {
	return string(l)
}

func (l Location) Validate() error {
	if l == "" {
		return fmt.Errorf("location is required")
	}

	parts := strings.Split(l.String(), ",")
	if len(parts) != 2 {
		return fmt.Errorf("location must be in the format of latitude,longitude")
	}
	latitude, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %w", err)
	}
	longitude, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %w", err)
	}
	if latitude <= -90 || latitude >= 90 {
		return fmt.Errorf("invalid latitude: %f", latitude)
	}
	if longitude <= -180 || longitude >= 180 {
		return fmt.Errorf("invalid longitude: %f", longitude)
	}

	return nil
}

type GetRoutesRequest struct {
	Source Location
	Destinations []Location
}

type GetRoutesResponse struct {
	Source Location	`json:"source"`
	Routes []*Route `json:"routes"`
}

type Route struct {
	Destination Location `json:"destination"`
	Distance    float64    `json:"distance"`
	Duration    float64    `json:"duration"`
}