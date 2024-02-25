// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/perimeterx/envite"
	"os"
)

// flagValues holds the command-line flags passed to the program.
// It includes various configuration options such as execution mode, file paths, and network identifiers.
type flagValues struct {
	mode            envite.ExecutionMode // Execution mode determines how the application will run.
	file            stringFlag           // File path to an environment YAML file.
	port            stringFlag           // Port number for the Web UI in daemon mode.
	envID           stringFlag           // Environment ID to override the default provided in the environment file.
	dockerNetworkID stringFlag           // Docker network identifier for environments with Docker components.
}

// parseFlags parses command-line arguments into flagValues.
// It initializes flagValues with command-line options for the application's configuration.
// This function also validates the execution mode and exits the program with an error message if the mode is invalid.
// Returns a populated flagValues struct with the parsed flags.
func parseFlags() flagValues {
	f := flagValues{}

	flag.Var(&f.file, "file", "Path to an environment yaml file (default: `envite.yml`)")
	flag.Var(&f.port, "port", "Web UI port to be used if mode is daemon (default: `4005`)")
	flag.Var(&f.envID, "id", "Override the environment ID provided by the environment yaml")
	flag.Var(&f.dockerNetworkID, "network", "Docker network identifier to be used. "+
		"Used only if docker components exist in the environment file. If not provided, ENVITE will create "+
		"a dedicated open docker network.")

	flag.Parse()
	mode, err := envite.ParseExecutionMode(flag.Arg(0))
	if err != nil {
		fmt.Printf("%s. %s\n", err.Error(), envite.DescribeAvailableModes())
		os.Exit(1)
	}

	f.mode = mode

	return f
}

// stringFlag is a custom flag type that supports checking the existence of a flag.
// It holds a string value and a boolean indicating whether the flag was provided.
type stringFlag struct {
	exist bool
	value string
}

func (s *stringFlag) Set(value string) error {
	s.value = value
	s.exist = true
	return nil
}

func (s *stringFlag) String() string {
	if s != nil {
		return s.value
	}
	return ""
}
