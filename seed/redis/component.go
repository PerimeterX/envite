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
	"time"
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

	err = client.FlushAll(ctx).Err()
	if err != nil {
		return err
	}

	for _, hashData := range r.config.Data {
		var count int
		err = client.HSet(ctx, hashData.Key, hashData.Fields).Err()

		if err != nil {
			return err
		}
		if hashData.TTL > 0 {
			err = client.Expire(ctx, hashData.Key, time.Duration(hashData.TTL)*time.Second).Err()
			if err != nil {
				return err
			}
		}
		count = len(hashData.Fields)

		r.writer.WriteString(fmt.Sprintf(
			"inserted %s fields to %s",
			r.writer.Color.Green(strconv.Itoa(count)),
			r.writer.Color.Green(hashData.Key),
		))
	}

	return nil
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
