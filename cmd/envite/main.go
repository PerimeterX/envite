// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/perimeterx/envite"
	"os"
)

// main is the entry point of the CLI.
// It executes the main application logic and exits with status code 1 in case of an error.
func main() {
	err := exec()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

// exec orchestrates the execution flow of the application.
// It parses command-line flags, initializes the environment,
// and starts the server based on the provided configuration.
// Returns an error if any step in the process fails.
func exec() error {
	flags := parseFlags()
	env, err := buildEnv(flags)
	if err != nil {
		return err
	}

	server := buildServer(env, flags)
	return envite.Execute(server, flags.mode)
}
