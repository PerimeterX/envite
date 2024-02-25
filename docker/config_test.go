// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package docker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainerConfig(t *testing.T) {
	newSampleConfig := func() Config {
		return Config{
			Name:  "test-container",
			Image: "nginx:latest",
		}
	}

	network := &Network{
		KeepStoppedContainers: true,
	}

	imageCloneTag := "nginx:cloned"

	// Test when valid config
	config := newSampleConfig()
	runConfig, err := config.initialize(network, imageCloneTag)

	assert.NoError(t, err)
	assert.NotNil(t, runConfig)

	// Test when Name is empty
	config = newSampleConfig()
	config.Name = ""
	_, err = config.initialize(network, imageCloneTag)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid docker config - property name: cannot be empty")

	// Test when Image is empty
	config = newSampleConfig()
	config.Image = ""
	_, err = config.initialize(network, imageCloneTag)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid docker config - property image: cannot be empty")

	// Test when ConsoleSize has more than 2 elements
	config = newSampleConfig()
	config.ConsoleSize = []uint{1, 2, 3}
	_, err = config.initialize(network, imageCloneTag)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid docker config - property console_size: must have exactly two elements")

	// Test when Waiters contain an invalid waiter type
	config = newSampleConfig()
	config.Waiters = []Waiter{
		{
			Type: "invalid",
		},
	}
	_, err = config.initialize(network, imageCloneTag)
	assert.Error(t, err)
	assert.EqualError(t, err, "invalid waiter type invalid")
}
