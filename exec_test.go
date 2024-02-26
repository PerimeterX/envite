package envite

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseExecutionMode(t *testing.T) {
	mode, err := ParseExecutionMode("start")
	assert.NoError(t, err)
	assert.Equal(t, ExecutionModeStart, mode)

	mode, err = ParseExecutionMode("stop")
	assert.NoError(t, err)
	assert.Equal(t, ExecutionModeStop, mode)

	mode, err = ParseExecutionMode("daemon")
	assert.NoError(t, err)
	assert.Equal(t, ExecutionModeDaemon, mode)

	mode, err = ParseExecutionMode("")
	assert.NoError(t, err)
	assert.Equal(t, ExecutionModeDaemon, mode)

	mode, err = ParseExecutionMode("invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid execution mode")
}

func TestExecute(t *testing.T) {
	component := &mockComponent{}
	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().AddLayer(map[string]Component{"component": component}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, env)

	server := NewServer("8080", env)

	err = Execute(server, ExecutionModeStart)
	assert.NoError(t, err)
	assert.Equal(t, ComponentStatusRunning, component.status)
	assert.True(t, component.prepareCalled)
	assert.True(t, component.startCalled)
	assert.False(t, component.stopCalled)
	assert.False(t, component.cleanupCalled)

	err = Execute(server, ExecutionModeStop)
	assert.NoError(t, err)
	assert.Equal(t, ComponentStatusStopped, component.status)
	assert.True(t, component.prepareCalled)
	assert.True(t, component.startCalled)
	assert.True(t, component.stopCalled)
	assert.True(t, component.cleanupCalled)
}

func TestDescribeAvailableModes(t *testing.T) {
	result := DescribeAvailableModes()
	assert.Contains(t, result, ExecutionModeStart, "missing start mode")
	assert.Contains(t, result, ExecutionModeStop, "missing stop mode")
	assert.Contains(t, result, ExecutionModeDaemon, "missing daemon mode")
}
