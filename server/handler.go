package server

import (
	"encoding/json"
	"net/http"

	"github.com/mrasoolmirzaei/delivery-route-system/service"
)

func (s *Server) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, "pong")
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

		// Convert server.Location to service.Location
		source := service.Location(req.Source)
		destinations := make([]service.Location, len(req.Destinations))
		for i, dst := range req.Destinations {
			destinations[i] = service.Location(dst)
		}

		// Call service with domain types
		serviceRoutes, err := s.serviceRouteService.GetRoutes(r.Context(), source, destinations)
		if err != nil {
			s.log.WithError(err).Error("failed to get routes")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert service.Route to server.Route
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
