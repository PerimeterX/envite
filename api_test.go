package envite

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPI(t *testing.T) {
	component := &mockComponent{}
	env, err := NewEnvironment(
		"test-env",
		NewComponentGraph().AddLayer(map[string]Component{"component": component}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, env)

	call := func(handler http.Handler, request, response any) int {
		var reqBody io.Reader
		if request != nil {
			var data []byte
			data, err = json.Marshal(request)
			assert.NoError(t, err)
			reqBody = bytes.NewBuffer(data)
		}
		req := httptest.NewRequest(http.MethodGet, "/foo", reqBody)
		if request != nil {
			req.Header.Set(contentType, applicationJSON)
		}
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		if response != nil {
			var data []byte
			data, err = io.ReadAll(res.Body)
			assert.NoError(t, err)
			err = json.Unmarshal(data, response)
			assert.NoError(t, err)
		}
		return res.Code
	}

	getStatusResponse := GetStatusResponse{}
	status := call(getStatusHandler{env: env}, nil, &getStatusResponse)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "test-env", getStatusResponse.ID)
	assert.Len(t, getStatusResponse.Components, 1)
	assert.Len(t, getStatusResponse.Components[0], 1)
	assert.Equal(t, "component", getStatusResponse.Components[0][0].ID)
	assert.Equal(t, "mock", getStatusResponse.Components[0][0].Type)

	status = call(postStartHandler{env: env}, postStartRequest{ComponentID: "invalid"}, nil)
	assert.Equal(t, http.StatusInternalServerError, status)
	status = call(postStartHandler{env: env}, postStartRequest{ComponentID: "component"}, nil)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, ComponentStatusRunning, component.status)

	status = call(postStopHandler{env: env}, postStopRequest{ComponentID: "invalid"}, nil)
	assert.Equal(t, http.StatusInternalServerError, status)
	status = call(postStopHandler{env: env}, postStopRequest{ComponentID: "component"}, nil)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, ComponentStatusStopped, component.status)

	status = call(postApplyHandler{env: env}, postApplyRequest{EnabledComponentIDs: []string{"component"}}, nil)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, ComponentStatusRunning, component.status)
	status = call(postApplyHandler{env: env}, postApplyRequest{EnabledComponentIDs: []string{}}, nil)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, ComponentStatusStopped, component.status)

	err = env.StartAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, ComponentStatusRunning, component.status)
	assert.False(t, component.cleanupCalled)
	status = call(postStopAllHandler{env: env}, postStopAllRequest{Cleanup: true}, nil)
	assert.Equal(t, http.StatusOK, status)
	assert.True(t, component.cleanupCalled)

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	res := httptest.NewRecorder()
	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)
	startTime := time.Now()
	go func() {
		// wait one second and close client connection
		time.Sleep(time.Second)
		cancel()
	}()
	outputHandler := getOutputHandler{env: env}
	outputHandler.ServeHTTP(res, req)
	assert.True(t, time.Since(startTime) >= time.Second)

	wHandler := newWebHandler()
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	wHandler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	req = httptest.NewRequest(http.MethodGet, "/favicon.ico", nil)
	res = httptest.NewRecorder()
	wHandler.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	req = httptest.NewRequest(http.MethodGet, "/", nil)
	res = httptest.NewRecorder()
	optionsHandler(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}
