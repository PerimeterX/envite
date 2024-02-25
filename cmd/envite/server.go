// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that

package main

import (
	"github.com/perimeterx/envite"
)

// defaultPort is the default port used to serve the UI in daemon mode unless
// explicitly provided otherwise via CLI flags.
const defaultPort = "4005"

// buildServer initializes a new server instance for the ENVITE environment if the execution mode is set to daemon.
// This function checks if the necessary flags for server operation are provided and returns an error if required
// information is missing.
//
// Returns a pointer to an initialized envite.Server instance ready to handle requests based on the
// provided environment configuration. This return value is nil if the execution mode is not
// set to daemon, indicating that a server instance is not required
//
// If a port flag is not provided, defaultPort is used.
func buildServer(env *envite.Environment, flags flagValues) *envite.Server {
	if flags.mode != envite.ExecutionModeDaemon {
		return nil
	}

	port := defaultPort
	if flags.port.exist {
		port = flags.port.value
	}

	return envite.NewServer(port, env)
}
