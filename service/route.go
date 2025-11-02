package service

import (
	"context"
	"sort"
)

type RouteService struct {
	routeFinder routeFinder
}

type routeFinder interface {
	FindNearestRoutes(ctx context.Context, source Location, destinations []Location) ([]*Route, error)
}

func NewRouteService(routeFinder routeFinder) *RouteService {
	return &RouteService{routeFinder: routeFinder}
}

func (s *RouteService) GetRoutes(ctx context.Context, source Location, destinations []Location) ([]*Route, error) {
	routes, err := s.routeFinder.FindNearestRoutes(ctx, source, destinations)
	if err != nil {
		return nil, err
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Duration == routes[j].Duration {
			return routes[i].Distance < routes[j].Distance
		}
		return routes[i].Duration < routes[j].Duration
	})

	return routes, nil
}
