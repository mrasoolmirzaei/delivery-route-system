package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/mrasoolmirzaei/delivery-route-system/service"
	"github.com/sirupsen/logrus"
)

const (
	defaultRequestTimeout  = 30 * time.Second
	defaultShutdownTimeout = 5 * time.Second
)

type Server struct {
	log             logrus.FieldLogger
	router          *http.ServeMux
	stopChan        chan struct{}
	routeService    service.RouteService
	requestTimeout  time.Duration
	shutdownTimeout time.Duration
}

type Config struct {
	Logger          logrus.FieldLogger
	RouteService    service.RouteService
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

func NewServer(config Config) (*Server, error) {
	if config.Logger == nil {
		return nil, errors.New("logger must be specified and cannot be nil")
	}
	if config.RouteService == nil {
		return nil, errors.New("routeService must be specified and cannot be nil")
	}

	requestTimeout := defaultRequestTimeout
	if config.RequestTimeout > 0 {
		requestTimeout = config.RequestTimeout
	}

	shutdownTimeout := defaultShutdownTimeout
	if config.ShutdownTimeout > 0 {
		shutdownTimeout = config.ShutdownTimeout
	}

	s := &Server{
		log:             config.Logger,
		router:          http.NewServeMux(),
		stopChan:        make(chan struct{}),
		routeService:    config.RouteService,
		requestTimeout:  requestTimeout,
		shutdownTimeout: shutdownTimeout,
	}

	s.SetupRoutes()
	return s, nil
}

func (s *Server) SetupRoutes() {
	s.router.HandleFunc("GET /health", s.health())
	s.router.HandleFunc("GET /routes", s.getRoutes())
}

func (s *Server) Serve(listen string) error {
	handler := s.recoveryMiddleware(s.router)
	handler = s.timeoutMiddleware(handler)
	handler = s.loggingMiddleware(handler)

	hs := http.Server{
		Addr:              listen,
		Handler:           handler,
		ReadTimeout:       s.requestTimeout,
		WriteTimeout:      s.requestTimeout,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		// Wait for stop signal.
		<-s.stopChan
		ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
		defer cancel()

		s.log.Info("Shutting down HTTP server.")
		if err := hs.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
			s.log.WithError(err).Error("failed to shutdown HTTP server")
		}
	}()

	if err := hs.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	select {
	case <-s.stopChan:
		// Already closed. Don't close again.
	default:
		// Safe to close here. We're the only closer.
		close(s.stopChan)
	}

	return nil
}
