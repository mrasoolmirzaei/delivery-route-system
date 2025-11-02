package service

import (
	"context"
	"sort"
)

type RouteService interface {
	GetFastestRoutes(ctx context.Context, source Location, destinations []Location) ([]*Route, error)
}

type routeServiceImpl struct {
	routeFinder routeFinder
}

type routeFinder interface {
	FindFastestRoutes(ctx context.Context, source Location, destinations []Location) ([]*Route, error)
}

func NewRouteService(routeFinder routeFinder) RouteService {
	return &routeServiceImpl{routeFinder: routeFinder}
}

func (s *routeServiceImpl) GetFastestRoutes(ctx context.Context, source Location, destinations []Location) ([]*Route, error) {
	routes, err := s.routeFinder.FindFastestRoutes(ctx, source, destinations)
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
