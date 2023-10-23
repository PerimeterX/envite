package fengshui

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	addr       string
	blueprint  *Blueprint
	httpServer *http.Server
	errHandler func(string)
}

func NewServer(port string, blueprint *Blueprint) *Server {
	if len(port) > 0 && port[0] != ':' {
		port = ":" + port
	}

	s := &Server{addr: port, blueprint: blueprint}

	router := mux.NewRouter()
	registerRoutes(router, blueprint)
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
