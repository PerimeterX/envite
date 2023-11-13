package fengshui

import (
	"context"
)

type Component interface {
	ID() string
	Type() string
	SetOutputWriter(ctx context.Context, writer *Writer) error
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
