package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type Network struct {
	Client         *client.Client
	envID          string
	ID             string
	OnNewComponent func(*Config)

	lock         sync.Mutex
	shouldDelete bool
	configure    func(config Config, runConfig *runConfig, containerName string)
}

func NewNetwork(cli *client.Client, networkIdentifier, envID string) (*Network, error) {
	if networkIdentifier != "" {
		return newClosedNetwork(cli, envID, networkIdentifier)
	} else if runtime.GOOS == "linux" {
		return newOpenLinuxNetwork(cli, envID)
	} else {
		return newOpenNetwork(cli, envID)
	}
}

func (n *Network) NewComponent(config Config) (*Component, error) {
	if n.OnNewComponent != nil {
		n.OnNewComponent(&config)
	}
	return newComponent(n.Client, n.envID, n, config)
}

func newClosedNetwork(cli *client.Client, envID, networkIdentifier string) (*Network, error) {
	networks, err := cli.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	nw, err := findNetwork(networks, networkIdentifier)
	if err != nil {
		return nil, err
	}

	return &Network{
		Client:       cli,
		envID:        envID,
		shouldDelete: false,
		ID:           nw.ID,
		configure: func(config Config, runConfig *runConfig, containerName string) {
			runConfig.hostConfig.NetworkMode = container.NetworkMode(nw.Driver)
			runConfig.networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{nw.ID: {NetworkID: nw.ID}},
			}
			runConfig.hostname = containerName
		},
	}, nil
}

func newOpenLinuxNetwork(cli *client.Client, envID string) (*Network, error) {
	id, err := createNetworkIfNotExist(cli, envID, "host")
	if err != nil {
		return nil, err
	}

	return &Network{
		Client:       cli,
		envID:        envID,
		shouldDelete: true,
		ID:           id,
		configure: func(config Config, runConfig *runConfig, containerName string) {
			runConfig.hostConfig.NetworkMode = "host"
			runConfig.networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{id: {NetworkID: id}},
			}
			runConfig.hostname = "localhost"
		},
	}, nil
}

func newOpenNetwork(cli *client.Client, envID string) (*Network, error) {
	err := validateHostsFile()
	if err != nil {
		return nil, err
	}

	id, err := createNetworkIfNotExist(cli, envID, "bridge")
	if err != nil {
		return nil, err
	}

	return &Network{
		Client:       cli,
		envID:        envID,
		shouldDelete: true,
		ID:           id,
		configure: func(config Config, runConfig *runConfig, containerName string) {
			runConfig.hostConfig.NetworkMode = "bridge"
			runConfig.networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{id: {NetworkID: id}},
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
	err := c.cli.NetworkRemove(ctx, c.envID)
	if err != nil &&
		!strings.Contains(err.Error(), "has active endpoints") &&
		!strings.Contains(err.Error(), "not found") {
		return err
	}

	return nil
}

func createNetworkIfNotExist(cli *client.Client, name, driver string) (string, error) {
	res, err := cli.NetworkCreate(context.Background(), name, types.NetworkCreate{
		CheckDuplicate: true,
		Driver:         driver,
	})
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return "", err
		}

		return name, nil
	}

	return res.ID, nil
}

func findNetwork(networks []types.NetworkResource, identifier string) (types.NetworkResource, error) {
	for _, current := range networks {
		if current.ID == identifier || current.Name == identifier {
			return current, nil
		}
	}
	return types.NetworkResource{}, ErrNetworkNotExist{network: identifier}
}

func validateHostsFile() error {
	valid, err := isHostsFileValid()
	if err != nil {
		return err
	}

	if valid {
		return nil
	}

	err = updateHostsFile()
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") {
			fmt.Println("missing permissions to add a required entry to /etc/hosts file.\n" +
				"without it, docker networking will not work as expected.\n" +
				"to fix that, either rerun with sudo permissions, " +
				"or manually add the following line to your /etc/hosts file:\n" +
				"127.0.0.1 host.docker.internal\n" +
				"and then rerun normally.")
			os.Exit(1)
		}

		return err
	}

	return nil
}

var hostsEntryRE = regexp.MustCompile(`127\.0\.0\.1\s+host\.docker\.internal`)

func isHostsFileValid() (bool, error) {
	data, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if hostsEntryRE.MatchString(strings.TrimSpace(line)) {
			return true, nil
		}
	}

	return false, nil
}

func updateHostsFile() error {
	f, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = f.WriteString("127.0.0.1 host.docker.internal\n")
	if err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}

type ErrNetworkNotExist struct {
	network string
}

func (e ErrNetworkNotExist) Error() string {
	return fmt.Sprintf("network %s does not exist", e.network)
}
