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
	Data []*SeedCollectionData `json:"data,omitempty"`
}

// SeedCollectionData represents data for a Postgres table.
type SeedCollectionData struct {
	// Table - the name of the target postgres table
	Table string `json:"collection,omitempty"`

	// Rows - a list of rows to insert using the postgres Exec function (a `column` tag is required for each field):
	Rows []any `json:"documents,omitempty"`
}
