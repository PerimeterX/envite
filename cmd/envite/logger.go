// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/perimeterx/envite"
	"os"
)

// logger provides a simple logging function for the application, outputting messages to the standard output.
// It prefixes log messages with their log level, except for info level messages, to make it easier to distinguish
// the severity of log messages. Fatal log messages cause the application to exit with a status code of 1.
// This logger is intended for the CLI.
func logger(level envite.LogLevel, message string) {
	var levelPrefix string
	if level != envite.LogLevelInfo {
		levelPrefix = fmt.Sprintf("[%s] ", level)
		fmt.Println(levelPrefix + message)
	}
	if level == envite.LogLevelFatal {
		os.Exit(1)
	}
}
