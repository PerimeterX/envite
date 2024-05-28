// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/perimeterx/envite"
	"github.com/perimeterx/envite/docker"
	"github.com/perimeterx/envite/seed/mongo"
	"github.com/perimeterx/envite/seed/redis"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
)

// defaultFile is the default filename for the environment configuration,
// unless explicitly provided otherwise via CLI flags.
const defaultFile = "envite.yml"

// buildEnv constructs an envite.Environment instance from the provided flags.
// It reads and parses the configuration file, constructs a component graph, and initializes an Environment.
// Returns an initialized Environment or an error if any step fails.
func buildEnv(flags flagValues) (*envite.Environment, error) {
	file := defaultFile
	if flags.file.exist {
		file = flags.file.value
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", file, err)
	}

	var envConfig environmentConfig
	err = yaml.Unmarshal(data, &envConfig)
	if err != nil {
		return nil, fmt.Errorf("could not parse file %s: %w", file, err)
	}

	envID := envConfig.DefaultID
	if flags.envID.exist {
		envID = flags.envID.value
	}

	graph, err := buildComponentGraph(flags, envConfig, envID)
	if err != nil {
		return nil, fmt.Errorf("could not build component graph: %w", err)
	}

	return envite.NewEnvironment(envID, graph, envite.WithLogger(logger))
}

// environmentConfig represents the structure of the environment configuration file.
type environmentConfig struct {
	DefaultID  string           `yaml:"default_id"`
	Components []map[string]any `yaml:"components"`
}

// builderFunc is a function type that constructs a Component from JSON data.
type builderFunc func(data []byte, flags flagValues, envID string) (envite.Component, error)

// mapping maps component types to their respective builder functions.
// This is the list of component types supported by the CLI,
// for each supported type, we need to map to a builderFunc.
//
// the CLI supports the following component types,
// each type supports additional config params as specified below:
// *type: "docker component", all config params are available in docker.Config - https://github.com/PerimeterX/envite/blob/b4e9f545226c990a1025b9ca198856faff8b5eed/docker/config.go#L23
// *type: "mongo seed", all config params are available in mongo.SeedConfig - https://github.com/PerimeterX/envite/blob/b4e9f545226c990a1025b9ca198856faff8b5eed/seed/mongo/config.go#L10
// *type: "redis seed", all config params are available in redis.SeedConfig
//
// a full YAML example can be found in the root README.md at
// https://github.com/PerimeterX/envite/blob/main/README.md#cli-usage
var mapping = map[string]builderFunc{
	docker.ComponentType: buildDocker,
	mongo.ComponentType:  buildMongoSeed,
	redis.ComponentType:  buildRedisSeed,
}

// buildComponent constructs a Component from raw YAML data.
// It marshals the YAML data into JSON, injects hostnames into the configuration, determines the component type,
// and uses the appropriate builder function from the mapping.
// Returns a constructed Component or an error if the process fails.
func buildComponent(
	rawValue any,
	flags flagValues,
	envID string,
	components map[string]envite.Component,
) (envite.Component, error) {
	data, err := json.Marshal(rawValue)
	if err != nil {
		return nil, fmt.Errorf("could not marshal yaml data: %w", err)
	}

	data, err = injectHostnames(data, components)
	if err != nil {
		return nil, fmt.Errorf("could not inject host names to config data: %w", err)
	}

	t, err := extractComponentType(err, data)
	if err != nil {
		return nil, fmt.Errorf("could not extract component type: %w", err)
	}

	f := mapping[t]
	if f == nil {
		return nil, ErrUnsupportedComponentType{Type: t}
	}

	return f(data, flags, envID)
}

// injectHostnames replaces placeholders in the configuration data with actual hostnames.
// It uses a regular expression to find placeholders and replaces them with host values
// obtained from previous components. Returns modified data or an error if a component is
// missing or not a Docker component.
func injectHostnames(data []byte, components map[string]envite.Component) ([]byte, error) {
	return injectValues(data, func(s string) (string, error) {
		component := components[s]
		if component == nil {
			return "", fmt.Errorf("could not find component %s in a previous layer", s)
		}

		dockerComponent, ok := component.(*docker.Component)
		if !ok {
			return "", fmt.Errorf("component %s is not a docker component", s)
		}

		return dockerComponent.Host(), nil
	})
}

// templateRegexp is the regular expression used to identify placeholders in the configuration data.
// e.g. it should identify {{ component_id }} in http://{{ component_id }}:8080
var templateRegexp = regexp.MustCompile(`{{\s*([^}\s]+)\s*}}`)

// injectValues replaces placeholders in the input data with values obtained using the valueMapper function.
// It processes each match found by the regular expression, obtaining replacement values from the valueMapper.
// Returns the modified data or an error if value mapping fails.
func injectValues(input []byte, valueMapper func(string) (string, error)) ([]byte, error) {
	var err error
	result := templateRegexp.ReplaceAllFunc(input, func(match []byte) []byte {
		variableName := templateRegexp.FindSubmatch(match)[1]
		value, e := valueMapper(string(variableName))
		if e != nil {
			errors.Join(err, e)
		}
		return []byte(value)
	})
	return result, err
}

// extractComponentType determines the type of component from its configuration data.
// It unmarshals the data into a struct to extract the "type" field, identifying the component type.
// Returns the component type as a string or an error if unmarshalling fails.
func extractComponentType(err error, data []byte) (string, error) {
	var t struct {
		Type string `json:"type"`
	}
	err = json.Unmarshal(data, &t)
	if err != nil {
		return "", err
	}

	return t.Type, nil
}

// buildComponentGraph constructs an envite.ComponentGraph from the environment configuration.
// It iterates through each component layer, constructing components and adding them to the graph.
// Components are built using the buildComponent function and are organized based on their dependencies.
// Returns a fully constructed ComponentGraph or an error if any component fails to build.
func buildComponentGraph(flags flagValues, envConfig environmentConfig, envID string) (*envite.ComponentGraph, error) {
	byID := make(map[string]envite.Component)
	graph := envite.NewComponentGraph()
	for _, layer := range envConfig.Components {
		components := make(map[string]envite.Component, len(layer))
		for id, rawValue := range layer {
			component, err := buildComponent(rawValue, flags, envID, byID)
			if err != nil {
				return nil, fmt.Errorf("could not build component %s: %w", id, err)
			}
			components[id] = component
			byID[id] = component
		}
		graph.AddLayer(components)
	}
	return graph, nil
}

// ErrUnsupportedComponentType represents an error for component types that are not supported.
type ErrUnsupportedComponentType struct {
	Type string
}

func (u ErrUnsupportedComponentType) Error() string {
	return fmt.Sprintf("unsupported component type %s", u.Type)
}
