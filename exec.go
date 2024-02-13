// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

import (
	"context"
	_ "embed"
	"fmt"
)

// ExecutionMode represents different execution modes for ENVITE.
// It can be used to specify the behavior when executing ENVITE commands.
type ExecutionMode string

const (
	// ExecutionModeStart indicates the start execution mode, which starts all components in the environment,
	// and then exists.
	ExecutionModeStart ExecutionMode = "start"

	// ExecutionModeStop indicates the stop execution mode, which stops all components in the environment,
	// performs cleanup, and then exits.
	ExecutionModeStop ExecutionMode = "stop"

	// ExecutionModeDaemon indicates the daemon execution mode, which starts ENVITE as a daemon and serving a web UI.
	ExecutionModeDaemon ExecutionMode = "daemon"
)

// ParseExecutionMode parses the provided string value into an ExecutionMode.
// It returns the parsed ExecutionMode or an error if the value is not a valid execution mode.
func ParseExecutionMode(value string) (ExecutionMode, error) {
	switch value {
	case "start":
		return ExecutionModeStart, nil
	case "stop":
		return ExecutionModeStop, nil
	case "daemon", "":
		return ExecutionModeDaemon, nil
	}
	return "", ErrInvalidExecutionMode{v: value}
}

// Execute performs the specified action based on the provided execution mode.
// It takes a Server instance and an ExecutionMode as parameters and executes the corresponding action.
// The available execution modes are ExecutionModeStart, ExecutionModeStop, and ExecutionModeDaemon.
func Execute(server *Server, executionMode ExecutionMode) error {
	switch executionMode {
	case ExecutionModeStart:
		return server.env.StartAll(context.Background())
	case ExecutionModeStop:
		err := server.env.StopAll(context.Background())
		if err != nil {
			return err
		}

		return server.env.Cleanup(context.Background())
	case ExecutionModeDaemon:
		fmt.Printf("%s\nstarting ENVITE daemon for %s at http://localhost%s\n", asciiArt, server.env.id, server.addr)
		return server.Start()
	}
	return ErrInvalidExecutionMode{v: string(executionMode)}
}

//go:embed ascii.txt
var asciiArt string

// ErrInvalidExecutionMode is an error type representing an invalid execution mode.
// It is returned when attempting to parse an unrecognized execution mode.
type ErrInvalidExecutionMode struct {
	v string
}

func (e ErrInvalidExecutionMode) Error() string {
	return fmt.Sprintf("invalid execution mode %s", e.v)
}
