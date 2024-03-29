// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package mongo

import "go.mongodb.org/mongo-driver/mongo"

// SeedConfig represents the configuration for the MongoDB seed component.
type SeedConfig struct {
	// URI - a valid MongoDB URI to connect to
	URI string `json:"uri,omitempty"`

	// ClientProvider - can be used as an alternative to URI, provides a mongo client to use.
	// available only via code, not available in config files.
	// if both ClientProvider and URI are provided, ClientProvider is used.
	ClientProvider func() (*mongo.Client, error) `json:"-"`

	// Data - a list of objects, each represents a single mongo collection and its data
	Data []*SeedCollectionData `json:"data,omitempty"`
}

// SeedCollectionData represents data for a MongoDB collection.
type SeedCollectionData struct {
	// DB - the name of the target mongo DB
	DB string `json:"db,omitempty"`

	// Collection - the name of the target mongo collection
	Collection string `json:"collection,omitempty"`

	// Documents - a list of documents to insert using the mongo InsertMany function:
	// https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#Collection.InsertMany
	Documents []any `json:"documents,omitempty"`
}
