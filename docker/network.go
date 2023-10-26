package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"runtime"
	"strings"
	"sync"
)

type Network struct {
	lock         sync.Mutex
	cli          *client.Client
	blueprintID  string
	shouldDelete bool
	id           string
	configure    func(config Config, runConfig *runConfig, containerName string)
}

func NewNetwork(cli *client.Client, networkIdentifier, blueprintID string) (*Network, error) {
	if networkIdentifier != "" {
		return newClosedNetwork(cli, blueprintID, networkIdentifier)
	} else if runtime.GOOS == "linux" {
		return newOpenLinuxNetwork(cli, blueprintID)
	} else {
		return newOpenNetwork(cli, blueprintID)
	}
}

func (n *Network) NewComponent(config Config) (*Component, error) {
	return newComponent(n.cli, n.blueprintID, n, config)
}

func newClosedNetwork(cli *client.Client, blueprintID, networkIdentifier string) (*Network, error) {
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	nw, err := findNetwork(networks, networkIdentifier)
	if err != nil {
		return nil, err
	}

	return &Network{
		cli:          cli,
		blueprintID:  blueprintID,
		shouldDelete: false,
		id:           nw.ID,
		configure: func(config Config, runConfig *runConfig, containerName string) {
			runConfig.hostConfig.NetworkMode = container.NetworkMode(nw.Driver)
			runConfig.networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{nw.ID: {NetworkID: nw.ID}},
			}
			runConfig.hostname = containerName
		},
	}, nil
}

func findNetwork(networks []types.NetworkResource, identifier string) (types.NetworkResource, error) {
	for _, current := range networks {
		if current.ID == identifier || current.Name == identifier {
			return current, nil
		}
	}
	return types.NetworkResource{}, ErrNetworkNotExist{network: identifier}
}

func newOpenLinuxNetwork(cli *client.Client, blueprintID string) (*Network, error) {
	res, err := cli.NetworkCreate(context.Background(), blueprintID, types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "host",
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, err
	}

	return &Network{
		cli:          cli,
		blueprintID:  blueprintID,
		shouldDelete: true,
		id:           res.ID,
		configure: func(config Config, runConfig *runConfig, containerName string) {
			runConfig.hostConfig.NetworkMode = "host"
			runConfig.networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{res.ID: {NetworkID: res.ID}},
			}
			runConfig.hostname = "localhost"
		},
	}, nil
}

func newOpenNetwork(cli *client.Client, blueprintID string) (*Network, error) {
	res, err := cli.NetworkCreate(context.Background(), blueprintID, types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         "bridge",
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, err
	}

	return &Network{
		cli:          cli,
		blueprintID:  blueprintID,
		shouldDelete: true,
		id:           res.ID,
		configure: func(config Config, runConfig *runConfig, containerName string) {
			runConfig.hostConfig.NetworkMode = "bridge"
			runConfig.networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{res.ID: {NetworkID: res.ID}},
			}
			runConfig.hostname = "host.docker.internal"
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
		},
	}, nil
}

func (n *Network) delete(ctx context.Context, c *Component) error {
	if !n.shouldDelete {
		return nil
	}

	n.lock.Lock()
	defer n.lock.Unlock()
	err := c.cli.NetworkRemove(ctx, c.blueprintID)
	if err != nil &&
		!strings.Contains(err.Error(), "has active endpoints") &&
		!strings.Contains(err.Error(), "not found") {
		return err
	}

	return nil
}

type ErrNetworkNotExist struct {
	network string
}

func (e ErrNetworkNotExist) Error() string {
	return fmt.Sprintf("network %s does not exist", e.network)
}
