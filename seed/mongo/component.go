package mongo

import (
	"context"
	"fmt"
	"github.com/perimeterx/fengshui"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"sync"
	"sync/atomic"
)

type SeedComponent struct {
	lock           sync.Mutex
	clientProvider func() (*mongo.Client, error)
	config         SeedConfig
	status         atomic.Value
	writer         *fengshui.Writer
}

func NewSeedComponent(
	clientProvider func() (*mongo.Client, error),
	config SeedConfig,
) *SeedComponent {
	m := &SeedComponent{
		clientProvider: clientProvider,
		config:         config,
	}

	m.status.Store(fengshui.ComponentStatusStopped)

	return m
}

func (m *SeedComponent) ID() string {
	return m.config.ID
}

func (m *SeedComponent) Type() string {
	return "mongo seed"
}

func (m *SeedComponent) SetOutputWriter(_ context.Context, writer *fengshui.Writer) error {
	m.writer = writer
	return nil
}

func (m *SeedComponent) Prepare(context.Context) error {
	return nil
}

func (m *SeedComponent) Start(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.writer.WriteString(fmt.Sprintf("starting mongo seed %s", m.config.ID))
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

	m.status.Store(fengshui.ComponentStatusFinished)

	return nil
}

func (m *SeedComponent) Stop(context.Context) error {
	return nil
}

func (m *SeedComponent) Cleanup(context.Context) error {
	return nil
}

func (m *SeedComponent) Status(context.Context) (fengshui.ComponentStatus, error) {
	return m.status.Load().(fengshui.ComponentStatus), nil
}

func (m *SeedComponent) Config() any {
	return m.config
}

func (m *SeedComponent) EnvVars() map[string]string {
	return nil
}
