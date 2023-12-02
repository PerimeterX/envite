package envite

import (
	"context"
	"fmt"
)

type ExecutionMode string

const (
	ExecutionModeStart  ExecutionMode = "start"
	ExecutionModeStop   ExecutionMode = "stop"
	ExecutionModeDaemon ExecutionMode = "daemon"
)

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

var asciiArt = `
▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓      ▓▓▓▓▓▓ ▓▓▓▓▓▓        ▓▓▓▓▓▓  ▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓    ▓▓▓▓▓▓ ▓▓▓▓▓▓▓      ▓▓▓▓▓▓▓  ▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓  ▓▓▓▓▓▓▓    ▓▓▓▓▓▓▓   ▓▓        ▓▓▓        ▓▓▓          ▓▓
▓▓ ▓▓▓▓▓▓▓          ▓▓▓▓▓▓▓▓▓▓ ▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓▓    ▓▓        ▓▓▓        ▓▓▓          ▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓     ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓  ▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓   ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓    ▓▓▓▓▓▓▓▓▓▓▓▓▓      ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓  ▓▓
▓▓ ▓▓▓▓▓▓▓          ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓     ▓▓▓▓▓▓▓▓▓▓▓▓      ▓▓        ▓▓▓        ▓▓▓          ▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓  ▓▓▓▓▓▓▓▓▓      ▓▓▓▓▓▓▓▓▓▓       ▓▓        ▓▓▓        ▓▓▓          ▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓    ▓▓▓▓▓▓▓       ▓▓▓▓▓▓▓▓        ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓
▓▓ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓  ▓▓▓▓▓▓     ▓▓▓▓▓▓       ▓▓▓▓▓▓▓         ▓▓        ▓▓▓        ▓▓▓▓▓▓▓▓▓▓▓▓ ▓▓
▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓
`

type ErrInvalidExecutionMode struct {
	v string
}

func (e ErrInvalidExecutionMode) Error() string {
	return fmt.Sprintf("invalid execution mode %s", e.v)
}
