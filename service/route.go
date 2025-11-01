package service

import (
	"context"
	"github.com/mrasoolmirzaei/delivery-route-system/server"
)

type RouteService struct {
}

type routeFinder interface {
	FindNearestRoute(ctx context.Context, source, destination string) (*Route, error)
}

func NewRouteService() *RouteService {
	return &RouteService{}
}

func (s *RouteService) GetAllRoutes(ctx context.Context, request *server.GetAllRoutesRequest) (*server.GetAllRoutesResponse, error) {
	return nil, nil
}