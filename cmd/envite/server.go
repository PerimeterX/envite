// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that

package main

import (
	"errors"
	"github.com/perimeterx/envite"
)

// buildServer initializes a new server instance for the ENVITE environment if the execution mode is set to daemon.
// This function checks if the necessary flags for server operation are provided and returns an error if required
// information is missing.
//
// Returns:
//   - *envite.Server: A pointer to an initialized envite.Server instance ready to handle requests based on the
//     provided environment configuration. This return value is nil if the execution mode is not
//     set to daemon, indicating that a server instance is not required.
//   - error: An error if initializing the server fails due to missing required information (e.g., the port number
//     in daemon mode) or any other issue that prevents the server from being properly initialized.
//
// Note: It is critical to provide the port flag when running in daemon mode, as the server needs to bind to a
//
//	specific port to accept incoming connections. If the port flag is not set, the function returns an error
//	to prevent the application from proceeding without a valid server configuration.
func buildServer(env *envite.Environment, flags flagValues) (*envite.Server, error) {
	if flags.mode != envite.ExecutionModeDaemon {
		return nil, nil
	}

	if !flags.port.exist {
		return nil, errors.New("in daemon execution mode, port flag is required")
	}

	return envite.NewServer(flags.port.value, env), nil
}
