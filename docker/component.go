package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/perimeterx/envite"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const ComponentType = "docker component"

type Component struct {
	lock             sync.Mutex
	envID            string
	cli              *client.Client
	config           Config
	runConfig        *runConfig
	network          *Network
	latestLogMessage time.Time
	containerName    string
	status           atomic.Value
	env              *envite.Environment
	writer           *envite.Writer
}

func newComponent(
	cli *client.Client,
	envID string,
	network *Network,
	config Config,
) (*Component, error) {
	runConf, err := config.initialize(network)
	if err != nil {
		return nil, err
	}

	containerName := fmt.Sprintf("%s_%s", envID, config.Name)
	network.configure(config, runConf, containerName)

	c := &Component{
		cli:           cli,
		config:        config,
		envID:         envID,
		runConfig:     runConf,
		network:       network,
		containerName: containerName,
	}

	c.status.Store(envite.ComponentStatusStopped)

	return c, nil
}

func (c *Component) ID() string {
	return c.config.Name
}

func (c *Component) Type() string {
	return ComponentType
}

func (c *Component) AttachEnvironment(ctx context.Context, env *envite.Environment, writer *envite.Writer) error {
	c.env = env
	c.writer = writer

	cont, err := c.findContainer(ctx)
	if err != nil {
		return err
	}
	if cont == nil {
		return nil
	}

	go c.writeLogs(cont.ID)
	go c.monitorStartingStatus(cont.ID, false)

	return nil
}

func (c *Component) Prepare(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	for _, mount := range c.config.Mounts {
		if mount.OnMount != nil {
			mount.OnMount()
		}
	}

	if c.config.ImagePullOptions != nil && c.config.ImagePullOptions.Disabled {
		c.Writer().WriteString(fmt.Sprintf("image pull disabled"))
		return nil
	}

	opts, err := c.config.imagePullOptions()
	if err != nil {
		return err
	}

	reader, err := c.cli.ImagePull(ctx, c.config.Image, opts)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		bytes := scanner.Bytes()
		msg := jsonmessage.JSONMessage{}
		err = json.Unmarshal(bytes, &msg)
		if err != nil {
			return fmt.Errorf("failed to parse image pull output: %w", err)
		}

		if msg.Progress == nil || msg.Progress.Total == 0 {
			if msg.ID == "" {
				c.Writer().WriteString(msg.Status)
			} else {
				c.Writer().WriteString(fmt.Sprintf(
					"%s %s",
					c.Writer().Color.Cyan(msg.ID),
					msg.Status,
				))
			}
		} else {
			c.Writer().WriteString(fmt.Sprintf(
				"%s %s %d%%",
				c.Writer().Color.Cyan(msg.ID),
				msg.Status,
				int(math.Ceil(float64(msg.Progress.Current)/float64(msg.Progress.Total)*100)),
			))
		}
	}

	return reader.Close()
}

func (c *Component) Start(ctx context.Context) error {
	id, err := c.startContainer(ctx)
	if err != nil {
		return err
	}

	c.monitorStartingStatus(id, true)
	return nil
}

func (c *Component) startContainer(ctx context.Context) (string, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var id string
	res, err := c.cli.ContainerCreate(
		ctx,
		c.runConfig.containerConfig,
		c.runConfig.hostConfig,
		c.runConfig.networkingConfig,
		c.runConfig.platformConfig,
		c.containerName,
	)
	if err == nil {
		id = res.ID
	} else if !errdefs.IsConflict(err) {
		return "", err
	} else {
		cont, err := c.findContainer(ctx)
		if err != nil {
			return "", err
		}

		id = cont.ID
	}

	err = c.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
	if err != nil {
		return "", err
	}

	go c.writeLogs(id)
	return id, nil
}

func (c *Component) Stop(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	cont, err := c.findContainer(ctx)
	if err != nil {
		return err
	}

	if cont == nil {
		return nil
	}

	err = c.cli.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return err
	}

	c.status.Store(envite.ComponentStatusStopped)
	return nil
}

func (c *Component) Cleanup(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.config.ImagePullOptions != nil && c.config.ImagePullOptions.Disabled {
		c.Writer().WriteString(fmt.Sprintf("image remove disabled"))
		return nil
	}

	_, err := c.cli.ImageRemove(ctx, c.config.Image, types.ImageRemoveOptions{})
	if errdefs.IsNotFound(err) {
		return nil
	}

	err = c.network.delete(ctx, c)
	if err != nil {
		return err
	}
	return err
}

func (c *Component) Status(context.Context) (envite.ComponentStatus, error) {
	status := c.status.Load().(envite.ComponentStatus)

	if status == envite.ComponentStatusRunning {
		// check if container stopped
		cont, err := c.findContainer(context.Background())
		if err != nil {
			return "", err
		}

		if cont == nil || cont.State != "running" {
			status = envite.ComponentStatusStopped
			c.status.Store(envite.ComponentStatusStopped)
		}
	}

	return status, nil
}

func (c *Component) monitorStartingStatus(containerID string, isNewContainer bool) {
	c.status.Store(envite.ComponentStatusStarting)
	for _, waiter := range c.runConfig.waiters {
		err := waiter(context.Background(), c.cli, containerID, isNewContainer)
		if err != nil {
			// container might have been manually stopped while we waited
			c.lock.Lock()
			if c.status.Load() == envite.ComponentStatusStarting {
				c.status.Store(envite.ComponentStatusFailed)
			}
			c.lock.Unlock()
			return
		}
	}
	c.status.Store(envite.ComponentStatusRunning)
}

func (c *Component) Config() any {
	return c.config
}

func (c *Component) EnvVars() map[string]string {
	return c.config.Env
}

func (c *Component) Exec(ctx context.Context, cmd []string) (int, error) {
	cont, err := c.findContainer(ctx)
	if err != nil {
		return 0, err
	}

	c.Writer().WriteString(c.Writer().Color.Cyan(fmt.Sprintf("executing: %s", strings.Join(cmd, " "))))
	response, err := c.cli.ContainerExecCreate(ctx, cont.ID, types.ExecConfig{
		Cmd:          cmd,
		Detach:       false,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return 0, err
	}

	hijack, err := c.cli.ContainerExecAttach(ctx, response.ID, types.ExecStartCheck{})
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(hijack.Reader)
	for scanner.Scan() {
		c.Writer().WriteString(c.Writer().Color.Cyan(fmt.Sprintf("exec output: %s", scanner.Text())))
	}

	hijack.Close()

	execResp, err := c.cli.ContainerExecInspect(ctx, response.ID)
	if err != nil {
		return 0, err
	}

	c.Writer().WriteString(c.Writer().Color.Cyan(fmt.Sprintf("exit code: %d", execResp.ExitCode)))
	return execResp.ExitCode, nil
}

func (c *Component) findContainer(ctx context.Context) (*types.Container, error) {
	containers, err := c.cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", c.containerName)),
	})
	if err != nil {
		return nil, err
	}

	for _, co := range containers {
		if len(co.Names) > 0 && co.Names[0][1:] == c.containerName {
			return &co, nil
		}
	}

	return nil, nil
}

func (c *Component) writeLogs(id string) {
	err := followLogs(
		context.Background(),
		c.cli,
		id,
		func(timestamp time.Time, text string, stream stdcopy.StdType) (stop bool) {
			if timestamp.Before(c.latestLogMessage) {
				return false
			}

			if stream == stdcopy.Stderr {
				text = c.Writer().Color.Red(text)
			}

			c.latestLogMessage = timestamp
			c.Writer().WriteStringWithTime(timestamp, text)
			return false
		},
	)
	if err != nil {
		c.Logger()(envite.LogLevelError, "could not read container logs for "+c.ID())
	}
}

func (c *Component) Host() string {
	return c.runConfig.hostname
}

func (c *Component) ContainerName() string {
	return c.containerName
}

func (c *Component) Writer() *envite.Writer {
	return c.writer
}

func (c *Component) Logger() envite.Logger {
	return c.env.Logger
}
