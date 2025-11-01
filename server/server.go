package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	log                  logrus.FieldLogger
	router               *http.ServeMux
	stopChan             chan struct{}
}

type Config struct {
	Logger               logrus.FieldLogger
}

func NewServer(config Config) (*Server, error) {
	if config.Logger == nil {
		return nil, errors.New("logger must be specified and cannot be nil")
	}
	s := &Server{
		log:                  config.Logger,
		router:               http.NewServeMux(),
		stopChan:             make(chan struct{}),
	}

	s.SetupRoutes()
	return s, nil
}

func (s *Server) SetupRoutes() {
	s.router.HandleFunc("GET /ping", s.ping())
}

func (s *Server) Serve(listen string) error {
	hs := http.Server{
		Addr:    listen,
		Handler: s.router,
	}

	go func() {
		// Wait for stop signal.
		<-s.stopChan
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
