package docker

import (
	"bufio"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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

type waiterFunc func(cli *client.Client, containerID string) error

func validateWaiter(w Waiter) (waiterFunc, error) {
	switch w.Type {
	case WaiterTypeString:
		return func(cli *client.Client, containerID string) error {
			containerReader, err := cli.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Timestamps: true,
				Follow:     true,
			})
			if err != nil {
				return err
			}

			scanner := bufio.NewScanner(containerReader)
			for scanner.Scan() {
				if strings.Contains(scanner.Text(), w.String) {
					return nil
				}
			}

			_ = containerReader.Close()

			return ErrContainerStopped{without: fmt.Sprintf("reaching log '%s'", w.String)}
		}, nil
	case WaiterTypeRegex:
		re, err := regexp.Compile(w.Regex)
		if err != nil {
			return nil, err
		}

		return func(cli *client.Client, containerID string) error {
			containerReader, err := cli.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Timestamps: true,
				Follow:     true,
			})
			if err != nil {
				return err
			}

			scanner := bufio.NewScanner(containerReader)
			for scanner.Scan() {
				if re.MatchString(scanner.Text()) {
					return nil
				}
			}

			_ = containerReader.Close()

			return ErrContainerStopped{without: fmt.Sprintf("reaching log regex '%s'", w.String)}
		}, nil
	case WaiterTypeDuration:
		d, err := time.ParseDuration(w.Duration)
		if err != nil {
			return nil, err
		}

		return func(cli *client.Client, containerID string) error {
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
