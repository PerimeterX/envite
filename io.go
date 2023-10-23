package fengshui

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

type outputManager struct {
	lock     sync.Mutex
	messages [][]byte
	readers  []*Reader
}

func newOutputManager() *outputManager {
	return &outputManager{}
}

func (o *outputManager) write(t time.Time, component, message string) {
	data := []byte(fmt.Sprintf("<component>%s<time>%s<msg>%s\n", component, t.Local().Format(timeFormat), message))
	o.lock.Lock()
	defer o.lock.Unlock()
	o.messages = append(o.messages, data)
	for _, reader := range o.readers {
		reader.ch <- data
	}
}

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

func (o *outputManager) writer(component string) *Writer {
	return &Writer{
		component:     component,
		outputManager: o,
	}
}

type Reader struct {
	ch    chan []byte
	close func()
}

func (o *Reader) Chan() chan []byte {
	return o.ch
}

func (o *Reader) Close() error {
	o.close()
	return nil
}

type Writer struct {
	Color AnsiColor

	component     string
	outputManager *outputManager
}

func (w *Writer) Write(message []byte) {
	w.WriteString(string(message))
}

func (w *Writer) WriteWithTime(t time.Time, message []byte) {
	w.WriteStringWithTime(t, string(message))
}

func (w *Writer) WriteString(message string) {
	w.WriteStringWithTime(time.Now(), message)
}

func (w *Writer) WriteStringWithTime(t time.Time, message string) {
	if strings.HasSuffix(message, "\r\n") {
		message = message[:len(message)-2]
	}
	if strings.HasSuffix(message, "\n") {
		message = message[:len(message)-1]
	}
	w.outputManager.write(t, w.component, message)
}

type AnsiColor struct{}

func (a AnsiColor) Red(s string) string {
	return fmt.Sprintf("\u001B[31m%s\u001B[39m", s)
}

func (a AnsiColor) Green(s string) string {
	return fmt.Sprintf("\u001B[32m%s\u001B[39m", s)
}

func (a AnsiColor) Yellow(s string) string {
	return fmt.Sprintf("\u001B[33m%s\u001B[39m", s)
}

func (a AnsiColor) Blue(s string) string {
	return fmt.Sprintf("\u001B[34m%s\u001B[39m", s)
}

func (a AnsiColor) Magenta(s string) string {
	return fmt.Sprintf("\u001B[35m%s\u001B[39m", s)
}

func (a AnsiColor) Cyan(s string) string {
	return fmt.Sprintf("\u001B[36m%s\u001B[39m", s)
}
