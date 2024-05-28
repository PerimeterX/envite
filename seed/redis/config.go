// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redis

import "github.com/go-redis/redis/v8"

// SeedConfig represents the configuration for the redis seed component.
type SeedConfig struct {
	// Address - a valid redis server address to connect to
	Address string `json:"address,omitempty"`

	// ClientProvider - can be used as an alternative to Address, provides a redis client to use.
	// available only via code, not available in config files.
	// if both ClientProvider and Address are provided, ClientProvider is used.
	ClientProvider func() (*redis.Client, error) `json:"-"`

	// Data - a list of objects, each represents a redis key and its data
	Data []*SeedData `json:"data,omitempty"`
}

// SeedData represents data for a redis hash.
type SeedData struct {
	// Key - the name of the redis key
	Key string `json:"key,omitempty"`

	// Fields - a map of field names and their values to insert using the redis HSet function:
	Fields []string `json:"fields,omitempty"`

	// TTL - the time to live for the key in seconds
	TTL int `json:"ttl,omitempty"`
}
