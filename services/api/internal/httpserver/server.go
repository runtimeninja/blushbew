package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

type ServerDeps struct {
	Logger       *slog.Logger
	Addr         string
	Handler      http.Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func New(deps ServerDeps) *Server {
	s := &http.Server{
		Addr:         deps.Addr,
		Handler:      deps.Handler,
		ReadTimeout:  deps.ReadTimeout,
		WriteTimeout: deps.WriteTimeout,
		IdleTimeout:  deps.IdleTimeout,
	}

	return &Server{
		httpServer: s,
		logger:     deps.Logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("http server starting", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.httpServer.Shutdown(ctx)
}
