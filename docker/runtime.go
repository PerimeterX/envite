package docker

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/client"
)

// runtimeInfos contains information about all known runtimes
var runtimeInfos = []*RuntimeInfo{
	{
		Runtime:          "Docker Desktop",
		SocketPath:       "/var/run/docker.sock",
		InternalHostname: "host.docker.internal",
	},
	{
		Runtime:          "Colima",
		SocketPath:       "~/.colima/default/docker.sock",
		InternalHostname: "host.lima.internal",
		NetworkLatency:   time.Second * 3,
	},
	{
		Runtime:          "Colima",
		SocketPath:       "~/.colima/docker.sock",
		InternalHostname: "host.lima.internal",
		NetworkLatency:   time.Second * 3,
	},
	{
		Runtime:          "Podman",
		SocketPath:       "~/.podman/podman.sock",
		InternalHostname: "host.containers.internal",
	},
	{
		Runtime:          "Rancher Desktop",
		SocketPath:       "~/.rd/docker.sock",
		InternalHostname: "host.docker.internal",
	},
	{
		Runtime:          "Lima",
		SocketPath:       "~/.lima/docker.sock",
		InternalHostname: "host.lima.internal",
	},
	{
		Runtime:          "OrbStack",
		SocketPath:       "~/.orbstack/run/docker.sock",
		InternalHostname: "host.docker.internal",
	},
	{
		Runtime:          "Minikube",
		SocketPath:       "~/.minikube/docker.sock",
		InternalHostname: "host.minikube.internal",
	},
	{
		Runtime:          "ContainerD",
		SocketPath:       "/run/containerd/containerd.sock",
		InternalHostname: "host.docker.internal",
	},
	{
		Runtime:          "Finch",
		SocketPath:       "~/.finch/docker.sock",
		InternalHostname: "host.docker.internal",
	},
}

// RuntimeInfo contains information about a runtime type
type RuntimeInfo struct {
	// Runtime is the name of the runtime
	Runtime string
	// SocketPath is the path to the docker socket
	SocketPath string
	// InternalHostname is the hostname to use for the docker daemon
	InternalHostname string
	// NetworkLatency some runtimes have a slight latency before network is ready when starting a container
	NetworkLatency time.Duration
}

var (
	mu          sync.Mutex
	runtimeInfo *RuntimeInfo
)

func GetRuntimeInfo() (*RuntimeInfo, error) {
	mu.Lock()
	defer mu.Unlock()

	if runtimeInfo == nil {
		var err error
		runtimeInfo, err = detectRuntime()
		if err != nil {
			return nil, err
		}
		fmt.Printf("ENVITE docker runtime detected: %s.\nIf you want to use another docker runtime, set the DOCKER_HOST environment variable.\n", runtimeInfo.Runtime)
	}
	return runtimeInfo, nil
}

// detectRuntime detects the Docker daemon implementation
func detectRuntime() (*RuntimeInfo, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	for _, runtimeInfo := range runtimeInfos {
		runtimeInfo.SocketPath = strings.ReplaceAll(runtimeInfo.SocketPath, "~/", home+"/")
	}

	// Check DOCKER_HOST env var
	dockerHost := os.Getenv("DOCKER_HOST")
	if dockerHost != "" {
		// Remove unix:// prefix if present
		dockerHost = strings.TrimPrefix(dockerHost, "unix://")

		// Check if the socket path exists
		for _, runtimeInfo := range runtimeInfos {
			if runtimeInfo.SocketPath == dockerHost {
				return runtimeInfo, nil
			}
		}

		// If no match, assume default behavior
		return &RuntimeInfo{
			Runtime:          "Unknown",
			SocketPath:       dockerHost,
			InternalHostname: "host.docker.internal",
		}, nil
	}

	// Try all known runtimes
	for _, runtimeInfo := range runtimeInfos {
		// Check if the socket path exists
		if _, err := os.Stat(runtimeInfo.SocketPath); err == nil {
			// Create a new client with the socket path
			cli, err := client.NewClientWithOpts(client.WithHost("unix://" + runtimeInfo.SocketPath))
			if err == nil {
				// Close the client
				err = cli.Close()
				if err != nil {
					return nil, fmt.Errorf("failed to close client: %w", err)
				}
				// Return the runtime info
				return runtimeInfo, nil
			}
		}
	}

	return nil, fmt.Errorf("could not detect a running Docker-compatible daemon")
}
