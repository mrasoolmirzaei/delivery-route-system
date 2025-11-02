package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrasoolmirzaei/delivery-route-system/pkg/httpclient"
	"github.com/mrasoolmirzaei/delivery-route-system/pkg/osrmclient"
	"github.com/mrasoolmirzaei/delivery-route-system/server"
	"github.com/mrasoolmirzaei/delivery-route-system/service"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	serverPort = ":8000"
)

func main() {
	logger := initLogger()
	routeService := service.NewRouteService(osrmclient.NewOSRMClient(&httpclient.Config{
		Log: logger.WithField("context", "osrmclient"),
	}))
	logger.Info("Creating server...")
	srv, err := server.NewServer(server.Config{
		Logger:              logger.WithField("context", "server"),
		ServiceRouteService: routeService,
	})
	if err != nil {
		logger.WithError(err).Fatal("failed to create server")
		return
	}

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		select {
		case sig := <-sigChan:
			logger.Infof("Received signal, exiting: %s", sig)
			return srv.Stop()
		case <-ctx.Done():
			logger.Infof("Received context cancel signal, exiting: %s", ctx.Err())
			return srv.Stop()
		}
	})

	g.Go(func() error {
		logger.Infof("Starting server on %s", envOrDefault("SERVER_PORT", serverPort, parseString))
		return srv.Serve(envOrDefault("SERVER_PORT", serverPort, parseString))
	})

	err = g.Wait()
	if err != nil {
		logger.WithError(err).Fatal("server error")
	}
}

func initLogger() *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stdout
	log.Level = logrus.DebugLevel

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    true,
		TimestampFormat:  time.RFC3339,
		PadLevelText:     true,
		QuoteEmptyFields: true,
	})

	return log.WithField("context", "main")
}

func envOrDefault[T any](env string, def T, parser func(string) (T, error)) T {
	e, ok := os.LookupEnv(env)
	if !ok {
		return def
	}
	if val, err := parser(e); err == nil {
		return val
	}
	return def
}

func parseString(s string) (string, error) {
	return s, nil
}
