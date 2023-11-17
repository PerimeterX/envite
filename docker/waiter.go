package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"regexp"
	"strings"
	"time"
)

func WaitForLog(s string) Waiter {
	return Waiter{
		Type:   WaiterTypeString,
		String: s,
	}
}

func WaitForLogRegex(regexp string) Waiter {
	return Waiter{
		Type:  WaiterTypeRegex,
		Regex: regexp,
	}
}

func WaitForDuration(duration string) Waiter {
	return Waiter{
		Type:     WaiterTypeDuration,
		Duration: duration,
	}
}

type waiterFunc func(ctx context.Context, cli *client.Client, containerID string, isNewContainer bool) error

func validateWaiter(w Waiter) (waiterFunc, error) {
	switch w.Type {
	case WaiterTypeString:
		return func(ctx context.Context, cli *client.Client, containerID string, _ bool) error {
			var reached bool
			err := followLogs(ctx, cli, containerID, func(_ time.Time, text string, _ stdcopy.StdType) (stop bool) {
				reached = strings.Contains(text, w.String)
				return reached
			})
			if err != nil {
				return err
			}

			if reached {
				return nil
			}

			return ErrContainerStopped{without: fmt.Sprintf("reaching log '%s'", w.String)}
		}, nil
	case WaiterTypeRegex:
		re, err := regexp.Compile(w.Regex)
		if err != nil {
			return nil, err
		}

		return func(ctx context.Context, cli *client.Client, containerID string, _ bool) error {
			var reached bool
			err := followLogs(ctx, cli, containerID, func(_ time.Time, text string, _ stdcopy.StdType) (stop bool) {
				reached = re.MatchString(text)
				return reached
			})
			if err != nil {
				return err
			}

			if reached {
				return nil
			}

			return ErrContainerStopped{without: fmt.Sprintf("reaching log regex '%s'", w.String)}
		}, nil
	case WaiterTypeDuration:
		d, err := time.ParseDuration(w.Duration)
		if err != nil {
			return nil, err
		}

		return func(_ context.Context, _ *client.Client, _ string, isNewContainer bool) error {
			if !isNewContainer {
				return nil
			}

			time.Sleep(d)
			return nil
		}, nil
	}

	return nil, ErrInvalidWaiterType{Type: w.Type}
}

type ErrInvalidWaiterType struct {
	Type WaiterType
}

func (e ErrInvalidWaiterType) Error() string {
	return fmt.Sprintf("invalid waiter type %s", e.Type)
}

type ErrContainerStopped struct {
	without string
}

func (e ErrContainerStopped) Error() string {
	return fmt.Sprintf("container stopped without %s", e.without)
}
