package server

import (
	"encoding/json"
	"net/http"
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

		response, err := s.routeService.GetRoutes(r.Context(), req)
		if err != nil {
			s.log.WithError(err).Error("failed to get routes")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
