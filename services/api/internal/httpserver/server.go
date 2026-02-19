package httpserver

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

type Deps struct {
	Logger       *slog.Logger
	Addr         string
	Router       http.Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func New(deps Deps) *Server {
	s := &http.Server{
		Addr:         deps.Addr,
		Handler:      deps.Router,
		ReadTimeout:  deps.ReadTimeout,
		WriteTimeout: deps.WriteTimeout,
		IdleTimeout:  deps.IdleTimeout,
	}

	return &Server{
		httpServer: s,
		logger:     deps.Logger,
	}
}

func NewRouter(logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Basic middleware-less start; we will add request-id/CORS in Step 2
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	return r
}

func (s *Server) Start() error {
	s.logger.Info("http server starting", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.httpServer.Shutdown(ctx)
}
