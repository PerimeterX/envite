// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/perimeterx/envite"
	"github.com/perimeterx/envite/seed/redis"
)

// buildRedisSeed is a builder function that constructs a new Redis seed component.
// It takes a byte slice of JSON data as input.
// The function attempts to parse the JSON data into a mongo.SeedConfig struct, which defines the configuration
// for a Redis seed component. If the JSON data is successfully parsed, it then uses this configuration
// to instantiate and return a new MongoDB seed component via the redis.NewSeedComponent function.
//
// Returns:
// - An envite.Component which is the redis.SeedComponent initialized with the provided configuration.
// - An error if the JSON data cannot be parsed into a redis.SeedConfig struct.
func buildRedisSeed(data []byte, _ flagValues, _ string) (envite.Component, error) {
	var config redis.SeedConfig
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not parse config: %w", err)
	}

	return redis.NewSeedComponent(config), nil
}
