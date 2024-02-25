// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package mongo

import (
	"context"
	"fmt"
	"github.com/perimeterx/envite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"sync"
	"sync/atomic"
)

// ComponentType represents the type of the MongoDB seed component.
const ComponentType = "mongo seed"

// SeedComponent is a component for seeding MongoDB with data.
type SeedComponent struct {
	lock   sync.Mutex
	config SeedConfig
	status atomic.Value
	writer *envite.Writer
}

// NewSeedComponent creates a new SeedComponent instance.
func NewSeedComponent(config SeedConfig) *SeedComponent {
	m := &SeedComponent{config: config}
	m.status.Store(envite.ComponentStatusStopped)
	return m
}

func (m *SeedComponent) Type() string {
	return ComponentType
}

func (m *SeedComponent) AttachEnvironment(_ context.Context, _ *envite.Environment, writer *envite.Writer) error {
	m.writer = writer
	return nil
}

func (m *SeedComponent) Prepare(context.Context) error {
	return nil
}

func (m *SeedComponent) Start(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.status.Store(envite.ComponentStatusStarting)

	err := m.seed(ctx)
	if err != nil {
		m.status.Store(envite.ComponentStatusFailed)
		return err
	}

	m.status.Store(envite.ComponentStatusFinished)

	return nil
}

func (m *SeedComponent) seed(ctx context.Context) error {
	m.writer.WriteString("starting mongo seed")
	client, err := m.clientProvider()
	if err != nil {
		return err
	}

	for _, collectionData := range m.config.Data {
		coll := client.Database(collectionData.DB).Collection(collectionData.Collection)
		_, err = coll.DeleteMany(context.Background(), map[string]interface{}{})
		if err != nil {
			return err
		}

		var count int
		if len(collectionData.Documents) > 0 {
			result, err := coll.InsertMany(ctx, collectionData.Documents)
			if err != nil {
				return err
			}

			count = len(result.InsertedIDs)
		}

		m.writer.WriteString(fmt.Sprintf(
			"inserted %s documents to %s:%s",
			m.writer.Color.Green(strconv.Itoa(count)),
			m.writer.Color.Green(collectionData.DB),
			m.writer.Color.Cyan(collectionData.Collection),
		))
	}
	return nil
}

func (m *SeedComponent) clientProvider() (*mongo.Client, error) {
	if m.config.ClientProvider != nil {
		return m.config.ClientProvider()
	}

	return mongo.Connect(context.Background(), options.Client().ApplyURI(m.config.URI))
}

func (m *SeedComponent) Stop(context.Context) error {
	m.status.Store(envite.ComponentStatusStopped)
	return nil
}

func (m *SeedComponent) Cleanup(context.Context) error {
	return nil
}

func (m *SeedComponent) Status(context.Context) (envite.ComponentStatus, error) {
	return m.status.Load().(envite.ComponentStatus), nil
}

func (m *SeedComponent) Config() any {
	return m.config
}
