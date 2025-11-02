package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}
}

func (s *Server) getRoutes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := &GetRoutesRequest{}
		params := r.URL.Query()
		source := params.Get("source")
		destinations := params.Get("destinations")
		request.Source = Location(source)
		request.Destinations = make([]Location, 0)
		for _, destination := range destinations {
			request.Destinations = append(request.Destinations, Location(destination))
		}
		response, err := s.routeService.GetRoutes(r.Context(), request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(response)
	}
}