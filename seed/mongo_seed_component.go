package seed

import (
	"context"
	"fmt"
	"github.com/perimeterx/fengshui"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
	"sync"
	"sync/atomic"
)

type MongoSeedComponent struct {
	lock           sync.Mutex
	clientProvider func() (*mongo.Client, error)
	config         MongoSeedConfig
	status         atomic.Value
	writer         *fengshui.Writer
}

func NewMongoSeedComponent(
	clientProvider func() (*mongo.Client, error),
	config MongoSeedConfig,
) *MongoSeedComponent {
	m := &MongoSeedComponent{
		clientProvider: clientProvider,
		config:         config,
	}

	m.status.Store(fengshui.ComponentStatusStopped)

	return m
}

func (m *MongoSeedComponent) ID() string {
	return m.config.ID
}

func (m *MongoSeedComponent) Type() string {
	return "mongo seed"
}

func (m *MongoSeedComponent) SetOutputWriter(_ context.Context, writer *fengshui.Writer) error {
	m.writer = writer
	return nil
}

func (m *MongoSeedComponent) Prepare(context.Context) error {
	return nil
}

func (m *MongoSeedComponent) Start(ctx context.Context) error {
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

func (m *MongoSeedComponent) Stop(context.Context) error {
	return nil
}

func (m *MongoSeedComponent) Cleanup(context.Context) error {
	return nil
}

func (m *MongoSeedComponent) Status(context.Context) (fengshui.ComponentStatus, error) {
	return m.status.Load().(fengshui.ComponentStatus), nil
}

func (m *MongoSeedComponent) Config() any {
	return m.config
}

func (m *MongoSeedComponent) EnvVars() map[string]string {
	return nil
}
