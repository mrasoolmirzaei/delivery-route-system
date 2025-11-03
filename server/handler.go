package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mrasoolmirzaei/delivery-route-system/service"
)

func (s *Server) health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":  "healthy",
			"service": "delivery-route-system",
		}

		// Check OSRM dependency if routeService is available
		if s.routeService != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()

			// Test with a simple health check location
			testSource := service.Location("0,0")
			testDest := []service.Location{service.Location("0.001,0.001")}

			_, err := s.routeService.GetFastestRoutes(ctx, testSource, testDest)
			if err != nil {
				// Check if it's a timeout or context cancellation
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					status["osrm"] = "timeout"
					status["status"] = "degraded"
				}
				status["osrm"] = "unavailable"
				status["status"] = "degraded"
			} else {
				status["osrm"] = "healthy"
			}
		}

		statusCode := http.StatusOK
		if status["status"] != "healthy" {
			statusCode = http.StatusServiceUnavailable
		}

		writeJSON(w, statusCode, status)
	}
}

func (s *Server) getRoutes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, validationErr := validateGetRoutesRequest(r)
		if validationErr != nil {
			s.log.WithError(validationErr).Error("failed to validate get routes request")
			writeJSON(w, http.StatusBadRequest, validationErr)
			return
		}

		source := service.Location(req.Source)
		destinations := make([]service.Location, len(req.Destinations))
		for i, dst := range req.Destinations {
			destinations[i] = service.Location(dst)
		}

		serviceRoutes, err := s.routeService.GetFastestRoutes(r.Context(), source, destinations)
		if err != nil {
			s.log.WithError(err).Error("failed to get routes")

			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				writeJSON(w, http.StatusRequestTimeout, map[string]string{
					"error": "request timeout",
				})
				return
			}

			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"error": getErrorMessage(err),
			})
			return
		}

		serverRoutes := make([]*Route, len(serviceRoutes))
		for i, route := range serviceRoutes {
			serverRoutes[i] = &Route{
				Destination: Location(route.Destination),
				Distance:    route.Distance,
				Duration:    route.Duration,
			}
		}

		response := &GetRoutesResponse{
			Source: req.Source,
			Routes: serverRoutes,
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getErrorMessage(err error) string {
	if err == nil {
		return "unknown error"
	}

	errStr := err.Error()
	if containsAny(errStr, []string{"connection refused", "timeout", "context deadline"}) {
		return "service temporarily unavailable"
	}

	return "route calculation failed"
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
