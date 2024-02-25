// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package docker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateWaiter(t *testing.T) {
	// Test case for WaiterTypeString
	waiterString := WaitForLog("test")
	funcString, err := validateWaiter(waiterString)
	assert.NoError(t, err)
	assert.NotNil(t, funcString)

	// Test case for WaiterTypeRegex
	waiterRegex := WaitForLogRegex("test.*")
	funcRegex, err := validateWaiter(waiterRegex)
	assert.NoError(t, err)
	assert.NotNil(t, funcRegex)

	// Test case for WaiterTypeDuration
	waiterDuration := WaitForDuration("1s")
	funcDuration, err := validateWaiter(waiterDuration)
	assert.NoError(t, err)
	assert.NotNil(t, funcDuration)

	// Test case for an invalid waiter type
	waiterInvalid := Waiter{
		Type: "invalid",
	}
	funcInvalid, err := validateWaiter(waiterInvalid)
	assert.Error(t, err)
	assert.Nil(t, funcInvalid)
}
