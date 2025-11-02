package osrmclient

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mrasoolmirzaei/delivery-route-system/pkg/httpclient"
	"github.com/mrasoolmirzaei/delivery-route-system/service"
	"github.com/sirupsen/logrus"
)

type TableResponse struct {
	Code      string      `json:"code"`
	Durations [][]float64 `json:"durations"`
	Distances [][]float64 `json:"distances"`
}

type OSRMClient struct {
	client *httpclient.HTTPClient
	log    logrus.FieldLogger
}

func NewOSRMClient(cfg *httpclient.Config) *OSRMClient {
	return &OSRMClient{
		client: httpclient.NewHTTPClient(cfg),
		log:    cfg.Log,
	}
}

func (c *OSRMClient) FindFastestRoutes(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error) {
	routes := make([]*service.Route, 0, len(destinations))
	sourceStr := source.String()
	destinationsStr := make([]string, 0, len(destinations))
	for _, d := range destinations {
		destinationsStr = append(destinationsStr, d.String())
	}
	url := findNearestRoutesURL(sourceStr, destinationsStr)

	tableResponse := &TableResponse{}
	err := c.client.Get(ctx, url, tableResponse)
	if err != nil {
		return nil, err
	}

	if tableResponse.Code != "Ok" {
		return nil, errors.New(tableResponse.Code)
	}

	if isResponseUnexpected(tableResponse, len(destinations)) {
		return nil, errors.New("unexpected table response")
	}

	durations, distances := tableResponse.Durations[0], tableResponse.Distances[0]

	for i, d := range destinations {
		routes = append(routes, &service.Route{
			Destination: d,
			Distance:    distances[i+1],
			Duration:    durations[i+1],
		})
	}

	return routes, nil
}

func findNearestRoutesURL(source string, destinations []string) string {
	destinationsStr := strings.Join(destinations, ";")
	return fmt.Sprintf("http://router.project-osrm.org/table/v1/driving/%s;%s?overview=false&sources=0&annotations=duration,distance", source, destinationsStr)
}

func isResponseUnexpected(tableResponse *TableResponse, totalDestinations int) bool {
	return (len(tableResponse.Durations) != 1 ||
		len(tableResponse.Distances) != 1 ||
		len(tableResponse.Durations[0]) != len(tableResponse.Distances[0]) ||
		len(tableResponse.Durations[0]) != totalDestinations+1)
}
