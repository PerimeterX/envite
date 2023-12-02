package envite

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	addr       string
	env        *Environment
	httpServer *http.Server
	errHandler func(string)
}

func NewServer(port string, env *Environment) *Server {
	if len(port) > 0 && port[0] != ':' {
		port = ":" + port
	}

	s := &Server{addr: port, env: env}

	router := mux.NewRouter()
	registerRoutes(router, env)
	s.httpServer = &http.Server{Addr: port, Handler: router}

	return s
}

func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Close() error {
	s.httpServer.SetKeepAlivesEnabled(false)
	return s.httpServer.Shutdown(context.Background())
}
