// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"strings"
)

// Environment represents a collection of components that can be managed together.
// Components within an environment can be started, stopped, and configured collectively or individually.
type Environment struct {
	id             string
	components     [][]Component
	componentsByID map[string]Component
	outputManager  *outputManager
	Logger         Logger
}

// NewEnvironment creates and initializes a new Environment with the specified id and component graph.
// It returns an error if the id is empty, the graph is nil, or if any components are misconfigured.
func NewEnvironment(id string, componentGraph *ComponentGraph, options ...Option) (*Environment, error) {
	if id == "" {
		return nil, ErrEmptyEnvID
	}

	if componentGraph == nil {
		return nil, ErrNilGraph
	}

	id = strings.ReplaceAll(id, " ", "_")

	om := newOutputManager()
	b := &Environment{
		id:             id,
		components:     componentGraph.components,
		componentsByID: make(map[string]Component),
		outputManager:  om,
	}

	for _, concurrentComponents := range componentGraph.components {
		for _, component := range concurrentComponents {
			componentID := component.ID()
			if componentID == "" {
				return nil, ErrInvalidComponentID{msg: "component id may not be empty"}
			}
			if strings.Contains(componentID, "|") || strings.Contains(componentID, " ") {
				return nil, ErrInvalidComponentID{id: componentID, msg: "component id may not contain '|' or ' '"}
			}

			_, exists := b.componentsByID[componentID]
			if exists {
				return nil, ErrInvalidComponentID{id: componentID, msg: "duplicate component id"}
			}

			err := component.AttachEnvironment(context.Background(), b, om.writer(component.ID()))
			if err != nil {
				return nil, err
			}

			b.componentsByID[componentID] = component
		}
	}

	for _, option := range options {
		option(b)
	}
	if b.Logger == nil {
		b.Logger = func(LogLevel, string) {}
	}

	return b, nil
}

// Components returns a slice of all components within the environment.
func (b *Environment) Components() []Component {
	result := make([]Component, 0, len(b.componentsByID))
	for _, component := range b.componentsByID {
		result = append(result, component)
	}
	return result
}

// Apply applies the specified configuration to the environment, enabling only the components with IDs in
// enabledComponentIDs.
// It returns an error if applying the configuration fails.
func (b *Environment) Apply(ctx context.Context, enabledComponentIDs []string) error {
	b.Logger(LogLevelInfo, "applying state")
	enabledComponents := make(map[string]struct{}, len(enabledComponentIDs))
	for _, id := range enabledComponentIDs {
		enabledComponents[id] = struct{}{}
	}
	err := b.apply(ctx, enabledComponents)
	if err != nil {
		return err
	}

	b.Logger(LogLevelInfo, "finished applying state")
	return nil
}

// StartAll starts all components in the environment concurrently.
// It returns an error if starting any component fails.
func (b *Environment) StartAll(ctx context.Context) error {
	b.Logger(LogLevelInfo, "starting all")
	all := make(map[string]struct{}, len(b.componentsByID))
	for id := range b.componentsByID {
		all[id] = struct{}{}
	}
	err := b.apply(ctx, all)
	if err != nil {
		return err
	}

	b.Logger(LogLevelInfo, "finished starting all")
	return nil
}

// StopAll stops all components in the environment in reverse order of their startup.
// It returns an error if stopping any component fails.
func (b *Environment) StopAll(ctx context.Context) error {
	b.Logger(LogLevelInfo, "stopping all")
	for i := len(b.components) - 1; i >= 0; i-- {
		concurrentComponents := b.components[i]
		g, ctx := errgroup.WithContext(ctx)
		for _, component := range concurrentComponents {
			component := component
			g.Go(func() error {
				b.Logger(LogLevelInfo, fmt.Sprintf("stopping %s", component.ID()))
				err := component.Stop(ctx)
				if err != nil {
					return fmt.Errorf("could not stop %s: %w", component.ID(), err)
				}

				return nil
			})
		}
		err := g.Wait()
		if err != nil {
			return err
		}
	}

	b.Logger(LogLevelInfo, "finished stopping all")
	return nil
}

// StartComponent starts a single component identified by componentID.
// It does nothing if the component is already running.
// Returns an error if the component fails to start.
func (b *Environment) StartComponent(ctx context.Context, componentID string) error {
	component, err := b.componentByID(componentID)
	if err != nil {
		return err
	}

	status, err := component.Status(ctx)
	if err != nil {
		return err
	}

	if status == ComponentStatusRunning || status == ComponentStatusStarting {
		return nil
	}

	b.Logger(LogLevelInfo, fmt.Sprintf("preparing %s", componentID))
	err = component.Prepare(ctx)
	if err != nil {
		return err
	}

	b.Logger(LogLevelInfo, fmt.Sprintf("starting %s", componentID))
	err = component.Start(ctx)
	if err != nil {
		return err
	}

	b.Logger(LogLevelInfo, fmt.Sprintf("finished starting %s", componentID))
	return nil
}

// StopComponent stops a single component identified by componentID.
// Returns an error if the component fails to stop.
func (b *Environment) StopComponent(ctx context.Context, componentID string) error {
	component, err := b.componentByID(componentID)
	if err != nil {
		return err
	}

	b.Logger(LogLevelInfo, fmt.Sprintf("stopping %s", componentID))
	err = component.Stop(ctx)
	if err != nil {
		return err
	}

	b.Logger(LogLevelInfo, fmt.Sprintf("finished stopping %s", componentID))
	return nil
}

// Status returns the current status of all components within the environment.
func (b *Environment) Status(ctx context.Context) (GetStatusResponse, error) {
	result := GetStatusResponse{ID: b.id, Components: make([][]GetStatusResponseComponent, len(b.components))}
	for i, concurrentComponents := range b.components {
		components := make([]GetStatusResponseComponent, len(concurrentComponents))
		for j, component := range concurrentComponents {
			status, err := component.Status(ctx)
			if err != nil {
				return GetStatusResponse{}, fmt.Errorf("could not get status for %s: %w", component.ID(), err)
			}
			components[j] = GetStatusResponseComponent{
				ID:      component.ID(),
				Type:    component.Type(),
				Status:  status,
				Info:    component.Config(),
				EnvVars: component.EnvVars(),
			}
		}
		result.Components[i] = components
	}
	return result, nil
}

// Output returns a reader for the environment's combined output from all components.
func (b *Environment) Output() *Reader {
	return b.outputManager.reader()
}

// Cleanup performs cleanup operations for all components within the environment.
// It returns an error if cleaning up any component fails.
func (b *Environment) Cleanup(ctx context.Context) error {
	b.Logger(LogLevelInfo, "cleaning up")
	g, ctx := errgroup.WithContext(ctx)
	for _, concurrentComponents := range b.components {
		for _, component := range concurrentComponents {
			component := component
			g.Go(func() error {
				b.Logger(LogLevelInfo, fmt.Sprintf("cleaning up %s", component.ID()))
				err := component.Cleanup(ctx)
				if err != nil {
					return fmt.Errorf("could not cleanup %s: %w", component.ID(), err)
				}

				return nil
			})
		}
	}
	err := g.Wait()
	if err != nil {
		return err
	}

	b.Logger(LogLevelInfo, "finished cleaning up")
	return nil
}

func (b *Environment) apply(ctx context.Context, enabledComponentIDs map[string]struct{}) error {
	err := b.prepare(ctx, enabledComponentIDs)
	if err != nil {
		return err
	}

	for _, concurrentComponents := range b.components {
		g, ctx := errgroup.WithContext(ctx)
		for _, component := range concurrentComponents {
			component := component
			_, ok := enabledComponentIDs[component.ID()]
			if ok {
				g.Go(func() error {
					status, err := component.Status(ctx)
					if err != nil {
						return fmt.Errorf("could not get status for %s: %w", component.ID(), err)
					}

					if status == ComponentStatusRunning || status == ComponentStatusStarting {
						return nil
					}

					b.Logger(LogLevelInfo, fmt.Sprintf("starting %s", component.ID()))
					err = component.Start(ctx)
					if err != nil {
						return fmt.Errorf("could not start %s: %w", component.ID(), err)
					}

					b.Logger(LogLevelInfo, fmt.Sprintf("finished starting %s", component.ID()))
					return nil
				})
			} else {
				g.Go(func() error {
					b.Logger(LogLevelInfo, fmt.Sprintf("stopping %s", component.ID()))
					err := component.Stop(ctx)
					if err != nil {
						return fmt.Errorf("could not stop %s: %w", component.ID(), err)
					}

					b.Logger(LogLevelInfo, fmt.Sprintf("finished stopping %s", component.ID()))
					return nil
				})
			}
		}
		err := g.Wait()
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Environment) prepare(ctx context.Context, enabledComponentIDs map[string]struct{}) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, concurrentComponents := range b.components {
		for _, component := range concurrentComponents {
			_, ok := enabledComponentIDs[component.ID()]
			if !ok {
				continue
			}
			component := component
			g.Go(func() error {
				status, err := component.Status(ctx)
				if err != nil {
					return fmt.Errorf("could not get status for %s: %w", component.ID(), err)
				}

				if status == ComponentStatusRunning || status == ComponentStatusStarting {
					return nil
				}

				b.Logger(LogLevelInfo, fmt.Sprintf("preparing %s", component.ID()))
				err = component.Prepare(ctx)
				if err != nil {
					return fmt.Errorf("could not prepare %s: %w", component.ID(), err)
				}

				b.Logger(LogLevelInfo, fmt.Sprintf("finished preparing %s", component.ID()))
				return nil
			})
		}
	}
	return g.Wait()
}

func (b *Environment) componentByID(componentID string) (Component, error) {
	component := b.componentsByID[componentID]
	if component == nil {
		return nil, ErrInvalidComponentID{id: componentID, msg: "not found"}
	}
	return component, nil
}

var (
	// ErrEmptyEnvID indicates that an empty environment ID was provided.
	ErrEmptyEnvID = errors.New("environment ID cannot be empty")

	// ErrNilGraph indicates that a nil component graph was provided.
	ErrNilGraph = errors.New("environment component graph cannot be nil")
)

// ErrInvalidComponentID represents an error when a component ID is invalid.
type ErrInvalidComponentID struct {
	id  string
	msg string
}

func (e ErrInvalidComponentID) Error() string {
	return fmt.Sprintf("component id '%s' is invalid: %s", e.id, e.msg)
}
