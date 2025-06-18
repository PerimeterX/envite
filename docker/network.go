// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package docker

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Network represents a Docker network.
// To create Docker components you must first create a Docker Network, then call NewComponent.
// example:
//
//	network, err := NewNetwork(cli, networkID, envID)
//	component, err := network.NewComponent(dockerComponentConfig)
type Network struct {
	client       *client.Client
	envID        string
	ID           string
	lock         sync.Mutex
	shouldDelete bool
	configure    func(config Config, runConfig *runConfig, containerName string)

	OnNewComponent        func(*Config)
	KeepStoppedContainers bool
}

// NewNetwork creates a new Docker network with given network id and environment id.
// If networkID is not empty, it will look for an existing network with the given id and attach new components to it.
// If networkID is empty, it will create a new open docker network depending on the OS you're running on:
//   - On linux, it will create a network with mode "host" and attach new components to it.
//   - On other types of OS, it will create a network in mode "bridge" and expose ports for all components.
func NewNetwork(cli *client.Client, networkID, envID string) (*Network, error) {
	if networkID != "" {
		return newClosedNetwork(cli, envID, networkID)
	}
	if runtime.GOOS == "linux" {
		return newOpenLinuxNetwork(cli, envID)
	}
	return newOpenNetwork(cli, envID)
}

// NewComponent creates a new Docker component within the network.
func (n *Network) NewComponent(config Config) (*Component, error) {
	if n.OnNewComponent != nil {
		n.OnNewComponent(&config)
	}
	return newComponent(n.client, n.envID, n, config)
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
		client:       cli,
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
		client:       cli,
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
	runtimeInfo, err := GetRuntimeInfo()
	if err != nil {
		return nil, err
	}

	err = validateHostsFile(runtimeInfo)
	if err != nil {
		return nil, err
	}

	id, err := createNetworkIfNotExist(cli, envID, "bridge")
	if err != nil {
		return nil, err
	}

	return &Network{
		client:       cli,
		envID:        envID,
		shouldDelete: true,
		ID:           id,
		configure: func(config Config, runConfig *runConfig, containerName string) {
			runConfig.hostConfig.NetworkMode = "bridge"
			runConfig.networkingConfig = &network.NetworkingConfig{
				EndpointsConfig: map[string]*network.EndpointSettings{id: {NetworkID: id}},
			}
			runConfig.hostname = runtimeInfo.InternalHostname
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

// delete tries to remove the network.
// Silently fails if components are still attached to it, or it has already deleted.
// Returns other errors if encountered.
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
	res, err := cli.NetworkCreate(context.Background(), name, network.CreateOptions{
		Driver: driver,
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

//go:embed setup-needed.txt
var setupNeeded string

//go:embed setup-finished.txt
var setupFinished string

// validateHostsFile installs necessary steps to /etc/hosts if needed
func validateHostsFile(runtimeInfo *RuntimeInfo) error {
	valid, err := isHostsFileValid(runtimeInfo)
	if err != nil {
		return err
	}

	if valid {
		return nil
	}

	err = updateHostsFile(runtimeInfo)
	if err == nil {
		fmt.Println(fmt.Sprintf(setupFinished, runtimeInfo.Runtime))
		os.Exit(0)
	}

	if strings.Contains(err.Error(), "permission denied") {
		fmt.Println(fmt.Sprintf(setupNeeded, runtimeInfo.Runtime, runtimeInfo.InternalHostname))
		os.Exit(1)
	}

	return err
}

func isHostsFileValid(runtimeInfo *RuntimeInfo) (bool, error) {
	data, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return false, err
	}

	hostsEntryRE, err := regexp.Compile(fmt.Sprintf(`^\s*127\.0\.0\.1\s+%s\s*$`, strings.ReplaceAll(runtimeInfo.InternalHostname, ".", "\\.")))
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

func updateHostsFile(runtimeInfo *RuntimeInfo) error {
	f, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	_, err = f.WriteString(fmt.Sprintf("\n# To allow ENVITE to create open docker networks\n127.0.0.1 %s\n# End of section\n", runtimeInfo.InternalHostname))
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
