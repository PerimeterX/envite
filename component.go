// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

import (
	"context"
)

// Component defines the interface for an environment component.
// It includes methods for lifecycle management, configuration, and status reporting.
type Component interface {
	// Type returns the type of the component.
	Type() string

	// AttachEnvironment associates the component with an environment and output writer.
	// It allows the component to interact with its environment and handle output properly.
	AttachEnvironment(ctx context.Context, env *Environment, writer *Writer) error

	// Prepare readies the component for operation. This may involve pre-start configuration or checks.
	Prepare(ctx context.Context) error

	// Start initiates the component's operation.
	// It should return any errors if encountered during startup.
	Start(ctx context.Context) error

	// Stop halts the component's operation.
	// It should return any errors if encountered during stop.
	Stop(ctx context.Context) error

	// Cleanup performs any necessary cleanup operations for the component,
	// such as removing temporary files or releasing external resources.
	Cleanup(ctx context.Context) error

	// Status reports the current operational status of the component.
	Status(ctx context.Context) (ComponentStatus, error)

	// Config returns the configuration of the component.
	// The exact return type can vary between component types.
	Config() any
}

// ComponentStatus represents the operational status of a component within the environment.
type ComponentStatus string

const (
	// ComponentStatusStopped indicates that the component is not currently running.
	ComponentStatusStopped ComponentStatus = "stopped"

	// ComponentStatusFailed indicates that the component has encountered an error and cannot continue operation.
	ComponentStatusFailed ComponentStatus = "failed"

	// ComponentStatusStarting indicates that the component is in the process of starting up but is not yet fully operational.
	ComponentStatusStarting ComponentStatus = "starting"

	// ComponentStatusRunning indicates that the component is currently operational and running as expected.
	ComponentStatusRunning ComponentStatus = "running"

	// ComponentStatusFinished indicates that the component has completed its operation successfully and has stopped running.
	ComponentStatusFinished ComponentStatus = "finished"
)
