package fengshui

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

func registerRoutes(router *mux.Router, blueprint *Blueprint) {
	apiRoute(router, http.MethodGet, "/status", getStatusHandler{blueprint: blueprint})
	apiRoute(router, http.MethodPost, "/start_component", postStartHandler{blueprint: blueprint})
	apiRoute(router, http.MethodPost, "/stop_component", postStopHandler{blueprint: blueprint})
	apiRoute(router, http.MethodPost, "/apply", postApplyHandler{blueprint: blueprint})
	apiRoute(router, http.MethodPost, "/stop_all", postStopAllHandler{blueprint: blueprint})
	apiRoute(router, http.MethodGet, "/output", getOutputHandler{blueprint: blueprint})
	router.PathPrefix("/").Handler(newWebHandler())
}

type getStatusHandler struct {
	blueprint *Blueprint
}

type GetStatusResponse struct {
	ID         string                         `json:"id"`
	Components [][]GetStatusResponseComponent `json:"components"`
}

type GetStatusResponseComponent struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Status  ComponentStatus   `json:"status"`
	Info    any               `json:"info"`
	EnvVars map[string]string `json:"env_vars"`
}

func (g getStatusHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	status, err := g.blueprint.Status(request.Context())
	if err != nil {
		apiError(g.blueprint, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(g.blueprint, writer, status, http.StatusOK)
}

type postApplyHandler struct {
	blueprint *Blueprint
}

type postApplyRequest struct {
	EnabledComponentIDs []string `json:"enabled_component_ids"`
}

var x = 0

func (p postApplyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postApplyRequest{}
	if !apiParse(p.blueprint, writer, request, &body) {
		return
	}

	err := p.blueprint.Apply(request.Context(), body.EnabledComponentIDs)
	if err != nil {
		apiError(p.blueprint, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(p.blueprint, writer, nil, http.StatusOK)
}

type postStopAllHandler struct {
	blueprint *Blueprint
}

type postStopAllRequest struct {
	Cleanup bool `json:"cleanup"`
}

func (p postStopAllHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postStopAllRequest{}
	if !apiParse(p.blueprint, writer, request, &body) {
		return
	}

	err := p.blueprint.StopAll(request.Context())
	if err != nil {
		apiError(p.blueprint, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	if body.Cleanup {
		err = p.blueprint.Cleanup(request.Context())
		if err != nil {
			apiError(p.blueprint, writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	apiSuccess(p.blueprint, writer, nil, http.StatusOK)
}

type postStartHandler struct {
	blueprint *Blueprint
}

type postStartRequest struct {
	ComponentID string `json:"component_id"`
}

func (p postStartHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postStartRequest{}
	if !apiParse(p.blueprint, writer, request, &body) {
		return
	}

	err := p.blueprint.StartComponent(request.Context(), body.ComponentID)
	if err != nil {
		apiError(p.blueprint, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(p.blueprint, writer, nil, http.StatusOK)
}

type postStopHandler struct {
	blueprint *Blueprint
}

type postStopRequest struct {
	ComponentID string `json:"component_id"`
}

func (p postStopHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	body := postStopRequest{}
	if !apiParse(p.blueprint, writer, request, &body) {
		return
	}

	err := p.blueprint.StopComponent(request.Context(), body.ComponentID)
	if err != nil {
		apiError(p.blueprint, writer, err.Error(), http.StatusInternalServerError)
		return
	}

	apiSuccess(p.blueprint, writer, nil, http.StatusOK)
}

type getOutputHandler struct {
	blueprint *Blueprint
}

func (g getOutputHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set(accessControl, accessControlValue)
	reader := g.blueprint.Output()

	ch := reader.Chan()

	for {
		select {
		case data := <-ch:
			_, err := writer.Write(data)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					g.blueprint.logger(LogLevelError, fmt.Sprintf("could not write output stream response: %v", err))
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
					g.blueprint.logger(LogLevelError, fmt.Sprintf("could not close output reader: %v", err))
				}
			}
			return
		}
	}
}

func apiParse(b *Blueprint, writer http.ResponseWriter, request *http.Request, target any) bool {
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

type apiErrorResponse struct {
	Error string `json:"error"`
}

func apiError(b *Blueprint, writer http.ResponseWriter, error string, status int) {
	if status >= 500 && !strings.Contains(error, "context canceled") {
		b.logger(LogLevelError, fmt.Sprintf("failed to serve request with status %d: %s", status, error))
	}
	writer.Header().Set(accessControl, accessControlValue)
	writer.Header().Set(contentType, applicationJSON)
	writer.WriteHeader(status)

	response := apiErrorResponse{Error: error}
	data, err := json.Marshal(response)
	if err != nil {
		b.logger(LogLevelError, fmt.Sprintf("could not marshal fail response: %v", err))
		return
	}

	_, err = writer.Write(data)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			b.logger(LogLevelError, fmt.Sprintf("could not write fail response: %v", err))
		}
	}
}

func apiSuccess(b *Blueprint, writer http.ResponseWriter, body any, status int) {
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
		b.logger(LogLevelError, fmt.Sprintf("could not write successful response: %v", err))
	}
}

func apiRoute(router *mux.Router, method, path string, handler http.Handler) {
	router.Methods(method).Path(path).Handler(handler)
	router.Methods(http.MethodOptions).Path(path).HandlerFunc(optionsHandler)
}

func optionsHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set(accessControl, accessControlValue)
	writer.Header().Set(accessControlAllowHeaders, accessControlAllowHeadersValue)
	writer.Header().Set(accessControlAllowMethods, accessControlAllowMethodsValue)
	writer.WriteHeader(http.StatusOK)
}

const indexFilePath = "index.html"

type webHandler struct {
	fileServer http.Handler
}

func newWebHandler() *webHandler {
	return &webHandler{
		fileServer: http.FileServer(AssetFile()),
	}
}

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
