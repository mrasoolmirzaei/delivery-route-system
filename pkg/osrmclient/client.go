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

// OSRM Error codes as defined in the API documentation
const (
	CodeOk             = "Ok"
	CodeInvalidUrl     = "InvalidUrl"
	CodeInvalidService = "InvalidService"
	CodeInvalidVersion = "InvalidVersion"
	CodeInvalidOptions = "InvalidOptions"
	CodeInvalidQuery   = "InvalidQuery"
	CodeInvalidValue   = "InvalidValue"
	CodeNoSegment      = "NoSegment"
	CodeTooBig         = "TooBig"
)

var (
	ErrInvalidUrl     = errors.New("OSRM: URL string is invalid")
	ErrInvalidService = errors.New("OSRM: service name is invalid")
	ErrInvalidVersion = errors.New("OSRM: version is not found")
	ErrInvalidOptions = errors.New("OSRM: options are invalid")
	ErrInvalidQuery   = errors.New("OSRM: query string is syntactically malformed")
	ErrInvalidValue   = errors.New("OSRM: query parameters are invalid")
	ErrNoSegment      = errors.New("OSRM: one of the supplied input coordinates could not snap to street segment")
	ErrTooBig         = errors.New("OSRM: request size violates service specific request size restrictions")
	ErrUnexpected     = errors.New("OSRM: unexpected table response structure")
)

type TableResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message,omitempty"`
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
		return nil, fmt.Errorf("failed to get table response from OSRM: %w", err)
	}

	if tableResponse.Code != CodeOk {
		return nil, handleOSRMError(tableResponse.Code, tableResponse.Message)
	}

	if isResponseUnexpected(tableResponse, len(destinations)) {
		var gotDurations int
		if len(tableResponse.Durations) > 0 && len(tableResponse.Durations[0]) > 0 {
			gotDurations = len(tableResponse.Durations[0])
		}
		return nil, fmt.Errorf("%w: expected %d destinations but got %d durations", ErrUnexpected, len(destinations)+1, gotDurations)
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
	return fmt.Sprintf("http://router.project-osrm.org/table/v1/driving/%s;%s?sources=0&annotations=duration,distance", source, destinationsStr)
}

func handleOSRMError(code, message string) error {
	var baseErr error
	switch code {
	case CodeInvalidUrl:
		baseErr = ErrInvalidUrl
	case CodeInvalidService:
		baseErr = ErrInvalidService
	case CodeInvalidVersion:
		baseErr = ErrInvalidVersion
	case CodeInvalidOptions:
		baseErr = ErrInvalidOptions
	case CodeInvalidQuery:
		baseErr = ErrInvalidQuery
	case CodeInvalidValue:
		baseErr = ErrInvalidValue
	case CodeNoSegment:
		baseErr = ErrNoSegment
	case CodeTooBig:
		baseErr = ErrTooBig
	default:
		if message != "" {
			return fmt.Errorf("OSRM error [%s]: %s", code, message)
		}
		return fmt.Errorf("OSRM error: %s", code)
	}

	if message != "" {
		return fmt.Errorf("%w: %s", baseErr, message)
	}

	return baseErr
}

func isResponseUnexpected(tableResponse *TableResponse, totalDestinations int) bool {
	if len(tableResponse.Durations) != 1 || len(tableResponse.Distances) != 1 {
		return true
	}
	if len(tableResponse.Durations[0]) != len(tableResponse.Distances[0]) {
		return true
	}
	if len(tableResponse.Durations[0]) != totalDestinations+1 {
		return true
	}
	return false
}
