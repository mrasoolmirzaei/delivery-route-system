package osrmclient

import (
	"context"
	"github.com/mrasoolmirzaei/delivery-route-system/service"
)

type MockOSRMClient struct {
	FindFastestRoutesFunc func(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error)
}

func (m *MockOSRMClient) FindFastestRoutes(ctx context.Context, source service.Location, destinations []service.Location) ([]*service.Route, error) {
	return m.FindFastestRoutesFunc(ctx, source, destinations)
}