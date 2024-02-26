// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type mockComponent struct {
	status        ComponentStatus
	shouldFail    bool
	prepareCalled bool
	startCalled   bool
	stopCalled    bool
	cleanupCalled bool
	onStart       func()
	onStop        func()
	w             *Writer
}

func (m *mockComponent) Type() string {
	return "mock"
}

func (m *mockComponent) AttachEnvironment(_ context.Context, _ *Environment, w *Writer) error {
	m.w = w
	return nil
}

func (m *mockComponent) Prepare(context.Context) error {
	m.prepareCalled = true
	return nil
}

func (m *mockComponent) Start(context.Context) error {
	if m.shouldFail {
		return errors.New("start error")
	}
	m.startCalled = true
	m.status = ComponentStatusRunning
	if m.onStart != nil {
		m.onStart()
	}
	return nil
}

func (m *mockComponent) Stop(context.Context) error {
	if m.shouldFail {
		return errors.New("stop error")
	}
	m.stopCalled = true
	m.status = ComponentStatusStopped
	if m.onStop != nil {
		m.onStop()
	}
	return nil
}

func (m *mockComponent) Cleanup(context.Context) error {
	m.cleanupCalled = true
	return nil
}

func (m *mockComponent) Status(context.Context) (ComponentStatus, error) {
	return m.status, nil
}

func (m *mockComponent) Config() any {
	return nil
}

func (m *mockComponent) EnvVars() map[string]string {
	return nil
}

func (m *mockComponent) initFlags() {
	m.prepareCalled = false
	m.startCalled = false
	m.stopCalled = false
	m.cleanupCalled = false
}

func TestEnvironmentManagement(t *testing.T) {
	// Create mock components
	component1 := &mockComponent{}
	component2 := &mockComponent{}
	component3 := &mockComponent{}

	component1.onStart = func() {
		assert.True(t, component1.prepareCalled)
		assert.True(t, component2.prepareCalled)
		assert.True(t, component3.prepareCalled)
		assert.True(t, component1.startCalled)
		assert.False(t, component2.startCalled)
		assert.False(t, component3.startCalled)
		assert.False(t, component1.stopCalled)
		assert.False(t, component2.stopCalled)
		assert.False(t, component3.stopCalled)
		assert.False(t, component1.cleanupCalled)
		assert.False(t, component2.cleanupCalled)
		assert.False(t, component3.cleanupCalled)
	}
	component2.onStart = func() {
		assert.True(t, component1.prepareCalled)
		assert.True(t, component2.prepareCalled)
		assert.True(t, component3.prepareCalled)
		assert.True(t, component1.startCalled)
		assert.True(t, component2.startCalled)
		assert.False(t, component3.startCalled)
		assert.False(t, component1.stopCalled)
		assert.False(t, component2.stopCalled)
		assert.False(t, component3.stopCalled)
		assert.False(t, component1.cleanupCalled)
		assert.False(t, component2.cleanupCalled)
		assert.False(t, component3.cleanupCalled)
	}
	component3.onStart = func() {
		assert.True(t, component1.prepareCalled)
		assert.True(t, component2.prepareCalled)
		assert.True(t, component3.prepareCalled)
		assert.True(t, component1.startCalled)
		assert.True(t, component2.startCalled)
		assert.True(t, component3.startCalled)
		assert.False(t, component1.stopCalled)
		assert.False(t, component2.stopCalled)
		assert.False(t, component3.stopCalled)
		assert.False(t, component1.cleanupCalled)
		assert.False(t, component2.cleanupCalled)
		assert.False(t, component3.cleanupCalled)
	}
	component1.onStop = func() {
		assert.True(t, component1.prepareCalled)
		assert.True(t, component2.prepareCalled)
		assert.True(t, component3.prepareCalled)
		assert.True(t, component1.startCalled)
		assert.True(t, component2.startCalled)
		assert.True(t, component3.startCalled)
		assert.True(t, component1.stopCalled)
		assert.True(t, component2.stopCalled)
		assert.True(t, component3.stopCalled)
		assert.False(t, component1.cleanupCalled)
		assert.False(t, component2.cleanupCalled)
		assert.False(t, component3.cleanupCalled)
	}
	component2.onStop = func() {
		assert.True(t, component1.prepareCalled)
		assert.True(t, component2.prepareCalled)
		assert.True(t, component3.prepareCalled)
		assert.True(t, component1.startCalled)
		assert.True(t, component2.startCalled)
		assert.True(t, component3.startCalled)
		assert.False(t, component1.stopCalled)
		assert.True(t, component2.stopCalled)
		assert.True(t, component3.stopCalled)
		assert.False(t, component1.cleanupCalled)
		assert.False(t, component2.cleanupCalled)
		assert.False(t, component3.cleanupCalled)
	}
	component3.onStop = func() {
		assert.True(t, component1.prepareCalled)
		assert.True(t, component2.prepareCalled)
		assert.True(t, component3.prepareCalled)
		assert.True(t, component1.startCalled)
		assert.True(t, component2.startCalled)
		assert.True(t, component3.startCalled)
		assert.False(t, component1.stopCalled)
		assert.False(t, component2.stopCalled)
		assert.True(t, component3.stopCalled)
		assert.False(t, component1.cleanupCalled)
		assert.False(t, component2.cleanupCalled)
		assert.False(t, component3.cleanupCalled)
	}

	// Create a mock environment
	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().
			AddLayer(map[string]Component{"component-1": component1}).
			AddLayer(map[string]Component{"component-2": component2}).
			AddLayer(map[string]Component{"component-3": component3}),
	)
	assert.NoError(t, err)

	// Validate env components
	components := env.Components()
	assert.Len(t, components, 3)
	assert.Contains(t, components, component1)
	assert.Contains(t, components, component2)
	assert.Contains(t, components, component3)
	c, err := env.componentByID("component-1")
	assert.NoError(t, err)
	assert.Equal(t, component1, c)
	c, err = env.componentByID("not exist")
	assert.Nil(t, c)
	assert.NotNil(t, err)

	// Start
	err = env.Apply(context.Background(), []string{"component-1", "component-2", "component-3"})
	assert.NoError(t, err)

	// Validate status
	status, err := env.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-env", status.ID)
	assert.Len(t, status.Components, 3)
	assert.Len(t, status.Components[0], 1)
	assert.Equal(t, "component-1", status.Components[0][0].ID)
	assert.Equal(t, "mock", status.Components[0][0].Type)
	assert.Equal(t, ComponentStatusRunning, status.Components[0][0].Status)
	assert.Equal(t, "component-2", status.Components[1][0].ID)
	assert.Equal(t, "mock", status.Components[1][0].Type)
	assert.Equal(t, ComponentStatusRunning, status.Components[1][0].Status)
	assert.Equal(t, "component-3", status.Components[2][0].ID)
	assert.Equal(t, "mock", status.Components[2][0].Type)
	assert.Equal(t, ComponentStatusRunning, status.Components[2][0].Status)

	// Stop and cleanup
	err = env.StopAll(context.Background())
	assert.NoError(t, err)
	err = env.Cleanup(context.Background())
	assert.NoError(t, err)

	// Validate cleanup
	assert.True(t, component1.prepareCalled)
	assert.True(t, component2.prepareCalled)
	assert.True(t, component3.prepareCalled)
	assert.True(t, component1.startCalled)
	assert.True(t, component2.startCalled)
	assert.True(t, component3.startCalled)
	assert.True(t, component1.stopCalled)
	assert.True(t, component2.stopCalled)
	assert.True(t, component3.stopCalled)
	assert.True(t, component1.cleanupCalled)
	assert.True(t, component2.cleanupCalled)
	assert.True(t, component3.cleanupCalled)

	// Validate status
	status, err = env.Status(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "test-env", status.ID)
	assert.Len(t, status.Components, 3)
	assert.Len(t, status.Components[0], 1)
	assert.Equal(t, "component-1", status.Components[0][0].ID)
	assert.Equal(t, "mock", status.Components[0][0].Type)
	assert.Equal(t, ComponentStatusStopped, status.Components[0][0].Status)
	assert.Equal(t, "component-2", status.Components[1][0].ID)
	assert.Equal(t, "mock", status.Components[1][0].Type)
	assert.Equal(t, ComponentStatusStopped, status.Components[1][0].Status)
	assert.Equal(t, "component-3", status.Components[2][0].ID)
	assert.Equal(t, "mock", status.Components[2][0].Type)
	assert.Equal(t, ComponentStatusStopped, status.Components[2][0].Status)

	// Clear state
	component1.initFlags()
	component2.initFlags()
	component3.initFlags()

	// Validate full start
	err = env.StartAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, ComponentStatusRunning, component1.status)
	assert.Equal(t, ComponentStatusRunning, component2.status)
	assert.Equal(t, ComponentStatusRunning, component3.status)

	component1.onStart = nil
	component2.onStart = nil
	component3.onStart = nil
	component1.onStop = nil
	component2.onStop = nil
	component3.onStop = nil

	// Validate manual stop
	err = env.StopComponent(context.Background(), "component-1")
	assert.NoError(t, err)
	assert.Equal(t, ComponentStatusStopped, component1.status)
	assert.Equal(t, ComponentStatusRunning, component2.status)
	assert.Equal(t, ComponentStatusRunning, component3.status)

	// Validate manual start
	err = env.StartComponent(context.Background(), "component-1")
	assert.NoError(t, err)
	assert.Equal(t, ComponentStatusRunning, component1.status)
	assert.Equal(t, ComponentStatusRunning, component2.status)
	assert.Equal(t, ComponentStatusRunning, component3.status)
}

func TestSelectiveAndZeroComponentApplication(t *testing.T) {
	// Setup
	component1 := &mockComponent{}
	component2 := &mockComponent{}
	component3 := &mockComponent{}

	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().
			AddLayer(map[string]Component{"component-1": component1}).
			AddLayer(map[string]Component{"component-2": component2}).
			AddLayer(map[string]Component{"component-3": component3}),
	)
	assert.NoError(t, err)

	// Start a subset of components
	err = env.Apply(context.Background(), []string{"component-1", "component-2"})
	assert.NoError(t, err)

	// Assert only the specified components are started
	assert.True(t, component1.startCalled, "Component 1 should be started")
	assert.True(t, component2.startCalled, "Component 2 should be started")
	assert.False(t, component3.startCalled, "Component 3 should not be started")

	// Init flags
	component1.initFlags()
	component2.initFlags()
	component3.initFlags()

	// Now apply zero components and check if all are stopped
	err = env.Apply(context.Background(), []string{})
	assert.NoError(t, err)

	// Validate that all components are stopped
	assert.True(t, component1.stopCalled, "Component 1 should be stopped after applying zero components")
	assert.True(t, component2.stopCalled, "Component 2 should be stopped after applying zero components")
	assert.False(t, component3.startCalled, "Component 3 should remain not started")
}

func TestErrorHandlingDuringComponentManagement(t *testing.T) {
	// Setup with multiple components, including one that will fail on start
	componentFailOnStart := &mockComponent{shouldFail: true}
	componentSuccess1 := &mockComponent{}
	componentSuccess2 := &mockComponent{}

	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().
			AddLayer(map[string]Component{"success-1": componentSuccess1}).
			// Ensure this component is in the middle to test error handling
			AddLayer(map[string]Component{"fail-start": componentFailOnStart}).
			AddLayer(map[string]Component{"success-2": componentSuccess2}),
	)
	assert.NoError(t, err)

	// Test applying all components, expecting failure due to "fail-start" component
	err = env.Apply(context.Background(), []string{"success-1", "fail-start", "success-2"})
	assert.Error(t, err, "Apply should fail due to the 'fail-start' component")
	assert.Contains(t, err.Error(), "start error", "Error should propagate when component fails to start")

	// Ensure all prepare calls occurred
	assert.True(t, componentSuccess1.prepareCalled, "Component 'success-1' should be prepared")
	assert.True(t, componentFailOnStart.prepareCalled, "Component 'fail-start' should be prepared")
	assert.True(t, componentSuccess2.prepareCalled, "Component 'success-2' should be prepared")

	// Ensure that the failure does not affect the initiation of other components
	assert.True(t, componentSuccess1.startCalled, "Component 'success-1' should be started before failure occurs")
	assert.False(t, componentSuccess2.startCalled, "Component 'success-2' should not be started due to the failure in 'fail-start'")

	// Test applying only the components that do not fail
	err = env.Apply(context.Background(), []string{"success-1", "success-2"})
	assert.Error(t, err, "Apply should fail due to the 'fail-start' component")
	assert.Contains(t, err.Error(), "stop error", "Error should propagate when component fails to start")
}

func TestErrorHandlingDuringConstruction(t *testing.T) {
	// Setup with empty ID
	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().
			AddLayer(map[string]Component{"valid": &mockComponent{}}).
			AddLayer(map[string]Component{"": &mockComponent{}}),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component id")
	assert.Nil(t, env)

	// Setup with duplicate ID
	env, err = NewEnvironment(
		"test-env",
		NewComponentGraph().
			AddLayer(map[string]Component{"valid": &mockComponent{}}).
			AddLayer(map[string]Component{"valid": &mockComponent{}}),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "component id")
	assert.Nil(t, env)

	// Setup with empty env id
	env, err = NewEnvironment(
		"",
		NewComponentGraph().AddLayer(map[string]Component{"valid": &mockComponent{}}),
	)
	assert.Equal(t, ErrEmptyEnvID, err)
	assert.Nil(t, env)

	// Setup with nil graph
	env, err = NewEnvironment(
		"test-env",
		nil,
	)
	assert.Equal(t, ErrNilGraph, err)
	assert.Nil(t, env)
}

func TestLogger(t *testing.T) {
	logMessages := 0
	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().AddLayer(map[string]Component{"valid": &mockComponent{}}),
		WithLogger(func(level LogLevel, message string) {
			logMessages++
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, env)

	// validate log messages
	err = env.StartAll(context.Background())
	assert.NoError(t, err)

	assert.True(t, logMessages >= 2)
}

func TestOutput(t *testing.T) {
	component := &mockComponent{}

	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().AddLayer(map[string]Component{"component": component}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, env)
	someTime := time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC)

	// send messages
	component.w.Write([]byte("bytes message\n"))
	component.w.WriteString("string message\r\n")
	component.w.WriteWithTime(someTime, []byte("bytes message with time"))
	component.w.WriteStringWithTime(someTime, "string message with time")
	component.w.WriteString(component.w.Color.Red("red message"))
	component.w.WriteString(component.w.Color.Green("green message"))
	component.w.WriteString(component.w.Color.Yellow("yellow message"))
	component.w.WriteString(component.w.Color.Blue("blue message"))
	component.w.WriteString(component.w.Color.Magenta("magenta message"))
	component.w.WriteString(component.w.Color.Cyan("cyan message"))

	// validate messages
	thisYear := fmt.Sprintf("<component>component<time>%d", time.Now().Year())
	someYear := fmt.Sprintf("<component>component<time>%d", someTime.Year())
	reader := env.Output()
	ch := reader.Chan()
	msg := <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "bytes message")

	msg = <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "string message")

	msg = <-ch
	assert.Contains(t, string(msg), someYear)
	assert.Contains(t, string(msg), "bytes message with time")

	msg = <-ch
	assert.Contains(t, string(msg), someYear)
	assert.Contains(t, string(msg), "string message with time")

	msg = <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "\u001B[31mred message\u001B[39m")

	msg = <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "\u001B[32mgreen message\u001B[39m")

	msg = <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "\u001B[33myellow message\u001B[39m")

	msg = <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "\u001B[34mblue message\u001B[39m")

	msg = <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "\u001B[35mmagenta message\u001B[39m")

	msg = <-ch
	assert.Contains(t, string(msg), thisYear)
	assert.Contains(t, string(msg), "\u001B[36mcyan message\u001B[39m")

	err = reader.Close()
	assert.NoError(t, err)
}
