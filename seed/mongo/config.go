// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package mongo

// SeedConfig represents the configuration for the MongoDB seed component.
type SeedConfig struct {
	// ID - a unique component name
	ID string `json:"id,omitempty" yaml:"id,omitempty"`

	// Data - a list of objects, each represents a single mongo collection and its data
	Data []*SeedCollectionData `json:"data,omitempty" yaml:"data,omitempty"`
}

// SeedCollectionData represents data for a MongoDB collection.
type SeedCollectionData struct {
	// DB - the name of the target mongo DB
	DB string `json:"db,omitempty" yaml:"db,omitempty"`

	// Collection - the name of the target mongo collection
	Collection string `json:"collection,omitempty" yaml:"collection,omitempty"`

	// Documents - a list of documents to insert using the mongo InsertMany function:
	// https://pkg.go.dev/go.mongodb.org/mongo-driver/mongo#Collection.InsertMany
	Documents []any `json:"documents,omitempty" yaml:"documents,omitempty"`
}
