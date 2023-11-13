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
	"github.com/perimeterx/fengshui"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const ComponentType = "docker component"

type Component struct {
	Writer *fengshui.Writer

	lock             sync.Mutex
	blueprintID      string
	cli              *client.Client
	config           Config
	runConfig        *runConfig
	network          *Network
	latestLogMessage time.Time
	containerName    string
	status           atomic.Value
	validatedStatus  bool
}

func newComponent(
	cli *client.Client,
	blueprintID string,
	network *Network,
	config Config,
) (*Component, error) {
	runConf, err := config.initialize()
	if err != nil {
		return nil, err
	}

	containerName := fmt.Sprintf("%s_%s", blueprintID, config.Name)
	network.configure(config, runConf, containerName)

	c := &Component{
		cli:           cli,
		config:        config,
		blueprintID:   blueprintID,
		runConfig:     runConf,
		network:       network,
		containerName: containerName,
	}

	c.status.Store(fengshui.ComponentStatusStopped)

	return c, nil
}

func (c *Component) ID() string {
	return c.config.Name
}

func (c *Component) Type() string {
	return ComponentType
}

func (c *Component) SetOutputWriter(ctx context.Context, writer *fengshui.Writer) error {
	c.Writer = writer

	cont, err := c.findContainer(ctx)
	if err != nil {
		return err
	}
	if cont == nil {
		return nil
	}

	c.writeLogs(cont.ID)
	c.monitorStatus(cont.ID)

	return nil
}

func (c *Component) monitorStatus(containerID string) {
	c.status.Store(fengshui.ComponentStatusStarting)
	go func() {
		for _, waiter := range c.runConfig.waiters {
			err := waiter(context.Background(), c.cli, containerID, false)
			if err != nil {
				c.status.Store(fengshui.ComponentStatusFailed)
				return
			}
		}
		c.status.Store(fengshui.ComponentStatusRunning)
	}()
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
		c.Writer.WriteString(fmt.Sprintf("image pull disabled"))
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
				c.Writer.WriteString(msg.Status)
			} else {
				c.Writer.WriteString(fmt.Sprintf(
					"%s %s",
					c.Writer.Color.Cyan(msg.ID),
					msg.Status,
				))
			}
		} else {
			c.Writer.WriteString(fmt.Sprintf(
				"%s %s %d%%",
				c.Writer.Color.Cyan(msg.ID),
				msg.Status,
				int(math.Ceil(float64(msg.Progress.Current)/float64(msg.Progress.Total)*100)),
			))
		}
	}

	return reader.Close()
}

func (c *Component) Start(ctx context.Context) error {
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
		return err
	} else {
		cont, err := c.findContainer(ctx)
		if err != nil {
			return err
		}

		id = cont.ID
	}

	c.status.Store(fengshui.ComponentStatusStarting)

	err = c.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	c.writeLogs(id)

	for _, waiter := range c.runConfig.waiters {
		err = waiter(context.Background(), c.cli, id, true)
		if err != nil {
			c.status.Store(fengshui.ComponentStatusFailed)
			return err
		}
	}

	c.status.Store(fengshui.ComponentStatusRunning)
	return nil
}

func (c *Component) startContainer(ctx context.Context, id string) error {

	return nil
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

	c.status.Store(fengshui.ComponentStatusStopped)
	return nil
}

func (c *Component) Cleanup(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.config.ImagePullOptions != nil && c.config.ImagePullOptions.Disabled {
		c.Writer.WriteString(fmt.Sprintf("image remove disabled"))
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

func (c *Component) Status(ctx context.Context) (fengshui.ComponentStatus, error) {
	cont, err := c.findContainer(ctx)
	if err != nil {
		return "", err
	}

	currentStatus := c.status.Load().(fengshui.ComponentStatus)
	if cont != nil && (cont.State == "running" || cont.State == "created") {
		return currentStatus, nil
	}

	if currentStatus == fengshui.ComponentStatusStarting {
		c.status.Store(fengshui.ComponentStatusFailed)
	} else if currentStatus == fengshui.ComponentStatusRunning {
		c.status.Store(fengshui.ComponentStatusFinished)
	}
	return c.status.Load().(fengshui.ComponentStatus), nil
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

	c.Writer.WriteString(c.Writer.Color.Cyan(fmt.Sprintf("executing: %s", strings.Join(cmd, " "))))
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
		c.Writer.WriteString(c.Writer.Color.Cyan(fmt.Sprintf("exec output: %s", scanner.Text())))
	}

	hijack.Close()

	execResp, err := c.cli.ContainerExecInspect(ctx, response.ID)
	if err != nil {
		return 0, err
	}

	c.Writer.WriteString(c.Writer.Color.Cyan(fmt.Sprintf("exit code: %d", execResp.ExitCode)))
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
	go followLogs(
		context.Background(),
		c.cli,
		id,
		func(timestamp time.Time, text string, stream stdcopy.StdType) (stop bool) {
			if timestamp.Before(c.latestLogMessage) {
				return false
			}

			if stream == stdcopy.Stderr {
				text = c.Writer.Color.Red(text)
			}

			c.latestLogMessage = timestamp
			c.Writer.WriteStringWithTime(timestamp, text)
			return false
		},
	)
}

func (c *Component) Host() string {
	return c.runConfig.hostname
}

func (c *Component) ContainerName() string {
	return c.containerName
}
