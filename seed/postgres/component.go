package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/perimeterx/envite"
)

// ComponentType represents the type of the Postgres seed component.
const ComponentType = "postgres seed"

// SeedComponent is a component for seeding Postgres with data.
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

	err := m.Seed()
	if err != nil {
		m.status.Store(envite.ComponentStatusFailed)
		return err
	}

	m.status.Store(envite.ComponentStatusFinished)

	return nil
}

func (m *SeedComponent) Seed() error {
	if m.writer != nil {
		m.writer.WriteString("starting postgres seed")
	}

	client, err := m.clientProvider()
	if err != nil {
		return err
	}

	if _, err = client.Exec(m.config.Setup); err != nil {
		return err
	}

	for _, collection := range m.config.Data {

		if _, err = client.Exec(fmt.Sprintf("DELETE FROM %s", collection.Table)); err != nil {
			return err
		}

		for _, row := range collection.Rows {
			sql, values := generateInsertSQL(collection.Table, row)
			_, err := client.Exec(sql, values...)
			if err != nil {
				return err
			}
		}

		if m.writer != nil {
			m.writer.WriteString(fmt.Sprintf(
				"inserted %s rows to %s",
				m.writer.Color.Green(strconv.Itoa(len(collection.Rows))),
				m.writer.Color.Cyan(collection.Table),
			))
		}
	}

	return nil
}

func (m *SeedComponent) clientProvider() (*sql.DB, error) {
	return m.config.ClientProvider()
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

func generateInsertSQL(table string, data any) (string, []any) {
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)
	var columns []string
	var placeholders []string
	var values []any
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		column := field.Tag.Get("column")
		if column != "" {
			columns = append(columns, column)
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
			values = append(values, v.Field(i).Interface())
		}
	}
	columnsPart := strings.Join(columns, ", ")
	placeholdersPart := strings.Join(placeholders, ", ")
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columnsPart, placeholdersPart)
	return sql, values
}
