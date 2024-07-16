package postgres

import "database/sql"

// SeedConfig represents the configuration for the Postgres seed component.
type SeedConfig struct {
	// ClientProvider - Provides a postgres client to use.
	// available only via code, not available in config files.
	ClientProvider func() (*sql.DB, error) `json:"-"`

	// Setup - a string that contains the SQL setup script to run before seeding the data.
	Setup string `json:"-"`

	// Data - a list of objects, each represents a single postgres table and its data
	Data []*SeedTableData `json:"data,omitempty"`
}

// SeedTableData represents data for a Postgres table.
type SeedTableData struct {
	// TableName - the name of the target postgres table
	TableName string `json:"table,omitempty"`

	// Rows - a list of rows to insert using the postgres Exec function (a `column` tag is required for each field):
	Rows []any `json:"rows,omitempty"`
}
