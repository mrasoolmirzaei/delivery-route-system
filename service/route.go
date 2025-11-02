package service

import (
	"context"
	"sort"

	"github.com/mrasoolmirzaei/delivery-route-system/server"
)

type RouteService struct {
	routeFinder routeFinder
}

type routeFinder interface {
	FindNearestRoute(ctx context.Context, source, destination string) (*Route, error)
}

func NewRouteService(routeFinder routeFinder) *RouteService {
	return &RouteService{routeFinder: routeFinder}
}

func (s *RouteService) GetRoutes(ctx context.Context, request *server.GetRoutesRequest) (*server.GetRoutesResponse, error) {
	routes := make([]*server.Route, 0)
	for _, destination := range request.Destinations {
		route, err := s.routeFinder.FindNearestRoute(ctx, request.Source.String(), destination.String())
		if err != nil {
			return nil, err
		}
		routes = append(routes, &server.Route{
			Destination: destination,
			Distance:    route.Distance,
			Duration:    route.Duration,
		})
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Duration == routes[j].Duration {
			return routes[i].Distance < routes[j].Distance
		}
		return routes[i].Duration < routes[j].Duration
	})

	return &server.GetRoutesResponse{
		Source: request.Source,
		Routes: routes,
	}, nil
}
