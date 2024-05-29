// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/perimeterx/envite"
	"strconv"
	"sync"
	"sync/atomic"
)

// ComponentType represents the type of the redis seed component.
const ComponentType = "redis seed"

// SeedComponent is a component for seeding redis with data.
type SeedComponent struct {
	lock   sync.Mutex
	config SeedConfig
	status atomic.Value
	writer *envite.Writer
}

// NewSeedComponent creates a new SeedComponent instance.
func NewSeedComponent(config SeedConfig) *SeedComponent {
	r := &SeedComponent{config: config}
	r.status.Store(envite.ComponentStatusStopped)
	return r
}

func (r *SeedComponent) Type() string {
	return ComponentType
}

func (r *SeedComponent) AttachEnvironment(_ context.Context, _ *envite.Environment, writer *envite.Writer) error {
	r.writer = writer
	return nil
}

func (r *SeedComponent) Prepare(context.Context) error {
	return nil
}

func (r *SeedComponent) Start(ctx context.Context) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.status.Store(envite.ComponentStatusStarting)

	err := r.seed(ctx)
	if err != nil {
		r.status.Store(envite.ComponentStatusFailed)
		return err
	}

	r.status.Store(envite.ComponentStatusFinished)

	return nil
}

func (r *SeedComponent) seed(ctx context.Context) error {
	r.writer.WriteString("starting redis seed")
	client, err := r.clientProvider()
	if err != nil {
		return err
	}

	if err = client.FlushAll(ctx).Err(); err != nil {
		return err
	}

	if err = r.setEntries(ctx, err, client); err != nil {
		return err
	}

	if err = r.hashSetEntries(ctx, err, client); err != nil {
		return err
	}
	r.logInsertions()

	return nil
}

func (r *SeedComponent) setEntries(ctx context.Context, err error, client *redis.Client) error {
	for _, entry := range r.config.Entries {
		err = client.Set(ctx, entry.Key, entry.Value, entry.TTL).Err()

		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SeedComponent) hashSetEntries(ctx context.Context, err error, client *redis.Client) error {
	for _, hEntry := range r.config.HEntries {
		err = client.HSet(ctx, hEntry.Key, hEntry.Values).Err()

		if err != nil {
			return err
		}
		if hEntry.TTL > 0 {
			err = client.Expire(ctx, hEntry.Key, hEntry.TTL).Err()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *SeedComponent) logInsertions() {
	count := len(r.config.Entries)
	hashedCount := len(r.config.HEntries)

	r.writer.WriteString(fmt.Sprintf(
		"inserted %s fields to %s and %s fields to %s",
		r.writer.Color.Green(strconv.Itoa(count)),
		r.writer.Color.Green("Entries"),
		r.writer.Color.Green(strconv.Itoa(hashedCount)),
		r.writer.Color.Green("Hashed Entries"),
	))
}

func (r *SeedComponent) clientProvider() (*redis.Client, error) {
	if r.config.ClientProvider != nil {
		return r.config.ClientProvider()
	}

	options, err := redis.ParseURL(r.config.Address)
	if err != nil {
		return nil, err
	}

	return redis.NewClient(options), nil
}

func (r *SeedComponent) Stop(context.Context) error {
	r.status.Store(envite.ComponentStatusStopped)
	return nil
}

func (r *SeedComponent) Cleanup(context.Context) error {
	return nil
}

func (r *SeedComponent) Status(context.Context) (envite.ComponentStatus, error) {
	return r.status.Load().(envite.ComponentStatus), nil
}

func (r *SeedComponent) Config() any {
	return r.config
}
