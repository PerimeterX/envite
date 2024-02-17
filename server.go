// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"net/http"
)

// Server is an HTTP server, serving UI and API requests to manage the Environment when running in ExecutionModeDaemon.
type Server struct {
	addr       string
	env        *Environment
	httpServer *http.Server
	errHandler func(string)
}

// NewServer creates a new Server instance for the given Environment.
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

// Start starts the HTTP server.
func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Close gracefully shuts down the HTTP server.
func (s *Server) Close() error {
	s.httpServer.SetKeepAlivesEnabled(false)
	return s.httpServer.Shutdown(context.Background())
}
