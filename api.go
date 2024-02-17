// Copyright 2024 HUMAN. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	contentType                    = "Content-Type"
	applicationJSON                = "application/json"
	accessControl                  = "Access-Control-Allow-Origin"
	accessControlValue             = "*"
	accessControlAllowHeaders      = "Access-Control-Allow-Headers"
	accessControlAllowHeadersValue = "Content-Type, Origin, Accept, token"
	accessControlAllowMethods      = "Access-Control-Allow-Methods"
	accessControlAllowMethodsValue = "GET,POST,PUT,DELETE,OPTIONS"
	invalidContentType             = "invalid content type"
	failedToReadBody               = "failed to read body"
)

// registerRoutes sets up the API endpoints using the provided router and environment.
// It defines routes for api to manage all components, and a fallback route to serve the UI.
func registerRoutes(router *mux.Router, env *Environment) {
	apiRoute(router, http.MethodGet, "/status", getStatusHandler{env: env})
	apiRoute(router, http.MethodPost, "/start_component", postStartHandler{env: env})
	apiRoute(router, http.MethodPost, "/stop_component", postStopHandler{env: env})
	apiRoute(router, http.MethodPost, "/apply", postApplyHandler{env: env})
	apiRoute(router, http.MethodPost, "/stop_all", postStopAllHandler{env: env})
	apiRoute(router, http.MethodGet, "/output", getOutputHandler{env: env})
	router.PathPrefix("/").Handler(newWebHandler())
}

// getStatusHandler handles requests to retrieve the current status of the environment or components within it.
type getStatusHandler struct {
	env *Environment
}

// GetStatusResponse defines the structure of the response for a status request.
// It includes details such as component ID, type, status, additional information, and environment variables.
type GetStatusResponse struct {
	ID         string                         `json:"id"`
	Components [][]GetStatusResponseComponent `json:"components"`
}

// GetStatusResponseComponent represents a single component's status within the environment.
// It provides detailed information about the component, including its ID, type, current status,
// additional information specific to the component type, and environment variables associated with it.
//
// Fields:
// - ID: A unique identifier for the component.
// - Type: The type of the component, indicating its role or function within the environment.
// - Status: The current status of the component, such as running, stopped, etc.
// - Config: The component config.
type GetStatusResponseComponent struct {
	ID     string          `json:"id"`
	Type   string          `json:"type"`
	Status ComponentStatus `json:"status"`
	Config map[string]any  `json:"config"`
}

// buildComponentInfo takes a Component and extracts its configuration object,
// injecting a field representing the component type.
func buildComponentInfo(c Component) (map[string]any, error) {
	data, err := json.Marshal(c.Config())
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	result["type"] = c.Type()
	return result, nil
}

func (g getStatusHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	status, err := g.env.Status(request.Context())
	if err != nil {
		apiError(g.env, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(g.env, writer, status, http.StatusOK)
}

// postApplyHandler handles requests to apply a given configuration to the environment.
type postApplyHandler struct {
	env *Environment
}

// postApplyRequest defines the expected request body for applying new configurations.
type postApplyRequest struct {
	EnabledComponentIDs []string `json:"enabled_component_ids"`
}

func (p postApplyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postApplyRequest{}
	if !apiParse(p.env, writer, request, &body) {
		return
	}

	err := p.env.Apply(request.Context(), body.EnabledComponentIDs)
	if err != nil {
		apiError(p.env, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(p.env, writer, nil, http.StatusOK)
}

// postStopAllHandler handles requests to stop all components in the environment, optionally cleaning up resources.
type postStopAllHandler struct {
	env *Environment
}

// postStopAllRequest defines the expected request body for stopping all components.
type postStopAllRequest struct {
	Cleanup bool `json:"cleanup"`
}

func (p postStopAllHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postStopAllRequest{}
	if !apiParse(p.env, writer, request, &body) {
		return
	}

	err := p.env.StopAll(request.Context())
	if err != nil {
		apiError(p.env, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if body.Cleanup {
		err = p.env.Cleanup(request.Context())
		if err != nil {
			apiError(p.env, writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	apiSuccess(p.env, writer, nil, http.StatusOK)
}

// postStartHandler handles requests to start a specific component within the environment.
type postStartHandler struct {
	env *Environment
}

// postStartRequest defines the expected request body for starting a component.
type postStartRequest struct {
	ComponentID string `json:"component_id"`
}

func (p postStartHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postStartRequest{}
	if !apiParse(p.env, writer, request, &body) {
		return
	}

	err := p.env.StartComponent(request.Context(), body.ComponentID)
	if err != nil {
		apiError(p.env, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(p.env, writer, nil, http.StatusOK)
}

// postStopHandler handles requests to stop a specific component within the environment.
type postStopHandler struct {
	env *Environment
}

// postStopRequest defines the expected request body for stopping a component.
type postStopRequest struct {
	ComponentID string `json:"component_id"`
}

func (p postStopHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postStopRequest{}
	if !apiParse(p.env, writer, request, &body) {
		return
	}

	err := p.env.StopComponent(request.Context(), body.ComponentID)
	if err != nil {
		apiError(p.env, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(p.env, writer, nil, http.StatusOK)
}

// getOutputHandler handles requests to stream the output from the environment or components.
type getOutputHandler struct {
	env *Environment
}

// ServeHTTP implements the http.Handler interface for getOutputHandler, streaming output to the client.
func (g getOutputHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set(accessControl, accessControlValue)
	reader := g.env.Output()

	ch := reader.Chan()

	for {
		select {
		case data := <-ch:
			_, err := writer.Write(data)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					g.env.Logger(LogLevelError, fmt.Sprintf("could not write output stream response: %v", err))
				}
				continue
			}
			if f, ok := writer.(http.Flusher); ok {
				f.Flush()
			}
		case <-request.Context().Done():
			err := reader.Close()
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					g.env.Logger(LogLevelError, fmt.Sprintf("could not close output reader: %v", err))
				}
			}
			return
		}
	}
}

// apiParse is a helper function to parse the JSON body of a request into a target struct.
// It returns true if parsing is successful, false otherwise.
func apiParse(b *Environment, writer http.ResponseWriter, request *http.Request, target any) bool {
	if strings.ToLower(request.Header.Get(contentType)) != applicationJSON {
		apiError(b, writer, invalidContentType, http.StatusBadRequest)
		return false
	}

	defer func() {
		_ = request.Body.Close()
	}()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		apiError(b, writer, failedToReadBody, http.StatusInternalServerError)
		return false
	}

	err = json.Unmarshal(body, target)
	if err != nil {
		apiError(b, writer, err.Error(), http.StatusInternalServerError)
		return false
	}

	return true
}

// apiErrorResponse represents the json response body returned in case of an error
type apiErrorResponse struct {
	Error string `json:"error"`
}

// apiError is a helper function to send an error response with a specific HTTP status code.
func apiError(b *Environment, writer http.ResponseWriter, error string, status int) {
	if status >= 500 && !strings.Contains(error, "context canceled") {
		b.Logger(LogLevelError, fmt.Sprintf("failed to serve request with status %d: %s", status, error))
	}
	writer.Header().Set(accessControl, accessControlValue)
	writer.Header().Set(contentType, applicationJSON)
	writer.WriteHeader(status)

	response := apiErrorResponse{Error: error}
	data, err := json.Marshal(response)
	if err != nil {
		b.Logger(LogLevelError, fmt.Sprintf("could not marshal fail response: %v", err))
		return
	}

	_, err = writer.Write(data)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			b.Logger(LogLevelError, fmt.Sprintf("could not write fail response: %v", err))
		}
	}
}

// apiSuccess is a helper function to send a success response, marshalling the provided body to JSON.
func apiSuccess(b *Environment, writer http.ResponseWriter, body any, status int) {
	var data []byte
	if body == nil {
		data = []byte(`{}`)
	} else {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			apiError(
				b,
				writer,
				fmt.Sprintf("could not marshal response body: %s", err.Error()),
				http.StatusInternalServerError,
			)
			return
		}
	}

	writer.Header().Set(accessControl, accessControlValue)
	writer.Header().Set(contentType, applicationJSON)
	writer.WriteHeader(status)
	_, err := writer.Write(data)
	if err != nil {
		b.Logger(LogLevelError, fmt.Sprintf("could not write successful response: %v", err))
	}
}

// apiRoute is a helper function to simplify API route registration, including setting up CORS options.
func apiRoute(router *mux.Router, method, path string, handler http.Handler) {
	router.Methods(method).Path(path).Handler(handler)
	router.Methods(http.MethodOptions).Path(path).HandlerFunc(optionsHandler)
}

// optionsHandler is a pre-configured HTTP handler for responding to CORS preflight requests.
func optionsHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set(accessControl, accessControlValue)
	writer.Header().Set(accessControlAllowHeaders, accessControlAllowHeadersValue)
	writer.Header().Set(accessControlAllowMethods, accessControlAllowMethodsValue)
	writer.WriteHeader(http.StatusOK)
}

const indexFilePath = "index.html"

// webHandler serves static files from the bundled assets, defaulting to index.html for unrecognized paths.
type webHandler struct {
	fileServer http.Handler
}

// newWebHandler creates a new instance of webHandler,
// initializing it with a file server that serves the bundled assets.
func newWebHandler() *webHandler {
	return &webHandler{
		fileServer: http.FileServer(AssetFile()),
	}
}

// ServeHTTP implements the http.Handler interface for webHandler, serving static files or defaulting to index.html.
func (h webHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	_, err := Asset(path)
	if err != nil {
		data, _ := Asset(indexFilePath)
		http.ServeContent(w, r, indexFilePath, time.Time{}, bytes.NewReader(data))
	} else {
		h.fileServer.ServeHTTP(w, r)
	}
}
