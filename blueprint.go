package fengshui

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"strings"
)

type Blueprint struct {
	id             string
	components     [][]Component
	componentsByID map[string]Component
	outputManager  *outputManager
	logger         Logger
}

func NewBlueprint(id string, components [][]Component, options ...Option) (*Blueprint, error) {
	if id == "" {
		id = "generic_blueprint"
	}
	id = strings.ReplaceAll(id, " ", "_")

	om := newOutputManager()
	componentsByID := make(map[string]Component)
	for _, concurrentComponents := range components {
		for _, component := range concurrentComponents {
			componentID := component.ID()
			if componentID == "" {
				return nil, ErrInvalidComponentID{msg: "component id may not be empty"}
			}
			if strings.Contains(componentID, "|") || strings.Contains(componentID, " ") {
				return nil, ErrInvalidComponentID{id: componentID, msg: "component id may not contain '|' or ' '"}
			}

			_, exists := componentsByID[componentID]
			if exists {
				return nil, ErrInvalidComponentID{id: componentID, msg: "duplicate component id"}
			}

			err := component.SetOutputWriter(context.Background(), om.writer(component.ID()))
			if err != nil {
				return nil, err
			}

			componentsByID[componentID] = component
		}
	}

	b := &Blueprint{id: id, components: components, componentsByID: componentsByID, outputManager: om}
	for _, option := range options {
		option(b)
	}
	if b.logger == nil {
		b.logger = func(LogLevel, string) {}
	}

	return b, nil
}

func (b *Blueprint) Components() []Component {
	result := make([]Component, 0, len(b.componentsByID))
	for _, component := range b.componentsByID {
		result = append(result, component)
	}
	return result
}

func (b *Blueprint) Apply(ctx context.Context, enabledComponentIDs []string) error {
	b.logger(LogLevelInfo, "applying state")
	enabledComponents := make(map[string]struct{}, len(enabledComponentIDs))
	for _, id := range enabledComponentIDs {
		enabledComponents[id] = struct{}{}
	}
	return b.apply(ctx, enabledComponents)
}

func (b *Blueprint) StartAll(ctx context.Context) error {
	b.logger(LogLevelInfo, "starting all")
	all := make(map[string]struct{}, len(b.componentsByID))
	for id := range b.componentsByID {
		all[id] = struct{}{}
	}
	return b.apply(ctx, all)
}

func (b *Blueprint) StopAll(ctx context.Context) error {
	b.logger(LogLevelInfo, "stopping all")
	for i := len(b.components) - 1; i >= 0; i-- {
		concurrentComponents := b.components[i]
		g, ctx := errgroup.WithContext(ctx)
		for _, component := range concurrentComponents {
			component := component
			g.Go(func() error {
				b.logger(LogLevelInfo, fmt.Sprintf("stopping %s", component.ID()))
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

	return nil
}

func (b *Blueprint) StartComponent(ctx context.Context, componentID string) error {
	component, err := b.componentByID(componentID)
	if err != nil {
		return err
	}

	status, err := component.Status(ctx)
	if err != nil {
		return err
	}

	if status == ComponentStatusRunning {
		return nil
	}

	b.logger(LogLevelInfo, fmt.Sprintf("preparing %s", componentID))
	err = component.Prepare(ctx)
	if err != nil {
		return err
	}

	b.logger(LogLevelInfo, fmt.Sprintf("starting %s", componentID))
	return component.Start(ctx)
}

func (b *Blueprint) StopComponent(ctx context.Context, componentID string) error {
	component, err := b.componentByID(componentID)
	if err != nil {
		return err
	}

	b.logger(LogLevelInfo, fmt.Sprintf("stopping %s", componentID))
	return component.Stop(ctx)
}

func (b *Blueprint) Status(ctx context.Context) (GetStatusResponse, error) {
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

func (b *Blueprint) Output() *Reader {
	return b.outputManager.reader()
}

func (b *Blueprint) Cleanup(ctx context.Context) error {
	b.logger(LogLevelInfo, "cleaning up")
	g, ctx := errgroup.WithContext(ctx)
	for _, concurrentComponents := range b.components {
		for _, component := range concurrentComponents {
			component := component
			g.Go(func() error {
				b.logger(LogLevelInfo, fmt.Sprintf("cleaning up %s", component.ID()))
				err := component.Cleanup(ctx)
				if err != nil {
					return fmt.Errorf("could not cleanup %s: %w", component.ID(), err)
				}

				return nil
			})
		}
	}
	return g.Wait()
}

func (b *Blueprint) apply(ctx context.Context, enabledComponentIDs map[string]struct{}) error {
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

					if status != ComponentStatusStopped {
						return nil
					}

					b.logger(LogLevelInfo, fmt.Sprintf("starting %s", component.ID()))
					err = component.Start(ctx)
					if err != nil {
						return fmt.Errorf("could not start %s: %w", component.ID(), err)
					}

					return nil
				})
			} else {
				g.Go(func() error {
					b.logger(LogLevelInfo, fmt.Sprintf("stopping %s", component.ID()))
					err := component.Stop(ctx)
					if err != nil {
						return fmt.Errorf("could not stop %s: %w", component.ID(), err)
					}

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

func (b *Blueprint) prepare(ctx context.Context, enabledComponentIDs map[string]struct{}) error {
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

				if status == ComponentStatusRunning {
					return nil
				}

				b.logger(LogLevelInfo, fmt.Sprintf("preparing %s", component.ID()))
				err = component.Prepare(ctx)
				if err != nil {
					return fmt.Errorf("could not prepare %s: %w", component.ID(), err)
				}

				return nil
			})
		}
	}
	return g.Wait()
}

func (b *Blueprint) componentByID(componentID string) (Component, error) {
	component := b.componentsByID[componentID]
	if component == nil {
		return nil, ErrInvalidComponentID{id: componentID, msg: "not found"}
	}
	return component, nil
}

type ErrInvalidComponentID struct {
	id  string
	msg string
}

func (e ErrInvalidComponentID) Error() string {
	return fmt.Sprintf("component id '%s' is invalid: %s", e.id, e.msg)
}
