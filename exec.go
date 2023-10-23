package fengshui

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
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
		return server.blueprint.StartAll(context.Background())
	case ExecutionModeStop:
		err := server.blueprint.StopAll(context.Background())
		if err != nil {
			return err
		}

		return server.blueprint.Cleanup(context.Background())
	case ExecutionModeDaemon:
		go func() {
			time.Sleep(time.Second * 2)
			err := openBrowser("http://localhost" + server.addr)
			if err != nil {
				server.errHandler(fmt.Sprintf("could not open browser window: %s", err.Error()))
			}
		}()

		return server.Start()
	}
	return ErrInvalidExecutionMode{v: string(executionMode)}
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

type ErrInvalidExecutionMode struct {
	v string
}

func (e ErrInvalidExecutionMode) Error() string {
	return fmt.Sprintf("invalid execution mode %s", e.v)
}
