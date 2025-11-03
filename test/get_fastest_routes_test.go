package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mrasoolmirzaei/delivery-route-system/server"
	"github.com/mrasoolmirzaei/delivery-route-system/service"
	"github.com/stretchr/testify/suite"
)

func (suite *testSuite) TestGetFastestRoutes() {
	cases := []struct {
		name                string
		request             *server.GetRoutesRequest
		expected            *server.GetRoutesResponse
		routeFinderResponse func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error)
	}{
		{
			name: "one destination",
			request: &server.GetRoutesRequest{
				Source:       "12.3456,78.9101",
				Destinations: []server.Location{"13.1234,12.7890"},
			},
			expected: &server.GetRoutesResponse{
				Source: "12.3456,78.9101",
				Routes: []*server.Route{
					{
						Destination: "13.1234,12.7890",
						Distance:    100,
						Duration:    100,
					},
				},
			},
			routeFinderResponse: func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error) {
				return []*service.Route{
					{
						Destination: "13.1234,12.7890",
						Distance:    100,
						Duration:    100,
					},
				}, nil
			},
		},
		{
			name: "multiple destinations",
			request: &server.GetRoutesRequest{
				Source:       "12.3456,78.9101",
				Destinations: []server.Location{"13.1234,12.7890", "14.1516,17.1819"},
			},
			expected: &server.GetRoutesResponse{
				Source: "12.3456,78.9101",
				Routes: []*server.Route{
					{
						Destination: "13.1234,12.7890",
						Distance:    100,
						Duration:    100,
					},
					{
						Destination: "14.1516,17.1819",
						Distance:    120,
						Duration:    120,
					},
				},
			},
			routeFinderResponse: func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error) {
				return []*service.Route{
					{
						Destination: "13.1234,12.7890",
						Distance:    100,
						Duration:    100,
					},
					{
						Destination: "14.1516,17.1819",
						Distance:    120,
						Duration:    120,
					},
				}, nil
			},
		},
	}

	for _, tc := range cases {
		suite.Run(tc.name, func() {
			suite.osrmMock.FindFastestRoutesFunc = tc.routeFinderResponse
			baseURL := fmt.Sprintf("http://localhost:8090/routes?src=%s", tc.request.Source)
			for _, dst := range tc.request.Destinations {
				baseURL += fmt.Sprintf("&dst=%s", dst)
			}
			resp, err := http.Get(baseURL)
			suite.NoError(err)
			suite.Equal(http.StatusOK, resp.StatusCode)
			var actual server.GetRoutesResponse
			err = json.NewDecoder(resp.Body).Decode(&actual)
			suite.NoError(err)
			suite.Equal(tc.expected, &actual)
		})
	}
}

func (suite *testSuite) TestGetFastestRoutes_Failures() {
	cases := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
		mockFunc       func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error)
	}{
		{
			name:           "missing source parameter",
			url:            "http://localhost:8090/routes?dst=12.3456,78.9101",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "src",
			mockFunc:       nil,
		},
		{
			name:           "missing destination parameter",
			url:            "http://localhost:8090/routes?src=12.3456,78.9101",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "dst",
			mockFunc:       nil,
		},
		{
			name:           "invalid location format",
			url:            "http://localhost:8090/routes?src=invalid&dst=12.3456,78.9101",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "format",
			mockFunc:       nil,
		},
		{
			name:           "route service error",
			url:            "http://localhost:8090/routes?src=12.3456,78.9101&dst=13.1234,12.7890",
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "failed to get routes",
			mockFunc: func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error) {
				return nil, fmt.Errorf("failed to get routes")
			},
		},
		{
			name:           "route service timeout error",
			url:            "http://localhost:8090/routes?src=12.3456,78.9101&dst=13.1234,12.7890",
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "",
			mockFunc: func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error) {
				return nil, fmt.Errorf("OSRM service timeout")
			},
		},
		{
			name:           "route service connection error",
			url:            "http://localhost:8090/routes?src=12.3456,78.9101&dst=13.1234,12.7890",
			expectedStatus: http.StatusServiceUnavailable,
			expectedError:  "",
			mockFunc: func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error) {
				return nil, fmt.Errorf("connection refused")
			},
		},
		{
			name:           "too many destinations",
			url:            buildURLWithManyDestinations("12.3456,78.9101", 81),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "too many destinations",
			mockFunc:       nil,
		},
	}

	for _, tc := range cases {
		suite.Run(tc.name, func() {
			if tc.mockFunc != nil {
				suite.osrmMock.FindFastestRoutesFunc = tc.mockFunc
			}

			resp, err := http.Get(tc.url)
			suite.NoError(err)
			defer resp.Body.Close()
			suite.Equal(tc.expectedStatus, resp.StatusCode)

			if tc.expectedStatus == http.StatusBadRequest {
				var validationErr map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&validationErr)
				suite.NoError(err)
				// Check that error message contains expected text
				errorStr := fmt.Sprintf("%v", validationErr)
				suite.Contains(errorStr, tc.expectedError)
			}
		})
	}
}

func buildURLWithManyDestinations(source string, count int) string {
	url := fmt.Sprintf("http://localhost:8090/routes?src=%s", source)
	for i := 0; i < count; i++ {
		url += fmt.Sprintf("&dst=12.%d,78.%d", i, i+100)
	}
	return url
}

func TestIntegration(t *testing.T) {
	suite.Run(t, new(testSuite))
}
