// Copyright 2024 HUMAN Security.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package envite

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

const (
	chanBufferSize = 100
	timeFormat     = "2006-01-02T15:04:05.000000000Z07:00"
)

// outputManager is responsible for managing and distributing log output messages.
type outputManager struct {
	lock     sync.Mutex
	messages [][]byte
	readers  []*Reader
}

// newOutputManager creates a new instance of outputManager.
func newOutputManager() *outputManager {
	return &outputManager{}
}

// write logs a message with the given timestamp, component, and message content.
func (o *outputManager) write(t time.Time, component, message string) {
	data := []byte(fmt.Sprintf("<component>%s<time>%s<msg>%s\n", component, t.Local().Format(timeFormat), message))
	o.lock.Lock()
	defer o.lock.Unlock()
	o.messages = append(o.messages, data)
	for _, reader := range o.readers {
		reader.ch <- data
	}
}

// reader creates and returns a new Reader instance to read log messages.
func (o *outputManager) reader() *Reader {
	o.lock.Lock()
	defer o.lock.Unlock()

	ch := make(chan []byte, chanBufferSize)
	messages := o.messages

	go func() {
		for _, message := range messages {
			ch <- message
		}
	}()

	reader := &Reader{ch: ch}
	o.readers = append(o.readers, reader)
	reader.close = func() {
		o.lock.Lock()
		defer o.lock.Unlock()
		for i, current := range o.readers {
			if current == reader {
				o.readers = append(o.readers[:i], o.readers[i+1:]...)
				return
			}
		}
	}

	return reader
}

// writer creates and returns a new Writer instance to write log messages for a specific component.
func (o *outputManager) writer(component string) *Writer {
	return &Writer{
		component:     component,
		outputManager: o,
	}
}

// Reader represents a reader for log messages.
type Reader struct {
	ch    chan []byte
	close func()
}

// Chan returns the channel for receiving log messages.
func (o *Reader) Chan() chan []byte {
	return o.ch
}

// Close closes the log message reader.
func (o *Reader) Close() error {
	o.close()
	return nil
}

// Writer represents a writer for log messages.
// Example:
//
//	writer.WriteString(writer.Color.Red("warning!"))
//
// this writes a red messages to the console with the text "warning!"
type Writer struct {
	Color AnsiColor

	component     string
	outputManager *outputManager
}

// Write writes a log message with the current timestamp.
func (w *Writer) Write(message []byte) {
	w.WriteString(string(message))
}

// WriteWithTime writes a log message with a specified timestamp.
func (w *Writer) WriteWithTime(t time.Time, message []byte) {
	w.WriteStringWithTime(t, string(message))
}

// WriteString writes a log message with the current timestamp.
func (w *Writer) WriteString(message string) {
	w.WriteStringWithTime(time.Now(), message)
}

// WriteStringWithTime writes a log message with a specified timestamp.
func (w *Writer) WriteStringWithTime(t time.Time, message string) {
	if strings.HasSuffix(message, "\r\n") {
		message = message[:len(message)-2]
	}
	if strings.HasSuffix(message, "\n") {
		message = message[:len(message)-1]
	}
	w.outputManager.write(t, w.component, message)
}

// AnsiColor provides ANSI color codes for console output.
type AnsiColor struct{}

// Red applies red color to the given string.
func (a AnsiColor) Red(s string) string {
	return fmt.Sprintf("\u001B[31m%s\u001B[39m", s)
}

// Green applies green color to the given string.
func (a AnsiColor) Green(s string) string {
	return fmt.Sprintf("\u001B[32m%s\u001B[39m", s)
}

// Yellow applies yellow color to the given string.
func (a AnsiColor) Yellow(s string) string {
	return fmt.Sprintf("\u001B[33m%s\u001B[39m", s)
}

// Blue applies blue color to the given string.
func (a AnsiColor) Blue(s string) string {
	return fmt.Sprintf("\u001B[34m%s\u001B[39m", s)
}

// Magenta applies magenta color to the given string.
func (a AnsiColor) Magenta(s string) string {
	return fmt.Sprintf("\u001B[35m%s\u001B[39m", s)
}

// Cyan applies cyan color to the given string.
func (a AnsiColor) Cyan(s string) string {
	return fmt.Sprintf("\u001B[36m%s\u001B[39m", s)
}
