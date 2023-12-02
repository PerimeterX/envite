package envite

import (
	"context"
)

type Component interface {
	ID() string
	Type() string
	AttachBlueprint(ctx context.Context, blueprint *Blueprint, writer *Writer) error
	Prepare(ctx context.Context) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Cleanup(ctx context.Context) error
	Status(ctx context.Context) (ComponentStatus, error)
	Config() any
	EnvVars() map[string]string
}

type ComponentStatus string

const (
	ComponentStatusStopped  ComponentStatus = "stopped"
	ComponentStatusFailed   ComponentStatus = "failed"
	ComponentStatusStarting ComponentStatus = "starting"
	ComponentStatusRunning  ComponentStatus = "running"
	ComponentStatusFinished ComponentStatus = "finished"
)
