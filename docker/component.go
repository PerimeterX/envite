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
	"time"
)

const ComponentType = "docker component"

type Component struct {
	Host          string
	ContainerName string
	Writer        *fengshui.Writer

	lock             sync.Mutex
	blueprintID      string
	cli              *client.Client
	config           Config
	networkMode      NetworkMode
	runConfig        *runConfig
	latestLogMessage time.Time
}

func NewComponent(
	cli *client.Client,
	blueprintID string,
	networkMode NetworkMode,
	config Config,
) (*Component, error) {
	runConf, err := config.validate(networkMode, blueprintID)
	if err != nil {
		return nil, err
	}

	containerName := fmt.Sprintf("%s_%s", blueprintID, config.Name)

	host, err := validateNetworkMode(networkMode, containerName)
	if err != nil {
		return nil, err
	}

	return &Component{
		cli:           cli,
		config:        config,
		blueprintID:   blueprintID,
		networkMode:   networkMode,
		runConfig:     runConf,
		Host:          host,
		ContainerName: containerName,
	}, nil
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

	err = c.followLogs(cont.ID)
	if err != nil {
		return err
	}

	return nil
}

func (c *Component) Prepare(ctx context.Context) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	err := createNetwork(ctx, c)
	if err != nil {
		return err
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
		c.ContainerName,
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

	err = c.cli.ContainerStart(ctx, id, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	err = c.followLogs(id)
	if err != nil {
		return err
	}

	for _, waiter := range c.runConfig.waiters {
		err = waiter(c.cli, id)
		if err != nil {
			return err
		}
	}

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

	return c.cli.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{Force: true})
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

	err = deleteNetwork(ctx, c)
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

	if cont != nil && cont.State == "running" {
		return fengshui.ComponentStatusRunning, nil
	}

	return fengshui.ComponentStatusStopped, nil
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
		Filters: filters.NewArgs(filters.Arg("name", c.ContainerName)),
	})
	if err != nil {
		return nil, err
	}

	for _, co := range containers {
		if len(co.Names) > 0 && co.Names[0][1:] == c.ContainerName {
			return &co, nil
		}
	}

	return nil, nil
}

func (c *Component) followLogs(id string) error {
	containerReader, err := c.cli.ContainerLogs(context.Background(), id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(containerReader)
		for scanner.Scan() {
			bytes := scanner.Bytes()
			var text string
			var t time.Time
			if len(bytes) > 8 {
				t, text = extractMessageTime(string(bytes[8:]))
				stream := stdcopy.StdType(bytes[0])
				if stream == stdcopy.Stderr {
					// stderr
					text = c.Writer.Color.Red(text)
				} else if stream == stdcopy.Systemerr {
					// docker system error, restart consume logs
					_ = containerReader.Close()
					err = c.followLogs(id)
					if err != nil {
						msg := fmt.Sprintf("failed to consume container logs: %v", err)
						c.Writer.WriteString(c.Writer.Color.Red(msg))
					}
					return
				}
			} else {
				t, text = extractMessageTime(string(bytes))
			}
			if t.Before(c.latestLogMessage) {
				continue
			}
			c.latestLogMessage = t
			c.Writer.WriteStringWithTime(t, text)
		}

		_ = containerReader.Close()
	}()
	return nil
}

func extractMessageTime(message string) (time.Time, string) {
	pos := strings.Index(message, " ")
	if pos > -1 {
		t, err := time.Parse(time.RFC3339Nano, message[:pos])
		if err == nil {
			return t, message[pos+1:]
		}
	}
	return time.Now(), message
}
