package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"runtime"
	"strings"
	"sync"
)

type NetworkMode string

const (
	NetworkModeClosed NetworkMode = "closed"
	NetworkModeOpen   NetworkMode = "open"
)

func ParseNetworkMode(value string) (NetworkMode, error) {
	switch strings.ToLower(value) {
	case "closed":
		return NetworkModeClosed, nil
	case "open", "":
		return NetworkModeOpen, nil
	}
	return "", ErrInvalidNetworkMode{v: value}
}

func validateNetworkMode(networkMode NetworkMode, containerName string) (host string, err error) {
	if networkMode == NetworkModeClosed {
		return containerName, nil
	}
	if networkMode == NetworkModeOpen {
		if runtime.GOOS == "linux" {
			return "localhost", nil
		}
		return "host.docker.internal", nil
	}
	return "", ErrInvalidNetworkMode{v: string(networkMode)}
}

func configureNetwork(networkMode NetworkMode, config Config, blueprintID string, runConfig *runConfig) {
	if networkMode == NetworkModeClosed {
		configureClosedNetwork(blueprintID, runConfig)
	} else if networkMode == NetworkModeOpen {
		if runtime.GOOS == "linux" {
			configureOpenLinuxNetwork(blueprintID, runConfig)
		} else {
			configureOpenNetwork(config, blueprintID, runConfig)
		}
	}
}

func configureClosedNetwork(blueprintID string, runConfig *runConfig) {
	runConfig.networkCreate = types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
	}
	runConfig.hostConfig.NetworkMode = "bridge"
	runConfig.networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{blueprintID: {NetworkID: blueprintID}},
	}
}

func configureOpenLinuxNetwork(blueprintID string, runConfig *runConfig) {
	runConfig.networkCreate = types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "host",
	}
	runConfig.hostConfig.NetworkMode = "host"
	runConfig.networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{blueprintID: {NetworkID: blueprintID}},
	}
}

func configureOpenNetwork(config Config, blueprintID string, runConfig *runConfig) {
	runConfig.networkCreate = types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
	}
	runConfig.hostConfig.NetworkMode = "bridge"
	runConfig.networkingConfig = &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{blueprintID: {NetworkID: blueprintID}},
	}
	runConfig.containerConfig.ExposedPorts = nat.PortSet{}
	runConfig.hostConfig.PortBindings = nat.PortMap{}
	for _, port := range config.Ports {
		protocol := port.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		p := nat.Port(fmt.Sprintf("%s/%s", port.Port, protocol))
		runConfig.containerConfig.ExposedPorts[p] = struct{}{}
		runConfig.hostConfig.PortBindings[p] = append(runConfig.hostConfig.PortBindings[p], nat.PortBinding{
			HostPort: port.Port,
		})
	}
}

var networkLock sync.Mutex

func createNetwork(ctx context.Context, c *Component) error {
	networkLock.Lock()
	defer networkLock.Unlock()
	_, err := c.cli.NetworkCreate(ctx, c.blueprintID, c.runConfig.networkCreate)
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return err
	}

	return nil
}

func deleteNetwork(ctx context.Context, c *Component) error {
	networkLock.Lock()
	defer networkLock.Unlock()
	err := c.cli.NetworkRemove(ctx, c.blueprintID)
	if err != nil &&
		!strings.Contains(err.Error(), "has active endpoints") &&
		!strings.Contains(err.Error(), "not found") {
		return err
	}

	return nil
}

type ErrInvalidNetworkMode struct {
	v string
}

func (e ErrInvalidNetworkMode) Error() string {
	return fmt.Sprintf("invalid network mode %s", e.v)
}
