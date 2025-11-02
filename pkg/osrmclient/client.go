package osrmclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/mrasoolmirzaei/delivery-route-system/pkg/httpclient"
	"github.com/mrasoolmirzaei/delivery-route-system/service"
)

type DrivingInfo struct {
	Routes []Route `json:"routes"`
	Code   string  `json:"code"`
}

type Route struct {
	Distance float64 `json:"distance"`
	Duration float64 `json:"duration"`
}

type OSRMClient struct {
	client *httpclient.HTTPClient
}

func NewOSRMClient(cfg *httpclient.Config) *OSRMClient {
	return &OSRMClient{client: httpclient.NewHTTPClient(cfg)}
}

func (c *OSRMClient) FindNearestRoute(ctx context.Context, source, destination string) (*service.Route, error) {
	url := findNearestRouteURL(source, destination)
	drivingInfo := &DrivingInfo{}
	err := c.client.Get(ctx, url, drivingInfo)
	if err != nil {
		return nil, err
	}

	if drivingInfo.Code != "Ok" {
		return nil, errors.New(drivingInfo.Code)
	}


	if len(drivingInfo.Routes) == 0 {
		return nil, errors.New("no routes found")
	}

	return &service.Route{
		Source: source,
		Destination: destination,
		Distance: drivingInfo.Routes[0].Distance,
		Duration: drivingInfo.Routes[0].Duration,
	}, nil
}

func findNearestRouteURL(source, destination string) string {
	return fmt.Sprintf("http://router.project-osrm.org/route/v1/driving/%s;%s", source, destination)
}