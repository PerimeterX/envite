// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redis

import (
	"github.com/go-redis/redis/v8"
	"time"
)

// SeedConfig represents the configuration for the redis seed component.
type SeedConfig struct {
	// Address - a valid redis server address to connect to
	Address string `json:"address,omitempty"`

	// ClientProvider - can be used as an alternative to Address, provides a redis client to use.
	// available only via code, not available in config files.
	// if both ClientProvider and Address are provided, ClientProvider is used.
	ClientProvider func() (*redis.Client, error) `json:"-"`

	// Entries - a list of entries to set in redis
	Entries *Entries `json:"entries"`
}

// Entries contains a list of HSet and Set entries to set in redis.
type Entries struct {
	HSet []*HSet `json:"hset,omitempty"`
	Set  []*Set  `json:"set,omitempty"`
}

// Set Represents a key-value pair to set in redis.
type Set struct {
	Key   string        `json:"key,omitempty"`
	Value string        `json:"value"`
	TTL   time.Duration `json:"ttl"`
}

// HSet Represents a key-value pair to set in a redis hash.
type HSet struct {
	Key    string            `json:"key"`
	Values map[string]string `json:"values"`
	TTL    time.Duration     `json:"ttl"`
}
