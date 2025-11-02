package test

import (
	"github.com/mrasoolmirzaei/delivery-route-system/server"
	"github.com/mrasoolmirzaei/delivery-route-system/pkg/osrmclient"
	"github.com/mrasoolmirzaei/delivery-route-system/service"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"os"
)

type testSuite struct {
	suite.Suite
	server *server.Server
	osrmMock *osrmclient.MockOSRMClient
}

func (suite *testSuite) SetupSuite() {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetOutput(os.Stderr)
	loggerEntry := logrus.NewEntry(logger)

	osrmClient := &osrmclient.MockOSRMClient{}
	routeService := service.NewRouteService(osrmClient)
	server, err := server.NewServer(server.Config{
		Logger:       loggerEntry,
		RouteService: routeService,
	})
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.server = server
	suite.osrmMock = osrmClient

	go func() {
		suite.NoError(server.Serve(":8090"))
	}()
}

func (suite *testSuite) SetupTest() {
	suite.osrmMock.FindFastestRoutesFunc = nil
}

func (suite *testSuite) TearDownSuite() {
	suite.NoError(suite.server.Stop())
}
